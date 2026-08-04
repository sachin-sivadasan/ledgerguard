package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// DuplicateJobError is returned when an active job already exists for the same app+type
var ErrDuplicateJob = errors.New("an active sync job already exists for this app and type")

// JobProgress contains progress info merged from DB + Redis
type JobProgress struct {
	Job       *entity.SyncJob `json:"job"`
	Total     int             `json:"total"`
	Completed int             `json:"completed"`
	Message   string          `json:"message"`
	Children  []*JobProgress  `json:"children,omitempty"`
}

// QueueSyncService manages async sync job lifecycle
type QueueSyncService struct {
	syncJobRepo repository.SyncJobRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
	redisClient *redis.Client
	lockManager *queue.LockManager
	progress    *queue.ProgressTracker
	tracker     domainservice.EventTracker
}

// SetTracker sets the event tracker for sync lifecycle events.
func (s *QueueSyncService) SetTracker(t domainservice.EventTracker) {
	s.tracker = t
}

// NewQueueSyncService creates a new queue sync service
func NewQueueSyncService(
	syncJobRepo repository.SyncJobRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	redisClient *redis.Client,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *QueueSyncService {
	return &QueueSyncService{
		syncJobRepo: syncJobRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
		redisClient: redisClient,
		lockManager: lockManager,
		progress:    progress,
	}
}

// EnqueueSync creates a sync job and enqueues it. Returns 409 on duplicate.
func (s *QueueSyncService) EnqueueSync(ctx context.Context, appID, userID, partnerAccountID uuid.UUID, jobType string, priority int) (*entity.SyncJob, error) {
	// Check for duplicate active job
	existing, err := s.syncJobRepo.FindActiveByAppIDAndType(ctx, appID, jobType)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing job: %w", err)
	}
	if existing != nil {
		return nil, ErrDuplicateJob
	}

	// Create job
	job := entity.NewSyncJob(appID, userID, partnerAccountID, jobType, priority)
	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create sync job: %w", err)
	}

	// Enqueue to Redis
	payload := &queue.SyncJobPayload{
		JobID:            job.ID,
		AppID:            job.AppID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          job.JobType,
		Priority:         job.Priority,
		EnqueuedAt:       time.Now().UTC(),
	}

	if err := queue.Enqueue(ctx, s.redisClient, payload); err != nil {
		// Mark the DB row as failed if Redis enqueue fails
		_ = s.syncJobRepo.MarkFailed(ctx, job.ID, fmt.Sprintf("enqueue failed: %v", err))
		return nil, fmt.Errorf("failed to enqueue sync job: %w", err)
	}

	if s.tracker != nil {
		s.tracker.Track(ctx, userID.String(), "sync_started", domainservice.EventProperties{
			"job_type": jobType,
			"app_id":   appID.String(),
			"job_id":   job.ID.String(),
		})
	}

	return job, nil
}

// EnqueueCatchupSync enqueues a sync job with a lookback window.
// Silently skips if a duplicate job is already running.
func (s *QueueSyncService) EnqueueCatchupSync(ctx context.Context, appID, userID, partnerAccountID uuid.UUID, jobType string, lookbackDays int) (*entity.SyncJob, error) {
	// Check for duplicate active job — silently skip
	existing, err := s.syncJobRepo.FindActiveByAppIDAndType(ctx, appID, jobType)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing job: %w", err)
	}
	if existing != nil {
		return nil, nil // already running — silently skip
	}

	// Create job
	job := entity.NewSyncJob(appID, userID, partnerAccountID, jobType, 0)
	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create sync job: %w", err)
	}

	// Enqueue to Redis with lookback days
	payload := &queue.SyncJobPayload{
		JobID:            job.ID,
		AppID:            job.AppID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          job.JobType,
		LookbackDays:     lookbackDays,
		Priority:         job.Priority,
		EnqueuedAt:       time.Now().UTC(),
	}

	if err := queue.Enqueue(ctx, s.redisClient, payload); err != nil {
		_ = s.syncJobRepo.MarkFailed(ctx, job.ID, fmt.Sprintf("enqueue failed: %v", err))
		return nil, fmt.Errorf("failed to enqueue sync job: %w", err)
	}

	return job, nil
}

// TriggerSync enqueues a high-priority full_sync job for a newly-selected app.
// Duplicate jobs are silently ignored.
func (s *QueueSyncService) TriggerSync(ctx context.Context, appID, userID, partnerAccountID uuid.UUID) error {
	_, err := s.EnqueueSync(ctx, appID, userID, partnerAccountID, entity.SyncJobTypeFullSync, 1)
	if errors.Is(err, ErrDuplicateJob) {
		return nil // already running — not an error
	}
	return err
}

// GetJobStatus returns a job by ID from DB
func (s *QueueSyncService) GetJobStatus(ctx context.Context, jobID uuid.UUID) (*entity.SyncJob, error) {
	return s.syncJobRepo.FindByID(ctx, jobID)
}

// GetJobProgress returns job progress merged from DB + Redis, with children if full_sync
func (s *QueueSyncService) GetJobProgress(ctx context.Context, jobID uuid.UUID) (*JobProgress, error) {
	job, err := s.syncJobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	progress := &JobProgress{
		Job:       job,
		Total:     job.TotalItems,
		Completed: job.CompletedItems,
	}

	// Overlay Redis progress if job is still processing
	if job.Status == entity.SyncJobStatusProcessing {
		redisProgress, err := s.progress.GetProgress(ctx, jobID)
		if err == nil && redisProgress != nil {
			progress.Total = redisProgress.Total
			progress.Completed = redisProgress.Completed
			progress.Message = redisProgress.Message
		}
	}

	// Include children for full_sync jobs
	if job.JobType == entity.SyncJobTypeFullSync {
		children, err := s.syncJobRepo.FindByParentJobID(ctx, jobID)
		if err == nil {
			for _, child := range children {
				childProgress := &JobProgress{
					Job:       child,
					Total:     child.TotalItems,
					Completed: child.CompletedItems,
				}

				if child.Status == entity.SyncJobStatusProcessing {
					rp, err := s.progress.GetProgress(ctx, child.ID)
					if err == nil && rp != nil {
						childProgress.Total = rp.Total
						childProgress.Completed = rp.Completed
						childProgress.Message = rp.Message
					}
				}

				progress.Children = append(progress.Children, childProgress)
			}
		}
	}

	return progress, nil
}

// ListJobs returns paginated sync jobs for an app
func (s *QueueSyncService) ListJobs(ctx context.Context, appID uuid.UUID, status, jobType string, limit, offset int) ([]*entity.SyncJob, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.syncJobRepo.ListByAppID(ctx, appID, status, jobType, limit, offset)
}

// CancelJob requests cooperative cancellation of a running job
func (s *QueueSyncService) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	job, err := s.syncJobRepo.FindByID(ctx, jobID)
	if err != nil {
		return err
	}

	if job.IsTerminal() {
		return fmt.Errorf("job is already in terminal state: %s", job.Status)
	}

	// Set cancellation flag in Redis
	if err := s.lockManager.RequestCancellation(ctx, jobID); err != nil {
		return fmt.Errorf("failed to request cancellation: %w", err)
	}

	// Also cancel children for full_sync
	if job.JobType == entity.SyncJobTypeFullSync {
		children, err := s.syncJobRepo.FindByParentJobID(ctx, jobID)
		if err == nil {
			for _, child := range children {
				if !child.IsTerminal() {
					_ = s.lockManager.RequestCancellation(ctx, child.ID)
				}
			}
		}
	}

	return s.syncJobRepo.UpdateStatus(ctx, jobID, entity.SyncJobStatusCancelled)
}

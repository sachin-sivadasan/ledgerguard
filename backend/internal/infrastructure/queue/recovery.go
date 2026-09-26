package queue

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"go.uber.org/zap"
)

const recoveryGracePeriod = 2 * time.Minute

// RecoveryService handles recovery of stuck or lost sync jobs
type RecoveryService struct {
	syncJobRepo repository.SyncJobRepository
	client      *redis.Client
	lockManager *LockManager
	interval    time.Duration
	stopCh      chan struct{}
}

// NewRecoveryService creates a new recovery service
func NewRecoveryService(
	syncJobRepo repository.SyncJobRepository,
	client *redis.Client,
	lockManager *LockManager,
	interval time.Duration,
) *RecoveryService {
	return &RecoveryService{
		syncJobRepo: syncJobRepo,
		client:      client,
		lockManager: lockManager,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

// RecoverOnStartup re-enqueues stuck jobs on server startup
func (rs *RecoveryService) RecoverOnStartup(ctx context.Context) {
	// Check Redis queue depths
	regularLen, _ := rs.client.LLen(ctx, RegularQueueKey).Result()
	fullLen, _ := rs.client.LLen(ctx, FullSyncQueueKey).Result()
	logging.FromContext(ctx).Info("recovery: queue depths",
		zap.Int64("regular", regularLen),
		zap.Int64("full_sync", fullLen))

	// Re-enqueue processing jobs without heartbeat
	processingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusProcessing)
	if err != nil {
		logging.FromContext(ctx).Error("recovery: failed to find processing jobs", zap.Error(err))
		return
	}

	staleCount := 0
	aliveCount := 0
	recovered := 0
	recoveredIDs := make(map[uuid.UUID]bool)
	// First pass: identify all stale jobs
	staleJobs := make([]*entity.SyncJob, 0)
	for _, job := range processingJobs {
		hasHB, err := rs.lockManager.HasHeartbeat(ctx, job.ID)
		if err != nil || hasHB {
			aliveCount++
			continue
		}
		staleCount++
		staleJobs = append(staleJobs, job)
		recoveredIDs[job.ID] = true
	}

	// Second pass: only re-enqueue parent jobs (full_sync) and orphaned jobs.
	// Skip child jobs whose parent is also being recovered — the parent will recreate them.
	for _, job := range staleJobs {
		if job.ParentJobID != nil && recoveredIDs[*job.ParentJobID] {
			logging.FromContext(ctx).Warn("recovery: skipping child job, parent will recreate it",
				zap.String("job_id", job.ID.String()),
				zap.String("job_type", job.JobType),
				zap.String("parent_job_id", job.ParentJobID.String()))
			// Mark back to failed so it doesn't get re-enqueued again
			_ = rs.syncJobRepo.MarkFailed(ctx, job.ID, "parent recovered — will be recreated")
			_ = rs.lockManager.ForceReleaseLock(ctx, job.AppID, job.JobType)
			_ = rs.lockManager.DeleteHeartbeat(ctx, job.ID)
			continue
		}

		_ = rs.lockManager.ForceReleaseLock(ctx, job.AppID, job.JobType)
		_ = rs.lockManager.DeleteHeartbeat(ctx, job.ID)
		logging.FromContext(ctx).Info("recovery: re-enqueuing stale job, no heartbeat",
			zap.String("job_id", job.ID.String()),
			zap.String("job_type", job.JobType),
			zap.String("app_id", job.AppID.String()))
		if err := rs.reEnqueueJob(ctx, job); err != nil {
			logging.FromContext(ctx).Error("recovery: failed to re-enqueue job",
				zap.String("job_id", job.ID.String()),
				zap.Error(err))
			continue
		}
		recovered++
	}
	logging.FromContext(ctx).Info("recovery: processing jobs summary",
		zap.Int("total", len(processingJobs)),
		zap.Int("alive", aliveCount),
		zap.Int("stale", staleCount),
		zap.Int("recovered", recovered))

	// Re-enqueue pending jobs (handles Redis flush scenarios)
	// Bug 14 fix: Do NOT release locks for pending jobs — they shouldn't hold locks.
	// If a lock exists for the same app+type, it belongs to an active worker.
	pendingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusPending)
	if err != nil {
		logging.FromContext(ctx).Error("recovery: failed to find pending jobs", zap.Error(err))
		return
	}

	pendingRecovered := 0
	for _, job := range pendingJobs {
		if recoveredIDs[job.ID] {
			continue // Already handled in stale processing recovery
		}
		// Skip pending child jobs whose parent is being recovered
		if job.ParentJobID != nil && recoveredIDs[*job.ParentJobID] {
			logging.FromContext(ctx).Warn("recovery: skipping pending child job, parent will recreate it",
				zap.String("job_id", job.ID.String()),
				zap.String("job_type", job.JobType),
				zap.String("parent_job_id", job.ParentJobID.String()))
			_ = rs.syncJobRepo.MarkFailed(ctx, job.ID, "parent recovered — will be recreated")
			continue
		}
		// Just re-enqueue to Redis without touching locks or status (already pending)
		payload := &SyncJobPayload{
			JobID:            job.ID,
			AppID:            job.AppID,
			UserID:           job.UserID,
			PartnerAccountID: job.PartnerAccountID,
			JobType:          job.JobType,
			ParentJobID:      job.ParentJobID,
			Priority:         job.Priority,
			EntityType:       job.EntityType,
			EnqueuedAt:       time.Now().UTC(),
		}
		if err := Enqueue(ctx, rs.client, payload); err != nil {
			logging.FromContext(ctx).Error("recovery: failed to re-enqueue pending job",
				zap.String("job_id", job.ID.String()),
				zap.Error(err))
			continue
		}
		pendingRecovered++
	}
	if pendingRecovered > 0 {
		logging.FromContext(ctx).Info("recovery: re-enqueued orphaned pending jobs",
			zap.Int("count", pendingRecovered))
	}

	logging.FromContext(ctx).Info("recovery: startup complete",
		zap.Int("processing_recovered", recovered),
		zap.Int("pending_recovered", pendingRecovered))
}

// StartPeriodicRecovery runs periodic recovery checks
func (rs *RecoveryService) StartPeriodicRecovery(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(rs.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rs.recoverStuckJobs(ctx)
			case <-rs.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the periodic recovery
func (rs *RecoveryService) Stop() {
	close(rs.stopCh)
}

func (rs *RecoveryService) recoverStuckJobs(ctx context.Context) {
	processingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusProcessing)
	if err != nil {
		logging.FromContext(ctx).Error("recovery: periodic check failed", zap.Error(err))
		return
	}

	recovered := 0
	for _, job := range processingJobs {
		// Bug 5 fix: Grace period — skip jobs that started recently
		// (worker may not have written first heartbeat yet)
		if job.StartedAt != nil && time.Since(*job.StartedAt) < recoveryGracePeriod {
			continue
		}

		// Check if job has been processing too long without heartbeat
		if job.StartedAt != nil && time.Since(*job.StartedAt) > lockTTL {
			hasHB, err := rs.lockManager.HasHeartbeat(ctx, job.ID)
			if err != nil || hasHB {
				continue
			}

			_ = rs.lockManager.ForceReleaseLock(ctx, job.AppID, job.JobType)
			_ = rs.lockManager.DeleteHeartbeat(ctx, job.ID)
			logging.FromContext(ctx).Info("recovery: periodic re-enqueuing dead job",
				zap.String("job_id", job.ID.String()),
				zap.String("job_type", job.JobType),
				zap.String("app_id", job.AppID.String()),
				zap.String("started", job.StartedAt.Format(time.RFC3339)))
			if err := rs.reEnqueueJob(ctx, job); err != nil {
				logging.FromContext(ctx).Error("recovery: failed to re-enqueue stuck job",
					zap.String("job_id", job.ID.String()),
					zap.Error(err))
				continue
			}
			recovered++
		}
	}

	if recovered > 0 {
		logging.FromContext(ctx).Info("recovery: periodic re-enqueued stuck jobs",
			zap.Int("count", recovered))
	}
}

// Bug 10 fix: reEnqueueJob uses conditional status update to avoid racing with workers
func (rs *RecoveryService) reEnqueueJob(ctx context.Context, job *entity.SyncJob) error {
	// Only reset to pending if currently processing (conditional update)
	if job.Status == entity.SyncJobStatusProcessing {
		if err := rs.syncJobRepo.MarkPendingIfProcessing(ctx, job.ID); err != nil {
			return err // Job already moved to another state — skip
		}
	}
	// If already pending, skip status update

	payload := &SyncJobPayload{
		JobID:            job.ID,
		AppID:            job.AppID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          job.JobType,
		ParentJobID:      job.ParentJobID,
		Priority:         job.Priority,
		EntityType:       job.EntityType,
		EnqueuedAt:       time.Now().UTC(),
	}

	return Enqueue(ctx, rs.client, payload)
}

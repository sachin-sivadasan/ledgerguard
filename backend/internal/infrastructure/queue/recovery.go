package queue

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

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
	// Re-enqueue processing jobs without heartbeat
	processingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusProcessing)
	if err != nil {
		log.Printf("Recovery: failed to find processing jobs: %v", err)
		return
	}

	recovered := 0
	for _, job := range processingJobs {
		hasHB, err := rs.lockManager.HasHeartbeat(ctx, job.ID)
		if err != nil || hasHB {
			continue // Still alive or can't check
		}

		// No heartbeat → re-enqueue
		if err := rs.reEnqueueJob(ctx, job); err != nil {
			log.Printf("Recovery: failed to re-enqueue job %s: %v", job.ID, err)
			continue
		}
		recovered++
	}

	// Re-enqueue pending jobs (handles Redis flush scenarios)
	pendingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusPending)
	if err != nil {
		log.Printf("Recovery: failed to find pending jobs: %v", err)
		return
	}

	for _, job := range pendingJobs {
		if err := rs.reEnqueueJob(ctx, job); err != nil {
			log.Printf("Recovery: failed to re-enqueue pending job %s: %v", job.ID, err)
			continue
		}
		recovered++
	}

	if recovered > 0 {
		log.Printf("Recovery: re-enqueued %d jobs on startup", recovered)
	}
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
		log.Printf("Recovery: periodic check failed: %v", err)
		return
	}

	recovered := 0
	for _, job := range processingJobs {
		// Check if job has been processing too long without heartbeat
		if job.StartedAt != nil && time.Since(*job.StartedAt) > lockTTL {
			hasHB, err := rs.lockManager.HasHeartbeat(ctx, job.ID)
			if err != nil || hasHB {
				continue
			}

			if err := rs.reEnqueueJob(ctx, job); err != nil {
				log.Printf("Recovery: failed to re-enqueue stuck job %s: %v", job.ID, err)
				continue
			}
			recovered++
		}
	}

	if recovered > 0 {
		log.Printf("Recovery: re-enqueued %d stuck jobs", recovered)
	}
}

func (rs *RecoveryService) reEnqueueJob(ctx context.Context, job *entity.SyncJob) error {
	// Reset status to pending
	if err := rs.syncJobRepo.UpdateStatus(ctx, job.ID, entity.SyncJobStatusPending); err != nil {
		return err
	}

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

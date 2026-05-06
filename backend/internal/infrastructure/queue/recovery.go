package queue

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
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
	log.Printf("[queue] Recovery: queue depths — regular: %d, full_sync: %d", regularLen, fullLen)

	// Re-enqueue processing jobs without heartbeat
	processingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusProcessing)
	if err != nil {
		log.Printf("[queue] Recovery: failed to find processing jobs: %v", err)
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
			log.Printf("[queue] Recovery: skipping child job %s (type=%s) — parent %s will recreate it", job.ID, job.JobType, *job.ParentJobID)
			// Mark back to failed so it doesn't get re-enqueued again
			_ = rs.syncJobRepo.MarkFailed(ctx, job.ID, "parent recovered — will be recreated")
			_ = rs.lockManager.ForceReleaseLock(ctx, job.AppID, job.JobType)
			_ = rs.lockManager.DeleteHeartbeat(ctx, job.ID)
			continue
		}

		_ = rs.lockManager.ForceReleaseLock(ctx, job.AppID, job.JobType)
		_ = rs.lockManager.DeleteHeartbeat(ctx, job.ID)
		log.Printf("[queue] Recovery: re-enqueuing stale job %s (type=%s, app=%s) — no heartbeat", job.ID, job.JobType, job.AppID)
		if err := rs.reEnqueueJob(ctx, job); err != nil {
			log.Printf("[queue] Recovery: failed to re-enqueue job %s: %v", job.ID, err)
			continue
		}
		recovered++
	}
	log.Printf("[queue] Recovery: processing jobs — %d total, %d alive (heartbeat ok), %d stale (no heartbeat), %d recovered", len(processingJobs), aliveCount, staleCount, recovered)

	// Re-enqueue pending jobs (handles Redis flush scenarios)
	// Bug 14 fix: Do NOT release locks for pending jobs — they shouldn't hold locks.
	// If a lock exists for the same app+type, it belongs to an active worker.
	pendingJobs, err := rs.syncJobRepo.FindByStatus(ctx, entity.SyncJobStatusPending)
	if err != nil {
		log.Printf("[queue] Recovery: failed to find pending jobs: %v", err)
		return
	}

	pendingRecovered := 0
	for _, job := range pendingJobs {
		if recoveredIDs[job.ID] {
			continue // Already handled in stale processing recovery
		}
		// Skip pending child jobs whose parent is being recovered
		if job.ParentJobID != nil && recoveredIDs[*job.ParentJobID] {
			log.Printf("[queue] Recovery: skipping pending child job %s (type=%s) — parent %s will recreate it", job.ID, job.JobType, *job.ParentJobID)
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
			log.Printf("[queue] Recovery: failed to re-enqueue pending job %s: %v", job.ID, err)
			continue
		}
		pendingRecovered++
	}
	if pendingRecovered > 0 {
		log.Printf("[queue] Recovery: re-enqueued %d orphaned pending jobs", pendingRecovered)
	}

	log.Printf("[queue] Recovery: startup complete — %d processing + %d pending jobs recovered", recovered, pendingRecovered)
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
		log.Printf("[queue] Recovery: periodic check failed: %v", err)
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
			log.Printf("[queue] Recovery: periodic — re-enqueuing dead job %s (type=%s, app=%s, started=%s)",
				job.ID, job.JobType, job.AppID, job.StartedAt.Format(time.RFC3339))
			if err := rs.reEnqueueJob(ctx, job); err != nil {
				log.Printf("[queue] Recovery: failed to re-enqueue stuck job %s: %v", job.ID, err)
				continue
			}
			recovered++
		}
	}

	if recovered > 0 {
		log.Printf("[queue] Recovery: periodic — re-enqueued %d stuck jobs", recovered)
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

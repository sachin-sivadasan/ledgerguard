package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"go.uber.org/zap"
)

const dequeueTimeout = 5 * time.Second

// WorkerPool manages a pool of workers that process sync jobs from a queue
type WorkerPool struct {
	name        string
	queueKey    string
	numWorkers  int
	client      *redis.Client
	syncJobRepo repository.SyncJobRepository
	lockManager *LockManager
	progress    *ProgressTracker
	registry    *ProcessorRegistry
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(
	name string,
	queueKey string,
	numWorkers int,
	client *redis.Client,
	syncJobRepo repository.SyncJobRepository,
	lockManager *LockManager,
	progress *ProgressTracker,
	registry *ProcessorRegistry,
) *WorkerPool {
	return &WorkerPool{
		name:        name,
		queueKey:    queueKey,
		numWorkers:  numWorkers,
		client:      client,
		syncJobRepo: syncJobRepo,
		lockManager: lockManager,
		progress:    progress,
		registry:    registry,
	}
}

// Start launches all workers in the pool
func (wp *WorkerPool) Start(ctx context.Context) {
	ctx, wp.cancel = context.WithCancel(ctx)

	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		workerID := fmt.Sprintf("%s-worker-%d", wp.name, i)
		go wp.workerLoop(ctx, workerID)
	}

	logging.FromContext(ctx).Info("worker pool started",
		zap.String("name", wp.name),
		zap.Int("num_workers", wp.numWorkers),
		zap.String("queue_key", wp.queueKey))
}

// Stop gracefully shuts down all workers
func (wp *WorkerPool) Stop() {
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.wg.Wait()
	zap.L().Info("worker pool stopped", zap.String("name", wp.name))
}

func (wp *WorkerPool) workerLoop(ctx context.Context, workerID string) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		payload, err := Dequeue(ctx, wp.client, wp.queueKey, dequeueTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return // Context cancelled
			}
			logging.FromContext(ctx).Warn("dequeue error",
				zap.String("worker_id", workerID),
				zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if payload == nil {
			continue // Timeout, no item
		}

		wp.processJob(ctx, workerID, payload)
	}
}

func (wp *WorkerPool) processJob(ctx context.Context, workerID string, payload *SyncJobPayload) {
	jobID := payload.JobID

	// Bug 2 fix: Acquire lock BEFORE MarkStarted to avoid bouncing processing→pending
	locked, err := wp.lockManager.AcquireLock(ctx, payload.AppID, payload.JobType, workerID)
	if err != nil {
		logging.FromContext(ctx).Warn("failed to acquire lock, re-enqueuing with backoff",
			zap.String("worker_id", workerID),
			zap.String("job_id", jobID.String()),
			zap.Error(err))
		// Job is still pending — re-enqueue with backoff
		wp.reEnqueueWithBackoff(ctx, workerID, jobID, payload)
		return
	}
	if !locked {
		// Check if the lock holder is dead (no heartbeat) — if so, we can steal
		existingHolder, _ := wp.lockManager.GetLockHolder(ctx, payload.AppID, payload.JobType)
		if existingHolder != "" {
			existingJob, _ := wp.syncJobRepo.FindActiveByAppIDAndType(ctx, payload.AppID, payload.JobType)
			if existingJob != nil && existingJob.ID != jobID {
				hasHB, _ := wp.lockManager.HasHeartbeat(ctx, existingJob.ID)
				if !hasHB {
					// Bug 4 fix: Atomic steal instead of separate release+acquire
					locked, _ = wp.lockManager.StealLock(ctx, payload.AppID, payload.JobType, existingHolder, workerID)
				}
			}
		}
		if !locked {
			logging.FromContext(ctx).Warn("could not acquire lock, re-enqueuing with 5s backoff",
				zap.String("worker_id", workerID),
				zap.String("job_id", jobID.String()))
			wp.reEnqueueWithBackoff(ctx, workerID, jobID, payload)
			return
		}
	}

	// Write initial heartbeat immediately after lock acquisition.
	// This prevents the steal-lock race: another worker checking HasHeartbeat
	// between our lock acquisition and the heartbeat goroutine starting.
	_ = wp.lockManager.Heartbeat(ctx, jobID)

	// Now mark job as started (after lock acquired)
	if err := wp.syncJobRepo.MarkStarted(ctx, jobID, workerID); err != nil {
		logging.FromContext(ctx).Error("failed to mark job started",
			zap.String("worker_id", workerID),
			zap.String("job_id", jobID.String()),
			zap.Error(err))
		// Release the lock we just acquired
		_, _ = wp.lockManager.ReleaseLockIfOwner(ctx, payload.AppID, payload.JobType, workerID)
		_ = wp.lockManager.DeleteHeartbeat(ctx, jobID)
		return
	}

	// Start heartbeat goroutine (continues renewing the heartbeat we just wrote)
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	go wp.heartbeatLoop(heartbeatCtx, jobID, payload.AppID, payload.JobType, workerID)

	// Look up processor
	processor, err := wp.registry.Get(payload.JobType)
	if err != nil {
		cancelHeartbeat()
		logging.FromContext(ctx).Error("no processor for job type",
			zap.String("worker_id", workerID),
			zap.String("job_type", payload.JobType),
			zap.Error(err))
		failCtx, failCancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = wp.syncJobRepo.MarkFailed(failCtx, jobID, err.Error())
		wp.cleanup(failCtx, jobID, payload, workerID)
		failCancel()
		return
	}

	// Execute processor
	err = processor.Process(ctx, payload)

	// Stop heartbeat
	cancelHeartbeat()

	// Use a background context for final state transitions and cleanup.
	// The parent ctx may be cancelled (server shutdown), but we MUST still
	// persist the final job state and release locks to avoid stuck jobs.
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	// Bug 3 fix: Centralized state transitions — worker handles MarkCompleted/MarkFailed
	if err != nil {
		if ctx.Err() != nil {
			// Server shutdown interrupted this job. Leave in 'processing' state
			// so recovery re-enqueues it on next startup (no heartbeat = stale).
			logging.FromContext(ctx).Warn("job interrupted by shutdown, will recover on restart",
				zap.String("job_id", jobID.String()),
				zap.String("job_type", payload.JobType))
			wp.cleanup(cleanupCtx, jobID, payload, workerID)
			return
		}
		// Bug 13 fix: Check if job was cancelled — don't overwrite with "failed"
		if cancelled, _ := wp.lockManager.IsCancelled(cleanupCtx, jobID); cancelled {
			logging.FromContext(ctx).Warn("job was cancelled, skipping MarkFailed",
				zap.String("worker_id", workerID),
				zap.String("job_id", jobID.String()))
		} else {
			logging.FromContext(ctx).Error("job failed",
				zap.String("worker_id", workerID),
				zap.String("job_id", jobID.String()),
				zap.Error(err))
			_ = wp.syncJobRepo.MarkFailed(cleanupCtx, jobID, err.Error())
		}
	} else {
		_ = wp.syncJobRepo.MarkCompleted(cleanupCtx, jobID)
	}

	// Cleanup
	wp.cleanup(cleanupCtx, jobID, payload, workerID)
}

// reEnqueueWithBackoff re-enqueues a job after a backoff delay
func (wp *WorkerPool) reEnqueueWithBackoff(ctx context.Context, workerID string, jobID uuid.UUID, payload *SyncJobPayload) {
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		// Bug 7 fix: On ctx cancel during backoff, best-effort enqueue before returning
		if err := Enqueue(context.Background(), wp.client, payload); err != nil {
			logging.FromContext(ctx).Error("failed to re-enqueue job on shutdown",
				zap.String("worker_id", workerID),
				zap.String("job_id", jobID.String()),
				zap.Error(err))
		}
		return
	}
	// Bug 7 fix: Log enqueue errors
	if err := Enqueue(ctx, wp.client, payload); err != nil {
		logging.FromContext(ctx).Error("failed to re-enqueue job",
			zap.String("worker_id", workerID),
			zap.String("job_id", jobID.String()),
			zap.Error(err))
	}
}

func (wp *WorkerPool) heartbeatLoop(ctx context.Context, jobID, appID uuid.UUID, syncType, workerID string) {
	hbTicker := time.NewTicker(wp.lockManager.HeartbeatInterval())
	lockTicker := time.NewTicker(wp.lockManager.LockExtensionInterval())
	defer hbTicker.Stop()
	defer lockTicker.Stop()

	// Initial heartbeat
	_ = wp.lockManager.Heartbeat(ctx, jobID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-hbTicker.C:
			_ = wp.lockManager.Heartbeat(ctx, jobID)
		case <-lockTicker.C:
			// Bug 12 fix: Ownership-aware lock extension
			_, _ = wp.lockManager.ExtendLockIfOwner(ctx, appID, syncType, workerID)
		}
	}
}

func (wp *WorkerPool) cleanup(ctx context.Context, jobID uuid.UUID, payload *SyncJobPayload, workerID string) {
	// Bug 1 fix: Ownership-aware lock release
	_, _ = wp.lockManager.ReleaseLockIfOwner(ctx, payload.AppID, payload.JobType, workerID)
	_ = wp.lockManager.DeleteHeartbeat(ctx, jobID)
	_ = wp.lockManager.CleanupCancellation(ctx, jobID)
	wp.progress.Cleanup(ctx, jobID)
}

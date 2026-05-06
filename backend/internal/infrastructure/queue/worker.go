package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
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

	log.Printf("[queue] Worker pool %q started with %d workers on queue %s", wp.name, wp.numWorkers, wp.queueKey)
}

// Stop gracefully shuts down all workers
func (wp *WorkerPool) Stop() {
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.wg.Wait()
	log.Printf("[queue] Worker pool %q stopped", wp.name)
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
			log.Printf("[%s] dequeue error: %v", workerID, err)
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
		log.Printf("[%s] failed to acquire lock for job %s: %v", workerID, jobID, err)
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
			log.Printf("[%s] could not acquire lock for job %s, re-enqueuing with 5s backoff", workerID, jobID)
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
		log.Printf("[%s] failed to mark job %s started: %v", workerID, jobID, err)
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
		log.Printf("[%s] no processor for job type %q: %v", workerID, payload.JobType, err)
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
			log.Printf("[queue] job %s (%s) interrupted by shutdown — will recover on restart", jobID, payload.JobType)
			wp.cleanup(cleanupCtx, jobID, payload, workerID)
			return
		}
		// Bug 13 fix: Check if job was cancelled — don't overwrite with "failed"
		if cancelled, _ := wp.lockManager.IsCancelled(cleanupCtx, jobID); cancelled {
			log.Printf("[%s] job %s was cancelled, skipping MarkFailed", workerID, jobID)
		} else {
			log.Printf("[%s] job %s failed: %v", workerID, jobID, err)
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
			log.Printf("[%s] failed to re-enqueue job %s on shutdown: %v", workerID, jobID, err)
		}
		return
	}
	// Bug 7 fix: Log enqueue errors
	if err := Enqueue(ctx, wp.client, payload); err != nil {
		log.Printf("[%s] failed to re-enqueue job %s: %v", workerID, jobID, err)
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

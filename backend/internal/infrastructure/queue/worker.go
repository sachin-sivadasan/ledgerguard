package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
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

	log.Printf("Worker pool %q started with %d workers on queue %s", wp.name, wp.numWorkers, wp.queueKey)
}

// Stop gracefully shuts down all workers
func (wp *WorkerPool) Stop() {
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.wg.Wait()
	log.Printf("Worker pool %q stopped", wp.name)
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

	// Mark job as started
	if err := wp.syncJobRepo.MarkStarted(ctx, jobID, workerID); err != nil {
		log.Printf("[%s] failed to mark job %s started: %v", workerID, jobID, err)
		return
	}

	// Acquire lock for the app+type combo
	locked, err := wp.lockManager.AcquireLock(ctx, payload.AppID, payload.JobType, workerID)
	if err != nil {
		log.Printf("[%s] failed to acquire lock for job %s: %v", workerID, jobID, err)
		_ = wp.syncJobRepo.MarkFailed(ctx, jobID, fmt.Sprintf("lock error: %v", err))
		return
	}
	if !locked {
		// Check if the lock holder is dead (no heartbeat) — if so, we can steal
		existingJob, _ := wp.syncJobRepo.FindActiveByAppIDAndType(ctx, payload.AppID, payload.JobType)
		if existingJob != nil && existingJob.ID != jobID {
			hasHB, _ := wp.lockManager.HasHeartbeat(ctx, existingJob.ID)
			if !hasHB {
				// Steal the lock
				_ = wp.lockManager.ReleaseLock(ctx, payload.AppID, payload.JobType)
				locked, _ = wp.lockManager.AcquireLock(ctx, payload.AppID, payload.JobType, workerID)
			}
		}
		if !locked {
			log.Printf("[%s] could not acquire lock for job %s, re-enqueuing", workerID, jobID)
			_ = wp.syncJobRepo.UpdateStatus(ctx, jobID, entity.SyncJobStatusPending)
			_ = Enqueue(ctx, wp.client, payload)
			return
		}
	}

	// Start heartbeat goroutine
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	go wp.heartbeatLoop(heartbeatCtx, jobID, payload.AppID, payload.JobType)

	// Look up processor
	processor, err := wp.registry.Get(payload.JobType)
	if err != nil {
		cancelHeartbeat()
		log.Printf("[%s] no processor for job type %q: %v", workerID, payload.JobType, err)
		_ = wp.syncJobRepo.MarkFailed(ctx, jobID, err.Error())
		wp.cleanup(ctx, jobID, payload)
		return
	}

	// Execute processor
	err = processor.Process(ctx, payload)

	// Stop heartbeat
	cancelHeartbeat()

	// Handle result
	if err != nil {
		log.Printf("[%s] job %s failed: %v", workerID, jobID, err)
		_ = wp.syncJobRepo.MarkFailed(ctx, jobID, err.Error())
	}
	// Note: successful processors call MarkCompleted themselves

	// Cleanup
	wp.cleanup(ctx, jobID, payload)
}

func (wp *WorkerPool) heartbeatLoop(ctx context.Context, jobID, appID uuid.UUID, syncType string) {
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
			_ = wp.lockManager.ExtendLock(ctx, appID, syncType)
		}
	}
}

func (wp *WorkerPool) cleanup(ctx context.Context, jobID uuid.UUID, payload *SyncJobPayload) {
	_ = wp.lockManager.ReleaseLock(ctx, payload.AppID, payload.JobType)
	_ = wp.lockManager.DeleteHeartbeat(ctx, jobID)
	_ = wp.lockManager.CleanupCancellation(ctx, jobID)
	wp.progress.Cleanup(ctx, jobID)
}

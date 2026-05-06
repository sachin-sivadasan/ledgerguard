package queue

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// --- Worker-specific mock processor ---

type mockProcessor struct {
	jobType    string
	processErr error
	called     int32
	processFunc func(ctx context.Context, payload *SyncJobPayload) error
}

func (m *mockProcessor) Type() string { return m.jobType }
func (m *mockProcessor) Process(ctx context.Context, payload *SyncJobPayload) error {
	atomic.AddInt32(&m.called, 1)
	if m.processFunc != nil {
		return m.processFunc(ctx, payload)
	}
	return m.processErr
}

// --- Worker-specific extended mock repo that tracks all operations ---

type workerMockRepo struct {
	mockSyncJobRepo
	jobs       map[uuid.UUID]*entity.SyncJob
	startedIDs []uuid.UUID
	failedIDs  []uuid.UUID
}

func newWorkerMockRepo() *workerMockRepo {
	return &workerMockRepo{
		mockSyncJobRepo: mockSyncJobRepo{progressUpdates: make(map[uuid.UUID][2]int)},
		jobs:            make(map[uuid.UUID]*entity.SyncJob),
	}
}

func (m *workerMockRepo) Create(_ context.Context, job *entity.SyncJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *workerMockRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.SyncJob, error) {
	if j, ok := m.jobs[id]; ok {
		return j, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *workerMockRepo) FindByStatus(_ context.Context, status entity.SyncJobStatus) ([]*entity.SyncJob, error) {
	var result []*entity.SyncJob
	for _, j := range m.jobs {
		if j.Status == status {
			result = append(result, j)
		}
	}
	return result, nil
}

func (m *workerMockRepo) FindActiveByAppIDAndType(_ context.Context, appID uuid.UUID, jobType string) (*entity.SyncJob, error) {
	for _, j := range m.jobs {
		if j.AppID == appID && j.JobType == jobType && !j.IsTerminal() {
			return j, nil
		}
	}
	return nil, nil
}

func (m *workerMockRepo) FindByParentJobID(_ context.Context, _ uuid.UUID) ([]*entity.SyncJob, error) {
	return nil, nil
}

func (m *workerMockRepo) UpdateStatus(_ context.Context, id uuid.UUID, status entity.SyncJobStatus) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = status
	}
	return nil
}

func (m *workerMockRepo) MarkStarted(_ context.Context, id uuid.UUID, workerID string) error {
	m.startedIDs = append(m.startedIDs, id)
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusProcessing
		j.WorkerID = workerID
		now := time.Now().UTC()
		j.StartedAt = &now
	}
	return nil
}

func (m *workerMockRepo) MarkCompleted(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusCompleted
		now := time.Now().UTC()
		j.CompletedAt = &now
	}
	return nil
}

func (m *workerMockRepo) MarkFailed(_ context.Context, id uuid.UUID, errMsg string) error {
	m.failedIDs = append(m.failedIDs, id)
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusFailed
		j.ErrorMessage = errMsg
	}
	return nil
}

func (m *workerMockRepo) ListByAppID(_ context.Context, _ uuid.UUID, _ string, _ string, _, _ int) ([]*entity.SyncJob, int, error) {
	return nil, 0, nil
}

func (m *workerMockRepo) MarkPendingIfProcessing(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		if j.Status == entity.SyncJobStatusProcessing {
			j.Status = entity.SyncJobStatusPending
			j.WorkerID = ""
			j.StartedAt = nil
		}
	}
	return nil
}

// --- Tests ---

func TestWorkerPool_ProcessesJob(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	appID := uuid.New()
	proc := &mockProcessor{
		jobType: "transaction_sync",
		// Processor returns nil → worker will call MarkCompleted
	}
	registry.Register(proc)

	// Create and enqueue a job
	job := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            appID,
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[job.ID] = job

	payload := &SyncJobPayload{
		JobID:            job.ID,
		AppID:            appID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          "transaction_sync",
		EnqueuedAt:       time.Now().UTC(),
	}
	err := Enqueue(ctx, client, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Start worker pool with 1 worker
	wp := NewWorkerPool("test", RegularQueueKey, 1, client, repo, lm, pt, registry)
	wp.Start(ctx)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	wp.Stop()

	// Verify job was started
	if len(repo.startedIDs) == 0 {
		t.Error("Expected job to be marked as started")
	}

	// Verify processor was called
	if atomic.LoadInt32(&proc.called) == 0 {
		t.Error("Expected processor to be called")
	}

	// Verify job completed
	if repo.jobs[job.ID].Status != entity.SyncJobStatusCompleted {
		t.Errorf("Expected completed, got %s", repo.jobs[job.ID].Status)
	}
}

func TestWorkerPool_FailedProcessor(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	proc := &mockProcessor{
		jobType:    "transaction_sync",
		processErr: fmt.Errorf("processing failed"),
	}
	registry.Register(proc)

	appID := uuid.New()
	job := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            appID,
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[job.ID] = job

	err := Enqueue(ctx, client, &SyncJobPayload{
		JobID:            job.ID,
		AppID:            appID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          "transaction_sync",
		EnqueuedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	wp := NewWorkerPool("test", RegularQueueKey, 1, client, repo, lm, pt, registry)
	wp.Start(ctx)

	time.Sleep(500 * time.Millisecond)
	wp.Stop()

	// Verify job was marked as failed
	if len(repo.failedIDs) == 0 {
		t.Error("Expected job to be marked as failed")
	}
	if repo.jobs[job.ID].Status != entity.SyncJobStatusFailed {
		t.Errorf("Expected failed, got %s", repo.jobs[job.ID].Status)
	}
}

func TestWorkerPool_NoProcessor(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry() // empty — no processors registered

	appID := uuid.New()
	job := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            appID,
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "unknown_type",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[job.ID] = job

	err := Enqueue(ctx, client, &SyncJobPayload{
		JobID:            job.ID,
		AppID:            appID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          "unknown_type",
		EnqueuedAt:       time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	wp := NewWorkerPool("test", RegularQueueKey, 1, client, repo, lm, pt, registry)
	wp.Start(ctx)

	time.Sleep(500 * time.Millisecond)
	wp.Stop()

	// Should fail because no processor is registered
	if repo.jobs[job.ID].Status != entity.SyncJobStatusFailed {
		t.Errorf("Expected failed (no processor), got %s", repo.jobs[job.ID].Status)
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	proc := &mockProcessor{jobType: "transaction_sync"}
	registry.Register(proc)

	wp := NewWorkerPool("test", RegularQueueKey, 2, client, repo, lm, pt, registry)
	wp.Start(ctx)

	// Stop should return without hanging
	done := make(chan struct{})
	go func() {
		wp.Stop()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(10 * time.Second):
		t.Fatal("Stop() did not return in time — workers may be stuck")
	}
}

func TestWorkerPool_LockConflictReEnqueues(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	appID := uuid.New()

	proc := &mockProcessor{
		jobType: "transaction_sync",
		// Processor returns nil → worker will call MarkCompleted
	}
	registry.Register(proc)

	// Pre-acquire lock for this app+type and write a heartbeat so it's not stolen
	otherJobID := uuid.New()
	otherJob := &entity.SyncJob{
		ID:      otherJobID,
		AppID:   appID,
		JobType: "transaction_sync",
		Status:  entity.SyncJobStatusProcessing,
	}
	repo.jobs[otherJobID] = otherJob
	_, _ = lm.AcquireLock(ctx, appID, "transaction_sync", "other-worker")
	_ = lm.Heartbeat(ctx, otherJobID)

	// Test processJob directly instead of running the full worker loop
	// This avoids the tight re-enqueue loop with miniredis
	job := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            appID,
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[job.ID] = job

	payload := &SyncJobPayload{
		JobID:            job.ID,
		AppID:            appID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          "transaction_sync",
		EnqueuedAt:       time.Now().UTC(),
	}

	wp := NewWorkerPool("test", RegularQueueKey, 1, client, repo, lm, pt, registry)
	wp.processJob(ctx, "test-worker-0", payload)

	// Job should have been re-enqueued (status reset to pending)
	if repo.jobs[job.ID].Status != entity.SyncJobStatusPending {
		t.Errorf("Expected re-enqueued (pending), got %s", repo.jobs[job.ID].Status)
	}

	// The job should be in the queue
	got, err := Dequeue(ctx, client, RegularQueueKey, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if got == nil {
		t.Fatal("Expected job to be re-enqueued to Redis")
	}
	if got.JobID != job.ID {
		t.Errorf("Re-enqueued job ID mismatch: got %s, want %s", got.JobID, job.ID)
	}
}

func TestWorkerPool_MultipleWorkers(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	processedCount := int32(0)
	proc := &mockProcessor{
		jobType: "transaction_sync",
		processFunc: func(_ context.Context, _ *SyncJobPayload) error {
			atomic.AddInt32(&processedCount, 1)
			time.Sleep(50 * time.Millisecond) // Simulate work
			return nil
		},
	}
	registry.Register(proc)

	// Enqueue 5 jobs with different appIDs (so locks don't conflict)
	for i := 0; i < 5; i++ {
		appID := uuid.New()
		job := &entity.SyncJob{
			ID:               uuid.New(),
			AppID:            appID,
			UserID:           uuid.New(),
			PartnerAccountID: uuid.New(),
			JobType:          "transaction_sync",
			Status:           entity.SyncJobStatusPending,
			CreatedAt:        time.Now().UTC(),
		}
		repo.jobs[job.ID] = job

		_ = Enqueue(ctx, client, &SyncJobPayload{
			JobID:            job.ID,
			AppID:            appID,
			UserID:           job.UserID,
			PartnerAccountID: job.PartnerAccountID,
			JobType:          "transaction_sync",
			EnqueuedAt:       time.Now().UTC(),
		})
	}

	wp := NewWorkerPool("test", RegularQueueKey, 3, client, repo, lm, pt, registry)
	wp.Start(ctx)

	time.Sleep(2 * time.Second)
	wp.Stop()

	processed := atomic.LoadInt32(&processedCount)
	if processed != 5 {
		t.Errorf("Expected 5 jobs processed, got %d", processed)
	}
}

func TestWorkerPool_HeartbeatSetDuringProcessing(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	jobID := uuid.New()
	appID := uuid.New()

	heartbeatSeen := int32(0)
	proc := &mockProcessor{
		jobType: "transaction_sync",
		processFunc: func(pctx context.Context, payload *SyncJobPayload) error {
			// Give heartbeat goroutine time to write initial heartbeat
			time.Sleep(100 * time.Millisecond)
			has, err := lm.HasHeartbeat(pctx, payload.JobID)
			if err != nil {
				return err
			}
			if has {
				atomic.StoreInt32(&heartbeatSeen, 1)
			}
			return nil
		},
	}
	registry.Register(proc)

	job := &entity.SyncJob{
		ID:               jobID,
		AppID:            appID,
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[jobID] = job

	_ = Enqueue(ctx, client, &SyncJobPayload{
		JobID:            jobID,
		AppID:            appID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          "transaction_sync",
		EnqueuedAt:       time.Now().UTC(),
	})

	wp := NewWorkerPool("test", RegularQueueKey, 1, client, repo, lm, pt, registry)
	wp.Start(ctx)

	time.Sleep(1 * time.Second)
	wp.Stop()

	if atomic.LoadInt32(&heartbeatSeen) != 1 {
		t.Error("Expected heartbeat to be set during processing")
	}
	if repo.jobs[jobID].Status != entity.SyncJobStatusCompleted {
		t.Errorf("Expected completed, got %s", repo.jobs[jobID].Status)
	}
}

func TestWorkerPool_CleanupAfterProcessing(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newWorkerMockRepo()
	lm := NewLockManager(client)
	pt := NewProgressTracker(client, repo, 0, 0)
	registry := NewProcessorRegistry()

	jobID := uuid.New()
	appID := uuid.New()

	proc := &mockProcessor{
		jobType: "transaction_sync",
		// Processor returns nil → worker handles MarkCompleted
	}
	registry.Register(proc)

	job := &entity.SyncJob{
		ID:               jobID,
		AppID:            appID,
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[jobID] = job

	_ = Enqueue(ctx, client, &SyncJobPayload{
		JobID:            jobID,
		AppID:            appID,
		UserID:           job.UserID,
		PartnerAccountID: job.PartnerAccountID,
		JobType:          "transaction_sync",
		EnqueuedAt:       time.Now().UTC(),
	})

	wp := NewWorkerPool("test", RegularQueueKey, 1, client, repo, lm, pt, registry)
	wp.Start(ctx)

	time.Sleep(500 * time.Millisecond)
	wp.Stop()

	// Lock should be released
	ok, _ := lm.AcquireLock(ctx, appID, "transaction_sync", "check-worker")
	if !ok {
		t.Error("Expected lock to be released after processing")
	}

	// Heartbeat should be cleaned up
	has, _ := lm.HasHeartbeat(ctx, jobID)
	if has {
		t.Error("Expected heartbeat to be cleaned up after processing")
	}
}

package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// extendedMockRepo adds stateful tracking for recovery tests
type extendedMockRepo struct {
	mockSyncJobRepo
	jobs       map[uuid.UUID]*entity.SyncJob
	statusHist []entity.SyncJobStatus
}

func newExtendedMockRepo() *extendedMockRepo {
	return &extendedMockRepo{
		mockSyncJobRepo: mockSyncJobRepo{progressUpdates: make(map[uuid.UUID][2]int)},
		jobs:            make(map[uuid.UUID]*entity.SyncJob),
	}
}

func (m *extendedMockRepo) FindByStatus(_ context.Context, status entity.SyncJobStatus) ([]*entity.SyncJob, error) {
	var result []*entity.SyncJob
	for _, j := range m.jobs {
		if j.Status == status {
			result = append(result, j)
		}
	}
	return result, nil
}

func (m *extendedMockRepo) UpdateStatus(_ context.Context, id uuid.UUID, status entity.SyncJobStatus) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = status
		m.statusHist = append(m.statusHist, status)
	}
	return nil
}

func (m *extendedMockRepo) MarkPendingIfProcessing(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		if j.Status == entity.SyncJobStatusProcessing {
			j.Status = entity.SyncJobStatusPending
			j.WorkerID = ""
			j.StartedAt = nil
			m.statusHist = append(m.statusHist, entity.SyncJobStatusPending)
		}
	}
	return nil
}

func TestRecoverOnStartup_ReEnqueuesPendingJobs(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	repo := newExtendedMockRepo()

	// Add a pending job
	pendingJob := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            uuid.New(),
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusPending,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[pendingJob.ID] = pendingJob

	rs := NewRecoveryService(repo, client, lm, 10*time.Minute)
	rs.RecoverOnStartup(ctx)

	// Job should have been re-enqueued to Redis
	got, err := Dequeue(ctx, client, RegularQueueKey, 1*time.Second)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if got == nil {
		t.Fatal("Expected pending job to be re-enqueued")
	}
	if got.JobID != pendingJob.ID {
		t.Errorf("Job ID mismatch: got %s, want %s", got.JobID, pendingJob.ID)
	}
}

func TestRecoverOnStartup_ReEnqueuesProcessingWithoutHeartbeat(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	repo := newExtendedMockRepo()

	// Add a processing job with no heartbeat
	processingJob := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            uuid.New(),
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusProcessing,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[processingJob.ID] = processingJob

	rs := NewRecoveryService(repo, client, lm, 10*time.Minute)
	rs.RecoverOnStartup(ctx)

	// Should be re-enqueued (no heartbeat)
	got, err := Dequeue(ctx, client, RegularQueueKey, 1*time.Second)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if got == nil {
		t.Fatal("Expected processing job without heartbeat to be re-enqueued")
	}

	// Status should have been reset to pending
	if processingJob.Status != entity.SyncJobStatusPending {
		t.Errorf("Expected status reset to pending, got %s", processingJob.Status)
	}
}

func TestRecoverOnStartup_SkipsProcessingWithHeartbeat(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	repo := newExtendedMockRepo()

	// Add a processing job WITH heartbeat
	aliveJob := &entity.SyncJob{
		ID:               uuid.New(),
		AppID:            uuid.New(),
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Status:           entity.SyncJobStatusProcessing,
		CreatedAt:        time.Now().UTC(),
	}
	repo.jobs[aliveJob.ID] = aliveJob

	// Write heartbeat
	_ = lm.Heartbeat(ctx, aliveJob.ID)

	rs := NewRecoveryService(repo, client, lm, 10*time.Minute)
	rs.RecoverOnStartup(ctx)

	// Queue should be empty — alive job should NOT be re-enqueued
	got, err := Dequeue(ctx, client, RegularQueueKey, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if got != nil {
		t.Error("Expected alive processing job to NOT be re-enqueued")
	}

	// Status should still be processing
	if aliveJob.Status != entity.SyncJobStatusProcessing {
		t.Errorf("Expected status to remain processing, got %s", aliveJob.Status)
	}
}

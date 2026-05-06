package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// mockSyncJobRepo implements repository.SyncJobRepository for testing
type mockSyncJobRepo struct {
	progressUpdates map[uuid.UUID][2]int // jobID -> [total, completed]
}

func newMockSyncJobRepo() *mockSyncJobRepo {
	return &mockSyncJobRepo{
		progressUpdates: make(map[uuid.UUID][2]int),
	}
}

func (m *mockSyncJobRepo) Create(_ context.Context, _ *entity.SyncJob) error { return nil }
func (m *mockSyncJobRepo) FindByID(_ context.Context, _ uuid.UUID) (*entity.SyncJob, error) {
	return nil, nil
}
func (m *mockSyncJobRepo) FindByStatus(_ context.Context, _ entity.SyncJobStatus) ([]*entity.SyncJob, error) {
	return nil, nil
}
func (m *mockSyncJobRepo) FindActiveByAppIDAndType(_ context.Context, _ uuid.UUID, _ string) (*entity.SyncJob, error) {
	return nil, nil
}
func (m *mockSyncJobRepo) FindByParentJobID(_ context.Context, _ uuid.UUID) ([]*entity.SyncJob, error) {
	return nil, nil
}
func (m *mockSyncJobRepo) ListByAppID(_ context.Context, _ uuid.UUID, _ string, _ string, _, _ int) ([]*entity.SyncJob, int, error) {
	return nil, 0, nil
}
func (m *mockSyncJobRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ entity.SyncJobStatus) error {
	return nil
}
func (m *mockSyncJobRepo) UpdateProgress(_ context.Context, id uuid.UUID, total, completed int) error {
	m.progressUpdates[id] = [2]int{total, completed}
	return nil
}
func (m *mockSyncJobRepo) MarkStarted(_ context.Context, _ uuid.UUID, _ string) error  { return nil }
func (m *mockSyncJobRepo) MarkCompleted(_ context.Context, _ uuid.UUID) error          { return nil }
func (m *mockSyncJobRepo) MarkFailed(_ context.Context, _ uuid.UUID, _ string) error   { return nil }
func (m *mockSyncJobRepo) MarkPendingIfProcessing(_ context.Context, _ uuid.UUID) error { return nil }

func TestProgressTrackerUpdate(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newMockSyncJobRepo()
	// Use zero intervals so throttle doesn't block in test
	pt := NewProgressTracker(client, repo, 0, 0)

	jobID := uuid.New()
	pt.Update(ctx, jobID, Progress{Total: 100, Completed: 50, Message: "halfway"})

	// Verify Redis
	got, err := pt.GetProgress(ctx, jobID)
	if err != nil {
		t.Fatalf("GetProgress failed: %v", err)
	}
	if got == nil {
		t.Fatal("Expected progress in Redis")
	}
	if got.Total != 100 || got.Completed != 50 {
		t.Errorf("Progress mismatch: got %d/%d, want 100/50", got.Total, got.Completed)
	}
	if got.Message != "halfway" {
		t.Errorf("Message mismatch: got %q, want %q", got.Message, "halfway")
	}

	// Verify DB
	dbProgress, ok := repo.progressUpdates[jobID]
	if !ok {
		t.Error("Expected DB progress update")
	}
	if dbProgress[0] != 100 || dbProgress[1] != 50 {
		t.Errorf("DB progress mismatch: got %d/%d, want 100/50", dbProgress[0], dbProgress[1])
	}
}

func TestProgressTrackerThrottling(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newMockSyncJobRepo()
	// Long intervals to verify throttling
	pt := NewProgressTracker(client, repo, 1*time.Hour, 1*time.Hour)

	jobID := uuid.New()

	// First update goes through
	pt.Update(ctx, jobID, Progress{Total: 100, Completed: 10, Message: "first"})

	// Second update throttled
	pt.Update(ctx, jobID, Progress{Total: 100, Completed: 20, Message: "second"})

	got, _ := pt.GetProgress(ctx, jobID)
	if got != nil && got.Message != "first" {
		t.Errorf("Expected 'first' due to throttle, got %q", got.Message)
	}
}

func TestProgressTrackerForceUpdate(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newMockSyncJobRepo()
	pt := NewProgressTracker(client, repo, 1*time.Hour, 1*time.Hour)

	jobID := uuid.New()

	// ForceUpdate bypasses throttle
	pt.ForceUpdate(ctx, jobID, Progress{Total: 200, Completed: 200, Message: "done"})

	got, _ := pt.GetProgress(ctx, jobID)
	if got == nil {
		t.Fatal("Expected forced progress in Redis")
	}
	if got.Completed != 200 {
		t.Errorf("ForceUpdate: got %d, want 200", got.Completed)
	}
}

func TestProgressTrackerCleanup(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newMockSyncJobRepo()
	pt := NewProgressTracker(client, repo, 0, 0)

	jobID := uuid.New()
	pt.Update(ctx, jobID, Progress{Total: 10, Completed: 10, Message: "done"})

	pt.Cleanup(ctx, jobID)

	got, _ := pt.GetProgress(ctx, jobID)
	if got != nil {
		t.Error("Expected nil after cleanup")
	}
}

func TestGetProgressNotFound(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	repo := newMockSyncJobRepo()
	pt := NewProgressTracker(client, repo, 0, 0)

	got, err := pt.GetProgress(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetProgress error: %v", err)
	}
	if got != nil {
		t.Error("Expected nil for non-existent progress")
	}
}

package queue

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return mr, client
}

func TestEnqueueDequeue(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	payload := &SyncJobPayload{
		JobID:            uuid.New(),
		AppID:            uuid.New(),
		UserID:           uuid.New(),
		PartnerAccountID: uuid.New(),
		JobType:          "transaction_sync",
		Priority:         0,
		EnqueuedAt:       time.Now().UTC(),
	}

	// Enqueue
	err := Enqueue(ctx, client, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Dequeue
	got, err := Dequeue(ctx, client, RegularQueueKey, 1*time.Second)
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}
	if got == nil {
		t.Fatal("Dequeue returned nil")
	}

	if got.JobID != payload.JobID {
		t.Errorf("JobID mismatch: got %s, want %s", got.JobID, payload.JobID)
	}
	if got.JobType != payload.JobType {
		t.Errorf("JobType mismatch: got %s, want %s", got.JobType, payload.JobType)
	}
}

func TestQueueKeyRouting(t *testing.T) {
	tests := []struct {
		jobType string
		want    string
	}{
		{"full_sync", FullSyncQueueKey},
		{"transaction_sync", RegularQueueKey},
		{"snapshot_sync", RegularQueueKey},
		{"event_sync", RegularQueueKey},
		{"review_sync", RegularQueueKey},
	}

	for _, tt := range tests {
		t.Run(tt.jobType, func(t *testing.T) {
			got := QueueKeyForJobType(tt.jobType)
			if got != tt.want {
				t.Errorf("QueueKeyForJobType(%q) = %q, want %q", tt.jobType, got, tt.want)
			}
		})
	}
}

func TestDequeueTimeout(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	// Dequeue from empty queue should return nil (timeout)
	got, err := Dequeue(ctx, client, RegularQueueKey, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Dequeue error: %v", err)
	}
	if got != nil {
		t.Error("Expected nil from empty queue dequeue")
	}
}

func TestEnqueueFullSyncGoesToFullQueue(t *testing.T) {
	mr, client := setupMiniredis(t)
	ctx := context.Background()

	payload := &SyncJobPayload{
		JobID:   uuid.New(),
		JobType: "full_sync",
	}

	err := Enqueue(ctx, client, payload)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Verify it went to the full sync queue
	len := mr.Exists(FullSyncQueueKey)
	if !len {
		t.Error("Expected payload in full sync queue")
	}

	// Regular queue should be empty
	regularLen := mr.Exists(RegularQueueKey)
	if regularLen {
		t.Error("Regular queue should be empty")
	}
}

func TestFIFOOrder(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()

	ids := make([]uuid.UUID, 3)
	for i := range ids {
		ids[i] = uuid.New()
		err := Enqueue(ctx, client, &SyncJobPayload{
			JobID:   ids[i],
			JobType: "transaction_sync",
		})
		if err != nil {
			t.Fatalf("Enqueue %d failed: %v", i, err)
		}
	}

	// Dequeue should return in FIFO order (LPUSH + BRPOP = FIFO)
	for i, wantID := range ids {
		got, err := Dequeue(ctx, client, RegularQueueKey, 1*time.Second)
		if err != nil {
			t.Fatalf("Dequeue %d failed: %v", i, err)
		}
		if got.JobID != wantID {
			t.Errorf("Dequeue %d: got %s, want %s", i, got.JobID, wantID)
		}
	}
}

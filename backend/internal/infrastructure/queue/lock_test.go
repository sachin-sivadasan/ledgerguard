package queue

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestAcquireReleaseLock(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	appID := uuid.New()
	syncType := "transaction_sync"

	// Acquire lock
	ok, err := lm.AcquireLock(ctx, appID, syncType, "worker-1")
	if err != nil {
		t.Fatalf("AcquireLock failed: %v", err)
	}
	if !ok {
		t.Fatal("Expected lock to be acquired")
	}

	// Second acquire should fail (lock already held)
	ok, err = lm.AcquireLock(ctx, appID, syncType, "worker-2")
	if err != nil {
		t.Fatalf("AcquireLock (2) failed: %v", err)
	}
	if ok {
		t.Error("Expected second acquire to fail")
	}

	// Release
	err = lm.ReleaseLock(ctx, appID, syncType)
	if err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}

	// Now acquire should succeed again
	ok, err = lm.AcquireLock(ctx, appID, syncType, "worker-2")
	if err != nil {
		t.Fatalf("AcquireLock (3) failed: %v", err)
	}
	if !ok {
		t.Fatal("Expected lock to be acquired after release")
	}
}

func TestHeartbeat(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	jobID := uuid.New()

	// No heartbeat initially
	has, err := lm.HasHeartbeat(ctx, jobID)
	if err != nil {
		t.Fatalf("HasHeartbeat failed: %v", err)
	}
	if has {
		t.Error("Expected no heartbeat initially")
	}

	// Write heartbeat
	err = lm.Heartbeat(ctx, jobID)
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}

	// Now should have heartbeat
	has, err = lm.HasHeartbeat(ctx, jobID)
	if err != nil {
		t.Fatalf("HasHeartbeat (2) failed: %v", err)
	}
	if !has {
		t.Error("Expected heartbeat to exist")
	}

	// Delete heartbeat
	err = lm.DeleteHeartbeat(ctx, jobID)
	if err != nil {
		t.Fatalf("DeleteHeartbeat failed: %v", err)
	}

	has, err = lm.HasHeartbeat(ctx, jobID)
	if err != nil {
		t.Fatalf("HasHeartbeat (3) failed: %v", err)
	}
	if has {
		t.Error("Expected heartbeat to be gone after delete")
	}
}

func TestCancellation(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	jobID := uuid.New()

	// Not cancelled initially
	cancelled, err := lm.IsCancelled(ctx, jobID)
	if err != nil {
		t.Fatalf("IsCancelled failed: %v", err)
	}
	if cancelled {
		t.Error("Expected not cancelled initially")
	}

	// Request cancellation
	err = lm.RequestCancellation(ctx, jobID)
	if err != nil {
		t.Fatalf("RequestCancellation failed: %v", err)
	}

	cancelled, err = lm.IsCancelled(ctx, jobID)
	if err != nil {
		t.Fatalf("IsCancelled (2) failed: %v", err)
	}
	if !cancelled {
		t.Error("Expected cancelled after request")
	}

	// Cleanup
	err = lm.CleanupCancellation(ctx, jobID)
	if err != nil {
		t.Fatalf("CleanupCancellation failed: %v", err)
	}

	cancelled, err = lm.IsCancelled(ctx, jobID)
	if err != nil {
		t.Fatalf("IsCancelled (3) failed: %v", err)
	}
	if cancelled {
		t.Error("Expected not cancelled after cleanup")
	}
}

func TestLockIsolation(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	appID := uuid.New()

	// Locks for different sync types should not interfere
	ok1, _ := lm.AcquireLock(ctx, appID, "transaction_sync", "w1")
	ok2, _ := lm.AcquireLock(ctx, appID, "review_sync", "w2")

	if !ok1 || !ok2 {
		t.Error("Locks for different sync types should not interfere")
	}
}

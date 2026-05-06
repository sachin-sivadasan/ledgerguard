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

	// Release (using ForceReleaseLock for test simplicity)
	err = lm.ForceReleaseLock(ctx, appID, syncType)
	if err != nil {
		t.Fatalf("ForceReleaseLock failed: %v", err)
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

func TestReleaseLockIfOwner(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	appID := uuid.New()
	syncType := "transaction_sync"

	// Acquire as worker-1
	ok, _ := lm.AcquireLock(ctx, appID, syncType, "worker-1")
	if !ok {
		t.Fatal("Expected lock acquired")
	}

	// worker-2 cannot release
	released, err := lm.ReleaseLockIfOwner(ctx, appID, syncType, "worker-2")
	if err != nil {
		t.Fatalf("ReleaseLockIfOwner failed: %v", err)
	}
	if released {
		t.Error("Expected worker-2 NOT to release worker-1's lock")
	}

	// worker-1 can release
	released, err = lm.ReleaseLockIfOwner(ctx, appID, syncType, "worker-1")
	if err != nil {
		t.Fatalf("ReleaseLockIfOwner failed: %v", err)
	}
	if !released {
		t.Error("Expected worker-1 to release its own lock")
	}

	// Lock should be free now
	ok, _ = lm.AcquireLock(ctx, appID, syncType, "worker-3")
	if !ok {
		t.Error("Expected lock to be free after owner release")
	}
}

func TestExtendLockIfOwner(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	appID := uuid.New()
	syncType := "transaction_sync"

	// Acquire as worker-1
	ok, _ := lm.AcquireLock(ctx, appID, syncType, "worker-1")
	if !ok {
		t.Fatal("Expected lock acquired")
	}

	// worker-2 cannot extend
	extended, err := lm.ExtendLockIfOwner(ctx, appID, syncType, "worker-2")
	if err != nil {
		t.Fatalf("ExtendLockIfOwner failed: %v", err)
	}
	if extended {
		t.Error("Expected worker-2 NOT to extend worker-1's lock")
	}

	// worker-1 can extend
	extended, err = lm.ExtendLockIfOwner(ctx, appID, syncType, "worker-1")
	if err != nil {
		t.Fatalf("ExtendLockIfOwner failed: %v", err)
	}
	if !extended {
		t.Error("Expected worker-1 to extend its own lock")
	}
}

func TestStealLock(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	appID := uuid.New()
	syncType := "transaction_sync"

	// Acquire as worker-1
	ok, _ := lm.AcquireLock(ctx, appID, syncType, "worker-1")
	if !ok {
		t.Fatal("Expected lock acquired")
	}

	// Steal with wrong expected holder should fail
	stolen, err := lm.StealLock(ctx, appID, syncType, "worker-99", "worker-2")
	if err != nil {
		t.Fatalf("StealLock failed: %v", err)
	}
	if stolen {
		t.Error("Expected steal to fail with wrong expected holder")
	}

	// Steal with correct expected holder should succeed
	stolen, err = lm.StealLock(ctx, appID, syncType, "worker-1", "worker-2")
	if err != nil {
		t.Fatalf("StealLock failed: %v", err)
	}
	if !stolen {
		t.Error("Expected steal to succeed")
	}

	// Verify worker-2 now holds the lock
	holder, _ := lm.GetLockHolder(ctx, appID, syncType)
	if holder != "worker-2" {
		t.Errorf("Expected worker-2 as holder, got %q", holder)
	}
}

func TestGetLockHolder(t *testing.T) {
	_, client := setupMiniredis(t)
	ctx := context.Background()
	lm := NewLockManager(client)

	appID := uuid.New()
	syncType := "transaction_sync"

	// No holder initially
	holder, err := lm.GetLockHolder(ctx, appID, syncType)
	if err != nil {
		t.Fatalf("GetLockHolder failed: %v", err)
	}
	if holder != "" {
		t.Errorf("Expected empty holder, got %q", holder)
	}

	// After acquire
	lm.AcquireLock(ctx, appID, syncType, "worker-1")
	holder, _ = lm.GetLockHolder(ctx, appID, syncType)
	if holder != "worker-1" {
		t.Errorf("Expected worker-1, got %q", holder)
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

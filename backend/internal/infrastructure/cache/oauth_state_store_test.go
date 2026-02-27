package cache

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOAuthStateStore_StoreAndValidate(t *testing.T) {
	store := NewOAuthStateStore(5 * time.Minute)

	userID := uuid.New()
	state := "test-state-123"

	// Store state
	store.Store(state, userID)

	// Validate should return the user ID
	returnedUserID, valid := store.Validate(state)
	if !valid {
		t.Error("expected state to be valid")
	}

	if returnedUserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, returnedUserID)
	}
}

func TestOAuthStateStore_ValidateConsumesState(t *testing.T) {
	store := NewOAuthStateStore(5 * time.Minute)

	userID := uuid.New()
	state := "test-state-456"

	store.Store(state, userID)

	// First validation should succeed
	_, valid := store.Validate(state)
	if !valid {
		t.Error("expected first validation to succeed")
	}

	// Second validation should fail (state consumed)
	_, valid = store.Validate(state)
	if valid {
		t.Error("expected second validation to fail (state should be consumed)")
	}
}

func TestOAuthStateStore_InvalidState(t *testing.T) {
	store := NewOAuthStateStore(5 * time.Minute)

	// Validate non-existent state
	_, valid := store.Validate("non-existent-state")
	if valid {
		t.Error("expected validation to fail for non-existent state")
	}
}

func TestOAuthStateStore_ExpiredState(t *testing.T) {
	// Use very short TTL for testing
	store := NewOAuthStateStore(1 * time.Millisecond)

	userID := uuid.New()
	state := "test-state-789"

	store.Store(state, userID)

	// Wait for state to expire
	time.Sleep(10 * time.Millisecond)

	// Validate should fail for expired state
	_, valid := store.Validate(state)
	if valid {
		t.Error("expected validation to fail for expired state")
	}
}

func TestOAuthStateStore_MultipleStates(t *testing.T) {
	store := NewOAuthStateStore(5 * time.Minute)

	user1 := uuid.New()
	user2 := uuid.New()
	state1 := "state-1"
	state2 := "state-2"

	store.Store(state1, user1)
	store.Store(state2, user2)

	// Both states should be valid
	returnedUser1, valid1 := store.Validate(state1)
	if !valid1 || returnedUser1 != user1 {
		t.Error("state1 validation failed")
	}

	returnedUser2, valid2 := store.Validate(state2)
	if !valid2 || returnedUser2 != user2 {
		t.Error("state2 validation failed")
	}
}

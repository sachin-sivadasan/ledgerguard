package cache

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// OAuthStateData holds the user info associated with an OAuth state
type OAuthStateData struct {
	UserID    uuid.UUID
	CreatedAt time.Time
}

// OAuthStateStore stores OAuth state tokens for CSRF protection.
// In production, this should be replaced with Redis or similar.
type OAuthStateStore struct {
	mu     sync.RWMutex
	states map[string]OAuthStateData
	ttl    time.Duration
}

// NewOAuthStateStore creates a new in-memory state store with the given TTL.
func NewOAuthStateStore(ttl time.Duration) *OAuthStateStore {
	store := &OAuthStateStore{
		states: make(map[string]OAuthStateData),
		ttl:    ttl,
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Store saves a state token with associated user ID.
func (s *OAuthStateStore) Store(state string, userID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state] = OAuthStateData{
		UserID:    userID,
		CreatedAt: time.Now(),
	}
}

// Validate checks if a state token is valid and returns the associated user ID.
// The state is consumed (deleted) upon validation to prevent replay attacks.
func (s *OAuthStateStore) Validate(state string) (uuid.UUID, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.states[state]
	if !exists {
		return uuid.Nil, false
	}

	// Delete the state (one-time use)
	delete(s.states, state)

	// Check if expired
	if time.Since(data.CreatedAt) > s.ttl {
		return uuid.Nil, false
	}

	return data.UserID, true
}

// cleanup removes expired states periodically
func (s *OAuthStateStore) cleanup() {
	ticker := time.NewTicker(s.ttl)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for state, data := range s.states {
			if now.Sub(data.CreatedAt) > s.ttl {
				delete(s.states, state)
			}
		}
		s.mu.Unlock()
	}
}

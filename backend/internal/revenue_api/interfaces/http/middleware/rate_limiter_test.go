package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

// erroringStore always fails Increment — used to document fail-open behavior.
type erroringStore struct{}

func (erroringStore) Increment(context.Context, string, time.Duration) (int64, error) {
	return 0, errors.New("store down")
}

func runLimiter(rl *RateLimiter, key *ValidatedAPIKey) *httptest.ResponseRecorder {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/v1/subscriptions/x", nil)
	if key != nil {
		req = req.WithContext(SetAPIKeyContext(req.Context(), key))
	}
	rec := httptest.NewRecorder()
	rl.Middleware(next).ServeHTTP(rec, req)
	return rec
}

// TestRateLimiter_PerKeyWindow: the per-key limit is enforced and the Nth+1 request
// gets 429 with the rate-limit + Retry-After headers.
func TestRateLimiter_PerKeyWindow(t *testing.T) {
	rl := NewRateLimiter(NewInMemoryRateLimitStore(), 60, 60)
	key := &ValidatedAPIKey{ID: uuid.New(), RateLimitPerMinute: 2}

	if rec := runLimiter(rl, key); rec.Code != http.StatusOK {
		t.Fatalf("request 1: expected 200, got %d", rec.Code)
	}
	if rec := runLimiter(rl, key); rec.Code != http.StatusOK {
		t.Fatalf("request 2: expected 200, got %d", rec.Code)
	}
	rec := runLimiter(rl, key)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request 3: expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429")
	}
	if rec.Header().Get("X-RateLimit-Limit") != "2" {
		t.Errorf("expected X-RateLimit-Limit=2, got %q", rec.Header().Get("X-RateLimit-Limit"))
	}
}

// TestRateLimiter_NoKey_Skips: without an API key in context the limiter is a no-op.
func TestRateLimiter_NoKey_Skips(t *testing.T) {
	rl := NewRateLimiter(NewInMemoryRateLimitStore(), 1, 60)
	for i := 0; i < 5; i++ {
		if rec := runLimiter(rl, nil); rec.Code != http.StatusOK {
			t.Fatalf("no-key request %d: expected 200 (skip), got %d", i+1, rec.Code)
		}
	}
}

// TestRateLimiter_StoreError_FailsOpen documents the CURRENT fail-open behavior: on a
// store error the request is allowed. (S6 follow-up: a Redis-backed store may choose
// fail-closed; this test pins today's behavior so a change is deliberate.)
func TestRateLimiter_StoreError_FailsOpen(t *testing.T) {
	rl := NewRateLimiter(erroringStore{}, 1, 60)
	key := &ValidatedAPIKey{ID: uuid.New(), RateLimitPerMinute: 1}
	if rec := runLimiter(rl, key); rec.Code != http.StatusOK {
		t.Fatalf("store error: expected fail-open 200, got %d", rec.Code)
	}
}

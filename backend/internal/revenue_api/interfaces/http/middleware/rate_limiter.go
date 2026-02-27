package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RateLimitStore is an interface for rate limit storage
type RateLimitStore interface {
	// Increment increments the counter for a key and returns the new count
	// Returns the count after increment and the TTL remaining
	Increment(ctx context.Context, key string, window time.Duration) (count int64, err error)
}

// RateLimiter is middleware that enforces rate limits per API key
type RateLimiter struct {
	store         RateLimitStore
	defaultLimit  int
	windowSeconds int
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter(store RateLimitStore, defaultLimit int, windowSeconds int) *RateLimiter {
	return &RateLimiter{
		store:         store,
		defaultLimit:  defaultLimit,
		windowSeconds: windowSeconds,
	}
}

// Middleware returns the HTTP middleware handler
func (m *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the API key from context (must come after APIKeyAuth middleware)
		apiKey := APIKeyFromContext(r.Context())
		if apiKey == nil {
			// No API key in context, skip rate limiting
			next.ServeHTTP(w, r)
			return
		}

		// Determine the rate limit
		limit := apiKey.RateLimitPerMinute
		if limit <= 0 {
			limit = m.defaultLimit
		}

		// Create a unique key for this API key + current window
		window := time.Duration(m.windowSeconds) * time.Second
		windowKey := m.getWindowKey(apiKey.ID, window)

		// Increment the counter
		count, err := m.store.Increment(r.Context(), windowKey, window)
		if err != nil {
			// On error, allow the request but log it
			// In production, you might want to fail closed instead
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, limit-int(count))))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

		// Check if over limit
		if int(count) > limit {
			w.Header().Set("Retry-After", strconv.Itoa(m.windowSeconds))
			writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *RateLimiter) getWindowKey(keyID uuid.UUID, window time.Duration) string {
	// Create a key based on the API key ID and the current window
	windowStart := time.Now().Truncate(window).Unix()
	return "ratelimit:" + keyID.String() + ":" + strconv.FormatInt(windowStart, 10)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// InMemoryRateLimitStore is an in-memory implementation of RateLimitStore
// Suitable for single-instance deployments or development
type InMemoryRateLimitStore struct {
	mu       sync.RWMutex
	counters map[string]*rateLimitEntry
}

type rateLimitEntry struct {
	count     int64
	expiresAt time.Time
}

// NewInMemoryRateLimitStore creates a new in-memory rate limit store
func NewInMemoryRateLimitStore() *InMemoryRateLimitStore {
	store := &InMemoryRateLimitStore{
		counters: make(map[string]*rateLimitEntry),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Increment increments the counter for a key
func (s *InMemoryRateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.counters[key]

	if !exists || now.After(entry.expiresAt) {
		// Create new entry
		s.counters[key] = &rateLimitEntry{
			count:     1,
			expiresAt: now.Add(window),
		}
		return 1, nil
	}

	// Increment existing entry
	entry.count++
	return entry.count, nil
}

// cleanup removes expired entries periodically
func (s *InMemoryRateLimitStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.counters {
			if now.After(entry.expiresAt) {
				delete(s.counters, key)
			}
		}
		s.mu.Unlock()
	}
}

// RedisRateLimitStore is a Redis-backed implementation of RateLimitStore
// This is a placeholder - actual implementation would use go-redis
type RedisRateLimitStore struct {
	// client *redis.Client
	// In production, inject redis.Client here
}

// NewRedisRateLimitStore creates a new Redis rate limit store
// func NewRedisRateLimitStore(client *redis.Client) *RedisRateLimitStore {
// 	return &RedisRateLimitStore{client: client}
// }

// Increment increments the counter for a key using Redis INCR + EXPIRE
// func (s *RedisRateLimitStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
// 	pipe := s.client.Pipeline()
// 	incr := pipe.Incr(ctx, key)
// 	pipe.Expire(ctx, key, window)
// 	_, err := pipe.Exec(ctx)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return incr.Val(), nil
// }

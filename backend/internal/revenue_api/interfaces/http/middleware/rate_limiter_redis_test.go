package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// TestRedisRateLimitStore_Increment verifies the Redis-backed store increments
// per-key, keeps keys independent, and sets a TTL (so a key can't live forever).
func TestRedisRateLimitStore_Increment(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	store := NewRedisRateLimitStore(client)
	ctx := context.Background()

	// Counts up per key.
	for want := int64(1); want <= 3; want++ {
		got, err := store.Increment(ctx, "k1", time.Minute)
		if err != nil {
			t.Fatalf("increment: %v", err)
		}
		if got != want {
			t.Errorf("increment: got %d, want %d", got, want)
		}
	}

	// A different key is independent (shared store, separate counters).
	if got, err := store.Increment(ctx, "k2", time.Minute); err != nil || got != 1 {
		t.Errorf("separate key: got %d err %v, want 1 nil", got, err)
	}

	// EXPIRE was applied (bounded lifetime).
	if ttl := mr.TTL("k1"); ttl <= 0 {
		t.Errorf("expected positive TTL on k1, got %v", ttl)
	}
}

package cache

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryCache_SetGet(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	// Set a value
	err := cache.Set(ctx, key, value, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the value
	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(got) != string(value) {
		t.Errorf("Get returned %q, want %q", got, value)
	}
}

func TestInMemoryCache_GetNotFound(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	if _, ok := err.(ErrNotFound); !ok {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "expiring-key"
	value := []byte("expiring-value")

	// Set with short TTL
	err := cache.Set(ctx, key, value, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should exist immediately
	_, err = cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Get should succeed immediately: %v", err)
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, key)
	if _, ok := err.(ErrNotFound); !ok {
		t.Errorf("Expected ErrNotFound after expiration, got %v", err)
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "delete-key"
	value := []byte("delete-value")

	// Set a value
	cache.Set(ctx, key, value, time.Minute)

	// Delete it
	err := cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should not exist
	_, err = cache.Get(ctx, key)
	if _, ok := err.(ErrNotFound); !ok {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestInMemoryCache_Exists(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "exists-key"
	value := []byte("exists-value")

	// Should not exist
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Key should not exist initially")
	}

	// Set a value
	cache.Set(ctx, key, value, time.Minute)

	// Should exist
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Key should exist after set")
	}
}

func TestInMemoryCache_Incr(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "counter"

	// First increment
	val, err := cache.Incr(ctx, key)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	// Second increment
	val, err = cache.Incr(ctx, key)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected 2, got %d", val)
	}

	// Third increment
	val, err = cache.Incr(ctx, key)
	if err != nil {
		t.Fatalf("Incr failed: %v", err)
	}
	if val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}
}

func TestInMemoryCache_IncrWithExpiry(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "rate-limit"

	// First increment with short TTL
	val, err := cache.IncrWithExpiry(ctx, key, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("IncrWithExpiry failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	// Second increment
	val, err = cache.IncrWithExpiry(ctx, key, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("IncrWithExpiry failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected 2, got %d", val)
	}

	// Wait for expiration
	time.Sleep(30 * time.Millisecond)

	// Should reset after expiration
	val, err = cache.IncrWithExpiry(ctx, key, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("IncrWithExpiry failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1 after expiration, got %d", val)
	}
}

func TestGetSetJSON(t *testing.T) {
	cache := NewInMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "json-key"

	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := testData{Name: "test", Value: 42}

	// Set JSON
	err := SetJSON(ctx, cache, key, original, time.Minute)
	if err != nil {
		t.Fatalf("SetJSON failed: %v", err)
	}

	// Get JSON
	got, err := GetJSON[testData](ctx, cache, key)
	if err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	if got.Name != original.Name || got.Value != original.Value {
		t.Errorf("GetJSON returned %+v, want %+v", got, original)
	}
}

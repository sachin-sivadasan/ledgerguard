package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with a TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Incr increments a counter and returns the new value
	Incr(ctx context.Context, key string) (int64, error)

	// IncrWithExpiry increments a counter and sets expiry (for rate limiting)
	IncrWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error)

	// Close cleans up resources
	Close() error
}

// ErrNotFound is returned when a key is not found in the cache
type ErrNotFound struct {
	Key string
}

func (e ErrNotFound) Error() string {
	return "cache: key not found: " + e.Key
}

// entry represents a cached item with expiration
type entry struct {
	value     []byte
	expiresAt time.Time
	counter   int64 // Used for Incr operations
	isCounter bool
}

// InMemoryCache implements Cache using an in-memory map
type InMemoryCache struct {
	mu      sync.RWMutex
	items   map[string]entry
	stopCh  chan struct{}
	stopped bool
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() *InMemoryCache {
	c := &InMemoryCache{
		items:  make(map[string]entry),
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, exists := c.items[key]
	if !exists {
		return nil, ErrNotFound{Key: key}
	}

	// Check if expired
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		return nil, ErrNotFound{Key: key}
	}

	return e.value, nil
}

// Set stores a value in the cache with a TTL
func (c *InMemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = entry{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Exists checks if a key exists in the cache
func (c *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, exists := c.items[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Incr increments a counter and returns the new value
func (c *InMemoryCache) Incr(ctx context.Context, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, exists := c.items[key]
	if !exists || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		c.items[key] = entry{
			counter:   1,
			isCounter: true,
		}
		return 1, nil
	}

	e.counter++
	c.items[key] = e
	return e.counter, nil
}

// IncrWithExpiry increments a counter and sets expiry
func (c *InMemoryCache) IncrWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, exists := c.items[key]
	now := time.Now()

	if !exists || (!e.expiresAt.IsZero() && now.After(e.expiresAt)) {
		// Create new counter with expiry
		c.items[key] = entry{
			counter:   1,
			isCounter: true,
			expiresAt: now.Add(ttl),
		}
		return 1, nil
	}

	e.counter++
	c.items[key] = e
	return e.counter, nil
}

// Close stops the cleanup goroutine
func (c *InMemoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.stopped {
		close(c.stopCh)
		c.stopped = true
	}
	return nil
}

// cleanup removes expired entries periodically
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, e := range c.items {
				if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

// Helper functions for common caching patterns

// GetJSON retrieves and unmarshals a JSON value from the cache
func GetJSON[T any](ctx context.Context, cache Cache, key string) (T, error) {
	var result T
	data, err := cache.Get(ctx, key)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	return result, nil
}

// SetJSON marshals and stores a JSON value in the cache
func SetJSON[T any](ctx context.Context, cache Cache, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return cache.Set(ctx, key, data, ttl)
}

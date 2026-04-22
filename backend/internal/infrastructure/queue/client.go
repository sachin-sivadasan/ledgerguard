package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/config"
)

// NewRedisClient creates a Redis client from configuration.
// Returns nil if Redis is not configured.
func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis address not configured")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", cfg.Addr, err)
	}

	log.Printf("Connected to Redis at %s", cfg.Addr)
	return client, nil
}

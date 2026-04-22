package queue

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// Progress represents the current progress of a sync job
type Progress struct {
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Message   string `json:"message"`
}

func progressRedisKey(jobID uuid.UUID) string {
	return fmt.Sprintf("lg:sync:progress:%s", jobID.String())
}

// ProgressTracker handles dual-write progress tracking to Redis (fast) and DB (durable)
type ProgressTracker struct {
	client        *redis.Client
	syncJobRepo   repository.SyncJobRepository
	redisInterval time.Duration
	dbInterval    time.Duration

	mu             sync.Mutex
	lastRedisWrite map[uuid.UUID]time.Time
	lastDBWrite    map[uuid.UUID]time.Time
}

// NewProgressTracker creates a new progress tracker with specified intervals
func NewProgressTracker(client *redis.Client, repo repository.SyncJobRepository, redisInterval, dbInterval time.Duration) *ProgressTracker {
	return &ProgressTracker{
		client:         client,
		syncJobRepo:    repo,
		redisInterval:  redisInterval,
		dbInterval:     dbInterval,
		lastRedisWrite: make(map[uuid.UUID]time.Time),
		lastDBWrite:    make(map[uuid.UUID]time.Time),
	}
}

// Update writes progress to Redis (throttled) and DB (throttled at longer interval)
func (pt *ProgressTracker) Update(ctx context.Context, jobID uuid.UUID, p Progress) {
	now := time.Now()

	pt.mu.Lock()
	lastRedis := pt.lastRedisWrite[jobID]
	lastDB := pt.lastDBWrite[jobID]
	pt.mu.Unlock()

	// Write to Redis if enough time has passed
	if now.Sub(lastRedis) >= pt.redisInterval {
		key := progressRedisKey(jobID)
		_ = pt.client.HSet(ctx, key, map[string]interface{}{
			"total":     p.Total,
			"completed": p.Completed,
			"message":   p.Message,
		}).Err()
		_ = pt.client.Expire(ctx, key, 2*time.Hour).Err()

		pt.mu.Lock()
		pt.lastRedisWrite[jobID] = now
		pt.mu.Unlock()
	}

	// Write to DB if enough time has passed
	if now.Sub(lastDB) >= pt.dbInterval {
		_ = pt.syncJobRepo.UpdateProgress(ctx, jobID, p.Total, p.Completed)

		pt.mu.Lock()
		pt.lastDBWrite[jobID] = now
		pt.mu.Unlock()
	}
}

// ForceUpdate writes progress immediately (for milestones like completion)
func (pt *ProgressTracker) ForceUpdate(ctx context.Context, jobID uuid.UUID, p Progress) {
	key := progressRedisKey(jobID)
	_ = pt.client.HSet(ctx, key, map[string]interface{}{
		"total":     p.Total,
		"completed": p.Completed,
		"message":   p.Message,
	}).Err()
	_ = pt.client.Expire(ctx, key, 2*time.Hour).Err()

	_ = pt.syncJobRepo.UpdateProgress(ctx, jobID, p.Total, p.Completed)

	pt.mu.Lock()
	now := time.Now()
	pt.lastRedisWrite[jobID] = now
	pt.lastDBWrite[jobID] = now
	pt.mu.Unlock()
}

// GetProgress reads progress from Redis. Returns nil if not found.
func (pt *ProgressTracker) GetProgress(ctx context.Context, jobID uuid.UUID) (*Progress, error) {
	key := progressRedisKey(jobID)
	result, err := pt.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil // Not found in Redis
	}

	total, _ := strconv.Atoi(result["total"])
	completed, _ := strconv.Atoi(result["completed"])

	return &Progress{
		Total:     total,
		Completed: completed,
		Message:   result["message"],
	}, nil
}

// Cleanup removes progress data from Redis and internal maps
func (pt *ProgressTracker) Cleanup(ctx context.Context, jobID uuid.UUID) {
	_ = pt.client.Del(ctx, progressRedisKey(jobID)).Err()

	pt.mu.Lock()
	delete(pt.lastRedisWrite, jobID)
	delete(pt.lastDBWrite, jobID)
	pt.mu.Unlock()
}

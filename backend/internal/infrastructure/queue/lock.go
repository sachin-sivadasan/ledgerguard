package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	lockTTL               = 2 * time.Hour
	heartbeatTTL          = 20 * time.Minute
	heartbeatInterval     = 10 * time.Minute
	lockExtensionInterval = 1 * time.Hour
)

// LockManager manages distributed locks and heartbeats for sync jobs
type LockManager struct {
	client *redis.Client
}

// NewLockManager creates a new lock manager
func NewLockManager(client *redis.Client) *LockManager {
	return &LockManager{client: client}
}

func lockKey(appID uuid.UUID, syncType string) string {
	return fmt.Sprintf("lg:sync:lock:%s:%s", appID.String(), syncType)
}

func heartbeatKey(jobID uuid.UUID) string {
	return fmt.Sprintf("lg:sync:heartbeat:%s", jobID.String())
}

func cancelKey(jobID uuid.UUID) string {
	return fmt.Sprintf("lg:sync:cancel:%s", jobID.String())
}

// AcquireLock attempts to acquire a distributed lock via SETNX
func (lm *LockManager) AcquireLock(ctx context.Context, appID uuid.UUID, syncType, workerID string) (bool, error) {
	key := lockKey(appID, syncType)
	ok, err := lm.client.SetNX(ctx, key, workerID, lockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock %s: %w", key, err)
	}
	return ok, nil
}

// ReleaseLock releases a distributed lock
func (lm *LockManager) ReleaseLock(ctx context.Context, appID uuid.UUID, syncType string) error {
	key := lockKey(appID, syncType)
	return lm.client.Del(ctx, key).Err()
}

// ExtendLock extends the lock TTL
func (lm *LockManager) ExtendLock(ctx context.Context, appID uuid.UUID, syncType string) error {
	key := lockKey(appID, syncType)
	return lm.client.Expire(ctx, key, lockTTL).Err()
}

// Heartbeat writes a heartbeat for a running job
func (lm *LockManager) Heartbeat(ctx context.Context, jobID uuid.UUID) error {
	key := heartbeatKey(jobID)
	return lm.client.Set(ctx, key, time.Now().UTC().Format(time.RFC3339), heartbeatTTL).Err()
}

// HasHeartbeat checks if a job has a recent heartbeat
func (lm *LockManager) HasHeartbeat(ctx context.Context, jobID uuid.UUID) (bool, error) {
	key := heartbeatKey(jobID)
	exists, err := lm.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// DeleteHeartbeat removes a heartbeat key
func (lm *LockManager) DeleteHeartbeat(ctx context.Context, jobID uuid.UUID) error {
	return lm.client.Del(ctx, heartbeatKey(jobID)).Err()
}

// RequestCancellation sets a cancellation flag for a job
func (lm *LockManager) RequestCancellation(ctx context.Context, jobID uuid.UUID) error {
	key := cancelKey(jobID)
	return lm.client.Set(ctx, key, "1", lockTTL).Err()
}

// IsCancelled checks if cancellation was requested for a job
func (lm *LockManager) IsCancelled(ctx context.Context, jobID uuid.UUID) (bool, error) {
	key := cancelKey(jobID)
	exists, err := lm.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// CleanupCancellation removes the cancellation flag
func (lm *LockManager) CleanupCancellation(ctx context.Context, jobID uuid.UUID) error {
	return lm.client.Del(ctx, cancelKey(jobID)).Err()
}

// HeartbeatInterval returns the interval between heartbeats
func (lm *LockManager) HeartbeatInterval() time.Duration {
	return heartbeatInterval
}

// LockExtensionInterval returns the interval between lock extensions
func (lm *LockManager) LockExtensionInterval() time.Duration {
	return lockExtensionInterval
}

package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Queue key constants
const (
	RegularQueueKey  = "lg:sync:queue"
	FullSyncQueueKey = "lg:sync:queue:full"
)

// SyncJobPayload is the JSON payload pushed into Redis queues
type SyncJobPayload struct {
	JobID            uuid.UUID `json:"job_id"`
	AppID            uuid.UUID `json:"app_id"`
	UserID           uuid.UUID `json:"user_id"`
	PartnerAccountID uuid.UUID `json:"partner_account_id"`
	JobType          string    `json:"job_type"`
	ParentJobID      *uuid.UUID `json:"parent_job_id,omitempty"`
	Priority         int       `json:"priority"`
	EntityType       string    `json:"entity_type,omitempty"`
	LookbackDays     int       `json:"lookback_days,omitempty"` // 0 = default window
	EnqueuedAt       time.Time `json:"enqueued_at"`
}

// QueueKeyForJobType returns the appropriate queue key for the given job type
func QueueKeyForJobType(jobType string) string {
	if jobType == "full_sync" {
		return FullSyncQueueKey
	}
	return RegularQueueKey
}

// Enqueue pushes a sync job payload to the appropriate Redis queue via LPUSH
func Enqueue(ctx context.Context, client *redis.Client, payload *SyncJobPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	queueKey := QueueKeyForJobType(payload.JobType)
	if err := client.LPush(ctx, queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue to %s: %w", queueKey, err)
	}

	return nil
}

// Dequeue pops a sync job payload from a Redis queue via BRPOP with timeout
func Dequeue(ctx context.Context, client *redis.Client, queueKey string, timeout time.Duration) (*SyncJobPayload, error) {
	result, err := client.BRPop(ctx, timeout, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout, no item
		}
		return nil, fmt.Errorf("failed to dequeue from %s: %w", queueKey, err)
	}

	if len(result) < 2 {
		return nil, nil
	}

	var payload SyncJobPayload
	if err := json.Unmarshal([]byte(result[1]), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &payload, nil
}

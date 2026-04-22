package entity

import (
	"time"

	"github.com/google/uuid"
)

// SyncJobStatus represents the lifecycle status of a sync job
type SyncJobStatus string

const (
	SyncJobStatusPending        SyncJobStatus = "pending"
	SyncJobStatusProcessing     SyncJobStatus = "processing"
	SyncJobStatusCompleted      SyncJobStatus = "completed"
	SyncJobStatusFailed         SyncJobStatus = "failed"
	SyncJobStatusCancelled      SyncJobStatus = "cancelled"
	SyncJobStatusPartialFailure SyncJobStatus = "partial_failure"
)

// SyncJob types
const (
	SyncJobTypeFullSync        = "full_sync"
	SyncJobTypeTransactionSync = "transaction_sync"
	SyncJobTypeSnapshotSync    = "snapshot_sync"
	SyncJobTypeEventSync       = "event_sync"
	SyncJobTypeStatusSync      = "status_sync"
	SyncJobTypeStoreSync       = "store_sync"
	SyncJobTypeReviewSync      = "review_sync"
)

// SyncJob represents an async sync job tracked in the database
type SyncJob struct {
	ID               uuid.UUID
	AppID            uuid.UUID
	UserID           uuid.UUID
	PartnerAccountID uuid.UUID
	JobType          string
	ParentJobID      *uuid.UUID
	Status           SyncJobStatus
	Priority         int // 0=normal, 1=high
	TotalItems       int
	CompletedItems   int
	EntityType       string
	ErrorMessage     string
	WorkerID         string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewSyncJob creates a new sync job in pending status
func NewSyncJob(appID, userID, partnerAccountID uuid.UUID, jobType string, priority int) *SyncJob {
	now := time.Now().UTC()
	return &SyncJob{
		ID:               uuid.New(),
		AppID:            appID,
		UserID:           userID,
		PartnerAccountID: partnerAccountID,
		JobType:          jobType,
		Status:           SyncJobStatusPending,
		Priority:         priority,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewChildSyncJob creates a child sync job linked to a parent
func NewChildSyncJob(parentJob *SyncJob, jobType, entityType string) *SyncJob {
	job := NewSyncJob(parentJob.AppID, parentJob.UserID, parentJob.PartnerAccountID, jobType, parentJob.Priority)
	job.ParentJobID = &parentJob.ID
	job.EntityType = entityType
	return job
}

// IsTerminal returns true if the job is in a terminal state
func (j *SyncJob) IsTerminal() bool {
	return j.Status == SyncJobStatusCompleted ||
		j.Status == SyncJobStatusFailed ||
		j.Status == SyncJobStatusCancelled ||
		j.Status == SyncJobStatusPartialFailure
}

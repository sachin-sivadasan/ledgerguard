package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// SyncJobRepository defines the interface for sync job persistence
type SyncJobRepository interface {
	// Create persists a new sync job
	Create(ctx context.Context, job *entity.SyncJob) error

	// FindByID returns a sync job by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.SyncJob, error)

	// FindByStatus returns all sync jobs with the given status
	FindByStatus(ctx context.Context, status entity.SyncJobStatus) ([]*entity.SyncJob, error)

	// FindActiveByAppIDAndType returns non-terminal jobs for an app+type combo
	FindActiveByAppIDAndType(ctx context.Context, appID uuid.UUID, jobType string) (*entity.SyncJob, error)

	// FindByParentJobID returns all child jobs of a parent
	FindByParentJobID(ctx context.Context, parentJobID uuid.UUID) ([]*entity.SyncJob, error)

	// ListByAppID returns paginated sync jobs for an app
	ListByAppID(ctx context.Context, appID uuid.UUID, status string, jobType string, limit, offset int) ([]*entity.SyncJob, int, error)

	// UpdateStatus sets the job status and updated_at
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SyncJobStatus) error

	// UpdateProgress updates total_items and completed_items
	UpdateProgress(ctx context.Context, id uuid.UUID, totalItems, completedItems int) error

	// MarkStarted sets status=processing, started_at, worker_id
	MarkStarted(ctx context.Context, id uuid.UUID, workerID string) error

	// MarkCompleted sets status=completed, completed_at
	MarkCompleted(ctx context.Context, id uuid.UUID) error

	// MarkFailed sets status=failed, error_message, completed_at
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error

	// MarkPendingIfProcessing atomically sets status=pending only if currently processing
	MarkPendingIfProcessing(ctx context.Context, id uuid.UUID) error
}

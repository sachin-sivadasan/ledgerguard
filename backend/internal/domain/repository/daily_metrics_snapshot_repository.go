package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// DailyMetricsSnapshotRepository defines operations for daily metrics snapshots
type DailyMetricsSnapshotRepository interface {
	// Upsert creates or updates a snapshot for the given app and date
	// Uses ON CONFLICT (app_id, date) DO UPDATE for idempotency
	Upsert(ctx context.Context, snapshot *entity.DailyMetricsSnapshot) error

	// FindByAppIDAndDate retrieves a snapshot for a specific app and date
	FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyMetricsSnapshot, error)

	// FindByAppIDRange retrieves snapshots for an app within a date range
	FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyMetricsSnapshot, error)

	// FindLatestByAppID retrieves the most recent snapshot for an app
	FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyMetricsSnapshot, error)
}

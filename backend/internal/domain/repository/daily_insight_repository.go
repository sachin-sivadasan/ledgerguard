package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// DailyInsightRepository defines operations for daily AI insights
type DailyInsightRepository interface {
	// Upsert creates or updates an insight for the given app and date
	// Uses ON CONFLICT (app_id, date) DO UPDATE for idempotency
	Upsert(ctx context.Context, insight *entity.DailyInsight) error

	// FindByAppIDAndDate retrieves an insight for a specific app and date
	FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyInsight, error)

	// FindByAppIDRange retrieves insights for an app within a date range
	FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyInsight, error)

	// FindLatestByAppID retrieves the most recent insight for an app
	FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyInsight, error)
}

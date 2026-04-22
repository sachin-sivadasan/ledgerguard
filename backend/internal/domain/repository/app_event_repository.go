package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// AppEventRepository defines the interface for app event persistence
type AppEventRepository interface {
	// UpsertBatch inserts or updates app events (idempotent by unique constraint)
	UpsertBatch(ctx context.Context, events []*entity.AppEvent) error

	// FindByAppAndShop returns events for a specific app+shop combo
	FindByAppAndShop(ctx context.Context, appID uuid.UUID, shopGID string) ([]*entity.AppEvent, error)

	// FindByAppID returns all events for an app
	FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppEvent, error)
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// EventFilters holds pagination and filter params for events.
type EventFilters struct {
	EventType   string    // optional filter
	StoreDomain string    // optional: used by handler to resolve GIDs before query
	ShopGIDs    []string  // optional: filter by shopify_shop_gid IN (...)
	Since       time.Time // optional: only events after this time
	Page        int
	PageSize    int
}

// EventPage is a paginated result set of events.
type EventPage struct {
	Events     []*entity.AppEvent
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// AppEventRepository defines the interface for app event persistence
type AppEventRepository interface {
	// UpsertBatch inserts or updates app events (idempotent by unique constraint)
	UpsertBatch(ctx context.Context, events []*entity.AppEvent) error

	// FindByAppAndShop returns events for a specific app+shop combo
	FindByAppAndShop(ctx context.Context, appID uuid.UUID, shopGID string) ([]*entity.AppEvent, error)

	// FindByAppID returns all events for an app
	FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppEvent, error)

	// FindByAppIDPaginated returns a paginated set of events for an app.
	FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters EventFilters) (*EventPage, error)
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// AppRepository defines the interface for app persistence.
//
// Error handling contract:
//   - Single-item Find methods (FindByID, FindByPartnerAppID) return a sentinel error
//     (e.g., ErrAppNotFound) when no matching record exists.
//   - List methods (FindByPartnerAccountID, FindAllByPartnerAppID) return an empty slice
//     when no records match; they do NOT return an error for "not found".
//   - All methods may return other errors for database/network failures.
type AppRepository interface {
	// Create persists a new app. Returns error if app with same partner_app_id exists.
	Create(ctx context.Context, app *entity.App) error

	// FindByID returns the app with the given ID.
	// Returns ErrAppNotFound if no app exists with that ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error)

	// FindByPartnerAccountID returns all apps belonging to a partner account.
	// Returns empty slice if no apps exist (not an error).
	FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error)

	// FindByPartnerAppID returns the app matching both partner account and Shopify app GID.
	// Returns ErrAppNotFound if no matching app exists.
	FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error)

	// FindAllByPartnerAppID finds all apps matching a partner app ID across all accounts.
	// Used for webhook processing where we only have the Shopify app GID.
	// Returns empty slice if no apps match (not an error).
	FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error)

	// Update updates an existing app. Returns error if app doesn't exist.
	// NOTE: this is a FULL-ROW overwrite — a caller holding a stale app object (e.g.
	// a long-running sync) will clobber fields changed concurrently by others. For a
	// single owned field, prefer a targeted updater (e.g. UpdateInstallCount).
	Update(ctx context.Context, app *entity.App) error

	// UpdateInstallCount sets ONLY install_count (+ updated_at) for an app, without
	// touching user-owned fields like app_store_slug or revenue_share_tier. Used by the
	// sync pipeline so a periodic install-count refresh can't wipe a slug/tier a user
	// set mid-sync (the read-modify-write race that full-row Update is prone to).
	UpdateInstallCount(ctx context.Context, appID uuid.UUID, count int) error

	// Delete removes an app by ID. Returns error if app doesn't exist.
	Delete(ctx context.Context, id uuid.UUID) error
}

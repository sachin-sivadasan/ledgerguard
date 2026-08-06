package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// MobileLinksRepository persists an app's mobile store identifiers.
type MobileLinksRepository interface {
	// FindByAppID returns the links for an app, or a zero-value MobileLinks (not an error)
	// when none are set yet.
	FindByAppID(ctx context.Context, appID uuid.UUID) (*entity.MobileLinks, error)
	// Upsert stores the links for an app (replacing any existing row).
	Upsert(ctx context.Context, links *entity.MobileLinks) error
}

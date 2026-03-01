package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type AppRepository interface {
	Create(ctx context.Context, app *entity.App) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error)
	FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error)
	FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error)
	// FindAllByPartnerAppID finds all apps matching a partner app ID across all accounts
	// Used for webhook processing where we only have the Shopify app GID
	FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error)
	Update(ctx context.Context, app *entity.App) error
	Delete(ctx context.Context, id uuid.UUID) error
}

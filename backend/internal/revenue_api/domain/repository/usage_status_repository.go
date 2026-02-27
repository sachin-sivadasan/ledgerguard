package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

// UsageStatusRepository defines the interface for usage status persistence (CQRS read model)
type UsageStatusRepository interface {
	// Upsert creates or updates a usage status
	Upsert(ctx context.Context, status *entity.UsageStatus) error

	// UpsertBatch creates or updates multiple usage statuses
	UpsertBatch(ctx context.Context, statuses []*entity.UsageStatus) error

	// GetByShopifyGID retrieves a usage status by Shopify GID
	GetByShopifyGID(ctx context.Context, shopifyGID string) (*entity.UsageStatus, error)

	// GetByShopifyGIDs retrieves multiple usage statuses by Shopify GIDs
	GetByShopifyGIDs(ctx context.Context, shopifyGIDs []string) ([]*entity.UsageStatus, error)

	// GetBySubscriptionID retrieves all usage statuses for a subscription
	GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.UsageStatus, error)

	// GetBySubscriptionShopifyGID retrieves all usage statuses for a subscription by Shopify GID
	GetBySubscriptionShopifyGID(ctx context.Context, subscriptionShopifyGID string) ([]*entity.UsageStatus, error)

	// GetUnbilledBySubscriptionID retrieves unbilled usage statuses for a subscription
	GetUnbilledBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.UsageStatus, error)

	// DeleteBySubscriptionID deletes all usage statuses for a subscription
	DeleteBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) error
}

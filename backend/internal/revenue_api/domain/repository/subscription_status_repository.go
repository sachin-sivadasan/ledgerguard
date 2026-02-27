package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

// SubscriptionStatusRepository defines the interface for subscription status persistence (CQRS read model)
type SubscriptionStatusRepository interface {
	// Upsert creates or updates a subscription status
	Upsert(ctx context.Context, status *entity.SubscriptionStatus) error

	// UpsertBatch creates or updates multiple subscription statuses
	UpsertBatch(ctx context.Context, statuses []*entity.SubscriptionStatus) error

	// GetByShopifyGID retrieves a subscription status by Shopify GID
	GetByShopifyGID(ctx context.Context, shopifyGID string) (*entity.SubscriptionStatus, error)

	// GetByShopifyGIDs retrieves multiple subscription statuses by Shopify GIDs
	GetByShopifyGIDs(ctx context.Context, shopifyGIDs []string) ([]*entity.SubscriptionStatus, error)

	// GetByDomain retrieves a subscription status by myshopify domain
	GetByDomain(ctx context.Context, appID uuid.UUID, domain string) (*entity.SubscriptionStatus, error)

	// GetByDomains retrieves multiple subscription statuses by domains
	GetByDomains(ctx context.Context, appID uuid.UUID, domains []string) ([]*entity.SubscriptionStatus, error)

	// GetByAppID retrieves all subscription statuses for an app
	GetByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.SubscriptionStatus, error)

	// GetByAppIDAndRiskState retrieves subscription statuses filtered by risk state
	GetByAppIDAndRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.SubscriptionStatus, error)

	// DeleteByAppID deletes all subscription statuses for an app (for rebuild)
	DeleteByAppID(ctx context.Context, appID uuid.UUID) error
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// SubscriptionFilters contains filter criteria for subscription queries
type SubscriptionFilters struct {
	RiskStates      []valueobject.RiskState
	PriceMinCents   *int64
	PriceMaxCents   *int64
	BillingInterval *valueobject.BillingInterval
	SearchTerm      string
	SortBy          string // "risk_state", "base_price_cents", "shop_name"
	SortOrder       string // "asc" or "desc"
	Page            int
	PageSize        int
}

// SubscriptionPage contains paginated subscription results
type SubscriptionPage struct {
	Subscriptions []*entity.Subscription
	Total         int
	Page          int
	PageSize      int
	TotalPages    int
}

// SubscriptionSummary contains aggregate subscription statistics
type SubscriptionSummary struct {
	ActiveCount   int
	AtRiskCount   int
	ChurnedCount  int
	AvgPriceCents int64
	TotalCount    int
}

// PricePoint represents a distinct price with its count
type PricePoint struct {
	PriceCents int64
	Count      int
}

// PriceStats contains price statistics and distinct prices for filtering
type PriceStats struct {
	MinCents int64
	MaxCents int64
	AvgCents int64
	Prices   []PricePoint // Distinct prices with counts, sorted by price
}

type SubscriptionRepository interface {
	Upsert(ctx context.Context, subscription *entity.Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error)
	FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Subscription, error)
	FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, myshopifyDomain string) (*entity.Subscription, error)
	FindByRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.Subscription, error)
	DeleteByAppID(ctx context.Context, appID uuid.UUID) error

	// Soft delete operations (preserves historical data)
	SoftDeleteByAppID(ctx context.Context, appID uuid.UUID) error
	FindDeletedByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error)
	RestoreByID(ctx context.Context, id uuid.UUID) error

	// Advanced querying
	FindWithFilters(ctx context.Context, appID uuid.UUID, filters SubscriptionFilters) (*SubscriptionPage, error)
	GetSummary(ctx context.Context, appID uuid.UUID) (*SubscriptionSummary, error)
	GetPriceStats(ctx context.Context, appID uuid.UUID) (*PriceStats, error)
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// EarningsSummary contains aggregated earnings by status
type EarningsSummary struct {
	PendingCents   int64
	AvailableCents int64
	PaidOutCents   int64
}

// EarningsByDate represents earnings for a specific date
type EarningsByDate struct {
	Date       time.Time
	AmountCents int64
}

type TransactionRepository interface {
	// Upsert inserts or updates a transaction (idempotent by shopify_gid)
	Upsert(ctx context.Context, tx *entity.Transaction) error

	// UpsertBatch inserts or updates multiple transactions
	UpsertBatch(ctx context.Context, txs []*entity.Transaction) error

	// FindByAppID returns transactions for an app within date range
	FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error)

	// FindByShopifyGID finds a transaction by its Shopify GID
	FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Transaction, error)

	// CountByAppID returns total transaction count for an app
	CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error)

	// GetEarningsSummary returns aggregated earnings by status
	GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*EarningsSummary, error)

	// GetPendingByAvailableDate returns pending earnings grouped by available_date
	GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]EarningsByDate, error)

	// GetUpcomingAvailability returns earnings becoming available in the next N days
	GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]EarningsByDate, error)
}

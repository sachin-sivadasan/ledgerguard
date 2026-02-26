package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

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
}

package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

var ErrUsageStatusNotFound = errors.New("usage status not found")

// PostgresUsageStatusRepository implements UsageStatusRepository using PostgreSQL
type PostgresUsageStatusRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUsageStatusRepository creates a new PostgresUsageStatusRepository
func NewPostgresUsageStatusRepository(pool *pgxpool.Pool) *PostgresUsageStatusRepository {
	return &PostgresUsageStatusRepository{pool: pool}
}

// Upsert creates or updates a usage status
func (r *PostgresUsageStatusRepository) Upsert(ctx context.Context, status *entity.UsageStatus) error {
	query := `
		INSERT INTO api_usage_status (
			id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			billed = EXCLUDED.billed,
			billing_date = EXCLUDED.billing_date,
			amount_cents = EXCLUDED.amount_cents,
			description = EXCLUDED.description,
			last_synced_at = EXCLUDED.last_synced_at
	`

	_, err := r.pool.Exec(ctx, query,
		status.ID,
		status.ShopifyGID,
		status.SubscriptionShopifyGID,
		status.SubscriptionID,
		status.Billed,
		status.BillingDate,
		status.AmountCents,
		status.Description,
		status.LastSyncedAt,
	)

	return err
}

// UpsertBatch creates or updates multiple usage statuses
func (r *PostgresUsageStatusRepository) UpsertBatch(ctx context.Context, statuses []*entity.UsageStatus) error {
	if len(statuses) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO api_usage_status (
			id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			billed = EXCLUDED.billed,
			billing_date = EXCLUDED.billing_date,
			amount_cents = EXCLUDED.amount_cents,
			description = EXCLUDED.description,
			last_synced_at = EXCLUDED.last_synced_at
	`

	for _, status := range statuses {
		batch.Queue(query,
			status.ID,
			status.ShopifyGID,
			status.SubscriptionShopifyGID,
			status.SubscriptionID,
			status.Billed,
			status.BillingDate,
			status.AmountCents,
			status.Description,
			status.LastSyncedAt,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range statuses {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

// GetByShopifyGID retrieves a usage status by Shopify GID
func (r *PostgresUsageStatusRepository) GetByShopifyGID(ctx context.Context, shopifyGID string) (*entity.UsageStatus, error) {
	query := `
		SELECT id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		FROM api_usage_status
		WHERE shopify_gid = $1
	`

	return r.scanStatus(r.pool.QueryRow(ctx, query, shopifyGID))
}

// GetByShopifyGIDs retrieves multiple usage statuses by Shopify GIDs
func (r *PostgresUsageStatusRepository) GetByShopifyGIDs(ctx context.Context, shopifyGIDs []string) ([]*entity.UsageStatus, error) {
	if len(shopifyGIDs) == 0 {
		return []*entity.UsageStatus{}, nil
	}

	query := `
		SELECT id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		FROM api_usage_status
		WHERE shopify_gid = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, shopifyGIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// GetBySubscriptionID retrieves all usage statuses for a subscription
func (r *PostgresUsageStatusRepository) GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.UsageStatus, error) {
	query := `
		SELECT id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		FROM api_usage_status
		WHERE subscription_id = $1
		ORDER BY last_synced_at DESC
	`

	rows, err := r.pool.Query(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// GetBySubscriptionShopifyGID retrieves all usage statuses by subscription Shopify GID
func (r *PostgresUsageStatusRepository) GetBySubscriptionShopifyGID(ctx context.Context, subscriptionShopifyGID string) ([]*entity.UsageStatus, error) {
	query := `
		SELECT id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		FROM api_usage_status
		WHERE subscription_shopify_gid = $1
		ORDER BY last_synced_at DESC
	`

	rows, err := r.pool.Query(ctx, query, subscriptionShopifyGID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// GetUnbilledBySubscriptionID retrieves unbilled usage statuses for a subscription
func (r *PostgresUsageStatusRepository) GetUnbilledBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.UsageStatus, error) {
	query := `
		SELECT id, shopify_gid, subscription_shopify_gid, subscription_id,
			billed, billing_date, amount_cents, description, last_synced_at
		FROM api_usage_status
		WHERE subscription_id = $1 AND billed = false
		ORDER BY last_synced_at DESC
	`

	rows, err := r.pool.Query(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// DeleteBySubscriptionID deletes all usage statuses for a subscription
func (r *PostgresUsageStatusRepository) DeleteBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) error {
	query := `DELETE FROM api_usage_status WHERE subscription_id = $1`
	_, err := r.pool.Exec(ctx, query, subscriptionID)
	return err
}

func (r *PostgresUsageStatusRepository) scanStatus(row pgx.Row) (*entity.UsageStatus, error) {
	var status entity.UsageStatus
	var description *string

	err := row.Scan(
		&status.ID,
		&status.ShopifyGID,
		&status.SubscriptionShopifyGID,
		&status.SubscriptionID,
		&status.Billed,
		&status.BillingDate,
		&status.AmountCents,
		&description,
		&status.LastSyncedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUsageStatusNotFound
		}
		return nil, err
	}

	if description != nil {
		status.Description = *description
	}

	return &status, nil
}

func (r *PostgresUsageStatusRepository) scanStatuses(rows pgx.Rows) ([]*entity.UsageStatus, error) {
	var statuses []*entity.UsageStatus

	for rows.Next() {
		var status entity.UsageStatus
		var description *string

		err := rows.Scan(
			&status.ID,
			&status.ShopifyGID,
			&status.SubscriptionShopifyGID,
			&status.SubscriptionID,
			&status.Billed,
			&status.BillingDate,
			&status.AmountCents,
			&description,
			&status.LastSyncedAt,
		)
		if err != nil {
			return nil, err
		}

		if description != nil {
			status.Description = *description
		}
		statuses = append(statuses, &status)
	}

	return statuses, rows.Err()
}

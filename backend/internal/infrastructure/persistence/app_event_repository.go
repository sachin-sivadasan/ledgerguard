package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresAppEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAppEventRepository(pool *pgxpool.Pool) *PostgresAppEventRepository {
	return &PostgresAppEventRepository{pool: pool}
}

func (r *PostgresAppEventRepository) UpsertBatch(ctx context.Context, events []*entity.AppEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO app_events (id, app_id, shopify_shop_gid, event_type, occurred_at, raw_data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (app_id, shopify_shop_gid, event_type, occurred_at) DO NOTHING
	`

	for _, ev := range events {
		batch.Queue(query, ev.ID, ev.AppID, ev.ShopifyShopGID, ev.EventType, ev.OccurredAt, ev.RawData, ev.CreatedAt)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range events {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresAppEventRepository) FindByAppAndShop(ctx context.Context, appID uuid.UUID, shopGID string) ([]*entity.AppEvent, error) {
	query := `
		SELECT id, app_id, shopify_shop_gid, event_type, occurred_at, raw_data, created_at
		FROM app_events
		WHERE app_id = $1 AND shopify_shop_gid = $2
		ORDER BY occurred_at DESC
	`
	rows, err := r.pool.Query(ctx, query, appID, shopGID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanEvents(rows)
}

func (r *PostgresAppEventRepository) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppEvent, error) {
	query := `
		SELECT id, app_id, shopify_shop_gid, event_type, occurred_at, raw_data, created_at
		FROM app_events
		WHERE app_id = $1
		ORDER BY occurred_at DESC
	`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanEvents(rows)
}

func (r *PostgresAppEventRepository) scanEvents(rows pgx.Rows) ([]*entity.AppEvent, error) {
	var events []*entity.AppEvent
	for rows.Next() {
		var ev entity.AppEvent
		err := rows.Scan(&ev.ID, &ev.AppID, &ev.ShopifyShopGID, &ev.EventType, &ev.OccurredAt, &ev.RawData, &ev.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &ev)
	}
	return events, rows.Err()
}

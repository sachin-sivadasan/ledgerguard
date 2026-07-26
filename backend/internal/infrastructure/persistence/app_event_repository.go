package persistence

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
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

func (r *PostgresAppEventRepository) FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters repository.EventFilters) (*repository.EventPage, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("app_id = $%d", argNum))
	args = append(args, appID)
	argNum++

	if filters.EventType != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argNum))
		args = append(args, filters.EventType)
		argNum++
	}

	if len(filters.ShopGIDs) > 0 {
		placeholders := make([]string, len(filters.ShopGIDs))
		for i, gid := range filters.ShopGIDs {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, gid)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("shopify_shop_gid IN (%s)", strings.Join(placeholders, ", ")))
	} else if filters.StoreDomain != "" {
		// Fallback: ILIKE match if no GIDs resolved (e.g., no subscriptions yet)
		conditions = append(conditions, fmt.Sprintf("shopify_shop_gid ILIKE $%d", argNum))
		args = append(args, "%"+filters.StoreDomain+"%")
		argNum++
	}

	if !filters.Since.IsZero() {
		conditions = append(conditions, fmt.Sprintf("occurred_at >= $%d", argNum))
		args = append(args, filters.Since)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM app_events WHERE %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count query failed: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, app_id, shopify_shop_gid, event_type, occurred_at, raw_data, created_at
		FROM app_events
		WHERE %s
		ORDER BY occurred_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select query failed: %w", err)
	}
	defer rows.Close()

	events, err := r.scanEvents(rows)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &repository.EventPage{
		Events:     events,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
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

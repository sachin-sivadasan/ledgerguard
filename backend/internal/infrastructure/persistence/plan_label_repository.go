package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type PostgresPlanLabelRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPlanLabelRepository(pool *pgxpool.Pool) *PostgresPlanLabelRepository {
	return &PostgresPlanLabelRepository{pool: pool}
}

// FindByAppID returns all plan labels for an app, ordered by price for stable display.
func (r *PostgresPlanLabelRepository) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.PlanLabel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, app_id, billing_interval, price_cents, label, created_at, updated_at
		FROM app_plan_labels
		WHERE app_id = $1
		ORDER BY price_cents ASC, billing_interval ASC
	`, appID)
	if err != nil {
		return nil, fmt.Errorf("query plan labels: %w", err)
	}
	defer rows.Close()

	labels := make([]*entity.PlanLabel, 0)
	for rows.Next() {
		var l entity.PlanLabel
		var interval string
		if err := rows.Scan(&l.ID, &l.AppID, &interval, &l.PriceCents, &l.Label, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plan label: %w", err)
		}
		l.BillingInterval = valueobject.BillingInterval(interval)
		labels = append(labels, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plan labels: %w", err)
	}
	return labels, nil
}

// ReplaceAll atomically swaps the app's entire label set for the provided one: delete all,
// then insert each. Runs in a transaction so a partial failure leaves the old set intact.
func (r *PostgresPlanLabelRepository) ReplaceAll(ctx context.Context, appID uuid.UUID, labels []*entity.PlanLabel) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful Commit

	if _, err := tx.Exec(ctx, `DELETE FROM app_plan_labels WHERE app_id = $1`, appID); err != nil {
		return fmt.Errorf("clear plan labels: %w", err)
	}

	for _, l := range labels {
		if _, err := tx.Exec(ctx, `
			INSERT INTO app_plan_labels (id, app_id, billing_interval, price_cents, label, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, uuid.New(), appID, string(l.BillingInterval), l.PriceCents, l.Label); err != nil {
			return fmt.Errorf("insert plan label: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit plan labels: %w", err)
	}
	return nil
}

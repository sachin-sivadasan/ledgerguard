package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresDailyMetricsSnapshotRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDailyMetricsSnapshotRepository(pool *pgxpool.Pool) *PostgresDailyMetricsSnapshotRepository {
	return &PostgresDailyMetricsSnapshotRepository{pool: pool}
}

func (r *PostgresDailyMetricsSnapshotRepository) Upsert(ctx context.Context, snapshot *entity.DailyMetricsSnapshot) error {
	query := `
		INSERT INTO daily_metrics_snapshot (
			id, app_id, date, active_mrr_cents, revenue_at_risk_cents,
			usage_revenue_cents, total_revenue_cents, renewal_success_rate,
			safe_count, one_cycle_missed_count, two_cycles_missed_count,
			churned_count, total_subscriptions, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (app_id, date) DO UPDATE SET
			active_mrr_cents = EXCLUDED.active_mrr_cents,
			revenue_at_risk_cents = EXCLUDED.revenue_at_risk_cents,
			usage_revenue_cents = EXCLUDED.usage_revenue_cents,
			total_revenue_cents = EXCLUDED.total_revenue_cents,
			renewal_success_rate = EXCLUDED.renewal_success_rate,
			safe_count = EXCLUDED.safe_count,
			one_cycle_missed_count = EXCLUDED.one_cycle_missed_count,
			two_cycles_missed_count = EXCLUDED.two_cycles_missed_count,
			churned_count = EXCLUDED.churned_count,
			total_subscriptions = EXCLUDED.total_subscriptions,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		snapshot.ID,
		snapshot.AppID,
		snapshot.Date,
		snapshot.ActiveMRRCents,
		snapshot.RevenueAtRiskCents,
		snapshot.UsageRevenueCents,
		snapshot.TotalRevenueCents,
		snapshot.RenewalSuccessRate,
		snapshot.SafeCount,
		snapshot.OneCycleMissedCount,
		snapshot.TwoCyclesMissedCount,
		snapshot.ChurnedCount,
		snapshot.TotalSubscriptions,
		snapshot.CreatedAt,
		snapshot.UpdatedAt,
	)

	return err
}

func (r *PostgresDailyMetricsSnapshotRepository) FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyMetricsSnapshot, error) {
	// Truncate to start of day
	truncatedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	query := `
		SELECT id, app_id, date, active_mrr_cents, revenue_at_risk_cents,
			usage_revenue_cents, total_revenue_cents, renewal_success_rate,
			safe_count, one_cycle_missed_count, two_cycles_missed_count,
			churned_count, total_subscriptions, created_at, updated_at
		FROM daily_metrics_snapshot
		WHERE app_id = $1 AND date = $2
	`

	snapshot := &entity.DailyMetricsSnapshot{}
	err := r.pool.QueryRow(ctx, query, appID, truncatedDate).Scan(
		&snapshot.ID,
		&snapshot.AppID,
		&snapshot.Date,
		&snapshot.ActiveMRRCents,
		&snapshot.RevenueAtRiskCents,
		&snapshot.UsageRevenueCents,
		&snapshot.TotalRevenueCents,
		&snapshot.RenewalSuccessRate,
		&snapshot.SafeCount,
		&snapshot.OneCycleMissedCount,
		&snapshot.TwoCyclesMissedCount,
		&snapshot.ChurnedCount,
		&snapshot.TotalSubscriptions,
		&snapshot.CreatedAt,
		&snapshot.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (r *PostgresDailyMetricsSnapshotRepository) FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyMetricsSnapshot, error) {
	query := `
		SELECT id, app_id, date, active_mrr_cents, revenue_at_risk_cents,
			usage_revenue_cents, total_revenue_cents, renewal_success_rate,
			safe_count, one_cycle_missed_count, two_cycles_missed_count,
			churned_count, total_subscriptions, created_at, updated_at
		FROM daily_metrics_snapshot
		WHERE app_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`

	rows, err := r.pool.Query(ctx, query, appID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*entity.DailyMetricsSnapshot
	for rows.Next() {
		snapshot := &entity.DailyMetricsSnapshot{}
		err := rows.Scan(
			&snapshot.ID,
			&snapshot.AppID,
			&snapshot.Date,
			&snapshot.ActiveMRRCents,
			&snapshot.RevenueAtRiskCents,
			&snapshot.UsageRevenueCents,
			&snapshot.TotalRevenueCents,
			&snapshot.RenewalSuccessRate,
			&snapshot.SafeCount,
			&snapshot.OneCycleMissedCount,
			&snapshot.TwoCyclesMissedCount,
			&snapshot.ChurnedCount,
			&snapshot.TotalSubscriptions,
			&snapshot.CreatedAt,
			&snapshot.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

func (r *PostgresDailyMetricsSnapshotRepository) FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyMetricsSnapshot, error) {
	query := `
		SELECT id, app_id, date, active_mrr_cents, revenue_at_risk_cents,
			usage_revenue_cents, total_revenue_cents, renewal_success_rate,
			safe_count, one_cycle_missed_count, two_cycles_missed_count,
			churned_count, total_subscriptions, created_at, updated_at
		FROM daily_metrics_snapshot
		WHERE app_id = $1
		ORDER BY date DESC
		LIMIT 1
	`

	snapshot := &entity.DailyMetricsSnapshot{}
	err := r.pool.QueryRow(ctx, query, appID).Scan(
		&snapshot.ID,
		&snapshot.AppID,
		&snapshot.Date,
		&snapshot.ActiveMRRCents,
		&snapshot.RevenueAtRiskCents,
		&snapshot.UsageRevenueCents,
		&snapshot.TotalRevenueCents,
		&snapshot.RenewalSuccessRate,
		&snapshot.SafeCount,
		&snapshot.OneCycleMissedCount,
		&snapshot.TwoCyclesMissedCount,
		&snapshot.ChurnedCount,
		&snapshot.TotalSubscriptions,
		&snapshot.CreatedAt,
		&snapshot.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

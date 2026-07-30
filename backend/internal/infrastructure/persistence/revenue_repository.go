package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// PostgresRevenueRepository implements RevenueRepository using PostgreSQL
type PostgresRevenueRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRevenueRepository creates a new PostgresRevenueRepository
func NewPostgresRevenueRepository(pool *pgxpool.Pool) *PostgresRevenueRepository {
	return &PostgresRevenueRepository{pool: pool}
}

// GetRevenueByDateRange retrieves aggregated revenue data for a date range
// Groups transactions by date and sums amounts by charge type
func (r *PostgresRevenueRepository) GetRevenueByDateRange(
	ctx context.Context,
	appID uuid.UUID,
	startDate, endDate time.Time,
) ([]repository.RevenueAggregation, error) {
	// Query to aggregate transactions by date and charge type
	query := `
		WITH daily_totals AS (
			SELECT
				DATE(transaction_date) as revenue_date,
				SUM(CASE WHEN charge_type = 'RECURRING' THEN amount_cents ELSE 0 END) as subscription_amount,
				SUM(CASE WHEN charge_type = 'USAGE' THEN amount_cents ELSE 0 END) as usage_amount,
				SUM(amount_cents) as total_amount
			FROM transactions
			WHERE app_id = $1
				AND transaction_date >= $2
				AND transaction_date <= $3
				AND charge_type IN ('RECURRING', 'USAGE')
			GROUP BY DATE(transaction_date)
			ORDER BY revenue_date ASC
		)
		SELECT
			TO_CHAR(revenue_date, 'YYYY-MM-DD') as date,
			total_amount,
			subscription_amount,
			usage_amount
		FROM daily_totals
	`

	rows, err := r.pool.Query(ctx, query, appID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aggregations []repository.RevenueAggregation

	for rows.Next() {
		var agg repository.RevenueAggregation
		err := rows.Scan(
			&agg.Date,
			&agg.TotalAmountCents,
			&agg.SubscriptionAmountCents,
			&agg.UsageAmountCents,
		)
		if err != nil {
			return nil, err
		}
		aggregations = append(aggregations, agg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return aggregations, nil
}

// GetMonthlyEarnings rolls transactions up by calendar month: gross =
// SUM(gross_amount_cents), net = SUM(net_amount_cents) (both falling back to
// amount_cents when the split column is null, matching the transactions list),
// plus per-earnings-status counts so the service can derive one month status.
// Revenue charge types only (RECURRING/USAGE/ONE_TIME); newest month first.
func (r *PostgresRevenueRepository) GetMonthlyEarnings(
	ctx context.Context,
	appID uuid.UUID,
	startDate, endDate time.Time,
) ([]repository.MonthlyEarningAggregation, error) {
	query := `
		WITH monthly AS (
			SELECT
				DATE_TRUNC('month', transaction_date) AS month_start,
				SUM(COALESCE(gross_amount_cents, amount_cents, 0)) AS gross,
				SUM(COALESCE(net_amount_cents, amount_cents, 0)) AS net,
				COUNT(*) FILTER (WHERE earnings_status = 'PENDING') AS pending_ct,
				COUNT(*) FILTER (WHERE earnings_status = 'AVAILABLE') AS available_ct,
				COUNT(*) FILTER (WHERE earnings_status = 'PAID_OUT') AS paid_ct
			FROM transactions
			WHERE app_id = $1
				AND transaction_date >= $2
				AND transaction_date <= $3
				AND charge_type IN ('RECURRING', 'USAGE', 'ONE_TIME')
			GROUP BY DATE_TRUNC('month', transaction_date)
		)
		SELECT
			TO_CHAR(month_start, 'Mon YYYY') AS month_label,
			TO_CHAR(month_start, 'YYYY-MM-DD') AS start_date,
			TO_CHAR(month_start + INTERVAL '1 month' - INTERVAL '1 day', 'YYYY-MM-DD') AS end_date,
			gross, net, pending_ct, available_ct, paid_ct
		FROM monthly
		ORDER BY month_start DESC
	`

	rows, err := r.pool.Query(ctx, query, appID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.MonthlyEarningAggregation
	for rows.Next() {
		var m repository.MonthlyEarningAggregation
		if err := rows.Scan(
			&m.MonthLabel,
			&m.StartDate,
			&m.EndDate,
			&m.GrossCents,
			&m.NetCents,
			&m.PendingCount,
			&m.AvailableCount,
			&m.PaidOutCount,
		); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

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

// GetMonthlyRevenue retrieves aggregated revenue data for a specific month
// Groups transactions by date and sums amounts by charge type
func (r *PostgresRevenueRepository) GetMonthlyRevenue(
	ctx context.Context,
	appID uuid.UUID,
	year, month int,
) ([]repository.RevenueAggregation, error) {
	// Calculate date range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0) // First day of next month

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
				AND transaction_date < $3
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

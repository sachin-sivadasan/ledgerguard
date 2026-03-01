package persistence

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type PostgresSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSubscriptionRepository(pool *pgxpool.Pool) *PostgresSubscriptionRepository {
	return &PostgresSubscriptionRepository{pool: pool}
}

func (r *PostgresSubscriptionRepository) Upsert(ctx context.Context, subscription *entity.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			shopify_shop_gid = EXCLUDED.shopify_shop_gid,
			shop_name = EXCLUDED.shop_name,
			plan_name = EXCLUDED.plan_name,
			base_price_cents = EXCLUDED.base_price_cents,
			currency = EXCLUDED.currency,
			billing_interval = EXCLUDED.billing_interval,
			status = EXCLUDED.status,
			last_recurring_charge_date = EXCLUDED.last_recurring_charge_date,
			expected_next_charge_date = EXCLUDED.expected_next_charge_date,
			risk_state = EXCLUDED.risk_state,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.pool.Exec(ctx, query,
		subscription.ID,
		subscription.AppID,
		subscription.ShopifyGID,
		subscription.ShopifyShopGID,
		subscription.MyshopifyDomain,
		subscription.ShopName,
		subscription.PlanName,
		subscription.BasePriceCents,
		subscription.Currency,
		subscription.BillingInterval.String(),
		subscription.Status,
		subscription.LastRecurringChargeDate,
		subscription.ExpectedNextChargeDate,
		subscription.RiskState.String(),
		subscription.CreatedAt,
		subscription.UpdatedAt,
		subscription.DeletedAt,
	)

	return err
}

func (r *PostgresSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanSubscription(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresSubscriptionRepository) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE app_id = $1 AND deleted_at IS NULL
		ORDER BY COALESCE(shop_name, myshopify_domain)
	`

	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSubscriptions(rows)
}

func (r *PostgresSubscriptionRepository) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE shopify_gid = $1 AND deleted_at IS NULL
	`

	return r.scanSubscription(r.pool.QueryRow(ctx, query, shopifyGID))
}

func (r *PostgresSubscriptionRepository) FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, myshopifyDomain string) (*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE app_id = $1 AND myshopify_domain = $2 AND deleted_at IS NULL
	`

	return r.scanSubscription(r.pool.QueryRow(ctx, query, appID, myshopifyDomain))
}

func (r *PostgresSubscriptionRepository) FindByRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE app_id = $1 AND risk_state = $2 AND deleted_at IS NULL
		ORDER BY COALESCE(shop_name, myshopify_domain)
	`

	rows, err := r.pool.Query(ctx, query, appID, riskState.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSubscriptions(rows)
}

func (r *PostgresSubscriptionRepository) DeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE app_id = $1`
	_, err := r.pool.Exec(ctx, query, appID)
	return err
}

func (r *PostgresSubscriptionRepository) scanSubscription(row pgx.Row) (*entity.Subscription, error) {
	var sub entity.Subscription
	var billingInterval string
	var riskState string
	var shopName *string
	var shopGID *string

	err := row.Scan(
		&sub.ID,
		&sub.AppID,
		&sub.ShopifyGID,
		&shopGID,
		&sub.MyshopifyDomain,
		&shopName,
		&sub.PlanName,
		&sub.BasePriceCents,
		&sub.Currency,
		&billingInterval,
		&sub.Status,
		&sub.LastRecurringChargeDate,
		&sub.ExpectedNextChargeDate,
		&riskState,
		&sub.CreatedAt,
		&sub.UpdatedAt,
		&sub.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}

	if shopGID != nil {
		sub.ShopifyShopGID = *shopGID
	}
	if shopName != nil {
		sub.ShopName = *shopName
	}
	sub.BillingInterval = valueobject.BillingInterval(billingInterval)
	sub.RiskState = valueobject.RiskState(riskState)

	return &sub, nil
}

func (r *PostgresSubscriptionRepository) scanSubscriptions(rows pgx.Rows) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription

	for rows.Next() {
		var sub entity.Subscription
		var billingInterval string
		var riskState string
		var shopName *string
		var shopGID *string

		err := rows.Scan(
			&sub.ID,
			&sub.AppID,
			&sub.ShopifyGID,
			&shopGID,
			&sub.MyshopifyDomain,
			&shopName,
			&sub.PlanName,
			&sub.BasePriceCents,
			&sub.Currency,
			&billingInterval,
			&sub.Status,
			&sub.LastRecurringChargeDate,
			&sub.ExpectedNextChargeDate,
			&riskState,
			&sub.CreatedAt,
			&sub.UpdatedAt,
			&sub.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if shopGID != nil {
			sub.ShopifyShopGID = *shopGID
		}
		if shopName != nil {
			sub.ShopName = *shopName
		}
		sub.BillingInterval = valueobject.BillingInterval(billingInterval)
		sub.RiskState = valueobject.RiskState(riskState)
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, rows.Err()
}

func (r *PostgresSubscriptionRepository) FindWithFilters(ctx context.Context, appID uuid.UUID, filters repository.SubscriptionFilters) (*repository.SubscriptionPage, error) {
	// Build dynamic WHERE clause
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("app_id = $%d", argNum))
	args = append(args, appID)
	argNum++

	// Exclude soft-deleted records
	conditions = append(conditions, "deleted_at IS NULL")

	// Risk states filter (multi-select)
	if len(filters.RiskStates) > 0 {
		placeholders := make([]string, len(filters.RiskStates))
		for i, rs := range filters.RiskStates {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, rs.String())
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("risk_state IN (%s)", strings.Join(placeholders, ", ")))
	}

	// Price range filter
	if filters.PriceMinCents != nil {
		conditions = append(conditions, fmt.Sprintf("base_price_cents >= $%d", argNum))
		args = append(args, *filters.PriceMinCents)
		argNum++
	}
	if filters.PriceMaxCents != nil {
		conditions = append(conditions, fmt.Sprintf("base_price_cents <= $%d", argNum))
		args = append(args, *filters.PriceMaxCents)
		argNum++
	}

	// Billing interval filter
	if filters.BillingInterval != nil {
		conditions = append(conditions, fmt.Sprintf("billing_interval = $%d", argNum))
		args = append(args, filters.BillingInterval.String())
		argNum++
	}

	// Search filter (shop_name or myshopify_domain)
	if filters.SearchTerm != "" {
		searchPattern := "%" + strings.ToLower(filters.SearchTerm) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(shop_name) LIKE $%d OR LOWER(myshopify_domain) LIKE $%d)", argNum, argNum))
		args = append(args, searchPattern)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Build ORDER BY clause
	orderBy := "COALESCE(shop_name, myshopify_domain)"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "risk_state":
			orderBy = "risk_state"
		case "base_price_cents":
			orderBy = "base_price_cents"
		case "shop_name":
			orderBy = "COALESCE(shop_name, myshopify_domain)"
		}
	}
	sortOrder := "ASC"
	if strings.ToLower(filters.SortOrder) == "desc" {
		sortOrder = "DESC"
	}

	// Set defaults for pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM subscriptions WHERE %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count query failed: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, sortOrder, argNum, argNum+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select query failed: %w", err)
	}
	defer rows.Close()

	subscriptions, err := r.scanSubscriptions(rows)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &repository.SubscriptionPage{
		Subscriptions: subscriptions,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
	}, nil
}

func (r *PostgresSubscriptionRepository) GetSummary(ctx context.Context, appID uuid.UUID) (*repository.SubscriptionSummary, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE risk_state = 'SAFE') as active_count,
			COUNT(*) FILTER (WHERE risk_state IN ('ONE_CYCLE_MISSED', 'TWO_CYCLES_MISSED')) as at_risk_count,
			COUNT(*) FILTER (WHERE risk_state = 'CHURNED') as churned_count,
			COALESCE(AVG(base_price_cents), 0)::bigint as avg_price_cents,
			COUNT(*) as total_count
		FROM subscriptions
		WHERE app_id = $1 AND deleted_at IS NULL
	`

	var summary repository.SubscriptionSummary
	err := r.pool.QueryRow(ctx, query, appID).Scan(
		&summary.ActiveCount,
		&summary.AtRiskCount,
		&summary.ChurnedCount,
		&summary.AvgPriceCents,
		&summary.TotalCount,
	)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *PostgresSubscriptionRepository) GetPriceStats(ctx context.Context, appID uuid.UUID) (*repository.PriceStats, error) {
	// Get min, max, avg
	statsQuery := `
		SELECT
			COALESCE(MIN(base_price_cents), 0),
			COALESCE(MAX(base_price_cents), 0),
			COALESCE(AVG(base_price_cents)::bigint, 0)
		FROM subscriptions
		WHERE app_id = $1 AND base_price_cents > 0 AND deleted_at IS NULL
	`

	var stats repository.PriceStats
	err := r.pool.QueryRow(ctx, statsQuery, appID).Scan(&stats.MinCents, &stats.MaxCents, &stats.AvgCents)
	if err != nil {
		return nil, err
	}

	// Get distinct prices with counts, sorted by price
	pricesQuery := `
		SELECT base_price_cents, COUNT(*) as count
		FROM subscriptions
		WHERE app_id = $1 AND base_price_cents > 0 AND deleted_at IS NULL
		GROUP BY base_price_cents
		ORDER BY base_price_cents ASC
	`

	rows, err := r.pool.Query(ctx, pricesQuery, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []repository.PricePoint
	for rows.Next() {
		var pp repository.PricePoint
		if err := rows.Scan(&pp.PriceCents, &pp.Count); err != nil {
			return nil, err
		}
		prices = append(prices, pp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats.Prices = prices
	return &stats, nil
}

// SoftDeleteByAppID marks all subscriptions for an app as deleted without removing them
// This preserves historical data for analytics and potential reactivation tracking
func (r *PostgresSubscriptionRepository) SoftDeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	query := `
		UPDATE subscriptions
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE app_id = $1 AND deleted_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, appID)
	return err
}

// FindDeletedByAppID returns all soft-deleted subscriptions for an app
// Useful for win-back campaigns and historical analysis
func (r *PostgresSubscriptionRepository) FindDeletedByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, shopify_shop_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at, deleted_at
		FROM subscriptions
		WHERE app_id = $1 AND deleted_at IS NOT NULL
		ORDER BY deleted_at DESC
	`

	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSubscriptions(rows)
}

// RestoreByID removes the soft-delete marker from a subscription
// Useful when a previously churned customer reactivates
func (r *PostgresSubscriptionRepository) RestoreByID(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE subscriptions
		SET deleted_at = NULL, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}
	return nil
}

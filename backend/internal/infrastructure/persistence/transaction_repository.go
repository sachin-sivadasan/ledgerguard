package persistence

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type PostgresTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{pool: pool}
}

func (r *PostgresTransactionRepository) Upsert(ctx context.Context, tx *entity.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, app_id, shopify_gid, myshopify_domain, shop_name, charge_type,
			gross_amount_cents, shopify_fee_cents, processing_fee_cents, tax_on_fees_cents,
			net_amount_cents, amount_cents, currency, transaction_date, created_at,
			created_date, available_date, earnings_status,
			shopify_shop_gid, shop_plan, subscription_gid, subscription_status,
			subscription_period_end, billing_interval
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			app_id = EXCLUDED.app_id,
			shop_name = EXCLUDED.shop_name,
			charge_type = EXCLUDED.charge_type,
			gross_amount_cents = EXCLUDED.gross_amount_cents,
			shopify_fee_cents = EXCLUDED.shopify_fee_cents,
			processing_fee_cents = EXCLUDED.processing_fee_cents,
			tax_on_fees_cents = EXCLUDED.tax_on_fees_cents,
			net_amount_cents = EXCLUDED.net_amount_cents,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			created_date = EXCLUDED.created_date,
			available_date = EXCLUDED.available_date,
			earnings_status = EXCLUDED.earnings_status,
			shopify_shop_gid = EXCLUDED.shopify_shop_gid,
			shop_plan = EXCLUDED.shop_plan,
			subscription_gid = EXCLUDED.subscription_gid,
			subscription_status = EXCLUDED.subscription_status,
			subscription_period_end = EXCLUDED.subscription_period_end,
			billing_interval = EXCLUDED.billing_interval
	`

	_, err := r.pool.Exec(ctx, query,
		tx.ID,
		tx.AppID,
		tx.ShopifyGID,
		tx.MyshopifyDomain,
		tx.ShopName,
		tx.ChargeType.String(),
		tx.GrossAmountCents,
		tx.ShopifyFeeCents,
		tx.ProcessingFeeCents,
		tx.TaxOnFeesCents,
		tx.NetAmountCents,
		tx.NetAmountCents, // amount_cents = net_amount_cents for backwards compatibility
		tx.Currency,
		tx.TransactionDate,
		tx.CreatedAt,
		tx.CreatedDate,
		tx.AvailableDate,
		string(tx.EarningsStatus),
		tx.ShopifyShopGID,
		tx.ShopPlan,
		tx.SubscriptionGID,
		tx.SubscriptionStatus,
		tx.SubscriptionPeriodEnd,
		tx.BillingInterval,
	)

	return err
}

func (r *PostgresTransactionRepository) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO transactions (
			id, app_id, shopify_gid, myshopify_domain, shop_name, charge_type,
			gross_amount_cents, shopify_fee_cents, processing_fee_cents, tax_on_fees_cents,
			net_amount_cents, amount_cents, currency, transaction_date, created_at,
			created_date, available_date, earnings_status,
			shopify_shop_gid, shop_plan, subscription_gid, subscription_status,
			subscription_period_end, billing_interval
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			app_id = EXCLUDED.app_id,
			shop_name = EXCLUDED.shop_name,
			charge_type = EXCLUDED.charge_type,
			gross_amount_cents = EXCLUDED.gross_amount_cents,
			shopify_fee_cents = EXCLUDED.shopify_fee_cents,
			processing_fee_cents = EXCLUDED.processing_fee_cents,
			tax_on_fees_cents = EXCLUDED.tax_on_fees_cents,
			net_amount_cents = EXCLUDED.net_amount_cents,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			created_date = EXCLUDED.created_date,
			available_date = EXCLUDED.available_date,
			earnings_status = EXCLUDED.earnings_status,
			shopify_shop_gid = EXCLUDED.shopify_shop_gid,
			shop_plan = EXCLUDED.shop_plan,
			subscription_gid = EXCLUDED.subscription_gid,
			subscription_status = EXCLUDED.subscription_status,
			subscription_period_end = EXCLUDED.subscription_period_end,
			billing_interval = EXCLUDED.billing_interval
	`

	for _, tx := range txs {
		batch.Queue(query,
			tx.ID,
			tx.AppID,
			tx.ShopifyGID,
			tx.MyshopifyDomain,
			tx.ShopName,
			tx.ChargeType.String(),
			tx.GrossAmountCents,
			tx.ShopifyFeeCents,
			tx.ProcessingFeeCents,
			tx.TaxOnFeesCents,
			tx.NetAmountCents,
			tx.NetAmountCents, // amount_cents = net_amount_cents for backwards compatibility
			tx.Currency,
			tx.TransactionDate,
			tx.CreatedAt,
			tx.CreatedDate,
			tx.AvailableDate,
			string(tx.EarningsStatus),
			tx.ShopifyShopGID,
			tx.ShopPlan,
			tx.SubscriptionGID,
			tx.SubscriptionStatus,
			tx.SubscriptionPeriodEnd,
			tx.BillingInterval,
		)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range txs {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresTransactionRepository) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, charge_type,
		       COALESCE(gross_amount_cents, 0), COALESCE(shopify_fee_cents, 0),
		       COALESCE(processing_fee_cents, 0), COALESCE(tax_on_fees_cents, 0),
		       COALESCE(net_amount_cents, amount_cents), currency, transaction_date, created_at,
		       created_date, available_date, earnings_status,
		       shopify_shop_gid, shop_plan, subscription_gid, subscription_status,
		       subscription_period_end, billing_interval
		FROM transactions
		WHERE app_id = $1 AND transaction_date >= $2 AND transaction_date <= $3
		ORDER BY transaction_date DESC
	`

	rows, err := r.pool.Query(ctx, query, appID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		tx, err := r.scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (r *PostgresTransactionRepository) FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters repository.TransactionFilters) (*repository.TransactionPage, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("app_id = $%d", argNum))
	args = append(args, appID)
	argNum++

	if !filters.From.IsZero() {
		conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argNum))
		args = append(args, filters.From)
		argNum++
	}
	if !filters.To.IsZero() {
		conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argNum))
		args = append(args, filters.To)
		argNum++
	}
	if filters.ChargeType != "" {
		conditions = append(conditions, fmt.Sprintf("charge_type = $%d", argNum))
		args = append(args, filters.ChargeType)
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

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM transactions WHERE %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count query failed: %w", err)
	}

	// Fetch page
	query := fmt.Sprintf(`
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, charge_type,
		       COALESCE(gross_amount_cents, 0), COALESCE(shopify_fee_cents, 0),
		       COALESCE(processing_fee_cents, 0), COALESCE(tax_on_fees_cents, 0),
		       COALESCE(net_amount_cents, amount_cents), currency, transaction_date, created_at,
		       created_date, available_date, earnings_status,
		       shopify_shop_gid, shop_plan, subscription_gid, subscription_status,
		       subscription_period_end, billing_interval
		FROM transactions
		WHERE %s
		ORDER BY transaction_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select query failed: %w", err)
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		tx, err := r.scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &repository.TransactionPage{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}, nil
}

// GetTransactionSummary sums gross/net and counts over the filtered set (same WHERE as
// FindByAppIDPaginated, minus pagination) so the UI totals reflect the whole set, not the
// loaded page. Net uses COALESCE(net_amount_cents, amount_cents), matching the list query.
func (r *PostgresTransactionRepository) GetTransactionSummary(ctx context.Context, appID uuid.UUID, filters repository.TransactionFilters) (*repository.TransactionSummary, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("app_id = $%d", argNum))
	args = append(args, appID)
	argNum++

	if !filters.From.IsZero() {
		conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argNum))
		args = append(args, filters.From)
		argNum++
	}
	if !filters.To.IsZero() {
		conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argNum))
		args = append(args, filters.To)
		argNum++
	}
	if filters.ChargeType != "" {
		conditions = append(conditions, fmt.Sprintf("charge_type = $%d", argNum))
		args = append(args, filters.ChargeType)
		argNum++
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(COALESCE(gross_amount_cents, 0)), 0),
		       COALESCE(SUM(COALESCE(net_amount_cents, amount_cents)), 0),
		       COUNT(*)
		FROM transactions WHERE %s
	`, strings.Join(conditions, " AND "))

	var summary repository.TransactionSummary
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&summary.GrossCents, &summary.NetCents, &summary.Count); err != nil {
		return nil, fmt.Errorf("transaction summary query failed: %w", err)
	}
	return &summary, nil
}

// scanTransaction scans a row into a Transaction entity
func (r *PostgresTransactionRepository) scanTransaction(rows pgx.Rows) (*entity.Transaction, error) {
	var tx entity.Transaction
	var chargeType string
	var shopName, shopifyShopGID, shopPlan, subscriptionGID, subscriptionStatus, billingInterval *string
	var earningsStatus string
	var subscriptionPeriodEnd *time.Time

	err := rows.Scan(
		&tx.ID,
		&tx.AppID,
		&tx.ShopifyGID,
		&tx.MyshopifyDomain,
		&shopName,
		&chargeType,
		&tx.GrossAmountCents,
		&tx.ShopifyFeeCents,
		&tx.ProcessingFeeCents,
		&tx.TaxOnFeesCents,
		&tx.NetAmountCents,
		&tx.Currency,
		&tx.TransactionDate,
		&tx.CreatedAt,
		&tx.CreatedDate,
		&tx.AvailableDate,
		&earningsStatus,
		&shopifyShopGID,
		&shopPlan,
		&subscriptionGID,
		&subscriptionStatus,
		&subscriptionPeriodEnd,
		&billingInterval,
	)
	if err != nil {
		return nil, err
	}

	tx.ChargeType = valueobject.ChargeType(chargeType)
	tx.EarningsStatus = entity.EarningsStatus(earningsStatus)
	if shopName != nil {
		tx.ShopName = *shopName
	}
	if shopifyShopGID != nil {
		tx.ShopifyShopGID = *shopifyShopGID
	}
	if shopPlan != nil {
		tx.ShopPlan = *shopPlan
	}
	if subscriptionGID != nil {
		tx.SubscriptionGID = *subscriptionGID
	}
	if subscriptionStatus != nil {
		tx.SubscriptionStatus = *subscriptionStatus
	}
	tx.SubscriptionPeriodEnd = subscriptionPeriodEnd
	if billingInterval != nil {
		tx.BillingInterval = *billingInterval
	}

	return &tx, nil
}

func (r *PostgresTransactionRepository) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Transaction, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, charge_type,
		       COALESCE(gross_amount_cents, 0), COALESCE(shopify_fee_cents, 0),
		       COALESCE(processing_fee_cents, 0), COALESCE(tax_on_fees_cents, 0),
		       COALESCE(net_amount_cents, amount_cents), currency, transaction_date, created_at,
		       created_date, available_date, earnings_status,
		       shopify_shop_gid, shop_plan, subscription_gid, subscription_status,
		       subscription_period_end, billing_interval
		FROM transactions
		WHERE shopify_gid = $1
	`

	var tx entity.Transaction
	var chargeType string
	var shopName, shopifyShopGID, shopPlan, subscriptionGID, subscriptionStatus, billingInterval *string
	var earningsStatus string
	var subscriptionPeriodEnd *time.Time

	err := r.pool.QueryRow(ctx, query, shopifyGID).Scan(
		&tx.ID,
		&tx.AppID,
		&tx.ShopifyGID,
		&tx.MyshopifyDomain,
		&shopName,
		&chargeType,
		&tx.GrossAmountCents,
		&tx.ShopifyFeeCents,
		&tx.ProcessingFeeCents,
		&tx.TaxOnFeesCents,
		&tx.NetAmountCents,
		&tx.Currency,
		&tx.TransactionDate,
		&tx.CreatedAt,
		&tx.CreatedDate,
		&tx.AvailableDate,
		&earningsStatus,
		&shopifyShopGID,
		&shopPlan,
		&subscriptionGID,
		&subscriptionStatus,
		&subscriptionPeriodEnd,
		&billingInterval,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	tx.ChargeType = valueobject.ChargeType(chargeType)
	tx.EarningsStatus = entity.EarningsStatus(earningsStatus)
	if shopName != nil {
		tx.ShopName = *shopName
	}
	if shopifyShopGID != nil {
		tx.ShopifyShopGID = *shopifyShopGID
	}
	if shopPlan != nil {
		tx.ShopPlan = *shopPlan
	}
	if subscriptionGID != nil {
		tx.SubscriptionGID = *subscriptionGID
	}
	if subscriptionStatus != nil {
		tx.SubscriptionStatus = *subscriptionStatus
	}
	tx.SubscriptionPeriodEnd = subscriptionPeriodEnd
	if billingInterval != nil {
		tx.BillingInterval = *billingInterval
	}
	return &tx, nil
}

func (r *PostgresTransactionRepository) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE app_id = $1`

	var count int64
	err := r.pool.QueryRow(ctx, query, appID).Scan(&count)
	return count, err
}

func (r *PostgresTransactionRepository) GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*repository.EarningsSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN earnings_status = 'PENDING' THEN net_amount_cents ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN earnings_status = 'AVAILABLE' THEN net_amount_cents ELSE 0 END), 0) as available,
			COALESCE(SUM(CASE WHEN earnings_status = 'PAID_OUT' THEN net_amount_cents ELSE 0 END), 0) as paid_out
		FROM transactions
		WHERE app_id = $1
	`

	var summary repository.EarningsSummary
	err := r.pool.QueryRow(ctx, query, appID).Scan(
		&summary.PendingCents,
		&summary.AvailableCents,
		&summary.PaidOutCents,
	)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *PostgresTransactionRepository) GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]repository.EarningsByDate, error) {
	query := `
		SELECT available_date, SUM(net_amount_cents) as amount
		FROM transactions
		WHERE app_id = $1 AND earnings_status = 'PENDING'
		GROUP BY available_date
		ORDER BY available_date ASC
	`

	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.EarningsByDate
	for rows.Next() {
		var entry repository.EarningsByDate
		if err := rows.Scan(&entry.Date, &entry.AmountCents); err != nil {
			return nil, err
		}
		results = append(results, entry)
	}

	return results, rows.Err()
}

func (r *PostgresTransactionRepository) GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]repository.EarningsByDate, error) {
	query := `
		SELECT available_date, SUM(net_amount_cents) as amount
		FROM transactions
		WHERE app_id = $1
			AND earnings_status = 'PENDING'
			AND available_date >= NOW()
			AND available_date <= NOW() + $2 * INTERVAL '1 day'
		GROUP BY available_date
		ORDER BY available_date ASC
	`

	rows, err := r.pool.Query(ctx, query, appID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.EarningsByDate
	for rows.Next() {
		var entry repository.EarningsByDate
		if err := rows.Scan(&entry.Date, &entry.AmountCents); err != nil {
			return nil, err
		}
		results = append(results, entry)
	}

	return results, rows.Err()
}

func (r *PostgresTransactionRepository) FindByDomain(ctx context.Context, appID uuid.UUID, domain string, from, to time.Time) ([]*entity.Transaction, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, charge_type,
		       COALESCE(gross_amount_cents, 0), COALESCE(shopify_fee_cents, 0),
		       COALESCE(processing_fee_cents, 0), COALESCE(tax_on_fees_cents, 0),
		       COALESCE(net_amount_cents, amount_cents), currency, transaction_date, created_at,
		       created_date, available_date, earnings_status,
		       shopify_shop_gid, shop_plan, subscription_gid, subscription_status,
		       subscription_period_end, billing_interval
		FROM transactions
		WHERE app_id = $1 AND myshopify_domain = $2 AND transaction_date >= $3 AND transaction_date <= $4
		ORDER BY transaction_date DESC
	`

	rows, err := r.pool.Query(ctx, query, appID, domain, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		tx, err := r.scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (r *PostgresTransactionRepository) GetEarningsSummaryByDomain(ctx context.Context, appID uuid.UUID, domain string) (*repository.EarningsSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN earnings_status = 'PENDING' THEN net_amount_cents ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN earnings_status = 'AVAILABLE' THEN net_amount_cents ELSE 0 END), 0) as available,
			COALESCE(SUM(CASE WHEN earnings_status = 'PAID_OUT' THEN net_amount_cents ELSE 0 END), 0) as paid_out
		FROM transactions
		WHERE app_id = $1 AND myshopify_domain = $2
	`

	var summary repository.EarningsSummary
	err := r.pool.QueryRow(ctx, query, appID, domain).Scan(
		&summary.PendingCents,
		&summary.AvailableCents,
		&summary.PaidOutCents,
	)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

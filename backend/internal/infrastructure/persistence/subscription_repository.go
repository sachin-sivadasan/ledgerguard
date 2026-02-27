package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
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
			id, app_id, shopify_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			shop_name = EXCLUDED.shop_name,
			plan_name = EXCLUDED.plan_name,
			base_price_cents = EXCLUDED.base_price_cents,
			currency = EXCLUDED.currency,
			billing_interval = EXCLUDED.billing_interval,
			status = EXCLUDED.status,
			last_recurring_charge_date = EXCLUDED.last_recurring_charge_date,
			expected_next_charge_date = EXCLUDED.expected_next_charge_date,
			risk_state = EXCLUDED.risk_state,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		subscription.ID,
		subscription.AppID,
		subscription.ShopifyGID,
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
	)

	return err
}

func (r *PostgresSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	return r.scanSubscription(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresSubscriptionRepository) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at
		FROM subscriptions
		WHERE app_id = $1
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
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at
		FROM subscriptions
		WHERE shopify_gid = $1
	`

	return r.scanSubscription(r.pool.QueryRow(ctx, query, shopifyGID))
}

func (r *PostgresSubscriptionRepository) FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, myshopifyDomain string) (*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at
		FROM subscriptions
		WHERE app_id = $1 AND myshopify_domain = $2
	`

	return r.scanSubscription(r.pool.QueryRow(ctx, query, appID, myshopifyDomain))
}

func (r *PostgresSubscriptionRepository) FindByRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.Subscription, error) {
	query := `
		SELECT id, app_id, shopify_gid, myshopify_domain, shop_name, plan_name,
			base_price_cents, currency, billing_interval, status,
			last_recurring_charge_date, expected_next_charge_date, risk_state,
			created_at, updated_at
		FROM subscriptions
		WHERE app_id = $1 AND risk_state = $2
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

	err := row.Scan(
		&sub.ID,
		&sub.AppID,
		&sub.ShopifyGID,
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
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
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

		err := rows.Scan(
			&sub.ID,
			&sub.AppID,
			&sub.ShopifyGID,
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
		)
		if err != nil {
			return nil, err
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

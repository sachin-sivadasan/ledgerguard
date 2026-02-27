package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

var ErrSubscriptionStatusNotFound = errors.New("subscription status not found")

// PostgresSubscriptionStatusRepository implements SubscriptionStatusRepository using PostgreSQL
type PostgresSubscriptionStatusRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSubscriptionStatusRepository creates a new PostgresSubscriptionStatusRepository
func NewPostgresSubscriptionStatusRepository(pool *pgxpool.Pool) *PostgresSubscriptionStatusRepository {
	return &PostgresSubscriptionStatusRepository{pool: pool}
}

// Upsert creates or updates a subscription status
func (r *PostgresSubscriptionStatusRepository) Upsert(ctx context.Context, status *entity.SubscriptionStatus) error {
	query := `
		INSERT INTO api_subscription_status (
			id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			shop_name = EXCLUDED.shop_name,
			plan_name = EXCLUDED.plan_name,
			risk_state = EXCLUDED.risk_state,
			is_paid_current_cycle = EXCLUDED.is_paid_current_cycle,
			months_overdue = EXCLUDED.months_overdue,
			last_successful_charge_date = EXCLUDED.last_successful_charge_date,
			expected_next_charge_date = EXCLUDED.expected_next_charge_date,
			status = EXCLUDED.status,
			last_synced_at = EXCLUDED.last_synced_at
	`

	_, err := r.pool.Exec(ctx, query,
		status.ID,
		status.ShopifyGID,
		status.AppID,
		status.MyshopifyDomain,
		status.ShopName,
		status.PlanName,
		string(status.RiskState),
		status.IsPaidCurrentCycle,
		status.MonthsOverdue,
		status.LastSuccessfulChargeDate,
		status.ExpectedNextChargeDate,
		status.Status,
		status.LastSyncedAt,
	)

	return err
}

// UpsertBatch creates or updates multiple subscription statuses
func (r *PostgresSubscriptionStatusRepository) UpsertBatch(ctx context.Context, statuses []*entity.SubscriptionStatus) error {
	if len(statuses) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO api_subscription_status (
			id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (shopify_gid) DO UPDATE SET
			shop_name = EXCLUDED.shop_name,
			plan_name = EXCLUDED.plan_name,
			risk_state = EXCLUDED.risk_state,
			is_paid_current_cycle = EXCLUDED.is_paid_current_cycle,
			months_overdue = EXCLUDED.months_overdue,
			last_successful_charge_date = EXCLUDED.last_successful_charge_date,
			expected_next_charge_date = EXCLUDED.expected_next_charge_date,
			status = EXCLUDED.status,
			last_synced_at = EXCLUDED.last_synced_at
	`

	for _, status := range statuses {
		batch.Queue(query,
			status.ID,
			status.ShopifyGID,
			status.AppID,
			status.MyshopifyDomain,
			status.ShopName,
			status.PlanName,
			string(status.RiskState),
			status.IsPaidCurrentCycle,
			status.MonthsOverdue,
			status.LastSuccessfulChargeDate,
			status.ExpectedNextChargeDate,
			status.Status,
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

// GetByShopifyGID retrieves a subscription status by Shopify GID
func (r *PostgresSubscriptionStatusRepository) GetByShopifyGID(ctx context.Context, shopifyGID string) (*entity.SubscriptionStatus, error) {
	query := `
		SELECT id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		FROM api_subscription_status
		WHERE shopify_gid = $1
	`

	return r.scanStatus(r.pool.QueryRow(ctx, query, shopifyGID))
}

// GetByShopifyGIDs retrieves multiple subscription statuses by Shopify GIDs
func (r *PostgresSubscriptionStatusRepository) GetByShopifyGIDs(ctx context.Context, shopifyGIDs []string) ([]*entity.SubscriptionStatus, error) {
	if len(shopifyGIDs) == 0 {
		return []*entity.SubscriptionStatus{}, nil
	}

	query := `
		SELECT id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		FROM api_subscription_status
		WHERE shopify_gid = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, shopifyGIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// GetByDomain retrieves a subscription status by myshopify domain
func (r *PostgresSubscriptionStatusRepository) GetByDomain(ctx context.Context, appID uuid.UUID, domain string) (*entity.SubscriptionStatus, error) {
	query := `
		SELECT id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		FROM api_subscription_status
		WHERE app_id = $1 AND myshopify_domain = $2
	`

	return r.scanStatus(r.pool.QueryRow(ctx, query, appID, domain))
}

// GetByDomains retrieves multiple subscription statuses by domains
func (r *PostgresSubscriptionStatusRepository) GetByDomains(ctx context.Context, appID uuid.UUID, domains []string) ([]*entity.SubscriptionStatus, error) {
	if len(domains) == 0 {
		return []*entity.SubscriptionStatus{}, nil
	}

	query := `
		SELECT id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		FROM api_subscription_status
		WHERE app_id = $1 AND myshopify_domain = ANY($2)
	`

	rows, err := r.pool.Query(ctx, query, appID, domains)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// GetByAppID retrieves all subscription statuses for an app
func (r *PostgresSubscriptionStatusRepository) GetByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.SubscriptionStatus, error) {
	query := `
		SELECT id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		FROM api_subscription_status
		WHERE app_id = $1
		ORDER BY COALESCE(shop_name, myshopify_domain)
	`

	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// GetByAppIDAndRiskState retrieves subscription statuses filtered by risk state
func (r *PostgresSubscriptionStatusRepository) GetByAppIDAndRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.SubscriptionStatus, error) {
	query := `
		SELECT id, shopify_gid, app_id, myshopify_domain, shop_name, plan_name,
			risk_state, is_paid_current_cycle, months_overdue,
			last_successful_charge_date, expected_next_charge_date, status, last_synced_at
		FROM api_subscription_status
		WHERE app_id = $1 AND risk_state = $2
		ORDER BY COALESCE(shop_name, myshopify_domain)
	`

	rows, err := r.pool.Query(ctx, query, appID, string(riskState))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatuses(rows)
}

// DeleteByAppID deletes all subscription statuses for an app
func (r *PostgresSubscriptionStatusRepository) DeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	query := `DELETE FROM api_subscription_status WHERE app_id = $1`
	_, err := r.pool.Exec(ctx, query, appID)
	return err
}

func (r *PostgresSubscriptionStatusRepository) scanStatus(row pgx.Row) (*entity.SubscriptionStatus, error) {
	var status entity.SubscriptionStatus
	var shopName, planName *string
	var riskState string

	err := row.Scan(
		&status.ID,
		&status.ShopifyGID,
		&status.AppID,
		&status.MyshopifyDomain,
		&shopName,
		&planName,
		&riskState,
		&status.IsPaidCurrentCycle,
		&status.MonthsOverdue,
		&status.LastSuccessfulChargeDate,
		&status.ExpectedNextChargeDate,
		&status.Status,
		&status.LastSyncedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionStatusNotFound
		}
		return nil, err
	}

	if shopName != nil {
		status.ShopName = *shopName
	}
	if planName != nil {
		status.PlanName = *planName
	}
	status.RiskState = valueobject.RiskState(riskState)

	return &status, nil
}

func (r *PostgresSubscriptionStatusRepository) scanStatuses(rows pgx.Rows) ([]*entity.SubscriptionStatus, error) {
	var statuses []*entity.SubscriptionStatus

	for rows.Next() {
		var status entity.SubscriptionStatus
		var shopName, planName *string
		var riskState string

		err := rows.Scan(
			&status.ID,
			&status.ShopifyGID,
			&status.AppID,
			&status.MyshopifyDomain,
			&shopName,
			&planName,
			&riskState,
			&status.IsPaidCurrentCycle,
			&status.MonthsOverdue,
			&status.LastSuccessfulChargeDate,
			&status.ExpectedNextChargeDate,
			&status.Status,
			&status.LastSyncedAt,
		)
		if err != nil {
			return nil, err
		}

		if shopName != nil {
			status.ShopName = *shopName
		}
		if planName != nil {
			status.PlanName = *planName
		}
		status.RiskState = valueobject.RiskState(riskState)
		statuses = append(statuses, &status)
	}

	return statuses, rows.Err()
}

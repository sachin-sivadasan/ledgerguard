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

var ErrBillingSubscriptionNotFound = errors.New("billing subscription not found")

type PostgresBillingSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBillingSubscriptionRepository(pool *pgxpool.Pool) *PostgresBillingSubscriptionRepository {
	return &PostgresBillingSubscriptionRepository{pool: pool}
}

func (r *PostgresBillingSubscriptionRepository) Create(ctx context.Context, bs *entity.BillingSubscription) error {
	query := `
		INSERT INTO billing_subscriptions (
			id, user_id, razorpay_subscription_id, razorpay_plan_id, razorpay_customer_id,
			plan, status, amount_cents, currency, current_period_start, current_period_end,
			short_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.pool.Exec(ctx, query,
		bs.ID,
		bs.UserID,
		bs.RazorpaySubscriptionID,
		bs.RazorpayPlanID,
		bs.RazorpayCustomerID,
		bs.Plan.String(),
		bs.Status.String(),
		bs.AmountCents,
		bs.Currency,
		bs.CurrentPeriodStart,
		bs.CurrentPeriodEnd,
		bs.ShortURL,
		bs.CreatedAt,
		bs.UpdatedAt,
	)
	return err
}

func (r *PostgresBillingSubscriptionRepository) Update(ctx context.Context, bs *entity.BillingSubscription) error {
	query := `
		UPDATE billing_subscriptions SET
			status = $2,
			amount_cents = $3,
			current_period_start = $4,
			current_period_end = $5,
			short_url = $6,
			updated_at = $7
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		bs.ID,
		bs.Status.String(),
		bs.AmountCents,
		bs.CurrentPeriodStart,
		bs.CurrentPeriodEnd,
		bs.ShortURL,
		bs.UpdatedAt,
	)
	return err
}

func (r *PostgresBillingSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.BillingSubscription, error) {
	query := `
		SELECT id, user_id, razorpay_subscription_id, razorpay_plan_id, razorpay_customer_id,
			plan, status, amount_cents, currency, current_period_start, current_period_end,
			short_url, created_at, updated_at
		FROM billing_subscriptions
		WHERE id = $1
	`
	return r.scanRow(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresBillingSubscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.BillingSubscription, error) {
	query := `
		SELECT id, user_id, razorpay_subscription_id, razorpay_plan_id, razorpay_customer_id,
			plan, status, amount_cents, currency, current_period_start, current_period_end,
			short_url, created_at, updated_at
		FROM billing_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*entity.BillingSubscription
	for rows.Next() {
		bs, err := r.scanRows(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, bs)
	}
	return results, rows.Err()
}

func (r *PostgresBillingSubscriptionRepository) FindByRazorpaySubscriptionID(ctx context.Context, razorpaySubID string) (*entity.BillingSubscription, error) {
	query := `
		SELECT id, user_id, razorpay_subscription_id, razorpay_plan_id, razorpay_customer_id,
			plan, status, amount_cents, currency, current_period_start, current_period_end,
			short_url, created_at, updated_at
		FROM billing_subscriptions
		WHERE razorpay_subscription_id = $1
	`
	return r.scanRow(r.pool.QueryRow(ctx, query, razorpaySubID))
}

func (r *PostgresBillingSubscriptionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.BillingSubscription, error) {
	query := `
		SELECT id, user_id, razorpay_subscription_id, razorpay_plan_id, razorpay_customer_id,
			plan, status, amount_cents, currency, current_period_start, current_period_end,
			short_url, created_at, updated_at
		FROM billing_subscriptions
		WHERE user_id = $1 AND status = 'ACTIVE'
		ORDER BY created_at DESC
		LIMIT 1
	`
	return r.scanRow(r.pool.QueryRow(ctx, query, userID))
}

func (r *PostgresBillingSubscriptionRepository) scanRow(row pgx.Row) (*entity.BillingSubscription, error) {
	var bs entity.BillingSubscription
	var plan, status string

	err := row.Scan(
		&bs.ID,
		&bs.UserID,
		&bs.RazorpaySubscriptionID,
		&bs.RazorpayPlanID,
		&bs.RazorpayCustomerID,
		&plan,
		&status,
		&bs.AmountCents,
		&bs.Currency,
		&bs.CurrentPeriodStart,
		&bs.CurrentPeriodEnd,
		&bs.ShortURL,
		&bs.CreatedAt,
		&bs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBillingSubscriptionNotFound
		}
		return nil, err
	}

	bs.Plan = valueobject.ParseBillingPlan(plan)
	bs.Status = valueobject.ParseBillingSubscriptionStatus(status)
	return &bs, nil
}

func (r *PostgresBillingSubscriptionRepository) scanRows(rows pgx.Rows) (*entity.BillingSubscription, error) {
	var bs entity.BillingSubscription
	var plan, status string

	err := rows.Scan(
		&bs.ID,
		&bs.UserID,
		&bs.RazorpaySubscriptionID,
		&bs.RazorpayPlanID,
		&bs.RazorpayCustomerID,
		&plan,
		&status,
		&bs.AmountCents,
		&bs.Currency,
		&bs.CurrentPeriodStart,
		&bs.CurrentPeriodEnd,
		&bs.ShortURL,
		&bs.CreatedAt,
		&bs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	bs.Plan = valueobject.ParseBillingPlan(plan)
	bs.Status = valueobject.ParseBillingSubscriptionStatus(status)
	return &bs, nil
}

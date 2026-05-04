package persistence

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

type PostgresAdminRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAdminRepository(pool *pgxpool.Pool) *PostgresAdminRepository {
	return &PostgresAdminRepository{pool: pool}
}

// ListUsers returns all users with app count and partner connection status.
func (r *PostgresAdminRepository) ListUsers(ctx context.Context) ([]repository.AdminUserRow, error) {
	query := `
		SELECT
			u.id,
			u.email,
			COALESCE(u.role, 'OWNER'),
			COALESCE(u.plan_tier, 'FREE'),
			u.created_at,
			u.onboarding_completed_at,
			COALESCE(app_counts.cnt, 0),
			(pa.id IS NOT NULL) AS partner_connected
		FROM users u
		LEFT JOIN partner_accounts pa ON pa.user_id = u.id
		LEFT JOIN (
			SELECT a.partner_account_id, COUNT(*) AS cnt
			FROM apps a
			GROUP BY a.partner_account_id
		) app_counts ON app_counts.partner_account_id = pa.id
		ORDER BY u.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []repository.AdminUserRow
	for rows.Next() {
		var row repository.AdminUserRow
		if err := rows.Scan(
			&row.ID,
			&row.Email,
			&row.Role,
			&row.PlanTier,
			&row.CreatedAt,
			&row.OnboardingCompletedAt,
			&row.AppCount,
			&row.PartnerConnected,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, rows.Err()
}

// GetOnboardingFunnel returns aggregate onboarding funnel metrics.
func (r *PostgresAdminRepository) GetOnboardingFunnel(ctx context.Context) (*repository.OnboardingFunnel, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(DISTINCT pa.user_id) FROM partner_accounts pa) AS partner_connected,
			(SELECT COUNT(DISTINCT pa.user_id)
			 FROM partner_accounts pa
			 JOIN apps a ON a.partner_account_id = pa.id) AS app_selected,
			(SELECT COUNT(*) FROM users WHERE onboarding_completed_at IS NOT NULL) AS onboarding_complete
	`

	var funnel repository.OnboardingFunnel
	err := r.pool.QueryRow(ctx, query).Scan(
		&funnel.TotalUsers,
		&funnel.PartnerConnected,
		&funnel.AppSelected,
		&funnel.OnboardingComplete,
	)
	if err != nil {
		return nil, err
	}

	return &funnel, nil
}

// ListSyncJobs returns recent sync jobs with app and user info.
func (r *PostgresAdminRepository) ListSyncJobs(ctx context.Context, limit int) ([]repository.AdminSyncJobRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	query := `
		SELECT
			sj.id,
			sj.app_id,
			COALESCE(a.name, ''),
			COALESCE(u.email, ''),
			sj.job_type,
			sj.status,
			COALESCE(sj.total_items, 0),
			COALESCE(sj.completed_items, 0),
			COALESCE(sj.error_message, ''),
			sj.started_at,
			sj.completed_at,
			sj.created_at
		FROM sync_jobs sj
		LEFT JOIN apps a ON a.id = sj.app_id
		LEFT JOIN users u ON u.id = sj.user_id
		ORDER BY sj.created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []repository.AdminSyncJobRow
	for rows.Next() {
		var row repository.AdminSyncJobRow
		var startedAt, completedAt *time.Time
		if err := rows.Scan(
			&row.ID,
			&row.AppID,
			&row.AppName,
			&row.UserEmail,
			&row.JobType,
			&row.Status,
			&row.TotalItems,
			&row.CompletedItems,
			&row.ErrorMessage,
			&startedAt,
			&completedAt,
			&row.CreatedAt,
		); err != nil {
			return nil, err
		}
		row.StartedAt = startedAt
		row.CompletedAt = completedAt
		result = append(result, row)
	}

	return result, rows.Err()
}

// ListBillingSubscriptions returns all billing subscriptions with user info.
func (r *PostgresAdminRepository) ListBillingSubscriptions(ctx context.Context) ([]repository.AdminBillingRow, error) {
	query := `
		SELECT
			bs.id,
			COALESCE(u.email, ''),
			bs.plan,
			bs.status,
			COALESCE(bs.amount_cents, 0),
			COALESCE(bs.currency, 'INR'),
			bs.current_period_start,
			bs.current_period_end,
			bs.created_at
		FROM billing_subscriptions bs
		LEFT JOIN users u ON u.id = bs.user_id
		ORDER BY bs.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []repository.AdminBillingRow
	for rows.Next() {
		var row repository.AdminBillingRow
		if err := rows.Scan(
			&row.ID,
			&row.UserEmail,
			&row.Plan,
			&row.Status,
			&row.AmountCents,
			&row.Currency,
			&row.CurrentPeriodStart,
			&row.CurrentPeriodEnd,
			&row.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, rows.Err()
}

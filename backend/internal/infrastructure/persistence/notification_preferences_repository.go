package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// ErrNotificationPreferencesNotFound is returned when notification preferences are not found
var ErrNotificationPreferencesNotFound = errors.New("notification preferences not found")

type PostgresNotificationPreferencesRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresNotificationPreferencesRepository(pool *pgxpool.Pool) *PostgresNotificationPreferencesRepository {
	return &PostgresNotificationPreferencesRepository{pool: pool}
}

func (r *PostgresNotificationPreferencesRepository) Create(ctx context.Context, prefs *entity.NotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (id, user_id, critical_enabled, daily_summary_enabled, daily_summary_time, slack_webhook_url,
			email_enabled, slack_enabled, churn_alerts_enabled, revenue_alerts_enabled, review_alerts_enabled, risk_threshold_days,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.pool.Exec(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Format("15:04:05"),
		prefs.SlackWebhookURL,
		prefs.EmailEnabled,
		prefs.SlackEnabled,
		prefs.ChurnAlertsEnabled,
		prefs.RevenueAlertsEnabled,
		prefs.ReviewAlertsEnabled,
		prefs.RiskThresholdDays,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	return err
}

func (r *PostgresNotificationPreferencesRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, critical_enabled, daily_summary_enabled, daily_summary_time, slack_webhook_url,
			email_enabled, slack_enabled, churn_alerts_enabled, revenue_alerts_enabled, review_alerts_enabled, risk_threshold_days,
			created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	var prefs entity.NotificationPreferences
	var summaryTime time.Time
	var slackURL *string

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.CriticalEnabled,
		&prefs.DailySummaryEnabled,
		&summaryTime,
		&slackURL,
		&prefs.EmailEnabled,
		&prefs.SlackEnabled,
		&prefs.ChurnAlertsEnabled,
		&prefs.RevenueAlertsEnabled,
		&prefs.ReviewAlertsEnabled,
		&prefs.RiskThresholdDays,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationPreferencesNotFound
		}
		return nil, err
	}

	prefs.DailySummaryTime = summaryTime

	if slackURL != nil {
		prefs.SlackWebhookURL = *slackURL
	}

	return &prefs, nil
}

func (r *PostgresNotificationPreferencesRepository) Update(ctx context.Context, prefs *entity.NotificationPreferences) error {
	query := `
		UPDATE notification_preferences
		SET critical_enabled = $2, daily_summary_enabled = $3, daily_summary_time = $4, slack_webhook_url = $5,
			email_enabled = $6, slack_enabled = $7, churn_alerts_enabled = $8, revenue_alerts_enabled = $9,
			review_alerts_enabled = $10, risk_threshold_days = $11, updated_at = $12
		WHERE user_id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Format("15:04:05"),
		prefs.SlackWebhookURL,
		prefs.EmailEnabled,
		prefs.SlackEnabled,
		prefs.ChurnAlertsEnabled,
		prefs.RevenueAlertsEnabled,
		prefs.ReviewAlertsEnabled,
		prefs.RiskThresholdDays,
		prefs.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotificationPreferencesNotFound
	}

	return nil
}

func (r *PostgresNotificationPreferencesRepository) Upsert(ctx context.Context, prefs *entity.NotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (id, user_id, critical_enabled, daily_summary_enabled, daily_summary_time, slack_webhook_url,
			email_enabled, slack_enabled, churn_alerts_enabled, revenue_alerts_enabled, review_alerts_enabled, risk_threshold_days,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (user_id) DO UPDATE SET
			critical_enabled = EXCLUDED.critical_enabled,
			daily_summary_enabled = EXCLUDED.daily_summary_enabled,
			daily_summary_time = EXCLUDED.daily_summary_time,
			slack_webhook_url = EXCLUDED.slack_webhook_url,
			email_enabled = EXCLUDED.email_enabled,
			slack_enabled = EXCLUDED.slack_enabled,
			churn_alerts_enabled = EXCLUDED.churn_alerts_enabled,
			revenue_alerts_enabled = EXCLUDED.revenue_alerts_enabled,
			review_alerts_enabled = EXCLUDED.review_alerts_enabled,
			risk_threshold_days = EXCLUDED.risk_threshold_days,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Format("15:04:05"),
		prefs.SlackWebhookURL,
		prefs.EmailEnabled,
		prefs.SlackEnabled,
		prefs.ChurnAlertsEnabled,
		prefs.RevenueAlertsEnabled,
		prefs.ReviewAlertsEnabled,
		prefs.RiskThresholdDays,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	return err
}

func (r *PostgresNotificationPreferencesRepository) FindUsersWithDailySummaryAtHour(ctx context.Context, hour int) ([]uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM notification_preferences
		WHERE daily_summary_enabled = true AND EXTRACT(HOUR FROM daily_summary_time) = $1
	`

	rows, err := r.pool.Query(ctx, query, hour)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

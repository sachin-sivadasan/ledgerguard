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
		INSERT INTO notification_preferences (id, user_id, critical_alerts, daily_summary, summary_hour, slack_webhook_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Hour(),
		prefs.SlackWebhookURL,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	return err
}

func (r *PostgresNotificationPreferencesRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, critical_alerts, daily_summary, summary_hour, slack_webhook_url, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	var prefs entity.NotificationPreferences
	var summaryHour int
	var slackURL *string

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.CriticalEnabled,
		&prefs.DailySummaryEnabled,
		&summaryHour,
		&slackURL,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationPreferencesNotFound
		}
		return nil, err
	}

	// Convert hour to time.Time
	prefs.DailySummaryTime = time.Date(0, 1, 1, summaryHour, 0, 0, 0, time.UTC)

	if slackURL != nil {
		prefs.SlackWebhookURL = *slackURL
	}

	return &prefs, nil
}

func (r *PostgresNotificationPreferencesRepository) Update(ctx context.Context, prefs *entity.NotificationPreferences) error {
	query := `
		UPDATE notification_preferences
		SET critical_alerts = $2, daily_summary = $3, summary_hour = $4, slack_webhook_url = $5, updated_at = $6
		WHERE user_id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Hour(),
		prefs.SlackWebhookURL,
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
		INSERT INTO notification_preferences (id, user_id, critical_alerts, daily_summary, summary_hour, slack_webhook_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			critical_alerts = EXCLUDED.critical_alerts,
			daily_summary = EXCLUDED.daily_summary,
			summary_hour = EXCLUDED.summary_hour,
			slack_webhook_url = EXCLUDED.slack_webhook_url,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Hour(),
		prefs.SlackWebhookURL,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	return err
}

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
		INSERT INTO notification_preferences (id, user_id, critical_enabled, daily_summary_enabled, daily_summary_time, slack_webhook_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Format("15:04:05"),
		prefs.SlackWebhookURL,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	return err
}

func (r *PostgresNotificationPreferencesRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, critical_enabled, daily_summary_enabled, daily_summary_time, slack_webhook_url, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	var prefs entity.NotificationPreferences
	var summaryTimeStr string
	var slackURL *string

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.CriticalEnabled,
		&prefs.DailySummaryEnabled,
		&summaryTimeStr,
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

	// Parse time from HH:MM:SS format
	parsedTime, err := parseTimeOfDay(summaryTimeStr)
	if err != nil {
		return nil, err
	}
	prefs.DailySummaryTime = parsedTime

	if slackURL != nil {
		prefs.SlackWebhookURL = *slackURL
	}

	return &prefs, nil
}

func (r *PostgresNotificationPreferencesRepository) Update(ctx context.Context, prefs *entity.NotificationPreferences) error {
	query := `
		UPDATE notification_preferences
		SET critical_enabled = $2, daily_summary_enabled = $3, daily_summary_time = $4, slack_webhook_url = $5, updated_at = $6
		WHERE user_id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Format("15:04:05"),
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
		INSERT INTO notification_preferences (id, user_id, critical_enabled, daily_summary_enabled, daily_summary_time, slack_webhook_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			critical_enabled = EXCLUDED.critical_enabled,
			daily_summary_enabled = EXCLUDED.daily_summary_enabled,
			daily_summary_time = EXCLUDED.daily_summary_time,
			slack_webhook_url = EXCLUDED.slack_webhook_url,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		prefs.ID,
		prefs.UserID,
		prefs.CriticalEnabled,
		prefs.DailySummaryEnabled,
		prefs.DailySummaryTime.Format("15:04:05"),
		prefs.SlackWebhookURL,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	return err
}

// parseTimeOfDay parses a time string in HH:MM:SS format to a time.Time
func parseTimeOfDay(s string) (t time.Time, err error) {
	return time.Parse("15:04:05", s)
}

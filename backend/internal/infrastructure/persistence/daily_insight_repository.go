package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresDailyInsightRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDailyInsightRepository(pool *pgxpool.Pool) *PostgresDailyInsightRepository {
	return &PostgresDailyInsightRepository{pool: pool}
}

func (r *PostgresDailyInsightRepository) Upsert(ctx context.Context, insight *entity.DailyInsight) error {
	query := `
		INSERT INTO daily_insight (id, app_id, date, insight_text, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (app_id, date) DO UPDATE SET
			insight_text = EXCLUDED.insight_text
	`

	_, err := r.pool.Exec(ctx, query,
		insight.ID,
		insight.AppID,
		insight.Date,
		insight.InsightText,
		insight.CreatedAt,
	)

	return err
}

func (r *PostgresDailyInsightRepository) FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyInsight, error) {
	// Truncate to start of day
	truncatedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	query := `
		SELECT id, app_id, date, insight_text, created_at
		FROM daily_insight
		WHERE app_id = $1 AND date = $2
	`

	insight := &entity.DailyInsight{}
	err := r.pool.QueryRow(ctx, query, appID, truncatedDate).Scan(
		&insight.ID,
		&insight.AppID,
		&insight.Date,
		&insight.InsightText,
		&insight.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return insight, nil
}

func (r *PostgresDailyInsightRepository) FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyInsight, error) {
	query := `
		SELECT id, app_id, date, insight_text, created_at
		FROM daily_insight
		WHERE app_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`

	rows, err := r.pool.Query(ctx, query, appID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*entity.DailyInsight
	for rows.Next() {
		insight := &entity.DailyInsight{}
		err := rows.Scan(
			&insight.ID,
			&insight.AppID,
			&insight.Date,
			&insight.InsightText,
			&insight.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		insights = append(insights, insight)
	}

	return insights, rows.Err()
}

func (r *PostgresDailyInsightRepository) FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyInsight, error) {
	query := `
		SELECT id, app_id, date, insight_text, created_at
		FROM daily_insight
		WHERE app_id = $1
		ORDER BY date DESC
		LIMIT 1
	`

	insight := &entity.DailyInsight{}
	err := r.pool.QueryRow(ctx, query, appID).Scan(
		&insight.ID,
		&insight.AppID,
		&insight.Date,
		&insight.InsightText,
		&insight.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return insight, nil
}

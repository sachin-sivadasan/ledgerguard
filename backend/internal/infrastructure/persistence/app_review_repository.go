package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresAppReviewRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAppReviewRepository(pool *pgxpool.Pool) *PostgresAppReviewRepository {
	return &PostgresAppReviewRepository{pool: pool}
}

func (r *PostgresAppReviewRepository) UpsertBatch(ctx context.Context, reviews []*entity.AppReview) error {
	if len(reviews) == 0 {
		return nil
	}

	query := `
		INSERT INTO app_reviews (id, app_id, source_review_id, author, rating, body, review_date, location, time_using, source, scraped_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (app_id, source_review_id) DO UPDATE SET
			rating = EXCLUDED.rating,
			body = EXCLUDED.body,
			location = EXCLUDED.location,
			time_using = EXCLUDED.time_using,
			scraped_at = EXCLUDED.scraped_at,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now().UTC()
	for _, review := range reviews {
		if review.ID == uuid.Nil {
			review.ID = uuid.New()
		}
		if review.CreatedAt.IsZero() {
			review.CreatedAt = now
		}
		review.UpdatedAt = now

		_, err := r.pool.Exec(ctx, query,
			review.ID,
			review.AppID,
			review.SourceReviewID,
			review.Author,
			review.Rating,
			review.Body,
			review.ReviewDate,
			review.Location,
			review.TimeUsing,
			review.Source,
			review.ScrapedAt,
			review.CreatedAt,
			review.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert review %s: %w", review.SourceReviewID, err)
		}
	}

	return nil
}

func (r *PostgresAppReviewRepository) FindByAppID(ctx context.Context, appID uuid.UUID, limit, offset int) ([]*entity.AppReview, error) {
	query := `
		SELECT id, app_id, source_review_id, author, rating, body, review_date,
		       location, time_using, source, scraped_at, created_at, updated_at
		FROM app_reviews
		WHERE app_id = $1
		ORDER BY review_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, appID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*entity.AppReview
	for rows.Next() {
		var review entity.AppReview
		err := rows.Scan(
			&review.ID,
			&review.AppID,
			&review.SourceReviewID,
			&review.Author,
			&review.Rating,
			&review.Body,
			&review.ReviewDate,
			&review.Location,
			&review.TimeUsing,
			&review.Source,
			&review.ScrapedAt,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, rows.Err()
}

func (r *PostgresAppReviewRepository) FindAllByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppReview, error) {
	query := `
		SELECT id, app_id, source_review_id, author, rating, body, review_date,
		       location, time_using, source, scraped_at, created_at, updated_at
		FROM app_reviews
		WHERE app_id = $1
		ORDER BY review_date DESC
	`

	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*entity.AppReview
	for rows.Next() {
		var review entity.AppReview
		err := rows.Scan(
			&review.ID,
			&review.AppID,
			&review.SourceReviewID,
			&review.Author,
			&review.Rating,
			&review.Body,
			&review.ReviewDate,
			&review.Location,
			&review.TimeUsing,
			&review.Source,
			&review.ScrapedAt,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, rows.Err()
}

func (r *PostgresAppReviewRepository) CountByAppID(ctx context.Context, appID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM app_reviews WHERE app_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, query, appID).Scan(&count)
	return count, err
}

package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

var ErrAppNotFound = errors.New("app not found")

type PostgresAppRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAppRepository(pool *pgxpool.Pool) *PostgresAppRepository {
	return &PostgresAppRepository{pool: pool}
}

func (r *PostgresAppRepository) Create(ctx context.Context, app *entity.App) error {
	query := `
		INSERT INTO apps (id, partner_account_id, partner_app_id, name, tracking_enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		app.ID,
		app.PartnerAccountID,
		app.PartnerAppID,
		app.Name,
		app.TrackingEnabled,
		app.CreatedAt,
	)

	return err
}

func (r *PostgresAppRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled, created_at
		FROM apps
		WHERE id = $1
	`

	var app entity.App
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&app.ID,
		&app.PartnerAccountID,
		&app.PartnerAppID,
		&app.Name,
		&app.TrackingEnabled,
		&app.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	return &app, nil
}

func (r *PostgresAppRepository) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled, created_at
		FROM apps
		WHERE partner_account_id = $1
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query, partnerAccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*entity.App
	for rows.Next() {
		var app entity.App
		err := rows.Scan(
			&app.ID,
			&app.PartnerAccountID,
			&app.PartnerAppID,
			&app.Name,
			&app.TrackingEnabled,
			&app.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}

	return apps, rows.Err()
}

func (r *PostgresAppRepository) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled, created_at
		FROM apps
		WHERE partner_account_id = $1 AND partner_app_id = $2
	`

	var app entity.App
	err := r.pool.QueryRow(ctx, query, partnerAccountID, partnerAppID).Scan(
		&app.ID,
		&app.PartnerAccountID,
		&app.PartnerAppID,
		&app.Name,
		&app.TrackingEnabled,
		&app.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	return &app, nil
}

func (r *PostgresAppRepository) Update(ctx context.Context, app *entity.App) error {
	query := `
		UPDATE apps
		SET name = $2, tracking_enabled = $3
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		app.ID,
		app.Name,
		app.TrackingEnabled,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAppNotFound
	}

	return nil
}

func (r *PostgresAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM apps WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAppNotFound
	}

	return nil
}

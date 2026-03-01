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

var ErrAppNotFound = errors.New("app not found")

type PostgresAppRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAppRepository(pool *pgxpool.Pool) *PostgresAppRepository {
	return &PostgresAppRepository{pool: pool}
}

func (r *PostgresAppRepository) Create(ctx context.Context, app *entity.App) error {
	query := `
		INSERT INTO apps (id, partner_account_id, partner_app_id, name, tracking_enabled, revenue_share_tier, install_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		app.ID,
		app.PartnerAccountID,
		app.PartnerAppID,
		app.Name,
		app.TrackingEnabled,
		string(app.RevenueShareTier),
		app.InstallCount,
		app.CreatedAt,
		app.UpdatedAt,
	)

	return err
}

func (r *PostgresAppRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled,
		       COALESCE(revenue_share_tier, 'DEFAULT_20'), COALESCE(install_count, 0),
		       created_at, COALESCE(updated_at, created_at)
		FROM apps
		WHERE id = $1
	`

	var app entity.App
	var tierStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&app.ID,
		&app.PartnerAccountID,
		&app.PartnerAppID,
		&app.Name,
		&app.TrackingEnabled,
		&tierStr,
		&app.InstallCount,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	app.RevenueShareTier = valueobject.ParseRevenueShareTier(tierStr)
	return &app, nil
}

func (r *PostgresAppRepository) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled,
		       COALESCE(revenue_share_tier, 'DEFAULT_20'), COALESCE(install_count, 0),
		       created_at, COALESCE(updated_at, created_at)
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
		var tierStr string
		err := rows.Scan(
			&app.ID,
			&app.PartnerAccountID,
			&app.PartnerAppID,
			&app.Name,
			&app.TrackingEnabled,
			&tierStr,
			&app.InstallCount,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		app.RevenueShareTier = valueobject.ParseRevenueShareTier(tierStr)
		apps = append(apps, &app)
	}

	return apps, rows.Err()
}

func (r *PostgresAppRepository) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled,
		       COALESCE(revenue_share_tier, 'DEFAULT_20'), COALESCE(install_count, 0),
		       created_at, COALESCE(updated_at, created_at)
		FROM apps
		WHERE partner_account_id = $1 AND partner_app_id = $2
	`

	var app entity.App
	var tierStr string
	err := r.pool.QueryRow(ctx, query, partnerAccountID, partnerAppID).Scan(
		&app.ID,
		&app.PartnerAccountID,
		&app.PartnerAppID,
		&app.Name,
		&app.TrackingEnabled,
		&tierStr,
		&app.InstallCount,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	app.RevenueShareTier = valueobject.ParseRevenueShareTier(tierStr)
	return &app, nil
}

func (r *PostgresAppRepository) Update(ctx context.Context, app *entity.App) error {
	query := `
		UPDATE apps
		SET name = $2, tracking_enabled = $3, revenue_share_tier = $4, install_count = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		app.ID,
		app.Name,
		app.TrackingEnabled,
		string(app.RevenueShareTier),
		app.InstallCount,
		app.UpdatedAt,
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

// FindAllByPartnerAppID finds all apps matching a partner app ID across all accounts
// Used for webhook processing where we only have the Shopify app GID
func (r *PostgresAppRepository) FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error) {
	query := `
		SELECT id, partner_account_id, partner_app_id, name, tracking_enabled,
		       COALESCE(revenue_share_tier, 'DEFAULT_20'), COALESCE(install_count, 0),
		       created_at, COALESCE(updated_at, created_at)
		FROM apps
		WHERE partner_app_id = $1
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query, partnerAppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*entity.App
	for rows.Next() {
		var app entity.App
		var tierStr string
		err := rows.Scan(
			&app.ID,
			&app.PartnerAccountID,
			&app.PartnerAppID,
			&app.Name,
			&app.TrackingEnabled,
			&tierStr,
			&app.InstallCount,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		app.RevenueShareTier = valueobject.ParseRevenueShareTier(tierStr)
		apps = append(apps, &app)
	}

	return apps, rows.Err()
}

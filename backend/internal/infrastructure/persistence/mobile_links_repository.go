package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresMobileLinksRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMobileLinksRepository(pool *pgxpool.Pool) *PostgresMobileLinksRepository {
	return &PostgresMobileLinksRepository{pool: pool}
}

// FindByAppID returns the app's links, or a zero-value (not an error) when unset.
func (r *PostgresMobileLinksRepository) FindByAppID(ctx context.Context, appID uuid.UUID) (*entity.MobileLinks, error) {
	links := &entity.MobileLinks{AppID: appID}
	err := r.pool.QueryRow(ctx, `
		SELECT ios_app_id, play_package FROM app_mobile_links WHERE app_id = $1
	`, appID).Scan(&links.IosAppID, &links.PlayPackage)
	if errors.Is(err, pgx.ErrNoRows) {
		return links, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query mobile links: %w", err)
	}
	return links, nil
}

// Upsert stores the links, replacing any existing row.
func (r *PostgresMobileLinksRepository) Upsert(ctx context.Context, links *entity.MobileLinks) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO app_mobile_links (app_id, ios_app_id, play_package, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (app_id) DO UPDATE SET
			ios_app_id = EXCLUDED.ios_app_id,
			play_package = EXCLUDED.play_package,
			updated_at = NOW()
	`, links.AppID, links.IosAppID, links.PlayPackage)
	if err != nil {
		return fmt.Errorf("upsert mobile links: %w", err)
	}
	return nil
}

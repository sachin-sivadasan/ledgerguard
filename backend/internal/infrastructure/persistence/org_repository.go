package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

var ErrOrgNotFound = errors.New("organization not found")

type PostgresOrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrganizationRepository(pool *pgxpool.Pool) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{pool: pool}
}

func (r *PostgresOrganizationRepository) Create(ctx context.Context, org *entity.Organization) error {
	query := `
		INSERT INTO organizations (
			id, name, slug, plan_tier, webhook_url, webhook_secret,
			sso_provider, sso_config, created_by, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		org.ID,
		org.Name,
		nilIfEmpty(org.Slug),
		string(org.PlanTier),
		nilIfEmpty(org.WebhookURL),
		nilIfEmpty(org.WebhookSecret),
		nilIfEmpty(org.SSOProvider),
		nullableJSON(org.SSOConfig),
		org.CreatedBy,
		org.CreatedAt,
		org.UpdatedAt,
	)
	return err
}

func (r *PostgresOrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	query := `
		SELECT id, name, COALESCE(slug, ''), plan_tier,
			COALESCE(webhook_url, ''), COALESCE(webhook_secret, ''),
			COALESCE(sso_provider, ''), sso_config,
			created_by, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	return r.scanOrg(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	query := `
		SELECT id, name, COALESCE(slug, ''), plan_tier,
			COALESCE(webhook_url, ''), COALESCE(webhook_secret, ''),
			COALESCE(sso_provider, ''), sso_config,
			created_by, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`
	return r.scanOrg(r.pool.QueryRow(ctx, query, slug))
}

func (r *PostgresOrganizationRepository) Update(ctx context.Context, org *entity.Organization) error {
	query := `
		UPDATE organizations
		SET name = $2, slug = $3, plan_tier = $4, webhook_url = $5, webhook_secret = $6,
			sso_provider = $7, sso_config = $8, updated_at = $9
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		org.ID,
		org.Name,
		nilIfEmpty(org.Slug),
		string(org.PlanTier),
		nilIfEmpty(org.WebhookURL),
		nilIfEmpty(org.WebhookSecret),
		nilIfEmpty(org.SSOProvider),
		nullableJSON(org.SSOConfig),
		org.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrgNotFound
	}
	return nil
}

func (r *PostgresOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrgNotFound
	}
	return nil
}

func (r *PostgresOrganizationRepository) scanOrg(row pgx.Row) (*entity.Organization, error) {
	var org entity.Organization
	var planTier string
	var ssoConfig []byte

	err := row.Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&planTier,
		&org.WebhookURL,
		&org.WebhookSecret,
		&org.SSOProvider,
		&ssoConfig,
		&org.CreatedBy,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	org.PlanTier = valueobject.PlanTier(planTier)
	if ssoConfig != nil {
		org.SSOConfig = json.RawMessage(ssoConfig)
	}

	return &org, nil
}

// nilIfEmpty returns nil for empty strings, the pointer otherwise.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullableJSON returns nil for nil/empty JSON, the raw bytes otherwise.
func nullableJSON(data json.RawMessage) []byte {
	if len(data) == 0 {
		return nil
	}
	return []byte(data)
}

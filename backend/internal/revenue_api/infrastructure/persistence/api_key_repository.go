package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

// PostgresAPIKeyRepository implements APIKeyRepository using PostgreSQL
type PostgresAPIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAPIKeyRepository creates a new PostgresAPIKeyRepository
func NewPostgresAPIKeyRepository(pool *pgxpool.Pool) *PostgresAPIKeyRepository {
	return &PostgresAPIKeyRepository{pool: pool}
}

// Create creates a new API key
func (r *PostgresAPIKeyRepository) Create(ctx context.Context, key *entity.APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, key_hash, name, rate_limit_per_minute, created_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		key.ID,
		key.UserID,
		key.KeyHash,
		key.Name,
		key.RateLimitPerMinute,
		key.CreatedAt,
		key.RevokedAt,
	)

	return err
}

// GetByHash retrieves an API key by its hash
func (r *PostgresAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, rate_limit_per_minute, created_at, revoked_at
		FROM api_keys
		WHERE key_hash = $1
	`

	return r.scanAPIKey(r.pool.QueryRow(ctx, query, keyHash))
}

// GetByID retrieves an API key by ID
func (r *PostgresAPIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, rate_limit_per_minute, created_at, revoked_at
		FROM api_keys
		WHERE id = $1
	`

	return r.scanAPIKey(r.pool.QueryRow(ctx, query, id))
}

// GetByUserID retrieves all API keys for a user
func (r *PostgresAPIKeyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, rate_limit_per_minute, created_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAPIKeys(rows)
}

// GetActiveByUserID retrieves only active (non-revoked) keys for a user
func (r *PostgresAPIKeyRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, name, rate_limit_per_minute, created_at, revoked_at
		FROM api_keys
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAPIKeys(rows)
}

// Revoke marks an API key as revoked
func (r *PostgresAPIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET revoked_at = $1
		WHERE id = $2 AND revoked_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}

	return nil
}

func (r *PostgresAPIKeyRepository) scanAPIKey(row pgx.Row) (*entity.APIKey, error) {
	var key entity.APIKey
	var name *string

	err := row.Scan(
		&key.ID,
		&key.UserID,
		&key.KeyHash,
		&name,
		&key.RateLimitPerMinute,
		&key.CreatedAt,
		&key.RevokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	if name != nil {
		key.Name = *name
	}

	return &key, nil
}

func (r *PostgresAPIKeyRepository) scanAPIKeys(rows pgx.Rows) ([]*entity.APIKey, error) {
	var keys []*entity.APIKey

	for rows.Next() {
		var key entity.APIKey
		var name *string

		err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.KeyHash,
			&name,
			&key.RateLimitPerMinute,
			&key.CreatedAt,
			&key.RevokedAt,
		)
		if err != nil {
			return nil, err
		}

		if name != nil {
			key.Name = *name
		}
		keys = append(keys, &key)
	}

	return keys, rows.Err()
}

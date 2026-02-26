package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// ErrDeviceTokenNotFound is returned when a device token is not found
var ErrDeviceTokenNotFound = errors.New("device token not found")

type PostgresDeviceTokenRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDeviceTokenRepository(pool *pgxpool.Pool) *PostgresDeviceTokenRepository {
	return &PostgresDeviceTokenRepository{pool: pool}
}

func (r *PostgresDeviceTokenRepository) Create(ctx context.Context, token *entity.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (id, user_id, device_token, platform, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.DeviceToken,
		string(token.Platform),
		token.CreatedAt,
		token.UpdatedAt,
	)

	return err
}

func (r *PostgresDeviceTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.DeviceToken, error) {
	query := `
		SELECT id, user_id, device_token, platform, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*entity.DeviceToken
	for rows.Next() {
		var token entity.DeviceToken
		var platform string

		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.DeviceToken,
			&platform,
			&token.CreatedAt,
			&token.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		token.Platform = entity.Platform(platform)
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *PostgresDeviceTokenRepository) FindByToken(ctx context.Context, deviceToken string) (*entity.DeviceToken, error) {
	query := `
		SELECT id, user_id, device_token, platform, created_at, updated_at
		FROM device_tokens
		WHERE device_token = $1
	`

	var token entity.DeviceToken
	var platform string

	err := r.pool.QueryRow(ctx, query, deviceToken).Scan(
		&token.ID,
		&token.UserID,
		&token.DeviceToken,
		&platform,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeviceTokenNotFound
		}
		return nil, err
	}

	token.Platform = entity.Platform(platform)

	return &token, nil
}

func (r *PostgresDeviceTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM device_tokens WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrDeviceTokenNotFound
	}

	return nil
}

func (r *PostgresDeviceTokenRepository) DeleteByToken(ctx context.Context, deviceToken string) error {
	query := `DELETE FROM device_tokens WHERE device_token = $1`

	result, err := r.pool.Exec(ctx, query, deviceToken)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrDeviceTokenNotFound
	}

	return nil
}

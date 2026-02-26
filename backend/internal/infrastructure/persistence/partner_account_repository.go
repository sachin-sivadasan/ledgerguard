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

var ErrPartnerAccountNotFound = errors.New("partner account not found")

type PostgresPartnerAccountRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPartnerAccountRepository(pool *pgxpool.Pool) *PostgresPartnerAccountRepository {
	return &PostgresPartnerAccountRepository{pool: pool}
}

func (r *PostgresPartnerAccountRepository) Create(ctx context.Context, account *entity.PartnerAccount) error {
	query := `
		INSERT INTO partner_accounts (id, user_id, integration_type, partner_id, encrypted_access_token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.UserID,
		account.IntegrationType.String(),
		account.PartnerID,
		account.EncryptedAccessToken,
		account.CreatedAt,
	)

	return err
}

func (r *PostgresPartnerAccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	query := `
		SELECT id, user_id, integration_type, partner_id, encrypted_access_token, created_at
		FROM partner_accounts
		WHERE id = $1
	`

	var account entity.PartnerAccount
	var integrationType string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&integrationType,
		&account.PartnerID,
		&account.EncryptedAccessToken,
		&account.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPartnerAccountNotFound
		}
		return nil, err
	}

	account.IntegrationType = valueobject.IntegrationType(integrationType)
	return &account, nil
}

func (r *PostgresPartnerAccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	query := `
		SELECT id, user_id, integration_type, partner_id, encrypted_access_token, created_at
		FROM partner_accounts
		WHERE user_id = $1
	`

	var account entity.PartnerAccount
	var integrationType string

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&account.ID,
		&account.UserID,
		&integrationType,
		&account.PartnerID,
		&account.EncryptedAccessToken,
		&account.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPartnerAccountNotFound
		}
		return nil, err
	}

	account.IntegrationType = valueobject.IntegrationType(integrationType)
	return &account, nil
}

func (r *PostgresPartnerAccountRepository) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	query := `
		SELECT id, user_id, integration_type, partner_id, encrypted_access_token, created_at
		FROM partner_accounts
		WHERE partner_id = $1
	`

	var account entity.PartnerAccount
	var integrationType string

	err := r.pool.QueryRow(ctx, query, partnerID).Scan(
		&account.ID,
		&account.UserID,
		&integrationType,
		&account.PartnerID,
		&account.EncryptedAccessToken,
		&account.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPartnerAccountNotFound
		}
		return nil, err
	}

	account.IntegrationType = valueobject.IntegrationType(integrationType)
	return &account, nil
}

func (r *PostgresPartnerAccountRepository) Update(ctx context.Context, account *entity.PartnerAccount) error {
	query := `
		UPDATE partner_accounts
		SET partner_id = $2, encrypted_access_token = $3
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.PartnerID,
		account.EncryptedAccessToken,
	)

	return err
}

func (r *PostgresPartnerAccountRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM partner_accounts WHERE user_id = $1`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrPartnerAccountNotFound
	}

	return nil
}

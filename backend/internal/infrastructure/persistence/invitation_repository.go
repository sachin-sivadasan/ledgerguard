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

var ErrInvitationNotFound = errors.New("invitation not found")

type PostgresInvitationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresInvitationRepository(pool *pgxpool.Pool) *PostgresInvitationRepository {
	return &PostgresInvitationRepository{pool: pool}
}

func (r *PostgresInvitationRepository) Create(ctx context.Context, inv *entity.OrgInvitation) error {
	query := `
		INSERT INTO org_invitations (
			id, org_id, email, role, invited_by, token, status, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		inv.ID,
		inv.OrgID,
		inv.Email,
		string(inv.Role),
		inv.InvitedBy,
		inv.Token,
		string(inv.Status),
		inv.ExpiresAt,
		inv.CreatedAt,
	)
	return err
}

func (r *PostgresInvitationRepository) FindByToken(ctx context.Context, token string) (*entity.OrgInvitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, status, expires_at, created_at
		FROM org_invitations
		WHERE token = $1
	`
	return r.scanInvitation(r.pool.QueryRow(ctx, query, token))
}

func (r *PostgresInvitationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.OrgInvitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, status, expires_at, created_at
		FROM org_invitations
		WHERE id = $1
	`
	return r.scanInvitation(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresInvitationRepository) FindPendingByOrgID(ctx context.Context, orgID uuid.UUID) ([]*entity.OrgInvitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, status, expires_at, created_at
		FROM org_invitations
		WHERE org_id = $1 AND status = 'PENDING'
		ORDER BY created_at DESC
	`
	return r.scanInvitations(ctx, query, orgID)
}

func (r *PostgresInvitationRepository) FindPendingByEmail(ctx context.Context, email string) ([]*entity.OrgInvitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, status, expires_at, created_at
		FROM org_invitations
		WHERE email = $1 AND status = 'PENDING'
		ORDER BY created_at DESC
	`
	return r.scanInvitations(ctx, query, email)
}

func (r *PostgresInvitationRepository) Update(ctx context.Context, inv *entity.OrgInvitation) error {
	query := `
		UPDATE org_invitations
		SET status = $2
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, inv.ID, string(inv.Status))
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

func (r *PostgresInvitationRepository) scanInvitation(row pgx.Row) (*entity.OrgInvitation, error) {
	var inv entity.OrgInvitation
	var role, status string

	err := row.Scan(
		&inv.ID,
		&inv.OrgID,
		&inv.Email,
		&role,
		&inv.InvitedBy,
		&inv.Token,
		&status,
		&inv.ExpiresAt,
		&inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	inv.Role = valueobject.OrgRole(role)
	inv.Status = valueobject.InvitationStatus(status)

	return &inv, nil
}

func (r *PostgresInvitationRepository) scanInvitations(ctx context.Context, query string, args ...any) ([]*entity.OrgInvitation, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*entity.OrgInvitation
	for rows.Next() {
		var inv entity.OrgInvitation
		var role, status string

		err := rows.Scan(
			&inv.ID,
			&inv.OrgID,
			&inv.Email,
			&role,
			&inv.InvitedBy,
			&inv.Token,
			&status,
			&inv.ExpiresAt,
			&inv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		inv.Role = valueobject.OrgRole(role)
		inv.Status = valueobject.InvitationStatus(status)
		invitations = append(invitations, &inv)
	}

	return invitations, rows.Err()
}

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

var ErrMemberNotFound = errors.New("org member not found")

type PostgresMemberRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMemberRepository(pool *pgxpool.Pool) *PostgresMemberRepository {
	return &PostgresMemberRepository{pool: pool}
}

func (r *PostgresMemberRepository) Create(ctx context.Context, member *entity.OrgMember) error {
	query := `
		INSERT INTO org_members (
			id, org_id, user_id, role, status, notification_prefs,
			invited_by, suspended_by, suspended_at, joined_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		member.ID,
		member.OrgID,
		member.UserID,
		string(member.Role),
		string(member.Status),
		nullableJSON(member.NotificationPrefs),
		member.InvitedBy,
		member.SuspendedBy,
		member.SuspendedAt,
		member.JoinedAt,
	)
	return err
}

func (r *PostgresMemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.OrgMember, error) {
	query := `
		SELECT id, org_id, user_id, role, status,
			COALESCE(notification_prefs, '{}'::jsonb),
			invited_by, suspended_by, suspended_at, joined_at
		FROM org_members
		WHERE id = $1
	`
	return r.scanMember(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresMemberRepository) FindByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*entity.OrgMember, error) {
	query := `
		SELECT id, org_id, user_id, role, status,
			COALESCE(notification_prefs, '{}'::jsonb),
			invited_by, suspended_by, suspended_at, joined_at
		FROM org_members
		WHERE org_id = $1 AND user_id = $2
	`
	return r.scanMember(r.pool.QueryRow(ctx, query, orgID, userID))
}

func (r *PostgresMemberRepository) FindByOrgID(ctx context.Context, orgID uuid.UUID) ([]*entity.OrgMember, error) {
	query := `
		SELECT id, org_id, user_id, role, status,
			COALESCE(notification_prefs, '{}'::jsonb),
			invited_by, suspended_by, suspended_at, joined_at
		FROM org_members
		WHERE org_id = $1
		ORDER BY joined_at
	`
	return r.scanMembers(ctx, query, orgID)
}

func (r *PostgresMemberRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.OrgMember, error) {
	query := `
		SELECT id, org_id, user_id, role, status,
			COALESCE(notification_prefs, '{}'::jsonb),
			invited_by, suspended_by, suspended_at, joined_at
		FROM org_members
		WHERE user_id = $1
		ORDER BY joined_at
	`
	return r.scanMembers(ctx, query, userID)
}

func (r *PostgresMemberRepository) CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM org_members WHERE org_id = $1 AND status = 'ACTIVE'`,
		orgID,
	).Scan(&count)
	return count, err
}

func (r *PostgresMemberRepository) Update(ctx context.Context, member *entity.OrgMember) error {
	query := `
		UPDATE org_members
		SET role = $2, status = $3, notification_prefs = $4,
			suspended_by = $5, suspended_at = $6
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		member.ID,
		string(member.Role),
		string(member.Status),
		nullableJSON(member.NotificationPrefs),
		member.SuspendedBy,
		member.SuspendedAt,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *PostgresMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM org_members WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *PostgresMemberRepository) scanMember(row pgx.Row) (*entity.OrgMember, error) {
	var m entity.OrgMember
	var role, status string
	var notifPrefs []byte

	err := row.Scan(
		&m.ID,
		&m.OrgID,
		&m.UserID,
		&role,
		&status,
		&notifPrefs,
		&m.InvitedBy,
		&m.SuspendedBy,
		&m.SuspendedAt,
		&m.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberNotFound
		}
		return nil, err
	}

	m.Role = valueobject.OrgRole(role)
	m.Status = valueobject.MemberStatus(status)
	if notifPrefs != nil {
		m.NotificationPrefs = json.RawMessage(notifPrefs)
	}

	return &m, nil
}

func (r *PostgresMemberRepository) scanMembers(ctx context.Context, query string, args ...any) ([]*entity.OrgMember, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*entity.OrgMember
	for rows.Next() {
		var m entity.OrgMember
		var role, status string
		var notifPrefs []byte

		err := rows.Scan(
			&m.ID,
			&m.OrgID,
			&m.UserID,
			&role,
			&status,
			&notifPrefs,
			&m.InvitedBy,
			&m.SuspendedBy,
			&m.SuspendedAt,
			&m.JoinedAt,
		)
		if err != nil {
			return nil, err
		}

		m.Role = valueobject.OrgRole(role)
		m.Status = valueobject.MemberStatus(status)
		if notifPrefs != nil {
			m.NotificationPrefs = json.RawMessage(notifPrefs)
		}

		members = append(members, &m)
	}

	return members, rows.Err()
}

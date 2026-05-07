package persistence

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresOrgAuditRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrgAuditRepository(pool *pgxpool.Pool) *PostgresOrgAuditRepository {
	return &PostgresOrgAuditRepository{pool: pool}
}

func (r *PostgresOrgAuditRepository) Append(ctx context.Context, entry *entity.OrgAuditEntry) error {
	query := `
		INSERT INTO org_audit_log (
			id, org_id, actor_id, action, target_type, target_id, metadata, ip_address, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::inet, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		entry.ID,
		entry.OrgID,
		entry.ActorID,
		entry.Action,
		nilIfEmpty(entry.TargetType),
		entry.TargetID,
		nullableJSON(entry.Metadata),
		nilIfEmpty(entry.IPAddress),
		entry.CreatedAt,
	)
	return err
}

func (r *PostgresOrgAuditRepository) FindByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*entity.OrgAuditEntry, error) {
	query := `
		SELECT id, org_id, actor_id, action,
			COALESCE(target_type, ''), target_id, metadata,
			COALESCE(host(ip_address), ''), created_at
		FROM org_audit_log
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*entity.OrgAuditEntry
	for rows.Next() {
		var e entity.OrgAuditEntry
		var metadata []byte

		err := rows.Scan(
			&e.ID,
			&e.OrgID,
			&e.ActorID,
			&e.Action,
			&e.TargetType,
			&e.TargetID,
			&metadata,
			&e.IPAddress,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadata != nil {
			e.Metadata = json.RawMessage(metadata)
		}

		entries = append(entries, &e)
	}

	return entries, rows.Err()
}

func (r *PostgresOrgAuditRepository) CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM org_audit_log WHERE org_id = $1`,
		orgID,
	).Scan(&count)
	return count, err
}

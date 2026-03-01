package persistence

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PostgresAuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditLogRepository(pool *pgxpool.Pool) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{pool: pool}
}

func (r *PostgresAuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	query := `
		INSERT INTO audit_log (
			id, user_id, action, resource_type, resource_id,
			details, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	_, err = r.pool.Exec(ctx, query,
		log.ID,
		log.UserID,
		string(log.Action),
		string(log.ResourceType),
		log.ResourceID,
		detailsJSON,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)

	return err
}

func (r *PostgresAuditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id,
		       details, ip_address, user_agent, created_at
		FROM audit_log
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, userID, from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *PostgresAuditLogRepository) FindByResourceID(ctx context.Context, resourceType entity.AuditResourceType, resourceID uuid.UUID) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id,
		       details, ip_address, user_agent, created_at
		FROM audit_log
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, string(resourceType), resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *PostgresAuditLogRepository) FindByAction(ctx context.Context, action entity.AuditAction, from, to time.Time, limit, offset int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id,
		       details, ip_address, user_agent, created_at
		FROM audit_log
		WHERE action = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, string(action), from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *PostgresAuditLogRepository) FindRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id,
		       details, ip_address, user_agent, created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *PostgresAuditLogRepository) Count(ctx context.Context, userID *uuid.UUID, action *entity.AuditAction, from, to time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM audit_log
		WHERE created_at >= $1 AND created_at <= $2
	`
	args := []any{from, to}
	argIdx := 3

	if userID != nil {
		query += ` AND user_id = $` + string(rune('0'+argIdx))
		args = append(args, *userID)
		argIdx++
	}

	if action != nil {
		query += ` AND action = $` + string(rune('0'+argIdx))
		args = append(args, string(*action))
	}

	var count int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

type auditRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func (r *PostgresAuditLogRepository) scanLogs(rows auditRows) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog
	for rows.Next() {
		var log entity.AuditLog
		var action, resourceType string
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&action,
			&resourceType,
			&log.ResourceID,
			&detailsJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		log.Action = entity.AuditAction(action)
		log.ResourceType = entity.AuditResourceType(resourceType)

		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &log.Details)
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

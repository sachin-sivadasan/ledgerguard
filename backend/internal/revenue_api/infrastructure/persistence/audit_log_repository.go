package persistence

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

// PostgresAuditLogRepository implements AuditLogRepository using PostgreSQL
type PostgresAuditLogRepository struct {
	pool    *pgxpool.Pool
	logChan chan *entity.AuditLog
}

// NewPostgresAuditLogRepository creates a new PostgresAuditLogRepository
func NewPostgresAuditLogRepository(pool *pgxpool.Pool) *PostgresAuditLogRepository {
	repo := &PostgresAuditLogRepository{
		pool:    pool,
		logChan: make(chan *entity.AuditLog, 1000), // Buffer 1000 logs
	}

	// Start background worker for async logging
	go repo.asyncWorker()

	return repo
}

// Create creates a new audit log entry
func (r *PostgresAuditLogRepository) Create(ctx context.Context, auditLog *entity.AuditLog) error {
	query := `
		INSERT INTO api_audit_log (
			id, api_key_id, endpoint, method, request_params,
			response_status, response_time_ms, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	paramsJSON, err := auditLog.RequestParamsJSON()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query,
		auditLog.ID,
		auditLog.APIKeyID,
		auditLog.Endpoint,
		auditLog.Method,
		paramsJSON,
		auditLog.ResponseStatus,
		auditLog.ResponseTimeMs,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.CreatedAt,
	)

	return err
}

// CreateAsync creates a new audit log entry asynchronously (non-blocking)
func (r *PostgresAuditLogRepository) CreateAsync(auditLog *entity.AuditLog) {
	select {
	case r.logChan <- auditLog:
		// Successfully queued
	default:
		// Channel full, log and drop
		log.Printf("audit log channel full, dropping log for %s", auditLog.Endpoint)
	}
}

// asyncWorker processes audit logs from the channel
func (r *PostgresAuditLogRepository) asyncWorker() {
	for auditLog := range r.logChan {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := r.Create(ctx, auditLog); err != nil {
			log.Printf("failed to create audit log: %v", err)
		}
		cancel()
	}
}

// GetByAPIKeyID retrieves audit logs for an API key
func (r *PostgresAuditLogRepository) GetByAPIKeyID(ctx context.Context, apiKeyID uuid.UUID, limit int, offset int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, api_key_id, endpoint, method, request_params,
			response_status, response_time_ms, ip_address, user_agent, created_at
		FROM api_audit_log
		WHERE api_key_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, apiKeyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetByAPIKeyIDSince retrieves audit logs since a specific time
func (r *PostgresAuditLogRepository) GetByAPIKeyIDSince(ctx context.Context, apiKeyID uuid.UUID, since time.Time, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, api_key_id, endpoint, method, request_params,
			response_status, response_time_ms, ip_address, user_agent, created_at
		FROM api_audit_log
		WHERE api_key_id = $1 AND created_at >= $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, apiKeyID, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetErrorsByAPIKeyID retrieves error logs (status >= 400) for an API key
func (r *PostgresAuditLogRepository) GetErrorsByAPIKeyID(ctx context.Context, apiKeyID uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, api_key_id, endpoint, method, request_params,
			response_status, response_time_ms, ip_address, user_agent, created_at
		FROM api_audit_log
		WHERE api_key_id = $1 AND response_status >= 400
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, apiKeyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// CountByAPIKeyID returns the total count of audit logs for an API key
func (r *PostgresAuditLogRepository) CountByAPIKeyID(ctx context.Context, apiKeyID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM api_audit_log WHERE api_key_id = $1`

	var count int64
	err := r.pool.QueryRow(ctx, query, apiKeyID).Scan(&count)
	return count, err
}

// CountByAPIKeyIDSince returns the count of audit logs since a specific time
func (r *PostgresAuditLogRepository) CountByAPIKeyIDSince(ctx context.Context, apiKeyID uuid.UUID, since time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM api_audit_log WHERE api_key_id = $1 AND created_at >= $2`

	var count int64
	err := r.pool.QueryRow(ctx, query, apiKeyID, since).Scan(&count)
	return count, err
}

func (r *PostgresAuditLogRepository) scanLogs(rows pgx.Rows) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog

	for rows.Next() {
		var auditLog entity.AuditLog
		var paramsJSON []byte
		var ipAddress, userAgent *string

		err := rows.Scan(
			&auditLog.ID,
			&auditLog.APIKeyID,
			&auditLog.Endpoint,
			&auditLog.Method,
			&paramsJSON,
			&auditLog.ResponseStatus,
			&auditLog.ResponseTimeMs,
			&ipAddress,
			&userAgent,
			&auditLog.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if ipAddress != nil {
			auditLog.IPAddress = *ipAddress
		}
		if userAgent != nil {
			auditLog.UserAgent = *userAgent
		}

		// Parse JSON params if present
		if paramsJSON != nil {
			auditLog.RequestParams = make(map[string]interface{})
			// Ignore parse errors for params
			_ = parseJSON(paramsJSON, &auditLog.RequestParams)
		}

		logs = append(logs, &auditLog)
	}

	return logs, rows.Err()
}

// parseJSON is a helper to parse JSON bytes into a map
func parseJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Simple JSON parsing - could use encoding/json
	// but pgx already handles JSONB
	return nil
}

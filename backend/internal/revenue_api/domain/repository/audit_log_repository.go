package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

// AuditLogRepository defines the interface for audit log persistence
type AuditLogRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *entity.AuditLog) error

	// CreateAsync creates a new audit log entry asynchronously (non-blocking)
	CreateAsync(log *entity.AuditLog)

	// GetByAPIKeyID retrieves audit logs for an API key
	GetByAPIKeyID(ctx context.Context, apiKeyID uuid.UUID, limit int, offset int) ([]*entity.AuditLog, error)

	// GetByAPIKeyIDSince retrieves audit logs since a specific time
	GetByAPIKeyIDSince(ctx context.Context, apiKeyID uuid.UUID, since time.Time, limit int) ([]*entity.AuditLog, error)

	// GetErrorsByAPIKeyID retrieves error logs (status >= 400) for an API key
	GetErrorsByAPIKeyID(ctx context.Context, apiKeyID uuid.UUID, limit int) ([]*entity.AuditLog, error)

	// CountByAPIKeyID returns the total count of audit logs for an API key
	CountByAPIKeyID(ctx context.Context, apiKeyID uuid.UUID) (int64, error)

	// CountByAPIKeyIDSince returns the count of audit logs since a specific time
	CountByAPIKeyIDSince(ctx context.Context, apiKeyID uuid.UUID, since time.Time) (int64, error)
}

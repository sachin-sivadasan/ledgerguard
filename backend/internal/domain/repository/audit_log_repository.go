package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// AuditLogRepository handles persistence of audit log entries
type AuditLogRepository interface {
	// Create stores a new audit log entry
	Create(ctx context.Context, log *entity.AuditLog) error

	// FindByUserID retrieves audit logs for a specific user
	FindByUserID(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset int) ([]*entity.AuditLog, error)

	// FindByResourceID retrieves audit logs for a specific resource
	FindByResourceID(ctx context.Context, resourceType entity.AuditResourceType, resourceID uuid.UUID) ([]*entity.AuditLog, error)

	// FindByAction retrieves audit logs for a specific action type
	FindByAction(ctx context.Context, action entity.AuditAction, from, to time.Time, limit, offset int) ([]*entity.AuditLog, error)

	// FindRecent retrieves the most recent audit logs
	FindRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error)

	// Count returns the total number of audit logs matching criteria
	Count(ctx context.Context, userID *uuid.UUID, action *entity.AuditAction, from, to time.Time) (int64, error)
}

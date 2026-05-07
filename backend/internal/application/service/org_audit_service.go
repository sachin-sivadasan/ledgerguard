package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// OrgAuditService handles creating and querying org-scoped audit log entries.
// This is separate from AuditService which handles user-level audit logs.
type OrgAuditService struct {
	auditRepo repository.OrgAuditRepository
}

func NewOrgAuditService(auditRepo repository.OrgAuditRepository) *OrgAuditService {
	return &OrgAuditService{auditRepo: auditRepo}
}

// LogAction records an action in the org audit log.
func (s *OrgAuditService) LogAction(ctx context.Context, orgID, actorID uuid.UUID, action, targetType string, targetID *uuid.UUID, metadata map[string]interface{}) error {
	var metadataJSON json.RawMessage
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = b
	}

	entry := entity.NewOrgAuditEntry(orgID, actorID, action, targetType, targetID, metadataJSON)
	return s.auditRepo.Append(ctx, entry)
}

// OrgAuditPage represents a paginated org audit log response.
type OrgAuditPage struct {
	Entries []*entity.OrgAuditEntry
	Total   int
	Limit   int
	Offset  int
}

// GetAuditLog returns paginated audit entries for an org.
func (s *OrgAuditService) GetAuditLog(ctx context.Context, orgID uuid.UUID, limit, offset int) (*OrgAuditPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	entries, err := s.auditRepo.FindByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.auditRepo.CountByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &OrgAuditPage{
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

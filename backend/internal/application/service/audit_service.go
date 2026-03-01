package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// AuditService provides audit logging functionality
type AuditService struct {
	repo repository.AuditLogRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo repository.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log creates an audit log entry
func (s *AuditService) Log(
	ctx context.Context,
	userID *uuid.UUID,
	action entity.AuditAction,
	resourceType entity.AuditResourceType,
	resourceID *uuid.UUID,
	details map[string]any,
	ipAddress string,
	userAgent string,
) {
	auditLog := entity.NewAuditLog(
		userID,
		action,
		resourceType,
		resourceID,
		details,
		ipAddress,
		userAgent,
	)

	// Log asynchronously to avoid blocking the main request
	go func() {
		if err := s.repo.Create(context.Background(), auditLog); err != nil {
			log.Printf("Failed to create audit log: %v", err)
		}
	}()
}

// LogSync creates an audit log entry synchronously (for critical actions)
func (s *AuditService) LogSync(
	ctx context.Context,
	userID *uuid.UUID,
	action entity.AuditAction,
	resourceType entity.AuditResourceType,
	resourceID *uuid.UUID,
	details map[string]any,
	ipAddress string,
	userAgent string,
) error {
	auditLog := entity.NewAuditLog(
		userID,
		action,
		resourceType,
		resourceID,
		details,
		ipAddress,
		userAgent,
	)

	return s.repo.Create(ctx, auditLog)
}

// LogLogin logs a user login event
func (s *AuditService) LogLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) {
	s.Log(ctx, &userID, entity.AuditActionLogin, entity.AuditResourceUser, &userID, nil, ipAddress, userAgent)
}

// LogLogout logs a user logout event
func (s *AuditService) LogLogout(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) {
	s.Log(ctx, &userID, entity.AuditActionLogout, entity.AuditResourceUser, &userID, nil, ipAddress, userAgent)
}

// LogIntegrationConnect logs a partner integration connection
func (s *AuditService) LogIntegrationConnect(ctx context.Context, userID uuid.UUID, partnerID string, integrationType string, ipAddress, userAgent string) {
	details := map[string]any{
		"partner_id":       partnerID,
		"integration_type": integrationType,
	}
	s.Log(ctx, &userID, entity.AuditActionIntegrationConnect, entity.AuditResourcePartner, nil, details, ipAddress, userAgent)
}

// LogIntegrationDisconnect logs a partner integration disconnection
func (s *AuditService) LogIntegrationDisconnect(ctx context.Context, userID uuid.UUID, partnerAccountID uuid.UUID, ipAddress, userAgent string) {
	s.Log(ctx, &userID, entity.AuditActionIntegrationDisconnect, entity.AuditResourcePartner, &partnerAccountID, nil, ipAddress, userAgent)
}

// LogAppSelect logs an app selection event
func (s *AuditService) LogAppSelect(ctx context.Context, userID uuid.UUID, appID uuid.UUID, appName string, ipAddress, userAgent string) {
	details := map[string]any{
		"app_name": appName,
	}
	s.Log(ctx, &userID, entity.AuditActionAppSelect, entity.AuditResourceApp, &appID, details, ipAddress, userAgent)
}

// LogTierChange logs a revenue share tier change
func (s *AuditService) LogTierChange(ctx context.Context, userID uuid.UUID, appID uuid.UUID, oldTier, newTier string, ipAddress, userAgent string) {
	details := map[string]any{
		"old_tier": oldTier,
		"new_tier": newTier,
	}
	s.Log(ctx, &userID, entity.AuditActionTierChange, entity.AuditResourceApp, &appID, details, ipAddress, userAgent)
}

// LogSyncStart logs the start of a sync operation
func (s *AuditService) LogSyncStart(ctx context.Context, userID *uuid.UUID, appID uuid.UUID, ipAddress, userAgent string) {
	s.Log(ctx, userID, entity.AuditActionSyncStart, entity.AuditResourceApp, &appID, nil, ipAddress, userAgent)
}

// LogSyncComplete logs the completion of a sync operation
func (s *AuditService) LogSyncComplete(ctx context.Context, userID *uuid.UUID, appID uuid.UUID, transactionCount int, ipAddress, userAgent string) {
	details := map[string]any{
		"transaction_count": transactionCount,
	}
	s.Log(ctx, userID, entity.AuditActionSyncComplete, entity.AuditResourceApp, &appID, details, ipAddress, userAgent)
}

// LogSyncFailed logs a failed sync operation
func (s *AuditService) LogSyncFailed(ctx context.Context, userID *uuid.UUID, appID uuid.UUID, errorMsg string, ipAddress, userAgent string) {
	details := map[string]any{
		"error": errorMsg,
	}
	s.Log(ctx, userID, entity.AuditActionSyncFailed, entity.AuditResourceApp, &appID, details, ipAddress, userAgent)
}

// LogExportRequest logs a data export request
func (s *AuditService) LogExportRequest(ctx context.Context, userID uuid.UUID, exportType string, appID *uuid.UUID, ipAddress, userAgent string) {
	details := map[string]any{
		"export_type": exportType,
	}
	s.Log(ctx, &userID, entity.AuditActionExportRequest, entity.AuditResourceApp, appID, details, ipAddress, userAgent)
}

// LogAPIKeyCreate logs an API key creation
func (s *AuditService) LogAPIKeyCreate(ctx context.Context, userID uuid.UUID, keyID uuid.UUID, keyName string, ipAddress, userAgent string) {
	details := map[string]any{
		"key_name": keyName,
	}
	s.Log(ctx, &userID, entity.AuditActionAPIKeyCreate, entity.AuditResourceAPIKey, &keyID, details, ipAddress, userAgent)
}

// LogAPIKeyRevoke logs an API key revocation
func (s *AuditService) LogAPIKeyRevoke(ctx context.Context, userID uuid.UUID, keyID uuid.UUID, ipAddress, userAgent string) {
	s.Log(ctx, &userID, entity.AuditActionAPIKeyRevoke, entity.AuditResourceAPIKey, &keyID, nil, ipAddress, userAgent)
}

// LogWebhookReceived logs a webhook receipt
func (s *AuditService) LogWebhookReceived(ctx context.Context, topic, shopDomain string) {
	details := map[string]any{
		"topic":       topic,
		"shop_domain": shopDomain,
	}
	s.Log(ctx, nil, entity.AuditActionWebhookReceived, entity.AuditResourceWebhook, nil, details, "", "")
}

// GetUserActivity retrieves audit logs for a user
func (s *AuditService) GetUserActivity(ctx context.Context, userID uuid.UUID, from, to time.Time, limit, offset int) ([]*entity.AuditLog, error) {
	return s.repo.FindByUserID(ctx, userID, from, to, limit, offset)
}

// GetResourceHistory retrieves audit logs for a resource
func (s *AuditService) GetResourceHistory(ctx context.Context, resourceType entity.AuditResourceType, resourceID uuid.UUID) ([]*entity.AuditLog, error) {
	return s.repo.FindByResourceID(ctx, resourceType, resourceID)
}

// GetRecentActivity retrieves the most recent audit logs
func (s *AuditService) GetRecentActivity(ctx context.Context, limit int) ([]*entity.AuditLog, error) {
	return s.repo.FindRecent(ctx, limit)
}

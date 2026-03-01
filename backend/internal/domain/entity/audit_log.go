package entity

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	// Authentication actions
	AuditActionLogin  AuditAction = "LOGIN"
	AuditActionLogout AuditAction = "LOGOUT"

	// Integration actions
	AuditActionIntegrationConnect    AuditAction = "INTEGRATION_CONNECT"
	AuditActionIntegrationDisconnect AuditAction = "INTEGRATION_DISCONNECT"
	AuditActionTokenAdd              AuditAction = "TOKEN_ADD"
	AuditActionTokenRevoke           AuditAction = "TOKEN_REVOKE"

	// App actions
	AuditActionAppSelect   AuditAction = "APP_SELECT"
	AuditActionAppDeselect AuditAction = "APP_DESELECT"
	AuditActionAppUpdate   AuditAction = "APP_UPDATE"
	AuditActionTierChange  AuditAction = "TIER_CHANGE"

	// Sync actions
	AuditActionSyncStart    AuditAction = "SYNC_START"
	AuditActionSyncComplete AuditAction = "SYNC_COMPLETE"
	AuditActionSyncFailed   AuditAction = "SYNC_FAILED"

	// Export actions
	AuditActionExportRequest AuditAction = "EXPORT_REQUEST"
	AuditActionExportComplete AuditAction = "EXPORT_COMPLETE"

	// Settings actions
	AuditActionSettingsUpdate      AuditAction = "SETTINGS_UPDATE"
	AuditActionPreferencesUpdate   AuditAction = "PREFERENCES_UPDATE"
	AuditActionNotificationUpdate  AuditAction = "NOTIFICATION_UPDATE"

	// API key actions
	AuditActionAPIKeyCreate AuditAction = "API_KEY_CREATE"
	AuditActionAPIKeyRevoke AuditAction = "API_KEY_REVOKE"

	// Webhook actions
	AuditActionWebhookReceived AuditAction = "WEBHOOK_RECEIVED"
	AuditActionWebhookProcessed AuditAction = "WEBHOOK_PROCESSED"
)

// AuditResourceType represents the type of resource being acted upon
type AuditResourceType string

const (
	AuditResourceUser         AuditResourceType = "USER"
	AuditResourcePartner      AuditResourceType = "PARTNER_ACCOUNT"
	AuditResourceApp          AuditResourceType = "APP"
	AuditResourceSubscription AuditResourceType = "SUBSCRIPTION"
	AuditResourceTransaction  AuditResourceType = "TRANSACTION"
	AuditResourceAPIKey       AuditResourceType = "API_KEY"
	AuditResourceSettings     AuditResourceType = "SETTINGS"
	AuditResourceWebhook      AuditResourceType = "WEBHOOK"
)

// AuditLog represents an audit trail entry for compliance and debugging
type AuditLog struct {
	ID           uuid.UUID
	UserID       *uuid.UUID        // Nullable for system-initiated actions
	Action       AuditAction       // What was done
	ResourceType AuditResourceType // What type of resource
	ResourceID   *uuid.UUID        // ID of the affected resource (nullable)
	Details      map[string]any    // Additional context as JSON
	IPAddress    string            // Client IP address
	UserAgent    string            // Client user agent
	CreatedAt    time.Time
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(
	userID *uuid.UUID,
	action AuditAction,
	resourceType AuditResourceType,
	resourceID *uuid.UUID,
	details map[string]any,
	ipAddress string,
	userAgent string,
) *AuditLog {
	return &AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now().UTC(),
	}
}

// IsUserAction returns true if this action was initiated by a user
func (a *AuditLog) IsUserAction() bool {
	return a.UserID != nil
}

// IsSystemAction returns true if this action was initiated by the system
func (a *AuditLog) IsSystemAction() bool {
	return a.UserID == nil
}

package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Organization is the top-level data-owning entity. All partner accounts,
// billing, and team members are scoped to an organization.
type Organization struct {
	ID            uuid.UUID
	Name          string
	Slug          string
	PlanTier      valueobject.PlanTier
	WebhookURL    string
	WebhookSecret string
	SSOProvider   string
	SSOConfig     json.RawMessage
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewOrganization(name string, createdBy uuid.UUID) *Organization {
	now := time.Now().UTC()
	return &Organization{
		ID:        uuid.New(),
		Name:      name,
		PlanTier:  valueobject.PlanTierFree,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MaxMembers returns the team size limit based on the org's plan tier.
func (o *Organization) MaxMembers() int {
	switch o.PlanTier {
	case valueobject.PlanTierFree:
		return 1
	case valueobject.PlanTierStarter:
		return 3
	case valueobject.PlanTierPro:
		return 10
	default:
		return 1
	}
}

// CanUseSSO returns true if the org's plan supports SSO/SAML.
func (o *Organization) CanUseSSO() bool {
	return o.PlanTier == valueobject.PlanTierPro
}

// CanUseWebhooks returns true if the org's plan supports org-level webhooks.
func (o *Organization) CanUseWebhooks() bool {
	return o.PlanTier == valueobject.PlanTierStarter || o.PlanTier == valueobject.PlanTierPro
}

// MaxApps returns the app limit for the org's plan tier. 0 means unlimited.
func (o *Organization) MaxApps() int {
	switch o.PlanTier {
	case valueobject.PlanTierFree, valueobject.PlanTierStarter:
		return 1
	case valueobject.PlanTierPro:
		return 0 // unlimited
	default:
		return 1
	}
}

// OrgMember represents a user's membership in an organization.
type OrgMember struct {
	ID                uuid.UUID
	OrgID             uuid.UUID
	UserID            uuid.UUID
	Role              valueobject.OrgRole
	Status            valueobject.MemberStatus
	NotificationPrefs json.RawMessage
	InvitedBy         *uuid.UUID
	SuspendedBy       *uuid.UUID
	SuspendedAt       *time.Time
	JoinedAt          time.Time
}

func NewOrgMember(orgID, userID uuid.UUID, role valueobject.OrgRole, invitedBy *uuid.UUID) *OrgMember {
	return &OrgMember{
		ID:                uuid.New(),
		OrgID:             orgID,
		UserID:            userID,
		Role:              role,
		Status:            valueobject.MemberStatusActive,
		NotificationPrefs: json.RawMessage(`{"email": true, "push": true}`),
		InvitedBy:         invitedBy,
		JoinedAt:          time.Now().UTC(),
	}
}

// IsActive returns true if the member can access org data.
func (m *OrgMember) IsActive() bool {
	return m.Status.IsActive()
}

// Suspend marks the member as suspended.
func (m *OrgMember) Suspend(suspendedBy uuid.UUID) {
	m.Status = valueobject.MemberStatusSuspended
	m.SuspendedBy = &suspendedBy
	now := time.Now().UTC()
	m.SuspendedAt = &now
}

// Unsuspend restores the member to active status.
func (m *OrgMember) Unsuspend() {
	m.Status = valueobject.MemberStatusActive
	m.SuspendedBy = nil
	m.SuspendedAt = nil
}

// OrgInvitation represents a pending invitation to join an organization.
type OrgInvitation struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Email     string
	Role      valueobject.OrgRole
	InvitedBy uuid.UUID
	Token     string
	Status    valueobject.InvitationStatus
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewOrgInvitation(orgID uuid.UUID, email string, role valueobject.OrgRole, invitedBy uuid.UUID, token string, expiresAt time.Time) *OrgInvitation {
	return &OrgInvitation{
		ID:        uuid.New(),
		OrgID:     orgID,
		Email:     email,
		Role:      role,
		InvitedBy: invitedBy,
		Token:     token,
		Status:    valueobject.InvitationStatusPending,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
}

// IsExpired returns true if the invitation has passed its expiry time.
func (i *OrgInvitation) IsExpired(now time.Time) bool {
	return now.After(i.ExpiresAt)
}

// OrgAuditEntry records a single action performed within an organization.
type OrgAuditEntry struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	ActorID    uuid.UUID
	Action     string
	TargetType string
	TargetID   *uuid.UUID
	Metadata   json.RawMessage
	IPAddress  string
	CreatedAt  time.Time
}

func NewOrgAuditEntry(orgID, actorID uuid.UUID, action, targetType string, targetID *uuid.UUID, metadata json.RawMessage) *OrgAuditEntry {
	return &OrgAuditEntry{
		ID:         uuid.New(),
		OrgID:      orgID,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
		CreatedAt:  time.Now().UTC(),
	}
}

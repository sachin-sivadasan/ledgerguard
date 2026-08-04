package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

var (
	ErrMemberLimitReached   = errors.New("organization member limit reached for current plan")
	ErrAlreadyMember        = errors.New("user is already a member of this organization")
	ErrCannotRemoveOwner    = errors.New("cannot remove the organization owner")
	ErrCannotSuspendOwner   = errors.New("cannot suspend the organization owner")
	ErrInsufficientRole     = errors.New("insufficient role for this action")
	ErrInvitationExpired    = errors.New("invitation has expired")
	ErrInvitationNotPending = errors.New("invitation is not in pending state")
	ErrMemberSuspended      = errors.New("member account is suspended")
)

// OrgService handles organization management: create, invite, accept, suspend, remove.
type OrgService struct {
	orgRepo        repository.OrganizationRepository
	memberRepo     repository.MemberRepository
	invitationRepo repository.InvitationRepository
	auditService   *OrgAuditService
}

func NewOrgService(
	orgRepo repository.OrganizationRepository,
	memberRepo repository.MemberRepository,
	invitationRepo repository.InvitationRepository,
	auditService *OrgAuditService,
) *OrgService {
	return &OrgService{
		orgRepo:        orgRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
		auditService:   auditService,
	}
}

// CreateOrganization creates a new org and adds the creator as OWNER.
func (s *OrgService) CreateOrganization(ctx context.Context, name string, creatorID uuid.UUID) (*entity.Organization, error) {
	org := entity.NewOrganization(name, creatorID)
	// Suffix the slug with a short unique fragment so orgs whose names slugify
	// identically (e.g. two invalid-email signups → "my-organization", or
	// alice@a.com / alice@b.io) don't collide on the UNIQUE slug column. Slug is
	// not used for routing (orgs resolve by ID), so the suffix is cosmetic-only.
	// Cap the base so base(≤91) + "-" + 8-char fragment fits slug's VARCHAR(100).
	base := generateSlug(name)
	if len(base) > 91 {
		base = base[:91]
	}
	org.Slug = base + "-" + org.ID.String()[:8]

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	member := entity.NewOrgMember(org.ID, creatorID, valueobject.OrgRoleOwner, nil)
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	_ = s.auditService.LogAction(ctx, org.ID, creatorID, "org.created", "org", &org.ID, nil)

	return org, nil
}

// GetOrganization returns an org by ID.
func (s *OrgService) GetOrganization(ctx context.Context, orgID uuid.UUID) (*entity.Organization, error) {
	return s.orgRepo.FindByID(ctx, orgID)
}

// ListUserOrganizations returns all orgs a user belongs to.
func (s *OrgService) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*entity.OrgMember, error) {
	return s.memberRepo.FindByUserID(ctx, userID)
}

// UpdateOrganization updates org name/settings. Only OWNER can do this.
func (s *OrgService) UpdateOrganization(ctx context.Context, org *entity.Organization, actorID uuid.UUID) error {
	org.UpdatedAt = time.Now().UTC()
	if err := s.orgRepo.Update(ctx, org); err != nil {
		return err
	}
	_ = s.auditService.LogAction(ctx, org.ID, actorID, "org.updated", "org", &org.ID, nil)
	return nil
}

// DeleteOrganization deletes an org. Only OWNER can do this. CASCADE handles cleanup.
func (s *OrgService) DeleteOrganization(ctx context.Context, orgID, actorID uuid.UUID) error {
	_ = s.auditService.LogAction(ctx, orgID, actorID, "org.deleted", "org", &orgID, nil)
	return s.orgRepo.Delete(ctx, orgID)
}

// ListMembers returns all members of an org.
func (s *OrgService) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*entity.OrgMember, error) {
	return s.memberRepo.FindByOrgID(ctx, orgID)
}

// InviteMember creates an invitation for a new member.
// Checks plan limits and duplicate membership.
func (s *OrgService) InviteMember(ctx context.Context, orgID uuid.UUID, email string, role valueobject.OrgRole, invitedBy uuid.UUID) (*entity.OrgInvitation, error) {
	// Check plan limits
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	count, err := s.memberRepo.CountByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Count pending invitations toward the limit
	pending, err := s.invitationRepo.FindPendingByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if count+len(pending) >= org.MaxMembers() {
		return nil, ErrMemberLimitReached
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour) // 7 days
	invitation := entity.NewOrgInvitation(orgID, email, role, invitedBy, token, expiresAt)

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	_ = s.auditService.LogAction(ctx, orgID, invitedBy, "member.invited", "invitation", &invitation.ID, map[string]interface{}{
		"email": email,
		"role":  string(role),
	})

	return invitation, nil
}

// AcceptInvitation processes an invitation token, creating a new member.
func (s *OrgService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*entity.OrgMember, error) {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !invitation.Status.IsPending() {
		return nil, ErrInvitationNotPending
	}

	if invitation.IsExpired(time.Now().UTC()) {
		invitation.Status = valueobject.InvitationStatusExpired
		_ = s.invitationRepo.Update(ctx, invitation)
		return nil, ErrInvitationExpired
	}

	// Check if user is already a member
	existing, err := s.memberRepo.FindByOrgAndUser(ctx, invitation.OrgID, userID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyMember
	}

	// Create member
	invitedBy := invitation.InvitedBy
	member := entity.NewOrgMember(invitation.OrgID, userID, invitation.Role, &invitedBy)
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	// Mark invitation as accepted
	invitation.Status = valueobject.InvitationStatusAccepted
	_ = s.invitationRepo.Update(ctx, invitation)

	_ = s.auditService.LogAction(ctx, invitation.OrgID, userID, "member.joined", "member", &member.ID, map[string]interface{}{
		"role": string(invitation.Role),
	})

	return member, nil
}

// RevokeInvitation cancels a pending invitation.
func (s *OrgService) RevokeInvitation(ctx context.Context, invitationID, actorID uuid.UUID) error {
	invitation, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		return err
	}

	if !invitation.Status.IsPending() {
		return ErrInvitationNotPending
	}

	invitation.Status = valueobject.InvitationStatusRevoked
	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		return err
	}

	_ = s.auditService.LogAction(ctx, invitation.OrgID, actorID, "invitation.revoked", "invitation", &invitationID, nil)
	return nil
}

// RemoveMember removes a member from the org. Cannot remove OWNER.
func (s *OrgService) RemoveMember(ctx context.Context, orgID uuid.UUID, targetMemberID, actorID uuid.UUID) error {
	member, err := s.memberRepo.FindByID(ctx, targetMemberID)
	if err != nil {
		return err
	}

	if member.Role == valueobject.OrgRoleOwner {
		return ErrCannotRemoveOwner
	}

	if err := s.memberRepo.Delete(ctx, targetMemberID); err != nil {
		return err
	}

	_ = s.auditService.LogAction(ctx, orgID, actorID, "member.removed", "member", &targetMemberID, map[string]interface{}{
		"user_id": member.UserID.String(),
		"role":    string(member.Role),
	})

	return nil
}

// SuspendMember suspends a member. Cannot suspend OWNER.
func (s *OrgService) SuspendMember(ctx context.Context, orgID uuid.UUID, targetMemberID, actorID uuid.UUID) error {
	member, err := s.memberRepo.FindByID(ctx, targetMemberID)
	if err != nil {
		return err
	}

	if member.Role == valueobject.OrgRoleOwner {
		return ErrCannotSuspendOwner
	}

	member.Suspend(actorID)
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	_ = s.auditService.LogAction(ctx, orgID, actorID, "member.suspended", "member", &targetMemberID, nil)
	return nil
}

// UnsuspendMember restores a suspended member to active status.
func (s *OrgService) UnsuspendMember(ctx context.Context, orgID uuid.UUID, targetMemberID, actorID uuid.UUID) error {
	member, err := s.memberRepo.FindByID(ctx, targetMemberID)
	if err != nil {
		return err
	}

	member.Unsuspend()
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	_ = s.auditService.LogAction(ctx, orgID, actorID, "member.unsuspended", "member", &targetMemberID, nil)
	return nil
}

// ChangeRole updates a member's role. Only OWNER can change roles.
func (s *OrgService) ChangeRole(ctx context.Context, orgID uuid.UUID, targetMemberID, actorID uuid.UUID, newRole valueobject.OrgRole) error {
	member, err := s.memberRepo.FindByID(ctx, targetMemberID)
	if err != nil {
		return err
	}

	if member.Role == valueobject.OrgRoleOwner {
		return ErrCannotRemoveOwner // Cannot demote owner
	}

	oldRole := member.Role
	member.Role = newRole
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	_ = s.auditService.LogAction(ctx, orgID, actorID, "role.changed", "member", &targetMemberID, map[string]interface{}{
		"old_role": string(oldRole),
		"new_role": string(newRole),
	})

	return nil
}

// UpdateNotificationPrefs updates a member's notification preferences.
func (s *OrgService) UpdateNotificationPrefs(ctx context.Context, memberID uuid.UUID, prefs []byte) error {
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		return err
	}
	member.NotificationPrefs = prefs
	return s.memberRepo.Update(ctx, member)
}

// ConfigureWebhook sets the org webhook URL and generates a secret.
func (s *OrgService) ConfigureWebhook(ctx context.Context, orgID uuid.UUID, webhookURL string, actorID uuid.UUID) error {
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return err
	}

	if !org.CanUseWebhooks() {
		return ErrMemberLimitReached // Plan doesn't support webhooks
	}

	org.WebhookURL = webhookURL
	if org.WebhookSecret == "" {
		secret, err := generateToken()
		if err != nil {
			return err
		}
		org.WebhookSecret = secret
	}

	org.UpdatedAt = time.Now().UTC()
	if err := s.orgRepo.Update(ctx, org); err != nil {
		return err
	}

	_ = s.auditService.LogAction(ctx, orgID, actorID, "webhook.configured", "org", &orgID, nil)
	return nil
}

// generateSlug creates a URL-safe slug from a name.
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// generateToken creates a cryptographically secure random token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

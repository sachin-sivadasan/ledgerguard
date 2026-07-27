package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// --- Mock Repositories ---

type mockOrgRepo struct {
	orgs map[uuid.UUID]*entity.Organization
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{orgs: make(map[uuid.UUID]*entity.Organization)}
}

func (r *mockOrgRepo) Create(_ context.Context, org *entity.Organization) error {
	r.orgs[org.ID] = org
	return nil
}

func (r *mockOrgRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.Organization, error) {
	org, ok := r.orgs[id]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return org, nil
}

func (r *mockOrgRepo) FindBySlug(_ context.Context, slug string) (*entity.Organization, error) {
	for _, org := range r.orgs {
		if org.Slug == slug {
			return org, nil
		}
	}
	return nil, errors.New("organization not found")
}

func (r *mockOrgRepo) Update(_ context.Context, org *entity.Organization) error {
	if _, ok := r.orgs[org.ID]; !ok {
		return errors.New("organization not found")
	}
	r.orgs[org.ID] = org
	return nil
}

func (r *mockOrgRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.orgs[id]; !ok {
		return errors.New("organization not found")
	}
	delete(r.orgs, id)
	return nil
}

type mockMemberRepo struct {
	members map[uuid.UUID]*entity.OrgMember
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[uuid.UUID]*entity.OrgMember)}
}

func (r *mockMemberRepo) Create(_ context.Context, member *entity.OrgMember) error {
	r.members[member.ID] = member
	return nil
}

func (r *mockMemberRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.OrgMember, error) {
	m, ok := r.members[id]
	if !ok {
		return nil, errors.New("org member not found")
	}
	return m, nil
}

func (r *mockMemberRepo) FindByOrgAndUser(_ context.Context, orgID, userID uuid.UUID) (*entity.OrgMember, error) {
	for _, m := range r.members {
		if m.OrgID == orgID && m.UserID == userID {
			return m, nil
		}
	}
	return nil, errors.New("org member not found")
}

func (r *mockMemberRepo) FindByOrgID(_ context.Context, orgID uuid.UUID) ([]*entity.OrgMember, error) {
	var result []*entity.OrgMember
	for _, m := range r.members {
		if m.OrgID == orgID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *mockMemberRepo) FindByUserID(_ context.Context, userID uuid.UUID) ([]*entity.OrgMember, error) {
	var result []*entity.OrgMember
	for _, m := range r.members {
		if m.UserID == userID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *mockMemberRepo) CountByOrgID(_ context.Context, orgID uuid.UUID) (int, error) {
	count := 0
	for _, m := range r.members {
		if m.OrgID == orgID && m.Status == valueobject.MemberStatusActive {
			count++
		}
	}
	return count, nil
}

func (r *mockMemberRepo) Update(_ context.Context, member *entity.OrgMember) error {
	if _, ok := r.members[member.ID]; !ok {
		return errors.New("org member not found")
	}
	r.members[member.ID] = member
	return nil
}

func (r *mockMemberRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.members[id]; !ok {
		return errors.New("org member not found")
	}
	delete(r.members, id)
	return nil
}

type mockInvitationRepo struct {
	invitations map[uuid.UUID]*entity.OrgInvitation
}

func newMockInvitationRepo() *mockInvitationRepo {
	return &mockInvitationRepo{invitations: make(map[uuid.UUID]*entity.OrgInvitation)}
}

func (r *mockInvitationRepo) Create(_ context.Context, inv *entity.OrgInvitation) error {
	r.invitations[inv.ID] = inv
	return nil
}

func (r *mockInvitationRepo) FindByToken(_ context.Context, token string) (*entity.OrgInvitation, error) {
	for _, inv := range r.invitations {
		if inv.Token == token {
			return inv, nil
		}
	}
	return nil, errors.New("invitation not found")
}

func (r *mockInvitationRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.OrgInvitation, error) {
	inv, ok := r.invitations[id]
	if !ok {
		return nil, errors.New("invitation not found")
	}
	return inv, nil
}

func (r *mockInvitationRepo) FindPendingByOrgID(_ context.Context, orgID uuid.UUID) ([]*entity.OrgInvitation, error) {
	var result []*entity.OrgInvitation
	for _, inv := range r.invitations {
		if inv.OrgID == orgID && inv.Status == valueobject.InvitationStatusPending {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *mockInvitationRepo) FindPendingByEmail(_ context.Context, email string) ([]*entity.OrgInvitation, error) {
	var result []*entity.OrgInvitation
	for _, inv := range r.invitations {
		if inv.Email == email && inv.Status == valueobject.InvitationStatusPending {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *mockInvitationRepo) Update(_ context.Context, inv *entity.OrgInvitation) error {
	if _, ok := r.invitations[inv.ID]; !ok {
		return errors.New("invitation not found")
	}
	r.invitations[inv.ID] = inv
	return nil
}

type mockOrgAuditRepo struct {
	entries []*entity.OrgAuditEntry
}

func newMockOrgAuditRepo() *mockOrgAuditRepo {
	return &mockOrgAuditRepo{}
}

func (r *mockOrgAuditRepo) Append(_ context.Context, entry *entity.OrgAuditEntry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func (r *mockOrgAuditRepo) FindByOrgID(_ context.Context, orgID uuid.UUID, limit, offset int) ([]*entity.OrgAuditEntry, error) {
	var result []*entity.OrgAuditEntry
	for _, e := range r.entries {
		if e.OrgID == orgID {
			result = append(result, e)
		}
	}
	if offset > len(result) {
		return nil, nil
	}
	result = result[offset:]
	if limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

func (r *mockOrgAuditRepo) CountByOrgID(_ context.Context, orgID uuid.UUID) (int, error) {
	count := 0
	for _, e := range r.entries {
		if e.OrgID == orgID {
			count++
		}
	}
	return count, nil
}

// --- Helper ---

func setupOrgService() (*OrgService, *mockOrgRepo, *mockMemberRepo, *mockInvitationRepo, *mockOrgAuditRepo) {
	orgRepo := newMockOrgRepo()
	memberRepo := newMockMemberRepo()
	invRepo := newMockInvitationRepo()
	auditRepo := newMockOrgAuditRepo()
	auditSvc := NewOrgAuditService(auditRepo)
	svc := NewOrgService(orgRepo, memberRepo, invRepo, auditSvc)
	return svc, orgRepo, memberRepo, invRepo, auditRepo
}

// --- Tests ---

func TestCreateOrganization(t *testing.T) {
	svc, orgRepo, memberRepo, _, auditRepo := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, err := svc.CreateOrganization(ctx, "My Org", creatorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if org.Name != "My Org" {
		t.Errorf("expected name 'My Org', got %q", org.Name)
	}
	// Slug is the generated base plus a short unique suffix (org ID fragment) to
	// avoid UNIQUE-slug collisions across same-named orgs.
	if !strings.HasPrefix(org.Slug, "my-org-") || len(org.Slug) <= len("my-org-") {
		t.Errorf("expected slug to start with 'my-org-' plus a unique suffix, got %q", org.Slug)
	}
	if org.PlanTier != valueobject.PlanTierFree {
		t.Errorf("expected FREE plan, got %s", org.PlanTier)
	}
	if org.CreatedBy != creatorID {
		t.Errorf("expected creator %s, got %s", creatorID, org.CreatedBy)
	}

	// Check org was persisted
	if _, ok := orgRepo.orgs[org.ID]; !ok {
		t.Error("org not persisted")
	}

	// Check OWNER member was created
	found := false
	for _, m := range memberRepo.members {
		if m.OrgID == org.ID && m.UserID == creatorID && m.Role == valueobject.OrgRoleOwner {
			found = true
		}
	}
	if !found {
		t.Error("OWNER member not created")
	}

	// Check audit log
	if len(auditRepo.entries) != 1 {
		t.Errorf("expected 1 audit entry, got %d", len(auditRepo.entries))
	}
	if auditRepo.entries[0].Action != "org.created" {
		t.Errorf("expected action 'org.created', got %q", auditRepo.entries[0].Action)
	}
}

// TestCreateOrganization_UniqueSlugPerSameName is the regression guard for the slug
// collision: two orgs with the same name must get distinct (both non-colliding) slugs.
func TestCreateOrganization_UniqueSlugPerSameName(t *testing.T) {
	svc, _, _, _, _ := setupOrgService()
	ctx := context.Background()

	org1, err := svc.CreateOrganization(ctx, "My Org", uuid.New())
	if err != nil {
		t.Fatalf("org1: %v", err)
	}
	org2, err := svc.CreateOrganization(ctx, "My Org", uuid.New())
	if err != nil {
		t.Fatalf("org2: %v", err)
	}
	if org1.Slug == org2.Slug {
		t.Errorf("expected distinct slugs for same-named orgs, both got %q", org1.Slug)
	}
}

// TestCreateOrganization_SlugFitsColumn verifies a very long name still yields a slug
// within the slug VARCHAR(100) column limit.
func TestCreateOrganization_SlugFitsColumn(t *testing.T) {
	svc, _, _, _, _ := setupOrgService()
	longName := strings.Repeat("a", 300)
	org, err := svc.CreateOrganization(context.Background(), longName, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(org.Slug) > 100 {
		t.Errorf("slug exceeds VARCHAR(100): len=%d", len(org.Slug))
	}
}

func TestInviteMember_Success(t *testing.T) {
	svc, _, _, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	// Create org (STARTER allows 3 members)
	org, _ := svc.CreateOrganization(ctx, "Test Org", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	// Invite a viewer
	inv, err := svc.InviteMember(ctx, org.ID, "viewer@example.com", valueobject.OrgRoleViewer, creatorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Email != "viewer@example.com" {
		t.Errorf("expected email 'viewer@example.com', got %q", inv.Email)
	}
	if inv.Role != valueobject.OrgRoleViewer {
		t.Errorf("expected VIEWER role, got %s", inv.Role)
	}
	if inv.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestInviteMember_FreePlanLimit(t *testing.T) {
	svc, _, _, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	// Create org on FREE plan (max 1 member = owner only)
	org, _ := svc.CreateOrganization(ctx, "Free Org", creatorID)

	// Try to invite — should fail
	_, err := svc.InviteMember(ctx, org.ID, "someone@example.com", valueobject.OrgRoleViewer, creatorID)
	if !errors.Is(err, ErrMemberLimitReached) {
		t.Errorf("expected ErrMemberLimitReached, got %v", err)
	}
}

func TestAcceptInvitation_Success(t *testing.T) {
	svc, _, memberRepo, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()
	inviteeID := uuid.New()

	// Create org on STARTER plan
	org, _ := svc.CreateOrganization(ctx, "Test Org", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	// Invite
	inv, _ := svc.InviteMember(ctx, org.ID, "invitee@example.com", valueobject.OrgRoleViewer, creatorID)

	// Accept
	member, err := svc.AcceptInvitation(ctx, inv.Token, inviteeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.Role != valueobject.OrgRoleViewer {
		t.Errorf("expected VIEWER role, got %s", member.Role)
	}
	if member.OrgID != org.ID {
		t.Errorf("expected org %s, got %s", org.ID, member.OrgID)
	}

	// Verify member was persisted
	if _, ok := memberRepo.members[member.ID]; !ok {
		t.Error("member not persisted")
	}
}

func TestAcceptInvitation_Expired(t *testing.T) {
	svc, _, _, invRepo, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	// Create org on STARTER
	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	// Create invitation that already expired
	inv, _ := svc.InviteMember(ctx, org.ID, "expired@example.com", valueobject.OrgRoleViewer, creatorID)
	inv.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)
	invRepo.invitations[inv.ID] = inv

	// Accept should fail
	_, err := svc.AcceptInvitation(ctx, inv.Token, uuid.New())
	if !errors.Is(err, ErrInvitationExpired) {
		t.Errorf("expected ErrInvitationExpired, got %v", err)
	}
}

func TestAcceptInvitation_AlreadyMember(t *testing.T) {
	svc, _, _, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	inv, _ := svc.InviteMember(ctx, org.ID, "dupe@example.com", valueobject.OrgRoleViewer, creatorID)

	// Accept with the creator's own ID (already an OWNER member)
	_, err := svc.AcceptInvitation(ctx, inv.Token, creatorID)
	if !errors.Is(err, ErrAlreadyMember) {
		t.Errorf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	svc, _, memberRepo, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)

	// Find the OWNER member
	var ownerMemberID uuid.UUID
	for _, m := range memberRepo.members {
		if m.OrgID == org.ID && m.Role == valueobject.OrgRoleOwner {
			ownerMemberID = m.ID
			break
		}
	}

	err := svc.RemoveMember(ctx, org.ID, ownerMemberID, creatorID)
	if !errors.Is(err, ErrCannotRemoveOwner) {
		t.Errorf("expected ErrCannotRemoveOwner, got %v", err)
	}
}

func TestSuspendMember_Success(t *testing.T) {
	svc, _, memberRepo, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()
	viewerID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	inv, _ := svc.InviteMember(ctx, org.ID, "viewer@test.com", valueobject.OrgRoleViewer, creatorID)
	member, _ := svc.AcceptInvitation(ctx, inv.Token, viewerID)

	// Suspend
	err := svc.SuspendMember(ctx, org.ID, member.ID, creatorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status
	updated := memberRepo.members[member.ID]
	if updated.Status != valueobject.MemberStatusSuspended {
		t.Errorf("expected SUSPENDED, got %s", updated.Status)
	}
	if updated.SuspendedBy == nil || *updated.SuspendedBy != creatorID {
		t.Error("expected SuspendedBy to be set to actor")
	}
}

func TestSuspendMember_CannotSuspendOwner(t *testing.T) {
	svc, _, memberRepo, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)

	var ownerMemberID uuid.UUID
	for _, m := range memberRepo.members {
		if m.OrgID == org.ID && m.Role == valueobject.OrgRoleOwner {
			ownerMemberID = m.ID
			break
		}
	}

	err := svc.SuspendMember(ctx, org.ID, ownerMemberID, creatorID)
	if !errors.Is(err, ErrCannotSuspendOwner) {
		t.Errorf("expected ErrCannotSuspendOwner, got %v", err)
	}
}

func TestUnsuspendMember(t *testing.T) {
	svc, _, memberRepo, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()
	viewerID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	inv, _ := svc.InviteMember(ctx, org.ID, "v@test.com", valueobject.OrgRoleViewer, creatorID)
	member, _ := svc.AcceptInvitation(ctx, inv.Token, viewerID)

	// Suspend then unsuspend
	_ = svc.SuspendMember(ctx, org.ID, member.ID, creatorID)
	err := svc.UnsuspendMember(ctx, org.ID, member.ID, creatorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := memberRepo.members[member.ID]
	if updated.Status != valueobject.MemberStatusActive {
		t.Errorf("expected ACTIVE, got %s", updated.Status)
	}
	if updated.SuspendedBy != nil {
		t.Error("expected SuspendedBy to be nil after unsuspend")
	}
}

func TestChangeRole(t *testing.T) {
	svc, _, memberRepo, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()
	viewerID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	inv, _ := svc.InviteMember(ctx, org.ID, "v@test.com", valueobject.OrgRoleViewer, creatorID)
	member, _ := svc.AcceptInvitation(ctx, inv.Token, viewerID)

	// Change VIEWER → ADMIN
	err := svc.ChangeRole(ctx, org.ID, member.ID, creatorID, valueobject.OrgRoleAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := memberRepo.members[member.ID]
	if updated.Role != valueobject.OrgRoleAdmin {
		t.Errorf("expected ADMIN, got %s", updated.Role)
	}
}

func TestRevokeInvitation(t *testing.T) {
	svc, _, _, invRepo, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	_ = svc.orgRepo.(*mockOrgRepo).Update(ctx, org)

	inv, _ := svc.InviteMember(ctx, org.ID, "revoke@test.com", valueobject.OrgRoleViewer, creatorID)

	err := svc.RevokeInvitation(ctx, inv.ID, creatorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := invRepo.invitations[inv.ID]
	if updated.Status != valueobject.InvitationStatusRevoked {
		t.Errorf("expected REVOKED, got %s", updated.Status)
	}
}

func TestConfigureWebhook_StarterPlan(t *testing.T) {
	svc, orgRepo, _, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Test", creatorID)
	org.PlanTier = valueobject.PlanTierStarter
	orgRepo.orgs[org.ID] = org

	err := svc.ConfigureWebhook(ctx, org.ID, "https://example.com/webhook", creatorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := orgRepo.orgs[org.ID]
	if updated.WebhookURL != "https://example.com/webhook" {
		t.Errorf("expected webhook URL, got %q", updated.WebhookURL)
	}
	if updated.WebhookSecret == "" {
		t.Error("expected non-empty webhook secret")
	}
}

func TestConfigureWebhook_FreePlanDenied(t *testing.T) {
	svc, _, _, _, _ := setupOrgService()
	ctx := context.Background()
	creatorID := uuid.New()

	org, _ := svc.CreateOrganization(ctx, "Free Org", creatorID)
	// Free plan by default

	err := svc.ConfigureWebhook(ctx, org.ID, "https://example.com/webhook", creatorID)
	if err == nil {
		t.Error("expected error for FREE plan webhook")
	}
}

func TestMaxMembers(t *testing.T) {
	tests := []struct {
		plan     valueobject.PlanTier
		expected int
	}{
		{valueobject.PlanTierFree, 1},
		{valueobject.PlanTierStarter, 3},
		{valueobject.PlanTierPro, 10},
	}

	for _, tt := range tests {
		org := &entity.Organization{PlanTier: tt.plan}
		if got := org.MaxMembers(); got != tt.expected {
			t.Errorf("plan %s: expected %d, got %d", tt.plan, tt.expected, got)
		}
	}
}

func TestOrgAuditService_GetAuditLog(t *testing.T) {
	auditRepo := newMockOrgAuditRepo()
	svc := NewOrgAuditService(auditRepo)
	ctx := context.Background()

	orgID := uuid.New()
	actorID := uuid.New()

	// Create 5 entries
	for i := 0; i < 5; i++ {
		_ = svc.LogAction(ctx, orgID, actorID, "test.action", "test", nil, nil)
	}

	page, err := svc.GetAuditLog(ctx, orgID, 3, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 5 {
		t.Errorf("expected total 5, got %d", page.Total)
	}
	if len(page.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(page.Entries))
	}
	if page.Limit != 3 {
		t.Errorf("expected limit 3, got %d", page.Limit)
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Org", "my-org"},
		{"Test Company 123", "test-company-123"},
		{"Special!@#Chars", "specialchars"},
		{"already-slugged", "already-slugged"},
	}

	for _, tt := range tests {
		got := generateSlug(tt.input)
		if got != tt.expected {
			t.Errorf("generateSlug(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestOrgRolePermissions(t *testing.T) {
	tests := []struct {
		role           valueobject.OrgRole
		canManage      bool
		canManageAdmin bool
		canSync        bool
		canManageOrg   bool
		canViewAudit   bool
	}{
		{valueobject.OrgRoleOwner, true, true, true, true, true},
		{valueobject.OrgRoleAdmin, true, false, true, false, true},
		{valueobject.OrgRoleViewer, false, false, false, false, false},
	}

	for _, tt := range tests {
		if got := tt.role.CanManageMembers(); got != tt.canManage {
			t.Errorf("%s.CanManageMembers() = %v, want %v", tt.role, got, tt.canManage)
		}
		if got := tt.role.CanManageAdmins(); got != tt.canManageAdmin {
			t.Errorf("%s.CanManageAdmins() = %v, want %v", tt.role, got, tt.canManageAdmin)
		}
		if got := tt.role.CanTriggerSync(); got != tt.canSync {
			t.Errorf("%s.CanTriggerSync() = %v, want %v", tt.role, got, tt.canSync)
		}
		if got := tt.role.CanManageOrg(); got != tt.canManageOrg {
			t.Errorf("%s.CanManageOrg() = %v, want %v", tt.role, got, tt.canManageOrg)
		}
		if got := tt.role.CanViewAuditLog(); got != tt.canViewAudit {
			t.Errorf("%s.CanViewAuditLog() = %v, want %v", tt.role, got, tt.canViewAudit)
		}
	}
}

func TestMemberStatus(t *testing.T) {
	if !valueobject.MemberStatusActive.IsActive() {
		t.Error("ACTIVE should be active")
	}
	if valueobject.MemberStatusSuspended.IsActive() {
		t.Error("SUSPENDED should not be active")
	}
	if !valueobject.MemberStatusActive.IsValid() {
		t.Error("ACTIVE should be valid")
	}
	if valueobject.MemberStatus("INVALID").IsValid() {
		t.Error("INVALID should not be valid")
	}
}

func TestInvitationStatus(t *testing.T) {
	if !valueobject.InvitationStatusPending.IsPending() {
		t.Error("PENDING should be pending")
	}
	if valueobject.InvitationStatusAccepted.IsPending() {
		t.Error("ACCEPTED should not be pending")
	}
}

func TestOrgMember_SuspendUnsuspend(t *testing.T) {
	member := entity.NewOrgMember(uuid.New(), uuid.New(), valueobject.OrgRoleViewer, nil)
	actorID := uuid.New()

	if !member.IsActive() {
		t.Error("new member should be active")
	}

	member.Suspend(actorID)
	if member.IsActive() {
		t.Error("suspended member should not be active")
	}
	if member.SuspendedBy == nil || *member.SuspendedBy != actorID {
		t.Error("SuspendedBy should be set")
	}
	if member.SuspendedAt == nil {
		t.Error("SuspendedAt should be set")
	}

	member.Unsuspend()
	if !member.IsActive() {
		t.Error("unsuspended member should be active")
	}
	if member.SuspendedBy != nil {
		t.Error("SuspendedBy should be nil after unsuspend")
	}
	if member.SuspendedAt != nil {
		t.Error("SuspendedAt should be nil after unsuspend")
	}
}

func TestOrgInvitation_IsExpired(t *testing.T) {
	inv := entity.NewOrgInvitation(uuid.New(), "test@test.com", valueobject.OrgRoleViewer, uuid.New(), "token", time.Now().UTC().Add(1*time.Hour))

	if inv.IsExpired(time.Now().UTC()) {
		t.Error("invitation should not be expired")
	}
	if !inv.IsExpired(time.Now().UTC().Add(2 * time.Hour)) {
		t.Error("invitation should be expired after 2 hours")
	}
}

func TestOrganization_Features(t *testing.T) {
	org := entity.NewOrganization("Test", uuid.New())

	// FREE plan
	if org.CanUseSSO() {
		t.Error("FREE should not support SSO")
	}
	if org.CanUseWebhooks() {
		t.Error("FREE should not support webhooks")
	}

	// STARTER plan
	org.PlanTier = valueobject.PlanTierStarter
	if org.CanUseSSO() {
		t.Error("STARTER should not support SSO")
	}
	if !org.CanUseWebhooks() {
		t.Error("STARTER should support webhooks")
	}

	// PRO plan
	org.PlanTier = valueobject.PlanTierPro
	if !org.CanUseSSO() {
		t.Error("PRO should support SSO")
	}
	if !org.CanUseWebhooks() {
		t.Error("PRO should support webhooks")
	}
}

// Verify that notification prefs JSON is correctly initialized
func TestOrgMember_DefaultNotificationPrefs(t *testing.T) {
	member := entity.NewOrgMember(uuid.New(), uuid.New(), valueobject.OrgRoleViewer, nil)

	var prefs map[string]bool
	if err := json.Unmarshal(member.NotificationPrefs, &prefs); err != nil {
		t.Fatalf("failed to unmarshal notification prefs: %v", err)
	}
	if !prefs["email"] {
		t.Error("expected email notifications enabled by default")
	}
	if !prefs["push"] {
		t.Error("expected push notifications enabled by default")
	}
}

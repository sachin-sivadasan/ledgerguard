package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
)

// --- minimal repo mocks (only the methods RequireOrg calls do real work) ---

type stubOrgRepo struct{ org *entity.Organization }

func (s *stubOrgRepo) Create(context.Context, *entity.Organization) error { return nil }
func (s *stubOrgRepo) FindByID(context.Context, uuid.UUID) (*entity.Organization, error) {
	return s.org, nil
}
func (s *stubOrgRepo) FindBySlug(context.Context, string) (*entity.Organization, error) {
	return s.org, nil
}
func (s *stubOrgRepo) Update(context.Context, *entity.Organization) error { return nil }
func (s *stubOrgRepo) Delete(context.Context, uuid.UUID) error            { return nil }

type stubMemberRepo struct{ member *entity.OrgMember }

func (s *stubMemberRepo) Create(context.Context, *entity.OrgMember) error { return nil }
func (s *stubMemberRepo) FindByID(context.Context, uuid.UUID) (*entity.OrgMember, error) {
	return s.member, nil
}
func (s *stubMemberRepo) FindByOrgAndUser(context.Context, uuid.UUID, uuid.UUID) (*entity.OrgMember, error) {
	return s.member, nil
}
func (s *stubMemberRepo) FindByOrgID(context.Context, uuid.UUID) ([]*entity.OrgMember, error) {
	return nil, nil
}
func (s *stubMemberRepo) FindByUserID(context.Context, uuid.UUID) ([]*entity.OrgMember, error) {
	return []*entity.OrgMember{s.member}, nil
}
func (s *stubMemberRepo) CountByOrgID(context.Context, uuid.UUID) (int, error) { return 1, nil }
func (s *stubMemberRepo) Update(context.Context, *entity.OrgMember) error      { return nil }
func (s *stubMemberRepo) Delete(context.Context, uuid.UUID) error              { return nil }

// TestRequireOrg_EnrichesLoggerWithTenant is the "one org's story" guard: once RequireOrg
// resolves the tenant, every downstream handler log line must carry org_id + user_id.
func TestRequireOrg_EnrichesLoggerWithTenant(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	user := entity.NewUser("firebase-uid", "alice@shop.com")
	orgID := uuid.New()
	member := entity.NewOrgMember(orgID, user.ID, valueobject.OrgRoleOwner, nil)
	org := entity.NewOrganization("Acme", user.ID)

	mw := NewOrgContextMiddleware(&stubOrgRepo{org: org}, &stubMemberRepo{member: member})

	handler := mw.RequireOrg(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.FromContext(r.Context()).Info("handled")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Org-Id", orgID.String())
	ctx := logging.ContextWithLogger(req.Context(), base)
	ctx = context.WithValue(ctx, userContextKey, user) // RequireOrg needs an authenticated user
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 handler log entry, got %d", len(entries))
	}
	fields := entries[0].ContextMap()
	if got := fields["org_id"]; got != orgID.String() {
		t.Errorf("org_id = %v, want %s", got, orgID)
	}
	if got := fields["user_id"]; got != user.ID.String() {
		t.Errorf("user_id = %v, want %s", got, user.ID)
	}
}

// TestRequireOrg_NoTenantLeakWithoutOrg is the negative guard for tenant isolation: a
// request that goes through RequestLogger but NOT RequireOrg must carry request_id but
// must NOT carry org_id/user_id — so enrichment can never leak one tenant's identity onto
// an unauthenticated or non-org-scoped line.
func TestRequireOrg_NoTenantLeakWithoutOrg(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.FromContext(r.Context()).Info("handled")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(logging.ContextWithLogger(req.Context(), base))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	fields := recorded.All()[0].ContextMap()
	if _, ok := fields["request_id"]; !ok {
		t.Error("expected request_id to be present via RequestLogger")
	}
	if _, ok := fields["org_id"]; ok {
		t.Errorf("org_id must NOT be present without RequireOrg, got %v", fields["org_id"])
	}
	if _, ok := fields["user_id"]; ok {
		t.Errorf("user_id must NOT be present without RequireOrg, got %v", fields["user_id"])
	}
}

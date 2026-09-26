package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

func TestResolvePartnerAccount_NotFound_Returns404(t *testing.T) {
	user := &entity.User{ID: uuid.New(), FirebaseUID: "test-uid"}
	repo := &mockPartnerRepoForApp{
		findErr: persistence.ErrPartnerAccountNotFound,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	_, lookupErr := resolvePartnerAccount(req, repo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", lookupErr.statusCode)
	}
}

func TestResolvePartnerAccount_DBError_Returns503(t *testing.T) {
	user := &entity.User{ID: uuid.New(), FirebaseUID: "test-uid"}
	repo := &mockPartnerRepoForApp{
		findErr: errors.New("connection refused"),
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	_, lookupErr := resolvePartnerAccount(req, repo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", lookupErr.statusCode)
	}
}

func TestResolvePartnerAccount_OrgNotFound_Returns404(t *testing.T) {
	org := &entity.Organization{ID: uuid.New(), Name: "Test Org"}
	repo := &mockPartnerRepoForApp{
		findErr: persistence.ErrPartnerAccountNotFound,
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.SetOrgContext(req.Context(), org)
	req = req.WithContext(ctx)

	_, lookupErr := resolvePartnerAccount(req, repo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", lookupErr.statusCode)
	}
}

func TestResolvePartnerAccount_OrgDBError_Returns503(t *testing.T) {
	org := &entity.Organization{ID: uuid.New(), Name: "Test Org"}
	repo := &mockPartnerRepoForApp{
		findErr: errors.New("connection timeout"),
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.SetOrgContext(req.Context(), org)
	req = req.WithContext(ctx)

	_, lookupErr := resolvePartnerAccount(req, repo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", lookupErr.statusCode)
	}
}

func TestResolveAppFromRequest_NotFound_Returns404(t *testing.T) {
	appID := uuid.New()
	appRepo := &mockAppRepo{findErr: persistence.ErrAppNotFound}
	partnerRepo := &mockPartnerRepoForApp{}

	req := httptest.NewRequest(http.MethodGet, "/apps/"+appID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, lookupErr := resolveAppFromRequest(req, partnerRepo, appRepo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", lookupErr.statusCode)
	}
}

func TestResolveAppFromRequest_DBError_Returns503(t *testing.T) {
	appID := uuid.New()
	appRepo := &mockAppRepo{findErr: errors.New("connection refused")}
	partnerRepo := &mockPartnerRepoForApp{}

	req := httptest.NewRequest(http.MethodGet, "/apps/"+appID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, lookupErr := resolveAppFromRequest(req, partnerRepo, appRepo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", lookupErr.statusCode)
	}
}

func TestResolveAppFromRequest_NilApp_Returns404(t *testing.T) {
	appID := uuid.New()
	appRepo := &mockAppRepo{app: nil, findErr: nil}
	partnerRepo := &mockPartnerRepoForApp{}

	req := httptest.NewRequest(http.MethodGet, "/apps/"+appID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	_, lookupErr := resolveAppFromRequest(req, partnerRepo, appRepo)
	if lookupErr == nil {
		t.Fatal("expected error, got nil")
	}
	if lookupErr.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", lookupErr.statusCode)
	}
}

// requestForApp builds a request carrying the appID URL param and the given org
// context (as OrgContextMW would set it).
func requestForApp(appID uuid.UUID, org *entity.Organization) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/apps/"+appID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", appID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(middleware.SetOrgContext(req.Context(), org))
	return req
}

// TestResolveAppFromRequest_CrossOrg_Returns404 is the S1 regression guard: a
// caller whose org owns partner-account Y must NOT be able to resolve an app that
// belongs to partner-account X (a different org).
func TestResolveAppFromRequest_CrossOrg_Returns404(t *testing.T) {
	appID := uuid.New()
	callerAccountID := uuid.New()
	otherAccountID := uuid.New() // the app's owning account — a DIFFERENT org

	appRepo := &mockAppRepo{app: &entity.App{ID: appID, PartnerAccountID: otherAccountID}}
	partnerRepo := &mockPartnerRepoForApp{account: &entity.PartnerAccount{ID: callerAccountID}}

	req := requestForApp(appID, &entity.Organization{ID: uuid.New(), Name: "Caller Org"})

	_, lookupErr := resolveAppFromRequest(req, partnerRepo, appRepo)
	if lookupErr == nil {
		t.Fatal("expected cross-org access to be denied, got nil error (app leaked)")
	}
	if lookupErr.statusCode != http.StatusNotFound {
		t.Errorf("expected 404 for cross-org app, got %d", lookupErr.statusCode)
	}
}

// TestResolveAppFromRequest_SameOrg_Succeeds ensures the legitimate flow still
// works: the app's partner account matches the caller's.
func TestResolveAppFromRequest_SameOrg_Succeeds(t *testing.T) {
	appID := uuid.New()
	accountID := uuid.New()

	appRepo := &mockAppRepo{app: &entity.App{ID: appID, PartnerAccountID: accountID}}
	partnerRepo := &mockPartnerRepoForApp{account: &entity.PartnerAccount{ID: accountID}}

	req := requestForApp(appID, &entity.Organization{ID: uuid.New(), Name: "Owner Org"})

	app, lookupErr := resolveAppFromRequest(req, partnerRepo, appRepo)
	if lookupErr != nil {
		t.Fatalf("expected success for same-org app, got error %d: %s", lookupErr.statusCode, lookupErr.message)
	}
	if app == nil || app.ID != appID {
		t.Errorf("expected the resolved app %s, got %+v", appID, app)
	}
}

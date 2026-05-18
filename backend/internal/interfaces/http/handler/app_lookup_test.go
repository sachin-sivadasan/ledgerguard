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

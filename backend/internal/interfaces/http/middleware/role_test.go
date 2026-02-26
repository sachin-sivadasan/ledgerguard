package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestRoleMiddleware_NoUserInContext(t *testing.T) {
	middleware := RequireRoles(valueobject.RoleAdmin)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRoleMiddleware_AdminAccessingAdminRoute(t *testing.T) {
	middleware := RequireRoles(valueobject.RoleAdmin)

	var called bool
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	user := &entity.User{
		Role: valueobject.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(setUserContext(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRoleMiddleware_OwnerAccessingAdminRoute(t *testing.T) {
	// OWNER should have access to ADMIN routes (OWNER is superset)
	middleware := RequireRoles(valueobject.RoleAdmin)

	var called bool
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	user := &entity.User{
		Role: valueobject.RoleOwner,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(setUserContext(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRoleMiddleware_AdminAccessingOwnerOnlyRoute(t *testing.T) {
	// ADMIN should NOT have access to OWNER-only routes
	middleware := RequireRoles(valueobject.RoleOwner)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	user := &entity.User{
		Role: valueobject.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(setUserContext(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRoleMiddleware_MultipleRolesAllowed(t *testing.T) {
	// Route allows both OWNER and ADMIN
	middleware := RequireRoles(valueobject.RoleOwner, valueobject.RoleAdmin)

	var called bool
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	user := &entity.User{
		Role: valueobject.RoleAdmin,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(setUserContext(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !called {
		t.Error("expected handler to be called")
	}
}

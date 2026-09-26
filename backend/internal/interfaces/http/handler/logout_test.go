package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type mockRevoker struct {
	gotUID string
	called bool
	err    error
}

func (m *mockRevoker) RevokeRefreshTokens(_ context.Context, uid string) error {
	m.called = true
	m.gotUID = uid
	return m.err
}

func TestLogout_NoUser_Returns401(t *testing.T) {
	rev := &mockRevoker{}
	h := NewLogoutHandler(rev)

	rec := httptest.NewRecorder()
	h.Logout(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", rec.Code)
	}
	if rev.called {
		t.Error("revoker must not be called when unauthenticated")
	}
}

func TestLogout_RevokesCallersTokens(t *testing.T) {
	rev := &mockRevoker{}
	h := NewLogoutHandler(rev)
	user := &entity.User{ID: uuid.New(), FirebaseUID: "uid-123"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req = req.WithContext(middleware.SetUserContext(req.Context(), user))
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if !rev.called || rev.gotUID != "uid-123" {
		t.Errorf("expected revoke for the caller's FirebaseUID, called=%v uid=%q", rev.called, rev.gotUID)
	}
}

func TestLogout_RevokerError_Returns503(t *testing.T) {
	rev := &mockRevoker{err: errors.New("firebase down")}
	h := NewLogoutHandler(rev)
	user := &entity.User{ID: uuid.New(), FirebaseUID: "uid-123"}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req = req.WithContext(middleware.SetUserContext(req.Context(), user))
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on revoke error, got %d", rec.Code)
	}
}

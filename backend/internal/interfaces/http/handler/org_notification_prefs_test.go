package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// notifPrefsRequest builds a PUT for member {userId}/notifications with the caller's
// org member injected (as OrgContextMW would).
func notifPrefsRequest(targetMemberID uuid.UUID, caller *entity.OrgMember) *http.Request {
	req := httptest.NewRequest(http.MethodPut, "/x", strings.NewReader(`{"daily":true}`))
	req = withURLParam(req, "userId", targetMemberID.String())
	if caller != nil {
		req = req.WithContext(middleware.SetOrgMemberContext(req.Context(), caller))
	}
	return req
}

// TestUpdateNotificationPrefs_OtherMember_ViewerDenied is the S2 self-or-admin guard:
// a VIEWER must not be able to modify another member's notification preferences.
// The denial returns before any service call, so a nil service is safe here.
func TestUpdateNotificationPrefs_OtherMember_ViewerDenied(t *testing.T) {
	h := NewOrgHandler(nil)
	caller := &entity.OrgMember{ID: uuid.New(), UserID: uuid.New(), Role: valueobject.OrgRoleViewer}
	target := uuid.New() // a DIFFERENT member

	rec := httptest.NewRecorder()
	h.UpdateNotificationPrefs(rec, notifPrefsRequest(target, caller))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer editing another member's prefs, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUpdateNotificationPrefs_NoMember_Denied: missing org context → 403 (not a panic).
func TestUpdateNotificationPrefs_NoMember_Denied(t *testing.T) {
	h := NewOrgHandler(nil)
	rec := httptest.NewRecorder()
	h.UpdateNotificationPrefs(rec, notifPrefsRequest(uuid.New(), nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without org context, got %d", rec.Code)
	}
}

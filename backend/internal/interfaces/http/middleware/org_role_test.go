package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// serve runs the given gate middleware over a handler that records a 200, with the
// supplied member (nil = no org context) injected, and returns the status code.
func serve(gate func(http.Handler) http.Handler, member *entity.OrgMember) int {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if member != nil {
		req = req.WithContext(SetOrgMemberContext(req.Context(), member))
	}
	rec := httptest.NewRecorder()
	gate(next).ServeHTTP(rec, req)
	return rec.Code
}

func member(role valueobject.OrgRole) *entity.OrgMember {
	return &entity.OrgMember{ID: uuid.New(), OrgID: uuid.New(), UserID: uuid.New(), Role: role}
}

func TestRequireOrgOwner(t *testing.T) {
	cases := []struct {
		name   string
		member *entity.OrgMember
		want   int
	}{
		{"owner allowed", member(valueobject.OrgRoleOwner), http.StatusOK},
		{"admin denied", member(valueobject.OrgRoleAdmin), http.StatusForbidden},
		{"viewer denied", member(valueobject.OrgRoleViewer), http.StatusForbidden},
		{"no member denied", nil, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := serve(RequireOrgOwner(), tc.member); got != tc.want {
				t.Errorf("RequireOrgOwner(%v) = %d, want %d", tc.member, got, tc.want)
			}
		})
	}
}

func TestRequireOrgAdmin(t *testing.T) {
	cases := []struct {
		name   string
		member *entity.OrgMember
		want   int
	}{
		{"owner allowed", member(valueobject.OrgRoleOwner), http.StatusOK},
		{"admin allowed", member(valueobject.OrgRoleAdmin), http.StatusOK},
		{"viewer denied", member(valueobject.OrgRoleViewer), http.StatusForbidden},
		{"no member denied", nil, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := serve(RequireOrgAdmin(), tc.member); got != tc.want {
				t.Errorf("RequireOrgAdmin(%v) = %d, want %d", tc.member, got, tc.want)
			}
		})
	}
}

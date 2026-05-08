package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

const orgMemberContextKey contextKey = "orgMember"
const orgContextKey contextKey = "org"

// OrgContextMiddleware resolves the organization from the request, checks
// membership, status, and minimum role before passing through to the handler.
type OrgContextMiddleware struct {
	orgRepo    repository.OrganizationRepository
	memberRepo repository.MemberRepository
}

func NewOrgContextMiddleware(
	orgRepo repository.OrganizationRepository,
	memberRepo repository.MemberRepository,
) *OrgContextMiddleware {
	return &OrgContextMiddleware{
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
	}
}

// RequireOrg resolves the org from URL param (:orgId) and verifies the user
// is an active member. Sets both org and member in context.
func (m *OrgContextMiddleware) RequireOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		orgIDStr := chi.URLParam(r, "orgId")
		if orgIDStr == "" {
			// Fallback: check X-Org-Id header
			orgIDStr = r.Header.Get("X-Org-Id")
		}

		if orgIDStr == "" {
			// Auto-select: if user has exactly 1 org, use it
			memberships, err := m.memberRepo.FindByUserID(r.Context(), user.ID)
			if err != nil || len(memberships) == 0 {
				writeError(w, http.StatusBadRequest, "organization selection required")
				return
			}
			if len(memberships) == 1 {
				orgIDStr = memberships[0].OrgID.String()
			} else {
				writeError(w, http.StatusBadRequest, "multiple organizations found; pass X-Org-Id header or use /orgs/:orgId path")
				return
			}
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid organization ID")
			return
		}

		org, err := m.orgRepo.FindByID(r.Context(), orgID)
		if err != nil {
			writeError(w, http.StatusNotFound, "organization not found")
			return
		}

		member, err := m.memberRepo.FindByOrgAndUser(r.Context(), orgID, user.ID)
		if err != nil {
			writeError(w, http.StatusForbidden, "not a member of this organization")
			return
		}

		if !member.IsActive() {
			writeError(w, http.StatusForbidden, "account_suspended")
			return
		}

		ctx := context.WithValue(r.Context(), orgContextKey, org)
		ctx = context.WithValue(ctx, orgMemberContextKey, member)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireOrgRole returns middleware that checks if the org member has the
// minimum required role. Must be used AFTER RequireOrg.
func RequireOrgRole(minRoles ...valueobject.OrgRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			member := OrgMemberFromContext(r.Context())
			if member == nil {
				writeError(w, http.StatusForbidden, "org context required")
				return
			}

			// OWNER always passes
			if member.Role == valueobject.OrgRoleOwner {
				next.ServeHTTP(w, r)
				return
			}

			for _, role := range minRoles {
				if member.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeError(w, http.StatusForbidden, "insufficient organization role")
		})
	}
}

// SetOrgContext sets the organization in context (exported for testing).
func SetOrgContext(ctx context.Context, org *entity.Organization) context.Context {
	return context.WithValue(ctx, orgContextKey, org)
}

// OrgFromContext retrieves the organization from the request context.
func OrgFromContext(ctx context.Context) *entity.Organization {
	org, ok := ctx.Value(orgContextKey).(*entity.Organization)
	if !ok {
		return nil
	}
	return org
}

// OrgMemberFromContext retrieves the org member from the request context.
func OrgMemberFromContext(ctx context.Context) *entity.OrgMember {
	member, ok := ctx.Value(orgMemberContextKey).(*entity.OrgMember)
	if !ok {
		return nil
	}
	return member
}

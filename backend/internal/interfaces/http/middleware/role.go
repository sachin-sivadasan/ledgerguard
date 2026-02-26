package middleware

import (
	"context"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// RequireRoles returns middleware that checks if the user has one of the required roles.
// OWNER role has access to all routes (superset of ADMIN).
func RequireRoles(allowedRoles ...valueobject.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if !hasRequiredRole(user.Role, allowedRoles) {
				writeError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasRequiredRole checks if the user's role is in the allowed roles list.
// OWNER has implicit access to all roles (superset).
func hasRequiredRole(userRole valueobject.Role, allowedRoles []valueobject.Role) bool {
	// OWNER has access to everything
	if userRole == valueobject.RoleOwner {
		return true
	}

	// Check if user's role is in the allowed list
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}

	return false
}

// setUserContext is a helper for testing - sets user in context
func setUserContext(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

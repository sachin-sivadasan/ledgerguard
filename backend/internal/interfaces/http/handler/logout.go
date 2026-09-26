package handler

import (
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// LogoutHandler revokes the caller's refresh tokens (server-side force-logout).
type LogoutHandler struct {
	revoker service.TokenRevoker
}

// NewLogoutHandler constructs a LogoutHandler.
func NewLogoutHandler(revoker service.TokenRevoker) *LogoutHandler {
	return &LogoutHandler{revoker: revoker}
}

// Logout revokes all refresh tokens for the authenticated user. With revocation-checked
// verification (S4) enabled, this invalidates their existing sessions within the ID-token
// TTL — the server-side counterpart to a client sign-out.
// POST /api/v1/auth/logout
func (h *LogoutHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := h.revoker.RevokeRefreshTokens(r.Context(), user.FirebaseUID); err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "failed to revoke session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

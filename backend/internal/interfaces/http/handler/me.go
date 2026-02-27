package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

// GetMe returns the current user's profile.
// GET /api/v1/me
func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          user.ID.String(),
		"email":       user.Email,
		"role":        user.Role.String(),
		"plan_tier":   user.PlanTier.String(),
		"created_at":  user.CreatedAt,
	})
}

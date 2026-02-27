package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type IntegrationStatusHandler struct {
	partnerRepo repository.PartnerAccountRepository
}

func NewIntegrationStatusHandler(partnerRepo repository.PartnerAccountRepository) *IntegrationStatusHandler {
	return &IntegrationStatusHandler{
		partnerRepo: partnerRepo,
	}
}

// GetStatus returns the current integration status for the user.
// GET /api/v1/integrations/shopify/status
func (h *IntegrationStatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	account, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil || account == nil {
		// Not connected
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"connected":  false,
			"partner_id": nil,
		})
		return
	}

	// Connected
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected":        true,
		"partner_id":       account.PartnerID,
		"integration_type": account.IntegrationType.String(),
		"connected_at":     account.CreatedAt,
	})
}

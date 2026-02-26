package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type ManualTokenHandler struct {
	encryptor   Encryptor
	partnerRepo repository.PartnerAccountRepository
}

func NewManualTokenHandler(encryptor Encryptor, partnerRepo repository.PartnerAccountRepository) *ManualTokenHandler {
	return &ManualTokenHandler{
		encryptor:   encryptor,
		partnerRepo: partnerRepo,
	}
}

type addTokenRequest struct {
	Token     string `json:"token"`
	PartnerID string `json:"partner_id"`
}

// AddToken handles adding a manual partner token.
// POST /api/v1/integrations/shopify/token
func (h *ManualTokenHandler) AddToken(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req addTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Token == "" {
		writeJSONError(w, http.StatusBadRequest, "token is required")
		return
	}

	if req.PartnerID == "" {
		writeJSONError(w, http.StatusBadRequest, "partner_id is required")
		return
	}

	// Encrypt token
	encryptedToken, err := h.encryptor.Encrypt([]byte(req.Token))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to encrypt token")
		return
	}

	// Check if account already exists
	existingAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err == nil && existingAccount != nil {
		// Update existing account
		existingAccount.PartnerID = req.PartnerID
		existingAccount.EncryptedAccessToken = encryptedToken
		existingAccount.IntegrationType = valueobject.IntegrationTypeManual

		if err := h.partnerRepo.Update(r.Context(), existingAccount); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to update partner account")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":      "Manual token updated successfully",
			"id":           existingAccount.ID.String(),
			"partner_id":   req.PartnerID,
			"masked_token": maskToken(req.Token),
		})
		return
	}

	// Create new account
	account := entity.NewPartnerAccount(user.ID, req.PartnerID, valueobject.IntegrationTypeManual, encryptedToken)

	if err := h.partnerRepo.Create(r.Context(), account); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save partner account")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Manual token added successfully",
		"id":           account.ID.String(),
		"partner_id":   req.PartnerID,
		"masked_token": maskToken(req.Token),
	})
}

// GetToken retrieves the current token info (masked).
// GET /api/v1/integrations/shopify/token
func (h *ManualTokenHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	account, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Decrypt token to show masked version
	decryptedToken, err := h.encryptor.Decrypt(account.EncryptedAccessToken)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to decrypt token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":               account.ID.String(),
		"partner_id":       account.PartnerID,
		"integration_type": account.IntegrationType.String(),
		"masked_token":     maskToken(string(decryptedToken)),
		"created_at":       account.CreatedAt,
	})
}

// RevokeToken deletes the partner account and token.
// DELETE /api/v1/integrations/shopify/token
func (h *ManualTokenHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := h.partnerRepo.Delete(r.Context(), user.ID); err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Token revoked successfully",
	})
}

// maskToken returns a masked version of the token showing only last 4 characters.
func maskToken(token string) string {
	if len(token) <= 4 {
		return "***..." + token
	}
	return "***..." + token[len(token)-4:]
}

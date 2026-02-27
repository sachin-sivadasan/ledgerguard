package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// PartnerClient interface for fetching apps from Shopify Partner API
type PartnerClient interface {
	FetchApps(ctx context.Context, organizationID, accessToken string) ([]external.PartnerApp, error)
}

type AppHandler struct {
	partnerClient PartnerClient
	partnerRepo   repository.PartnerAccountRepository
	appRepo       repository.AppRepository
	decryptor     Encryptor
}

func NewAppHandler(
	partnerClient PartnerClient,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
	decryptor Encryptor,
) *AppHandler {
	return &AppHandler{
		partnerClient: partnerClient,
		partnerRepo:   partnerRepo,
		appRepo:       appRepo,
		decryptor:     decryptor,
	}
}

// GetAvailableApps fetches apps from Shopify Partner API
// GET /api/v1/apps/available
func (h *AppHandler) GetAvailableApps(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Check if partner client is configured
	if h.partnerClient == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "Shopify Partner API not configured")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Decrypt access token
	decryptedToken, err := h.decryptor.Decrypt(partnerAccount.EncryptedAccessToken)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to decrypt token")
		return
	}

	// Fetch apps from Partner API
	apps, err := h.partnerClient.FetchApps(r.Context(), partnerAccount.PartnerID, string(decryptedToken))
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "failed to fetch apps from Partner API")
		return
	}

	// Convert to response format
	appResponses := make([]map[string]interface{}, len(apps))
	for i, app := range apps {
		appResponses[i] = map[string]interface{}{
			"id":   app.ID,
			"name": app.Name,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apps": appResponses,
	})
}

type selectAppRequest struct {
	PartnerAppID string `json:"partner_app_id"`
	Name         string `json:"name"`
}

// SelectApp stores the selected app
// POST /api/v1/apps/select
func (h *AppHandler) SelectApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req selectAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PartnerAppID == "" {
		writeJSONError(w, http.StatusBadRequest, "partner_app_id is required")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Check if app already exists
	existingApp, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, req.PartnerAppID)
	if err == nil && existingApp != nil {
		writeJSONError(w, http.StatusConflict, "app already tracked")
		return
	}

	// Create new app
	app := entity.NewApp(partnerAccount.ID, req.PartnerAppID, req.Name)

	if err := h.appRepo.Create(r.Context(), app); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "App added successfully",
		"id":             app.ID.String(),
		"partner_app_id": app.PartnerAppID,
		"name":           app.Name,
	})
}

// ListApps returns user's tracked apps
// GET /api/v1/apps
func (h *AppHandler) ListApps(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Get apps
	apps, err := h.appRepo.FindByPartnerAccountID(r.Context(), partnerAccount.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch apps")
		return
	}

	// Convert to response format
	appResponses := make([]map[string]interface{}, len(apps))
	for i, app := range apps {
		appResponses[i] = map[string]interface{}{
			"id":               app.ID.String(),
			"partner_app_id":   app.PartnerAppID,
			"name":             app.Name,
			"tracking_enabled": app.TrackingEnabled,
			"created_at":       app.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apps": appResponses,
	})
}

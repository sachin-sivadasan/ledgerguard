package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// PartnerClient interface for fetching apps from Shopify Partner API
type PartnerClient interface {
	FetchApps(ctx context.Context, organizationID, accessToken string) ([]external.PartnerApp, error)
	FetchInstallCount(ctx context.Context, organizationID, accessToken, partnerAppID string) (int, error)
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

	// Extract numeric ID for use with other endpoints
	appID := extractNumericAppID(app.PartnerAppID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "App added successfully",
		"id":                 appID,
		"name":               app.Name,
		"revenue_share_tier": app.RevenueShareTier.String(),
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
		// Extract numeric ID from GID (e.g., "gid://partners/App/4599915" -> "4599915")
		// This ID can be used directly with other endpoints like /apps/{id}/subscriptions
		appID := extractNumericAppID(app.PartnerAppID)

		appResponses[i] = map[string]interface{}{
			"id":                 appID,
			"name":               app.Name,
			"tracking_enabled":   app.TrackingEnabled,
			"revenue_share_tier": app.RevenueShareTier.String(),
			"install_count":      app.InstallCount,
			"created_at":         app.CreatedAt,
			"updated_at":         app.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apps": appResponses,
	})
}

// extractNumericAppID extracts the numeric ID from a Shopify GID
// e.g., "gid://partners/App/4599915" -> "4599915"
func extractNumericAppID(gid string) string {
	parts := strings.Split(gid, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return gid
}

type updateAppTierRequest struct {
	RevenueShareTier string `json:"revenue_share_tier"`
}

// UpdateAppTier updates the revenue share tier for an app
// PATCH /api/v1/apps/{appID}/tier
// appID can be internal UUID or Shopify GID (gid://partners/App/xxx or just the numeric part)
func (h *AppHandler) UpdateAppTier(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get app ID from URL
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app_id is required")
		return
	}

	var req updateAppTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Try to find app - first by UUID, then by partner app ID
	var app *entity.App
	var err error

	appID, uuidErr := uuid.Parse(appIDStr)
	if uuidErr == nil {
		// Valid UUID - find by ID
		app, err = h.appRepo.FindByID(r.Context(), appID)
	} else {
		// Not a UUID - try as partner app ID (GID or numeric)
		// Get partner account for this user
		partnerAccount, paErr := h.partnerRepo.FindByUserID(r.Context(), user.ID)
		if paErr != nil {
			writeJSONError(w, http.StatusNotFound, "partner account not found")
			return
		}

		// Try with full GID format first
		partnerAppID := appIDStr
		if !strings.HasPrefix(partnerAppID, "gid://") {
			// If just numeric, convert to GID format
			partnerAppID = "gid://partners/App/" + appIDStr
		}

		app, err = h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, partnerAppID)
	}

	if err != nil || app == nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	// Validate and set tier
	tier := valueobject.ParseRevenueShareTier(req.RevenueShareTier)
	if !tier.IsValid() {
		writeJSONError(w, http.StatusBadRequest, "invalid revenue_share_tier")
		return
	}

	app.SetRevenueShareTier(tier)

	if err := h.appRepo.Update(r.Context(), app); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "Tier updated successfully",
		"revenue_share_tier": app.RevenueShareTier.String(),
		"display_name":       app.RevenueShareTier.DisplayName(),
		"description":        app.RevenueShareTier.Description(),
		"revenue_share_pct":  app.RevenueShareTier.RevenueSharePercent(),
	})
}

// RefreshInstallCount refreshes the install count for an app from the Partner API
// POST /api/v1/apps/{appID}/refresh-install-count
func (h *AppHandler) RefreshInstallCount(w http.ResponseWriter, r *http.Request) {
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

	// Get app ID from URL
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app_id is required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "partner account not found")
		return
	}

	// Find app
	var app *entity.App
	appID, uuidErr := uuid.Parse(appIDStr)
	if uuidErr == nil {
		app, err = h.appRepo.FindByID(r.Context(), appID)
	} else {
		// Try as partner app ID
		partnerAppID := appIDStr
		if !strings.HasPrefix(partnerAppID, "gid://") {
			partnerAppID = "gid://partners/App/" + appIDStr
		}
		app, err = h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, partnerAppID)
	}

	if err != nil || app == nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	// Decrypt access token
	decryptedToken, err := h.decryptor.Decrypt(partnerAccount.EncryptedAccessToken)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to decrypt token")
		return
	}

	// Fetch install count from Partner API
	installCount, err := h.partnerClient.FetchInstallCount(
		r.Context(),
		partnerAccount.PartnerID,
		string(decryptedToken),
		app.PartnerAppID,
	)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "failed to fetch install count from Partner API: "+err.Error())
		return
	}

	// Update app with new install count
	app.InstallCount = installCount
	if err := h.appRepo.Update(r.Context(), app); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Install count refreshed successfully",
		"install_count": installCount,
	})
}

// GetInstallCount returns the current install count for an app
// GET /api/v1/apps/{appID}/install-count
func (h *AppHandler) GetInstallCount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get app ID from URL
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app_id is required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "partner account not found")
		return
	}

	// Find app
	var app *entity.App
	appID, uuidErr := uuid.Parse(appIDStr)
	if uuidErr == nil {
		app, err = h.appRepo.FindByID(r.Context(), appID)
	} else {
		// Try as partner app ID
		partnerAppID := appIDStr
		if !strings.HasPrefix(partnerAppID, "gid://") {
			partnerAppID = "gid://partners/App/" + appIDStr
		}
		app, err = h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, partnerAppID)
	}

	if err != nil || app == nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"app_id":        extractNumericAppID(app.PartnerAppID),
		"name":          app.Name,
		"install_count": app.InstallCount,
	})
}

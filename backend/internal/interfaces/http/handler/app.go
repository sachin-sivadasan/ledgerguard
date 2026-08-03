package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// PartnerClient interface for fetching apps from Shopify Partner API
type PartnerClient interface {
	FetchApps(ctx context.Context, organizationID, accessToken string) ([]external.PartnerApp, error)
	FetchInstallCount(ctx context.Context, organizationID, accessToken, partnerAppID string) (int, error)
}

// SyncTrigger triggers an initial sync for a newly-selected app (fire-and-forget)
type SyncTrigger interface {
	TriggerSync(ctx context.Context, appID, userID, partnerAccountID uuid.UUID) error
}

type AppHandler struct {
	partnerClient PartnerClient
	partnerRepo   repository.PartnerAccountRepository
	appRepo       repository.AppRepository
	decryptor     Encryptor
	syncTrigger   SyncTrigger
	tracker       domainservice.EventTracker
}

// SetSyncTrigger sets the sync trigger for auto-syncing on app selection
func (h *AppHandler) SetSyncTrigger(st SyncTrigger) {
	h.syncTrigger = st
}

// SetTracker sets the event tracker for analytics.
func (h *AppHandler) SetTracker(t domainservice.EventTracker) {
	h.tracker = t
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
	partnerAccount, lookupErr := resolvePartnerAccount(r, h.partnerRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
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
		// Log the underlying cause (e.g. Partner API status / GraphQL error) — the 502
		// response is intentionally generic, but the reason must stay diagnosable.
		log.Printf("apps: FetchApps failed for partner %s: %v", partnerAccount.PartnerID, err)
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
	partnerAccount, lookupErr := resolvePartnerAccount(r, h.partnerRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Enforce app limit based on org plan tier
	org := middleware.OrgFromContext(r.Context())
	if org != nil {
		maxApps := org.MaxApps()
		if maxApps > 0 {
			existingApps, err := h.appRepo.FindByPartnerAccountID(r.Context(), partnerAccount.ID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to check app count")
				return
			}
			if len(existingApps) >= maxApps {
				writeJSONError(w, http.StatusForbidden,
					fmt.Sprintf("app limit reached: %s plan allows %d app(s)", org.PlanTier, maxApps))
				return
			}
		}
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

	// Track app selection event
	if h.tracker != nil {
		h.tracker.Track(r.Context(), user.ID.String(), "app_selected", domainservice.EventProperties{
			"app_name": app.Name,
			"app_id":   app.PartnerAppID,
		})
	}

	// Trigger initial sync (fire-and-forget — don't block the response)
	syncTriggered := false
	if h.syncTrigger != nil {
		if err := h.syncTrigger.TriggerSync(r.Context(), app.ID, user.ID, partnerAccount.ID); err != nil {
			log.Printf("WARNING: auto-sync trigger failed for app %s: %v", app.ID, err)
		} else {
			syncTriggered = true
		}
	}

	// Extract numeric ID for use with other endpoints
	appID := extractNumericAppID(app.PartnerAppID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "App added successfully",
		"id":                 appID,
		"uuid":               app.ID.String(),
		"name":               app.Name,
		"revenue_share_tier": app.RevenueShareTier.String(),
		"sync_triggered":     syncTriggered,
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
	partnerAccount, lookupErr := resolvePartnerAccount(r, h.partnerRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Get apps
	apps, err := h.appRepo.FindByPartnerAccountID(r.Context(), partnerAccount.ID)
	if err != nil {
		log.Printf("ListApps: failed to fetch apps for partner %s: %v", partnerAccount.ID, err)
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
			"uuid":               app.ID.String(),
			"name":               app.Name,
			"tracking_enabled":   app.TrackingEnabled,
			"revenue_share_tier": app.RevenueShareTier.String(),
			"install_count":      app.InstallCount,
			"app_store_slug":     app.AppStoreSlug,
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

	var req updateAppTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
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

// GetInstallCount returns the current install count for an app
// GET /api/v1/apps/{appID}/install-count
func (h *AppHandler) GetInstallCount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"app_id":        extractNumericAppID(app.PartnerAppID),
		"name":          app.Name,
		"install_count": app.InstallCount,
	})
}

type updateStoreSlugRequest struct {
	AppStoreSlug string `json:"app_store_slug"`
}

// appStoreSlugPattern matches a Shopify App Store slug — the last path segment of
// apps.shopify.com/<slug> — i.e. lowercase alphanumerics in hyphen-separated words
// (e.g. "zoko", "klaviyo-email-marketing"). No leading/trailing/double hyphens.
var appStoreSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// normalizeAppStoreSlug trims a user-supplied slug and, if they pasted the full app
// listing URL, extracts the slug segment (apps.shopify.com/<slug>[/...]).
func normalizeAppStoreSlug(raw string) string {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "apps.shopify.com/"); i >= 0 {
		s = s[i+len("apps.shopify.com/"):]
	}
	s = strings.Trim(s, "/")
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i] // drop any /reviews, query, or fragment
	}
	return strings.ToLower(s)
}

// UpdateStoreSlug sets the app's Shopify App Store slug — the key that unblocks review
// sync (the scraper hits apps.shopify.com/<slug>/reviews). The slug can't be fetched
// from the Partner API, so it's set manually here.
// PATCH /api/v1/apps/{appID}/store-slug   body: {"app_store_slug": "zoko"}
func (h *AppHandler) UpdateStoreSlug(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req updateStoreSlugRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slug := normalizeAppStoreSlug(req.AppStoreSlug)
	if !appStoreSlugPattern.MatchString(slug) {
		writeJSONError(w, http.StatusBadRequest,
			"invalid app_store_slug — expected the apps.shopify.com/<slug> segment (lowercase letters, digits, hyphens)")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	app.AppStoreSlug = slug
	app.UpdatedAt = time.Now().UTC()

	if err := h.appRepo.Update(r.Context(), app); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "App store slug updated — reviews will populate on the next sync",
		"app_store_slug": app.AppStoreSlug,
	})
}

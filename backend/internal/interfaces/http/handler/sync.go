package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type SyncHandler struct {
	syncService *service.SyncService
	partnerRepo repository.PartnerAccountRepository
	appRepo     repository.AppRepository
}

func NewSyncHandler(
	syncService *service.SyncService,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		partnerRepo: partnerRepo,
		appRepo:     appRepo,
	}
}

// SyncAllApps triggers sync for all user's apps
// POST /api/v1/sync
func (h *SyncHandler) SyncAllApps(w http.ResponseWriter, r *http.Request) {
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

	// Sync all apps
	results, err := h.syncService.SyncAllApps(r.Context(), partnerAccount.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "sync failed")
		return
	}

	// Convert to response format
	syncResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		res := map[string]interface{}{
			"app_id":            result.AppID.String(),
			"app_name":          result.AppName,
			"transaction_count": result.TransactionCount,
			"synced_at":         result.SyncedAt,
		}
		if result.Error != nil {
			res["error"] = result.Error.Error()
		}
		syncResults[i] = res
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Sync completed",
		"results": syncResults,
	})
}

// SyncApp triggers sync for a specific app
// POST /api/v1/sync/{appID}
func (h *SyncHandler) SyncApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appIDStr := chi.URLParam(r, "appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid app ID")
		return
	}

	// Get user's partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Tenant isolation: verify the app belongs to the user's partner account
	app, err := h.appRepo.FindByID(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	if app.PartnerAccountID != partnerAccount.ID {
		writeJSONError(w, http.StatusForbidden, "access denied")
		return
	}

	// Sync the app
	result, err := h.syncService.SyncApp(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "Sync completed",
		"app_id":            result.AppID.String(),
		"app_name":          result.AppName,
		"transaction_count": result.TransactionCount,
		"synced_at":         result.SyncedAt,
	})
}

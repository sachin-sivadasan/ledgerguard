package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/middleware"
)

// UsageStatusHandler handles usage status endpoints
type UsageStatusHandler struct {
	service *service.UsageStatusService
}

// NewUsageStatusHandler creates a new UsageStatusHandler
func NewUsageStatusHandler(svc *service.UsageStatusService) *UsageStatusHandler {
	return &UsageStatusHandler{service: svc}
}

// GetByGID retrieves a usage status by Shopify GID
// GET /v1/usage/{shopify_gid}/status
func (h *UsageStatusHandler) GetByGID(w http.ResponseWriter, r *http.Request) {
	apiKey := middleware.APIKeyFromContext(r.Context())
	if apiKey == nil {
		writeJSONError(w, http.StatusUnauthorized, "API key required")
		return
	}

	shopifyGID := chi.URLParam(r, "shopify_gid")
	if shopifyGID == "" {
		writeJSONError(w, http.StatusBadRequest, "shopify_gid is required")
		return
	}

	status, err := h.service.GetByShopifyGID(r.Context(), apiKey.UserID, shopifyGID)
	if err != nil {
		switch err {
		case service.ErrUsageNotFound:
			writeJSONError(w, http.StatusNotFound, "usage record not found")
		case service.ErrAppAccessDenied:
			writeJSONError(w, http.StatusForbidden, "access denied")
		default:
			writeJSONError(w, http.StatusInternalServerError, "failed to get usage status")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// UsageBatchRequest is the request body for batch lookups
type UsageBatchRequest struct {
	IDs []string `json:"ids"`
}

// GetBatch retrieves multiple usage statuses
// POST /v1/usage/status/batch
func (h *UsageStatusHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	apiKey := middleware.APIKeyFromContext(r.Context())
	if apiKey == nil {
		writeJSONError(w, http.StatusUnauthorized, "API key required")
		return
	}

	var req UsageBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "ids array is required")
		return
	}

	if len(req.IDs) > 100 {
		writeJSONError(w, http.StatusBadRequest, "maximum 100 IDs per batch request")
		return
	}

	resp, err := h.service.GetByShopifyGIDs(r.Context(), apiKey.UserID, req.IDs)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to get usage statuses")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

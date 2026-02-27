package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/middleware"
)

// SubscriptionStatusHandler handles subscription status endpoints
type SubscriptionStatusHandler struct {
	service *service.SubscriptionStatusService
}

// NewSubscriptionStatusHandler creates a new SubscriptionStatusHandler
func NewSubscriptionStatusHandler(svc *service.SubscriptionStatusService) *SubscriptionStatusHandler {
	return &SubscriptionStatusHandler{service: svc}
}

// GetByGID retrieves a subscription status by Shopify GID
// GET /v1/subscription/{shopify_gid}/status
func (h *SubscriptionStatusHandler) GetByGID(w http.ResponseWriter, r *http.Request) {
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
		case service.ErrSubscriptionNotFound:
			writeJSONError(w, http.StatusNotFound, "subscription not found")
		case service.ErrAppAccessDenied:
			writeJSONError(w, http.StatusForbidden, "access denied")
		default:
			writeJSONError(w, http.StatusInternalServerError, "failed to get subscription status")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status.ToResponse())
}

// GetByDomain retrieves a subscription status by myshopify domain
// GET /v1/subscription/status?domain={domain}
func (h *SubscriptionStatusHandler) GetByDomain(w http.ResponseWriter, r *http.Request) {
	apiKey := middleware.APIKeyFromContext(r.Context())
	if apiKey == nil {
		writeJSONError(w, http.StatusUnauthorized, "API key required")
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeJSONError(w, http.StatusBadRequest, "domain query parameter is required")
		return
	}

	status, err := h.service.GetByDomain(r.Context(), apiKey.UserID, domain)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			writeJSONError(w, http.StatusNotFound, "subscription not found")
		case service.ErrAppAccessDenied:
			writeJSONError(w, http.StatusForbidden, "access denied")
		default:
			writeJSONError(w, http.StatusInternalServerError, "failed to get subscription status")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status.ToResponse())
}

// BatchRequest is the request body for batch lookups
type BatchRequest struct {
	IDs []string `json:"ids"`
}

// GetBatch retrieves multiple subscription statuses
// POST /v1/subscriptions/status/batch
func (h *SubscriptionStatusHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	apiKey := middleware.APIKeyFromContext(r.Context())
	if apiKey == nil {
		writeJSONError(w, http.StatusUnauthorized, "API key required")
		return
	}

	var req BatchRequest
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
		writeJSONError(w, http.StatusInternalServerError, "failed to get subscription statuses")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

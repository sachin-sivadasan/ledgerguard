package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/application/service"
)

// APIKeyHandler handles API key management endpoints
type APIKeyHandler struct {
	service *service.APIKeyService
}

// NewAPIKeyHandler creates a new APIKeyHandler
func NewAPIKeyHandler(svc *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: svc}
}

// CreateRequest is the request body for creating an API key
type CreateRequest struct {
	Name               string `json:"name"`
	RateLimitPerMinute int    `json:"rate_limit_per_minute,omitempty"`
}

// APIKeyResponse is the response format for an API key
type APIKeyResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	KeyPrefix string  `json:"key_prefix"`
	CreatedAt string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at"`
}

// CreateResponse is the response format after creating an API key
type CreateResponse struct {
	APIKey  APIKeyResponse `json:"api_key"`
	FullKey string         `json:"full_key"`
}

// Create creates a new API key
// POST /api/v1/api-keys
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Only OWNER role can create API keys
	if user.Role != "OWNER" {
		writeJSONError(w, http.StatusForbidden, "only account owners can manage API keys")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Create(r.Context(), service.CreateKeyRequest{
		UserID:             user.ID,
		Name:               req.Name,
		RateLimitPerMinute: req.RateLimitPerMinute,
	})

	if err != nil {
		switch err {
		case service.ErrRateLimitInvalid:
			writeJSONError(w, http.StatusBadRequest, err.Error())
		default:
			writeJSONError(w, http.StatusInternalServerError, "failed to create API key")
		}
		return
	}

	// Format response for frontend
	keyPrefix := resp.RawKey
	if len(keyPrefix) > 12 {
		keyPrefix = keyPrefix[:12] + "..."
	}

	createResp := CreateResponse{
		APIKey: APIKeyResponse{
			ID:         resp.ID.String(),
			Name:       resp.Name,
			KeyPrefix:  keyPrefix,
			CreatedAt:  resp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			LastUsedAt: nil,
		},
		FullKey: resp.RawKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createResp)
}

// List returns all API keys for the authenticated user
// GET /api/v1/api-keys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Only OWNER role can list API keys
	if user.Role != "OWNER" {
		writeJSONError(w, http.StatusForbidden, "only account owners can manage API keys")
		return
	}

	keys, err := h.service.List(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	// Format keys for frontend
	apiKeys := make([]APIKeyResponse, len(keys))
	for i, k := range keys {
		// Only include active keys
		if !k.IsActive {
			continue
		}
		var lastUsed *string
		// The service doesn't currently track last_used_at, so we leave it nil
		apiKeys[i] = APIKeyResponse{
			ID:         k.ID.String(),
			Name:       k.Name,
			KeyPrefix:  "lgk_" + k.ID.String()[:8] + "...", // Use part of ID as visual prefix
			CreatedAt:  k.CreatedAt.Format("2006-01-02T15:04:05Z"),
			LastUsedAt: lastUsed,
		}
	}

	// Filter out any empty entries from inactive keys
	activeKeys := make([]APIKeyResponse, 0, len(apiKeys))
	for _, k := range apiKeys {
		if k.ID != "" {
			activeKeys = append(activeKeys, k)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"api_keys": activeKeys,
	})
}

// Revoke revokes an API key
// DELETE /api/v1/api-keys/{id}
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Only OWNER role can revoke API keys
	if user.Role != "OWNER" {
		writeJSONError(w, http.StatusForbidden, "only account owners can manage API keys")
		return
	}

	keyIDStr := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid key ID")
		return
	}

	err = h.service.Revoke(r.Context(), user.ID, keyID)
	if err != nil {
		switch err {
		case service.ErrAPIKeyNotFound:
			writeJSONError(w, http.StatusNotFound, "API key not found")
		case service.ErrUnauthorized:
			writeJSONError(w, http.StatusForbidden, "you don't own this API key")
		case service.ErrAPIKeyRevoked:
			writeJSONError(w, http.StatusConflict, "API key is already revoked")
		default:
			writeJSONError(w, http.StatusInternalServerError, "failed to revoke API key")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}

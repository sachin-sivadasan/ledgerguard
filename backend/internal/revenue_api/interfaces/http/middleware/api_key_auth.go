package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/application/service"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// APIKeyContextKey is the context key for the validated API key
	APIKeyContextKey contextKey = "api_key"
)

// ValidatedAPIKey contains info about the validated API key in context
type ValidatedAPIKey struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	RateLimitPerMinute int
}

// APIKeyFromContext retrieves the validated API key from context
func APIKeyFromContext(ctx context.Context) *ValidatedAPIKey {
	key, ok := ctx.Value(APIKeyContextKey).(*ValidatedAPIKey)
	if !ok {
		return nil
	}
	return key
}

// SetAPIKeyContext sets the validated API key in context
func SetAPIKeyContext(ctx context.Context, key *ValidatedAPIKey) context.Context {
	return context.WithValue(ctx, APIKeyContextKey, key)
}

// APIKeyAuth is middleware that validates API keys
type APIKeyAuth struct {
	keyService *service.APIKeyService
}

// NewAPIKeyAuth creates a new APIKeyAuth middleware
func NewAPIKeyAuth(keyService *service.APIKeyService) *APIKeyAuth {
	return &APIKeyAuth{keyService: keyService}
}

// Middleware returns the HTTP middleware handler
func (m *APIKeyAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header with Bearer prefix
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			writeJSONError(w, http.StatusUnauthorized, "API key required")
			return
		}

		// Validate the key
		validatedKey, err := m.keyService.ValidateKey(r.Context(), apiKey)
		if err != nil {
			switch err {
			case service.ErrAPIKeyNotFound:
				writeJSONError(w, http.StatusUnauthorized, "invalid API key")
			case service.ErrAPIKeyRevoked:
				writeJSONError(w, http.StatusUnauthorized, "API key has been revoked")
			default:
				writeJSONError(w, http.StatusInternalServerError, "authentication error")
			}
			return
		}

		// Add validated key to context
		ctx := SetAPIKeyContext(r.Context(), &ValidatedAPIKey{
			ID:                 validatedKey.ID,
			UserID:             validatedKey.UserID,
			RateLimitPerMinute: validatedKey.RateLimitPerMinute,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Simple JSON encoding to avoid import cycle
	w.Write([]byte(`{"error":{"code":"` + http.StatusText(status) + `","message":"` + message + `"}}`))
}

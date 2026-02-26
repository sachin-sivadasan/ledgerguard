package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type OAuthService interface {
	GenerateAuthURL(state string) string
	ExchangeCodeForToken(ctx context.Context, code string) (string, error)
}

type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type OAuthHandler struct {
	oauthService OAuthService
	encryptor    Encryptor
	partnerRepo  repository.PartnerAccountRepository
}

func NewOAuthHandler(oauthService OAuthService, encryptor Encryptor, partnerRepo repository.PartnerAccountRepository) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		encryptor:    encryptor,
		partnerRepo:  partnerRepo,
	}
}

// StartOAuth generates the OAuth URL and returns it to the client.
// GET /api/v1/integrations/shopify/oauth
func (h *OAuthHandler) StartOAuth(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	url := h.oauthService.GenerateAuthURL(state)

	// TODO: Store state in session/cache for verification in callback

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":   url,
		"state": state,
	})
}

// Callback handles the OAuth callback from Shopify.
// GET /api/v1/integrations/shopify/callback
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "missing code parameter")
		return
	}

	// TODO: Verify state parameter matches stored state

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Exchange code for token
	token, err := h.oauthService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "failed to exchange code for token")
		return
	}

	// Encrypt token
	encryptedToken, err := h.encryptor.Encrypt([]byte(token))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to encrypt token")
		return
	}

	// Create partner account
	// TODO: Extract partner ID from Shopify API response
	partnerID := "pending" // Will be updated after first API call
	account := entity.NewPartnerAccount(user.ID, partnerID, valueobject.IntegrationTypeOAuth, encryptedToken)

	if err := h.partnerRepo.Create(r.Context(), account); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save partner account")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Partner account connected successfully",
		"id":      account.ID.String(),
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

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

// contextWithUser is a helper for testing
func contextWithUser(ctx context.Context, user *entity.User) context.Context {
	return middleware.SetUserContext(ctx, user)
}

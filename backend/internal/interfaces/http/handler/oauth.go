package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type OAuthService interface {
	GenerateAuthURL(state string) string
	ExchangeCodeForToken(ctx context.Context, code string) (string, error)
	FetchOrganizationID(ctx context.Context, accessToken string) (string, error)
}

type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// OAuthStateStore interface for storing and validating OAuth states
type OAuthStateStore interface {
	Store(state string, userID uuid.UUID)
	Validate(state string) (uuid.UUID, bool)
}

type OAuthHandler struct {
	oauthService OAuthService
	encryptor    Encryptor
	partnerRepo  repository.PartnerAccountRepository
	userRepo     repository.UserRepository
	stateStore   OAuthStateStore
}

func NewOAuthHandler(
	oauthService OAuthService,
	encryptor Encryptor,
	partnerRepo repository.PartnerAccountRepository,
	userRepo repository.UserRepository,
	stateStore OAuthStateStore,
) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		encryptor:    encryptor,
		partnerRepo:  partnerRepo,
		userRepo:     userRepo,
		stateStore:   stateStore,
	}
}

// StartOAuth generates the OAuth URL and returns it to the client.
// GET /api/v1/integrations/shopify/oauth
func (h *OAuthHandler) StartOAuth(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	state := generateState()
	url := h.oauthService.GenerateAuthURL(state)

	// Store state with user ID for validation in callback
	h.stateStore.Store(state, user.ID)

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

	state := r.URL.Query().Get("state")
	if state == "" {
		writeJSONError(w, http.StatusBadRequest, "missing state parameter")
		return
	}

	// Validate state and get associated user ID
	userID, valid := h.stateStore.Validate(state)
	if !valid {
		writeJSONError(w, http.StatusBadRequest, "invalid or expired state parameter")
		return
	}

	// Retrieve user from database
	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	// Exchange code for token
	token, err := h.oauthService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "failed to exchange code for token")
		return
	}

	// Fetch organization ID using the token
	partnerID, err := h.oauthService.FetchOrganizationID(r.Context(), token)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "failed to fetch organization info")
		return
	}

	// Encrypt token
	encryptedToken, err := h.encryptor.Encrypt([]byte(token))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to encrypt token")
		return
	}

	// Create partner account
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
	if _, err := rand.Read(b); err != nil {
		// Cryptographic random failure is critical - panic rather than use weak state
		panic("crypto/rand failed: " + err.Error())
	}
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

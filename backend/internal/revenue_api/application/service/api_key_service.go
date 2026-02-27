package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/repository"
)

var (
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrAPIKeyRevoked    = errors.New("api key has been revoked")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrRateLimitInvalid = errors.New("rate limit must be between 1 and 1000")
)

// APIKeyService handles API key management
type APIKeyService struct {
	repo repository.APIKeyRepository
}

// NewAPIKeyService creates a new APIKeyService
func NewAPIKeyService(repo repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// CreateKeyRequest contains the request data for creating an API key
type CreateKeyRequest struct {
	UserID             uuid.UUID
	Name               string
	RateLimitPerMinute int
}

// CreateKeyResponse contains the response data after creating an API key
type CreateKeyResponse struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name,omitempty"`
	RateLimitPerMinute int       `json:"rate_limit_per_minute"`
	CreatedAt          time.Time `json:"created_at"`
	RawKey             string    `json:"api_key"` // Only returned on creation
}

// Create creates a new API key for a user
func (s *APIKeyService) Create(ctx context.Context, req CreateKeyRequest) (*CreateKeyResponse, error) {
	// Validate rate limit
	if req.RateLimitPerMinute <= 0 {
		req.RateLimitPerMinute = 60 // Default
	}
	if req.RateLimitPerMinute > 1000 {
		return nil, ErrRateLimitInvalid
	}

	// Generate new key
	keyWithRaw, err := entity.NewAPIKey(req.UserID, req.Name, req.RateLimitPerMinute)
	if err != nil {
		return nil, err
	}

	// Store the key
	if err := s.repo.Create(ctx, &keyWithRaw.APIKey); err != nil {
		return nil, err
	}

	return &CreateKeyResponse{
		ID:                 keyWithRaw.ID,
		Name:               keyWithRaw.Name,
		RateLimitPerMinute: keyWithRaw.RateLimitPerMinute,
		CreatedAt:          keyWithRaw.CreatedAt,
		RawKey:             keyWithRaw.RawKey,
	}, nil
}

// APIKeyInfo contains info about an API key (without the raw key)
type APIKeyInfo struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name,omitempty"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	CreatedAt          time.Time  `json:"created_at"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
	IsActive           bool       `json:"is_active"`
}

// List returns all API keys for a user
func (s *APIKeyService) List(ctx context.Context, userID uuid.UUID) ([]APIKeyInfo, error) {
	keys, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]APIKeyInfo, len(keys))
	for i, k := range keys {
		result[i] = APIKeyInfo{
			ID:                 k.ID,
			Name:               k.Name,
			RateLimitPerMinute: k.RateLimitPerMinute,
			CreatedAt:          k.CreatedAt,
			RevokedAt:          k.RevokedAt,
			IsActive:           k.IsActive(),
		}
	}

	return result, nil
}

// Revoke revokes an API key
func (s *APIKeyService) Revoke(ctx context.Context, userID uuid.UUID, keyID uuid.UUID) error {
	// First verify the key belongs to the user
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil {
		return ErrAPIKeyNotFound
	}

	if key.UserID != userID {
		return ErrUnauthorized
	}

	if key.IsRevoked() {
		return ErrAPIKeyRevoked
	}

	return s.repo.Revoke(ctx, keyID)
}

// ValidatedKey contains info about a validated API key
type ValidatedKey struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	RateLimitPerMinute int
}

// ValidateKey validates an API key and returns the associated user
func (s *APIKeyService) ValidateKey(ctx context.Context, rawKey string) (*ValidatedKey, error) {
	// Hash the provided key
	keyHash := entity.HashKey(rawKey)

	// Look up the key
	key, err := s.repo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	// Check if revoked
	if key.IsRevoked() {
		return nil, ErrAPIKeyRevoked
	}

	return &ValidatedKey{
		ID:                 key.ID,
		UserID:             key.UserID,
		RateLimitPerMinute: key.RateLimitPerMinute,
	}, nil
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
)

// APIKeyRepository defines the interface for API key persistence
type APIKeyRepository interface {
	// Create creates a new API key
	Create(ctx context.Context, key *entity.APIKey) error

	// GetByHash retrieves an API key by its hash (for authentication)
	GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error)

	// GetByID retrieves an API key by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error)

	// GetByUserID retrieves all API keys for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.APIKey, error)

	// GetActiveByUserID retrieves only active (non-revoked) keys for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.APIKey, error)

	// Revoke marks an API key as revoked
	Revoke(ctx context.Context, id uuid.UUID) error
}

package entity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const (
	// APIKeyPrefix is the prefix for all LedgerGuard API keys
	APIKeyPrefix = "lgk_"
	// APIKeyLength is the length of the raw key (32 bytes = 256 bits)
	APIKeyLength = 32
)

// APIKey represents an API key for Revenue API authentication
type APIKey struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	KeyHash            string // SHA-256 hash of raw key
	Name               string
	RateLimitPerMinute int
	CreatedAt          time.Time
	RevokedAt          *time.Time
}

// APIKeyWithRaw is used only during key creation to return the raw key
type APIKeyWithRaw struct {
	APIKey
	RawKey string // Only available at creation time
}

// NewAPIKey creates a new API key with a randomly generated key
func NewAPIKey(userID uuid.UUID, name string, rateLimitPerMinute int) (*APIKeyWithRaw, error) {
	// Generate random bytes for the key
	keyBytes := make([]byte, APIKeyLength)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	// Create raw key with prefix
	rawKey := APIKeyPrefix + hex.EncodeToString(keyBytes)

	// Hash the raw key
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// Set default rate limit if not provided
	if rateLimitPerMinute <= 0 {
		rateLimitPerMinute = 60
	}

	now := time.Now().UTC()

	return &APIKeyWithRaw{
		APIKey: APIKey{
			ID:                 uuid.New(),
			UserID:             userID,
			KeyHash:            keyHash,
			Name:               name,
			RateLimitPerMinute: rateLimitPerMinute,
			CreatedAt:          now,
			RevokedAt:          nil,
		},
		RawKey: rawKey,
	}, nil
}

// HashKey hashes a raw API key using SHA-256
func HashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// IsRevoked returns true if the key has been revoked
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

// IsActive returns true if the key is active (not revoked)
func (k *APIKey) IsActive() bool {
	return k.RevokedAt == nil
}

// Revoke marks the key as revoked
func (k *APIKey) Revoke() {
	now := time.Now().UTC()
	k.RevokedAt = &now
}

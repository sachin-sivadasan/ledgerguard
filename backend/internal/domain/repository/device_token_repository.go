package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// DeviceTokenRepository defines operations for device tokens
type DeviceTokenRepository interface {
	// Create creates a new device token
	Create(ctx context.Context, token *entity.DeviceToken) error

	// FindByUserID retrieves all device tokens for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.DeviceToken, error)

	// FindByToken retrieves a device token by its token string
	FindByToken(ctx context.Context, deviceToken string) (*entity.DeviceToken, error)

	// Delete removes a device token
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByToken removes a device token by its token string
	DeleteByToken(ctx context.Context, deviceToken string) error
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// NotificationPreferencesRepository defines operations for notification preferences
type NotificationPreferencesRepository interface {
	// Create creates notification preferences for a user
	Create(ctx context.Context, prefs *entity.NotificationPreferences) error

	// FindByUserID retrieves notification preferences for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error)

	// Update updates notification preferences
	Update(ctx context.Context, prefs *entity.NotificationPreferences) error

	// Upsert creates or updates notification preferences
	Upsert(ctx context.Context, prefs *entity.NotificationPreferences) error
}

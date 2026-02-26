package entity

import (
	"time"

	"github.com/google/uuid"
)

// Platform represents the device platform
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformWeb     Platform = "web"
)

// DeviceToken represents a push notification token for a user's device
type DeviceToken struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	DeviceToken string
	Platform    Platform
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDeviceToken creates a new device token
func NewDeviceToken(userID uuid.UUID, deviceToken string, platform Platform) *DeviceToken {
	now := time.Now().UTC()
	return &DeviceToken{
		ID:          uuid.New(),
		UserID:      userID,
		DeviceToken: deviceToken,
		Platform:    platform,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsValid returns true if the platform is valid
func (p Platform) IsValid() bool {
	switch p {
	case PlatformIOS, PlatformAndroid, PlatformWeb:
		return true
	default:
		return false
	}
}

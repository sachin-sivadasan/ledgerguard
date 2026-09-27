package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrPushTokenUnregistered signals that a device token is no longer valid (app
// uninstalled / token expired). Push providers return it so the notification
// service prunes the token instead of retrying. Lives here (domain) so both the
// application service and the infrastructure FCM client can reference it without
// an import cycle.
var ErrPushTokenUnregistered = errors.New("push token unregistered")

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

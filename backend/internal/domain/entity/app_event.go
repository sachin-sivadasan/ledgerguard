package entity

import (
	"time"

	"github.com/google/uuid"
)

// AppEvent represents a Shopify app lifecycle event (install, uninstall, etc.)
type AppEvent struct {
	ID             uuid.UUID
	AppID          uuid.UUID
	ShopifyShopGID string
	EventType      string
	OccurredAt     time.Time
	RawData        []byte // JSON
	CreatedAt      time.Time
}

// NewAppEvent creates a new app event
func NewAppEvent(appID uuid.UUID, shopGID, eventType string, occurredAt time.Time, rawData []byte) *AppEvent {
	now := time.Now().UTC()
	return &AppEvent{
		ID:             uuid.New(),
		AppID:          appID,
		ShopifyShopGID: shopGID,
		EventType:      eventType,
		OccurredAt:     occurredAt,
		RawData:        rawData,
		CreatedAt:      now,
	}
}

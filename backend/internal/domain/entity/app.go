package entity

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID               uuid.UUID
	PartnerAccountID uuid.UUID
	PartnerAppID     string // Shopify app GID
	Name             string
	TrackingEnabled  bool
	CreatedAt        time.Time
}

func NewApp(partnerAccountID uuid.UUID, partnerAppID, name string) *App {
	return &App{
		ID:               uuid.New(),
		PartnerAccountID: partnerAccountID,
		PartnerAppID:     partnerAppID,
		Name:             name,
		TrackingEnabled:  true,
		CreatedAt:        time.Now(),
	}
}

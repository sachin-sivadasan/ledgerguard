package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type App struct {
	ID               uuid.UUID
	PartnerAccountID uuid.UUID
	PartnerAppID     string // Shopify app GID
	Name             string
	TrackingEnabled  bool
	RevenueShareTier valueobject.RevenueShareTier // Shopify revenue share tier
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewApp(partnerAccountID uuid.UUID, partnerAppID, name string) *App {
	now := time.Now()
	return &App{
		ID:               uuid.New(),
		PartnerAccountID: partnerAccountID,
		PartnerAppID:     partnerAppID,
		Name:             name,
		TrackingEnabled:  true,
		RevenueShareTier: valueobject.RevenueShareTierDefault, // Default to 20% tier
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// SetRevenueShareTier updates the revenue share tier for this app
func (a *App) SetRevenueShareTier(tier valueobject.RevenueShareTier) {
	if tier.IsValid() {
		a.RevenueShareTier = tier
		a.UpdatedAt = time.Now()
	}
}

// CalculateFeeBreakdown calculates the expected fee breakdown for a given gross amount
func (a *App) CalculateFeeBreakdown(grossAmountCents int64, taxRate float64) valueobject.FeeBreakdown {
	return a.RevenueShareTier.CalculateFeeBreakdown(grossAmountCents, taxRate)
}

package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// PlanLabel is a developer-assigned friendly name for a price tier (billing interval +
// exact price), so plan-based reports can show "Starter" instead of the derived
// "$29.00/mo" pseudo-label. Keyed by the price tier, a mid-life price change produces a new
// tier — letting the same plan be named at each price it has had ("Starter" vs "Starter
// (old)").
type PlanLabel struct {
	ID              uuid.UUID
	AppID           uuid.UUID
	BillingInterval valueobject.BillingInterval
	PriceCents      int64
	Label           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

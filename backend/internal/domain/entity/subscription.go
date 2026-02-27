package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type Subscription struct {
	ID                      uuid.UUID
	AppID                   uuid.UUID
	ShopifyGID              string // Shopify subscription GID
	MyshopifyDomain         string
	ShopName                string // Human-readable shop name
	PlanName                string
	BasePriceCents          int64
	Currency                string
	BillingInterval         valueobject.BillingInterval
	Status                  string // ACTIVE, CANCELLED, FROZEN, PENDING
	LastRecurringChargeDate *time.Time
	ExpectedNextChargeDate  *time.Time
	RiskState               valueobject.RiskState
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func NewSubscription(
	appID uuid.UUID,
	shopifyGID string,
	myshopifyDomain string,
	shopName string,
	planName string,
	basePriceCents int64,
	currency string,
	billingInterval valueobject.BillingInterval,
) *Subscription {
	now := time.Now().UTC()
	return &Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		ShopifyGID:      shopifyGID,
		MyshopifyDomain: myshopifyDomain,
		ShopName:        shopName,
		PlanName:        planName,
		BasePriceCents:  basePriceCents,
		Currency:        currency,
		BillingInterval: billingInterval,
		Status:          "ACTIVE",
		RiskState:       valueobject.RiskStateSafe,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// UpdateFromRecurringCharge updates subscription state based on a recurring charge
func (s *Subscription) UpdateFromRecurringCharge(chargeDate time.Time, amountCents int64) {
	s.LastRecurringChargeDate = &chargeDate
	s.BasePriceCents = amountCents

	// Calculate next expected charge date
	nextCharge := s.BillingInterval.NextChargeDate(chargeDate)
	s.ExpectedNextChargeDate = &nextCharge

	s.UpdatedAt = time.Now().UTC()
}

// ClassifyRisk determines the risk state based on payment history
// This is the authoritative risk classification per CLAUDE.md
func (s *Subscription) ClassifyRisk(now time.Time) {
	// Active status with recent charge is always safe
	if s.Status == "ACTIVE" && s.ExpectedNextChargeDate != nil {
		if now.Before(*s.ExpectedNextChargeDate) || now.Equal(*s.ExpectedNextChargeDate) {
			s.RiskState = valueobject.RiskStateSafe
			return
		}
	}

	// If no expected charge date, can't classify
	if s.ExpectedNextChargeDate == nil {
		s.RiskState = valueobject.RiskStateSafe
		return
	}

	// Calculate days past due
	daysPastDue := int(now.Sub(*s.ExpectedNextChargeDate).Hours() / 24)

	switch {
	case daysPastDue <= 0:
		s.RiskState = valueobject.RiskStateSafe
	case daysPastDue <= 30:
		s.RiskState = valueobject.RiskStateSafe // Grace period
	case daysPastDue <= 60:
		s.RiskState = valueobject.RiskStateOneCycleMissed
	case daysPastDue <= 90:
		s.RiskState = valueobject.RiskStateTwoCyclesMissed
	default:
		s.RiskState = valueobject.RiskStateChurned
	}

	s.UpdatedAt = time.Now().UTC()
}

// IsActive returns true if the subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == "ACTIVE"
}

// MRRCents returns the monthly recurring revenue in cents
// For annual subscriptions, divides by 12
func (s *Subscription) MRRCents() int64 {
	if s.BillingInterval == valueobject.BillingIntervalAnnual {
		return s.BasePriceCents / 12
	}
	return s.BasePriceCents
}

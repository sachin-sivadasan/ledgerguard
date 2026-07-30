package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type Subscription struct {
	ID                      uuid.UUID
	AppID                   uuid.UUID
	ShopifyGID              string // Shopify subscription GID (gid://shopify/AppSubscription/...)
	ShopifyShopGID          string // Shopify shop GID for events lookup
	StableDomainKey         string // Deterministic ID from domain (lg_sub_...), stable across reinstalls
	MyshopifyDomain         string
	ShopName                string // Human-readable shop name
	PlanName                string
	BasePriceCents          int64
	Currency                string
	BillingInterval         valueobject.BillingInterval
	Status                  string // ACTIVE, CANCELLED, FROZEN, PENDING, UNINSTALLED
	LastRecurringChargeDate *time.Time
	ExpectedNextChargeDate  *time.Time
	RiskState               valueobject.RiskState
	// ActivatedAt is the real business subscription-start date (the earliest charge
	// date, set at rebuild). Distinct from CreatedAt, which is the record-created /
	// ingestion timestamp (reset on every ledger rebuild). Nil when unknown.
	ActivatedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // Soft delete timestamp (nil = not deleted)
}

// StartDate returns the business subscription-start date: ActivatedAt when known,
// falling back to CreatedAt (the record-created timestamp) otherwise. Use this — NOT
// CreatedAt — for "new subscription" / signup-cohort / start-date logic in reports.
func (s *Subscription) StartDate() time.Time {
	if s.ActivatedAt != nil {
		return *s.ActivatedAt
	}
	return s.CreatedAt
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

// ApplyEventStatus reconciles an event-derived status (from the Partner API app
// events) against billing reality and re-derives the risk state. Returns true if
// anything changed.
//
// The "cancel trap" defence: a terminal CANCELLED/UNINSTALLED event only churns
// the subscription when there has been NO recurring charge since that event. If a
// recurring charge is dated after the terminal event (eventAt), the cancel was a
// plan-change/stale event and the merchant is still being billed — treat as ACTIVE.
// For every non-terminal outcome the risk state is recomputed from charge recency
// (ClassifyRisk), so a status refresh can never leave a stale ACTIVE/CHURNED mix.
func (s *Subscription) ApplyEventStatus(newStatus string, eventAt, now time.Time) bool {
	if newStatus == "" {
		return false
	}

	terminal := newStatus == "UNINSTALLED" || newStatus == "CANCELLED"

	// A terminal event is overridden only when the subscription shows *current*
	// billing that post-dates it — a stale/plan-change cancel that active billing
	// has outlived. See billedSince for the recency bound and the unknown-event-time
	// fallback.
	if terminal && s.billedSince(eventAt, now) {
		newStatus = "ACTIVE"
		terminal = false
	}

	changed := false
	if s.Status != newStatus {
		s.Status = newStatus
		changed = true
	}

	oldRisk := s.RiskState
	if terminal {
		s.RiskState = valueobject.RiskStateChurned
	} else {
		// Re-derive risk from status + charge recency (fixes stale ACTIVE/CHURNED).
		s.ClassifyRisk(now)
	}
	if s.RiskState != oldRisk {
		changed = true
	}

	if changed {
		s.UpdatedAt = now
	}
	return changed
}

// billedSince reports whether the subscription shows active billing that should
// override a terminal (CANCELLED/UNINSTALLED) app event — i.e. the "cancel trap"
// where Shopify emits an old-plan cancel but the merchant keeps being charged.
//
// It requires a recurring charge that is both (a) after the terminal event, so a
// genuine cancellation with no later billing still churns, and (b) recent — within
// one billing interval + a 30-day grace of now — so a single stale trailing charge
// can't resurrect a long-dead sub. When eventAt is unknown (unparseable OccurredAt,
// zero), it falls back to charge recency alone so a missing timestamp doesn't
// silently reinstate the cancel trap.
func (s *Subscription) billedSince(eventAt, now time.Time) bool {
	if s.LastRecurringChargeDate == nil {
		return false
	}
	last := *s.LastRecurringChargeDate

	cycleDays := 30
	if s.BillingInterval == valueobject.BillingIntervalAnnual {
		cycleDays = 365
	}
	window := time.Duration(cycleDays+30) * 24 * time.Hour
	recent := now.Sub(last) <= window

	if eventAt.IsZero() {
		return recent
	}
	return last.After(eventAt) && recent
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

// SoftDelete marks the subscription as deleted without removing the record
func (s *Subscription) SoftDelete() {
	now := time.Now().UTC()
	s.DeletedAt = &now
	s.UpdatedAt = now
}

// IsDeleted returns true if the subscription has been soft-deleted
func (s *Subscription) IsDeleted() bool {
	return s.DeletedAt != nil
}

// Restore removes the soft-delete marker, making the subscription active again
func (s *Subscription) Restore() {
	s.DeletedAt = nil
	s.UpdatedAt = time.Now().UTC()
}

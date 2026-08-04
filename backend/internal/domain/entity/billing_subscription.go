package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// BillingSubscription represents a LedgerSpear B2B subscription managed via Razorpay.
// This is separate from the Shopify Subscription entity which tracks monitored app subscriptions.
type BillingSubscription struct {
	ID                     uuid.UUID
	UserID                 uuid.UUID
	RazorpaySubscriptionID string
	RazorpayPlanID         string
	RazorpayCustomerID     string
	Plan                   valueobject.BillingPlan
	Status                 valueobject.BillingSubscriptionStatus
	AmountCents            int
	Currency               string
	CurrentPeriodStart     *time.Time
	CurrentPeriodEnd       *time.Time
	ShortURL               string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// NewBillingSubscription creates a new billing subscription in CREATED status.
func NewBillingSubscription(
	userID uuid.UUID,
	razorpaySubscriptionID string,
	razorpayPlanID string,
	razorpayCustomerID string,
	plan valueobject.BillingPlan,
	shortURL string,
) *BillingSubscription {
	now := time.Now().UTC()
	return &BillingSubscription{
		ID:                     uuid.New(),
		UserID:                 userID,
		RazorpaySubscriptionID: razorpaySubscriptionID,
		RazorpayPlanID:         razorpayPlanID,
		RazorpayCustomerID:     razorpayCustomerID,
		Plan:                   plan,
		Status:                 valueobject.BillingSubscriptionStatusCreated,
		AmountCents:            plan.PriceUSDCents(),
		Currency:               "USD",
		ShortURL:               shortURL,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

// Activate transitions the subscription to ACTIVE status and sets the billing period.
func (bs *BillingSubscription) Activate(periodStart, periodEnd time.Time) {
	bs.Status = valueobject.BillingSubscriptionStatusActive
	bs.CurrentPeriodStart = &periodStart
	bs.CurrentPeriodEnd = &periodEnd
	bs.UpdatedAt = time.Now().UTC()
}

// Cancel transitions the subscription to CANCELLED status.
func (bs *BillingSubscription) Cancel() {
	bs.Status = valueobject.BillingSubscriptionStatusCancelled
	bs.UpdatedAt = time.Now().UTC()
}

// Halt transitions the subscription to HALTED status (payment failure).
func (bs *BillingSubscription) Halt() {
	bs.Status = valueobject.BillingSubscriptionStatusHalted
	bs.UpdatedAt = time.Now().UTC()
}

// MarkPending transitions the subscription to PENDING status.
func (bs *BillingSubscription) MarkPending() {
	bs.Status = valueobject.BillingSubscriptionStatusPending
	bs.UpdatedAt = time.Now().UTC()
}

// UpdatePeriod updates the current billing period (e.g., on charge).
func (bs *BillingSubscription) UpdatePeriod(periodStart, periodEnd time.Time) {
	bs.CurrentPeriodStart = &periodStart
	bs.CurrentPeriodEnd = &periodEnd
	bs.UpdatedAt = time.Now().UTC()
}

// MapToPlanTier maps the billing plan to the corresponding PlanTier value object.
func (bs *BillingSubscription) MapToPlanTier() valueobject.PlanTier {
	switch bs.Plan {
	case valueobject.BillingPlanStarter:
		return valueobject.PlanTierStarter
	case valueobject.BillingPlanPro:
		return valueobject.PlanTierPro
	default:
		return valueobject.PlanTierFree
	}
}

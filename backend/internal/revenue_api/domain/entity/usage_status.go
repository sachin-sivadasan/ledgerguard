package entity

import (
	"time"

	"github.com/google/uuid"
)

// UsageStatus represents the billing status of a usage record (CQRS read model)
type UsageStatus struct {
	ID                    uuid.UUID
	ShopifyGID            string // e.g., gid://shopify/AppUsageRecord/456
	SubscriptionShopifyGID string // Parent subscription GID
	SubscriptionID        uuid.UUID
	Billed                bool
	BillingDate           *time.Time
	AmountCents           int
	Description           string
	LastSyncedAt          time.Time
}

// NewUsageStatus creates a new usage status
func NewUsageStatus(
	shopifyGID string,
	subscriptionShopifyGID string,
	subscriptionID uuid.UUID,
	billed bool,
	billingDate *time.Time,
	amountCents int,
	description string,
) *UsageStatus {
	return &UsageStatus{
		ID:                    uuid.New(),
		ShopifyGID:            shopifyGID,
		SubscriptionShopifyGID: subscriptionShopifyGID,
		SubscriptionID:        subscriptionID,
		Billed:                billed,
		BillingDate:           billingDate,
		AmountCents:           amountCents,
		Description:           description,
		LastSyncedAt:          time.Now().UTC(),
	}
}

// UsageStatusResponse is the API response format
type UsageStatusResponse struct {
	UsageID      string                              `json:"usage_id"`
	Billed       bool                                `json:"billed"`
	BillingDate  *time.Time                          `json:"billing_date,omitempty"`
	AmountCents  int                                 `json:"amount_cents"`
	Description  string                              `json:"description,omitempty"`
	Subscription *UsageSubscriptionStatusResponse    `json:"subscription,omitempty"`
}

// UsageSubscriptionStatusResponse is the nested subscription info in usage response
type UsageSubscriptionStatusResponse struct {
	SubscriptionID     string `json:"subscription_id"`
	MyshopifyDomain    string `json:"myshopify_domain"`
	RiskState          string `json:"risk_state"`
	IsPaidCurrentCycle bool   `json:"is_paid_current_cycle"`
}

// ToResponse converts the entity to API response format
func (u *UsageStatus) ToResponse() UsageStatusResponse {
	return UsageStatusResponse{
		UsageID:     u.ShopifyGID,
		Billed:      u.Billed,
		BillingDate: u.BillingDate,
		AmountCents: u.AmountCents,
		Description: u.Description,
		// Subscription will be populated by the service layer
	}
}

// ToResponseWithSubscription creates response with nested subscription
func (u *UsageStatus) ToResponseWithSubscription(sub *SubscriptionStatus) UsageStatusResponse {
	resp := u.ToResponse()
	if sub != nil {
		resp.Subscription = &UsageSubscriptionStatusResponse{
			SubscriptionID:     sub.ShopifyGID,
			MyshopifyDomain:    sub.MyshopifyDomain,
			RiskState:          string(sub.RiskState),
			IsPaidCurrentCycle: sub.IsPaidCurrentCycle,
		}
	}
	return resp
}

// UsageStatusBatchResponse is the batch API response format
type UsageStatusBatchResponse struct {
	Results  []UsageStatusResponse `json:"results"`
	NotFound []string              `json:"not_found"`
}

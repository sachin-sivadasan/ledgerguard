package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// SubscriptionStatus represents the payment status of a subscription (CQRS read model)
type SubscriptionStatus struct {
	ID                       uuid.UUID
	ShopifyGID               string // e.g., gid://shopify/AppSubscription/123
	AppID                    uuid.UUID
	MyshopifyDomain          string
	ShopName                 string
	PlanName                 string
	RiskState                valueobject.RiskState
	IsPaidCurrentCycle       bool
	MonthsOverdue            int
	LastSuccessfulChargeDate *time.Time
	ExpectedNextChargeDate   *time.Time
	Status                   string // ACTIVE, CANCELLED, FROZEN, PENDING
	LastSyncedAt             time.Time
}

// NewSubscriptionStatus creates a new subscription status from a subscription
func NewSubscriptionStatus(
	shopifyGID string,
	appID uuid.UUID,
	myshopifyDomain string,
	shopName string,
	planName string,
	riskState valueobject.RiskState,
	status string,
	lastRecurringChargeDate *time.Time,
	expectedNextChargeDate *time.Time,
) *SubscriptionStatus {
	now := time.Now().UTC()

	// Calculate months overdue
	monthsOverdue := 0
	if expectedNextChargeDate != nil && now.After(*expectedNextChargeDate) {
		monthsOverdue = int(now.Sub(*expectedNextChargeDate).Hours() / (24 * 30))
	}

	// Determine if paid current cycle
	isPaidCurrentCycle := status == "ACTIVE" && riskState == valueobject.RiskStateSafe

	return &SubscriptionStatus{
		ID:                       uuid.New(),
		ShopifyGID:               shopifyGID,
		AppID:                    appID,
		MyshopifyDomain:          myshopifyDomain,
		ShopName:                 shopName,
		PlanName:                 planName,
		RiskState:                riskState,
		IsPaidCurrentCycle:       isPaidCurrentCycle,
		MonthsOverdue:            monthsOverdue,
		LastSuccessfulChargeDate: lastRecurringChargeDate,
		ExpectedNextChargeDate:   expectedNextChargeDate,
		Status:                   status,
		LastSyncedAt:             now,
	}
}

// SubscriptionStatusResponse is the API response format
type SubscriptionStatusResponse struct {
	SubscriptionID           string     `json:"subscription_id"`
	MyshopifyDomain          string     `json:"myshopify_domain"`
	ShopName                 string     `json:"shop_name,omitempty"`
	PlanName                 string     `json:"plan_name,omitempty"`
	RiskState                string     `json:"risk_state"`
	IsPaidCurrentCycle       bool       `json:"is_paid_current_cycle"`
	MonthsOverdue            int        `json:"months_overdue"`
	LastSuccessfulChargeDate *time.Time `json:"last_successful_charge_date,omitempty"`
	ExpectedNextChargeDate   *time.Time `json:"expected_next_charge_date,omitempty"`
	Status                   string     `json:"status"`
}

// ToResponse converts the entity to API response format
func (s *SubscriptionStatus) ToResponse() SubscriptionStatusResponse {
	return SubscriptionStatusResponse{
		SubscriptionID:           s.ShopifyGID,
		MyshopifyDomain:          s.MyshopifyDomain,
		ShopName:                 s.ShopName,
		PlanName:                 s.PlanName,
		RiskState:                string(s.RiskState),
		IsPaidCurrentCycle:       s.IsPaidCurrentCycle,
		MonthsOverdue:            s.MonthsOverdue,
		LastSuccessfulChargeDate: s.LastSuccessfulChargeDate,
		ExpectedNextChargeDate:   s.ExpectedNextChargeDate,
		Status:                   s.Status,
	}
}

// SubscriptionStatusBatchResponse is the batch API response format
type SubscriptionStatusBatchResponse struct {
	Results  []SubscriptionStatusResponse `json:"results"`
	NotFound []string                     `json:"not_found"`
}

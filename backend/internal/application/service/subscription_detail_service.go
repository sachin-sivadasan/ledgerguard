package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// SubscriptionDetail contains full subscription information with computed fields
type SubscriptionDetail struct {
	ID                      uuid.UUID                  `json:"id"`
	ShopifyGID              string                     `json:"shopify_gid"`
	ShopDomain              string                     `json:"shop_domain"`
	ShopName                string                     `json:"shop_name"`
	PlanName                string                     `json:"plan_name"`
	BasePriceCents          int64                      `json:"base_price_cents"`
	MRRCents                int64                      `json:"mrr_cents"`
	Currency                string                     `json:"currency"`
	BillingInterval         valueobject.BillingInterval `json:"billing_interval"`
	Status                  string                     `json:"status"`
	RiskState               valueobject.RiskState      `json:"risk_state"`
	LastRecurringChargeDate *time.Time                 `json:"last_recurring_charge_date,omitempty"`
	ExpectedNextChargeDate  *time.Time                 `json:"expected_next_charge_date,omitempty"`
	DaysSinceLastPayment    *int                       `json:"days_since_last_payment,omitempty"`
	DaysUntilNextPayment    *int                       `json:"days_until_next_payment,omitempty"`
	CreatedAt               time.Time                  `json:"created_at"`
	UpdatedAt               time.Time                  `json:"updated_at"`
}

// PaymentHistoryEntry represents a payment in the subscription history
type PaymentHistoryEntry struct {
	ID              uuid.UUID                `json:"id"`
	TransactionDate time.Time                `json:"transaction_date"`
	ChargeType      valueobject.ChargeType   `json:"charge_type"`
	GrossAmountCents int64                   `json:"gross_amount_cents"`
	NetAmountCents  int64                    `json:"net_amount_cents"`
	Currency        string                   `json:"currency"`
	EarningsStatus  entity.EarningsStatus    `json:"earnings_status"`
}

// RiskTimelineEntry represents a risk state change
type RiskTimelineEntry struct {
	ID            uuid.UUID             `json:"id"`
	FromRiskState string                `json:"from_risk_state"`
	ToRiskState   string                `json:"to_risk_state"`
	FromStatus    string                `json:"from_status"`
	ToStatus      string                `json:"to_status"`
	EventType     string                `json:"event_type"`
	Reason        string                `json:"reason,omitempty"`
	OccurredAt    time.Time             `json:"occurred_at"`
}

// SubscriptionDetailService provides subscription detail operations
type SubscriptionDetailService struct {
	subscriptionRepo repository.SubscriptionRepository
	transactionRepo  repository.TransactionRepository
	eventRepo        repository.SubscriptionEventRepository
}

// NewSubscriptionDetailService creates a new subscription detail service
func NewSubscriptionDetailService(
	subscriptionRepo repository.SubscriptionRepository,
	transactionRepo repository.TransactionRepository,
	eventRepo repository.SubscriptionEventRepository,
) *SubscriptionDetailService {
	return &SubscriptionDetailService{
		subscriptionRepo: subscriptionRepo,
		transactionRepo:  transactionRepo,
		eventRepo:        eventRepo,
	}
}

// GetSubscriptionDetail retrieves full subscription details
func (s *SubscriptionDetailService) GetSubscriptionDetail(ctx context.Context, subscriptionID uuid.UUID) (*SubscriptionDetail, error) {
	sub, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	detail := &SubscriptionDetail{
		ID:                      sub.ID,
		ShopifyGID:              sub.ShopifyGID,
		ShopDomain:              sub.MyshopifyDomain,
		ShopName:                sub.ShopName,
		PlanName:                sub.PlanName,
		BasePriceCents:          sub.BasePriceCents,
		MRRCents:                sub.MRRCents(),
		Currency:                sub.Currency,
		BillingInterval:         sub.BillingInterval,
		Status:                  sub.Status,
		RiskState:               sub.RiskState,
		LastRecurringChargeDate: sub.LastRecurringChargeDate,
		ExpectedNextChargeDate:  sub.ExpectedNextChargeDate,
		CreatedAt:               sub.CreatedAt,
		UpdatedAt:               sub.UpdatedAt,
	}

	// Calculate days since last payment
	if sub.LastRecurringChargeDate != nil {
		days := int(now.Sub(*sub.LastRecurringChargeDate).Hours() / 24)
		detail.DaysSinceLastPayment = &days
	}

	// Calculate days until next payment
	if sub.ExpectedNextChargeDate != nil {
		days := int(sub.ExpectedNextChargeDate.Sub(now).Hours() / 24)
		detail.DaysUntilNextPayment = &days
	}

	return detail, nil
}

// GetPaymentHistory retrieves payment history for a subscription
func (s *SubscriptionDetailService) GetPaymentHistory(ctx context.Context, subscriptionID uuid.UUID, limit int) ([]*PaymentHistoryEntry, error) {
	// First get the subscription to find its domain
	sub, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	// Get transactions for this subscription's domain
	// Look back 12 months
	to := time.Now().UTC()
	from := to.AddDate(-1, 0, 0)

	transactions, err := s.transactionRepo.FindByDomain(ctx, sub.AppID, sub.MyshopifyDomain, from, to)
	if err != nil {
		return nil, err
	}

	// Convert to history entries
	history := make([]*PaymentHistoryEntry, 0, len(transactions))
	for _, tx := range transactions {
		history = append(history, &PaymentHistoryEntry{
			ID:               tx.ID,
			TransactionDate:  tx.TransactionDate,
			ChargeType:       tx.ChargeType,
			GrossAmountCents: tx.GrossAmountCents,
			NetAmountCents:   tx.NetAmountCents,
			Currency:         tx.Currency,
			EarningsStatus:   tx.EarningsStatus,
		})
	}

	// Limit results
	if limit > 0 && len(history) > limit {
		history = history[:limit]
	}

	return history, nil
}

// GetRiskTimeline retrieves risk state changes for a subscription
func (s *SubscriptionDetailService) GetRiskTimeline(ctx context.Context, subscriptionID uuid.UUID) ([]*RiskTimelineEntry, error) {
	events, err := s.eventRepo.FindBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}

	timeline := make([]*RiskTimelineEntry, len(events))
	for i, event := range events {
		timeline[i] = &RiskTimelineEntry{
			ID:            event.ID,
			FromRiskState: string(event.FromRiskState),
			ToRiskState:   string(event.ToRiskState),
			FromStatus:    event.FromStatus,
			ToStatus:      event.ToStatus,
			EventType:     string(event.EventType),
			Reason:        event.Reason,
			OccurredAt:    event.OccurredAt,
		}
	}

	return timeline, nil
}

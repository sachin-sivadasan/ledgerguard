package service

import (
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// RiskEngine handles risk classification for subscriptions
// This is the authoritative risk classification per CLAUDE.md
type RiskEngine struct{}

func NewRiskEngine() *RiskEngine {
	return &RiskEngine{}
}

// ClassifyRisk determines the risk state based on payment history
// Risk States:
//   - SAFE: Active subscription or ≤30 days past due (grace period)
//   - ONE_CYCLE_MISSED: 31-60 days past due
//   - TWO_CYCLES_MISSED: 61-90 days past due
//   - CHURNED: >90 days past due
func (r *RiskEngine) ClassifyRisk(subscription *entity.Subscription, now time.Time) valueobject.RiskState {
	// Active status with future or current charge date is always safe
	if subscription.Status == "ACTIVE" && subscription.ExpectedNextChargeDate != nil {
		if now.Before(*subscription.ExpectedNextChargeDate) || now.Equal(*subscription.ExpectedNextChargeDate) {
			return valueobject.RiskStateSafe
		}
	}

	// If no expected charge date, can't classify - default to safe
	if subscription.ExpectedNextChargeDate == nil {
		return valueobject.RiskStateSafe
	}

	// Calculate days past due
	daysPastDue := r.DaysPastDue(subscription, now)

	return r.RiskStateFromDaysPastDue(daysPastDue)
}

// DaysPastDue calculates the number of days past the expected charge date
func (r *RiskEngine) DaysPastDue(subscription *entity.Subscription, now time.Time) int {
	if subscription.ExpectedNextChargeDate == nil {
		return 0
	}

	hours := now.Sub(*subscription.ExpectedNextChargeDate).Hours()
	if hours < 0 {
		return 0
	}
	return int(hours / 24)
}

// RiskStateFromDaysPastDue converts days past due to a risk state
func (r *RiskEngine) RiskStateFromDaysPastDue(daysPastDue int) valueobject.RiskState {
	switch {
	case daysPastDue <= 0:
		return valueobject.RiskStateSafe
	case daysPastDue <= 30:
		return valueobject.RiskStateSafe // Grace period
	case daysPastDue <= 60:
		return valueobject.RiskStateOneCycleMissed
	case daysPastDue <= 90:
		return valueobject.RiskStateTwoCyclesMissed
	default:
		return valueobject.RiskStateChurned
	}
}

// ClassifyAll classifies risk for multiple subscriptions
func (r *RiskEngine) ClassifyAll(subscriptions []*entity.Subscription, now time.Time) {
	for _, sub := range subscriptions {
		sub.RiskState = r.ClassifyRisk(sub, now)
	}
}

// CalculateRiskSummary calculates risk distribution across subscriptions
func (r *RiskEngine) CalculateRiskSummary(subscriptions []*entity.Subscription) RiskSummary {
	summary := RiskSummary{}

	for _, sub := range subscriptions {
		switch sub.RiskState {
		case valueobject.RiskStateSafe:
			summary.SafeCount++
		case valueobject.RiskStateOneCycleMissed:
			summary.OneCycleMissedCount++
		case valueobject.RiskStateTwoCyclesMissed:
			summary.TwoCyclesMissedCount++
		case valueobject.RiskStateChurned:
			summary.ChurnedCount++
		}
	}

	return summary
}

// CalculateRevenueAtRisk calculates the MRR at risk (ONE_CYCLE_MISSED + TWO_CYCLES_MISSED)
func (r *RiskEngine) CalculateRevenueAtRisk(subscriptions []*entity.Subscription) int64 {
	var atRisk int64

	for _, sub := range subscriptions {
		if sub.RiskState.IsAtRisk() {
			atRisk += sub.MRRCents()
		}
	}

	return atRisk
}

// IsAtRisk returns true if the subscription is at risk (ONE_CYCLE_MISSED or TWO_CYCLES_MISSED)
func (r *RiskEngine) IsAtRisk(subscription *entity.Subscription) bool {
	return subscription.RiskState.IsAtRisk()
}

// IsChurned returns true if the subscription has churned
func (r *RiskEngine) IsChurned(subscription *entity.Subscription) bool {
	return subscription.RiskState.IsChurned()
}

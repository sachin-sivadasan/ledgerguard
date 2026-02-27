package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// MetricsEngine computes KPIs and creates daily snapshots
type MetricsEngine struct{}

func NewMetricsEngine() *MetricsEngine {
	return &MetricsEngine{}
}

// CalculateActiveMRR computes the MRR from SAFE subscriptions only
// This represents the healthy, renewing revenue
func (m *MetricsEngine) CalculateActiveMRR(subscriptions []*entity.Subscription) int64 {
	var total int64
	for _, sub := range subscriptions {
		if sub.RiskState == valueobject.RiskStateSafe {
			total += sub.MRRCents()
		}
	}
	return total
}

// CalculateRevenueAtRisk computes MRR from ONE_CYCLE_MISSED + TWO_CYCLES_MISSED
// This is revenue that may be lost if intervention doesn't occur
func (m *MetricsEngine) CalculateRevenueAtRisk(subscriptions []*entity.Subscription) int64 {
	var total int64
	for _, sub := range subscriptions {
		if sub.RiskState.IsAtRisk() {
			total += sub.MRRCents()
		}
	}
	return total
}

// CalculateUsageRevenue computes total revenue from USAGE transactions
func (m *MetricsEngine) CalculateUsageRevenue(transactions []*entity.Transaction) int64 {
	var total int64
	for _, tx := range transactions {
		if tx.ChargeType == valueobject.ChargeTypeUsage {
			total += tx.AmountCents()
		}
	}
	return total
}

// CalculateTotalRevenue computes RECURRING + USAGE + ONE_TIME - REFUNDS
func (m *MetricsEngine) CalculateTotalRevenue(transactions []*entity.Transaction) int64 {
	var total int64
	for _, tx := range transactions {
		switch tx.ChargeType {
		case valueobject.ChargeTypeRecurring, valueobject.ChargeTypeUsage, valueobject.ChargeTypeOneTime:
			total += tx.AmountCents()
		case valueobject.ChargeTypeRefund:
			total -= tx.AmountCents()
		}
	}
	return total
}

// CalculateRenewalSuccessRate computes SAFE / Total subscriptions as a decimal
// Returns 0 if no subscriptions
func (m *MetricsEngine) CalculateRenewalSuccessRate(subscriptions []*entity.Subscription) float64 {
	if len(subscriptions) == 0 {
		return 0
	}

	safeCount := 0
	for _, sub := range subscriptions {
		if sub.RiskState == valueobject.RiskStateSafe {
			safeCount++
		}
	}

	return float64(safeCount) / float64(len(subscriptions))
}

// ComputeAllMetrics computes all KPIs and returns a DailyMetricsSnapshot
func (m *MetricsEngine) ComputeAllMetrics(
	appID uuid.UUID,
	subscriptions []*entity.Subscription,
	transactions []*entity.Transaction,
	now time.Time,
) *entity.DailyMetricsSnapshot {
	snapshot := entity.NewDailyMetricsSnapshot(appID, now)

	// Compute metrics
	activeMRR := m.CalculateActiveMRR(subscriptions)
	revenueAtRisk := m.CalculateRevenueAtRisk(subscriptions)
	usageRevenue := m.CalculateUsageRevenue(transactions)
	totalRevenue := m.CalculateTotalRevenue(transactions)
	renewalRate := m.CalculateRenewalSuccessRate(subscriptions)

	// Count by risk state
	var safeCount, oneCycleMissedCount, twoCyclesMissedCount, churnedCount int
	for _, sub := range subscriptions {
		switch sub.RiskState {
		case valueobject.RiskStateSafe:
			safeCount++
		case valueobject.RiskStateOneCycleMissed:
			oneCycleMissedCount++
		case valueobject.RiskStateTwoCyclesMissed:
			twoCyclesMissedCount++
		case valueobject.RiskStateChurned:
			churnedCount++
		}
	}

	snapshot.SetMetrics(
		activeMRR,
		revenueAtRisk,
		usageRevenue,
		totalRevenue,
		renewalRate,
		safeCount,
		oneCycleMissedCount,
		twoCyclesMissedCount,
		churnedCount,
	)

	return snapshot
}

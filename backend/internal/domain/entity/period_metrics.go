package entity

import (
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// MetricsSummary contains aggregated metrics for a time period
type MetricsSummary struct {
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
	ActiveMRRCents       int64     `json:"active_mrr_cents"`
	RevenueAtRiskCents   int64     `json:"revenue_at_risk_cents"`
	UsageRevenueCents    int64     `json:"usage_revenue_cents"`
	TotalRevenueCents    int64     `json:"total_revenue_cents"`
	RenewalSuccessRate   float64   `json:"renewal_success_rate"`
	SafeCount            int       `json:"safe_count"`
	OneCycleMissedCount  int       `json:"one_cycle_missed_count"`
	TwoCyclesMissedCount int       `json:"two_cycles_missed_count"`
	ChurnedCount         int       `json:"churned_count"`
}

// MetricsDelta contains percentage changes between periods
// Nil values indicate no previous data available for comparison
type MetricsDelta struct {
	ActiveMRRPercent       *float64 `json:"active_mrr_percent,omitempty"`
	RevenueAtRiskPercent   *float64 `json:"revenue_at_risk_percent,omitempty"`
	UsageRevenuePercent    *float64 `json:"usage_revenue_percent,omitempty"`
	TotalRevenuePercent    *float64 `json:"total_revenue_percent,omitempty"`
	RenewalSuccessPercent  *float64 `json:"renewal_success_rate_percent,omitempty"`
	ChurnCountPercent      *float64 `json:"churn_count_percent,omitempty"`
}

// DeltaSemantic indicates whether a positive delta is good or bad
type DeltaSemantic string

const (
	DeltaSemanticHigherIsGood DeltaSemantic = "HIGHER_IS_GOOD"
	DeltaSemanticLowerIsGood  DeltaSemantic = "LOWER_IS_GOOD"
)

// PeriodMetrics contains current period metrics, previous period metrics, and deltas
type PeriodMetrics struct {
	Period   valueobject.DateRange `json:"period"`
	Current  *MetricsSummary       `json:"current"`
	Previous *MetricsSummary       `json:"previous,omitempty"`
	Delta    *MetricsDelta         `json:"delta,omitempty"`
}

// NewPeriodMetrics creates a new PeriodMetrics with calculated deltas
func NewPeriodMetrics(
	dateRange valueobject.DateRange,
	current *MetricsSummary,
	previous *MetricsSummary,
) *PeriodMetrics {
	pm := &PeriodMetrics{
		Period:   dateRange,
		Current:  current,
		Previous: previous,
	}

	if current != nil && previous != nil {
		pm.Delta = calculateDeltas(current, previous)
	}

	return pm
}

// calculateDeltas computes percentage changes between two periods
func calculateDeltas(current, previous *MetricsSummary) *MetricsDelta {
	delta := &MetricsDelta{}

	// Active MRR - point-in-time, higher is good
	delta.ActiveMRRPercent = calculatePercentChange(
		float64(previous.ActiveMRRCents),
		float64(current.ActiveMRRCents),
	)

	// Revenue at Risk - point-in-time, lower is good
	delta.RevenueAtRiskPercent = calculatePercentChange(
		float64(previous.RevenueAtRiskCents),
		float64(current.RevenueAtRiskCents),
	)

	// Usage Revenue - cumulative, higher is good
	delta.UsageRevenuePercent = calculatePercentChange(
		float64(previous.UsageRevenueCents),
		float64(current.UsageRevenueCents),
	)

	// Total Revenue - cumulative, higher is good
	delta.TotalRevenuePercent = calculatePercentChange(
		float64(previous.TotalRevenueCents),
		float64(current.TotalRevenueCents),
	)

	// Renewal Success Rate - point-in-time, higher is good
	delta.RenewalSuccessPercent = calculatePercentChange(
		previous.RenewalSuccessRate,
		current.RenewalSuccessRate,
	)

	// Churn Count - point-in-time, lower is good
	delta.ChurnCountPercent = calculatePercentChange(
		float64(previous.ChurnedCount),
		float64(current.ChurnedCount),
	)

	return delta
}

// calculatePercentChange calculates the percentage change from previous to current
// Returns nil if previous is 0 (to avoid division by zero)
func calculatePercentChange(previous, current float64) *float64 {
	if previous == 0 {
		// If previous is 0 but current has value, it's "new" (infinite growth)
		// We return nil to indicate no comparison possible
		if current != 0 {
			return nil
		}
		// Both 0, no change
		zero := 0.0
		return &zero
	}

	change := ((current - previous) / previous) * 100
	return &change
}

// IsPositive returns true if the delta value is positive (for display purposes)
func (d *MetricsDelta) IsPositive(field string) *bool {
	var val *float64
	switch field {
	case "active_mrr":
		val = d.ActiveMRRPercent
	case "revenue_at_risk":
		val = d.RevenueAtRiskPercent
	case "usage_revenue":
		val = d.UsageRevenuePercent
	case "total_revenue":
		val = d.TotalRevenuePercent
	case "renewal_success":
		val = d.RenewalSuccessPercent
	case "churn_count":
		val = d.ChurnCountPercent
	default:
		return nil
	}

	if val == nil {
		return nil
	}

	isPos := *val > 0
	return &isPos
}

// IsGood returns true if the delta is considered "good" for the given metric
func (d *MetricsDelta) IsGood(field string) *bool {
	var val *float64
	var higherIsGood bool

	switch field {
	case "active_mrr":
		val = d.ActiveMRRPercent
		higherIsGood = true
	case "revenue_at_risk":
		val = d.RevenueAtRiskPercent
		higherIsGood = false // Lower is good
	case "usage_revenue":
		val = d.UsageRevenuePercent
		higherIsGood = true
	case "total_revenue":
		val = d.TotalRevenuePercent
		higherIsGood = true
	case "renewal_success":
		val = d.RenewalSuccessPercent
		higherIsGood = true
	case "churn_count":
		val = d.ChurnCountPercent
		higherIsGood = false // Lower is good
	default:
		return nil
	}

	if val == nil {
		return nil
	}

	var isGood bool
	if higherIsGood {
		isGood = *val >= 0
	} else {
		isGood = *val <= 0
	}
	return &isGood
}

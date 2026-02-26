package entity

import (
	"time"

	"github.com/google/uuid"
)

// DailyMetricsSnapshot represents a daily snapshot of app metrics
// These are immutable audit records - never deleted
type DailyMetricsSnapshot struct {
	ID                 uuid.UUID
	AppID              uuid.UUID
	Date               time.Time // Date of snapshot (truncated to day)
	ActiveMRRCents     int64     // MRR from SAFE subscriptions
	RevenueAtRiskCents int64     // MRR from ONE_CYCLE_MISSED + TWO_CYCLES_MISSED
	UsageRevenueCents  int64     // Sum of USAGE transactions (12-month window)
	TotalRevenueCents  int64     // RECURRING + USAGE + ONE_TIME - REFUNDS
	RenewalSuccessRate float64   // SAFE / (SAFE + at-risk + churned) as decimal
	SafeCount          int       // Subscriptions in SAFE state
	OneCycleMissedCount int      // Subscriptions in ONE_CYCLE_MISSED state
	TwoCyclesMissedCount int     // Subscriptions in TWO_CYCLES_MISSED state
	ChurnedCount       int       // Subscriptions in CHURNED state
	TotalSubscriptions int       // Total subscription count
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewDailyMetricsSnapshot creates a new daily metrics snapshot
func NewDailyMetricsSnapshot(appID uuid.UUID, date time.Time) *DailyMetricsSnapshot {
	now := time.Now().UTC()
	// Truncate date to start of day
	truncatedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	return &DailyMetricsSnapshot{
		ID:        uuid.New(),
		AppID:     appID,
		Date:      truncatedDate,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetMetrics sets all metric values
func (s *DailyMetricsSnapshot) SetMetrics(
	activeMRR int64,
	revenueAtRisk int64,
	usageRevenue int64,
	totalRevenue int64,
	renewalSuccessRate float64,
	safeCount int,
	oneCycleMissedCount int,
	twoCyclesMissedCount int,
	churnedCount int,
) {
	s.ActiveMRRCents = activeMRR
	s.RevenueAtRiskCents = revenueAtRisk
	s.UsageRevenueCents = usageRevenue
	s.TotalRevenueCents = totalRevenue
	s.RenewalSuccessRate = renewalSuccessRate
	s.SafeCount = safeCount
	s.OneCycleMissedCount = oneCycleMissedCount
	s.TwoCyclesMissedCount = twoCyclesMissedCount
	s.ChurnedCount = churnedCount
	s.TotalSubscriptions = safeCount + oneCycleMissedCount + twoCyclesMissedCount + churnedCount
	s.UpdatedAt = time.Now().UTC()
}

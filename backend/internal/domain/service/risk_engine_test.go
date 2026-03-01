package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestRiskEngine_ClassifyRisk_Safe_ActiveWithFutureCharge(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	futureCharge := now.AddDate(0, 0, 5) // 5 days from now

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &futureCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateSafe {
		t.Errorf("expected SAFE, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Safe_GracePeriod(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// 15 days past due (within 30-day grace period)
	pastCharge := now.AddDate(0, 0, -15)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateSafe {
		t.Errorf("expected SAFE (grace period), got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_OneCycleMissed(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// 45 days past due (31-60 range)
	pastCharge := now.AddDate(0, 0, -45)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateOneCycleMissed {
		t.Errorf("expected ONE_CYCLE_MISSED, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_TwoCyclesMissed(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// 75 days past due (61-90 range)
	pastCharge := now.AddDate(0, 0, -75)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateTwoCyclesMissed {
		t.Errorf("expected TWO_CYCLES_MISSED, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Churned(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// 120 days past due (>90 range)
	pastCharge := now.AddDate(0, 0, -120)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateChurned {
		t.Errorf("expected CHURNED, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_NoExpectedChargeDate(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: nil,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateSafe {
		t.Errorf("expected SAFE (no expected charge), got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_CancelledStatus(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	futureCharge := now.AddDate(0, 0, 5)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "CANCELLED",
		ExpectedNextChargeDate: &futureCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateChurned {
		t.Errorf("expected CHURNED for CANCELLED status, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_ExpiredStatus(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	futureCharge := now.AddDate(0, 0, 5)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "EXPIRED",
		ExpectedNextChargeDate: &futureCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateChurned {
		t.Errorf("expected CHURNED for EXPIRED status, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_FrozenStatus(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	futureCharge := now.AddDate(0, 0, 5)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "FROZEN",
		ExpectedNextChargeDate: &futureCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateOneCycleMissed {
		t.Errorf("expected ONE_CYCLE_MISSED for FROZEN status, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_PendingStatus(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "PENDING",
		ExpectedNextChargeDate: nil,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateSafe {
		t.Errorf("expected SAFE for PENDING status, got %s", result)
	}
}

// Boundary tests for risk classification thresholds
func TestRiskEngine_ClassifyRisk_Exactly30Days(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Exactly 30 days past due (boundary - should be SAFE)
	pastCharge := now.AddDate(0, 0, -30)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateSafe {
		t.Errorf("expected SAFE at exactly 30 days (grace period boundary), got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Exactly31Days(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Exactly 31 days past due (boundary - should be ONE_CYCLE_MISSED)
	pastCharge := now.AddDate(0, 0, -31)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateOneCycleMissed {
		t.Errorf("expected ONE_CYCLE_MISSED at exactly 31 days, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Exactly60Days(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Exactly 60 days past due (boundary - should be ONE_CYCLE_MISSED)
	pastCharge := now.AddDate(0, 0, -60)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateOneCycleMissed {
		t.Errorf("expected ONE_CYCLE_MISSED at exactly 60 days (boundary), got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Exactly61Days(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Exactly 61 days past due (boundary - should be TWO_CYCLES_MISSED)
	pastCharge := now.AddDate(0, 0, -61)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateTwoCyclesMissed {
		t.Errorf("expected TWO_CYCLES_MISSED at exactly 61 days, got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Exactly90Days(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Exactly 90 days past due (boundary - should be TWO_CYCLES_MISSED)
	pastCharge := now.AddDate(0, 0, -90)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateTwoCyclesMissed {
		t.Errorf("expected TWO_CYCLES_MISSED at exactly 90 days (boundary), got %s", result)
	}
}

func TestRiskEngine_ClassifyRisk_Exactly91Days(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Exactly 91 days past due (boundary - should be CHURNED)
	pastCharge := now.AddDate(0, 0, -91)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "ACTIVE",
		ExpectedNextChargeDate: &pastCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateChurned {
		t.Errorf("expected CHURNED at exactly 91 days, got %s", result)
	}
}

func TestRiskEngine_DaysPastDue(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		chargeDate   *time.Time
		expectedDays int
	}{
		{
			name:         "nil charge date",
			chargeDate:   nil,
			expectedDays: 0,
		},
		{
			name:         "future charge date",
			chargeDate:   func() *time.Time { t := now.AddDate(0, 0, 5); return &t }(),
			expectedDays: 0,
		},
		{
			name:         "today",
			chargeDate:   &now,
			expectedDays: 0,
		},
		{
			name:         "30 days ago",
			chargeDate:   func() *time.Time { t := now.AddDate(0, 0, -30); return &t }(),
			expectedDays: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &entity.Subscription{ExpectedNextChargeDate: tt.chargeDate}
			days := engine.DaysPastDue(sub, now)
			if days != tt.expectedDays {
				t.Errorf("expected %d days, got %d", tt.expectedDays, days)
			}
		})
	}
}

func TestRiskEngine_RiskStateFromDaysPastDue(t *testing.T) {
	engine := NewRiskEngine()

	tests := []struct {
		days     int
		expected valueobject.RiskState
	}{
		{-5, valueobject.RiskStateSafe},
		{0, valueobject.RiskStateSafe},
		{15, valueobject.RiskStateSafe},
		{30, valueobject.RiskStateSafe},
		{31, valueobject.RiskStateOneCycleMissed},
		{45, valueobject.RiskStateOneCycleMissed},
		{60, valueobject.RiskStateOneCycleMissed},
		{61, valueobject.RiskStateTwoCyclesMissed},
		{75, valueobject.RiskStateTwoCyclesMissed},
		{90, valueobject.RiskStateTwoCyclesMissed},
		{91, valueobject.RiskStateChurned},
		{120, valueobject.RiskStateChurned},
		{365, valueobject.RiskStateChurned},
	}

	for _, tt := range tests {
		result := engine.RiskStateFromDaysPastDue(tt.days)
		if result != tt.expected {
			t.Errorf("days=%d: expected %s, got %s", tt.days, tt.expected, result)
		}
	}
}

func TestRiskEngine_ClassifyAll(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	futureCharge := now.AddDate(0, 0, 5)
	pastCharge45 := now.AddDate(0, 0, -45)

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), Status: "ACTIVE", ExpectedNextChargeDate: &futureCharge},
		{ID: uuid.New(), Status: "ACTIVE", ExpectedNextChargeDate: &pastCharge45},
	}

	engine.ClassifyAll(subscriptions, now)

	if subscriptions[0].RiskState != valueobject.RiskStateSafe {
		t.Errorf("expected first subscription SAFE, got %s", subscriptions[0].RiskState)
	}

	if subscriptions[1].RiskState != valueobject.RiskStateOneCycleMissed {
		t.Errorf("expected second subscription ONE_CYCLE_MISSED, got %s", subscriptions[1].RiskState)
	}
}

func TestRiskEngine_CalculateRiskSummary(t *testing.T) {
	engine := NewRiskEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), RiskState: valueobject.RiskStateOneCycleMissed},
		{ID: uuid.New(), RiskState: valueobject.RiskStateTwoCyclesMissed},
		{ID: uuid.New(), RiskState: valueobject.RiskStateChurned},
	}

	summary := engine.CalculateRiskSummary(subscriptions)

	if summary.SafeCount != 2 {
		t.Errorf("expected SafeCount=2, got %d", summary.SafeCount)
	}
	if summary.OneCycleMissedCount != 1 {
		t.Errorf("expected OneCycleMissedCount=1, got %d", summary.OneCycleMissedCount)
	}
	if summary.TwoCyclesMissedCount != 1 {
		t.Errorf("expected TwoCyclesMissedCount=1, got %d", summary.TwoCyclesMissedCount)
	}
	if summary.ChurnedCount != 1 {
		t.Errorf("expected ChurnedCount=1, got %d", summary.ChurnedCount)
	}
}

func TestRiskEngine_CalculateRevenueAtRisk(t *testing.T) {
	engine := NewRiskEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 1000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateOneCycleMissed, BasePriceCents: 2000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateTwoCyclesMissed, BasePriceCents: 3000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateChurned, BasePriceCents: 4000, BillingInterval: valueobject.BillingIntervalMonthly},
	}

	atRisk := engine.CalculateRevenueAtRisk(subscriptions)

	// Only ONE_CYCLE_MISSED (2000) + TWO_CYCLES_MISSED (3000) = 5000
	expected := int64(5000)
	if atRisk != expected {
		t.Errorf("expected revenue at risk %d, got %d", expected, atRisk)
	}
}

func TestRiskEngine_IsAtRisk(t *testing.T) {
	engine := NewRiskEngine()

	tests := []struct {
		state    valueobject.RiskState
		expected bool
	}{
		{valueobject.RiskStateSafe, false},
		{valueobject.RiskStateOneCycleMissed, true},
		{valueobject.RiskStateTwoCyclesMissed, true},
		{valueobject.RiskStateChurned, false},
	}

	for _, tt := range tests {
		sub := &entity.Subscription{RiskState: tt.state}
		result := engine.IsAtRisk(sub)
		if result != tt.expected {
			t.Errorf("state=%s: expected IsAtRisk=%v, got %v", tt.state, tt.expected, result)
		}
	}
}

// Test that CANCELLED status always returns CHURNED regardless of payment timing
func TestRiskEngine_ClassifyRisk_CancelledIgnoresPaymentTiming(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	recentCharge := now.AddDate(0, 0, -5) // 5 days ago (normally would be SAFE)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "CANCELLED",
		ExpectedNextChargeDate: &recentCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	if result != valueobject.RiskStateChurned {
		t.Errorf("CANCELLED status should be CHURNED regardless of payment timing, got %s", result)
	}
}

// Test that FROZEN status takes precedence over days calculation
func TestRiskEngine_ClassifyRisk_FrozenWithRecentPayment(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	// Payment is only 5 days late (normally SAFE in grace period)
	recentCharge := now.AddDate(0, 0, -5)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "FROZEN",
		ExpectedNextChargeDate: &recentCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	// FROZEN should be ONE_CYCLE_MISSED regardless of days
	if result != valueobject.RiskStateOneCycleMissed {
		t.Errorf("FROZEN status should be ONE_CYCLE_MISSED, got %s", result)
	}
}

// Test unknown status defaults to ACTIVE behavior
func TestRiskEngine_ClassifyRisk_UnknownStatus(t *testing.T) {
	engine := NewRiskEngine()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	futureCharge := now.AddDate(0, 0, 10)

	sub := &entity.Subscription{
		ID:                     uuid.New(),
		Status:                 "UNKNOWN_STATUS",
		ExpectedNextChargeDate: &futureCharge,
	}

	result := engine.ClassifyRisk(sub, now)

	// Unknown status should default to ACTIVE behavior (SAFE for future charge)
	if result != valueobject.RiskStateSafe {
		t.Errorf("Unknown status with future charge should be SAFE, got %s", result)
	}
}

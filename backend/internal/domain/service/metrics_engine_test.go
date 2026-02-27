package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestMetricsEngine_CalculateActiveMRR(t *testing.T) {
	engine := NewMetricsEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 1000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 2000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateOneCycleMissed, BasePriceCents: 3000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateChurned, BasePriceCents: 4000, BillingInterval: valueobject.BillingIntervalMonthly},
	}

	// Only SAFE subscriptions count toward Active MRR
	activeMRR := engine.CalculateActiveMRR(subscriptions)
	expected := int64(3000) // 1000 + 2000

	if activeMRR != expected {
		t.Errorf("expected active MRR %d, got %d", expected, activeMRR)
	}
}

func TestMetricsEngine_CalculateActiveMRR_Annual(t *testing.T) {
	engine := NewMetricsEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 12000, BillingInterval: valueobject.BillingIntervalAnnual},
	}

	// Annual subscriptions should be divided by 12
	activeMRR := engine.CalculateActiveMRR(subscriptions)
	expected := int64(1000) // 12000 / 12

	if activeMRR != expected {
		t.Errorf("expected active MRR %d, got %d", expected, activeMRR)
	}
}

func TestMetricsEngine_CalculateRevenueAtRisk(t *testing.T) {
	engine := NewMetricsEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 1000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateOneCycleMissed, BasePriceCents: 2000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateTwoCyclesMissed, BasePriceCents: 3000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateChurned, BasePriceCents: 4000, BillingInterval: valueobject.BillingIntervalMonthly},
	}

	// Only ONE_CYCLE_MISSED + TWO_CYCLES_MISSED count
	atRisk := engine.CalculateRevenueAtRisk(subscriptions)
	expected := int64(5000) // 2000 + 3000

	if atRisk != expected {
		t.Errorf("expected revenue at risk %d, got %d", expected, atRisk)
	}
}

func TestMetricsEngine_CalculateUsageRevenue(t *testing.T) {
	engine := NewMetricsEngine()

	transactions := []*entity.Transaction{
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 500},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 750},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeRecurring, NetAmountCents: 2000},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeOneTime, NetAmountCents: 100},
	}

	// Only USAGE transactions count
	usageRevenue := engine.CalculateUsageRevenue(transactions)
	expected := int64(1250) // 500 + 750

	if usageRevenue != expected {
		t.Errorf("expected usage revenue %d, got %d", expected, usageRevenue)
	}
}

func TestMetricsEngine_CalculateTotalRevenue(t *testing.T) {
	engine := NewMetricsEngine()

	transactions := []*entity.Transaction{
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeRecurring, NetAmountCents: 2000},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 500},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeOneTime, NetAmountCents: 100},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeRefund, NetAmountCents: 200},
	}

	// RECURRING + USAGE + ONE_TIME - REFUNDS
	totalRevenue := engine.CalculateTotalRevenue(transactions)
	expected := int64(2400) // 2000 + 500 + 100 - 200

	if totalRevenue != expected {
		t.Errorf("expected total revenue %d, got %d", expected, totalRevenue)
	}
}

func TestMetricsEngine_CalculateRenewalSuccessRate(t *testing.T) {
	engine := NewMetricsEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), RiskState: valueobject.RiskStateOneCycleMissed},
		{ID: uuid.New(), RiskState: valueobject.RiskStateChurned},
	}

	// SAFE / Total = 3/5 = 0.6
	rate := engine.CalculateRenewalSuccessRate(subscriptions)
	expected := 0.6

	if rate != expected {
		t.Errorf("expected renewal success rate %f, got %f", expected, rate)
	}
}

func TestMetricsEngine_CalculateRenewalSuccessRate_NoSubscriptions(t *testing.T) {
	engine := NewMetricsEngine()

	subscriptions := []*entity.Subscription{}

	// Empty should return 0
	rate := engine.CalculateRenewalSuccessRate(subscriptions)

	if rate != 0 {
		t.Errorf("expected 0 for empty subscriptions, got %f", rate)
	}
}

func TestMetricsEngine_CalculateRenewalSuccessRate_AllSafe(t *testing.T) {
	engine := NewMetricsEngine()

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe},
	}

	rate := engine.CalculateRenewalSuccessRate(subscriptions)

	if rate != 1.0 {
		t.Errorf("expected 1.0 for all safe, got %f", rate)
	}
}

func TestMetricsEngine_ComputeAllMetrics(t *testing.T) {
	engine := NewMetricsEngine()
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 1000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateSafe, BasePriceCents: 2000, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateOneCycleMissed, BasePriceCents: 1500, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateTwoCyclesMissed, BasePriceCents: 500, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), RiskState: valueobject.RiskStateChurned, BasePriceCents: 1000, BillingInterval: valueobject.BillingIntervalMonthly},
	}

	transactions := []*entity.Transaction{
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeRecurring, NetAmountCents: 5000},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 1000},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeOneTime, NetAmountCents: 200},
		{ID: uuid.New(), ChargeType: valueobject.ChargeTypeRefund, NetAmountCents: 100},
	}

	snapshot := engine.ComputeAllMetrics(appID, subscriptions, transactions, now)

	// Verify snapshot
	if snapshot.AppID != appID {
		t.Errorf("expected appID %s, got %s", appID, snapshot.AppID)
	}

	// Active MRR = 1000 + 2000 = 3000
	if snapshot.ActiveMRRCents != 3000 {
		t.Errorf("expected active MRR 3000, got %d", snapshot.ActiveMRRCents)
	}

	// Revenue at risk = 1500 + 500 = 2000
	if snapshot.RevenueAtRiskCents != 2000 {
		t.Errorf("expected revenue at risk 2000, got %d", snapshot.RevenueAtRiskCents)
	}

	// Usage revenue = 1000
	if snapshot.UsageRevenueCents != 1000 {
		t.Errorf("expected usage revenue 1000, got %d", snapshot.UsageRevenueCents)
	}

	// Total revenue = 5000 + 1000 + 200 - 100 = 6100
	if snapshot.TotalRevenueCents != 6100 {
		t.Errorf("expected total revenue 6100, got %d", snapshot.TotalRevenueCents)
	}

	// Renewal success rate = 2/5 = 0.4
	if snapshot.RenewalSuccessRate != 0.4 {
		t.Errorf("expected renewal success rate 0.4, got %f", snapshot.RenewalSuccessRate)
	}

	// Risk summary counts
	if snapshot.SafeCount != 2 {
		t.Errorf("expected safe count 2, got %d", snapshot.SafeCount)
	}
	if snapshot.OneCycleMissedCount != 1 {
		t.Errorf("expected one cycle missed count 1, got %d", snapshot.OneCycleMissedCount)
	}
	if snapshot.TwoCyclesMissedCount != 1 {
		t.Errorf("expected two cycles missed count 1, got %d", snapshot.TwoCyclesMissedCount)
	}
	if snapshot.ChurnedCount != 1 {
		t.Errorf("expected churned count 1, got %d", snapshot.ChurnedCount)
	}
	if snapshot.TotalSubscriptions != 5 {
		t.Errorf("expected total subscriptions 5, got %d", snapshot.TotalSubscriptions)
	}
}

func TestMetricsEngine_ComputeAllMetrics_EmptyInputs(t *testing.T) {
	engine := NewMetricsEngine()
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	snapshot := engine.ComputeAllMetrics(appID, nil, nil, now)

	if snapshot.ActiveMRRCents != 0 {
		t.Errorf("expected 0 active MRR, got %d", snapshot.ActiveMRRCents)
	}
	if snapshot.TotalRevenueCents != 0 {
		t.Errorf("expected 0 total revenue, got %d", snapshot.TotalRevenueCents)
	}
	if snapshot.RenewalSuccessRate != 0 {
		t.Errorf("expected 0 renewal rate, got %f", snapshot.RenewalSuccessRate)
	}
}

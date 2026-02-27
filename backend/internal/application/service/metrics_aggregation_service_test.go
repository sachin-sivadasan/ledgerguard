package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock implementation of DailyMetricsSnapshotRepository
type mockSnapshotRepo struct {
	snapshots []*entity.DailyMetricsSnapshot
	err       error
}

func (m *mockSnapshotRepo) Upsert(ctx context.Context, snapshot *entity.DailyMetricsSnapshot) error {
	return m.err
}

func (m *mockSnapshotRepo) FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyMetricsSnapshot, error) {
	for _, s := range m.snapshots {
		if s.AppID == appID && s.Date.Equal(date) {
			return s, nil
		}
	}
	return nil, m.err
}

func (m *mockSnapshotRepo) FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyMetricsSnapshot, error) {
	var result []*entity.DailyMetricsSnapshot
	for _, s := range m.snapshots {
		if s.AppID == appID && !s.Date.Before(from) && !s.Date.After(to) {
			result = append(result, s)
		}
	}
	return result, m.err
}

func (m *mockSnapshotRepo) FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyMetricsSnapshot, error) {
	var latest *entity.DailyMetricsSnapshot
	for _, s := range m.snapshots {
		if s.AppID == appID {
			if latest == nil || s.Date.After(latest.Date) {
				latest = s
			}
		}
	}
	return latest, m.err
}

// Helper to create a snapshot with all fields
func createSnapshot(appID uuid.UUID, date time.Time, activeMRR, revenueAtRisk, usageRevenue, totalRevenue int64, renewalRate float64, safe, oneCycle, twoCycle, churned int) *entity.DailyMetricsSnapshot {
	s := entity.NewDailyMetricsSnapshot(appID, date)
	s.SetMetrics(activeMRR, revenueAtRisk, usageRevenue, totalRevenue, renewalRate, safe, oneCycle, twoCycle, churned)
	return s
}

func TestGetPeriodMetrics_CurrentPeriodOnly(t *testing.T) {
	appID := uuid.New()

	// Create snapshots for current period (Feb 1-15)
	snapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), 100000, 10000, 5000, 115000, 0.90, 45, 3, 2, 0),
		createSnapshot(appID, time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), 110000, 12000, 6000, 128000, 0.91, 46, 3, 1, 0),
		createSnapshot(appID, time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), 125000, 15000, 7000, 147000, 0.92, 47, 2, 1, 0),
	}

	repo := &mockSnapshotRepo{snapshots: snapshots}
	service := NewMetricsAggregationService(repo)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := service.GetPeriodMetrics(context.Background(), appID, dateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify current period exists
	if result.Current == nil {
		t.Fatal("expected current metrics, got nil")
	}

	// Point-in-time metrics should be from latest snapshot (Feb 15)
	if result.Current.ActiveMRRCents != 125000 {
		t.Errorf("expected ActiveMRR 125000, got %d", result.Current.ActiveMRRCents)
	}
	if result.Current.RevenueAtRiskCents != 15000 {
		t.Errorf("expected RevenueAtRisk 15000, got %d", result.Current.RevenueAtRiskCents)
	}
	if result.Current.RenewalSuccessRate != 0.92 {
		t.Errorf("expected RenewalSuccessRate 0.92, got %f", result.Current.RenewalSuccessRate)
	}
	if result.Current.SafeCount != 47 {
		t.Errorf("expected SafeCount 47, got %d", result.Current.SafeCount)
	}

	// Cumulative metrics should be summed
	expectedUsageRevenue := int64(5000 + 6000 + 7000)
	if result.Current.UsageRevenueCents != expectedUsageRevenue {
		t.Errorf("expected UsageRevenue %d, got %d", expectedUsageRevenue, result.Current.UsageRevenueCents)
	}

	expectedTotalRevenue := int64(115000 + 128000 + 147000)
	if result.Current.TotalRevenueCents != expectedTotalRevenue {
		t.Errorf("expected TotalRevenue %d, got %d", expectedTotalRevenue, result.Current.TotalRevenueCents)
	}

	// Previous period should be nil (no data)
	if result.Previous != nil {
		t.Error("expected previous to be nil, got data")
	}

	// Delta should be nil (no previous data)
	if result.Delta != nil {
		t.Error("expected delta to be nil, got data")
	}
}

func TestGetPeriodMetrics_WithPreviousPeriod(t *testing.T) {
	appID := uuid.New()

	// Current period: Feb 1-15 (15 days)
	// Previous period: Jan 17-31 (15 days)
	currentSnapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), 125000, 15000, 35000, 175000, 0.92, 45, 5, 2, 3),
	}

	previousSnapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), 118000, 16300, 31150, 165450, 0.90, 44, 5, 2, 4),
	}

	allSnapshots := append(currentSnapshots, previousSnapshots...)
	repo := &mockSnapshotRepo{snapshots: allSnapshots}
	service := NewMetricsAggregationService(repo)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := service.GetPeriodMetrics(context.Background(), appID, dateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both current and previous should exist
	if result.Current == nil {
		t.Fatal("expected current metrics")
	}
	if result.Previous == nil {
		t.Fatal("expected previous metrics")
	}

	// Delta should be calculated
	if result.Delta == nil {
		t.Fatal("expected delta metrics")
	}

	// Verify delta calculations
	// Active MRR: (125000 - 118000) / 118000 * 100 ≈ 5.93%
	if result.Delta.ActiveMRRPercent == nil {
		t.Fatal("expected ActiveMRRPercent")
	}
	expectedMRRDelta := ((125000.0 - 118000.0) / 118000.0) * 100
	if !floatClose(*result.Delta.ActiveMRRPercent, expectedMRRDelta, 0.01) {
		t.Errorf("expected ActiveMRRPercent %.2f, got %.2f", expectedMRRDelta, *result.Delta.ActiveMRRPercent)
	}

	// Revenue at Risk: (15000 - 16300) / 16300 * 100 ≈ -7.98%
	if result.Delta.RevenueAtRiskPercent == nil {
		t.Fatal("expected RevenueAtRiskPercent")
	}
	expectedRiskDelta := ((15000.0 - 16300.0) / 16300.0) * 100
	if !floatClose(*result.Delta.RevenueAtRiskPercent, expectedRiskDelta, 0.01) {
		t.Errorf("expected RevenueAtRiskPercent %.2f, got %.2f", expectedRiskDelta, *result.Delta.RevenueAtRiskPercent)
	}

	// Churn Count: (3 - 4) / 4 * 100 = -25%
	if result.Delta.ChurnCountPercent == nil {
		t.Fatal("expected ChurnCountPercent")
	}
	expectedChurnDelta := ((3.0 - 4.0) / 4.0) * 100
	if !floatClose(*result.Delta.ChurnCountPercent, expectedChurnDelta, 0.01) {
		t.Errorf("expected ChurnCountPercent %.2f, got %.2f", expectedChurnDelta, *result.Delta.ChurnCountPercent)
	}
}

func TestGetPeriodMetrics_DivideByZeroProtection(t *testing.T) {
	appID := uuid.New()

	// Previous period has 0 for some metrics
	currentSnapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), 100000, 10000, 5000, 115000, 0.90, 45, 3, 2, 2),
	}

	previousSnapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), 0, 0, 0, 0, 0, 0, 0, 0, 0),
	}

	allSnapshots := append(currentSnapshots, previousSnapshots...)
	repo := &mockSnapshotRepo{snapshots: allSnapshots}
	service := NewMetricsAggregationService(repo)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := service.GetPeriodMetrics(context.Background(), appID, dateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delta should exist
	if result.Delta == nil {
		t.Fatal("expected delta")
	}

	// Metrics with 0 previous should have nil delta (can't compute percentage)
	if result.Delta.ActiveMRRPercent != nil {
		t.Errorf("expected nil ActiveMRRPercent for 0 previous, got %f", *result.Delta.ActiveMRRPercent)
	}
	if result.Delta.RevenueAtRiskPercent != nil {
		t.Errorf("expected nil RevenueAtRiskPercent for 0 previous, got %f", *result.Delta.RevenueAtRiskPercent)
	}
}

func TestGetPeriodMetrics_EmptySnapshots(t *testing.T) {
	appID := uuid.New()

	repo := &mockSnapshotRepo{snapshots: []*entity.DailyMetricsSnapshot{}}
	service := NewMetricsAggregationService(repo)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := service.GetPeriodMetrics(context.Background(), appID, dateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Current should be nil (no data)
	if result.Current != nil {
		t.Error("expected nil current for empty snapshots")
	}

	// Previous should be nil
	if result.Previous != nil {
		t.Error("expected nil previous for empty snapshots")
	}

	// Delta should be nil
	if result.Delta != nil {
		t.Error("expected nil delta for empty snapshots")
	}
}

func TestGetPeriodMetricsWithPreset(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)

	// Create snapshot for this month
	snapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), 100000, 10000, 5000, 115000, 0.90, 45, 3, 2, 0),
	}

	repo := &mockSnapshotRepo{snapshots: snapshots}
	service := NewMetricsAggregationService(repo)

	result, err := service.GetPeriodMetricsWithPreset(context.Background(), appID, valueobject.TimeRangeThisMonth, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify period is this month
	if result.Period.Start.Month() != 2 || result.Period.Start.Day() != 1 {
		t.Errorf("expected period start Feb 1, got %v", result.Period.Start)
	}
	if result.Period.End.Month() != 2 || result.Period.End.Day() != 15 {
		t.Errorf("expected period end Feb 15, got %v", result.Period.End)
	}

	if result.Current == nil {
		t.Fatal("expected current metrics")
	}
}

func TestDeltaIsGood_Semantics(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		delta    float64
		expected bool
	}{
		// Higher is good
		{"positive active_mrr is good", "active_mrr", 5.0, true},
		{"negative active_mrr is bad", "active_mrr", -5.0, false},
		{"positive renewal_success is good", "renewal_success", 2.0, true},
		{"negative renewal_success is bad", "renewal_success", -2.0, false},
		{"positive usage_revenue is good", "usage_revenue", 10.0, true},
		{"zero usage_revenue is good", "usage_revenue", 0.0, true},

		// Lower is good
		{"positive revenue_at_risk is bad", "revenue_at_risk", 5.0, false},
		{"negative revenue_at_risk is good", "revenue_at_risk", -5.0, true},
		{"positive churn_count is bad", "churn_count", 10.0, false},
		{"negative churn_count is good", "churn_count", -10.0, true},
		{"zero churn_count is good", "churn_count", 0.0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			delta := &entity.MetricsDelta{}

			// Set the appropriate field
			switch tc.field {
			case "active_mrr":
				delta.ActiveMRRPercent = &tc.delta
			case "revenue_at_risk":
				delta.RevenueAtRiskPercent = &tc.delta
			case "usage_revenue":
				delta.UsageRevenuePercent = &tc.delta
			case "total_revenue":
				delta.TotalRevenuePercent = &tc.delta
			case "renewal_success":
				delta.RenewalSuccessPercent = &tc.delta
			case "churn_count":
				delta.ChurnCountPercent = &tc.delta
			}

			result := delta.IsGood(tc.field)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if *result != tc.expected {
				t.Errorf("expected IsGood=%v for %s delta %.1f, got %v", tc.expected, tc.field, tc.delta, *result)
			}
		})
	}
}

// floatClose checks if two floats are within tolerance
func floatClose(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

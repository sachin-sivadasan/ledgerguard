package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
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

func (m *mockSnapshotRepo) UpsertBatch(ctx context.Context, snapshots []*entity.DailyMetricsSnapshot) error {
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

// Mock implementation of TransactionRepository
type mockTxRepo struct {
	transactions []*entity.Transaction
	err          error
}

func (m *mockTxRepo) Upsert(ctx context.Context, tx *entity.Transaction) error {
	return m.err
}

func (m *mockTxRepo) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	return m.err
}

func (m *mockTxRepo) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	var result []*entity.Transaction
	for _, tx := range m.transactions {
		if tx.AppID == appID && !tx.TransactionDate.Before(from) && tx.TransactionDate.Before(to) {
			result = append(result, tx)
		}
	}
	return result, m.err
}

func (m *mockTxRepo) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Transaction, error) {
	return nil, m.err
}

func (m *mockTxRepo) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	return int64(len(m.transactions)), nil
}

func (m *mockTxRepo) GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}

func (m *mockTxRepo) GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]repository.EarningsByDate, error) {
	return nil, nil
}

func (m *mockTxRepo) GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]repository.EarningsByDate, error) {
	return nil, nil
}

func (m *mockTxRepo) FindByDomain(ctx context.Context, appID uuid.UUID, domain string, from, to time.Time) ([]*entity.Transaction, error) {
	return nil, m.err
}

func (m *mockTxRepo) GetEarningsSummaryByDomain(ctx context.Context, appID uuid.UUID, domain string) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}
func (m *mockTxRepo) FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters repository.TransactionFilters) (*repository.TransactionPage, error) {
	return &repository.TransactionPage{}, nil
}

// Helper to create a snapshot with all fields
func createSnapshot(appID uuid.UUID, date time.Time, activeMRR, revenueAtRisk, usageRevenue, totalRevenue int64, renewalRate float64, safe, oneCycle, twoCycle, churned int) *entity.DailyMetricsSnapshot {
	s := entity.NewDailyMetricsSnapshot(appID, date)
	s.SetMetrics(activeMRR, revenueAtRisk, usageRevenue, totalRevenue, renewalRate, safe, oneCycle, twoCycle, churned)
	return s
}

// Helper to create transactions for a period
func createTransactions(appID uuid.UUID, date time.Time, usageAmount, recurringAmount int64) []*entity.Transaction {
	return []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ChargeType:      valueobject.ChargeTypeUsage,
			NetAmountCents:  usageAmount,
			TransactionDate: date,
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:  recurringAmount,
			TransactionDate: date,
		},
	}
}

func TestGetPeriodMetrics_CurrentPeriodOnly(t *testing.T) {
	appID := uuid.New()

	// Create snapshots for current period (Feb 1-15)
	snapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), 125000, 15000, 0, 0, 0.92, 47, 2, 1, 0),
	}

	// Create transactions for the period
	transactions := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 5000, TransactionDate: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 7000, TransactionDate: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, NetAmountCents: 100000, TransactionDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	snapshotRepo := &mockSnapshotRepo{snapshots: snapshots}
	txRepo := &mockTxRepo{transactions: transactions}
	metricsEngine := service.NewMetricsEngine()

	svc := NewMetricsAggregationService(snapshotRepo, txRepo, metricsEngine)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := svc.GetPeriodMetrics(context.Background(), appID, dateRange)
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

	// Usage revenue should be calculated from transactions (5000 + 7000)
	expectedUsageRevenue := int64(12000)
	if result.Current.UsageRevenueCents != expectedUsageRevenue {
		t.Errorf("expected UsageRevenue %d, got %d", expectedUsageRevenue, result.Current.UsageRevenueCents)
	}

	// Total revenue = usage + recurring (12000 + 100000)
	expectedTotalRevenue := int64(112000)
	if result.Current.TotalRevenueCents != expectedTotalRevenue {
		t.Errorf("expected TotalRevenue %d, got %d", expectedTotalRevenue, result.Current.TotalRevenueCents)
	}

	// Previous period should be nil (no data)
	if result.Previous != nil {
		t.Error("expected previous to be nil, got data")
	}
}

func TestGetPeriodMetrics_WithPreviousPeriod(t *testing.T) {
	appID := uuid.New()

	// Current period: Feb 1-15 (15 days)
	// Previous period: Jan 17-31 (15 days)
	currentSnapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), 125000, 15000, 0, 0, 0.92, 45, 5, 2, 3),
	}

	previousSnapshots := []*entity.DailyMetricsSnapshot{
		createSnapshot(appID, time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), 118000, 16300, 0, 0, 0.90, 44, 5, 2, 4),
	}

	allSnapshots := append(currentSnapshots, previousSnapshots...)

	// Transactions for current period
	currentTxs := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 35000, TransactionDate: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, NetAmountCents: 140000, TransactionDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	// Transactions for previous period
	previousTxs := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 31150, TransactionDate: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, NetAmountCents: 134300, TransactionDate: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)},
	}

	allTxs := append(currentTxs, previousTxs...)

	snapshotRepo := &mockSnapshotRepo{snapshots: allSnapshots}
	txRepo := &mockTxRepo{transactions: allTxs}
	metricsEngine := service.NewMetricsEngine()

	svc := NewMetricsAggregationService(snapshotRepo, txRepo, metricsEngine)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := svc.GetPeriodMetrics(context.Background(), appID, dateRange)
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

	// Verify current usage revenue from transactions
	if result.Current.UsageRevenueCents != 35000 {
		t.Errorf("expected current UsageRevenue 35000, got %d", result.Current.UsageRevenueCents)
	}

	// Verify previous usage revenue from transactions
	if result.Previous.UsageRevenueCents != 31150 {
		t.Errorf("expected previous UsageRevenue 31150, got %d", result.Previous.UsageRevenueCents)
	}

	// Delta should be calculated
	if result.Delta == nil {
		t.Fatal("expected delta metrics")
	}
}

func TestGetPeriodMetrics_EmptySnapshots(t *testing.T) {
	appID := uuid.New()

	snapshotRepo := &mockSnapshotRepo{snapshots: []*entity.DailyMetricsSnapshot{}}
	txRepo := &mockTxRepo{transactions: []*entity.Transaction{}}
	metricsEngine := service.NewMetricsEngine()

	svc := NewMetricsAggregationService(snapshotRepo, txRepo, metricsEngine)

	dateRange := valueobject.NewDateRange(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	)

	result, err := svc.GetPeriodMetrics(context.Background(), appID, dateRange)
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
		createSnapshot(appID, time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), 100000, 10000, 0, 0, 0.90, 45, 3, 2, 0),
	}

	transactions := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, NetAmountCents: 5000, TransactionDate: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)},
	}

	snapshotRepo := &mockSnapshotRepo{snapshots: snapshots}
	txRepo := &mockTxRepo{transactions: transactions}
	metricsEngine := service.NewMetricsEngine()

	svc := NewMetricsAggregationService(snapshotRepo, txRepo, metricsEngine)

	result, err := svc.GetPeriodMetricsWithPreset(context.Background(), appID, valueobject.TimeRangeThisMonth, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify period is this month
	if result.Period.Start.Month() != 2 || result.Period.Start.Day() != 1 {
		t.Errorf("expected period start Feb 1, got %v", result.Period.Start)
	}

	if result.Current == nil {
		t.Fatal("expected current metrics")
	}

	// Usage revenue should come from transactions
	if result.Current.UsageRevenueCents != 5000 {
		t.Errorf("expected UsageRevenue 5000, got %d", result.Current.UsageRevenueCents)
	}
}

func TestDownsample_Weekly(t *testing.T) {
	appID := uuid.New()

	// Create 30 daily snapshots
	snapshots := make([]*entity.DailyMetricsSnapshot, 30)
	for i := 0; i < 30; i++ {
		date := time.Date(2026, 3, 1+i, 0, 0, 0, 0, time.UTC)
		s := entity.NewDailyMetricsSnapshot(appID, date)
		s.SetMetrics(int64(100000+i*1000), 0, 0, 0, 0.9, 10, 0, 0, 0)
		snapshots[i] = s
	}

	result := DownsampleSnapshots(snapshots, valueobject.GranularityWeekly)

	// 30 days spans ~4-5 ISO weeks
	if len(result) < 4 || len(result) > 6 {
		t.Errorf("expected 4-6 weekly snapshots from 30 daily, got %d", len(result))
	}

	// Each weekly snapshot should be the last day in its ISO week
	for i := 1; i < len(result); i++ {
		if !result[i].Date.After(result[i-1].Date) {
			t.Errorf("weekly snapshots not in order: %v not after %v", result[i].Date, result[i-1].Date)
		}
	}
}

func TestDownsample_Monthly(t *testing.T) {
	appID := uuid.New()

	// Create 90 daily snapshots spanning 3 months
	snapshots := make([]*entity.DailyMetricsSnapshot, 0)
	for i := 0; i < 90; i++ {
		date := time.Date(2026, 1, 1+i, 0, 0, 0, 0, time.UTC)
		s := entity.NewDailyMetricsSnapshot(appID, date)
		s.SetMetrics(int64(100000+i*1000), 0, 0, 0, 0.9, 10, 0, 0, 0)
		snapshots = append(snapshots, s)
	}

	result := DownsampleSnapshots(snapshots, valueobject.GranularityMonthly)

	// 90 days from Jan 1 → 3 months (Jan, Feb, Mar + possibly Apr)
	if len(result) < 3 || len(result) > 4 {
		t.Errorf("expected 3-4 monthly snapshots from 90 daily, got %d", len(result))
	}
}

func TestDownsample_Daily(t *testing.T) {
	appID := uuid.New()
	snapshots := []*entity.DailyMetricsSnapshot{
		func() *entity.DailyMetricsSnapshot {
			s := entity.NewDailyMetricsSnapshot(appID, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
			s.SetMetrics(100000, 0, 0, 0, 0.9, 10, 0, 0, 0)
			return s
		}(),
		func() *entity.DailyMetricsSnapshot {
			s := entity.NewDailyMetricsSnapshot(appID, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
			s.SetMetrics(100000, 0, 0, 0, 0.9, 10, 0, 0, 0)
			return s
		}(),
	}

	result := DownsampleSnapshots(snapshots, valueobject.GranularityDaily)

	if len(result) != 2 {
		t.Errorf("expected 2 daily snapshots (passthrough), got %d", len(result))
	}
}

func TestDownsample_Empty(t *testing.T) {
	result := DownsampleSnapshots(nil, valueobject.GranularityWeekly)
	if len(result) != 0 {
		t.Errorf("expected 0 for empty input, got %d", len(result))
	}

	result = DownsampleSnapshots([]*entity.DailyMetricsSnapshot{}, valueobject.GranularityMonthly)
	if len(result) != 0 {
		t.Errorf("expected 0 for empty slice, got %d", len(result))
	}
}

func TestDownsample_PicksLastInPeriod(t *testing.T) {
	appID := uuid.New()

	// Create snapshots for a single month with increasing MRR
	snapshots := make([]*entity.DailyMetricsSnapshot, 0)
	for i := 1; i <= 31; i++ {
		date := time.Date(2026, 1, i, 0, 0, 0, 0, time.UTC)
		s := entity.NewDailyMetricsSnapshot(appID, date)
		s.SetMetrics(int64(i*1000), 0, 0, 0, 0.9, 10, 0, 0, 0)
		snapshots = append(snapshots, s)
	}

	result := DownsampleSnapshots(snapshots, valueobject.GranularityMonthly)

	// Single month → 1 snapshot
	if len(result) != 1 {
		t.Fatalf("expected 1 monthly snapshot, got %d", len(result))
	}

	// Should pick the last day (Jan 31), which has MRR = 31000
	if result[0].ActiveMRRCents != 31000 {
		t.Errorf("expected last-of-month MRR 31000, got %d", result[0].ActiveMRRCents)
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

func (m *mockTxRepo) GetTransactionSummary(_ context.Context, _ uuid.UUID, _ repository.TransactionFilters) (*repository.TransactionSummary, error) {
	return &repository.TransactionSummary{}, nil
}

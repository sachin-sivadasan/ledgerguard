package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

func generateSnapshots(count int, startMRR int64, dailyGrowth int64) []*entity.DailyMetricsSnapshot {
	appID := uuid.New()
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshots := make([]*entity.DailyMetricsSnapshot, count)
	for i := 0; i < count; i++ {
		snapshots[i] = &entity.DailyMetricsSnapshot{
			ID:             uuid.New(),
			AppID:          appID,
			Date:           baseDate.AddDate(0, 0, i),
			ActiveMRRCents: startMRR + int64(i)*dailyGrowth,
		}
	}
	return snapshots
}

func TestLinearRegressionForecast_InsufficientData(t *testing.T) {
	engine := service.NewForecastingEngine()
	snapshots := generateSnapshots(50, 500000, 100) // Only 50 points

	_, err := engine.LinearRegressionForecast(snapshots, 12, uuid.New())
	if err != service.ErrInsufficientData {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestLinearRegressionForecast_GrowingMRR(t *testing.T) {
	engine := service.NewForecastingEngine()
	appID := uuid.New()
	// 180 days of data, starting at $5000 MRR, growing $10/day
	snapshots := generateSnapshots(180, 500000, 1000)
	for i := range snapshots {
		snapshots[i].AppID = appID
	}

	result, err := engine.LinearRegressionForecast(snapshots, 12, appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Model != service.ForecastModelLinear {
		t.Errorf("expected model 'linear', got %q", result.Model)
	}
	if result.DataPointsUsed != 180 {
		t.Errorf("expected 180 data points, got %d", result.DataPointsUsed)
	}
	if len(result.Points) != 12 {
		t.Fatalf("expected 12 forecast points, got %d", len(result.Points))
	}

	// Expected MRR should be increasing month over month
	for i := 1; i < len(result.Points); i++ {
		if result.Points[i].ExpectedCents <= result.Points[i-1].ExpectedCents {
			t.Errorf("month %d expected MRR (%d) should be > month %d (%d)",
				i+1, result.Points[i].ExpectedCents, i, result.Points[i-1].ExpectedCents)
		}
	}

	// Optimistic should be > expected > pessimistic
	for i, p := range result.Points {
		if p.OptimisticCents <= p.ExpectedCents {
			t.Errorf("month %d: optimistic (%d) should be > expected (%d)",
				i+1, p.OptimisticCents, p.ExpectedCents)
		}
		if p.PessimisticCents >= p.ExpectedCents {
			t.Errorf("month %d: pessimistic (%d) should be < expected (%d)",
				i+1, p.PessimisticCents, p.ExpectedCents)
		}
	}
}

func TestLinearRegressionForecast_FlatMRR(t *testing.T) {
	engine := service.NewForecastingEngine()
	appID := uuid.New()
	// 120 days of flat MRR at $5000
	snapshots := generateSnapshots(120, 500000, 0)
	for i := range snapshots {
		snapshots[i].AppID = appID
	}

	result, err := engine.LinearRegressionForecast(snapshots, 6, appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With flat MRR, all forecast points should be approximately the same
	for i := 1; i < len(result.Points); i++ {
		diff := result.Points[i].ExpectedCents - result.Points[0].ExpectedCents
		if diff < -1000 || diff > 1000 {
			t.Errorf("flat MRR should produce stable forecast, got diff %d between months", diff)
		}
	}
}

func TestLinearRegressionForecast_DecliningMRR(t *testing.T) {
	engine := service.NewForecastingEngine()
	appID := uuid.New()
	// 120 days of declining MRR
	snapshots := generateSnapshots(120, 1000000, -500)
	for i := range snapshots {
		snapshots[i].AppID = appID
	}

	result, err := engine.LinearRegressionForecast(snapshots, 6, appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected MRR should be decreasing
	for i := 1; i < len(result.Points); i++ {
		if result.Points[i].ExpectedCents >= result.Points[i-1].ExpectedCents {
			t.Errorf("declining MRR: month %d (%d) should be < month %d (%d)",
				i+1, result.Points[i].ExpectedCents, i, result.Points[i-1].ExpectedCents)
		}
	}

	// Pessimistic should never go below 0
	for i, p := range result.Points {
		if p.PessimisticCents < 0 {
			t.Errorf("month %d: pessimistic (%d) should not be negative", i+1, p.PessimisticCents)
		}
	}
}

func TestExponentialSmoothingForecast_InsufficientData(t *testing.T) {
	engine := service.NewForecastingEngine()
	snapshots := generateSnapshots(50, 500000, 100)

	_, err := engine.ExponentialSmoothingForecast(snapshots, 12, 0.3, uuid.New())
	if err != service.ErrInsufficientData {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestExponentialSmoothingForecast_GrowingMRR(t *testing.T) {
	engine := service.NewForecastingEngine()
	appID := uuid.New()
	snapshots := generateSnapshots(180, 500000, 1000)
	for i := range snapshots {
		snapshots[i].AppID = appID
	}

	result, err := engine.ExponentialSmoothingForecast(snapshots, 12, 0.3, appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Model != service.ForecastModelExponential {
		t.Errorf("expected model 'exponential', got %q", result.Model)
	}
	if len(result.Points) != 12 {
		t.Fatalf("expected 12 forecast points, got %d", len(result.Points))
	}

	// Growing data should produce increasing forecast
	for i := 1; i < len(result.Points); i++ {
		if result.Points[i].ExpectedCents <= result.Points[i-1].ExpectedCents {
			t.Errorf("month %d expected (%d) should be > month %d (%d)",
				i+1, result.Points[i].ExpectedCents, i, result.Points[i-1].ExpectedCents)
		}
	}
}

func TestExponentialSmoothingForecast_AlphaClamping(t *testing.T) {
	engine := service.NewForecastingEngine()
	appID := uuid.New()
	snapshots := generateSnapshots(100, 500000, 500)
	for i := range snapshots {
		snapshots[i].AppID = appID
	}

	// Alpha too low should be clamped to 0.1
	result1, err := engine.ExponentialSmoothingForecast(snapshots, 3, 0.01, appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Alpha too high should be clamped to 0.9
	result2, err := engine.ExponentialSmoothingForecast(snapshots, 3, 1.5, appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should produce valid non-zero results
	if result1.Points[0].ExpectedCents <= 0 || result2.Points[0].ExpectedCents <= 0 {
		t.Error("clamped alpha should still produce valid forecasts")
	}
}

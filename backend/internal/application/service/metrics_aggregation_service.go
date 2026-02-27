package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// MetricsAggregationService aggregates daily metrics into period summaries
type MetricsAggregationService struct {
	snapshotRepo repository.DailyMetricsSnapshotRepository
}

// NewMetricsAggregationService creates a new MetricsAggregationService
func NewMetricsAggregationService(snapshotRepo repository.DailyMetricsSnapshotRepository) *MetricsAggregationService {
	return &MetricsAggregationService{
		snapshotRepo: snapshotRepo,
	}
}

// GetPeriodMetrics retrieves aggregated metrics for a date range with delta comparison
func (s *MetricsAggregationService) GetPeriodMetrics(
	ctx context.Context,
	appID uuid.UUID,
	dateRange valueobject.DateRange,
) (*entity.PeriodMetrics, error) {
	// Get snapshots for current period
	currentSnapshots, err := s.snapshotRepo.FindByAppIDRange(ctx, appID, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}

	// Calculate previous period range
	previousRange := dateRange.PreviousPeriod()

	// Get snapshots for previous period
	previousSnapshots, err := s.snapshotRepo.FindByAppIDRange(ctx, appID, previousRange.Start, previousRange.End)
	if err != nil {
		return nil, err
	}

	// Aggregate current period
	var currentSummary *entity.MetricsSummary
	if len(currentSnapshots) > 0 {
		currentSummary = s.aggregateSnapshots(currentSnapshots, dateRange)
	}

	// Aggregate previous period
	var previousSummary *entity.MetricsSummary
	if len(previousSnapshots) > 0 {
		previousSummary = s.aggregateSnapshots(previousSnapshots, previousRange)
	}

	return entity.NewPeriodMetrics(dateRange, currentSummary, previousSummary), nil
}

// aggregateSnapshots combines multiple daily snapshots into a period summary
// Uses end-of-period snapshot for point-in-time metrics, sum for cumulative metrics
func (s *MetricsAggregationService) aggregateSnapshots(
	snapshots []*entity.DailyMetricsSnapshot,
	dateRange valueobject.DateRange,
) *entity.MetricsSummary {
	if len(snapshots) == 0 {
		return nil
	}

	// Sort snapshots by date to find the latest (end-of-period)
	var latestSnapshot *entity.DailyMetricsSnapshot
	var latestDate time.Time

	// Cumulative totals
	var totalUsageRevenue int64
	var totalRevenue int64

	for _, snap := range snapshots {
		// Track latest snapshot for point-in-time metrics
		if snap.Date.After(latestDate) {
			latestDate = snap.Date
			latestSnapshot = snap
		}

		// Sum cumulative metrics
		totalUsageRevenue += snap.UsageRevenueCents
		totalRevenue += snap.TotalRevenueCents
	}

	if latestSnapshot == nil {
		return nil
	}

	return &entity.MetricsSummary{
		PeriodStart: dateRange.Start,
		PeriodEnd:   dateRange.End,
		// Point-in-time metrics from end-of-period snapshot
		ActiveMRRCents:       latestSnapshot.ActiveMRRCents,
		RevenueAtRiskCents:   latestSnapshot.RevenueAtRiskCents,
		RenewalSuccessRate:   latestSnapshot.RenewalSuccessRate,
		SafeCount:            latestSnapshot.SafeCount,
		OneCycleMissedCount:  latestSnapshot.OneCycleMissedCount,
		TwoCyclesMissedCount: latestSnapshot.TwoCyclesMissedCount,
		ChurnedCount:         latestSnapshot.ChurnedCount,
		// Cumulative metrics summed across period
		UsageRevenueCents: totalUsageRevenue,
		TotalRevenueCents: totalRevenue,
	}
}

// GetPeriodMetricsWithPreset retrieves metrics using a preset time range
func (s *MetricsAggregationService) GetPeriodMetricsWithPreset(
	ctx context.Context,
	appID uuid.UUID,
	preset valueobject.TimeRangePreset,
	now time.Time,
) (*entity.PeriodMetrics, error) {
	dateRange := valueobject.DateRangeForPreset(preset, now)
	return s.GetPeriodMetrics(ctx, appID, dateRange)
}

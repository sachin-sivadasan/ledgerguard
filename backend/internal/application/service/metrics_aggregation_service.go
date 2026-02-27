package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// MetricsAggregationService aggregates daily metrics into period summaries
type MetricsAggregationService struct {
	snapshotRepo repository.DailyMetricsSnapshotRepository
	txRepo       repository.TransactionRepository
	metrics      *service.MetricsEngine
}

// NewMetricsAggregationService creates a new MetricsAggregationService
func NewMetricsAggregationService(
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	txRepo repository.TransactionRepository,
	metrics *service.MetricsEngine,
) *MetricsAggregationService {
	return &MetricsAggregationService{
		snapshotRepo: snapshotRepo,
		txRepo:       txRepo,
		metrics:      metrics,
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

	// Get transactions for current period (for usage and total revenue)
	currentTxs, err := s.txRepo.FindByAppID(ctx, appID, dateRange.Start, dateRange.End.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	// Get transactions for previous period
	previousTxs, err := s.txRepo.FindByAppID(ctx, appID, previousRange.Start, previousRange.End.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	// Aggregate current period
	var currentSummary *entity.MetricsSummary
	if len(currentSnapshots) > 0 {
		currentSummary = s.aggregateSnapshots(currentSnapshots, currentTxs, dateRange)
	}

	// Aggregate previous period
	var previousSummary *entity.MetricsSummary
	if len(previousSnapshots) > 0 {
		previousSummary = s.aggregateSnapshots(previousSnapshots, previousTxs, previousRange)
	}

	return entity.NewPeriodMetrics(dateRange, currentSummary, previousSummary), nil
}

// aggregateSnapshots combines multiple daily snapshots into a period summary
// Uses end-of-period snapshot for point-in-time metrics
// Calculates revenue from transactions for the specific period
func (s *MetricsAggregationService) aggregateSnapshots(
	snapshots []*entity.DailyMetricsSnapshot,
	transactions []*entity.Transaction,
	dateRange valueobject.DateRange,
) *entity.MetricsSummary {
	if len(snapshots) == 0 {
		return nil
	}

	// Find the latest snapshot for point-in-time metrics
	var latestSnapshot *entity.DailyMetricsSnapshot
	var latestDate time.Time

	for _, snap := range snapshots {
		if snap.Date.After(latestDate) {
			latestDate = snap.Date
			latestSnapshot = snap
		}
	}

	if latestSnapshot == nil {
		return nil
	}

	// Calculate revenue from transactions for this specific period
	usageRevenue := s.metrics.CalculateUsageRevenue(transactions)
	totalRevenue := s.metrics.CalculateTotalRevenue(transactions)

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
		// Revenue calculated from transactions for this specific period
		UsageRevenueCents: usageRevenue,
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

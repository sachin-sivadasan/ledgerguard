package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// MetricsCacheTTL is the default TTL for cached metrics
const MetricsCacheTTL = 5 * time.Minute

// MetricsCache provides caching for daily metrics and KPIs
type MetricsCache struct {
	cache Cache
	ttl   time.Duration
}

// NewMetricsCache creates a new metrics cache
func NewMetricsCache(cache Cache) *MetricsCache {
	return &MetricsCache{
		cache: cache,
		ttl:   MetricsCacheTTL,
	}
}

// NewMetricsCacheWithTTL creates a new metrics cache with custom TTL
func NewMetricsCacheWithTTL(cache Cache, ttl time.Duration) *MetricsCache {
	return &MetricsCache{
		cache: cache,
		ttl:   ttl,
	}
}

// metricsKey generates the cache key for metrics
func metricsKey(appID uuid.UUID, date time.Time) string {
	return fmt.Sprintf("metrics:%s:%s", appID.String(), date.Format("2006-01-02"))
}

// latestMetricsKey generates the cache key for latest metrics
func latestMetricsKey(appID uuid.UUID) string {
	return fmt.Sprintf("metrics:latest:%s", appID.String())
}

// dashboardKey generates the cache key for dashboard data
func dashboardKey(appID uuid.UUID) string {
	return fmt.Sprintf("dashboard:%s", appID.String())
}

// GetDailyMetrics retrieves cached daily metrics for an app and date
func (c *MetricsCache) GetDailyMetrics(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyMetricsSnapshot, error) {
	key := metricsKey(appID, date)
	return GetJSON[*entity.DailyMetricsSnapshot](ctx, c.cache, key)
}

// SetDailyMetrics caches daily metrics for an app and date
func (c *MetricsCache) SetDailyMetrics(ctx context.Context, appID uuid.UUID, date time.Time, metrics *entity.DailyMetricsSnapshot) error {
	key := metricsKey(appID, date)
	return SetJSON(ctx, c.cache, key, metrics, c.ttl)
}

// GetLatestMetrics retrieves the cached latest metrics for an app
func (c *MetricsCache) GetLatestMetrics(ctx context.Context, appID uuid.UUID) (*entity.DailyMetricsSnapshot, error) {
	key := latestMetricsKey(appID)
	return GetJSON[*entity.DailyMetricsSnapshot](ctx, c.cache, key)
}

// SetLatestMetrics caches the latest metrics for an app
func (c *MetricsCache) SetLatestMetrics(ctx context.Context, appID uuid.UUID, metrics *entity.DailyMetricsSnapshot) error {
	key := latestMetricsKey(appID)
	return SetJSON(ctx, c.cache, key, metrics, c.ttl)
}

// InvalidateAppMetrics removes all cached metrics for an app
func (c *MetricsCache) InvalidateAppMetrics(ctx context.Context, appID uuid.UUID) error {
	// Delete latest metrics
	key := latestMetricsKey(appID)
	if err := c.cache.Delete(ctx, key); err != nil {
		return err
	}

	// Delete dashboard cache
	dashKey := dashboardKey(appID)
	return c.cache.Delete(ctx, dashKey)
}

// DashboardData represents cached dashboard data
type DashboardData struct {
	MRRCents           int64   `json:"mrr_cents"`
	TotalRevenueCents  int64   `json:"total_revenue_cents"`
	RevenueAtRiskCents int64   `json:"revenue_at_risk_cents"`
	RenewalRate        float64 `json:"renewal_rate"`
	SafeCount          int     `json:"safe_count"`
	AtRiskCount        int     `json:"at_risk_count"`
	ChurnedCount       int     `json:"churned_count"`
	TotalSubscriptions int     `json:"total_subscriptions"`
	LastUpdated        string  `json:"last_updated"`
}

// GetDashboardData retrieves cached dashboard data for an app
func (c *MetricsCache) GetDashboardData(ctx context.Context, appID uuid.UUID) (*DashboardData, error) {
	key := dashboardKey(appID)
	return GetJSON[*DashboardData](ctx, c.cache, key)
}

// SetDashboardData caches dashboard data for an app
func (c *MetricsCache) SetDashboardData(ctx context.Context, appID uuid.UUID, data *DashboardData) error {
	key := dashboardKey(appID)
	return SetJSON(ctx, c.cache, key, data, c.ttl)
}

// NewDashboardDataFromMetrics creates dashboard data from a metrics snapshot
func NewDashboardDataFromMetrics(m *entity.DailyMetricsSnapshot) *DashboardData {
	return &DashboardData{
		MRRCents:           m.ActiveMRRCents,
		TotalRevenueCents:  m.TotalRevenueCents,
		RevenueAtRiskCents: m.RevenueAtRiskCents,
		RenewalRate:        m.RenewalSuccessRate,
		SafeCount:          m.SafeCount,
		AtRiskCount:        m.OneCycleMissedCount + m.TwoCyclesMissedCount,
		ChurnedCount:       m.ChurnedCount,
		TotalSubscriptions: m.TotalSubscriptions,
		LastUpdated:        time.Now().UTC().Format(time.RFC3339),
	}
}

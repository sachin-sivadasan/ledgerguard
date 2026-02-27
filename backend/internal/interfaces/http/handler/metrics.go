package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

const appGIDPrefix = "gid://partners/App/"

// MetricsAggregator interface for aggregating metrics across periods
type MetricsAggregator interface {
	GetPeriodMetrics(ctx context.Context, appID uuid.UUID, dateRange valueobject.DateRange) (*entity.PeriodMetrics, error)
}

type MetricsHandler struct {
	aggregator  MetricsAggregator
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

func NewMetricsHandler(
	aggregator MetricsAggregator,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *MetricsHandler {
	return &MetricsHandler{
		aggregator:  aggregator,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// GetLatestMetrics returns the latest metrics for an app.
// GET /api/v1/apps/{appID}/metrics/latest
// appID is numeric (e.g., "4599915"), backend constructs full GID
func (h *MetricsHandler) GetLatestMetrics(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appID := chi.URLParam(r, "appID")
	if appID == "" {
		writeJSONError(w, http.StatusBadRequest, "app ID is required")
		return
	}

	// Construct full GID for internal use
	fullAppGID := appGIDPrefix + appID

	// TODO: Calculate real metrics from transactions using fullAppGID
	// For now, return sample metrics based on app ID
	metrics := map[string]interface{}{
		"app_id":                  fullAppGID,
		"active_mrr_cents":        125000,  // $1,250.00
		"revenue_at_risk_cents":   15000,   // $150.00
		"usage_revenue_cents":     35000,   // $350.00
		"total_revenue_cents":     175000,  // $1,750.00
		"renewal_success_rate":    0.92,    // 92%
		"safe_count":              45,
		"one_cycle_missed_count":  5,
		"two_cycles_missed_count": 2,
		"churned_count":           3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// metricsResponse represents the JSON response for period metrics
type metricsResponse struct {
	Period   periodResponse          `json:"period"`
	Current  *metricsSummaryResponse `json:"current,omitempty"`
	Previous *metricsSummaryResponse `json:"previous,omitempty"`
	Delta    *metricsDeltaResponse   `json:"delta,omitempty"`
}

type periodResponse struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type metricsSummaryResponse struct {
	ActiveMRRCents       int64   `json:"active_mrr_cents"`
	RevenueAtRiskCents   int64   `json:"revenue_at_risk_cents"`
	UsageRevenueCents    int64   `json:"usage_revenue_cents"`
	TotalRevenueCents    int64   `json:"total_revenue_cents"`
	RenewalSuccessRate   float64 `json:"renewal_success_rate"`
	SafeCount            int     `json:"safe_count"`
	OneCycleMissedCount  int     `json:"one_cycle_missed_count"`
	TwoCyclesMissedCount int     `json:"two_cycles_missed_count"`
	ChurnedCount         int     `json:"churned_count"`
}

type metricsDeltaResponse struct {
	ActiveMRRPercent      *float64 `json:"active_mrr_percent,omitempty"`
	RevenueAtRiskPercent  *float64 `json:"revenue_at_risk_percent,omitempty"`
	UsageRevenuePercent   *float64 `json:"usage_revenue_percent,omitempty"`
	TotalRevenuePercent   *float64 `json:"total_revenue_percent,omitempty"`
	RenewalSuccessPercent *float64 `json:"renewal_success_rate_percent,omitempty"`
	ChurnCountPercent     *float64 `json:"churn_count_percent,omitempty"`
}

// GetMetricsByPeriod returns aggregated metrics for a time period with delta comparison.
// GET /api/v1/apps/{appID}/metrics?start=YYYY-MM-DD&end=YYYY-MM-DD
// If start/end not provided, defaults to this month
func (h *MetricsHandler) GetMetricsByPeriod(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appID := chi.URLParam(r, "appID")
	if appID == "" {
		writeJSONError(w, http.StatusBadRequest, "app ID is required")
		return
	}

	// Construct full GID for lookup
	fullAppGID := appGIDPrefix + appID

	// Get partner account to verify access
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, "no partner account found")
		return
	}

	// Parse date range from query parameters
	now := time.Now().UTC()
	dateRange, err := h.parseDateRange(r, now)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find app by partner app ID
	app, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		// App not found - return mock data for development
		// In production, this would return 404
		h.writeMockPeriodMetrics(w, dateRange)
		return
	}

	// Check if aggregator is configured
	if h.aggregator == nil {
		// Return mock data if aggregator not configured
		h.writeMockPeriodMetrics(w, dateRange)
		return
	}

	// Get period metrics
	periodMetrics, err := h.aggregator.GetPeriodMetrics(r.Context(), app.ID, dateRange)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch metrics")
		return
	}

	// Convert to response
	resp := h.toMetricsResponse(periodMetrics)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// parseDateRange extracts and validates start/end from query params
func (h *MetricsHandler) parseDateRange(r *http.Request, now time.Time) (valueobject.DateRange, error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	// Default to this month if not provided
	if startStr == "" || endStr == "" {
		return valueobject.DateRangeForPreset(valueobject.TimeRangeThisMonth, now), nil
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return valueobject.DateRange{}, err
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return valueobject.DateRange{}, err
	}

	return valueobject.NewDateRange(start, end), nil
}

// toMetricsResponse converts domain model to JSON response
func (h *MetricsHandler) toMetricsResponse(pm *entity.PeriodMetrics) metricsResponse {
	resp := metricsResponse{
		Period: periodResponse{
			Start: pm.Period.Start.Format("2006-01-02"),
			End:   pm.Period.End.Format("2006-01-02"),
		},
	}

	if pm.Current != nil {
		resp.Current = &metricsSummaryResponse{
			ActiveMRRCents:       pm.Current.ActiveMRRCents,
			RevenueAtRiskCents:   pm.Current.RevenueAtRiskCents,
			UsageRevenueCents:    pm.Current.UsageRevenueCents,
			TotalRevenueCents:    pm.Current.TotalRevenueCents,
			RenewalSuccessRate:   pm.Current.RenewalSuccessRate,
			SafeCount:            pm.Current.SafeCount,
			OneCycleMissedCount:  pm.Current.OneCycleMissedCount,
			TwoCyclesMissedCount: pm.Current.TwoCyclesMissedCount,
			ChurnedCount:         pm.Current.ChurnedCount,
		}
	}

	if pm.Previous != nil {
		resp.Previous = &metricsSummaryResponse{
			ActiveMRRCents:       pm.Previous.ActiveMRRCents,
			RevenueAtRiskCents:   pm.Previous.RevenueAtRiskCents,
			UsageRevenueCents:    pm.Previous.UsageRevenueCents,
			TotalRevenueCents:    pm.Previous.TotalRevenueCents,
			RenewalSuccessRate:   pm.Previous.RenewalSuccessRate,
			SafeCount:            pm.Previous.SafeCount,
			OneCycleMissedCount:  pm.Previous.OneCycleMissedCount,
			TwoCyclesMissedCount: pm.Previous.TwoCyclesMissedCount,
			ChurnedCount:         pm.Previous.ChurnedCount,
		}
	}

	if pm.Delta != nil {
		resp.Delta = &metricsDeltaResponse{
			ActiveMRRPercent:      pm.Delta.ActiveMRRPercent,
			RevenueAtRiskPercent:  pm.Delta.RevenueAtRiskPercent,
			UsageRevenuePercent:   pm.Delta.UsageRevenuePercent,
			TotalRevenuePercent:   pm.Delta.TotalRevenuePercent,
			RenewalSuccessPercent: pm.Delta.RenewalSuccessPercent,
			ChurnCountPercent:     pm.Delta.ChurnCountPercent,
		}
	}

	return resp
}

// writeMockPeriodMetrics writes mock data when aggregator is not configured
func (h *MetricsHandler) writeMockPeriodMetrics(w http.ResponseWriter, dateRange valueobject.DateRange) {
	// Calculate mock delta (5.93% increase)
	activeMRRDelta := 5.93
	revenueAtRiskDelta := -8.5
	usageRevenueDelta := 12.3
	totalRevenueDelta := 8.7
	renewalSuccessDelta := 2.1
	churnDelta := -15.0

	resp := metricsResponse{
		Period: periodResponse{
			Start: dateRange.Start.Format("2006-01-02"),
			End:   dateRange.End.Format("2006-01-02"),
		},
		Current: &metricsSummaryResponse{
			ActiveMRRCents:       125000,
			RevenueAtRiskCents:   15000,
			UsageRevenueCents:    35000,
			TotalRevenueCents:    175000,
			RenewalSuccessRate:   0.92,
			SafeCount:            45,
			OneCycleMissedCount:  5,
			TwoCyclesMissedCount: 2,
			ChurnedCount:         3,
		},
		Previous: &metricsSummaryResponse{
			ActiveMRRCents:       118000,
			RevenueAtRiskCents:   16400,
			UsageRevenueCents:    31200,
			TotalRevenueCents:    161000,
			RenewalSuccessRate:   0.90,
			SafeCount:            44,
			OneCycleMissedCount:  5,
			TwoCyclesMissedCount: 2,
			ChurnedCount:         4,
		},
		Delta: &metricsDeltaResponse{
			ActiveMRRPercent:      &activeMRRDelta,
			RevenueAtRiskPercent:  &revenueAtRiskDelta,
			UsageRevenuePercent:   &usageRevenueDelta,
			TotalRevenuePercent:   &totalRevenueDelta,
			RenewalSuccessPercent: &renewalSuccessDelta,
			ChurnCountPercent:     &churnDelta,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

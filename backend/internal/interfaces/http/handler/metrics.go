package handler

import (
	"context"
	"encoding/json"
	"log"
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
		log.Printf("Failed to get period metrics for app %s: %v", app.ID, err)
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch metrics")
		return
	}

	// If no data found for the period, fall back to mock data
	if periodMetrics.Current == nil && periodMetrics.Previous == nil {
		log.Printf("No metrics data found for app %s in period %s to %s, returning mock data",
			app.ID, dateRange.Start.Format("2006-01-02"), dateRange.End.Format("2006-01-02"))
		h.writeMockPeriodMetrics(w, dateRange)
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
// Mock data varies based on the month to simulate realistic period-over-period changes
func (h *MetricsHandler) writeMockPeriodMetrics(w http.ResponseWriter, dateRange valueobject.DateRange) {
	// Use month to generate different mock values (simulates growth over time)
	month := dateRange.Start.Month()
	year := dateRange.Start.Year()

	// Base values that grow by ~5% each month
	baseMultiplier := 1.0 + (float64(month-1) * 0.05) + (float64(year-2025) * 0.60)

	currentMRR := int64(float64(100000) * baseMultiplier)
	currentAtRisk := int64(float64(12000) * baseMultiplier * 0.9) // At risk grows slower
	currentUsage := int64(float64(28000) * baseMultiplier * 1.1)  // Usage grows faster
	currentTotal := currentMRR + currentUsage
	currentSafe := int(float64(40) * baseMultiplier)
	currentOneCycle := int(float64(4) * baseMultiplier * 0.8)
	currentTwoCycle := int(float64(2) * baseMultiplier * 0.7)
	currentChurned := int(float64(3) * baseMultiplier * 0.6)
	currentRenewalRate := 0.88 + (float64(month) * 0.005) // Slowly improving
	if currentRenewalRate > 0.98 {
		currentRenewalRate = 0.98
	}

	// Previous period (approximately 5% less)
	prevMultiplier := baseMultiplier * 0.95
	prevMRR := int64(float64(100000) * prevMultiplier)
	prevAtRisk := int64(float64(12000) * prevMultiplier * 0.95)
	prevUsage := int64(float64(28000) * prevMultiplier * 1.05)
	prevTotal := prevMRR + prevUsage
	prevSafe := int(float64(40) * prevMultiplier)
	prevOneCycle := int(float64(4) * prevMultiplier * 0.85)
	prevTwoCycle := int(float64(2) * prevMultiplier * 0.75)
	prevChurned := int(float64(3) * prevMultiplier * 0.7)
	prevRenewalRate := currentRenewalRate - 0.02

	// Calculate deltas
	activeMRRDelta := (float64(currentMRR-prevMRR) / float64(prevMRR)) * 100
	revenueAtRiskDelta := (float64(currentAtRisk-prevAtRisk) / float64(prevAtRisk)) * 100
	usageRevenueDelta := (float64(currentUsage-prevUsage) / float64(prevUsage)) * 100
	totalRevenueDelta := (float64(currentTotal-prevTotal) / float64(prevTotal)) * 100
	renewalSuccessDelta := (currentRenewalRate - prevRenewalRate) / prevRenewalRate * 100
	churnDelta := 0.0
	if prevChurned > 0 {
		churnDelta = (float64(currentChurned-prevChurned) / float64(prevChurned)) * 100
	}

	resp := metricsResponse{
		Period: periodResponse{
			Start: dateRange.Start.Format("2006-01-02"),
			End:   dateRange.End.Format("2006-01-02"),
		},
		Current: &metricsSummaryResponse{
			ActiveMRRCents:       currentMRR,
			RevenueAtRiskCents:   currentAtRisk,
			UsageRevenueCents:    currentUsage,
			TotalRevenueCents:    currentTotal,
			RenewalSuccessRate:   currentRenewalRate,
			SafeCount:            currentSafe,
			OneCycleMissedCount:  currentOneCycle,
			TwoCyclesMissedCount: currentTwoCycle,
			ChurnedCount:         currentChurned,
		},
		Previous: &metricsSummaryResponse{
			ActiveMRRCents:       prevMRR,
			RevenueAtRiskCents:   prevAtRisk,
			UsageRevenueCents:    prevUsage,
			TotalRevenueCents:    prevTotal,
			RenewalSuccessRate:   prevRenewalRate,
			SafeCount:            prevSafe,
			OneCycleMissedCount:  prevOneCycle,
			TwoCyclesMissedCount: prevTwoCycle,
			ChurnedCount:         prevChurned,
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

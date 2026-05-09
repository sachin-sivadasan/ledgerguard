package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// RevenueHandler handles earnings timeline endpoints
type RevenueHandler struct {
	revenueService *service.RevenueMetricsService
	partnerRepo    repository.PartnerAccountRepository
	appRepo        repository.AppRepository
}

// NewRevenueHandler creates a new RevenueHandler
func NewRevenueHandler(
	revenueService *service.RevenueMetricsService,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *RevenueHandler {
	return &RevenueHandler{
		revenueService: revenueService,
		partnerRepo:    partnerRepo,
		appRepo:        appRepo,
	}
}

// GetEarningsStatus handles GET /api/v1/apps/{appID}/earnings/status
// Returns earnings availability status (pending, available, paid out)
func (h *RevenueHandler) GetEarningsStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user
	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeJSONErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONErrorResponse(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Get earnings status
	status, err := h.revenueService.GetEarningsStatus(ctx, app.ID)
	if err != nil {
		log.Printf("GetEarningsStatus: failed for app %s: %v", app.ID, err)
		writeJSONErrorResponse(w, http.StatusInternalServerError, "failed to fetch earnings status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetEarnings handles GET /api/v1/apps/{appID}/earnings
// Query params: start (required, YYYY-MM-DD), end (required, YYYY-MM-DD), mode (optional: combined|split)
func (h *RevenueHandler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user
	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeJSONErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONErrorResponse(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Parse query parameters
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	mode := r.URL.Query().Get("mode")

	if startStr == "" || endStr == "" {
		writeJSONErrorResponse(w, http.StatusBadRequest, "start and end dates are required (format: YYYY-MM-DD)")
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		writeJSONErrorResponse(w, http.StatusBadRequest, "invalid start date format, use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		writeJSONErrorResponse(w, http.StatusBadRequest, "invalid end date format, use YYYY-MM-DD")
		return
	}

	// Default to combined mode
	revenueMode := service.RevenueModeCombined
	if mode == "split" {
		revenueMode = service.RevenueModeSplit
	}

	// Get earnings metrics
	metrics, err := h.revenueService.GetEarningsByDateRange(ctx, app.ID, startDate, endDate, revenueMode)
	if err != nil {
		if err == service.ErrInvalidDateRange {
			writeJSONErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("GetEarnings: failed for app %s: %v", app.ID, err)
		writeJSONErrorResponse(w, http.StatusInternalServerError, "failed to fetch earnings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetRevenueConcentration handles GET /api/v1/apps/{appID}/revenue/concentration?top=10
// Returns top stores by revenue for the app
func (h *RevenueHandler) GetRevenueConcentration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeJSONErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONErrorResponse(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Parse params
	limit := 10
	if topStr := r.URL.Query().Get("top"); topStr != "" {
		if parsed, err := json.Number(topStr).Int64(); err == nil && parsed >= 1 && parsed <= 100 {
			limit = int(parsed)
		}
	}

	// Default to last 90 days
	now := time.Now()
	start := now.AddDate(0, -3, 0)
	end := now

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsed
		}
	}
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsed.Add(24*time.Hour - time.Second)
		}
	}

	result, err := h.revenueService.GetRevenueConcentration(ctx, app.ID, start, end, limit)
	if err != nil {
		log.Printf("GetRevenueConcentration: failed for app %s: %v", app.ID, err)
		writeJSONErrorResponse(w, http.StatusInternalServerError, "failed to compute revenue concentration")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func writeJSONErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}

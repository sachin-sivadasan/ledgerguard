package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	// Parse app ID from URL
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONErrorResponse(w, http.StatusBadRequest, "app ID is required")
		return
	}

	// Convert numeric app ID to UUID by looking up the app
	app, err := h.lookupAppByNumericID(ctx, user.ID, appIDStr)
	if err != nil {
		writeJSONErrorResponse(w, http.StatusNotFound, "app not found")
		return
	}

	// Get earnings status
	status, err := h.revenueService.GetEarningsStatus(ctx, app.ID)
	if err != nil {
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

	// Parse app ID from URL
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONErrorResponse(w, http.StatusBadRequest, "app ID is required")
		return
	}

	// Convert numeric app ID to UUID by looking up the app
	app, err := h.lookupAppByNumericID(ctx, user.ID, appIDStr)
	if err != nil {
		writeJSONErrorResponse(w, http.StatusNotFound, "app not found")
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
		writeJSONErrorResponse(w, http.StatusInternalServerError, "failed to fetch earnings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// lookupAppByNumericID finds an app by its numeric ID (extracted from Shopify GID)
func (h *RevenueHandler) lookupAppByNumericID(ctx context.Context, userID uuid.UUID, numericID string) (*struct {
	ID uuid.UUID
}, error) {
	// Get partner account for user
	partner, err := h.partnerRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Search for app in partner account
	apps, err := h.appRepo.FindByPartnerAccountID(ctx, partner.ID)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		// Extract numeric ID from GID (e.g., "gid://partners/App/12345" -> "12345")
		if extractNumericID(app.PartnerAppID) == numericID {
			return &struct{ ID uuid.UUID }{ID: app.ID}, nil
		}
	}

	return nil, ErrAppNotFound
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

// extractNumericID extracts the numeric ID from a Shopify GID
func extractNumericID(gid string) string {
	// GID format: gid://partners/App/12345
	// We want: 12345
	for i := len(gid) - 1; i >= 0; i-- {
		if gid[i] == '/' {
			return gid[i+1:]
		}
	}
	return gid
}

// ErrAppNotFound is returned when app is not found
var ErrAppNotFound = &handlerError{message: "app not found"}

type handlerError struct {
	message string
}

func (e *handlerError) Error() string {
	return e.message
}

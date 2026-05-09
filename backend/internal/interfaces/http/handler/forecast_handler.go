package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type ForecastHandler struct {
	snapshotRepo repository.DailyMetricsSnapshotRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
	engine       *service.ForecastingEngine
}

func NewForecastHandler(
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *ForecastHandler {
	return &ForecastHandler{
		snapshotRepo: snapshotRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
		engine:       service.NewForecastingEngine(),
	}
}

// GetForecast returns a revenue forecast for the specified app.
// GET /api/v1/apps/{appID}/forecast?months=12&model=linear|exponential&alpha=0.3
func (h *ForecastHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Parse query params
	months := 12
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 36 {
			months = parsed
		}
	}

	model := r.URL.Query().Get("model")
	if model == "" {
		model = service.ForecastModelLinear
	}
	if model != service.ForecastModelLinear && model != service.ForecastModelExponential {
		writeJSONError(w, http.StatusBadRequest, "model must be 'linear' or 'exponential'")
		return
	}

	alpha := 0.3
	if a := r.URL.Query().Get("alpha"); a != "" {
		if parsed, err := strconv.ParseFloat(a, 64); err == nil {
			alpha = parsed
		}
	}

	// Fetch last 12 months of snapshots
	now := time.Now()
	from := now.AddDate(-1, 0, 0)
	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, now)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch snapshot data")
		return
	}

	writeInsufficientData := func() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":       "insufficient data for forecasting",
			"data_points": len(snapshots),
			"required":    service.MinDataPointsForForecast,
		})
	}

	var result interface{}
	switch model {
	case service.ForecastModelLinear:
		forecast, err := h.engine.LinearRegressionForecast(snapshots, months, app.ID)
		if err == service.ErrInsufficientData {
			writeInsufficientData()
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "forecasting error")
			return
		}
		result = forecast
	case service.ForecastModelExponential:
		forecast, err := h.engine.ExponentialSmoothingForecast(snapshots, months, alpha, app.ID)
		if err == service.ErrInsufficientData {
			writeInsufficientData()
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "forecasting error")
			return
		}
		result = forecast
	}

	fr := result.(*entity.ForecastResult) // safe: switch exhausts cases above
	// Build JSON response
	type forecastPointJSON struct {
		Date             string `json:"date"`
		ExpectedCents    int64  `json:"expected_cents"`
		OptimisticCents  int64  `json:"optimistic_cents"`
		PessimisticCents int64  `json:"pessimistic_cents"`
	}
	points := make([]forecastPointJSON, len(fr.Points))
	for i, p := range fr.Points {
		points[i] = forecastPointJSON{
			Date:             p.Date.Format("2006-01-02"),
			ExpectedCents:    p.ExpectedCents,
			OptimisticCents:  p.OptimisticCents,
			PessimisticCents: p.PessimisticCents,
		}
	}

	resp := map[string]interface{}{
		"model":            fr.Model,
		"generated_at":     fr.GeneratedAt.Format(time.RFC3339),
		"data_points_used": fr.DataPointsUsed,
		"forecast":         points,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

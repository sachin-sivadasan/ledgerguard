package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

const appGIDPrefix = "gid://partners/App/"

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
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
		"app_id":               fullAppGID,
		"active_mrr_cents":     125000,  // $1,250.00
		"revenue_at_risk_cents": 15000,  // $150.00
		"usage_revenue_cents":   35000,  // $350.00
		"total_revenue_cents":   175000, // $1,750.00
		"renewal_success_rate":  0.92,   // 92%
		"safe_count":            45,
		"one_cycle_missed_count": 5,
		"two_cycles_missed_count": 2,
		"churned_count":          3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

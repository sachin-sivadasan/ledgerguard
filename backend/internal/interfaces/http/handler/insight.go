package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type InsightHandler struct {
	insightRepo    repository.DailyInsightRepository
	appRepo        repository.AppRepository
	partnerRepo    repository.PartnerAccountRepository
}

func NewInsightHandler(
	insightRepo repository.DailyInsightRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *InsightHandler {
	return &InsightHandler{
		insightRepo:    insightRepo,
		appRepo:        appRepo,
		partnerRepo:    partnerRepo,
	}
}

// GetDailyInsight returns the daily AI insight for an app
func (h *InsightHandler) GetDailyInsight(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Get today's insight (or the latest one)
	insight, err := h.insightRepo.FindLatestByAppID(r.Context(), app.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrDailyInsightNotFound) {
			// Return a default insight if none exists
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(DailyInsightResponse{
				Summary:     "No insights available yet. Insights are generated daily after your first sync.",
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				KeyPoints:   []string{},
			})
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to fetch insight")
		return
	}

	// Parse key points from insight text if formatted with bullets
	keyPoints := extractKeyPoints(insight.InsightText)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DailyInsightResponse{
		Summary:     insight.InsightText,
		GeneratedAt: insight.CreatedAt.Format(time.RFC3339),
		KeyPoints:   keyPoints,
	})
}

// DailyInsightResponse represents the response for daily insight
type DailyInsightResponse struct {
	Summary     string   `json:"summary"`
	GeneratedAt string   `json:"generated_at"`
	KeyPoints   []string `json:"key_points"`
}

// extractKeyPoints attempts to parse bullet points from insight text
func extractKeyPoints(text string) []string {
	var points []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for bullet point patterns
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "• ") || strings.HasPrefix(line, "* ") {
			point := strings.TrimPrefix(line, "- ")
			point = strings.TrimPrefix(point, "• ")
			point = strings.TrimPrefix(point, "* ")
			if point != "" {
				points = append(points, point)
			}
		}
	}
	return points
}

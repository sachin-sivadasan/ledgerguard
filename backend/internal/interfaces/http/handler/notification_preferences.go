package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type NotificationPreferencesHandler struct {
	repo repository.NotificationPreferencesRepository
}

func NewNotificationPreferencesHandler(repo repository.NotificationPreferencesRepository) *NotificationPreferencesHandler {
	return &NotificationPreferencesHandler{repo: repo}
}

// GetNotificationPreferences returns the current user's notification preferences
func (h *NotificationPreferencesHandler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	prefs, err := h.repo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, persistence.ErrNotificationPreferencesNotFound) {
			// Return default preferences if none exist
			prefs = entity.NewNotificationPreferences(user.ID)
		} else {
			writeJSONError(w, http.StatusInternalServerError, "Failed to fetch notification preferences")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toNotificationPreferencesResponse(prefs))
}

// SaveNotificationPreferences updates the current user's notification preferences
func (h *NotificationPreferencesHandler) SaveNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req NotificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse time from HH:MM format
	summaryTime, err := parseTimeOfDay(req.DailySummaryTime)
	if err != nil {
		// Default to 9:00 AM if parsing fails
		summaryTime = time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	}

	now := time.Now().UTC()
	prefs := &entity.NotificationPreferences{
		ID:                  uuid.New(),
		UserID:              user.ID,
		CriticalEnabled:     req.CriticalAlertsEnabled,
		DailySummaryEnabled: req.DailySummaryEnabled,
		DailySummaryTime:    summaryTime,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := h.repo.Upsert(r.Context(), prefs); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to save notification preferences")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toNotificationPreferencesResponse(prefs))
}

// NotificationPreferencesRequest represents the request body for updating preferences
type NotificationPreferencesRequest struct {
	CriticalAlertsEnabled bool   `json:"critical_alerts_enabled"`
	DailySummaryEnabled   bool   `json:"daily_summary_enabled"`
	DailySummaryTime      string `json:"daily_summary_time"` // HH:MM format
}

// NotificationPreferencesResponse represents the response for notification preferences
type NotificationPreferencesResponse struct {
	CriticalAlertsEnabled bool   `json:"critical_alerts_enabled"`
	DailySummaryEnabled   bool   `json:"daily_summary_enabled"`
	DailySummaryTime      string `json:"daily_summary_time"` // HH:MM format
}

func toNotificationPreferencesResponse(prefs *entity.NotificationPreferences) NotificationPreferencesResponse {
	return NotificationPreferencesResponse{
		CriticalAlertsEnabled: prefs.CriticalEnabled,
		DailySummaryEnabled:   prefs.DailySummaryEnabled,
		DailySummaryTime:      prefs.DailySummaryTime.Format("15:04"),
	}
}

// parseTimeOfDay parses a time string in HH:MM format to a time.Time
func parseTimeOfDay(s string) (time.Time, error) {
	return time.Parse("15:04", s)
}

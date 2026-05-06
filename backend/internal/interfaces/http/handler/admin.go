package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/scheduler"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// AdminHandler serves admin dashboard endpoints.
type AdminHandler struct {
	adminRepo  repository.AdminRepository
	notifSched *scheduler.NotificationScheduler
}

func NewAdminHandler(adminRepo repository.AdminRepository) *AdminHandler {
	return &AdminHandler{adminRepo: adminRepo}
}

// WithNotificationScheduler adds the notification scheduler for admin triggers.
func (h *AdminHandler) WithNotificationScheduler(s *scheduler.NotificationScheduler) *AdminHandler {
	h.notifSched = s
	return h
}

// ListUsers returns all users with onboarding status, plan tier, and app count.
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.adminRepo.ListUsers(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

// OnboardingFunnel returns aggregate onboarding funnel metrics.
// GET /api/v1/admin/onboarding
func (h *AdminHandler) OnboardingFunnel(w http.ResponseWriter, r *http.Request) {
	funnel, err := h.adminRepo.GetOnboardingFunnel(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch onboarding funnel")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(funnel)
}

// ListSyncJobs returns recent sync jobs across all apps.
// GET /api/v1/admin/sync?limit=50
func (h *AdminHandler) ListSyncJobs(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	jobs, err := h.adminRepo.ListSyncJobs(r.Context(), limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch sync jobs")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

// ListBilling returns all billing subscriptions with user info.
// GET /api/v1/admin/billing
func (h *AdminHandler) ListBilling(w http.ResponseWriter, r *http.Request) {
	subs, err := h.adminRepo.ListBillingSubscriptions(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch billing subscriptions")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subs,
		"total":         len(subs),
	})
}

// ResetAppData deletes all sync artifacts for a given app.
// DELETE /api/v1/admin/apps/{appID}/data
func (h *AdminHandler) ResetAppData(w http.ResponseWriter, r *http.Request) {
	appID, err := uuid.Parse(chi.URLParam(r, "appID"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid appID")
		return
	}

	deleted, err := h.adminRepo.ResetAppData(r.Context(), appID)
	if err != nil {
		log.Printf("[admin] ResetAppData error for app %s: %v", appID, err)
		writeJSONError(w, http.StatusInternalServerError, "failed to reset app data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": deleted,
	})
}

// TriggerDailySummary triggers daily summary notifications for a given hour.
// POST /api/v1/admin/notifications/daily-summary?hour=8
// Defaults to current UTC hour if not specified.
func (h *AdminHandler) TriggerDailySummary(w http.ResponseWriter, r *http.Request) {
	if h.notifSched == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "notification scheduler not configured")
		return
	}

	hour := time.Now().UTC().Hour()
	if v := r.URL.Query().Get("hour"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 23 {
			hour = n
		} else {
			writeJSONError(w, http.StatusBadRequest, "hour must be 0-23")
			return
		}
	}

	usersNotified := h.notifSched.RunForHour(r.Context(), hour)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hour":           hour,
		"users_notified": usersNotified,
	})
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// OnboardingHandler handles onboarding endpoints
type OnboardingHandler struct {
	userRepo    repository.UserRepository
	partnerRepo repository.PartnerAccountRepository
	appRepo     repository.AppRepository
}

// NewOnboardingHandler creates a new OnboardingHandler
func NewOnboardingHandler(
	userRepo repository.UserRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *OnboardingHandler {
	return &OnboardingHandler{
		userRepo:    userRepo,
		partnerRepo: partnerRepo,
		appRepo:     appRepo,
	}
}

// OnboardingStatusResponse represents the onboarding status response
type OnboardingStatusResponse struct {
	IsComplete        bool   `json:"is_complete"`
	CompletedAt       string `json:"completed_at,omitempty"`
	HasPartnerAccount bool   `json:"has_partner_account"`
	HasApps           bool   `json:"has_apps"`
	NextStep          string `json:"next_step,omitempty"`
}

// GetStatus handles GET /api/v1/users/onboarding-status
// Returns the user's onboarding status and next step
func (h *OnboardingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeOnboardingError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Fetch full user data to check onboarding status
	fullUser, err := h.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		writeOnboardingError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	status := OnboardingStatusResponse{
		IsComplete: fullUser.IsOnboardingComplete(),
	}

	if fullUser.OnboardingCompletedAt != nil {
		status.CompletedAt = fullUser.OnboardingCompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	// Check if user has a partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(ctx, user.ID)
	if err == nil && partnerAccount != nil {
		status.HasPartnerAccount = true

		// Check if user has apps
		apps, err := h.appRepo.FindByPartnerAccountID(ctx, partnerAccount.ID)
		if err == nil && len(apps) > 0 {
			status.HasApps = true
		}
	}

	// Determine next step
	if !status.IsComplete {
		if !status.HasPartnerAccount {
			status.NextStep = "connect_partner"
		} else if !status.HasApps {
			status.NextStep = "select_app"
		} else {
			status.NextStep = "complete"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Complete handles POST /api/v1/users/onboarding-complete
// Marks the user's onboarding as complete
func (h *OnboardingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeOnboardingError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Fetch full user data
	fullUser, err := h.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		writeOnboardingError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	// Check if already completed
	if fullUser.IsOnboardingComplete() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"message":      "onboarding already completed",
			"completed_at": fullUser.OnboardingCompletedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
		return
	}

	// Validate that prerequisites are met
	partnerAccount, err := h.partnerRepo.FindByUserID(ctx, user.ID)
	if err != nil || partnerAccount == nil {
		writeOnboardingError(w, http.StatusBadRequest, "must connect partner account before completing onboarding")
		return
	}

	apps, err := h.appRepo.FindByPartnerAccountID(ctx, partnerAccount.ID)
	if err != nil || len(apps) == 0 {
		writeOnboardingError(w, http.StatusBadRequest, "must select at least one app before completing onboarding")
		return
	}

	// Mark onboarding as complete
	fullUser.CompleteOnboarding()
	if err := h.userRepo.Update(ctx, fullUser); err != nil {
		writeOnboardingError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "onboarding completed successfully",
		"completed_at": fullUser.OnboardingCompletedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// writeOnboardingError writes an error response
func writeOnboardingError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}

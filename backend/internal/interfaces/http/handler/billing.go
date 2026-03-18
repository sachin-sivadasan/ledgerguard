package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	appservice "github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// BillingHandler handles billing-related HTTP endpoints.
type BillingHandler struct {
	billingService *appservice.BillingService
}

// NewBillingHandler creates a new BillingHandler.
func NewBillingHandler(billingService *appservice.BillingService) *BillingHandler {
	return &BillingHandler{billingService: billingService}
}

// CreateCheckout handles POST /api/v1/billing/checkout
func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan := valueobject.ParseBillingPlan(req.Plan)
	if !plan.IsValid() {
		writeJSONError(w, http.StatusBadRequest, "invalid plan: must be STARTER or PRO")
		return
	}

	result, err := h.billingService.CreateCheckout(r.Context(), user.ID, plan)
	if err != nil {
		log.Printf("billing: checkout error for user %s: %v", user.ID, err)
		writeJSONError(w, http.StatusInternalServerError, "failed to create checkout")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetStatus handles GET /api/v1/billing/status
func (h *BillingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	status, err := h.billingService.GetBillingStatus(r.Context(), user.ID)
	if err != nil {
		log.Printf("billing: status error for user %s: %v", user.ID, err)
		writeJSONError(w, http.StatusInternalServerError, "failed to get billing status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleWebhook handles POST /webhooks/razorpay
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("billing webhook: failed to read body: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	signature := r.Header.Get("X-Razorpay-Signature")

	if err := h.billingService.HandleWebhookEvent(r.Context(), body, signature); err != nil {
		log.Printf("billing webhook: signature verification failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Always return 200 to prevent Razorpay retries
	w.WriteHeader(http.StatusOK)
}

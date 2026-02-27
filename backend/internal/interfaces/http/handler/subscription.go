package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// subscriptionAppGIDPrefix is the Shopify Partner App GID prefix
const subscriptionAppGIDPrefix = "gid://partners/App/"

type SubscriptionHandler struct {
	subscriptionRepo repository.SubscriptionRepository
	partnerRepo      repository.PartnerAccountRepository
	appRepo          repository.AppRepository
}

func NewSubscriptionHandler(
	subscriptionRepo repository.SubscriptionRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionRepo: subscriptionRepo,
		partnerRepo:      partnerRepo,
		appRepo:          appRepo,
	}
}

// List returns subscriptions for an app with optional filtering
// GET /api/v1/apps/{appID}/subscriptions?risk_state=SAFE&limit=50&offset=0
// appID is numeric (e.g., "4599915"), backend constructs full GID
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Get numeric appID from URL and construct full GID
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app ID is required")
		return
	}
	fullAppGID := subscriptionAppGIDPrefix + appIDStr

	// Find app by partner app ID (GID)
	app, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	// App ownership is already verified by FindByPartnerAppID (requires partnerAccountID)
	appID := app.ID

	// Parse query parameters
	riskStateStr := r.URL.Query().Get("risk_state")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Fetch subscriptions
	var subscriptions []*entity.Subscription
	if riskStateStr != "" {
		riskState, valid := parseRiskState(riskStateStr)
		if !valid {
			writeJSONError(w, http.StatusBadRequest, "invalid risk_state")
			return
		}
		subscriptions, err = h.subscriptionRepo.FindByRiskState(r.Context(), appID, riskState)
	} else {
		subscriptions, err = h.subscriptionRepo.FindByAppID(r.Context(), appID)
	}

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch subscriptions")
		return
	}

	// Apply pagination
	total := len(subscriptions)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	paginated := subscriptions[start:end]

	// Build response
	subResponses := make([]map[string]interface{}, len(paginated))
	for i, sub := range paginated {
		subResponses[i] = subscriptionToJSON(sub)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subResponses,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetByID returns a single subscription by ID
// GET /api/v1/apps/{appID}/subscriptions/{subscriptionID}
// appID is numeric (e.g., "4599915"), backend constructs full GID
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Get numeric appID from URL and construct full GID
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app ID is required")
		return
	}
	fullAppGID := subscriptionAppGIDPrefix + appIDStr

	// Find app by partner app ID (GID)
	app, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	// App ownership is already verified by FindByPartnerAppID
	appID := app.ID

	// Parse subscriptionID from URL
	subIDStr := chi.URLParam(r, "subscriptionID")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid subscription ID")
		return
	}

	// Fetch subscription
	subscription, err := h.subscriptionRepo.FindByID(r.Context(), subID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "subscription not found")
		return
	}

	// Verify subscription belongs to the app
	if subscription.AppID != appID {
		writeJSONError(w, http.StatusForbidden, "subscription does not belong to this app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscription": subscriptionToJSON(subscription),
	})
}

// parseRiskState converts string to RiskState
func parseRiskState(s string) (valueobject.RiskState, bool) {
	switch s {
	case "SAFE":
		return valueobject.RiskStateSafe, true
	case "ONE_CYCLE_MISSED":
		return valueobject.RiskStateOneCycleMissed, true
	case "TWO_CYCLES_MISSED":
		return valueobject.RiskStateTwoCyclesMissed, true
	case "CHURNED":
		return valueobject.RiskStateChurned, true
	default:
		return "", false
	}
}

// subscriptionToJSON converts a subscription to JSON response format
func subscriptionToJSON(sub *entity.Subscription) map[string]interface{} {
	resp := map[string]interface{}{
		"id":               sub.ID.String(),
		"shopify_gid":      sub.ShopifyGID,
		"myshopify_domain": sub.MyshopifyDomain,
		"plan_name":        sub.PlanName,
		"base_price_cents": sub.BasePriceCents,
		"billing_interval": string(sub.BillingInterval),
		"risk_state":       string(sub.RiskState),
		"status":           sub.Status,
		"created_at":       sub.CreatedAt,
	}

	if sub.ExpectedNextChargeDate != nil {
		resp["expected_next_charge"] = sub.ExpectedNextChargeDate
	}

	if sub.LastRecurringChargeDate != nil {
		resp["last_charge_date"] = sub.LastRecurringChargeDate
	}

	return resp
}

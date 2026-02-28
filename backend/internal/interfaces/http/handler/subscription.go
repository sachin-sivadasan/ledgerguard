package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
// GET /api/v1/apps/{appID}/subscriptions
// Query params: status (comma-sep), priceMin, priceMax, billingInterval, search, sortBy, sortOrder, page, pageSize
// appID is numeric (e.g., "4599915"), backend constructs full GID
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	appID, err := h.getAppIDFromRequest(r)
	if err != nil {
		writeJSONError(w, err.statusCode, err.message)
		return
	}

	// Parse filter parameters
	filters := repository.SubscriptionFilters{
		Page:     1,
		PageSize: 25,
	}

	// Status filter (comma-separated risk states)
	statusStr := r.URL.Query().Get("status")
	if statusStr != "" {
		statuses := strings.Split(statusStr, ",")
		for _, s := range statuses {
			riskState, valid := parseRiskState(strings.TrimSpace(s))
			if valid {
				filters.RiskStates = append(filters.RiskStates, riskState)
			}
		}
	}

	// Legacy support: single risk_state param
	riskStateStr := r.URL.Query().Get("risk_state")
	if riskStateStr != "" && len(filters.RiskStates) == 0 {
		riskState, valid := parseRiskState(riskStateStr)
		if valid {
			filters.RiskStates = append(filters.RiskStates, riskState)
		}
	}

	// Price range filter
	if priceMinStr := r.URL.Query().Get("priceMin"); priceMinStr != "" {
		if parsed, err := strconv.ParseInt(priceMinStr, 10, 64); err == nil && parsed >= 0 {
			filters.PriceMinCents = &parsed
		}
	}
	if priceMaxStr := r.URL.Query().Get("priceMax"); priceMaxStr != "" {
		if parsed, err := strconv.ParseInt(priceMaxStr, 10, 64); err == nil && parsed >= 0 {
			filters.PriceMaxCents = &parsed
		}
	}

	// Billing interval filter
	if intervalStr := r.URL.Query().Get("billingInterval"); intervalStr != "" {
		interval := valueobject.BillingInterval(intervalStr)
		if interval.IsValid() {
			filters.BillingInterval = &interval
		}
	}

	// Search filter
	filters.SearchTerm = r.URL.Query().Get("search")

	// Sort parameters
	filters.SortBy = r.URL.Query().Get("sortBy")
	filters.SortOrder = r.URL.Query().Get("sortOrder")

	// Pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			filters.Page = parsed
		}
	}
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= 100 {
			filters.PageSize = parsed
		}
	}

	// Legacy support: limit/offset
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			filters.PageSize = parsed
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			filters.Page = (parsed / filters.PageSize) + 1
		}
	}

	// Fetch subscriptions with filters
	result, err2 := h.subscriptionRepo.FindWithFilters(r.Context(), appID, filters)
	if err2 != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch subscriptions")
		return
	}

	// Build response
	subResponses := make([]map[string]interface{}, len(result.Subscriptions))
	for i, sub := range result.Subscriptions {
		subResponses[i] = subscriptionToJSON(sub)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subResponses,
		"total":         result.Total,
		"page":          result.Page,
		"pageSize":      result.PageSize,
		"totalPages":    result.TotalPages,
	})
}

// Summary returns aggregate subscription statistics
// GET /api/v1/apps/{appID}/subscriptions/summary
func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	appID, err := h.getAppIDFromRequest(r)
	if err != nil {
		writeJSONError(w, err.statusCode, err.message)
		return
	}

	summary, err2 := h.subscriptionRepo.GetSummary(r.Context(), appID)
	if err2 != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch summary")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"activeCount":   summary.ActiveCount,
		"atRiskCount":   summary.AtRiskCount,
		"churnedCount":  summary.ChurnedCount,
		"avgPriceCents": summary.AvgPriceCents,
		"totalCount":    summary.TotalCount,
	})
}

// PriceStats returns price statistics and distinct prices for filtering
// GET /api/v1/apps/{appID}/subscriptions/price-stats
func (h *SubscriptionHandler) PriceStats(w http.ResponseWriter, r *http.Request) {
	appID, err := h.getAppIDFromRequest(r)
	if err != nil {
		writeJSONError(w, err.statusCode, err.message)
		return
	}

	stats, err2 := h.subscriptionRepo.GetPriceStats(r.Context(), appID)
	if err2 != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch price stats")
		return
	}

	// Build prices array for response
	prices := make([]map[string]interface{}, len(stats.Prices))
	for i, p := range stats.Prices {
		prices[i] = map[string]interface{}{
			"priceCents": p.PriceCents,
			"count":      p.Count,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"minCents": stats.MinCents,
		"maxCents": stats.MaxCents,
		"avgCents": stats.AvgCents,
		"prices":   prices,
	})
}

// subHandlerError is a helper struct for subscription error responses
type subHandlerError struct {
	statusCode int
	message    string
}

// getAppIDFromRequest extracts and validates app ID from request
func (h *SubscriptionHandler) getAppIDFromRequest(r *http.Request) (uuid.UUID, *subHandlerError) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		return uuid.Nil, &subHandlerError{statusCode: http.StatusUnauthorized, message: "authentication required"}
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		return uuid.Nil, &subHandlerError{statusCode: http.StatusNotFound, message: "no partner account found"}
	}

	// Get numeric appID from URL and construct full GID
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		return uuid.Nil, &subHandlerError{statusCode: http.StatusBadRequest, message: "app ID is required"}
	}
	fullAppGID := subscriptionAppGIDPrefix + appIDStr

	// Find app by partner app ID (GID)
	app, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		return uuid.Nil, &subHandlerError{statusCode: http.StatusNotFound, message: "app not found"}
	}

	return app.ID, nil
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
		"shop_name":        sub.ShopName,
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

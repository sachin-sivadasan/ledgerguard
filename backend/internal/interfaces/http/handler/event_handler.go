package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

type EventHandler struct {
	appEventRepo repository.AppEventRepository
	partnerRepo  repository.PartnerAccountRepository
	appRepo      repository.AppRepository
}

func NewEventHandler(
	appEventRepo repository.AppEventRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *EventHandler {
	return &EventHandler{
		appEventRepo: appEventRepo,
		partnerRepo:  partnerRepo,
		appRepo:      appRepo,
	}
}

// List returns paginated app events for an app.
// GET /api/v1/apps/{appID}/events
// Query params: page (default 1), pageSize (default 20, max 100), eventType, storeDomain
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	filters := repository.EventFilters{
		Page:     1,
		PageSize: 20,
	}

	if s := r.URL.Query().Get("eventType"); s != "" {
		filters.EventType = s
	}
	if s := r.URL.Query().Get("storeDomain"); s != "" {
		filters.StoreDomain = s
	}
	if s := r.URL.Query().Get("page"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 {
			filters.Page = parsed
		}
	}
	if s := r.URL.Query().Get("pageSize"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 100 {
			filters.PageSize = parsed
		}
	}

	result, err := h.appEventRepo.FindByAppIDPaginated(r.Context(), app.ID, filters)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	type eventJSON struct {
		ID          string `json:"id"`
		Date        string `json:"date"`
		Type        string `json:"type"`
		AppID       string `json:"app_id"`
		StoreDomain string `json:"store_domain"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	items := make([]eventJSON, 0, len(result.Events))
	for _, ev := range result.Events {
		mappedType := mapEventType(ev.EventType)
		title, desc := eventTitleDescription(mappedType, ev.ShopifyShopGID)

		items = append(items, eventJSON{
			ID:          ev.ID.String(),
			Date:        ev.OccurredAt.Format(time.RFC3339),
			Type:        mappedType,
			AppID:       app.ID.String(),
			StoreDomain: extractDomainFromGID(ev.ShopifyShopGID),
			Title:       title,
			Description: desc,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":     items,
		"total":      result.Total,
		"page":       result.Page,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	})
}

// mapEventType maps Shopify event types to Flutter-expected types
func mapEventType(eventType string) string {
	switch strings.ToUpper(eventType) {
	case "RELATIONSHIP_INSTALLED":
		return "APP_INSTALL"
	case "RELATIONSHIP_UNINSTALLED":
		return "APP_UNINSTALL"
	case "RELATIONSHIP_REACTIVATED":
		return "APP_REACTIVATED"
	case "RELATIONSHIP_DEACTIVATED":
		return "APP_DEACTIVATED"
	case "SUBSCRIPTION_CHARGE_ACCEPTED":
		return "SUBSCRIPTION_ACTIVATED"
	case "SUBSCRIPTION_CHARGE_CANCELED":
		return "SUBSCRIPTION_CANCELLED"
	case "SUBSCRIPTION_CHARGE_FROZEN":
		return "SUBSCRIPTION_FROZEN"
	case "SUBSCRIPTION_CHARGE_UNFROZEN":
		return "SUBSCRIPTION_UNFROZEN"
	default:
		return eventType
	}
}

func eventTitleDescription(eventType, shopGID string) (string, string) {
	domain := extractDomainFromGID(shopGID)
	switch eventType {
	case "APP_INSTALL":
		return "App Installed", domain + " installed the app"
	case "APP_UNINSTALL":
		return "App Uninstalled", domain + " uninstalled the app"
	case "APP_REACTIVATED":
		return "App Reactivated", domain + " reactivated the app"
	case "APP_DEACTIVATED":
		return "App Deactivated", domain + " deactivated the app"
	case "SUBSCRIPTION_ACTIVATED":
		return "Subscription Activated", domain + " activated a subscription"
	case "SUBSCRIPTION_CANCELLED":
		return "Subscription Cancelled", domain + " cancelled their subscription"
	case "SUBSCRIPTION_FROZEN":
		return "Subscription Frozen", domain + " subscription was frozen"
	case "SUBSCRIPTION_UNFROZEN":
		return "Subscription Unfrozen", domain + " subscription was unfrozen"
	default:
		return eventType, domain + " triggered " + eventType
	}
}

// extractDomainFromGID extracts a readable domain from a Shopify shop GID or returns it as-is
func extractDomainFromGID(shopGID string) string {
	// If it's already a domain, return as-is
	if strings.Contains(shopGID, ".myshopify.com") {
		return shopGID
	}
	// For GIDs like "gid://shopify/Shop/12345", just return the GID as identifier
	return shopGID
}

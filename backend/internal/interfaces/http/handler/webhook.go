package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
)

// WebhookHandler handles incoming Shopify webhooks
type WebhookHandler struct {
	webhookService *service.WebhookService
}

func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// HandleWebhook processes incoming Shopify webhook events
// POST /webhooks/shopify
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the body for HMAC validation
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Webhook: failed to read body: %v", err)
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	// Get required headers
	topic := r.Header.Get("X-Shopify-Topic")
	shopID := r.Header.Get("X-Shopify-Shop-Domain")
	hmacSignature := r.Header.Get("X-Shopify-Hmac-Sha256")
	appID := r.Header.Get("X-Shopify-API-Version") // We'll use shop domain to look up app

	if topic == "" {
		log.Printf("Webhook: missing X-Shopify-Topic header")
		writeJSONError(w, http.StatusBadRequest, "missing topic header")
		return
	}

	// Note: In production, validate HMAC using the webhook secret
	// For now, we log but don't reject to support development
	if hmacSignature == "" {
		log.Printf("Webhook: warning - missing HMAC signature for topic %s", topic)
	}

	// Build webhook event
	event := service.WebhookEvent{
		Topic:     topic,
		ShopID:    shopID,
		AppID:     appID,
		Payload:   body,
		Timestamp: time.Now().UTC(),
	}

	// Process the webhook
	if err := h.webhookService.ProcessEvent(r.Context(), event); err != nil {
		log.Printf("Webhook: failed to process event (topic=%s): %v", topic, err)
		// Return 200 to prevent Shopify from retrying
		// Log the error for investigation
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Webhook: processed event (topic=%s, shop=%s)", topic, shopID)
	w.WriteHeader(http.StatusOK)
}

// HandleSubscriptionUpdate handles subscription update webhooks
// POST /webhooks/shopify/subscriptions
func (h *WebhookHandler) HandleSubscriptionUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	event := service.WebhookEvent{
		Topic:     "app_subscriptions/update",
		ShopID:    r.Header.Get("X-Shopify-Shop-Domain"),
		AppID:     r.Header.Get("X-Shopify-Webhook-Id"),
		Payload:   body,
		Timestamp: time.Now().UTC(),
	}

	if err := h.webhookService.ProcessSubscriptionUpdate(r.Context(), event); err != nil {
		log.Printf("Webhook: subscription update failed: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleAppUninstalled handles app uninstallation webhooks
// POST /webhooks/shopify/uninstalled
func (h *WebhookHandler) HandleAppUninstalled(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	event := service.WebhookEvent{
		Topic:     "app/uninstalled",
		ShopID:    r.Header.Get("X-Shopify-Shop-Domain"),
		AppID:     r.Header.Get("X-Shopify-Webhook-Id"),
		Payload:   body,
		Timestamp: time.Now().UTC(),
	}

	if err := h.webhookService.ProcessAppUninstalled(r.Context(), event); err != nil {
		log.Printf("Webhook: app uninstalled failed: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleBillingFailure handles billing failure webhooks
// POST /webhooks/shopify/billing-failure
func (h *WebhookHandler) HandleBillingFailure(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	event := service.WebhookEvent{
		Topic:     "subscription_billing_attempts/failure",
		ShopID:    r.Header.Get("X-Shopify-Shop-Domain"),
		AppID:     r.Header.Get("X-Shopify-Webhook-Id"),
		Payload:   body,
		Timestamp: time.Now().UTC(),
	}

	if err := h.webhookService.ProcessBillingFailure(r.Context(), event); err != nil {
		log.Printf("Webhook: billing failure processing failed: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}

// WebhookStats returns stats about processed webhooks
// GET /api/v1/webhooks/stats (admin only)
func (h *WebhookHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// This would query the subscription_events table
	// For now, return a placeholder
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "webhook stats endpoint",
	})
}

package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
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
		logging.FromContext(r.Context()).Error("failed to read webhook body", zap.Error(err))
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	// Get required headers
	topic := r.Header.Get("X-Shopify-Topic")
	shopID := r.Header.Get("X-Shopify-Shop-Domain")
	hmacSignature := r.Header.Get("X-Shopify-Hmac-Sha256")
	appID := r.Header.Get("X-Shopify-API-Version") // We'll use shop domain to look up app

	if topic == "" {
		logging.FromContext(r.Context()).Warn("missing X-Shopify-Topic header")
		writeJSONError(w, http.StatusBadRequest, "missing topic header")
		return
	}

	// Note: In production, validate HMAC using the webhook secret
	// For now, we log but don't reject to support development
	if hmacSignature == "" {
		logging.FromContext(r.Context()).Warn("missing HMAC signature", zap.String("topic", topic))
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
		logging.FromContext(r.Context()).Error("failed to process event", zap.String("topic", topic), zap.Error(err))
		// Return 200 to prevent Shopify from retrying
		// Log the error for investigation
		w.WriteHeader(http.StatusOK)
		return
	}

	logging.FromContext(r.Context()).Info("processed event", zap.String("topic", topic), zap.String("shop", shopID))
	w.WriteHeader(http.StatusOK)
}

// HandleAppInstalled handles app installation webhooks
// POST /webhooks/shopify/installed
func (h *WebhookHandler) HandleAppInstalled(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	event := service.WebhookEvent{
		Topic:     "app/installed",
		ShopID:    r.Header.Get("X-Shopify-Shop-Domain"),
		AppID:     r.Header.Get("X-Shopify-Webhook-Id"),
		Payload:   body,
		Timestamp: time.Now().UTC(),
	}

	if err := h.webhookService.ProcessAppInstalled(r.Context(), event); err != nil {
		logging.FromContext(r.Context()).Error("app installed processing failed", zap.Error(err))
	}

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
		logging.FromContext(r.Context()).Error("subscription update processing failed", zap.Error(err))
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
		logging.FromContext(r.Context()).Error("app uninstalled processing failed", zap.Error(err))
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
		logging.FromContext(r.Context()).Error("billing failure processing failed", zap.Error(err))
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

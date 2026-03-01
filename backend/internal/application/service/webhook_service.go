package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// WebhookEvent represents a parsed webhook event from Shopify
type WebhookEvent struct {
	Topic     string          // e.g., "app_subscriptions/update", "app/uninstalled"
	ShopID    string          // Shopify shop GID
	AppID     string          // Shopify app GID
	Payload   json.RawMessage // Raw event payload
	Timestamp time.Time
}

// SubscriptionUpdatePayload represents the payload for subscription update webhooks
type SubscriptionUpdatePayload struct {
	ID                string  `json:"admin_graphql_api_id"`
	Name              string  `json:"name"`
	Status            string  `json:"status"` // ACTIVE, CANCELLED, FROZEN, EXPIRED
	CreatedAt         string  `json:"created_at"`
	BillingOn         *string `json:"billing_on"`
	TrialDays         int     `json:"trial_days"`
	Test              bool    `json:"test"`
	CappedAmount      string  `json:"capped_amount"`
	BalanceUsed       float64 `json:"balance_used"`
	BalanceRemaining  float64 `json:"balance_remaining"`
	RiskLevel         float64 `json:"risk_level"`
	LineItems         []struct {
		Plan struct {
			PricingDetails struct {
				Interval string `json:"interval"`
			} `json:"pricingDetails"`
		} `json:"plan"`
	} `json:"line_items"`
}

// AppUninstalledPayload represents the payload for app uninstalled webhooks
type AppUninstalledPayload struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Domain          string `json:"domain"`
	MyshopifyDomain string `json:"myshopify_domain"`
}

// WebhookService handles webhook event processing
type WebhookService struct {
	subRepo         repository.SubscriptionRepository
	subEventRepo    repository.SubscriptionEventRepository
	appRepo         repository.AppRepository
	webhookSecrets  map[string]string // app_id -> webhook secret
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	subRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
) *WebhookService {
	return &WebhookService{
		subRepo:        subRepo,
		appRepo:        appRepo,
		webhookSecrets: make(map[string]string),
	}
}

// WithSubscriptionEventRepo adds subscription event repository for lifecycle tracking
func (s *WebhookService) WithSubscriptionEventRepo(repo repository.SubscriptionEventRepository) *WebhookService {
	s.subEventRepo = repo
	return s
}

// RegisterWebhookSecret registers a webhook secret for HMAC validation
func (s *WebhookService) RegisterWebhookSecret(appID, secret string) {
	s.webhookSecrets[appID] = secret
}

// ValidateHMAC validates the webhook HMAC signature
func (s *WebhookService) ValidateHMAC(appID string, body []byte, signature string) bool {
	secret, ok := s.webhookSecrets[appID]
	if !ok {
		log.Printf("No webhook secret registered for app %s", appID)
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// ProcessSubscriptionUpdate handles subscription status change webhooks
func (s *WebhookService) ProcessSubscriptionUpdate(ctx context.Context, event WebhookEvent) error {
	var payload SubscriptionUpdatePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription update payload: %w", err)
	}

	log.Printf("Processing subscription update: %s -> status=%s", payload.ID, payload.Status)

	// Find subscription by Shopify GID
	sub, err := s.subRepo.FindByShopifyGID(ctx, payload.ID)
	if err != nil {
		// Subscription might not exist yet (new subscription)
		log.Printf("Subscription %s not found: %v", payload.ID, err)
		return nil
	}

	oldStatus := sub.Status
	oldRiskState := sub.RiskState

	// Update subscription status
	sub.Status = payload.Status
	sub.UpdatedAt = time.Now().UTC()

	// Update risk state based on new status
	switch payload.Status {
	case "ACTIVE":
		sub.RiskState = valueobject.RiskStateSafe
	case "CANCELLED", "EXPIRED":
		sub.RiskState = valueobject.RiskStateChurned
	case "FROZEN":
		sub.RiskState = valueobject.RiskStateTwoCyclesMissed
	}

	// Save updated subscription
	if err := s.subRepo.Upsert(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Record lifecycle event if repository is configured
	if s.subEventRepo != nil && oldStatus != sub.Status {
		subEvent := entity.NewSubscriptionEvent(
			sub.ID,
			oldStatus,
			sub.Status,
			oldRiskState,
			sub.RiskState,
			"webhook",
			"",
		)
		if err := s.subEventRepo.Create(ctx, subEvent); err != nil {
			log.Printf("Failed to record subscription event: %v", err)
			// Don't fail the webhook processing for event recording failures
		}
	}

	log.Printf("Subscription %s updated: %s -> %s", payload.ID, oldStatus, payload.Status)
	return nil
}

// ProcessAppUninstalled handles app uninstallation webhooks
func (s *WebhookService) ProcessAppUninstalled(ctx context.Context, event WebhookEvent) error {
	var payload AppUninstalledPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse app uninstalled payload: %w", err)
	}

	log.Printf("Processing app uninstalled: shop=%s", payload.MyshopifyDomain)

	// Find all apps matching this Shopify app GID (across all accounts)
	apps, err := s.appRepo.FindAllByPartnerAppID(ctx, event.AppID)
	if err != nil {
		log.Printf("Failed to find app for %s: %v", event.AppID, err)
		return nil
	}

	// Find and soft-delete subscriptions for this shop across all apps
	for _, app := range apps {
		sub, err := s.subRepo.FindByAppIDAndDomain(ctx, app.ID, payload.MyshopifyDomain)
		if err != nil {
			continue // Subscription might not exist
		}

		oldStatus := sub.Status
		oldRiskState := sub.RiskState

		// Mark as uninstalled and churned
		sub.Status = "UNINSTALLED"
		sub.RiskState = valueobject.RiskStateChurned
		sub.UpdatedAt = time.Now().UTC()

		// Soft delete the subscription
		sub.SoftDelete()

		if err := s.subRepo.Upsert(ctx, sub); err != nil {
			log.Printf("Failed to update subscription for %s: %v", payload.MyshopifyDomain, err)
			continue
		}

		// Record lifecycle event
		if s.subEventRepo != nil {
			subEvent := entity.NewSubscriptionEvent(
				sub.ID,
				oldStatus,
				"UNINSTALLED",
				oldRiskState,
				valueobject.RiskStateChurned,
				"app_uninstalled",
				"Shop uninstalled the app",
			)
			if err := s.subEventRepo.Create(ctx, subEvent); err != nil {
				log.Printf("Failed to record subscription event: %v", err)
			}
		}

		log.Printf("Subscription soft-deleted for shop %s", payload.MyshopifyDomain)
	}

	return nil
}

// ProcessBillingFailure handles billing attempt failure webhooks
func (s *WebhookService) ProcessBillingFailure(ctx context.Context, event WebhookEvent) error {
	// Parse billing failure payload
	var payload struct {
		ID                string `json:"admin_graphql_api_id"`
		SubscriptionID    string `json:"subscription_contract_id"`
		ErrorCode         string `json:"error_code"`
		ErrorMessage      string `json:"error_message"`
		Ready             bool   `json:"ready"`
		CompletedAt       string `json:"completed_at"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse billing failure payload: %w", err)
	}

	log.Printf("Processing billing failure: subscription=%s, error=%s", payload.SubscriptionID, payload.ErrorCode)

	// Find subscription
	sub, err := s.subRepo.FindByShopifyGID(ctx, payload.SubscriptionID)
	if err != nil {
		log.Printf("Subscription %s not found: %v", payload.SubscriptionID, err)
		return nil
	}

	oldRiskState := sub.RiskState

	// Escalate risk state on billing failure
	switch sub.RiskState {
	case valueobject.RiskStateSafe:
		sub.RiskState = valueobject.RiskStateOneCycleMissed
	case valueobject.RiskStateOneCycleMissed:
		sub.RiskState = valueobject.RiskStateTwoCyclesMissed
	case valueobject.RiskStateTwoCyclesMissed:
		sub.RiskState = valueobject.RiskStateChurned
	}

	// Mark churn as involuntary
	sub.UpdatedAt = time.Now().UTC()

	if err := s.subRepo.Upsert(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Record billing failure event
	if s.subEventRepo != nil && oldRiskState != sub.RiskState {
		reason := fmt.Sprintf("Billing failure: %s - %s", payload.ErrorCode, payload.ErrorMessage)
		subEvent := entity.NewSubscriptionEvent(
			sub.ID,
			sub.Status,
			sub.Status,
			oldRiskState,
			sub.RiskState,
			"billing_failure",
			reason,
		)
		if err := s.subEventRepo.Create(ctx, subEvent); err != nil {
			log.Printf("Failed to record billing failure event: %v", err)
		}
	}

	log.Printf("Subscription %s risk escalated: %s -> %s due to billing failure",
		payload.SubscriptionID, oldRiskState, sub.RiskState)
	return nil
}

// ProcessEvent routes webhook events to appropriate handlers
func (s *WebhookService) ProcessEvent(ctx context.Context, event WebhookEvent) error {
	switch event.Topic {
	case "app_subscriptions/update":
		return s.ProcessSubscriptionUpdate(ctx, event)
	case "app/uninstalled":
		return s.ProcessAppUninstalled(ctx, event)
	case "subscription_billing_attempts/failure":
		return s.ProcessBillingFailure(ctx, event)
	default:
		log.Printf("Unhandled webhook topic: %s", event.Topic)
		return nil
	}
}

// Helper to get app internal ID from Shopify GID
func (s *WebhookService) getAppByPartnerID(ctx context.Context, partnerAppID string) (*entity.App, error) {
	apps, err := s.appRepo.FindAllByPartnerAppID(ctx, partnerAppID)
	if err != nil || len(apps) == 0 {
		return nil, fmt.Errorf("app not found for partner ID %s", partnerAppID)
	}
	return apps[0], nil
}

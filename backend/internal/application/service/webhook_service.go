package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"go.uber.org/zap"
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
	ID               string  `json:"admin_graphql_api_id"`
	Name             string  `json:"name"`
	Status           string  `json:"status"` // ACTIVE, CANCELLED, FROZEN, EXPIRED
	CreatedAt        string  `json:"created_at"`
	BillingOn        *string `json:"billing_on"`
	TrialDays        int     `json:"trial_days"`
	Test             bool    `json:"test"`
	CappedAmount     string  `json:"capped_amount"`
	BalanceUsed      float64 `json:"balance_used"`
	BalanceRemaining float64 `json:"balance_remaining"`
	RiskLevel        float64 `json:"risk_level"`
	LineItems        []struct {
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
	subRepo            repository.SubscriptionRepository
	subEventRepo       repository.SubscriptionEventRepository
	appEventRepo       repository.AppEventRepository
	appRepo            repository.AppRepository
	partnerAccountRepo repository.PartnerAccountRepository
	notificationSvc    *NotificationService
	webhookSecrets     map[string]string // app_id -> webhook secret
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

// WithAppEventRepo adds app event repository for recording install/uninstall events
func (s *WebhookService) WithAppEventRepo(repo repository.AppEventRepository) *WebhookService {
	s.appEventRepo = repo
	return s
}

// WithNotificationService adds notification support for risk state changes
func (s *WebhookService) WithNotificationService(
	partnerRepo repository.PartnerAccountRepository,
	notifSvc *NotificationService,
) *WebhookService {
	s.partnerAccountRepo = partnerRepo
	s.notificationSvc = notifSvc
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
		zap.L().Warn("No webhook secret registered for app", zap.String("app_id", appID))
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// ProcessAppInstalled handles app installation webhooks from Partner API.
// Logs the event and records a lifecycle event if a subscription already exists for this shop.
func (s *WebhookService) ProcessAppInstalled(ctx context.Context, event WebhookEvent) error {
	var payload AppUninstalledPayload // Same shop payload structure
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse app installed payload: %w", err)
	}

	domain := payload.MyshopifyDomain
	logging.FromContext(ctx).Info("Processing app installed", zap.String("shop", domain))

	// Find apps matching this webhook's app ID
	apps, err := s.appRepo.FindAllByPartnerAppID(ctx, event.AppID)
	if err != nil || len(apps) == 0 {
		logging.FromContext(ctx).Warn("App installed webhook: no matching app found", zap.String("app_id", event.AppID), zap.String("shop", domain))
		return nil
	}

	for _, app := range apps {
		// Record install in app_events (always — for both new installs and reinstalls)
		if s.appEventRepo != nil {
			appEvent := entity.NewAppEvent(app.ID, event.ShopID, "RELATIONSHIP_INSTALLED", event.Timestamp, event.Payload)
			if err := s.appEventRepo.UpsertBatch(ctx, []*entity.AppEvent{appEvent}); err != nil {
				logging.FromContext(ctx).Warn("Failed to record app install event", zap.Error(err))
			}
		}

		// Check if a subscription already exists (reinstall case)
		sub, err := s.subRepo.FindByAppIDAndDomain(ctx, app.ID, domain)
		if err != nil {
			logging.FromContext(ctx).Info("App installed (new) — subscription will be created on first sync", zap.String("shop", domain), zap.String("app_id", app.ID.String()))
			continue
		}

		// Existing subscription found — this is a reinstall
		oldStatus := sub.Status
		oldRiskState := sub.RiskState

		// Reactivate if it was previously uninstalled/churned
		if sub.Status == "UNINSTALLED" || sub.Status == "CANCELLED" {
			sub.Status = "PENDING"
			sub.RiskState = valueobject.RiskStateSafe
			sub.UpdatedAt = time.Now().UTC()
			sub.Restore()

			if err := s.subRepo.Upsert(ctx, sub); err != nil {
				logging.FromContext(ctx).Error("Failed to reactivate subscription", zap.String("shop", domain), zap.Error(err))
				continue
			}
		}

		// Record lifecycle event on the subscription
		if s.subEventRepo != nil {
			subEvent := entity.NewSubscriptionEvent(
				sub.ID,
				oldStatus,
				sub.Status,
				oldRiskState,
				sub.RiskState,
				"app_installed",
				"Shop installed the app",
			)
			if err := s.subEventRepo.Create(ctx, subEvent); err != nil {
				logging.FromContext(ctx).Warn("Failed to record install event", zap.Error(err))
			}
		}

		logging.FromContext(ctx).Info("App installed (reinstall)", zap.String("shop", domain), zap.String("previous_status", oldStatus), zap.String("new_status", sub.Status))
	}

	return nil
}

// ProcessSubscriptionUpdate handles subscription status change webhooks
func (s *WebhookService) ProcessSubscriptionUpdate(ctx context.Context, event WebhookEvent) error {
	var payload SubscriptionUpdatePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse subscription update payload: %w", err)
	}

	logging.FromContext(ctx).Info("Processing subscription update", zap.String("subscription_id", payload.ID), zap.String("status", payload.Status))

	// Find subscription by Shopify GID
	sub, err := s.subRepo.FindByShopifyGID(ctx, payload.ID)
	if err != nil {
		// Subscription might not exist yet (new subscription)
		logging.FromContext(ctx).Warn("Subscription not found", zap.String("subscription_id", payload.ID), zap.Error(err))
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
			logging.FromContext(ctx).Warn("Failed to record subscription event", zap.Error(err))
			// Don't fail the webhook processing for event recording failures
		}
	}

	// Send notification if risk state changed
	if oldRiskState != sub.RiskState {
		app, err := s.appRepo.FindByID(ctx, sub.AppID)
		if err == nil && app != nil {
			s.sendRiskChangeNotification(ctx, sub, app, oldRiskState, sub.RiskState)
		}
	}

	logging.FromContext(ctx).Info("Subscription updated", zap.String("subscription_id", payload.ID), zap.String("previous_status", oldStatus), zap.String("new_status", payload.Status))
	return nil
}

// ProcessAppUninstalled handles app uninstallation webhooks
func (s *WebhookService) ProcessAppUninstalled(ctx context.Context, event WebhookEvent) error {
	var payload AppUninstalledPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse app uninstalled payload: %w", err)
	}

	logging.FromContext(ctx).Info("Processing app uninstalled", zap.String("shop", payload.MyshopifyDomain))

	// Find all apps matching this Shopify app GID (across all accounts)
	apps, err := s.appRepo.FindAllByPartnerAppID(ctx, event.AppID)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to find app", zap.String("app_id", event.AppID), zap.Error(err))
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
			logging.FromContext(ctx).Error("Failed to update subscription", zap.String("shop", payload.MyshopifyDomain), zap.Error(err))
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
				logging.FromContext(ctx).Warn("Failed to record subscription event", zap.Error(err))
			}
		}

		// Send notification for app uninstall (risk changed to churned)
		if oldRiskState != valueobject.RiskStateChurned {
			s.sendRiskChangeNotification(ctx, sub, app, oldRiskState, valueobject.RiskStateChurned)
		}

		logging.FromContext(ctx).Info("Subscription soft-deleted", zap.String("shop", payload.MyshopifyDomain))
	}

	return nil
}

// ProcessBillingFailure handles billing attempt failure webhooks
func (s *WebhookService) ProcessBillingFailure(ctx context.Context, event WebhookEvent) error {
	// Parse billing failure payload
	var payload struct {
		ID             string `json:"admin_graphql_api_id"`
		SubscriptionID string `json:"subscription_contract_id"`
		ErrorCode      string `json:"error_code"`
		ErrorMessage   string `json:"error_message"`
		Ready          bool   `json:"ready"`
		CompletedAt    string `json:"completed_at"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse billing failure payload: %w", err)
	}

	logging.FromContext(ctx).Info("Processing billing failure", zap.String("subscription_id", payload.SubscriptionID), zap.String("error_code", payload.ErrorCode))

	// Find subscription
	sub, err := s.subRepo.FindByShopifyGID(ctx, payload.SubscriptionID)
	if err != nil {
		logging.FromContext(ctx).Warn("Subscription not found", zap.String("subscription_id", payload.SubscriptionID), zap.Error(err))
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
			logging.FromContext(ctx).Warn("Failed to record billing failure event", zap.Error(err))
		}
	}

	// Send notification for billing failure risk escalation
	if oldRiskState != sub.RiskState {
		app, err := s.appRepo.FindByID(ctx, sub.AppID)
		if err == nil && app != nil {
			s.sendRiskChangeNotification(ctx, sub, app, oldRiskState, sub.RiskState)
		}
	}

	logging.FromContext(ctx).Info("Subscription risk escalated due to billing failure",
		zap.String("subscription_id", payload.SubscriptionID), zap.String("previous_risk", oldRiskState.String()), zap.String("new_risk", sub.RiskState.String()))
	return nil
}

// ProcessEvent routes webhook events to appropriate handlers
func (s *WebhookService) ProcessEvent(ctx context.Context, event WebhookEvent) error {
	switch event.Topic {
	case "app/installed":
		return s.ProcessAppInstalled(ctx, event)
	case "app_subscriptions/update":
		return s.ProcessSubscriptionUpdate(ctx, event)
	case "app/uninstalled":
		return s.ProcessAppUninstalled(ctx, event)
	case "subscription_billing_attempts/failure":
		return s.ProcessBillingFailure(ctx, event)
	default:
		logging.FromContext(ctx).Warn("Unhandled webhook topic", zap.String("topic", event.Topic))
		return nil
	}
}

// sendRiskChangeNotification sends a critical alert when risk state changes
func (s *WebhookService) sendRiskChangeNotification(
	ctx context.Context,
	sub *entity.Subscription,
	app *entity.App,
	oldRiskState valueobject.RiskState,
	newRiskState valueobject.RiskState,
) {
	// Skip if notification service not configured
	if s.notificationSvc == nil || s.partnerAccountRepo == nil {
		return
	}

	// Only notify on actual risk state changes
	if oldRiskState == newRiskState {
		return
	}

	// Resolve user ID: Subscription -> App -> PartnerAccount -> UserID
	partnerAccount, err := s.partnerAccountRepo.FindByID(ctx, app.PartnerAccountID)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to find partner account for notification", zap.Error(err))
		return
	}

	// Send critical alert
	if err := s.notificationSvc.SendCriticalAlert(
		ctx,
		partnerAccount.UserID,
		app.ID,
		sub.ID,
		app.Name,
		sub.MyshopifyDomain,
		oldRiskState,
		newRiskState,
	); err != nil {
		logging.FromContext(ctx).Error("Failed to send risk change notification", zap.Error(err))
	}
}

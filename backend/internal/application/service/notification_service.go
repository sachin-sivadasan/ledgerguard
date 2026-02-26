package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// ErrDeviceTokenNotFound is returned when a device token is not found
var ErrDeviceTokenNotFound = errors.New("device token not found")

// ErrInvalidPlatform is returned when an invalid platform is provided
var ErrInvalidPlatform = errors.New("invalid platform")

// PushNotificationProvider defines the interface for sending push notifications
type PushNotificationProvider interface {
	// SendPush sends a push notification to a device
	SendPush(ctx context.Context, deviceToken string, platform entity.Platform, title string, body string) error
}

// SlackNotifier defines the interface for sending Slack notifications
type SlackNotifier interface {
	// SendSlack sends a message to a Slack webhook
	SendSlack(ctx context.Context, webhookURL string, title string, body string, color string) error
}

// Slack color constants
const (
	SlackColorDanger  = "#dc3545" // Red - for critical alerts
	SlackColorWarning = "#ffc107" // Yellow - for warnings
	SlackColorSuccess = "#28a745" // Green - for success
	SlackColorInfo    = "#17a2b8" // Blue - for info
)

// NotificationService handles sending notifications to users
type NotificationService struct {
	deviceTokenRepo repository.DeviceTokenRepository
	prefsRepo       repository.NotificationPreferencesRepository
	pushProvider    PushNotificationProvider
	slackNotifier   SlackNotifier
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	deviceTokenRepo repository.DeviceTokenRepository,
	prefsRepo repository.NotificationPreferencesRepository,
	pushProvider PushNotificationProvider,
) *NotificationService {
	return &NotificationService{
		deviceTokenRepo: deviceTokenRepo,
		prefsRepo:       prefsRepo,
		pushProvider:    pushProvider,
	}
}

// WithSlackNotifier adds Slack notification support
func (s *NotificationService) WithSlackNotifier(notifier SlackNotifier) *NotificationService {
	s.slackNotifier = notifier
	return s
}

// RegisterDevice registers a device token for push notifications
func (s *NotificationService) RegisterDevice(ctx context.Context, userID uuid.UUID, deviceToken string, platform entity.Platform) error {
	if !platform.IsValid() {
		return ErrInvalidPlatform
	}

	// Check if token already exists
	existing, err := s.deviceTokenRepo.FindByToken(ctx, deviceToken)
	if err == nil && existing != nil {
		// Token exists, update if needed
		if existing.UserID != userID {
			// Token belongs to different user, delete and recreate
			if err := s.deviceTokenRepo.Delete(ctx, existing.ID); err != nil {
				return fmt.Errorf("failed to delete existing token: %w", err)
			}
		} else {
			// Same user, no action needed
			return nil
		}
	}

	// Create new device token
	token := entity.NewDeviceToken(userID, deviceToken, platform)
	if err := s.deviceTokenRepo.Create(ctx, token); err != nil {
		return fmt.Errorf("failed to create device token: %w", err)
	}

	// Ensure user has notification preferences
	_, err = s.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		// Create default preferences if not found
		prefs := entity.NewNotificationPreferences(userID)
		if err := s.prefsRepo.Create(ctx, prefs); err != nil {
			// Ignore if preferences already exist (race condition)
			return nil
		}
	}

	return nil
}

// UnregisterDevice removes a device token
func (s *NotificationService) UnregisterDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error {
	existing, err := s.deviceTokenRepo.FindByToken(ctx, deviceToken)
	if err != nil {
		return ErrDeviceTokenNotFound
	}

	// Only allow deleting own tokens
	if existing.UserID != userID {
		return ErrDeviceTokenNotFound
	}

	return s.deviceTokenRepo.DeleteByToken(ctx, deviceToken)
}

// SendCriticalAlert sends a critical alert when risk state changes
func (s *NotificationService) SendCriticalAlert(
	ctx context.Context,
	userID uuid.UUID,
	appName string,
	storeDomain string,
	oldState valueobject.RiskState,
	newState valueobject.RiskState,
) error {
	// Check user preferences
	prefs, err := s.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		// No preferences, use defaults (critical enabled)
		prefs = entity.NewNotificationPreferences(userID)
	}

	if !prefs.ShouldSendCritical() {
		return nil // User has disabled critical alerts
	}

	// Build notification content
	title := fmt.Sprintf("🚨 Risk Alert: %s", appName)
	body := fmt.Sprintf("%s changed from %s to %s", storeDomain, oldState, newState)

	var lastErr error

	// Send to Slack if configured
	if s.slackNotifier != nil && prefs.SlackWebhookURL != "" {
		if err := s.slackNotifier.SendSlack(ctx, prefs.SlackWebhookURL, title, body, SlackColorDanger); err != nil {
			lastErr = err
		}
	}

	// Get user's device tokens
	tokens, err := s.deviceTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get device tokens: %w", err)
	}

	// Send to all devices
	for _, token := range tokens {
		if err := s.pushProvider.SendPush(ctx, token.DeviceToken, token.Platform, title, body); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// SendDailySummary sends a daily summary notification
func (s *NotificationService) SendDailySummary(
	ctx context.Context,
	userID uuid.UUID,
	appName string,
	snapshot *entity.DailyMetricsSnapshot,
) error {
	// Check user preferences
	prefs, err := s.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		// No preferences, use defaults (daily summary enabled)
		prefs = entity.NewNotificationPreferences(userID)
	}

	if !prefs.ShouldSendDailySummary() {
		return nil // User has disabled daily summaries
	}

	// Build notification content
	title := fmt.Sprintf("📊 Daily Summary: %s", appName)
	mrrDollars := float64(snapshot.ActiveMRRCents) / 100
	atRiskDollars := float64(snapshot.RevenueAtRiskCents) / 100
	body := fmt.Sprintf("MRR: $%.2f | At Risk: $%.2f | Renewal Rate: %.1f%%",
		mrrDollars, atRiskDollars, snapshot.RenewalSuccessRate*100)

	var lastErr error

	// Send to Slack if configured
	if s.slackNotifier != nil && prefs.SlackWebhookURL != "" {
		if err := s.slackNotifier.SendSlack(ctx, prefs.SlackWebhookURL, title, body, SlackColorInfo); err != nil {
			lastErr = err
		}
	}

	// Get user's device tokens
	tokens, err := s.deviceTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get device tokens: %w", err)
	}

	// Send to all devices
	for _, token := range tokens {
		if err := s.pushProvider.SendPush(ctx, token.DeviceToken, token.Platform, title, body); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetPreferences retrieves notification preferences for a user
func (s *NotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error) {
	prefs, err := s.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		// Return default preferences if not found
		return entity.NewNotificationPreferences(userID), nil
	}
	return prefs, nil
}

// UpdatePreferences updates notification preferences for a user
func (s *NotificationService) UpdatePreferences(ctx context.Context, prefs *entity.NotificationPreferences) error {
	return s.prefsRepo.Upsert(ctx, prefs)
}

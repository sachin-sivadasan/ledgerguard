package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock implementations

type mockDeviceTokenRepository struct {
	tokens       map[string]*entity.DeviceToken
	tokensByUser map[uuid.UUID][]*entity.DeviceToken
	createErr    error
	deleteErr    error
}

func newMockDeviceTokenRepository() *mockDeviceTokenRepository {
	return &mockDeviceTokenRepository{
		tokens:       make(map[string]*entity.DeviceToken),
		tokensByUser: make(map[uuid.UUID][]*entity.DeviceToken),
	}
}

func (m *mockDeviceTokenRepository) Create(ctx context.Context, token *entity.DeviceToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.tokens[token.DeviceToken] = token
	m.tokensByUser[token.UserID] = append(m.tokensByUser[token.UserID], token)
	return nil
}

func (m *mockDeviceTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.DeviceToken, error) {
	tokens := m.tokensByUser[userID]
	if tokens == nil {
		return []*entity.DeviceToken{}, nil
	}
	return tokens, nil
}

func (m *mockDeviceTokenRepository) FindByToken(ctx context.Context, deviceToken string) (*entity.DeviceToken, error) {
	token, ok := m.tokens[deviceToken]
	if !ok {
		return nil, errors.New("not found")
	}
	return token, nil
}

func (m *mockDeviceTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for tokenStr, token := range m.tokens {
		if token.ID == id {
			delete(m.tokens, tokenStr)
			// Remove from user's list
			userTokens := m.tokensByUser[token.UserID]
			for i, t := range userTokens {
				if t.ID == id {
					m.tokensByUser[token.UserID] = append(userTokens[:i], userTokens[i+1:]...)
					break
				}
			}
			break
		}
	}
	return nil
}

func (m *mockDeviceTokenRepository) DeleteByToken(ctx context.Context, deviceToken string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	token, ok := m.tokens[deviceToken]
	if !ok {
		return errors.New("not found")
	}
	delete(m.tokens, deviceToken)
	// Remove from user's list
	userTokens := m.tokensByUser[token.UserID]
	for i, t := range userTokens {
		if t.DeviceToken == deviceToken {
			m.tokensByUser[token.UserID] = append(userTokens[:i], userTokens[i+1:]...)
			break
		}
	}
	return nil
}

type mockNotificationPreferencesRepository struct {
	prefs     map[uuid.UUID]*entity.NotificationPreferences
	createErr error
	updateErr error
}

func newMockNotificationPreferencesRepository() *mockNotificationPreferencesRepository {
	return &mockNotificationPreferencesRepository{
		prefs: make(map[uuid.UUID]*entity.NotificationPreferences),
	}
}

func (m *mockNotificationPreferencesRepository) Create(ctx context.Context, prefs *entity.NotificationPreferences) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.prefs[prefs.UserID] = prefs
	return nil
}

func (m *mockNotificationPreferencesRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error) {
	prefs, ok := m.prefs[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return prefs, nil
}

func (m *mockNotificationPreferencesRepository) Update(ctx context.Context, prefs *entity.NotificationPreferences) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.prefs[prefs.UserID] = prefs
	return nil
}

func (m *mockNotificationPreferencesRepository) Upsert(ctx context.Context, prefs *entity.NotificationPreferences) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.prefs[prefs.UserID] = prefs
	return nil
}

type mockPushNotificationProvider struct {
	sentNotifications []sentNotification
	sendErr           error
}

type sentNotification struct {
	deviceToken string
	platform    entity.Platform
	title       string
	body        string
}

func newMockPushNotificationProvider() *mockPushNotificationProvider {
	return &mockPushNotificationProvider{
		sentNotifications: make([]sentNotification, 0),
	}
}

func (m *mockPushNotificationProvider) SendPush(ctx context.Context, deviceToken string, platform entity.Platform, title string, body string) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentNotifications = append(m.sentNotifications, sentNotification{
		deviceToken: deviceToken,
		platform:    platform,
		title:       title,
		body:        body,
	})
	return nil
}

// Tests

func TestNotificationService_RegisterDevice(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("registers new device successfully", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		err := svc.RegisterDevice(ctx, userID, "fcm-token-123", entity.PlatformAndroid)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify token was stored
		token, err := tokenRepo.FindByToken(ctx, "fcm-token-123")
		if err != nil {
			t.Fatalf("expected to find token, got %v", err)
		}
		if token.UserID != userID {
			t.Errorf("expected user ID %s, got %s", userID, token.UserID)
		}
		if token.Platform != entity.PlatformAndroid {
			t.Errorf("expected platform Android, got %s", token.Platform)
		}
	})

	t.Run("rejects invalid platform", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		err := svc.RegisterDevice(ctx, userID, "token", entity.Platform("invalid"))
		if !errors.Is(err, ErrInvalidPlatform) {
			t.Errorf("expected ErrInvalidPlatform, got %v", err)
		}
	})

	t.Run("handles duplicate registration from same user", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// First registration
		err := svc.RegisterDevice(ctx, userID, "fcm-token-123", entity.PlatformIOS)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Same token, same user - should succeed without error
		err = svc.RegisterDevice(ctx, userID, "fcm-token-123", entity.PlatformIOS)
		if err != nil {
			t.Fatalf("expected no error for duplicate, got %v", err)
		}
	})

	t.Run("transfers token to new user", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		user1 := uuid.New()
		user2 := uuid.New()

		// Register for user1
		err := svc.RegisterDevice(ctx, user1, "fcm-token-123", entity.PlatformIOS)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Register same token for user2
		err = svc.RegisterDevice(ctx, user2, "fcm-token-123", entity.PlatformIOS)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify token now belongs to user2
		token, _ := tokenRepo.FindByToken(ctx, "fcm-token-123")
		if token.UserID != user2 {
			t.Errorf("expected token to belong to user2 %s, got %s", user2, token.UserID)
		}
	})
}

func TestNotificationService_UnregisterDevice(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("unregisters device successfully", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Register first
		_ = svc.RegisterDevice(ctx, userID, "fcm-token-123", entity.PlatformAndroid)

		// Unregister
		err := svc.UnregisterDevice(ctx, userID, "fcm-token-123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify token was removed
		_, err = tokenRepo.FindByToken(ctx, "fcm-token-123")
		if err == nil {
			t.Error("expected token to be removed")
		}
	})

	t.Run("returns error for non-existent token", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		err := svc.UnregisterDevice(ctx, userID, "non-existent-token")
		if !errors.Is(err, ErrDeviceTokenNotFound) {
			t.Errorf("expected ErrDeviceTokenNotFound, got %v", err)
		}
	})

	t.Run("prevents unregistering other user's token", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		user1 := uuid.New()
		user2 := uuid.New()

		// Register for user1
		_ = svc.RegisterDevice(ctx, user1, "fcm-token-123", entity.PlatformIOS)

		// Try to unregister as user2
		err := svc.UnregisterDevice(ctx, user2, "fcm-token-123")
		if !errors.Is(err, ErrDeviceTokenNotFound) {
			t.Errorf("expected ErrDeviceTokenNotFound, got %v", err)
		}
	})
}

func TestNotificationService_SendCriticalAlert(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("sends critical alert to all devices", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Register two devices
		_ = svc.RegisterDevice(ctx, userID, "token-1", entity.PlatformIOS)
		_ = svc.RegisterDevice(ctx, userID, "token-2", entity.PlatformAndroid)

		// Send critical alert
		err := svc.SendCriticalAlert(ctx, userID, "MyApp", "store.myshopify.com",
			valueobject.RiskStateSafe, valueobject.RiskStateOneCycleMissed)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify notifications were sent
		if len(pushProvider.sentNotifications) != 2 {
			t.Errorf("expected 2 notifications, got %d", len(pushProvider.sentNotifications))
		}

		// Verify content
		notif := pushProvider.sentNotifications[0]
		if notif.title != "🚨 Risk Alert: MyApp" {
			t.Errorf("unexpected title: %s", notif.title)
		}
		if notif.body != "store.myshopify.com changed from SAFE to ONE_CYCLE_MISSED" {
			t.Errorf("unexpected body: %s", notif.body)
		}
	})

	t.Run("respects disabled critical alerts preference", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Register device
		_ = svc.RegisterDevice(ctx, userID, "token-1", entity.PlatformIOS)

		// Disable critical alerts
		prefs := entity.NewNotificationPreferences(userID)
		prefs.CriticalEnabled = false
		_ = prefsRepo.Upsert(ctx, prefs)

		// Send critical alert
		err := svc.SendCriticalAlert(ctx, userID, "MyApp", "store.myshopify.com",
			valueobject.RiskStateSafe, valueobject.RiskStateOneCycleMissed)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify no notifications were sent
		if len(pushProvider.sentNotifications) != 0 {
			t.Errorf("expected 0 notifications when disabled, got %d", len(pushProvider.sentNotifications))
		}
	})

	t.Run("handles no registered devices gracefully", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		err := svc.SendCriticalAlert(ctx, userID, "MyApp", "store.myshopify.com",
			valueobject.RiskStateSafe, valueobject.RiskStateOneCycleMissed)
		if err != nil {
			t.Fatalf("expected no error when no devices, got %v", err)
		}
	})
}

func TestNotificationService_SendDailySummary(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	t.Run("sends daily summary to all devices", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Register device
		_ = svc.RegisterDevice(ctx, userID, "token-1", entity.PlatformIOS)

		// Create snapshot
		snapshot := &entity.DailyMetricsSnapshot{
			ID:                 uuid.New(),
			AppID:              appID,
			Date:               time.Now().UTC().Truncate(24 * time.Hour),
			ActiveMRRCents:     500000, // $5,000
			RevenueAtRiskCents: 50000,  // $500
			RenewalSuccessRate: 0.95,
		}

		// Send daily summary
		err := svc.SendDailySummary(ctx, userID, "MyApp", snapshot)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify notification was sent
		if len(pushProvider.sentNotifications) != 1 {
			t.Errorf("expected 1 notification, got %d", len(pushProvider.sentNotifications))
		}

		notif := pushProvider.sentNotifications[0]
		if notif.title != "📊 Daily Summary: MyApp" {
			t.Errorf("unexpected title: %s", notif.title)
		}
		expectedBody := "MRR: $5000.00 | At Risk: $500.00 | Renewal Rate: 95.0%"
		if notif.body != expectedBody {
			t.Errorf("unexpected body: %s, expected: %s", notif.body, expectedBody)
		}
	})

	t.Run("respects disabled daily summary preference", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Register device
		_ = svc.RegisterDevice(ctx, userID, "token-1", entity.PlatformIOS)

		// Disable daily summary
		prefs := entity.NewNotificationPreferences(userID)
		prefs.DailySummaryEnabled = false
		_ = prefsRepo.Upsert(ctx, prefs)

		snapshot := &entity.DailyMetricsSnapshot{
			ID:                 uuid.New(),
			AppID:              appID,
			Date:               time.Now().UTC().Truncate(24 * time.Hour),
			ActiveMRRCents:     500000,
			RevenueAtRiskCents: 50000,
			RenewalSuccessRate: 0.95,
		}

		err := svc.SendDailySummary(ctx, userID, "MyApp", snapshot)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify no notifications were sent
		if len(pushProvider.sentNotifications) != 0 {
			t.Errorf("expected 0 notifications when disabled, got %d", len(pushProvider.sentNotifications))
		}
	})
}

func TestNotificationService_GetPreferences(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns existing preferences", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Create custom preferences
		prefs := entity.NewNotificationPreferences(userID)
		prefs.CriticalEnabled = false
		_ = prefsRepo.Create(ctx, prefs)

		// Get preferences
		result, err := svc.GetPreferences(ctx, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.CriticalEnabled != false {
			t.Error("expected critical alerts to be disabled")
		}
	})

	t.Run("returns default preferences when none exist", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Get preferences without creating any
		result, err := svc.GetPreferences(ctx, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should return defaults
		if !result.CriticalEnabled {
			t.Error("expected critical alerts enabled by default")
		}
		if !result.DailySummaryEnabled {
			t.Error("expected daily summary enabled by default")
		}
	})
}

func TestNotificationService_UpdatePreferences(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("updates preferences successfully", func(t *testing.T) {
		tokenRepo := newMockDeviceTokenRepository()
		prefsRepo := newMockNotificationPreferencesRepository()
		pushProvider := newMockPushNotificationProvider()
		svc := NewNotificationService(tokenRepo, prefsRepo, pushProvider)

		// Create and update preferences
		prefs := entity.NewNotificationPreferences(userID)
		prefs.CriticalEnabled = false
		prefs.DailySummaryEnabled = false

		err := svc.UpdatePreferences(ctx, prefs)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify update
		result, _ := prefsRepo.FindByUserID(ctx, userID)
		if result.CriticalEnabled {
			t.Error("expected critical alerts to be disabled after update")
		}
		if result.DailySummaryEnabled {
			t.Error("expected daily summary to be disabled after update")
		}
	})
}

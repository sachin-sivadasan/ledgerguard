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

// Mock AI Provider
type mockAIProvider struct {
	response string
	err      error
}

func (m *mockAIProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	return m.response, m.err
}

// Mock Daily Insight Repository
type mockDailyInsightRepo struct {
	insight *entity.DailyInsight
	err     error
}

func (m *mockDailyInsightRepo) Upsert(ctx context.Context, insight *entity.DailyInsight) error {
	m.insight = insight
	return m.err
}

func (m *mockDailyInsightRepo) FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyInsight, error) {
	return m.insight, m.err
}

func (m *mockDailyInsightRepo) FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyInsight, error) {
	if m.insight != nil {
		return []*entity.DailyInsight{m.insight}, nil
	}
	return nil, m.err
}

func (m *mockDailyInsightRepo) FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyInsight, error) {
	return m.insight, m.err
}

// Mock User Repository for plan tier check
type mockUserRepoForInsight struct {
	user *entity.User
	err  error
}

func (m *mockUserRepoForInsight) Create(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUserRepoForInsight) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.user, m.err
}

func (m *mockUserRepoForInsight) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	return m.user, m.err
}

func (m *mockUserRepoForInsight) Update(ctx context.Context, user *entity.User) error {
	return nil
}

func TestAIInsightService_GenerateInsight_Success(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	snapshot := &entity.DailyMetricsSnapshot{
		AppID:              appID,
		Date:               now,
		ActiveMRRCents:     500000, // $5,000
		RevenueAtRiskCents: 50000,  // $500
		RenewalSuccessRate: 0.85,
		SafeCount:          85,
		OneCycleMissedCount: 10,
		TwoCyclesMissedCount: 3,
		ChurnedCount:       2,
		TotalSubscriptions: 100,
	}

	user := &entity.User{
		ID:       userID,
		PlanTier: valueobject.PlanTierPro,
	}

	aiProvider := &mockAIProvider{
		response: "Your app shows strong performance with $5,000 MRR and 85% renewal success rate. However, 13 subscriptions are at risk representing $500 in potential lost revenue. Focus on re-engaging the 10 one-cycle-missed customers before they progress to higher risk states. Consider implementing automated retry logic for failed payments.",
	}
	insightRepo := &mockDailyInsightRepo{}
	userRepo := &mockUserRepoForInsight{user: user}

	service := NewAIInsightService(aiProvider, insightRepo, userRepo)

	insight, err := service.GenerateInsight(context.Background(), userID, appID, snapshot, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if insight == nil {
		t.Fatal("expected insight, got nil")
	}

	if insight.AppID != appID {
		t.Errorf("expected appID %s, got %s", appID, insight.AppID)
	}

	if insight.InsightText == "" {
		t.Error("expected insight text, got empty")
	}

	// Verify insight was stored
	if insightRepo.insight == nil {
		t.Error("expected insight to be stored")
	}
}

func TestAIInsightService_GenerateInsight_FreeTierBlocked(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	snapshot := &entity.DailyMetricsSnapshot{
		AppID: appID,
		Date:  now,
	}

	user := &entity.User{
		ID:       userID,
		PlanTier: valueobject.PlanTierFree, // FREE tier
	}

	aiProvider := &mockAIProvider{response: "This should not be called"}
	insightRepo := &mockDailyInsightRepo{}
	userRepo := &mockUserRepoForInsight{user: user}

	service := NewAIInsightService(aiProvider, insightRepo, userRepo)

	_, err := service.GenerateInsight(context.Background(), userID, appID, snapshot, now)
	if err == nil {
		t.Fatal("expected error for FREE tier, got nil")
	}

	if err != ErrProTierRequired {
		t.Errorf("expected ErrProTierRequired, got %v", err)
	}

	// Verify AI was not called (insight not stored)
	if insightRepo.insight != nil {
		t.Error("expected no insight to be stored for FREE tier")
	}
}

func TestAIInsightService_GenerateInsight_AIError(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	snapshot := &entity.DailyMetricsSnapshot{
		AppID: appID,
		Date:  now,
	}

	user := &entity.User{
		ID:       userID,
		PlanTier: valueobject.PlanTierPro,
	}

	aiProvider := &mockAIProvider{err: errors.New("AI service unavailable")}
	insightRepo := &mockDailyInsightRepo{}
	userRepo := &mockUserRepoForInsight{user: user}

	service := NewAIInsightService(aiProvider, insightRepo, userRepo)

	_, err := service.GenerateInsight(context.Background(), userID, appID, snapshot, now)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAIInsightService_GenerateInsight_UserNotFound(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	snapshot := &entity.DailyMetricsSnapshot{
		AppID: appID,
		Date:  now,
	}

	aiProvider := &mockAIProvider{}
	insightRepo := &mockDailyInsightRepo{}
	userRepo := &mockUserRepoForInsight{err: errors.New("user not found")}

	service := NewAIInsightService(aiProvider, insightRepo, userRepo)

	_, err := service.GenerateInsight(context.Background(), userID, appID, snapshot, now)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAIInsightService_BuildPrompt(t *testing.T) {
	service := NewAIInsightService(nil, nil, nil)

	snapshot := &entity.DailyMetricsSnapshot{
		AppID:              uuid.New(),
		Date:               time.Date(2026, 2, 26, 0, 0, 0, 0, time.UTC),
		ActiveMRRCents:     500000,
		RevenueAtRiskCents: 50000,
		UsageRevenueCents:  10000,
		TotalRevenueCents:  560000,
		RenewalSuccessRate: 0.85,
		SafeCount:          85,
		OneCycleMissedCount: 10,
		TwoCyclesMissedCount: 3,
		ChurnedCount:       2,
		TotalSubscriptions: 100,
	}

	prompt := service.BuildPrompt(snapshot)

	// Verify prompt contains key data
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	// Check for key metrics in prompt
	expectedPhrases := []string{
		"$5000.00", // Active MRR
		"$500.00",  // Revenue at risk
		"85.00%",   // Renewal rate
		"100",      // Total subscriptions
	}

	for _, phrase := range expectedPhrases {
		if !contains(prompt, phrase) {
			t.Errorf("expected prompt to contain %q", phrase)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

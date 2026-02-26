package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// ErrProTierRequired is returned when a FREE tier user tries to access Pro features
var ErrProTierRequired = errors.New("pro tier required for AI insights")

// AIProvider interface for LLM API calls (OpenAI, Claude, etc.)
type AIProvider interface {
	GenerateCompletion(ctx context.Context, prompt string) (string, error)
}

// AIInsightService generates AI-powered daily insights for Pro users
type AIInsightService struct {
	aiProvider  AIProvider
	insightRepo repository.DailyInsightRepository
	userRepo    repository.UserRepository
}

func NewAIInsightService(
	aiProvider AIProvider,
	insightRepo repository.DailyInsightRepository,
	userRepo repository.UserRepository,
) *AIInsightService {
	return &AIInsightService{
		aiProvider:  aiProvider,
		insightRepo: insightRepo,
		userRepo:    userRepo,
	}
}

// GenerateInsight generates an AI insight from a metrics snapshot
// Returns ErrProTierRequired if user is not on Pro plan
func (s *AIInsightService) GenerateInsight(
	ctx context.Context,
	userID uuid.UUID,
	appID uuid.UUID,
	snapshot *entity.DailyMetricsSnapshot,
	now time.Time,
) (*entity.DailyInsight, error) {
	// Check user plan tier
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.PlanTier != valueobject.PlanTierPro {
		return nil, ErrProTierRequired
	}

	// Build prompt from snapshot
	prompt := s.BuildPrompt(snapshot)

	// Generate AI completion
	insightText, err := s.aiProvider.GenerateCompletion(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate insight: %w", err)
	}

	// Create and store insight
	insight := entity.NewDailyInsight(appID, now, insightText)

	if err := s.insightRepo.Upsert(ctx, insight); err != nil {
		return nil, fmt.Errorf("failed to store insight: %w", err)
	}

	return insight, nil
}

// BuildPrompt constructs the AI prompt from a metrics snapshot
func (s *AIInsightService) BuildPrompt(snapshot *entity.DailyMetricsSnapshot) string {
	// Format currency values
	activeMRR := float64(snapshot.ActiveMRRCents) / 100
	revenueAtRisk := float64(snapshot.RevenueAtRiskCents) / 100
	usageRevenue := float64(snapshot.UsageRevenueCents) / 100
	totalRevenue := float64(snapshot.TotalRevenueCents) / 100
	renewalRate := snapshot.RenewalSuccessRate * 100

	prompt := fmt.Sprintf(`You are a revenue intelligence analyst for a Shopify app developer.

Generate a concise executive brief (80-120 words) based on these daily metrics:

## Key Metrics for %s

**Revenue:**
- Active MRR: $%.2f
- Revenue at Risk: $%.2f
- Usage Revenue: $%.2f
- Total Revenue (12mo): $%.2f

**Subscription Health:**
- Renewal Success Rate: %.2f%%
- Total Subscriptions: %d
  - Safe: %d
  - One Cycle Missed: %d
  - Two Cycles Missed: %d
  - Churned: %d

## Instructions:
1. Summarize the overall health in one sentence
2. Highlight the most critical insight or trend
3. Provide ONE actionable recommendation
4. Keep the tone professional but conversational
5. Do NOT use bullet points in your response
6. Write exactly 80-120 words`,
		snapshot.Date.Format("January 2, 2006"),
		activeMRR,
		revenueAtRisk,
		usageRevenue,
		totalRevenue,
		renewalRate,
		snapshot.TotalSubscriptions,
		snapshot.SafeCount,
		snapshot.OneCycleMissedCount,
		snapshot.TwoCyclesMissedCount,
		snapshot.ChurnedCount,
	)

	return prompt
}

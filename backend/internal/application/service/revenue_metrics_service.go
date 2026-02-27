package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// RevenueMode specifies how revenue data should be returned
type RevenueMode string

const (
	RevenueModeCombined RevenueMode = "combined"
	RevenueModeSplit    RevenueMode = "split"
)

// RevenueMetricsService handles revenue aggregation logic
type RevenueMetricsService struct {
	revenueRepo repository.RevenueRepository
}

// NewRevenueMetricsService creates a new RevenueMetricsService
func NewRevenueMetricsService(revenueRepo repository.RevenueRepository) *RevenueMetricsService {
	return &RevenueMetricsService{
		revenueRepo: revenueRepo,
	}
}

// EarningsEntryResponse represents a single day's earnings in the API response
type EarningsEntryResponse struct {
	Date                    string `json:"date"`
	TotalAmountCents        int64  `json:"total_amount_cents"`
	SubscriptionAmountCents int64  `json:"subscription_amount_cents,omitempty"`
	UsageAmountCents        int64  `json:"usage_amount_cents,omitempty"`
}

// EarningsTimelineResponse represents the API response for monthly earnings
type EarningsTimelineResponse struct {
	Month    string                  `json:"month"`
	Earnings []EarningsEntryResponse `json:"earnings"`
}

// GetMonthlyEarnings retrieves earnings timeline for a specific month
func (s *RevenueMetricsService) GetMonthlyEarnings(
	ctx context.Context,
	appID uuid.UUID,
	year, month int,
	mode RevenueMode,
) (*EarningsTimelineResponse, error) {
	// Validate month
	if month < 1 || month > 12 {
		return nil, ErrInvalidMonth
	}

	// Don't allow future months
	now := time.Now()
	requestedMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if requestedMonth.After(currentMonth) {
		return nil, ErrFutureMonth
	}

	// Get aggregated revenue data from repository
	aggregations, err := s.revenueRepo.GetMonthlyRevenue(ctx, appID, year, month)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	response := &EarningsTimelineResponse{
		Month:    time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01"),
		Earnings: make([]EarningsEntryResponse, 0, len(aggregations)),
	}

	for _, agg := range aggregations {
		entry := EarningsEntryResponse{
			Date:             agg.Date,
			TotalAmountCents: agg.TotalAmountCents,
		}

		// Include split data only in split mode
		if mode == RevenueModeSplit {
			entry.SubscriptionAmountCents = agg.SubscriptionAmountCents
			entry.UsageAmountCents = agg.UsageAmountCents
		}

		response.Earnings = append(response.Earnings, entry)
	}

	return response, nil
}

// Errors
var (
	ErrInvalidMonth = &RevenueError{Message: "invalid month: must be 1-12"}
	ErrFutureMonth  = &RevenueError{Message: "cannot request future months"}
)

// RevenueError represents a revenue service error
type RevenueError struct {
	Message string
}

func (e *RevenueError) Error() string {
	return e.Message
}

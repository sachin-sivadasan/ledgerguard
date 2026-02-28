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

// EarningsTimelineResponse represents the API response for earnings timeline
type EarningsTimelineResponse struct {
	StartDate string                  `json:"start_date"`
	EndDate   string                  `json:"end_date"`
	Earnings  []EarningsEntryResponse `json:"earnings"`
}

// GetEarningsByDateRange retrieves earnings timeline for a date range
func (s *RevenueMetricsService) GetEarningsByDateRange(
	ctx context.Context,
	appID uuid.UUID,
	startDate, endDate time.Time,
	mode RevenueMode,
) (*EarningsTimelineResponse, error) {
	// Don't allow future end dates
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if endDate.After(today) {
		endDate = today
	}

	// Validate date range
	if startDate.After(endDate) {
		return nil, ErrInvalidDateRange
	}

	// Get aggregated revenue data from repository
	aggregations, err := s.revenueRepo.GetRevenueByDateRange(ctx, appID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	response := &EarningsTimelineResponse{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Earnings:  make([]EarningsEntryResponse, 0, len(aggregations)),
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
	ErrInvalidDateRange = &RevenueError{Message: "invalid date range: start date must be before end date"}
)

// RevenueError represents a revenue service error
type RevenueError struct {
	Message string
}

func (e *RevenueError) Error() string {
	return e.Message
}

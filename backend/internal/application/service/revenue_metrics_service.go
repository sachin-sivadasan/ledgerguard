package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

const upcomingAvailabilityDays = 30

// RevenueMode specifies how revenue data should be returned
type RevenueMode string

const (
	RevenueModeCombined RevenueMode = "combined"
	RevenueModeSplit    RevenueMode = "split"
)

// RevenueMetricsService handles revenue aggregation logic
type RevenueMetricsService struct {
	revenueRepo     repository.RevenueRepository
	transactionRepo repository.TransactionRepository
}

// NewRevenueMetricsService creates a new RevenueMetricsService
func NewRevenueMetricsService(revenueRepo repository.RevenueRepository) *RevenueMetricsService {
	return &RevenueMetricsService{
		revenueRepo: revenueRepo,
	}
}

// NewRevenueMetricsServiceWithTransactions creates a RevenueMetricsService with transaction repo
func NewRevenueMetricsServiceWithTransactions(
	revenueRepo repository.RevenueRepository,
	transactionRepo repository.TransactionRepository,
) *RevenueMetricsService {
	return &RevenueMetricsService{
		revenueRepo:     revenueRepo,
		transactionRepo: transactionRepo,
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

// EarningsStatusResponse represents earnings availability status
type EarningsStatusResponse struct {
	TotalPendingCents      int64                     `json:"total_pending_cents"`
	TotalAvailableCents    int64                     `json:"total_available_cents"`
	TotalPaidOutCents      int64                     `json:"total_paid_out_cents"`
	PendingByDate          []EarningsDateEntry       `json:"pending_by_date"`
	UpcomingAvailability   []EarningsDateEntry       `json:"upcoming_availability"`
}

// EarningsDateEntry represents earnings for a specific date
type EarningsDateEntry struct {
	Date        string `json:"date"`
	AmountCents int64  `json:"amount_cents"`
}

// GetEarningsStatus retrieves earnings availability status for an app
func (s *RevenueMetricsService) GetEarningsStatus(ctx context.Context, appID uuid.UUID) (*EarningsStatusResponse, error) {
	if s.transactionRepo == nil {
		return nil, ErrTransactionRepoRequired
	}

	// Get summary totals
	summary, err := s.transactionRepo.GetEarningsSummary(ctx, appID)
	if err != nil {
		return nil, err
	}

	// Get pending by available date
	pendingByDate, err := s.transactionRepo.GetPendingByAvailableDate(ctx, appID)
	if err != nil {
		return nil, err
	}

	// Get upcoming availability (next 30 days)
	upcoming, err := s.transactionRepo.GetUpcomingAvailability(ctx, appID, upcomingAvailabilityDays)
	if err != nil {
		return nil, err
	}

	// Build response
	response := &EarningsStatusResponse{
		TotalPendingCents:   summary.PendingCents,
		TotalAvailableCents: summary.AvailableCents,
		TotalPaidOutCents:   summary.PaidOutCents,
		PendingByDate:       make([]EarningsDateEntry, 0, len(pendingByDate)),
		UpcomingAvailability: make([]EarningsDateEntry, 0, len(upcoming)),
	}

	for _, entry := range pendingByDate {
		response.PendingByDate = append(response.PendingByDate, EarningsDateEntry{
			Date:        entry.Date.Format("2006-01-02"),
			AmountCents: entry.AmountCents,
		})
	}

	for _, entry := range upcoming {
		response.UpcomingAvailability = append(response.UpcomingAvailability, EarningsDateEntry{
			Date:        entry.Date.Format("2006-01-02"),
			AmountCents: entry.AmountCents,
		})
	}

	return response, nil
}

// StoreRevenue represents revenue for a single store
type StoreRevenue struct {
	Domain           string  `json:"domain"`
	ShopName         string  `json:"shop_name"`
	RevenueCents     int64   `json:"revenue_cents"`
	TransactionCount int     `json:"transaction_count"`
	PctOfTotal       float64 `json:"pct_of_total"`
}

// RevenueConcentrationResponse represents the concentration analysis
type RevenueConcentrationResponse struct {
	TotalRevenueCents int64          `json:"total_revenue_cents"`
	Stores            []StoreRevenue `json:"stores"`
}

// GetRevenueConcentration returns top stores by revenue for an app
func (s *RevenueMetricsService) GetRevenueConcentration(
	ctx context.Context,
	appID uuid.UUID,
	start, end time.Time,
	limit int,
) (*RevenueConcentrationResponse, error) {
	if s.transactionRepo == nil {
		return nil, ErrTransactionRepoRequired
	}

	transactions, err := s.transactionRepo.FindByAppID(ctx, appID, start, end)
	if err != nil {
		return nil, err
	}

	// Group by domain
	type domainAgg struct {
		shopName string
		revenue  int64
		count    int
	}
	byDomain := make(map[string]*domainAgg)
	var totalRevenue int64

	for _, tx := range transactions {
		if tx.NetAmountCents <= 0 {
			continue
		}
		domain := tx.MyshopifyDomain
		if domain == "" {
			domain = "unknown"
		}
		agg, ok := byDomain[domain]
		if !ok {
			agg = &domainAgg{shopName: tx.ShopName}
			byDomain[domain] = agg
		}
		agg.revenue += tx.NetAmountCents
		agg.count++
		totalRevenue += tx.NetAmountCents
	}

	// Sort by revenue descending
	stores := make([]StoreRevenue, 0, len(byDomain))
	for domain, agg := range byDomain {
		var pct float64
		if totalRevenue > 0 {
			pct = float64(agg.revenue) / float64(totalRevenue) * 100
		}
		stores = append(stores, StoreRevenue{
			Domain:           domain,
			ShopName:         agg.shopName,
			RevenueCents:     agg.revenue,
			TransactionCount: agg.count,
			PctOfTotal:       pct,
		})
	}

	// Sort descending by revenue
	for i := 0; i < len(stores); i++ {
		for j := i + 1; j < len(stores); j++ {
			if stores[j].RevenueCents > stores[i].RevenueCents {
				stores[i], stores[j] = stores[j], stores[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(stores) > limit {
		stores = stores[:limit]
	}

	return &RevenueConcentrationResponse{
		TotalRevenueCents: totalRevenue,
		Stores:            stores,
	}, nil
}

// Errors
var (
	ErrInvalidDateRange       = &RevenueError{Message: "invalid date range: start date must be before end date"}
	ErrTransactionRepoRequired = &RevenueError{Message: "transaction repository required for earnings status"}
)

// RevenueError represents a revenue service error
type RevenueError struct {
	Message string
}

func (e *RevenueError) Error() string {
	return e.Message
}

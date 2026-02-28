package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RevenueAggregation represents a single date's aggregated revenue data
type RevenueAggregation struct {
	Date                    string // "YYYY-MM-DD"
	TotalAmountCents        int64
	SubscriptionAmountCents int64
	UsageAmountCents        int64
}

// RevenueRepository defines the interface for revenue data access
type RevenueRepository interface {
	// GetRevenueByDateRange retrieves aggregated revenue data for a date range
	// Groups transactions by date and sums amounts by charge type
	GetRevenueByDateRange(ctx context.Context, appID uuid.UUID, startDate, endDate time.Time) ([]RevenueAggregation, error)
}

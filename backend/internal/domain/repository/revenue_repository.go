package repository

import (
	"context"

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
	// GetMonthlyRevenue retrieves aggregated revenue data for a specific month
	// Groups transactions by date and sums amounts by charge type
	GetMonthlyRevenue(ctx context.Context, appID uuid.UUID, year, month int) ([]RevenueAggregation, error)
}

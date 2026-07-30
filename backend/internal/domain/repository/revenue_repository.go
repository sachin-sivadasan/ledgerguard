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

// MonthlyEarningAggregation is one month's raw earnings rollup (gross/net plus the
// per-earnings-status counts used to derive a single month-level status badge).
type MonthlyEarningAggregation struct {
	MonthLabel     string // "May 2026" (display)
	StartDate      string // "2026-05-01"
	EndDate        string // "2026-05-31"
	GrossCents     int64
	NetCents       int64
	PendingCount   int64
	AvailableCount int64
	PaidOutCount   int64
}

// RevenueRepository defines the interface for revenue data access
type RevenueRepository interface {
	// GetRevenueByDateRange retrieves aggregated revenue data for a date range
	// Groups transactions by date and sums amounts by charge type
	GetRevenueByDateRange(ctx context.Context, appID uuid.UUID, startDate, endDate time.Time) ([]RevenueAggregation, error)

	// GetMonthlyEarnings rolls transactions up by calendar month with gross + net
	// totals and per-status counts, newest month first. Powers the Earnings page's
	// monthly period cards (gross / Shopify fee / net + status badge).
	GetMonthlyEarnings(ctx context.Context, appID uuid.UUID, startDate, endDate time.Time) ([]MonthlyEarningAggregation, error)
}

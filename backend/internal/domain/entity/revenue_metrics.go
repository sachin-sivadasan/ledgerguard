package entity

import "time"

// RevenueEntry represents a single day's aggregated revenue
type RevenueEntry struct {
	Date                    time.Time
	TotalAmountCents        int64
	SubscriptionAmountCents int64
	UsageAmountCents        int64
}

// RevenueMetrics represents monthly revenue aggregation
type RevenueMetrics struct {
	Month    string         // Format: "YYYY-MM"
	Revenues []RevenueEntry // Sorted by Date ascending
}

// NewRevenueMetrics creates a new RevenueMetrics for a given month
func NewRevenueMetrics(year, month int) *RevenueMetrics {
	return &RevenueMetrics{
		Month:    formatMonth(year, month),
		Revenues: make([]RevenueEntry, 0),
	}
}

// AddRevenue adds a revenue entry to the metrics
func (rm *RevenueMetrics) AddRevenue(entry RevenueEntry) {
	rm.Revenues = append(rm.Revenues, entry)
}

// TotalRevenue returns the total revenue for the month
func (rm *RevenueMetrics) TotalRevenue() int64 {
	var total int64
	for _, r := range rm.Revenues {
		total += r.TotalAmountCents
	}
	return total
}

// TotalSubscriptionRevenue returns total subscription revenue for the month
func (rm *RevenueMetrics) TotalSubscriptionRevenue() int64 {
	var total int64
	for _, r := range rm.Revenues {
		total += r.SubscriptionAmountCents
	}
	return total
}

// TotalUsageRevenue returns total usage revenue for the month
func (rm *RevenueMetrics) TotalUsageRevenue() int64 {
	var total int64
	for _, r := range rm.Revenues {
		total += r.UsageAmountCents
	}
	return total
}

func formatMonth(year, month int) string {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

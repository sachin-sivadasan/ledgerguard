package service

import (
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// DefaultEarningsDelayDays is the standard delay for earnings availability
// Shopify holds earnings for 7-37 days depending on charge type and billing cycle
// We use 7 days as the default (minimum) for simplicity in MVP
const DefaultEarningsDelayDays = 7

// EarningsCalculator handles earnings availability calculations
// It determines when transaction earnings become available for payout
type EarningsCalculator struct{}

// NewEarningsCalculator creates a new EarningsCalculator
func NewEarningsCalculator() *EarningsCalculator {
	return &EarningsCalculator{}
}

// EarningsSummary contains aggregated earnings information
type EarningsSummary struct {
	PendingCents   int64 // Earnings not yet available
	AvailableCents int64 // Earnings ready for payout
	PaidOutCents   int64 // Earnings already disbursed
	PendingCount   int   // Number of pending transactions
	AvailableCount int   // Number of available transactions
	PaidOutCount   int   // Number of paid out transactions
}

// TotalCents returns the total earnings across all statuses
func (s EarningsSummary) TotalCents() int64 {
	return s.PendingCents + s.AvailableCents + s.PaidOutCents
}

// TotalCount returns the total number of transactions
func (s EarningsSummary) TotalCount() int {
	return s.PendingCount + s.AvailableCount + s.PaidOutCount
}

// CalculateAvailableDate determines when earnings become available based on charge type
// - RECURRING: 7 days after creation (Shopify holds 7-37 days, we use minimum)
// - ONE_TIME: 7 days after creation
// - USAGE: 7 days after creation
// - REFUND: Immediate (same day)
func (c *EarningsCalculator) CalculateAvailableDate(chargeType valueobject.ChargeType, createdDate time.Time) time.Time {
	switch chargeType {
	case valueobject.ChargeTypeRefund:
		// Refunds are processed immediately
		return createdDate
	default:
		// All other charge types have a 7-day delay
		return createdDate.AddDate(0, 0, DefaultEarningsDelayDays)
	}
}

// DetermineEarningsStatus determines the status based on current time vs available date
// - PENDING: now < availableDate
// - AVAILABLE: now >= availableDate
func (c *EarningsCalculator) DetermineEarningsStatus(availableDate time.Time, now time.Time) entity.EarningsStatus {
	if now.Before(availableDate) {
		return entity.EarningsStatusPending
	}
	return entity.EarningsStatusAvailable
}

// ProcessTransaction sets earnings tracking fields on a transaction
// It calculates the available date and determines current status
func (c *EarningsCalculator) ProcessTransaction(tx *entity.Transaction, createdDate time.Time, now time.Time) {
	availableDate := c.CalculateAvailableDate(tx.ChargeType, createdDate)
	status := c.DetermineEarningsStatus(availableDate, now)

	tx.SetEarningsTracking(createdDate, availableDate, status)
}

// ProcessTransactions sets earnings tracking for multiple transactions
func (c *EarningsCalculator) ProcessTransactions(txs []*entity.Transaction, now time.Time) {
	for _, tx := range txs {
		// Use TransactionDate as the created date if CreatedDate is not set
		createdDate := tx.CreatedDate
		if createdDate.IsZero() {
			createdDate = tx.TransactionDate
		}
		c.ProcessTransaction(tx, createdDate, now)
	}
}

// SummarizeEarnings calculates aggregate earnings by status
func (c *EarningsCalculator) SummarizeEarnings(txs []*entity.Transaction) EarningsSummary {
	summary := EarningsSummary{}

	for _, tx := range txs {
		switch tx.EarningsStatus {
		case entity.EarningsStatusPending:
			summary.PendingCents += tx.NetAmountCents
			summary.PendingCount++
		case entity.EarningsStatusAvailable:
			summary.AvailableCents += tx.NetAmountCents
			summary.AvailableCount++
		case entity.EarningsStatusPaidOut:
			summary.PaidOutCents += tx.NetAmountCents
			summary.PaidOutCount++
		}
	}

	return summary
}

// UpdateStatuses updates the earnings status of transactions based on current time
// This is useful for batch updates when time has passed
func (c *EarningsCalculator) UpdateStatuses(txs []*entity.Transaction, now time.Time) {
	for _, tx := range txs {
		tx.UpdateEarningsStatus(now)
	}
}

package valueobject

import "time"

type BillingInterval string

const (
	BillingIntervalMonthly BillingInterval = "MONTHLY"
	BillingIntervalAnnual  BillingInterval = "ANNUAL"
)

func (b BillingInterval) String() string {
	return string(b)
}

func (b BillingInterval) IsValid() bool {
	switch b {
	case BillingIntervalMonthly, BillingIntervalAnnual:
		return true
	}
	return false
}

// NextChargeDate calculates the next expected charge date from the last charge date
func (b BillingInterval) NextChargeDate(lastChargeDate time.Time) time.Time {
	switch b {
	case BillingIntervalMonthly:
		return lastChargeDate.AddDate(0, 1, 0) // +1 month
	case BillingIntervalAnnual:
		return lastChargeDate.AddDate(1, 0, 0) // +1 year
	default:
		return lastChargeDate.AddDate(0, 1, 0) // Default to monthly
	}
}

// DaysInCycle returns the number of days in a billing cycle
func (b BillingInterval) DaysInCycle() int {
	switch b {
	case BillingIntervalMonthly:
		return 30
	case BillingIntervalAnnual:
		return 365
	default:
		return 30
	}
}

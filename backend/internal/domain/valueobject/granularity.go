package valueobject

import "strings"

// Granularity represents the time granularity for metrics trend data.
// Daily snapshots are the atomic unit; weekly and monthly are derived at query time.
type Granularity string

const (
	GranularityDaily   Granularity = "DAILY"
	GranularityWeekly  Granularity = "WEEKLY"
	GranularityMonthly Granularity = "MONTHLY"
)

// ParseGranularity converts a string to a Granularity, defaulting to DAILY.
func ParseGranularity(s string) Granularity {
	switch strings.ToUpper(s) {
	case "WEEKLY":
		return GranularityWeekly
	case "MONTHLY":
		return GranularityMonthly
	case "DAILY", "":
		return GranularityDaily
	default:
		return GranularityDaily
	}
}

// IsValid returns true if the granularity is a recognized value.
func (g Granularity) IsValid() bool {
	switch g {
	case GranularityDaily, GranularityWeekly, GranularityMonthly:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (g Granularity) String() string {
	return string(g)
}

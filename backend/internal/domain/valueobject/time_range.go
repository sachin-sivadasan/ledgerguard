package valueobject

import (
	"time"
)

// TimeRangePreset represents predefined time range options
type TimeRangePreset string

const (
	TimeRangeThisMonth  TimeRangePreset = "THIS_MONTH"
	TimeRangeLastMonth  TimeRangePreset = "LAST_MONTH"
	TimeRangeLast30Days TimeRangePreset = "LAST_30_DAYS"
	TimeRangeLast90Days TimeRangePreset = "LAST_90_DAYS"
	TimeRangeCustom     TimeRangePreset = "CUSTOM"
)

func (t TimeRangePreset) String() string {
	return string(t)
}

func (t TimeRangePreset) IsValid() bool {
	switch t {
	case TimeRangeThisMonth, TimeRangeLastMonth, TimeRangeLast30Days, TimeRangeLast90Days, TimeRangeCustom:
		return true
	}
	return false
}

// DisplayName returns a human-readable name for the preset
func (t TimeRangePreset) DisplayName() string {
	switch t {
	case TimeRangeThisMonth:
		return "This Month"
	case TimeRangeLastMonth:
		return "Last Month"
	case TimeRangeLast30Days:
		return "Last 30 Days"
	case TimeRangeLast90Days:
		return "Last 90 Days"
	case TimeRangeCustom:
		return "Custom"
	default:
		return string(t)
	}
}

// DateRange represents a start and end date range
type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange creates a new DateRange with validation
func NewDateRange(start, end time.Time) DateRange {
	// Normalize to start of day in UTC
	start = truncateToDay(start)
	end = truncateToDay(end)

	// Swap if start > end
	if start.After(end) {
		start, end = end, start
	}

	return DateRange{Start: start, End: end}
}

// Days returns the number of days in the range (inclusive)
func (d DateRange) Days() int {
	return int(d.End.Sub(d.Start).Hours()/24) + 1
}

// PreviousPeriod returns a DateRange of equal length ending just before this range starts
func (d DateRange) PreviousPeriod() DateRange {
	days := d.Days()
	prevEnd := d.Start.AddDate(0, 0, -1)
	prevStart := prevEnd.AddDate(0, 0, -(days - 1))
	return DateRange{Start: prevStart, End: prevEnd}
}

// Contains checks if a date falls within this range (inclusive)
func (d DateRange) Contains(date time.Time) bool {
	date = truncateToDay(date)
	return !date.Before(d.Start) && !date.After(d.End)
}

// DateRangeForPreset calculates the date range for a given preset
func DateRangeForPreset(preset TimeRangePreset, now time.Time) DateRange {
	now = truncateToDay(now)

	switch preset {
	case TimeRangeThisMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return DateRange{Start: start, End: now}

	case TimeRangeLastMonth:
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := firstOfThisMonth.AddDate(0, 0, -1)
		start := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
		return DateRange{Start: start, End: end}

	case TimeRangeLast30Days:
		start := now.AddDate(0, 0, -29) // 30 days including today
		return DateRange{Start: start, End: now}

	case TimeRangeLast90Days:
		start := now.AddDate(0, 0, -89) // 90 days including today
		return DateRange{Start: start, End: now}

	case TimeRangeCustom:
		// For custom, return last 30 days as default
		start := now.AddDate(0, 0, -29)
		return DateRange{Start: start, End: now}

	default:
		// Default to this month
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return DateRange{Start: start, End: now}
	}
}

// truncateToDay normalizes a time to the start of day in UTC
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

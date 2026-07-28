package handler

import (
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// TrendInterval is the granularity of a report's time-series trend.
type TrendInterval string

const (
	IntervalDay   TrendInterval = "day"
	IntervalWeek  TrendInterval = "week"
	IntervalMonth TrendInterval = "month"
)

// resolveTrendInterval picks the trend granularity from the selected window span — the
// single ladder every report shares (mantle-brain: don't duplicate the ladder per
// screen). ≤ 31 days → daily, ≤ 92 days → weekly, otherwise monthly. This keeps the
// number of trend points bounded (a 12-month window → ~12 points, not ~365).
func resolveTrendInterval(from, to time.Time) TrendInterval {
	days := to.Sub(from).Hours() / 24
	switch {
	case days <= 31:
		return IntervalDay
	case days <= 92:
		return IntervalWeek
	default:
		return IntervalMonth
	}
}

// bucketKeyOf returns the bucket a time falls in, formatted as a full YYYY-MM-DD date so
// the frontend can always parse it: day → that day, week → its Monday, month → the first
// of the month. Bucket keys sort lexically in chronological order.
func bucketKeyOf(t time.Time, interval TrendInterval) string {
	switch interval {
	case IntervalWeek:
		return mondayOf(t).Format(dateLayout)
	case IntervalMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateLayout)
	default:
		return t.UTC().Format(dateLayout)
	}
}

// downsampleSnapshots collapses a daily-snapshot series to the given interval, keeping
// the LAST snapshot in each bucket. This is correct for as-of / stock metrics (MRR,
// renewal rate, churn rate) — a weekly/monthly point reflects the state at the end of the
// bucket. Input is assumed ascending by Date; output stays ascending, one per bucket.
func downsampleSnapshots(snapshots []*entity.DailyMetricsSnapshot, interval TrendInterval) []*entity.DailyMetricsSnapshot {
	if interval == IntervalDay || len(snapshots) == 0 {
		return snapshots
	}
	byBucket := map[string]*entity.DailyMetricsSnapshot{}
	order := make([]string, 0)
	for _, s := range snapshots {
		key := bucketKeyOf(s.Date, interval)
		if _, ok := byBucket[key]; !ok {
			order = append(order, key)
		}
		byBucket[key] = s // last-in-bucket wins (ascending input → end-of-bucket state)
	}
	out := make([]*entity.DailyMetricsSnapshot, 0, len(order))
	for _, key := range order {
		out = append(out, byBucket[key])
	}
	return out
}

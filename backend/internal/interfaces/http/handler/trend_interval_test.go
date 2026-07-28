package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// TestResolveTrendInterval pins the single range→interval ladder: ≤31d day, ≤92d week,
// else month.
func TestResolveTrendInterval(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		days int
		want TrendInterval
	}{
		{7, IntervalDay}, {30, IntervalDay}, {31, IntervalDay},
		{45, IntervalWeek}, {90, IntervalWeek}, {92, IntervalWeek},
		{120, IntervalMonth}, {365, IntervalMonth},
	}
	for _, c := range cases {
		if got := resolveTrendInterval(base, base.AddDate(0, 0, c.days)); got != c.want {
			t.Errorf("%d-day span: expected %q, got %q", c.days, c.want, got)
		}
	}
}

// TestBucketKeyOf pins bucket keys: day → the day; week → its Monday; month → the first.
func TestBucketKeyOf(t *testing.T) {
	// 2026-07-15 is a Wednesday → Monday is 2026-07-13.
	d := time.Date(2026, 7, 15, 18, 0, 0, 0, time.UTC)
	if got := bucketKeyOf(d, IntervalDay); got != "2026-07-15" {
		t.Errorf("day bucket: expected 2026-07-15, got %q", got)
	}
	if got := bucketKeyOf(d, IntervalWeek); got != "2026-07-13" {
		t.Errorf("week bucket: expected Monday 2026-07-13, got %q", got)
	}
	if got := bucketKeyOf(d, IntervalMonth); got != "2026-07-01" {
		t.Errorf("month bucket: expected 2026-07-01, got %q", got)
	}
}

// TestDownsampleSnapshots keeps the LAST snapshot per bucket (end-of-bucket, as-of state).
func TestDownsampleSnapshots(t *testing.T) {
	appID := uuid.New()
	snap := func(date time.Time, mrr int64) *entity.DailyMetricsSnapshot {
		return &entity.DailyMetricsSnapshot{ID: uuid.New(), AppID: appID, Date: date, ActiveMRRCents: mrr}
	}
	// Two July days + one August day, ascending.
	snaps := []*entity.DailyMetricsSnapshot{
		snap(time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC), 100),
		snap(time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC), 200), // later July → wins July bucket
		snap(time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC), 300),
	}
	out := downsampleSnapshots(snaps, IntervalMonth)
	if len(out) != 2 {
		t.Fatalf("expected 2 monthly buckets, got %d", len(out))
	}
	if out[0].ActiveMRRCents != 200 { // last-in-July
		t.Errorf("July bucket: expected the last snapshot (200), got %d", out[0].ActiveMRRCents)
	}
	if out[1].ActiveMRRCents != 300 {
		t.Errorf("Aug bucket: expected 300, got %d", out[1].ActiveMRRCents)
	}
	// Day interval is a passthrough.
	if got := downsampleSnapshots(snaps, IntervalDay); len(got) != 3 {
		t.Errorf("day interval should passthrough all 3, got %d", len(got))
	}
}

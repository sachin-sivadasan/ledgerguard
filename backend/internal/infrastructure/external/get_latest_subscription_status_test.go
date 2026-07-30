package external

import (
	"testing"
	"time"
)

// These tests pin down the "cancel trap": on Shopify, an upgrade/downgrade emits
// a SUBSCRIPTION_CHARGE_CANCELED (old plan) plus a SUBSCRIPTION_CHARGE_ACCEPTED
// (new plan). Reading the cancel as churn over-counts churn. Status must be
// decided by the latest event by OccurredAt (not slice order), with an
// accepted-charge winning a same-timestamp tie against a cancel.
func TestGetLatestSubscriptionStatus(t *testing.T) {
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	ev := func(typ string, t time.Time) AppEvent { return AppEvent{Type: typ, OccurredAt: t} }

	tests := []struct {
		name   string
		events []AppEvent
		want   string
	}{
		{
			name: "upgrade is not churn — later ACCEPTED wins over earlier CANCELED (slice order misleading)",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(2*time.Hour)), // old plan (listed first)
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(3*time.Hour)), // new plan (newer)
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base),
				ev("RELATIONSHIP_INSTALLED", base.Add(-1*time.Hour)),
			},
			want: "ACTIVE",
		},
		{
			name: "upgrade at same timestamp — ACCEPTED beats CANCELED on tie",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(2*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(2*time.Hour)),
			},
			want: "ACTIVE",
		},
		{
			name: "genuine churn — latest event is a CANCELED with no later ACCEPTED",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base),
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(3*time.Hour)),
			},
			want: "CANCELLED",
		},
		{
			name: "uninstall is terminal when it is the latest event",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base),
				ev("RELATIONSHIP_UNINSTALLED", base.Add(5*time.Hour)),
			},
			want: "UNINSTALLED",
		},
		{
			name: "win-back — reinstall + accept after an uninstall reads as ACTIVE",
			events: []AppEvent{
				ev("RELATIONSHIP_UNINSTALLED", base.Add(1*time.Hour)),
				ev("RELATIONSHIP_INSTALLED", base.Add(2*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(3*time.Hour)),
			},
			want: "ACTIVE",
		},
		{
			name: "unsorted input still resolves to the newest status",
			events: []AppEvent{
				ev("RELATIONSHIP_INSTALLED", base),
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(4*time.Hour)), // newest
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(1*time.Hour)),
			},
			want: "CANCELLED",
		},
		{
			name: "frozen is the latest event — reads FROZEN, not the older ACTIVE",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(1*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_FROZEN", base.Add(4*time.Hour)),
			},
			want: "FROZEN",
		},
		{
			name: "unfrozen restores ACTIVE",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_FROZEN", base.Add(2*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_UNFROZEN", base.Add(5*time.Hour)),
			},
			want: "ACTIVE",
		},
		{
			name: "accept then a LATER uninstall stays terminal (accept is not sticky)",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(1*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(2*time.Hour)),
				ev("RELATIONSHIP_UNINSTALLED", base.Add(6*time.Hour)),
			},
			want: "UNINSTALLED",
		},
		{
			name: "genuine churn — cancel later than an EARLIER accept (tie-break must not over-fire)",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(1*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(5*time.Hour)),
			},
			want: "CANCELLED",
		},
		{
			name: "multiple upgrades in sequence — latest accept wins",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base),
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(1*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(1*time.Hour)), // same-ts upgrade
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(2*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(3*time.Hour)), // newest
			},
			want: "ACTIVE",
		},
		{
			name: "all zero timestamps — degrades to priority order (accept wins over cancel)",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_CANCELED", time.Time{}),
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", time.Time{}),
				ev("RELATIONSHIP_INSTALLED", time.Time{}),
			},
			want: "ACTIVE",
		},
		{
			name: "unknown/future event type as latest falls through to the newest handled event",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base.Add(1*time.Hour)),
				ev("SOME_FUTURE_EVENT_TYPE", base.Add(9*time.Hour)), // newest but unhandled
			},
			want: "ACTIVE",
		},
		{
			name:   "only unhandled event types — unknown",
			events: []AppEvent{ev("SOME_FUTURE_EVENT_TYPE", base)},
			want:   "",
		},
		{
			name:   "installed only — pending",
			events: []AppEvent{ev("RELATIONSHIP_INSTALLED", base)},
			want:   "PENDING",
		},
		// --- missing-event-type regression: recurring renewals emit
		// SUBSCRIPTION_CHARGE_ACTIVATED (not ACCEPTED), and reinstalls emit
		// RELATIONSHIP_REACTIVATED (not INSTALLED). Before these were handled, an
		// active/reinstalled shop's newest event was skipped and the loop fell through
		// to a stale UNINSTALLED/CANCELED — mass-mislabelling live subs as churned.
		{
			name: "recurring activation is ACTIVE even after an older uninstall",
			events: []AppEvent{
				ev("RELATIONSHIP_UNINSTALLED", base.Add(1*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACTIVATED", base.Add(4*time.Hour)), // newest
			},
			want: "ACTIVE",
		},
		{
			name: "activated wins a same-timestamp tie against a cancel (plan change)",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_CANCELED", base.Add(2*time.Hour)),
				ev("SUBSCRIPTION_CHARGE_ACTIVATED", base.Add(2*time.Hour)),
			},
			want: "ACTIVE",
		},
		{
			name: "reactivation (reinstall) after an uninstall is not churn",
			events: []AppEvent{
				ev("RELATIONSHIP_UNINSTALLED", base.Add(1*time.Hour)),
				ev("RELATIONSHIP_REACTIVATED", base.Add(3*time.Hour)), // newest, no later charge yet
			},
			want: "PENDING",
		},
		{
			name: "relationship deactivated is terminal like uninstall",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACTIVATED", base.Add(1*time.Hour)),
				ev("RELATIONSHIP_DEACTIVATED", base.Add(5*time.Hour)),
			},
			want: "UNINSTALLED",
		},
		{
			name: "expired charge reads as cancelled when latest",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACCEPTED", base),
				ev("SUBSCRIPTION_CHARGE_EXPIRED", base.Add(3*time.Hour)),
			},
			want: "CANCELLED",
		},
		{
			name: "declined charge reads as a payment issue (frozen), recoverable",
			events: []AppEvent{
				ev("SUBSCRIPTION_CHARGE_ACTIVATED", base),
				ev("SUBSCRIPTION_CHARGE_DECLINED", base.Add(3*time.Hour)),
			},
			want: "FROZEN",
		},
		{
			name:   "empty",
			events: nil,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLatestSubscriptionStatus(tt.events); got != tt.want {
				t.Errorf("GetLatestSubscriptionStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

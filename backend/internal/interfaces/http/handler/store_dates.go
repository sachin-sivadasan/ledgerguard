package handler

import (
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// storeDates holds the real, event-sourced dates for a single shop: the earliest
// install and the most-recent interaction of any kind. Both are zero when the
// shop has no matching app events, so callers fall back to subscription dates.
type storeDates struct {
	firstInstall    time.Time
	lastInteraction time.Time
}

// isInstallEvent reports whether an event type marks a (re)install — the signals
// that establish a shop's install date. Note this intentionally counts
// RELATIONSHIP_REACTIVATED, unlike the installs *report* (installEventKind), which
// excludes it to avoid inflating new-install counts: here we only want the
// earliest available install-ish date, and a reactivation is a better estimate
// than the subscription fallback when no original INSTALLED event was fetched.
func isInstallEvent(eventType string) bool {
	return eventType == "RELATIONSHIP_INSTALLED" || eventType == "RELATIONSHIP_REACTIVATED"
}

// buildStoreDatesFromEvents indexes app events by shop identifier (the myshopify
// domain for charged shops, which is how EventProcessor stores them), deriving
// each shop's earliest install event and most-recent interaction. Returns an
// empty map for nil/empty input so callers can safely fall back.
func buildStoreDatesFromEvents(events []*entity.AppEvent) map[string]storeDates {
	dates := make(map[string]storeDates, len(events))
	for _, ev := range events {
		key := ev.ShopifyShopGID
		if key == "" {
			continue
		}
		d := dates[key]
		if isInstallEvent(ev.EventType) && (d.firstInstall.IsZero() || ev.OccurredAt.Before(d.firstInstall)) {
			d.firstInstall = ev.OccurredAt
		}
		if ev.OccurredAt.After(d.lastInteraction) {
			d.lastInteraction = ev.OccurredAt
		}
		dates[key] = d
	}
	return dates
}

// resolveFirstInstall picks the real install date: the earliest install event
// when known, else the subscription business start date (ActivatedAt/CreatedAt).
func resolveFirstInstall(eventInstall, subStart time.Time) time.Time {
	if !eventInstall.IsZero() {
		return eventInstall
	}
	return subStart
}

// resolveLastInteraction returns a shop's most-recent REAL activity — the latest
// of its last app event and its last recurring charge. The record UpdatedAt (a
// rebuild timestamp that is always ~now) is a last resort only, used when no real
// signal exists; it must never compete with real dates or it always wins.
func resolveLastInteraction(eventLast time.Time, lastCharge *time.Time, updatedAt time.Time) time.Time {
	real := eventLast
	if lastCharge != nil && lastCharge.After(real) {
		real = *lastCharge
	}
	if !real.IsZero() {
		return real
	}
	return updatedAt
}

package external

import (
	"testing"
	"time"
)

// TestCountInstalls pins the per-shop install state machine: active = latest
// relationship event is INSTALLED/REACTIVATED; total = shops ever installed.
func TestCountInstalls(t *testing.T) {
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	ev := func(typ, shop string, t time.Time) AppEvent {
		return AppEvent{Type: typ, ShopID: shop, OccurredAt: t}
	}

	events := []AppEvent{
		// s1: installed, still active
		ev("RELATIONSHIP_INSTALLED", "s1", base),
		// s2: installed then uninstalled -> inactive, but counts toward total
		ev("RELATIONSHIP_INSTALLED", "s2", base),
		ev("RELATIONSHIP_UNINSTALLED", "s2", base.Add(2*time.Hour)),
		// s3: installed, uninstalled, reactivated (out of order) -> active
		ev("RELATIONSHIP_UNINSTALLED", "s3", base.Add(2*time.Hour)),
		ev("RELATIONSHIP_INSTALLED", "s3", base),
		ev("RELATIONSHIP_REACTIVATED", "s3", base.Add(4*time.Hour)),
		// s4: installed then deactivated -> inactive
		ev("RELATIONSHIP_INSTALLED", "s4", base),
		ev("RELATIONSHIP_DEACTIVATED", "s4", base.Add(1*time.Hour)),
		// a non-relationship event is ignored
		ev("SUBSCRIPTION_CHARGE_ACTIVATED", "s1", base.Add(5*time.Hour)),
	}

	active, total := CountInstalls(events)
	if active != 2 { // s1, s3
		t.Errorf("active: expected 2, got %d", active)
	}
	if total != 4 { // s1, s2, s3, s4 all installed at some point
		t.Errorf("total: expected 4, got %d", total)
	}
}

func TestCountInstalls_Empty(t *testing.T) {
	a, tot := CountInstalls(nil)
	if a != 0 || tot != 0 {
		t.Errorf("empty: expected 0/0, got %d/%d", a, tot)
	}
}

// Reactivation-only (install was before our window) still counts as active, but
// not toward total-installed (no INSTALLED event seen).
func TestCountInstalls_ReactivatedOnly(t *testing.T) {
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	events := []AppEvent{{Type: "RELATIONSHIP_REACTIVATED", ShopID: "s9", OccurredAt: base}}
	active, total := CountInstalls(events)
	if active != 1 {
		t.Errorf("active: expected 1, got %d", active)
	}
	if total != 0 {
		t.Errorf("total: expected 0 (no INSTALLED event), got %d", total)
	}
}

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

// TestCountLifecycle pins the richer lifecycle snapshot (APPS-1b tiles).
func TestCountLifecycle(t *testing.T) {
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	ev := func(typ, shop string, t time.Time) AppEvent {
		return AppEvent{Type: typ, ShopID: shop, OccurredAt: t}
	}
	events := []AppEvent{
		ev("RELATIONSHIP_INSTALLED", "s1", base), // active
		ev("RELATIONSHIP_INSTALLED", "s2", base), // -> uninstalled
		ev("RELATIONSHIP_UNINSTALLED", "s2", base.Add(2*time.Hour)),
		ev("RELATIONSHIP_INSTALLED", "s3", base), // -> reactivated (active, returning)
		ev("RELATIONSHIP_REACTIVATED", "s3", base.Add(4*time.Hour)),
		ev("RELATIONSHIP_INSTALLED", "s4", base), // -> deactivated
		ev("RELATIONSHIP_DEACTIVATED", "s4", base.Add(1*time.Hour)),
		ev("SUBSCRIPTION_CHARGE_ACTIVATED", "s1", base.Add(5*time.Hour)), // ignored
	}

	lc := CountLifecycle(events)
	if lc.Active != 2 {
		t.Errorf("Active = %d, want 2 (s1, s3)", lc.Active)
	}
	if lc.EverInstalled != 4 {
		t.Errorf("EverInstalled = %d, want 4", lc.EverInstalled)
	}
	if lc.Uninstalled != 1 {
		t.Errorf("Uninstalled = %d, want 1 (s2)", lc.Uninstalled)
	}
	if lc.Deactivated != 1 {
		t.Errorf("Deactivated = %d, want 1 (s4)", lc.Deactivated)
	}
	if lc.Reactivated != 1 {
		t.Errorf("Reactivated = %d, want 1 (s3)", lc.Reactivated)
	}

	// CountInstalls stays consistent with the lifecycle snapshot.
	active, total := CountInstalls(events)
	if active != lc.Active || total != lc.EverInstalled {
		t.Errorf("CountInstalls (%d,%d) diverged from CountLifecycle (Active=%d, EverInstalled=%d)", active, total, lc.Active, lc.EverInstalled)
	}
}

// Same-timestamp install+uninstall must resolve deterministically (active wins),
// regardless of slice order — the API doesn't guarantee order.
func TestCountInstalls_SameTimestampTieBreak(t *testing.T) {
	at := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	// Order A: uninstall before install in the slice.
	a1, _ := CountInstalls([]AppEvent{
		{Type: "RELATIONSHIP_UNINSTALLED", ShopID: "s", OccurredAt: at},
		{Type: "RELATIONSHIP_INSTALLED", ShopID: "s", OccurredAt: at},
	})
	// Order B: install before uninstall.
	a2, _ := CountInstalls([]AppEvent{
		{Type: "RELATIONSHIP_INSTALLED", ShopID: "s", OccurredAt: at},
		{Type: "RELATIONSHIP_UNINSTALLED", ShopID: "s", OccurredAt: at},
	})
	if a1 != 1 || a2 != 1 {
		t.Errorf("same-timestamp tie-break not deterministic/active: got %d and %d, want 1 and 1", a1, a2)
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

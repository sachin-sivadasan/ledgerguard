package handler

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

func ev(shop, eventType string, occurredAt time.Time) *entity.AppEvent {
	return &entity.AppEvent{
		ID:             uuid.New(),
		AppID:          uuid.New(),
		ShopifyShopGID: shop,
		EventType:      eventType,
		OccurredAt:     occurredAt,
	}
}

func TestBuildStoreDatesFromEvents(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		// shop-a: reinstalled after an uninstall — earliest install wins for install date,
		// latest event (uninstall) wins for last interaction.
		ev("shop-a", "RELATIONSHIP_INSTALLED", base),
		ev("shop-a", "SUBSCRIPTION_CHARGE_ACTIVATED", base.AddDate(0, 1, 0)),
		ev("shop-a", "RELATIONSHIP_REACTIVATED", base.AddDate(0, 2, 0)),
		ev("shop-a", "RELATIONSHIP_UNINSTALLED", base.AddDate(0, 3, 0)),
		// shop-b: only a non-install event → firstInstall stays zero, lastInteraction set.
		ev("shop-b", "SUBSCRIPTION_CHARGE_CANCELED", base.AddDate(0, 5, 0)),
		// empty shop id is ignored.
		ev("", "RELATIONSHIP_INSTALLED", base),
	}

	dates := buildStoreDatesFromEvents(events)

	a := dates["shop-a"]
	if !a.firstInstall.Equal(base) {
		t.Errorf("shop-a firstInstall = %v, want %v (earliest install)", a.firstInstall, base)
	}
	if want := base.AddDate(0, 3, 0); !a.lastInteraction.Equal(want) {
		t.Errorf("shop-a lastInteraction = %v, want %v (latest event)", a.lastInteraction, want)
	}

	b := dates["shop-b"]
	if !b.firstInstall.IsZero() {
		t.Errorf("shop-b firstInstall = %v, want zero (no install event)", b.firstInstall)
	}
	if want := base.AddDate(0, 5, 0); !b.lastInteraction.Equal(want) {
		t.Errorf("shop-b lastInteraction = %v, want %v", b.lastInteraction, want)
	}

	if _, ok := dates[""]; ok {
		t.Error("empty shop id should be skipped")
	}
}

func TestBuildStoreDatesFromEvents_Empty(t *testing.T) {
	if got := buildStoreDatesFromEvents(nil); len(got) != 0 {
		t.Errorf("expected empty map for nil events, got %d entries", len(got))
	}
}

func TestResolveFirstInstall(t *testing.T) {
	subStart := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	eventInstall := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Event install date wins when present.
	if got := resolveFirstInstall(eventInstall, subStart); !got.Equal(eventInstall) {
		t.Errorf("with event = %v, want %v", got, eventInstall)
	}
	// Falls back to the subscription start date when no install event.
	if got := resolveFirstInstall(time.Time{}, subStart); !got.Equal(subStart) {
		t.Errorf("no event = %v, want %v (sub start)", got, subStart)
	}
}

func TestResolveLastInteraction(t *testing.T) {
	updated := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) // rebuild "today"
	eventLast := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	charge := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

	// A real event date must win over the always-recent rebuild UpdatedAt — the
	// core STORE-2 bug was UpdatedAt (≈now) beating real signals.
	if got := resolveLastInteraction(eventLast, nil, updated); !got.Equal(eventLast) {
		t.Errorf("event vs updated = %v, want %v (real beats rebuild time)", got, eventLast)
	}
	// Latest of event and charge wins.
	if got := resolveLastInteraction(eventLast, &charge, updated); !got.Equal(eventLast) {
		t.Errorf("event newer than charge = %v, want %v", got, eventLast)
	}
	newerCharge := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	if got := resolveLastInteraction(eventLast, &newerCharge, updated); !got.Equal(newerCharge) {
		t.Errorf("charge newer than event = %v, want %v", got, newerCharge)
	}
	// No real signal at all → record UpdatedAt as last resort.
	if got := resolveLastInteraction(time.Time{}, nil, updated); !got.Equal(updated) {
		t.Errorf("no real signal = %v, want %v (fallback)", got, updated)
	}
}

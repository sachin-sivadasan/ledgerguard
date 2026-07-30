package entity

import (
	"testing"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// TestSubscription_ApplyEventStatus pins the cancel-trap reconciliation: a terminal
// CANCELLED/UNINSTALLED event only churns when no recurring charge post-dates it,
// and risk is always re-derived from charge recency.
func TestSubscription_ApplyEventStatus(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	ptr := func(t time.Time) *time.Time { return &t }
	// A future expected-next-charge = actively billed (would grade SAFE).
	futureCharge := now.AddDate(0, 0, 23)

	t.Run("empty status is a no-op", func(t *testing.T) {
		s := &Subscription{Status: "ACTIVE", RiskState: valueobject.RiskStateSafe}
		if s.ApplyEventStatus("", time.Time{}, now) {
			t.Fatal("expected no change for empty status")
		}
	})

	t.Run("cancel billed AFTER the event is a stale/plan-change cancel -> ACTIVE + re-graded SAFE", func(t *testing.T) {
		cancelAt := now.AddDate(0, 0, -20)  // cancel 20d ago
		lastCharge := now.AddDate(0, 0, -8) // ...but billed 8d ago (after the cancel)
		s := &Subscription{
			Status:                  "ACTIVE",
			RiskState:               valueobject.RiskStateChurned, // stale
			LastRecurringChargeDate: ptr(lastCharge),
			ExpectedNextChargeDate:  ptr(futureCharge),
		}
		changed := s.ApplyEventStatus("CANCELLED", cancelAt, now)
		if !changed {
			t.Fatal("expected a change")
		}
		if s.Status != "ACTIVE" {
			t.Errorf("status: got %q, want ACTIVE (billed after cancel)", s.Status)
		}
		if s.RiskState != valueobject.RiskStateSafe {
			t.Errorf("risk: got %q, want SAFE", s.RiskState)
		}
	})

	t.Run("genuine cancel with no charge after the event -> CANCELLED + CHURNED", func(t *testing.T) {
		cancelAt := now.AddDate(0, 0, -2)    // cancel 2d ago
		lastCharge := now.AddDate(0, 0, -40) // last charge well before the cancel
		s := &Subscription{
			Status:                  "ACTIVE",
			RiskState:               valueobject.RiskStateSafe,
			LastRecurringChargeDate: ptr(lastCharge),
			ExpectedNextChargeDate:  ptr(now.AddDate(0, 0, -10)),
		}
		if !s.ApplyEventStatus("CANCELLED", cancelAt, now) {
			t.Fatal("expected a change")
		}
		if s.Status != "CANCELLED" || s.RiskState != valueobject.RiskStateChurned {
			t.Errorf("got %q/%q, want CANCELLED/CHURNED", s.Status, s.RiskState)
		}
	})

	t.Run("uninstall with no charge to reconcile against -> CHURNED", func(t *testing.T) {
		s := &Subscription{Status: "ACTIVE", RiskState: valueobject.RiskStateSafe}
		if !s.ApplyEventStatus("UNINSTALLED", now.AddDate(0, 0, -1), now) {
			t.Fatal("expected a change")
		}
		if s.Status != "UNINSTALLED" || s.RiskState != valueobject.RiskStateChurned {
			t.Errorf("got %q/%q, want UNINSTALLED/CHURNED", s.Status, s.RiskState)
		}
	})

	t.Run("re-classifies risk even when status is unchanged (fixes stale ACTIVE/CHURNED)", func(t *testing.T) {
		s := &Subscription{
			Status:                 "ACTIVE",
			RiskState:              valueobject.RiskStateChurned, // stale, contradicts ACTIVE
			ExpectedNextChargeDate: ptr(futureCharge),
		}
		changed := s.ApplyEventStatus("ACTIVE", now.AddDate(0, 0, -1), now)
		if !changed {
			t.Fatal("expected risk re-classification even though status was already ACTIVE")
		}
		if s.RiskState != valueobject.RiskStateSafe {
			t.Errorf("risk: got %q, want SAFE", s.RiskState)
		}
	})
}

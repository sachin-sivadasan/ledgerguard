package handler

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func labSub(plan string, interval valueobject.BillingInterval, priceCents int64) *entity.Subscription {
	return &entity.Subscription{
		ID:              uuid.New(),
		PlanName:        plan,
		BillingInterval: interval,
		BasePriceCents:  priceCents,
	}
}

func TestPlanLabeler_RealNameWins(t *testing.T) {
	l := newPlanLabeler(map[string]string{"MONTHLY:13999": "Should not be used"})
	got := l.label(labSub("Growth", valueobject.BillingIntervalMonthly, 13999))
	if got != "Growth" {
		t.Errorf("label = %q, want the real PlanName %q", got, "Growth")
	}
}

func TestPlanLabeler_PseudoLabelWhenEmpty(t *testing.T) {
	l := newPlanLabeler(nil)
	cases := []struct {
		interval valueobject.BillingInterval
		price    int64
		want     string
	}{
		{valueobject.BillingIntervalMonthly, 13999, "$139.99/mo"},
		{valueobject.BillingIntervalAnnual, 140000, "$1400.00/yr"},
		{valueobject.BillingIntervalMonthly, 0, "Free / unknown"},
	}
	for _, c := range cases {
		if got := l.label(labSub("", c.interval, c.price)); got != c.want {
			t.Errorf("pseudo label(%s,%d) = %q, want %q", c.interval, c.price, got, c.want)
		}
	}
}

func TestPlanLabeler_MapOverridesPseudo(t *testing.T) {
	// A mid-life price change: same "Starter" tier at two prices → two distinct keys, each
	// nameable independently. This is the "Starter" vs "Starter (old)" case.
	l := newPlanLabeler(map[string]string{
		"MONTHLY:2900": "Starter",
		"MONTHLY:1900": "Starter (old)",
	})
	if got := l.label(labSub("", valueobject.BillingIntervalMonthly, 2900)); got != "Starter" {
		t.Errorf("label($29) = %q, want Starter", got)
	}
	if got := l.label(labSub("", valueobject.BillingIntervalMonthly, 1900)); got != "Starter (old)" {
		t.Errorf("label($19) = %q, want Starter (old)", got)
	}
	// A price with no mapping falls through to the pseudo-label.
	if got := l.label(labSub("", valueobject.BillingIntervalMonthly, 4900)); got != "$49.00/mo" {
		t.Errorf("unmapped label($49) = %q, want $49.00/mo", got)
	}
}

func TestPlanKey_DistinctPerPriceAndInterval(t *testing.T) {
	if planKey(valueobject.BillingIntervalMonthly, 2900) == planKey(valueobject.BillingIntervalMonthly, 1900) {
		t.Error("different prices must produce different plan keys")
	}
	if planKey(valueobject.BillingIntervalMonthly, 2900) == planKey(valueobject.BillingIntervalAnnual, 2900) {
		t.Error("different intervals must produce different plan keys")
	}
}

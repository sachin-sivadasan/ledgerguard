package handler

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// The Partner API's app-subscription transactions carry no plan NAME (only a chargeId GID
// + billing interval), so the ledger rebuild leaves Subscription.PlanName empty for most
// apps (see ledger_service.go). These helpers give plan-based reports a usable plan
// dimension anyway: a stable price-tier KEY plus a display LABEL that a developer can
// optionally override with a friendly name.

// planKey is the stable identity of a price tier: billing interval + exact base price.
// Distinct prices produce distinct keys, so a mid-life price change naturally splits into
// two plans (e.g. a $19 "Starter (old)" and a $29 "Starter") rather than merging them.
//
// The interval half uses the canonical BillingInterval enum string ("MONTHLY"/"ANNUAL");
// developer plan-label maps (Phase 2) must key on the same canonical values, and an empty
// interval collapses to a ":<cents>" key (treated as monthly downstream).
func planKey(interval valueobject.BillingInterval, basePriceCents int64) string {
	return string(interval) + ":" + strconv.FormatInt(basePriceCents, 10)
}

// pseudoPlanLabel synthesizes a human label from the price tier when there's no real plan
// name — e.g. "$139.99/mo" or "$1400.00/yr". A zero/negative price is a free or unknown
// tier. Only ANNUAL is special-cased; every other interval (including an empty/unknown
// one) defaults to "/mo", matching Subscription.MRRCents()'s monthly default.
func pseudoPlanLabel(interval valueobject.BillingInterval, basePriceCents int64) string {
	if basePriceCents <= 0 {
		return "Free / unknown"
	}
	amount := fmt.Sprintf("$%.2f", float64(basePriceCents)/100)
	if interval == valueobject.BillingIntervalAnnual {
		return amount + "/yr"
	}
	return amount + "/mo"
}

// planLabeler resolves a subscription to its display plan label, in priority order:
//  1. the real PlanName, when the sync captured one;
//  2. a developer-assigned name for the price tier (from the app's plan-label map);
//  3. a synthesized price-tier pseudo-label ("$139.99/mo").
//
// The map is keyed by planKey. A nil/empty map simply falls through to the pseudo-label,
// so the zero value is a valid pseudo-only labeler.
type planLabeler struct {
	labels map[string]string
}

// newPlanLabeler builds a labeler over an optional developer plan-label map (may be nil).
func newPlanLabeler(labels map[string]string) planLabeler {
	return planLabeler{labels: labels}
}

// isPseudoPlanLabel reports whether a label is a synthesized price-tier label (rather than
// a real or developer-assigned name). The tail-collapse never folds a NON-pseudo (named)
// tier — a developer who named a low-volume plan must still see it in the reports, matching
// the plan-labels settings which always keep named tiers.
func isPseudoPlanLabel(label string) bool {
	if label == "Free / unknown" {
		return true
	}
	return strings.HasPrefix(label, "$") &&
		(strings.HasSuffix(label, "/mo") || strings.HasSuffix(label, "/yr"))
}

// minTiersToCollapse is the tier count above which the long-tail collapse kicks in. Below
// it there's no proration noise worth hiding (a real app has only a handful of plans), so
// every tier is shown as-is.
const minTiersToCollapse = 12

// tierSignificanceThreshold is the minimum customer count for a price tier to count as a
// real plan rather than proration/refund noise. BasePriceCents is the last CHARGED amount,
// so mid-cycle upgrades (prorated partials) and refund/adjustment subs spawn many 1–2
// customer phantom tiers; the genuine plans carry far more. Floor of 3, scaled to 0.5% of
// the active base so it stays meaningful as an app grows.
func tierSignificanceThreshold(totalActive int) int {
	t := int(math.Ceil(float64(totalActive) * 0.005))
	if t < 3 {
		t = 3
	}
	return t
}

// planLabelMapFor loads an app's saved plan labels as a planKey→label map for the labeler.
// A nil repo (feature not wired) or a load error yields nil — plan labels are a display
// nicety, so the labeler must degrade to pseudo-labels rather than fail the report.
func planLabelMapFor(ctx context.Context, repo repository.PlanLabelRepository, appID uuid.UUID) map[string]string {
	if repo == nil {
		return nil
	}
	labels, err := repo.FindByAppID(ctx, appID)
	if err != nil {
		log.Printf("plan-labels: load map failed (falling back to pseudo-labels): %v", err)
		return nil
	}
	m := make(map[string]string, len(labels))
	for _, l := range labels {
		m[planKey(l.BillingInterval, l.PriceCents)] = l.Label
	}
	return m
}

// label returns the display plan label for a subscription.
func (l planLabeler) label(s *entity.Subscription) string {
	if s.PlanName != "" {
		return s.PlanName
	}
	if l.labels != nil {
		if name, ok := l.labels[planKey(s.BillingInterval, s.BasePriceCents)]; ok && name != "" {
			return name
		}
	}
	return pseudoPlanLabel(s.BillingInterval, s.BasePriceCents)
}

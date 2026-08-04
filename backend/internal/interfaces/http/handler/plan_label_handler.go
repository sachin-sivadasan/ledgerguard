package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// planLabelMaxLen caps a label to the DB column width; planLabelsMax caps the tier count in
// one save (an app has a handful of price tiers, never hundreds).
const (
	planLabelMaxLen = 120
	planLabelsMax   = 200
)

// PlanLabelHandler manages developer-assigned names for price tiers (see plan_label.go).
// GET returns the distinct price tiers present in the app's un-named subscriptions (the
// ones a pseudo-label is synthesized for) so the settings UI can name each; PUT saves the
// full set.
type PlanLabelHandler struct {
	subRepo       repository.SubscriptionRepository
	planLabelRepo repository.PlanLabelRepository
	appRepo       repository.AppRepository
	partnerRepo   repository.PartnerAccountRepository
}

// NewPlanLabelHandler constructs a PlanLabelHandler.
func NewPlanLabelHandler(
	subRepo repository.SubscriptionRepository,
	planLabelRepo repository.PlanLabelRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *PlanLabelHandler {
	return &PlanLabelHandler{subRepo: subRepo, planLabelRepo: planLabelRepo, appRepo: appRepo, partnerRepo: partnerRepo}
}

type planTierJSON struct {
	BillingInterval string `json:"billingInterval"`
	PriceCents      int64  `json:"priceCents"`
	Key             string `json:"key"`         // planKey — the stable tier identity
	PseudoLabel     string `json:"pseudoLabel"` // derived "$29.00/mo"
	Label           string `json:"label"`       // developer-assigned name, "" when unset
	Customers       int    `json:"customers"`   // active (non-churned) subs on this tier
}

type planLabelsResponse struct {
	Tiers []planTierJSON `json:"tiers"`
	// HiddenTiers is the number of minor price tiers omitted (prorations / one-off / refund
	// amounts below the significance threshold and never named) — surfaced so the UI can
	// note them rather than silently dropping them.
	HiddenTiers int `json:"hiddenTiers"`
}

type planLabelInput struct {
	BillingInterval string `json:"billingInterval"`
	PriceCents      int64  `json:"priceCents"`
	Label           string `json:"label"`
}

// GetPlanLabels lists the app's price tiers with their current (saved or pseudo) labels.
// GET /api/v1/apps/{appID}/plan-labels
func (h *PlanLabelHandler) GetPlanLabels(w http.ResponseWriter, r *http.Request) {
	if user := middleware.UserFromContext(r.Context()); user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	subs, err := h.subRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		log.Printf("plan-labels: repo error in FindByAppID(subs): %v", err)
		writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}
	saved, err := h.planLabelRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		log.Printf("plan-labels: repo error in FindByAppID(labels): %v", err)
		writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}
	savedByKey := make(map[string]string, len(saved))
	for _, l := range saved {
		savedByKey[planKey(l.BillingInterval, l.PriceCents)] = l.Label
	}

	// Distinct tiers among subs WITHOUT a real plan name (the ones that get a pseudo-label),
	// counting only non-churned customers (matching what the reports segment).
	type tierAgg struct {
		interval  valueobject.BillingInterval
		price     int64
		customers int
	}
	byKey := map[string]*tierAgg{}
	totalActive := 0
	for _, s := range subs {
		if s.PlanName != "" || s.RiskState.IsChurned() {
			continue
		}
		key := planKey(s.BillingInterval, s.BasePriceCents)
		a, ok := byKey[key]
		if !ok {
			a = &tierAgg{interval: s.BillingInterval, price: s.BasePriceCents}
			byKey[key] = a
		}
		a.customers++
		totalActive++
	}

	// Only surface tiers worth naming: a real plan (≥ the significance threshold and a
	// positive price) or one that already has a saved label. The rest (prorations, refund
	// amounts) are counted as hidden so the UI can note them.
	// Below a real long tail, every tier is nameable; past it, filter to the significant
	// ones (or already-named) and count the rest as hidden.
	collapse := len(byKey) > minTiersToCollapse
	threshold := tierSignificanceThreshold(totalActive)
	tiers := make([]planTierJSON, 0, len(byKey))
	hidden := 0
	for key, a := range byKey {
		saved := savedByKey[key]
		significant := !collapse || (a.price > 0 && a.customers >= threshold)
		if !significant && saved == "" {
			hidden++
			continue
		}
		tiers = append(tiers, planTierJSON{
			BillingInterval: string(a.interval),
			PriceCents:      a.price,
			Key:             key,
			PseudoLabel:     pseudoPlanLabel(a.interval, a.price),
			Label:           saved,
			Customers:       a.customers,
		})
	}
	// Highest-value tiers first (most customers, then price) so the important ones lead.
	sort.SliceStable(tiers, func(i, j int) bool {
		if tiers[i].Customers != tiers[j].Customers {
			return tiers[i].Customers > tiers[j].Customers
		}
		return tiers[i].PriceCents > tiers[j].PriceCents
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(planLabelsResponse{Tiers: tiers, HiddenTiers: hidden}); err != nil {
		log.Printf("plan-labels: encode: %v", err)
	}
}

// PutPlanLabels replaces the app's plan-label set with the submitted one.
// PUT /api/v1/apps/{appID}/plan-labels   body: {"labels":[{billingInterval,priceCents,label}]}
func (h *PlanLabelHandler) PutPlanLabels(w http.ResponseWriter, r *http.Request) {
	if user := middleware.UserFromContext(r.Context()); user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	var body struct {
		Labels []planLabelInput `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.Labels) > planLabelsMax {
		writeJSONError(w, http.StatusBadRequest, "too many tiers")
		return
	}

	labels := make([]*entity.PlanLabel, 0, len(body.Labels))
	seen := map[string]bool{}
	for _, in := range body.Labels {
		label := strings.TrimSpace(in.Label)
		if label == "" {
			continue // a cleared tier — simply absent from the replacement set
		}
		if len(label) > planLabelMaxLen {
			writeJSONError(w, http.StatusBadRequest, "label too long (max 120 chars)")
			return
		}
		if in.PriceCents < 0 {
			writeJSONError(w, http.StatusBadRequest, "priceCents must be non-negative")
			return
		}
		interval := valueobject.BillingInterval(in.BillingInterval)
		// Allow empty (some synced subs carry no interval — planKey treats it as monthly),
		// but reject a malformed non-empty value that could never match a real tier.
		if in.BillingInterval != "" && !interval.IsValid() {
			writeJSONError(w, http.StatusBadRequest, "invalid billingInterval")
			return
		}
		key := planKey(interval, in.PriceCents)
		if seen[key] {
			writeJSONError(w, http.StatusBadRequest, "duplicate tier in request")
			return
		}
		seen[key] = true
		labels = append(labels, &entity.PlanLabel{
			AppID:           app.ID,
			BillingInterval: interval,
			PriceCents:      in.PriceCents,
			Label:           label,
		})
	}

	if err := h.planLabelRepo.ReplaceAll(r.Context(), app.ID, labels); err != nil {
		log.Printf("plan-labels: repo error in ReplaceAll: %v", err)
		writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int{"saved": len(labels)}); err != nil {
		log.Printf("plan-labels: encode save result: %v", err)
	}
}

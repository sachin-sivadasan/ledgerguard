package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type RiskHandler struct {
	subscriptionRepo repository.SubscriptionRepository
	appEventRepo     repository.AppEventRepository // optional; nil-tolerant
	partnerRepo      repository.PartnerAccountRepository
	appRepo          repository.AppRepository
	riskEngine       *domainservice.RiskEngine
}

func NewRiskHandler(
	subscriptionRepo repository.SubscriptionRepository,
	appEventRepo repository.AppEventRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
	riskEngine *domainservice.RiskEngine,
) *RiskHandler {
	return &RiskHandler{
		subscriptionRepo: subscriptionRepo,
		appEventRepo:     appEventRepo,
		partnerRepo:      partnerRepo,
		appRepo:          appRepo,
		riskEngine:       riskEngine,
	}
}

// Summary returns the risk distribution and at-risk stores for an app.
// GET /api/v1/apps/{appID}/risk/summary
func (h *RiskHandler) Summary(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	subs, err := h.subscriptionRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch subscriptions")
		return
	}

	// Count the PERSISTED risk_state — the reconciled value written by the sync
	// pipeline (ledger rebuild + StatusProcessor.ApplyEventStatus cancel-trap
	// reconciliation), the same source Subscriptions /summary and Dashboard
	// /metrics read. Do NOT re-run RiskEngine.ClassifyAll here (RISK-1b): its
	// naive status→risk rule lacks the cancel-trap reconciliation and diverged
	// from the other two pages (safe 1,076 vs 1,061).
	summary := h.riskEngine.CalculateRiskSummary(subs)

	// Real install / last-interaction dates from the app-event stream; subscription
	// business dates are the fallback. CreatedAt/UpdatedAt are record timestamps
	// (reset on rebuild) and must not be used as install/interaction dates (STORE-2).
	var eventDates map[string]storeDates
	if h.appEventRepo != nil {
		if events, err := h.appEventRepo.FindByAppID(r.Context(), app.ID); err == nil {
			eventDates = buildStoreDatesFromEvents(events)
		}
	}

	// Build at-risk stores list
	type atRiskStoreJSON struct {
		ID                 string   `json:"id"`
		ShopDomain         string   `json:"shop_domain"`
		InstalledAppIDs    []string `json:"installed_app_ids"`
		HealthScore        int      `json:"health_score"`
		LifetimeValueCents int64    `json:"lifetime_value_cents"`
		FirstInstallDate   string   `json:"first_install_date"`
		LastInteraction    string   `json:"last_interaction"`
		RiskState          string   `json:"risk_state"`
	}

	atRiskStores := make([]atRiskStoreJSON, 0)
	for _, sub := range subs {
		if sub.RiskState == valueobject.RiskStateOneCycleMissed ||
			sub.RiskState == valueobject.RiskStateTwoCyclesMissed {
			ed := eventDates[sub.MyshopifyDomain] // zero-value when the shop has no events
			firstInstall := resolveFirstInstall(ed.firstInstall, sub.StartDate())
			lastInteraction := resolveLastInteraction(ed.lastInteraction, sub.LastRecurringChargeDate, sub.UpdatedAt)
			atRiskStores = append(atRiskStores, atRiskStoreJSON{
				ID:         sub.ID.String(),
				ShopDomain: sub.MyshopifyDomain,
				// Populate installed_app_ids with the current app (RISK-2) so this
				// endpoint's store serialization matches /stores (was empty []).
				InstalledAppIDs:    []string{app.ID.String()},
				HealthScore:        healthScoreFromRisk(sub.RiskState),
				LifetimeValueCents: sub.BasePriceCents,
				FirstInstallDate:   firstInstall.Format(time.RFC3339),
				LastInteraction:    lastInteraction.Format(time.RFC3339),
				RiskState:          string(sub.RiskState),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"distribution": map[string]int{
			"safe":      summary.SafeCount,
			"one_cycle": summary.OneCycleMissedCount,
			"two_cycle": summary.TwoCyclesMissedCount,
			"churned":   summary.ChurnedCount,
		},
		"at_risk_stores": atRiskStores,
	})
}

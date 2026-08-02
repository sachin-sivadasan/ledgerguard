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
	partnerRepo      repository.PartnerAccountRepository
	appRepo          repository.AppRepository
	riskEngine       *domainservice.RiskEngine
}

func NewRiskHandler(
	subscriptionRepo repository.SubscriptionRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
	riskEngine *domainservice.RiskEngine,
) *RiskHandler {
	return &RiskHandler{
		subscriptionRepo: subscriptionRepo,
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

	now := time.Now().UTC()
	h.riskEngine.ClassifyAll(subs, now)
	summary := h.riskEngine.CalculateRiskSummary(subs)

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
			atRiskStores = append(atRiskStores, atRiskStoreJSON{
				ID:                 sub.ID.String(),
				ShopDomain:         sub.MyshopifyDomain,
				InstalledAppIDs:    []string{},
				HealthScore:        healthScoreFromRisk(sub.RiskState),
				LifetimeValueCents: sub.BasePriceCents,
				FirstInstallDate:   sub.CreatedAt.Format(time.RFC3339),
				LastInteraction:    sub.UpdatedAt.Format(time.RFC3339),
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

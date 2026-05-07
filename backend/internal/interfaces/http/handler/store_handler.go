package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type StoreHandler struct {
	subscriptionRepo repository.SubscriptionRepository
	txRepo           repository.TransactionRepository
	partnerRepo      repository.PartnerAccountRepository
	appRepo          repository.AppRepository
	riskEngine       *domainservice.RiskEngine
}

func NewStoreHandler(
	subscriptionRepo repository.SubscriptionRepository,
	txRepo repository.TransactionRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
	riskEngine *domainservice.RiskEngine,
) *StoreHandler {
	return &StoreHandler{
		subscriptionRepo: subscriptionRepo,
		txRepo:           txRepo,
		partnerRepo:      partnerRepo,
		appRepo:          appRepo,
		riskEngine:       riskEngine,
	}
}

// List returns a deduplicated list of stores for an app, enriched with risk and LTV.
// GET /api/v1/apps/{appID}/stores
func (h *StoreHandler) List(w http.ResponseWriter, r *http.Request) {
	appID, lookupErr := lookupAppID(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	subs, err := h.subscriptionRepo.FindByAppID(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch subscriptions")
		return
	}

	now := time.Now().UTC()
	h.riskEngine.ClassifyAll(subs, now)

	// Group subscriptions by domain
	type storeInfo struct {
		domain       string
		riskState    valueobject.RiskState
		firstInstall time.Time
		lastActivity time.Time
	}
	storeMap := make(map[string]*storeInfo)
	for _, sub := range subs {
		domain := sub.MyshopifyDomain
		if domain == "" {
			continue
		}
		existing, ok := storeMap[domain]
		if !ok {
			storeMap[domain] = &storeInfo{
				domain:       domain,
				riskState:    sub.RiskState,
				firstInstall: sub.CreatedAt,
				lastActivity: sub.UpdatedAt,
			}
		} else {
			// Use worst risk state
			if riskPriority(sub.RiskState) > riskPriority(existing.riskState) {
				existing.riskState = sub.RiskState
			}
			if sub.CreatedAt.Before(existing.firstInstall) {
				existing.firstInstall = sub.CreatedAt
			}
			if sub.UpdatedAt.After(existing.lastActivity) {
				existing.lastActivity = sub.UpdatedAt
			}
		}
	}

	// Calculate LTV per domain from transactions (last 3 years)
	from := now.AddDate(-3, 0, 0)
	txs, _ := h.txRepo.FindByAppID(r.Context(), appID, from, now)
	ltvMap := make(map[string]int64)
	for _, tx := range txs {
		ltvMap[tx.MyshopifyDomain] += tx.NetAmountCents
	}

	type storeJSON struct {
		ID                string   `json:"id"`
		ShopDomain        string   `json:"shop_domain"`
		InstalledAppIDs   []string `json:"installed_app_ids"`
		HealthScore       int      `json:"health_score"`
		LifetimeValueCents int64   `json:"lifetime_value_cents"`
		FirstInstallDate  string   `json:"first_install_date"`
		LastInteraction   string   `json:"last_interaction"`
		RiskState         string   `json:"risk_state"`
	}

	items := make([]storeJSON, 0, len(storeMap))
	for domain, info := range storeMap {
		items = append(items, storeJSON{
			ID:                 domain,
			ShopDomain:         domain,
			InstalledAppIDs:    []string{appID.String()},
			HealthScore:        healthScoreFromRisk(info.riskState),
			LifetimeValueCents: ltvMap[domain],
			FirstInstallDate:   info.firstInstall.Format(time.RFC3339),
			LastInteraction:    info.lastActivity.Format(time.RFC3339),
			RiskState:          string(info.riskState),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stores": items,
	})
}

func riskPriority(rs valueobject.RiskState) int {
	switch rs {
	case valueobject.RiskStateSafe:
		return 0
	case valueobject.RiskStateOneCycleMissed:
		return 1
	case valueobject.RiskStateTwoCyclesMissed:
		return 2
	case valueobject.RiskStateChurned:
		return 3
	default:
		return 0
	}
}

func healthScoreFromRisk(rs valueobject.RiskState) int {
	switch rs {
	case valueobject.RiskStateSafe:
		return 90
	case valueobject.RiskStateOneCycleMissed:
		return 50
	case valueobject.RiskStateTwoCyclesMissed:
		return 25
	case valueobject.RiskStateChurned:
		return 10
	default:
		return 50
	}
}

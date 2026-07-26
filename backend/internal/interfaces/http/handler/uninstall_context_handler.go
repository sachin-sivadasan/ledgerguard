package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// UninstallContextHandler serves the "Uninstall Context" report (REPORTS.md — who
// uninstalled in a period, what state they were in right before, their plan and
// tenure). It correlates uninstall app-events to subscriptions via ShopifyShopGID.
// Mirrors the RetentionHandler structure (minus snapshots — this report is derived
// entirely from events + subscriptions).
type UninstallContextHandler struct {
	subRepo     repository.SubscriptionRepository
	eventRepo   repository.AppEventRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewUninstallContextHandler constructs an UninstallContextHandler.
func NewUninstallContextHandler(
	subRepo repository.SubscriptionRepository,
	eventRepo repository.AppEventRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *UninstallContextHandler {
	return &UninstallContextHandler{
		subRepo:     subRepo,
		eventRepo:   eventRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// uninstallStore is a single per-shop uninstall row in the report.
type uninstallStore struct {
	Domain               string  `json:"domain"`
	StateBeforeUninstall string  `json:"stateBeforeUninstall"`
	PlanName             string  `json:"planName"`
	TenureMonths         float64 `json:"tenureMonths"`
	UninstalledDate      string  `json:"uninstalledDate"`
}

// uninstallReport is the full JSON contract for the Uninstall Context report.
type uninstallReport struct {
	Uninstalls         int              `json:"uninstalls"`
	WereAtRiskPct      float64          `json:"wereAtRiskPct"`
	MedianTenureMonths float64          `json:"medianTenureMonths"`
	Stores             []uninstallStore `json:"stores"`
}

// GetUninstallContext returns the Uninstall Context report for an app.
// GET /api/v1/apps/{appID}/reports/uninstall-context?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *UninstallContextHandler) GetUninstallContext(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	now := time.Now().UTC()
	from, to := parseDateRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"), now)

	subs, err := h.subRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeUninstallRepoError(w, "FindByAppID", err)
		return
	}

	events, err := h.eventRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeUninstallRepoError(w, "FindByAppID(events)", err)
		return
	}

	subsByShop := indexSubsByShop(subs)
	report := buildUninstallReport(events, subsByShop, from, to)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeUninstallStoresCSV(w, report.Stores)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("uninstall_context: encode report: %v", err)
	}
}

// writeUninstallRepoError logs a repository failure and responds 503. These repos
// have no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeUninstallRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("uninstall_context: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// indexSubsByShop keys subscriptions by ShopifyShopGID (the join key with app
// events). Subscriptions with an empty GID are skipped — they cannot be correlated.
func indexSubsByShop(subs []*entity.Subscription) map[string]*entity.Subscription {
	byShop := make(map[string]*entity.Subscription, len(subs))
	for _, s := range subs {
		if s.ShopifyShopGID == "" {
			continue
		}
		byShop[s.ShopifyShopGID] = s
	}
	return byShop
}

// stateBeforeUninstall maps a correlated subscription to the human-readable state
// it was in right before uninstalling. A nil sub (no correlation) is "Unknown";
// a FROZEN sub is "Frozen"; otherwise the RiskState decides: SAFE → "Healthy",
// any missed/churned state → "At-Risk". A non-nil sub whose RiskState is empty or
// unrecognized also yields "Unknown".
func stateBeforeUninstall(sub *entity.Subscription) string {
	if sub == nil {
		return "Unknown"
	}
	if sub.Status == "FROZEN" {
		return "Frozen"
	}
	switch sub.RiskState {
	case valueobject.RiskStateSafe:
		return "Healthy"
	case valueobject.RiskStateOneCycleMissed, valueobject.RiskStateTwoCyclesMissed, valueobject.RiskStateChurned:
		return "At-Risk"
	default:
		return "Unknown"
	}
}

// wereAtRiskRate returns (at-risk-or-frozen) ÷ correlated clamped to [0,1], guarding
// divide-by-zero. The numerator counts shops whose inferred state was "At-Risk" OR
// "Frozen". The denominator is correlated-uninstalls-only (shops matched to a
// subscription, i.e. non-nil sub); uncorrelated shops are excluded from both. Note a
// correlated sub can still resolve to state "Unknown" (empty/unrecognized RiskState) —
// those remain in the denominator but not the numerator. Logs if it clamps (parity
// with retention's renewalRate diagnostics).
func wereAtRiskRate(atRisk, correlated int) float64 {
	if correlated <= 0 {
		return 0
	}
	rate := float64(atRisk) / float64(correlated)
	if rate < 0 || rate > 1 {
		log.Printf("uninstall_context: wereAtRiskPct %.4f outside [0,1] — clamping (unexpected counts atRisk=%d correlated=%d)", rate, atRisk, correlated)
	}
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}

// tenureMonths returns the sub's tenure at uninstall time in months (days ÷ 30),
// rounded to 1 decimal. A nil sub yields 0 (the CSV writer blanks a 0; JSON keeps the
// literal 0). Floored at 0 so a clock-skewed/backfilled uninstall predating CreatedAt
// can't produce a negative tenure that skews the median.
func tenureMonths(sub *entity.Subscription, uninstalledAt time.Time) float64 {
	if sub == nil {
		return 0
	}
	days := uninstalledAt.Sub(sub.CreatedAt).Hours() / 24
	if days < 0 {
		days = 0
	}
	return round1(days / 30.0)
}

// round1 rounds to 1 decimal place.
func round1(v float64) float64 {
	rounded, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', 1, 64), 64)
	return rounded
}

// buildUninstallReport filters uninstall events into the [from,to] window (whole
// `to` day inclusive, matching countReactivations' boundary), dedupes to the LATEST
// uninstall event per shop, correlates each to a subscription, and aggregates the
// KPIs. Stores are sorted by uninstalledDate descending (newest first).
func buildUninstallReport(events []*entity.AppEvent, subsByShop map[string]*entity.Subscription, from, to time.Time) uninstallReport {
	toExclusive := to.AddDate(0, 0, 1)

	// Dedup to the latest uninstall event per shop within range.
	latestByShop := map[string]*entity.AppEvent{}
	for _, e := range events {
		if !strings.Contains(strings.ToUpper(e.EventType), "UNINSTALL") {
			continue
		}
		if e.OccurredAt.Before(from) || !e.OccurredAt.Before(toExclusive) {
			continue
		}
		if e.ShopifyShopGID == "" {
			// Skip uncorrelatable events — otherwise every empty-GID event would
			// collapse into a single phantom row and under-count Uninstalls.
			continue
		}
		existing, ok := latestByShop[e.ShopifyShopGID]
		if !ok || e.OccurredAt.After(existing.OccurredAt) {
			latestByShop[e.ShopifyShopGID] = e
		}
	}

	stores := make([]uninstallStore, 0, len(latestByShop))
	var atRisk, correlated int
	tenures := make([]float64, 0, len(latestByShop))
	for shopGID, e := range latestByShop {
		sub := subsByShop[shopGID]
		state := stateBeforeUninstall(sub)

		domain := shopGID
		plan := ""
		tenure := 0.0
		if sub != nil {
			if sub.MyshopifyDomain != "" {
				domain = sub.MyshopifyDomain
			}
			plan = sub.PlanName
			tenure = tenureMonths(sub, e.OccurredAt)
			correlated++
			tenures = append(tenures, tenure)
			if state == "At-Risk" || state == "Frozen" {
				atRisk++
			}
		}

		stores = append(stores, uninstallStore{
			Domain:               domain,
			StateBeforeUninstall: state,
			PlanName:             plan,
			TenureMonths:         tenure,
			UninstalledDate:      e.OccurredAt.Format(dateLayout),
		})
	}

	sort.SliceStable(stores, func(i, j int) bool {
		return stores[i].UninstalledDate > stores[j].UninstalledDate
	})

	return uninstallReport{
		Uninstalls:         len(stores),
		WereAtRiskPct:      wereAtRiskRate(atRisk, correlated),
		MedianTenureMonths: medianTenure(tenures),
		Stores:             stores,
	}
}

// medianTenure returns the median of the tenure values (avg of the two middles for
// an even count), rounded to 1 decimal. 0 when there are no correlated tenures.
func medianTenure(tenures []float64) float64 {
	n := len(tenures)
	if n == 0 {
		return 0
	}
	sorted := make([]float64, n)
	copy(sorted, tenures)
	sort.Float64s(sorted)
	mid := n / 2
	if n%2 == 1 {
		return round1(sorted[mid])
	}
	return round1((sorted[mid-1] + sorted[mid]) / 2)
}

// writeUninstallStoresCSV writes the per-shop uninstall table as a CSV attachment.
// Uses encoding/csv so free-text domains/plan names with commas/quotes stay one
// column. An empty tenure (0, no correlated sub) is rendered as a blank cell.
func writeUninstallStoresCSV(w http.ResponseWriter, stores []uninstallStore) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="uninstall-context.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"domain", "stateBeforeUninstall", "planName", "tenureMonths", "uninstalledDate"})
	for _, s := range stores {
		tenure := ""
		if s.TenureMonths != 0 {
			tenure = strconv.FormatFloat(s.TenureMonths, 'f', 1, 64)
		}
		_ = cw.Write([]string{
			s.Domain,
			s.StateBeforeUninstall,
			s.PlanName,
			tenure,
			s.UninstalledDate,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("uninstall_context: write CSV: %v", err)
	}
}

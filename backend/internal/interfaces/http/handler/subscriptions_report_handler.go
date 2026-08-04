package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"math"
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

// SubscriptionsReportHandler serves the "Subscriptions (ARPU / LTV)" report
// (REPORTS.md — Archetype B, Composition): the active subscription base with
// ARPU (average revenue per user) and LTV (lifetime value) headlines and a
// per-plan composition/breakdown. Subscriptions are inherently RECURRING, so
// their MRR never mixes USAGE per the Revenue Classification rule. Mirrors the
// MRRReportHandler structure (subs + snapshot for the churn denominator).
type SubscriptionsReportHandler struct {
	subRepo      repository.SubscriptionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
}

// NewSubscriptionsReportHandler constructs a SubscriptionsReportHandler.
func NewSubscriptionsReportHandler(
	subRepo repository.SubscriptionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *SubscriptionsReportHandler {
	return &SubscriptionsReportHandler{
		subRepo:      subRepo,
		snapshotRepo: snapshotRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
	}
}

// subscriptionsPlan is a single per-plan row in the composition/breakdown.
type subscriptionsPlan struct {
	PlanName   string  `json:"planName"`
	ActiveSubs int     `json:"activeSubs"`
	MrrCents   int64   `json:"mrrCents"`
	ArpuCents  int64   `json:"arpuCents"`
	LtvCents   int64   `json:"ltvCents"`
	PctOfSubs  float64 `json:"pctOfSubs"`
}

// subscriptionsReport is the full JSON contract for the Subscriptions report.
type subscriptionsReport struct {
	Currency       string              `json:"currency"`
	ActiveSubs     int                 `json:"activeSubs"`
	ActiveMrrCents int64               `json:"activeMrrCents"`
	ArpuCents      int64               `json:"arpuCents"`
	LtvCents       int64               `json:"ltvCents"`
	ChurnRate      float64             `json:"churnRate"`
	Plans          []subscriptionsPlan `json:"plans"`
}

// GetSubscriptions returns the Subscriptions (ARPU / LTV) report for an app.
// GET /api/v1/apps/{appID}/reports/subscriptions?format=csv
func (h *SubscriptionsReportHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
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

	subs, err := h.subRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeSubscriptionsRepoError(w, "FindByAppID", err)
		return
	}

	// The churn rate for LTV uses the latest snapshot's total-subscription count as its
	// denominator via the shared churnRate helper — the same churn *definition* as the
	// Churn report. Both pick the newest snapshot ending at `now`, but the windows
	// differ (here 90d, Churn's default 30d), so the two can diverge if the latest
	// snapshot is older than 30 days.
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -90)
	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, now)
	if err != nil {
		writeSubscriptionsRepoError(w, "FindByAppIDRange", err)
		return
	}
	latest := latestSnapshot(snapshots)
	if latest == nil {
		// No snapshot in the trailing window → the churn denominator is unavailable, so
		// LTV will be undefined (rendered "—"). Log so this data-freshness gap stays
		// distinguishable from a genuine zero churn rate.
		log.Printf("subscriptions: no snapshot in 90d window for app %s — churn denominator unavailable, LTV undefined", app.ID)
	}

	report := buildSubscriptionsReport(subs, latest, newPlanLabeler(nil))

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeSubscriptionsPlansCSV(w, report.Plans)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("subscriptions: encode report: %v", err)
	}
}

// writeSubscriptionsRepoError logs a repository failure and responds 503. Neither the
// subscription nor the snapshot repo has a not-found sentinel — every error is an
// infrastructure failure (ADR-042).
func writeSubscriptionsRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("subscriptions: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// buildSubscriptionsReport computes the active base, ARPU, LTV and per-plan composition.
//   - Active subscriptions = SAFE subs (those contributing to MRR), matching how the MRR
//     report counts "active" so ARPU = ActiveMRR ÷ activeSubs stays coherent across reports.
//   - ARPU (monthly) = ActiveMRR ÷ activeSubs, integer cents (floored); 0 when no active subs.
//   - LTV = ARPU ÷ monthly churn rate; 0 when the churn rate is 0 (LTV mathematically
//     undefined — the frontend renders this as "—", not $0).
//   - Per-plan ARPU = plan MRR ÷ plan active subs; per-plan LTV uses the SAME app-level
//     churn rate (we don't have reliable per-plan churn), documented as an approximation.
//
// Plans are sorted by active-sub count descending (the composition axis).
func buildSubscriptionsReport(subs []*entity.Subscription, latest *entity.DailyMetricsSnapshot, labeler planLabeler) subscriptionsReport {
	type agg struct {
		activeSubs int
		mrr        int64
	}
	byPlan := map[string]*agg{}
	order := make([]string, 0)

	currency := "USD"
	var activeSubs int
	var activeMrrCents int64
	var churnedCount int

	for _, s := range subs {
		if currency == "USD" && s.Currency != "" {
			currency = s.Currency
		}
		if s.RiskState == valueobject.RiskStateChurned {
			churnedCount++
		}

		plan := labeler.label(s)
		a, ok := byPlan[plan]
		if !ok {
			a = &agg{}
			byPlan[plan] = a
			order = append(order, plan)
		}
		if s.RiskState == valueobject.RiskStateSafe {
			a.activeSubs++
			a.mrr += s.MRRCents()
			activeSubs++
			activeMrrCents += s.MRRCents()
		}
	}

	// Monthly churn rate via the shared helper — same definition as the Churn report.
	total := 0
	if latest != nil {
		total = latest.TotalSubscriptions
	}
	// Mirror the Churn report's drift log: when the live churned count exceeds the
	// snapshot total the snapshot is stale/behind, and churnRate silently clamps to 1.0
	// (collapsing LTV to ARPU). Log so the clamp never hides the drift.
	if total > 0 && churnedCount > total {
		log.Printf("subscriptions: live churned count %d exceeds latest snapshot total %d — clamping churn to 1.0 (stale snapshot?)", churnedCount, total)
	}
	rate := churnRate(churnedCount, total)

	plans := make([]subscriptionsPlan, 0, len(byPlan))
	for _, name := range order {
		a := byPlan[name]
		arpu := arpuCents(a.mrr, a.activeSubs)
		plans = append(plans, subscriptionsPlan{
			PlanName:   name,
			ActiveSubs: a.activeSubs,
			MrrCents:   a.mrr,
			ArpuCents:  arpu,
			LtvCents:   ltvCents(arpu, rate),
			PctOfSubs:  subsShare(a.activeSubs, activeSubs),
		})
	}
	sort.SliceStable(plans, func(i, j int) bool {
		return plans[i].ActiveSubs > plans[j].ActiveSubs
	})

	arpu := arpuCents(activeMrrCents, activeSubs)
	return subscriptionsReport{
		Currency:       currency,
		ActiveSubs:     activeSubs,
		ActiveMrrCents: activeMrrCents,
		ArpuCents:      arpu,
		LtvCents:       ltvCents(arpu, rate),
		ChurnRate:      rate,
		Plans:          plans,
	}
}

// arpuCents returns monthly MRR ÷ active subs in whole cents (floored via integer
// division), or 0 when there are no active subs (divide-by-zero guard).
func arpuCents(mrrCents int64, activeSubs int) int64 {
	if activeSubs <= 0 {
		return 0
	}
	return mrrCents / int64(activeSubs)
}

// ltvCents returns ARPU ÷ monthly churn rate rounded to the nearest whole cent, or 0
// when the churn rate is 0 (LTV is mathematically undefined — an infinite lifetime — so
// we surface 0 and let the frontend show "—" rather than a misleading $0 or ∞). Rounds
// (not truncates) because ARPU ÷ churn rarely lands on a whole cent (e.g. 1000 ÷ 0.7 =
// 1428.57…); a plain int64() cast would truncate, systematically understating LTV.
func ltvCents(arpuCents int64, churnRate float64) int64 {
	if churnRate <= 0 {
		return 0
	}
	return int64(math.Round(float64(arpuCents) / churnRate))
}

// subsShare returns a plan's share of total active subs, clamped to [0,1], guarding
// divide-by-zero.
func subsShare(planActiveSubs, totalActiveSubs int) float64 {
	if totalActiveSubs <= 0 {
		return 0
	}
	return float64(planActiveSubs) / float64(totalActiveSubs)
}

// writeSubscriptionsPlansCSV writes the per-plan breakdown as a CSV attachment. Uses
// encoding/csv so free-text plan names with commas/quotes stay one column.
func writeSubscriptionsPlansCSV(w http.ResponseWriter, plans []subscriptionsPlan) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="subscriptions.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"plan", "activeSubs", "mrrCents", "arpuCents", "ltvCents", "pctOfSubs"})
	for _, p := range plans {
		_ = cw.Write([]string{
			p.PlanName,
			strconv.Itoa(p.ActiveSubs),
			strconv.FormatInt(p.MrrCents, 10),
			strconv.FormatInt(p.ArpuCents, 10),
			strconv.FormatInt(p.LtvCents, 10),
			strconv.FormatFloat(p.PctOfSubs, 'f', 4, 64),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("subscriptions: write CSV: %v", err)
	}
}

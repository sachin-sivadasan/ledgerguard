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

// RetentionHandler serves the "Retention/Renewal" report (REPORTS.md — renewal
// rate, retained MRR, reactivations, per-plan renewal breakdown). Mirrors the
// ChurnHandler structure.
type RetentionHandler struct {
	subRepo      repository.SubscriptionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	eventRepo    repository.AppEventRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
}

// NewRetentionHandler constructs a RetentionHandler.
func NewRetentionHandler(
	subRepo repository.SubscriptionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	eventRepo repository.AppEventRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *RetentionHandler {
	return &RetentionHandler{
		subRepo:      subRepo,
		snapshotRepo: snapshotRepo,
		eventRepo:    eventRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
	}
}

// retentionPlan is a single per-plan renewal row in the report.
type retentionPlan struct {
	PlanName         string  `json:"planName"`
	ActiveSubs       int     `json:"activeSubs"`
	RenewalRate      float64 `json:"renewalRate"`
	RetainedMrrCents int64   `json:"retainedMrrCents"`
}

// retentionTrendPoint is a single point in the renewal-rate time series.
type retentionTrendPoint struct {
	Date        string  `json:"date"`
	RenewalRate float64 `json:"renewalRate"`
}

// retentionReport is the full JSON contract for the Retention report.
type retentionReport struct {
	Currency         string                `json:"currency"`
	RenewalRate      float64               `json:"renewalRate"`
	RetainedMrrCents int64                 `json:"retainedMrrCents"`
	Reactivations    int                   `json:"reactivations"`
	Trend            []retentionTrendPoint `json:"trend"`
	Plans            []retentionPlan       `json:"plans"`
}

// GetRetention returns the Retention/Renewal report for an app.
// GET /api/v1/apps/{appID}/reports/retention?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *RetentionHandler) GetRetention(w http.ResponseWriter, r *http.Request) {
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
		writeRetentionRepoError(w, "FindByAppID", err)
		return
	}

	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, to)
	if err != nil {
		writeRetentionRepoError(w, "FindByAppIDRange", err)
		return
	}

	events, err := h.eventRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeRetentionRepoError(w, "FindByAppID(events)", err)
		return
	}

	plans := buildRetentionPlans(subs)
	report := buildRetentionReport(subs, plans, latestSnapshot(snapshots))
	report.Reactivations = countReactivations(events, from, to)
	report.Trend = buildRetentionTrend(snapshots)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeRetentionPlansCSV(w, plans)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("retention: encode report: %v", err)
	}
}

// writeRetentionRepoError logs a repository failure and responds 503. These repos
// have no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeRetentionRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("retention: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// renewalRate clamps an already-computed rate to [0,1] so the UI never shows a
// value outside that band. It performs no division — callers (the snapshot
// headline, per-plan planRenewalRate, and each trend point) supply the rate.
func renewalRate(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}

// planRenewalRate returns safeCount ÷ total clamped to [0,1], guarding divide-by-zero.
func planRenewalRate(safeCount, total int) float64 {
	if total <= 0 {
		return 0
	}
	return renewalRate(float64(safeCount) / float64(total))
}

// buildRetentionReport aggregates renewal rate, retained MRR and currency. The
// headline renewalRate is the latest snapshot's RenewalSuccessRate (clamped),
// making it equal the last trend point; 0 when there is no snapshot in range.
func buildRetentionReport(subs []*entity.Subscription, plans []retentionPlan, latest *entity.DailyMetricsSnapshot) retentionReport {
	var retainedMrrCents int64
	currency := "USD"
	for _, s := range subs {
		if currency == "USD" && s.Currency != "" {
			currency = s.Currency
		}
		if s.RiskState == valueobject.RiskStateSafe {
			retainedMrrCents += s.MRRCents()
		}
	}

	rate := 0.0
	if latest != nil {
		// Log if the stored rate is outside [0,1] before renewalRate clamps it, so a
		// corrupt/stale snapshot value stays diagnosable rather than silently capped
		// (parity with the churn drift log).
		if latest.RenewalSuccessRate < 0 || latest.RenewalSuccessRate > 1 {
			log.Printf("retention: snapshot RenewalSuccessRate %.4f outside [0,1] — clamping (stale/corrupt snapshot?)", latest.RenewalSuccessRate)
		}
		rate = renewalRate(latest.RenewalSuccessRate)
	}

	return retentionReport{
		Currency:         currency,
		RenewalRate:      rate,
		RetainedMrrCents: retainedMrrCents,
		Reactivations:    0,
		Trend:            []retentionTrendPoint{},
		Plans:            plans,
	}
}

// buildRetentionPlans groups ALL subscriptions by plan name and computes, per plan,
// the SAFE (active) count, renewal rate (safe ÷ total in plan) and retained MRR
// (Σ MRR of SAFE subs). Result is sorted by retained MRR descending. An empty plan
// name is kept as its own bucket (not special-cased), mirroring churn.
func buildRetentionPlans(subs []*entity.Subscription) []retentionPlan {
	type agg struct {
		total       int
		safeCount   int
		retainedMrr int64
	}
	byPlan := map[string]*agg{}
	order := make([]string, 0)
	for _, s := range subs {
		a, ok := byPlan[s.PlanName]
		if !ok {
			a = &agg{}
			byPlan[s.PlanName] = a
			order = append(order, s.PlanName)
		}
		a.total++
		if s.RiskState == valueobject.RiskStateSafe {
			a.safeCount++
			a.retainedMrr += s.MRRCents()
		}
	}

	plans := make([]retentionPlan, 0, len(byPlan))
	for _, name := range order {
		a := byPlan[name]
		plans = append(plans, retentionPlan{
			PlanName:         name,
			ActiveSubs:       a.safeCount,
			RenewalRate:      planRenewalRate(a.safeCount, a.total),
			RetainedMrrCents: a.retainedMrr,
		})
	}
	sort.SliceStable(plans, func(i, j int) bool {
		return plans[i].RetainedMrrCents > plans[j].RetainedMrrCents
	})
	return plans
}

// countReactivations counts distinct shops (ShopifyShopGID) that had a reactivation
// event within the [from,to] date range, inclusive of the entire `to` day. Event
// types are matched case-insensitively on the "REACTIVAT" stem (e.g. REACTIVATED,
// reactivation). This is period-scoped — reactivations are inherently a within-window
// count. The upper bound is the day AFTER `to` (exclusive): when `to` is an explicit
// date param, parseDateRange returns it as midnight UTC, so +1 day keeps the whole
// `to` day inclusive (a plain `.After(to)` would drop every event on that day); when
// `to` defaults to `now`, the window is inclusive of the current instant (up to ~24h
// past it), which is acceptable for this coarse in-range count.
func countReactivations(events []*entity.AppEvent, from, to time.Time) int {
	toExclusive := to.AddDate(0, 0, 1)
	seen := map[string]struct{}{}
	for _, e := range events {
		if !strings.Contains(strings.ToUpper(e.EventType), "REACTIVAT") {
			continue
		}
		if e.OccurredAt.Before(from) || !e.OccurredAt.Before(toExclusive) {
			continue
		}
		seen[e.ShopifyShopGID] = struct{}{}
	}
	return len(seen)
}

// buildRetentionTrend converts daily snapshots to the renewal-rate time series.
func buildRetentionTrend(snapshots []*entity.DailyMetricsSnapshot) []retentionTrendPoint {
	trend := make([]retentionTrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		trend = append(trend, retentionTrendPoint{
			Date:        snap.Date.Format(dateLayout),
			RenewalRate: renewalRate(snap.RenewalSuccessRate),
		})
	}
	return trend
}

// writeRetentionPlansCSV writes the per-plan renewal table as a CSV attachment.
// Uses encoding/csv so free-text plan names with commas/quotes stay one column.
func writeRetentionPlansCSV(w http.ResponseWriter, plans []retentionPlan) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="retention.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"plan", "activeSubs", "renewalRate", "retainedMrrCents"})
	for _, p := range plans {
		_ = cw.Write([]string{
			p.PlanName,
			strconv.Itoa(p.ActiveSubs),
			strconv.FormatFloat(p.RenewalRate, 'f', 4, 64),
			strconv.FormatInt(p.RetainedMrrCents, 10),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("retention: write CSV: %v", err)
	}
}

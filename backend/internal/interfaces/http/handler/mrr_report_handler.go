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

// MRRReportHandler serves the "MRR" report (REPORTS.md — monthly recurring
// revenue headline, month-over-month growth, new/churned MRR movement, MRR
// trend, per-plan MRR breakdown). Mirrors the RetentionHandler structure.
type MRRReportHandler struct {
	subRepo      repository.SubscriptionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
}

// NewMRRReportHandler constructs an MRRReportHandler.
func NewMRRReportHandler(
	subRepo repository.SubscriptionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *MRRReportHandler {
	return &MRRReportHandler{
		subRepo:      subRepo,
		snapshotRepo: snapshotRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
	}
}

// mrrPlan is a single per-plan MRR row in the report.
type mrrPlan struct {
	PlanName   string  `json:"planName"`
	ActiveSubs int     `json:"activeSubs"`
	MrrCents   int64   `json:"mrrCents"`
	PctOfTotal float64 `json:"pctOfTotal"`
}

// mrrTrendPoint is a single point in the MRR time series.
type mrrTrendPoint struct {
	Date     string `json:"date"`
	MrrCents int64  `json:"mrrCents"`
}

// mrrReport is the full JSON contract for the MRR report.
type mrrReport struct {
	Currency        string          `json:"currency"`
	MrrCents        int64           `json:"mrrCents"`
	MomChangePct    float64         `json:"momChangePct"`
	NewMrrCents     int64           `json:"newMrrCents"`
	ChurnedMrrCents int64           `json:"churnedMrrCents"`
	Trend           []mrrTrendPoint `json:"trend"`
	Plans           []mrrPlan       `json:"plans"`
}

// GetMRRReport returns the MRR report for an app.
// GET /api/v1/apps/{appID}/reports/mrr?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *MRRReportHandler) GetMRRReport(w http.ResponseWriter, r *http.Request) {
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
		writeMRRRepoError(w, "FindByAppID", err)
		return
	}

	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, to)
	if err != nil {
		writeMRRRepoError(w, "FindByAppIDRange", err)
		return
	}

	plans := buildMRRPlans(subs)
	report := buildMRRReport(subs, plans, snapshots, from, to)
	report.Trend = buildMRRTrend(snapshots)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeMRRPlansCSV(w, plans)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("mrr: encode report: %v", err)
	}
}

// writeMRRRepoError logs a repository failure and responds 503. These repos have
// no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeMRRRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("mrr: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// momChangePct returns the signed growth ratio of MRR: (latest − baseline) / baseline,
// where baseline is the FIRST in-range snapshot's ActiveMRRCents and latest is the last.
// The "mom" name reflects the default 30-day range (start-vs-latest ≈ month-over-month);
// with a wider selected range this is range-start-to-latest, not a trailing calendar
// month. This is a growth ratio, NOT a rate — intentionally not clamped to [0,1] (a
// doubling → 1.0, a halving → -0.5). Returns 0 when there are fewer than 2 snapshots or
// the baseline is <= 0 (the frontend hides the delta when there are <2 snapshots).
func momChangePct(snapshots []*entity.DailyMetricsSnapshot) float64 {
	if len(snapshots) < 2 {
		return 0
	}
	baseline := snapshots[0].ActiveMRRCents
	if baseline <= 0 {
		return 0
	}
	latest := snapshots[len(snapshots)-1].ActiveMRRCents
	return float64(latest-baseline) / float64(baseline)
}

// buildMRRReport aggregates the MRR headline, MoM growth, new/churned MRR movement
// and currency. The headline mrrCents is the latest in-range snapshot's
// ActiveMRRCents (0 when no snapshot in range).
func buildMRRReport(subs []*entity.Subscription, plans []mrrPlan, snapshots []*entity.DailyMetricsSnapshot, from, to time.Time) mrrReport {
	currency := "USD"
	for _, s := range subs {
		if s.Currency != "" {
			currency = s.Currency
			break
		}
	}

	var mrrCents int64
	if latest := latestSnapshot(snapshots); latest != nil {
		mrrCents = latest.ActiveMRRCents
	}

	// New/churned MRR are period-scoped movements over the [from,to] window,
	// inclusive of the entire `to` day (matching retention's boundary). "New" counts
	// only subs that are CURRENTLY SAFE and STARTED in-range (StartDate() = the real
	// business start, NOT the record-created CreatedAt which resets on every rebuild) —
	// a sub started in-range that has since gone at-risk/churned is excluded from New
	// (and, if it churned in-range, shows under Churned instead).
	toExclusive := to.AddDate(0, 0, 1)
	var newMrrCents, churnedMrrCents int64
	for _, s := range subs {
		start := s.StartDate()
		if s.RiskState == valueobject.RiskStateSafe &&
			!start.Before(from) && start.Before(toExclusive) {
			newMrrCents += s.MRRCents()
		}
		if s.RiskState == valueobject.RiskStateChurned {
			churned := churnedDateOf(s)
			if churned != nil && !churned.Before(from) && churned.Before(toExclusive) {
				churnedMrrCents += s.MRRCents()
			}
		}
	}

	return mrrReport{
		Currency:        currency,
		MrrCents:        mrrCents,
		MomChangePct:    momChangePct(snapshots),
		NewMrrCents:     newMrrCents,
		ChurnedMrrCents: churnedMrrCents,
		Trend:           []mrrTrendPoint{},
		Plans:           plans,
	}
}

// buildMRRPlans groups ALL subscriptions by plan name and computes, per plan, the
// SAFE (active) count, MRR (Σ MRR of SAFE subs) and its share of total plan MRR.
// Result is sorted by MRR descending. An empty plan name is kept as its own bucket
// (not special-cased), mirroring retention.
func buildMRRPlans(subs []*entity.Subscription) []mrrPlan {
	type agg struct {
		safeCount int
		mrr       int64
	}
	byPlan := map[string]*agg{}
	order := make([]string, 0)
	var totalMrr int64
	for _, s := range subs {
		a, ok := byPlan[s.PlanName]
		if !ok {
			a = &agg{}
			byPlan[s.PlanName] = a
			order = append(order, s.PlanName)
		}
		if s.RiskState == valueobject.RiskStateSafe {
			a.safeCount++
			a.mrr += s.MRRCents()
			totalMrr += s.MRRCents()
		}
	}

	plans := make([]mrrPlan, 0, len(byPlan))
	for _, name := range order {
		a := byPlan[name]
		pct := 0.0
		if totalMrr > 0 {
			pct = float64(a.mrr) / float64(totalMrr)
		}
		plans = append(plans, mrrPlan{
			PlanName:   name,
			ActiveSubs: a.safeCount,
			MrrCents:   a.mrr,
			PctOfTotal: pct,
		})
	}
	sort.SliceStable(plans, func(i, j int) bool {
		return plans[i].MrrCents > plans[j].MrrCents
	})
	return plans
}

// buildMRRTrend converts daily snapshots to the MRR time series (ascending).
func buildMRRTrend(snapshots []*entity.DailyMetricsSnapshot) []mrrTrendPoint {
	trend := make([]mrrTrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		trend = append(trend, mrrTrendPoint{
			Date:     snap.Date.Format(dateLayout),
			MrrCents: snap.ActiveMRRCents,
		})
	}
	return trend
}

// writeMRRPlansCSV writes the per-plan MRR table as a CSV attachment. Uses
// encoding/csv so free-text plan names with commas/quotes stay one column.
func writeMRRPlansCSV(w http.ResponseWriter, plans []mrrPlan) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="mrr.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"plan", "activeSubs", "mrrCents", "pctOfTotal"})
	for _, p := range plans {
		_ = cw.Write([]string{
			p.PlanName,
			strconv.Itoa(p.ActiveSubs),
			strconv.FormatInt(p.MrrCents, 10),
			strconv.FormatFloat(p.PctOfTotal, 'f', 4, 64),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("mrr: write CSV: %v", err)
	}
}

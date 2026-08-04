package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
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

// ActiveCustomersReportHandler serves the "Active Customers" report (REPORTS.md,
// Customers · Archetype A). Headline = current active (non-churned) paying subs;
// New / Net-Change movement over the window; an active-customers trend from daily
// snapshots; and an "active by plan" breakdown. Mirrors MRRReportHandler.
//
// "Active" = non-churned paying subscriptions (SAFE + at-risk), i.e. every risk
// state except CHURNED — matching the wireframe's "Safe + at-risk".
type ActiveCustomersReportHandler struct {
	subRepo       repository.SubscriptionRepository
	snapshotRepo  repository.DailyMetricsSnapshotRepository
	appRepo       repository.AppRepository
	partnerRepo   repository.PartnerAccountRepository
	planLabelRepo repository.PlanLabelRepository
}

// NewActiveCustomersReportHandler constructs an ActiveCustomersReportHandler. planLabelRepo
// may be nil (plan labels then fall back to pseudo-labels).
func NewActiveCustomersReportHandler(
	subRepo repository.SubscriptionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	planLabelRepo repository.PlanLabelRepository,
) *ActiveCustomersReportHandler {
	return &ActiveCustomersReportHandler{
		subRepo:       subRepo,
		snapshotRepo:  snapshotRepo,
		appRepo:       appRepo,
		partnerRepo:   partnerRepo,
		planLabelRepo: planLabelRepo,
	}
}

// activeCustomersPlan is a single per-plan row (active count + MRR + share of active).
type activeCustomersPlan struct {
	PlanName    string  `json:"planName"`
	ActiveSubs  int     `json:"activeSubs"`
	MrrCents    int64   `json:"mrrCents"`
	PctOfActive float64 `json:"pctOfActive"`
}

// activeCustomersTrendPoint is a single point in the active-customers time series.
type activeCustomersTrendPoint struct {
	Date            string `json:"date"`
	ActiveCustomers int    `json:"activeCustomers"`
}

// activeCustomersReport is the full JSON contract for the Active Customers report.
type activeCustomersReport struct {
	Currency        string `json:"currency"`
	ActiveCustomers int    `json:"activeCustomers"`
	NewCount        int    `json:"newCount"`
	ChurnedCount    int    `json:"churnedCount"`
	NetChange       int    `json:"netChange"`
	// Interval is the trend granularity: day / week / month.
	Interval string                      `json:"interval"`
	Trend    []activeCustomersTrendPoint `json:"trend"`
	Plans    []activeCustomersPlan       `json:"plans"`
}

// GetActiveCustomersReport returns the Active Customers report for an app.
// GET /api/v1/apps/{appID}/reports/active-customers?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *ActiveCustomersReportHandler) GetActiveCustomersReport(w http.ResponseWriter, r *http.Request) {
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
		writeActiveCustomersRepoError(w, "FindByAppID", err)
		return
	}

	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, to)
	if err != nil {
		writeActiveCustomersRepoError(w, "FindByAppIDRange", err)
		return
	}

	interval := resolveTrendInterval(from, to)
	labeler := newPlanLabeler(planLabelMapFor(r.Context(), h.planLabelRepo, app.ID))
	plans := buildActiveCustomersPlans(subs, labeler)
	report := buildActiveCustomersReport(subs, plans, from, to)
	report.Interval = string(interval)
	report.Trend = buildActiveCustomersTrend(downsampleSnapshots(snapshots, interval))

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeActiveCustomersPlansCSV(w, plans)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("active-customers: encode report: %v", err)
	}
}

// writeActiveCustomersRepoError logs a repository failure and responds 503. These
// repos have no not-found sentinel — every error is an infra failure (ADR-042).
func writeActiveCustomersRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("active-customers: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// snapshotActiveCount is the active (non-churned) customer count for a snapshot:
// TotalSubscriptions − CHURNED (= SAFE + at-risk).
func snapshotActiveCount(s *entity.DailyMetricsSnapshot) int {
	return s.TotalSubscriptions - s.ChurnedCount
}

// buildActiveCustomersReport aggregates the headline active count, in-window New /
// Churned / Net-Change movement, and currency.
//
// The headline activeCustomers is the current live non-churned subscription count,
// taken as the sum of the per-plan breakdown so the big number ALWAYS reconciles with
// the "Active by plan" table (and with New/Churned/Net, which are also live). The
// snapshot series feeds the trend line only — mixing snapshot (headline) with live
// (table) sources previously let the two disagree whenever the latest snapshot's active
// tally drifted from the current subscription state. New counts subs whose StartDate()
// (the real business start = activated_at, NOT the record-created CreatedAt that resets
// on rebuild) falls in [from, to]; Churned counts subs that churned in-range. A sub that
// started AND churned in-window counts toward both (net nets to zero).
func buildActiveCustomersReport(subs []*entity.Subscription, plans []activeCustomersPlan, from, to time.Time) activeCustomersReport {
	currency := "USD"
	for _, s := range subs {
		if s.Currency != "" {
			currency = s.Currency
			break
		}
	}

	active := 0
	for _, p := range plans {
		active += p.ActiveSubs
	}

	toExclusive := to.AddDate(0, 0, 1)
	newCount, churnedCount := 0, 0
	for _, s := range subs {
		start := s.StartDate()
		if !start.Before(from) && start.Before(toExclusive) {
			newCount++
		}
		if s.RiskState == valueobject.RiskStateChurned {
			if churned := churnedDateOf(s); churned != nil && !churned.Before(from) && churned.Before(toExclusive) {
				churnedCount++
			}
		}
	}

	return activeCustomersReport{
		Currency:        currency,
		ActiveCustomers: active,
		NewCount:        newCount,
		ChurnedCount:    churnedCount,
		NetChange:       newCount - churnedCount,
		Trend:           []activeCustomersTrendPoint{},
		Plans:           plans,
	}
}

// buildActiveCustomersPlans groups the ACTIVE (non-churned) subscriptions by plan and
// computes, per plan, the active count, MRR (Σ MRR of active subs), and the plan's
// share of the total active COUNT (matching the wireframe's "% OF ACTIVE"). Result is
// sorted by MRR descending. Plans with only churned subs don't appear.
func buildActiveCustomersPlans(subs []*entity.Subscription, labeler planLabeler) []activeCustomersPlan {
	type agg struct {
		count int
		mrr   int64
	}
	byPlan := map[string]*agg{}
	order := make([]string, 0)
	totalActive := 0
	for _, s := range subs {
		if s.RiskState == valueobject.RiskStateChurned {
			continue
		}
		plan := labeler.label(s)
		a, ok := byPlan[plan]
		if !ok {
			a = &agg{}
			byPlan[plan] = a
			order = append(order, plan)
		}
		a.count++
		a.mrr += s.MRRCents()
		totalActive++
	}

	plans := make([]activeCustomersPlan, 0, len(byPlan))
	for _, name := range order {
		a := byPlan[name]
		pct := 0.0
		if totalActive > 0 {
			pct = float64(a.count) / float64(totalActive)
		}
		plans = append(plans, activeCustomersPlan{
			PlanName:    name,
			ActiveSubs:  a.count,
			MrrCents:    a.mrr,
			PctOfActive: pct,
		})
	}
	sort.SliceStable(plans, func(i, j int) bool {
		return plans[i].MrrCents > plans[j].MrrCents
	})

	// Fold minor tiers (proration/refund noise) into a trailing "Other" row — only when
	// there are enough tiers for a real long tail.
	if len(plans) <= minTiersToCollapse {
		return plans
	}
	threshold := tierSignificanceThreshold(totalActive)
	kept := make([]activeCustomersPlan, 0, len(plans))
	other := activeCustomersPlan{PlanName: "Other"}
	otherTiers := 0
	for _, p := range plans {
		if p.ActiveSubs >= threshold {
			kept = append(kept, p)
			continue
		}
		other.ActiveSubs += p.ActiveSubs
		other.MrrCents += p.MrrCents
		other.PctOfActive += p.PctOfActive
		otherTiers++
	}
	if otherTiers > 0 {
		other.PlanName = fmt.Sprintf("Other (%d tiers)", otherTiers)
		kept = append(kept, other)
	}
	return kept
}

// buildActiveCustomersTrend converts daily snapshots to the active-customers series
// (ascending), one point per (downsampled) snapshot.
func buildActiveCustomersTrend(snapshots []*entity.DailyMetricsSnapshot) []activeCustomersTrendPoint {
	trend := make([]activeCustomersTrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		trend = append(trend, activeCustomersTrendPoint{
			Date:            snap.Date.Format(dateLayout),
			ActiveCustomers: snapshotActiveCount(snap),
		})
	}
	return trend
}

// writeActiveCustomersPlansCSV writes the per-plan active-customers table as a CSV
// attachment. Uses encoding/csv so free-text plan names stay one column.
func writeActiveCustomersPlansCSV(w http.ResponseWriter, plans []activeCustomersPlan) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="active-customers.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"plan", "activeSubs", "mrrCents", "pctOfActive"})
	for _, p := range plans {
		_ = cw.Write([]string{
			p.PlanName,
			strconv.Itoa(p.ActiveSubs),
			strconv.FormatInt(p.MrrCents, 10),
			strconv.FormatFloat(p.PctOfActive, 'f', 4, 64),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("active-customers: write CSV: %v", err)
	}
}

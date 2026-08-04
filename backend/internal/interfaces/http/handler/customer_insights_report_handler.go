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

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// CustomerInsightsReportHandler serves the "Customer Insights" report (REPORTS.md,
// Customers · Archetype B — segmentation). It slices the current customer base three
// cross-cutting ways the single-dimension reports don't: by revenue band (how MRR
// concentrates by customer size), by risk state, and a plan × risk crosstab (where the
// at-risk revenue sits) — plus the top customers by MRR.
//
// The "Shopify plan" cut named in the catalog is intentionally omitted: transactions.
// shop_plan is never populated (the Partner API's read-only transaction stream carries no
// merchant plan), so segmenting on it would be a dead, all-empty dimension.
//
// The customer base = non-churned subscriptions (SAFE + at-risk), matching Active
// Customers; churned subs appear only in the risk breakdown for context.
type CustomerInsightsReportHandler struct {
	subRepo       repository.SubscriptionRepository
	appRepo       repository.AppRepository
	partnerRepo   repository.PartnerAccountRepository
	planLabelRepo repository.PlanLabelRepository
}

// NewCustomerInsightsReportHandler constructs a CustomerInsightsReportHandler. planLabelRepo
// may be nil (plan labels then fall back to pseudo-labels).
func NewCustomerInsightsReportHandler(
	subRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	planLabelRepo repository.PlanLabelRepository,
) *CustomerInsightsReportHandler {
	return &CustomerInsightsReportHandler{subRepo: subRepo, appRepo: appRepo, partnerRepo: partnerRepo, planLabelRepo: planLabelRepo}
}

// revenueBand is one MRR bucket over the active base: how many customers sit in it and
// how much MRR they represent.
type revenueBand struct {
	Label          string  `json:"label"`
	Customers      int     `json:"customers"`
	MrrCents       int64   `json:"mrrCents"`
	PctOfCustomers float64 `json:"pctOfCustomers"`
}

// riskSegment is one risk bucket (SAFE / AT_RISK / CHURNED) across the whole base.
type riskSegment struct {
	RiskState string `json:"riskState"`
	Customers int    `json:"customers"`
	MrrCents  int64  `json:"mrrCents"`
}

// planRiskRow is one plan's slice of the active base, split by risk so the report shows
// WHERE the at-risk revenue concentrates (the crosstab the other reports lack).
type planRiskRow struct {
	PlanName       string `json:"planName"`
	Customers      int    `json:"customers"` // active (safe + at-risk)
	SafeCount      int    `json:"safeCount"`
	AtRiskCount    int    `json:"atRiskCount"`
	MrrCents       int64  `json:"mrrCents"`
	AtRiskMrrCents int64  `json:"atRiskMrrCents"`
}

// topCustomer is one of the highest-MRR active customers (the whales).
type topCustomer struct {
	ShopName  string `json:"shopName"`
	PlanName  string `json:"planName"`
	MrrCents  int64  `json:"mrrCents"`
	RiskState string `json:"riskState"`
}

// customerInsightsReport is the full JSON contract for the Customer Insights report.
type customerInsightsReport struct {
	Currency        string        `json:"currency"`
	TotalCustomers  int           `json:"totalCustomers"` // active base (non-churned)
	ActiveMrrCents  int64         `json:"activeMrrCents"`
	AtRiskCustomers int           `json:"atRiskCustomers"`
	AtRiskMrrCents  int64         `json:"atRiskMrrCents"`
	RevenueBands    []revenueBand `json:"revenueBands"`
	RiskSegments    []riskSegment `json:"riskSegments"`
	PlanRisk        []planRiskRow `json:"planRisk"`
	TopCustomers    []topCustomer `json:"topCustomers"`
}

// topCustomersLimit caps the whales list so a huge base doesn't bloat the payload.
const topCustomersLimit = 10

// revenueBandDefs are the fixed MRR buckets (in cents), low→high. A customer lands in the
// first band whose upper bound it's under; the last band is open-ended.
var revenueBandDefs = []struct {
	label string
	upper int64 // exclusive; 0 ⇒ open-ended (last band)
}{
	{"< $25", 2500},
	{"$25–$50", 5000},
	{"$50–$100", 10000},
	{"$100–$250", 25000},
	{"$250+", 0},
}

// GetCustomerInsights returns the Customer Insights report for an app.
// GET /api/v1/apps/{appID}/reports/customer-insights?format=csv
func (h *CustomerInsightsReportHandler) GetCustomerInsights(w http.ResponseWriter, r *http.Request) {
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
		// No not-found sentinel on this repo — every error is an infra failure (ADR-042).
		log.Printf("customer-insights: repo error in FindByAppID: %v", err)
		writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
		return
	}

	labeler := newPlanLabeler(planLabelMapFor(r.Context(), h.planLabelRepo, app.ID))
	report := buildCustomerInsights(subs, labeler)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeCustomerInsightsCSV(w, report)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("customer-insights: encode report: %v", err)
	}
}

// buildCustomerInsights segments the subscriptions into the report's four views. The
// active base (SAFE + at-risk) drives revenue bands, the plan crosstab, top customers, and
// the headline KPIs; churned subs contribute only the CHURNED risk segment.
func buildCustomerInsights(subs []*entity.Subscription, labeler planLabeler) customerInsightsReport {
	currency := "USD"
	for _, s := range subs {
		if s.Currency != "" {
			currency = s.Currency
			break
		}
	}

	report := customerInsightsReport{Currency: currency}

	// Risk segments across the WHOLE base (churned included).
	safeSeg := riskSegment{RiskState: "SAFE"}
	atRiskSeg := riskSegment{RiskState: "AT_RISK"}
	churnedSeg := riskSegment{RiskState: "CHURNED"}

	// Plan crosstab over the active base only.
	planAgg := map[string]*planRiskRow{}
	planOrder := make([]string, 0)
	bandCustomers := make([]int, len(revenueBandDefs))
	bandMrr := make([]int64, len(revenueBandDefs))
	actives := make([]*entity.Subscription, 0, len(subs))

	for _, s := range subs {
		mrr := s.MRRCents()
		switch {
		case s.RiskState.IsChurned():
			churnedSeg.Customers++
			churnedSeg.MrrCents += mrr
			continue // churned subs don't enter the active-base views
		case s.RiskState.IsAtRisk():
			atRiskSeg.Customers++
			atRiskSeg.MrrCents += mrr
		default: // SAFE
			safeSeg.Customers++
			safeSeg.MrrCents += mrr
		}

		// Active base: KPIs, plan crosstab, revenue bands, whales.
		report.TotalCustomers++
		report.ActiveMrrCents += mrr
		if s.RiskState.IsAtRisk() {
			report.AtRiskCustomers++
			report.AtRiskMrrCents += mrr
		}

		plan := labeler.label(s)
		row, ok := planAgg[plan]
		if !ok {
			row = &planRiskRow{PlanName: plan}
			planAgg[plan] = row
			planOrder = append(planOrder, plan)
		}
		row.Customers++
		row.MrrCents += mrr
		if s.RiskState.IsAtRisk() {
			row.AtRiskCount++
			row.AtRiskMrrCents += mrr
		} else {
			row.SafeCount++
		}

		bandCustomers[revenueBandIndex(mrr)]++
		bandMrr[revenueBandIndex(mrr)] += mrr
		actives = append(actives, s)
	}

	report.RiskSegments = []riskSegment{safeSeg, atRiskSeg, churnedSeg}
	report.RevenueBands = buildRevenueBands(bandCustomers, bandMrr, report.TotalCustomers)
	report.PlanRisk = buildPlanRisk(planAgg, planOrder, report.TotalCustomers)
	report.TopCustomers = buildTopCustomers(actives, labeler)
	return report
}

// revenueBandIndex returns the band a given MRR falls into (the first band whose exclusive
// upper bound it's under; the open-ended last band catches the rest).
func revenueBandIndex(mrr int64) int {
	for i, b := range revenueBandDefs {
		if b.upper == 0 || mrr < b.upper {
			return i
		}
	}
	return len(revenueBandDefs) - 1
}

// buildRevenueBands materializes the band rows with each band's share of the active count.
func buildRevenueBands(customers []int, mrr []int64, totalActive int) []revenueBand {
	bands := make([]revenueBand, len(revenueBandDefs))
	for i, def := range revenueBandDefs {
		pct := 0.0
		if totalActive > 0 {
			pct = float64(customers[i]) / float64(totalActive)
		}
		bands[i] = revenueBand{
			Label:          def.label,
			Customers:      customers[i],
			MrrCents:       mrr[i],
			PctOfCustomers: pct,
		}
	}
	return bands
}

// buildPlanRisk flattens the plan crosstab, sorted by MRR descending, then folds minor
// tiers (proration/refund noise — see tierSignificanceThreshold) into a single trailing
// "Other (N tiers)" row so the table shows the real plans, not dozens of one-off prices.
func buildPlanRisk(agg map[string]*planRiskRow, order []string, totalActive int) []planRiskRow {
	rows := make([]planRiskRow, 0, len(agg))
	for _, name := range order {
		rows = append(rows, *agg[name])
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].MrrCents != rows[j].MrrCents {
			return rows[i].MrrCents > rows[j].MrrCents
		}
		return rows[i].PlanName < rows[j].PlanName // stable tie-break across rebuilds
	})

	if len(rows) <= minTiersToCollapse {
		return rows
	}
	threshold := tierSignificanceThreshold(totalActive)
	significant := make([]planRiskRow, 0, len(rows))
	other := planRiskRow{PlanName: "Other"}
	otherTiers := 0
	for _, r := range rows {
		if r.Customers >= threshold {
			significant = append(significant, r)
			continue
		}
		other.Customers += r.Customers
		other.SafeCount += r.SafeCount
		other.AtRiskCount += r.AtRiskCount
		other.MrrCents += r.MrrCents
		other.AtRiskMrrCents += r.AtRiskMrrCents
		otherTiers++
	}
	if otherTiers > 0 {
		other.PlanName = fmt.Sprintf("Other (%d tiers)", otherTiers)
		significant = append(significant, other)
	}
	return significant
}

// buildTopCustomers returns the highest-MRR active customers, capped. Ties break by shop
// name so the ordering is deterministic across rebuilds.
func buildTopCustomers(actives []*entity.Subscription, labeler planLabeler) []topCustomer {
	sorted := make([]*entity.Subscription, len(actives))
	copy(sorted, actives)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].MRRCents() != sorted[j].MRRCents() {
			return sorted[i].MRRCents() > sorted[j].MRRCents()
		}
		return sorted[i].ShopName < sorted[j].ShopName
	})
	limit := topCustomersLimit
	if len(sorted) < limit {
		limit = len(sorted)
	}
	top := make([]topCustomer, 0, limit)
	for _, s := range sorted[:limit] {
		top = append(top, topCustomer{
			ShopName:  s.ShopName,
			PlanName:  labeler.label(s),
			MrrCents:  s.MRRCents(),
			RiskState: string(s.RiskState),
		})
	}
	return top
}

// writeCustomerInsightsCSV writes the plan × risk crosstab (the report's richest table) as
// a CSV attachment.
func writeCustomerInsightsCSV(w http.ResponseWriter, report customerInsightsReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="customer-insights.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"plan", "customers", "safe", "at_risk", "mrr_cents", "at_risk_mrr_cents"})
	for _, p := range report.PlanRisk {
		_ = cw.Write([]string{
			p.PlanName,
			strconv.Itoa(p.Customers),
			strconv.Itoa(p.SafeCount),
			strconv.Itoa(p.AtRiskCount),
			strconv.FormatInt(p.MrrCents, 10),
			strconv.FormatInt(p.AtRiskMrrCents, 10),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("customer-insights: write CSV: %v", err)
	}
}

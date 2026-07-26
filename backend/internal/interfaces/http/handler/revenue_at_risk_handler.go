package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// Recovery rates used to estimate recoverable revenue from at-risk subscriptions.
// A 1-cycle-late store recovers far more often than a 2-cycle-late one.
// These start as constants and can later be learned from reactivation events.
const (
	recoveryRateOneCycle = 0.60
	recoveryRateTwoCycle = 0.25
)

const dateLayout = "2006-01-02"

// RevenueAtRiskHandler serves the "Revenue at Risk" report.
type RevenueAtRiskHandler struct {
	subRepo      repository.SubscriptionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
}

// NewRevenueAtRiskHandler constructs a RevenueAtRiskHandler.
func NewRevenueAtRiskHandler(
	subRepo repository.SubscriptionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *RevenueAtRiskHandler {
	return &RevenueAtRiskHandler{
		subRepo:      subRepo,
		snapshotRepo: snapshotRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
	}
}

// revenueAtRiskStore is a single at-risk store row in the report.
type revenueAtRiskStore struct {
	Domain             string `json:"domain"`
	ShopName           string `json:"shopName"`
	MRRCents           int64  `json:"mrrCents"`
	RiskState          string `json:"riskState"`
	DaysLate           int    `json:"daysLate"`
	ExpectedChargeDate string `json:"expectedChargeDate"`
	PlanName           string `json:"planName"`
	RecoverableCents   int64  `json:"recoverableCents"`
}

// revenueAtRiskTrendPoint is a single point in the at-risk MRR time series.
type revenueAtRiskTrendPoint struct {
	Date       string `json:"date"`
	AtRiskCents int64  `json:"atRiskCents"`
}

// revenueAtRiskReport is the full JSON contract (REPORTS.md §3.4).
type revenueAtRiskReport struct {
	Currency         string                    `json:"currency"`
	TotalAtRiskCents int64                     `json:"totalAtRiskCents"`
	RecoverableCents int64                     `json:"recoverableCents"`
	ByState          revenueAtRiskByState      `json:"byState"`
	Counts           revenueAtRiskCounts       `json:"counts"`
	Trend            []revenueAtRiskTrendPoint `json:"trend"`
	Stores           []revenueAtRiskStore      `json:"stores"`
}

type revenueAtRiskByState struct {
	OneCycleCents int64 `json:"oneCycleCents"`
	TwoCycleCents int64 `json:"twoCycleCents"`
}

type revenueAtRiskCounts struct {
	OneCycle int `json:"oneCycle"`
	TwoCycle int `json:"twoCycle"`
}

// GetRevenueAtRisk returns the Revenue at Risk report for an app.
// GET /api/v1/apps/{appID}/reports/revenue-at-risk?from=YYYY-MM-DD&to=YYYY-MM-DD&segment=all|plan:<name>&format=csv
func (h *RevenueAtRiskHandler) GetRevenueAtRisk(w http.ResponseWriter, r *http.Request) {
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
	planFilter := parseSegment(r.URL.Query().Get("segment"))

	// Fetch at-risk subscriptions (1-cycle + 2-cycle missed).
	oneCycle, err := h.subRepo.FindByRiskState(r.Context(), app.ID, valueobject.RiskStateOneCycleMissed)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch at-risk subscriptions")
		return
	}
	twoCycle, err := h.subRepo.FindByRiskState(r.Context(), app.ID, valueobject.RiskStateTwoCyclesMissed)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch at-risk subscriptions")
		return
	}

	atRisk := make([]*entity.Subscription, 0, len(oneCycle)+len(twoCycle))
	atRisk = append(atRisk, oneCycle...)
	atRisk = append(atRisk, twoCycle...)
	if planFilter != "" {
		atRisk = filterByPlan(atRisk, planFilter)
	}

	stores := buildStores(atRisk, now)
	report := buildReport(atRisk, stores)

	// Trend from daily snapshots.
	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, to)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch snapshot data")
		return
	}
	report.Trend = buildTrend(snapshots)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeStoresCSV(w, stores)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// parseDateRange parses from/to query params (YYYY-MM-DD). Defaults: to=today, from=today-30d.
func parseDateRange(fromStr, toStr string, now time.Time) (time.Time, time.Time) {
	to := now
	if t, err := time.Parse(dateLayout, toStr); err == nil {
		to = t
	}
	from := to.AddDate(0, 0, -30)
	if f, err := time.Parse(dateLayout, fromStr); err == nil {
		from = f
	}
	return from, to
}

// parseSegment parses an optional "plan:<name>" segment, returning the plan name (or "").
func parseSegment(segment string) string {
	if segment == "" || segment == "all" {
		return ""
	}
	if strings.HasPrefix(segment, "plan:") {
		return strings.TrimPrefix(segment, "plan:")
	}
	return ""
}

// filterByPlan keeps only subscriptions on the given plan.
func filterByPlan(subs []*entity.Subscription, plan string) []*entity.Subscription {
	filtered := make([]*entity.Subscription, 0, len(subs))
	for _, s := range subs {
		if s.PlanName == plan {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// recoveryRateFor returns the historical recovery rate for a risk state.
func recoveryRateFor(state valueobject.RiskState) float64 {
	switch state {
	case valueobject.RiskStateOneCycleMissed:
		return recoveryRateOneCycle
	case valueobject.RiskStateTwoCyclesMissed:
		return recoveryRateTwoCycle
	default:
		return 0
	}
}

// daysLate returns whole days between the expected charge date and now (0 if unset).
func daysLate(expectedNextCharge *time.Time, now time.Time) int {
	if expectedNextCharge == nil {
		return 0
	}
	return int(now.Sub(*expectedNextCharge).Hours() / 24)
}

// buildStores converts at-risk subscriptions to sorted report rows (MRR desc).
func buildStores(subs []*entity.Subscription, now time.Time) []revenueAtRiskStore {
	stores := make([]revenueAtRiskStore, 0, len(subs))
	for _, s := range subs {
		expected := ""
		if s.ExpectedNextChargeDate != nil {
			expected = s.ExpectedNextChargeDate.Format(dateLayout)
		}
		recoverable := int64(math.Round(float64(s.BasePriceCents) * recoveryRateFor(s.RiskState)))
		stores = append(stores, revenueAtRiskStore{
			Domain:             s.MyshopifyDomain,
			ShopName:           s.ShopName,
			MRRCents:           s.BasePriceCents,
			RiskState:          s.RiskState.String(),
			DaysLate:           daysLate(s.ExpectedNextChargeDate, now),
			ExpectedChargeDate: expected,
			PlanName:           s.PlanName,
			RecoverableCents:   recoverable,
		})
	}
	sort.SliceStable(stores, func(i, j int) bool {
		return stores[i].MRRCents > stores[j].MRRCents
	})
	return stores
}

// buildReport aggregates totals, per-state sums, counts, recoverable and currency.
func buildReport(subs []*entity.Subscription, stores []revenueAtRiskStore) revenueAtRiskReport {
	var oneCycleCents, twoCycleCents int64
	var oneCycleCount, twoCycleCount int
	currency := "USD"
	for _, s := range subs {
		if currency == "USD" && s.Currency != "" {
			currency = s.Currency
		}
		switch s.RiskState {
		case valueobject.RiskStateOneCycleMissed:
			oneCycleCents += s.BasePriceCents
			oneCycleCount++
		case valueobject.RiskStateTwoCyclesMissed:
			twoCycleCents += s.BasePriceCents
			twoCycleCount++
		}
	}

	recoverable := int64(math.Round(float64(oneCycleCents)*recoveryRateOneCycle + float64(twoCycleCents)*recoveryRateTwoCycle))

	return revenueAtRiskReport{
		Currency:         currency,
		TotalAtRiskCents: oneCycleCents + twoCycleCents,
		RecoverableCents: recoverable,
		ByState: revenueAtRiskByState{
			OneCycleCents: oneCycleCents,
			TwoCycleCents: twoCycleCents,
		},
		Counts: revenueAtRiskCounts{
			OneCycle: oneCycleCount,
			TwoCycle: twoCycleCount,
		},
		Trend:  []revenueAtRiskTrendPoint{},
		Stores: stores,
	}
}

// buildTrend converts daily snapshots to the at-risk MRR time series.
func buildTrend(snapshots []*entity.DailyMetricsSnapshot) []revenueAtRiskTrendPoint {
	trend := make([]revenueAtRiskTrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		trend = append(trend, revenueAtRiskTrendPoint{
			Date:        snap.Date.Format(dateLayout),
			AtRiskCents: snap.RevenueAtRiskCents,
		})
	}
	return trend
}

// writeStoresCSV writes the ranked stores as a CSV attachment.
func writeStoresCSV(w http.ResponseWriter, stores []revenueAtRiskStore) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="revenue-at-risk.csv"`)
	fmt.Fprintln(w, "domain,shopName,mrrCents,riskState,daysLate,expectedChargeDate,planName,recoverableCents")
	for _, s := range stores {
		fmt.Fprintf(w, "%s,%s,%d,%s,%d,%s,%s,%d\n",
			s.Domain, s.ShopName, s.MRRCents, s.RiskState, s.DaysLate,
			s.ExpectedChargeDate, s.PlanName, s.RecoverableCents)
	}
}

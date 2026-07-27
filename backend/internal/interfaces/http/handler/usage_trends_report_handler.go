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

// UsageTrendsReportHandler serves the "Usage Trends" report (REPORTS.md — Archetype A,
// momentum: weekly usage-revenue buckets + week-over-week growth). It measures the
// momentum of USAGE-based (metered) revenue only, kept strictly separated from MRR per
// the Revenue Classification rule: RECURRING, ONE_TIME and REFUND transactions are
// ignored entirely by design. Unlike the Usage & One-Time report it takes NO snapshot
// repo — the weekly trend is derived directly from the in-window USAGE transactions
// (bucketed by ISO-week Monday), so the report is self-contained for the selected range.
type UsageTrendsReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewUsageTrendsReportHandler constructs a UsageTrendsReportHandler.
func NewUsageTrendsReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *UsageTrendsReportHandler {
	return &UsageTrendsReportHandler{
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// usageTrendsStore is a single per-store row: total USAGE revenue in the window and the
// store's own week-over-week momentum.
type usageTrendsStore struct {
	Domain     string  `json:"domain"`
	ShopName   string  `json:"shopName"`
	UsageCents int64   `json:"usageCents"`
	WowPct     float64 `json:"wowPct"`
}

// usageTrendsWeekPoint is a single weekly bucket in the usage-revenue trend.
type usageTrendsWeekPoint struct {
	WeekStart  string `json:"weekStart"`
	UsageCents int64  `json:"usageCents"`
}

// usageTrendsReport is the full JSON contract for the Usage Trends report.
type usageTrendsReport struct {
	Currency           string                 `json:"currency"`
	UsageMrrEquivCents int64                  `json:"usageMrrEquivCents"`
	WowChangePct       float64                `json:"wowChangePct"`
	ActiveStores       int                    `json:"activeStores"`
	WeeklyTrend        []usageTrendsWeekPoint `json:"weeklyTrend"`
	Stores             []usageTrendsStore     `json:"stores"`
}

// GetUsageTrends returns the Usage Trends report for an app.
// GET /api/v1/apps/{appID}/reports/usage-trends?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *UsageTrendsReportHandler) GetUsageTrends(w http.ResponseWriter, r *http.Request) {
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

	txs, err := h.txRepo.FindByAppID(r.Context(), app.ID, from, to)
	if err != nil {
		writeUsageTrendsRepoError(w, "FindByAppID", err)
		return
	}

	report := buildUsageTrendsReport(txs)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeUsageTrendsStoresCSV(w, report.Stores)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("usage-trends: encode report: %v", err)
	}
}

// writeUsageTrendsRepoError logs a repository failure and responds 503. The transaction
// repo has no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeUsageTrendsRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("usage-trends: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// mondayOf truncates t to the Monday (00:00:00 UTC) of its ISO week. ISO weeks start on
// Monday; Go's time.Weekday puts Sunday at 0, so Sunday is treated as 6 days after the
// Monday. The returned time is normalized to UTC midnight so it serializes to a stable
// dateLayout string usable as a bucket key.
func mondayOf(t time.Time) time.Time {
	u := t.UTC()
	day := time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
	// Days since Monday: Mon=0 … Sun=6.
	offset := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -offset)
}

// wowPct returns the SIGNED week-over-week growth ratio (latest − prior) / prior. This
// is a growth ratio, NOT a rate — it is intentionally NOT clamped, so a doubling week
// yields +1.0 and a halving week yields -0.5, and it may be negative or exceed 1.
// Returns 0 when prior <= 0 (no meaningful baseline).
func wowPct(latest, prior int64) float64 {
	if prior <= 0 {
		return 0
	}
	return float64(latest-prior) / float64(prior)
}

// weeklyUsage buckets USAGE transactions into ISO-week buckets keyed by the Monday of
// each tx's TransactionDate (formatted dateLayout), summing AmountCents() (net). It
// skips every non-USAGE type itself (callers may pass mixed txs), so the buckets stay
// strictly MRR-separated.
func weeklyUsage(txs []*entity.Transaction) map[string]int64 {
	buckets := map[string]int64{}
	for _, tx := range txs {
		if tx.ChargeType != valueobject.ChargeTypeUsage {
			continue
		}
		key := mondayOf(tx.TransactionDate).Format(dateLayout)
		buckets[key] += tx.AmountCents()
	}
	return buckets
}

// sortedWeeks returns the bucket keys (weekStart date strings) sorted ascending. Because
// keys are dateLayout ("2006-01-02") strings, lexical sort equals chronological sort.
func sortedWeeks(buckets map[string]int64) []string {
	weeks := make([]string, 0, len(buckets))
	for k := range buckets {
		weeks = append(weeks, k)
	}
	sort.Strings(weeks)
	return weeks
}

// bucketWow computes the SIGNED WoW ratio from the two most-recent weekly buckets. It is
// 0 when there are fewer than 2 buckets or the prior week's total is <= 0.
func bucketWow(buckets map[string]int64) float64 {
	weeks := sortedWeeks(buckets)
	if len(weeks) < 2 {
		return 0
	}
	prior := buckets[weeks[len(weeks)-2]]
	latest := buckets[weeks[len(weeks)-1]]
	return wowPct(latest, prior)
}

// buildUsageTrendsReport aggregates the Usage Trends report from the in-window
// transactions. USAGE is the only charge type considered — RECURRING (which belongs to
// MRR), ONE_TIME and REFUND are ignored entirely by design, keeping USAGE strictly
// separated from MRR. usageMrrEquivCents is the Σ of net USAGE amounts across the whole
// window — for the default 30-day range this trailing-window usage total ≈ a monthly
// MRR-equivalent figure. The weekly trend and headline WoW come from ISO-week buckets;
// per-store rows carry each store's own window total and WoW.
func buildUsageTrendsReport(txs []*entity.Transaction) usageTrendsReport {
	type agg struct {
		usageCents int64
		shopName   string
		weeks      map[string]int64
	}
	byDomain := map[string]*agg{}
	order := make([]string, 0)

	var totalUsageCents int64
	// App-level weekly buckets drive both the weeklyTrend and the headline WoW.
	appWeeks := weeklyUsage(txs)

	for _, tx := range txs {
		// USAGE only — skip RECURRING/ONE_TIME/REFUND (and any unrecognized type) so
		// usage momentum stays strictly separated from MRR.
		if tx.ChargeType != valueobject.ChargeTypeUsage {
			continue
		}

		net := tx.AmountCents()
		totalUsageCents += net
		week := mondayOf(tx.TransactionDate).Format(dateLayout)

		a, ok := byDomain[tx.MyshopifyDomain]
		if !ok {
			a = &agg{weeks: map[string]int64{}}
			byDomain[tx.MyshopifyDomain] = a
			order = append(order, tx.MyshopifyDomain)
		}
		if a.shopName == "" && tx.ShopName != "" {
			a.shopName = tx.ShopName
		}
		a.usageCents += net
		a.weeks[week] += net
	}

	// Weekly trend: one point per ISO-week bucket, ascending by weekStart.
	weeklyTrend := make([]usageTrendsWeekPoint, 0, len(appWeeks))
	for _, wk := range sortedWeeks(appWeeks) {
		weeklyTrend = append(weeklyTrend, usageTrendsWeekPoint{
			WeekStart:  wk,
			UsageCents: appWeeks[wk],
		})
	}

	// Per-store rows: window total + each store's own WoW. activeStores is the count of
	// distinct domains with >= 1 USAGE tx in range (== number of buckets in byDomain).
	stores := make([]usageTrendsStore, 0, len(byDomain))
	for _, domain := range order {
		a := byDomain[domain]
		stores = append(stores, usageTrendsStore{
			Domain:     domain,
			ShopName:   a.shopName,
			UsageCents: a.usageCents,
			WowPct:     bucketWow(a.weeks),
		})
	}
	sort.SliceStable(stores, func(i, j int) bool {
		return stores[i].UsageCents > stores[j].UsageCents
	})

	return usageTrendsReport{
		Currency:           usageCurrency(txs),
		UsageMrrEquivCents: totalUsageCents,
		WowChangePct:       bucketWow(appWeeks),
		ActiveStores:       len(byDomain),
		WeeklyTrend:        weeklyTrend,
		Stores:             stores,
	}
}

// writeUsageTrendsStoresCSV writes the per-store table as a CSV attachment. Uses
// encoding/csv so free-text domains/shop names with commas/quotes stay one column.
func writeUsageTrendsStoresCSV(w http.ResponseWriter, stores []usageTrendsStore) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="usage-trends.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"domain", "shopName", "usageCents", "wowPct"})
	for _, s := range stores {
		_ = cw.Write([]string{
			s.Domain,
			s.ShopName,
			strconv.FormatInt(s.UsageCents, 10),
			strconv.FormatFloat(s.WowPct, 'f', 4, 64),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("usage-trends: write CSV: %v", err)
	}
}

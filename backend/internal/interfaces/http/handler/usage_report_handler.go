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

// UsageReportHandler serves the "Usage & One-Time Charges" report (REPORTS.md —
// Archetype A, Trend + Ranked Table): the non-recurring revenue streams (USAGE and
// ONE_TIME) with a per-store ranked breakdown and a usage-revenue trend from daily
// snapshots. USAGE is kept strictly separated from MRR per the Revenue Classification
// rule; RECURRING and REFUND transactions are ignored entirely. Mirrors the
// RetentionHandler structure (snapshot trend + aggregation + ranked table).
type UsageReportHandler struct {
	txRepo       repository.TransactionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
}

// NewUsageReportHandler constructs a UsageReportHandler.
func NewUsageReportHandler(
	txRepo repository.TransactionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *UsageReportHandler {
	return &UsageReportHandler{
		txRepo:       txRepo,
		snapshotRepo: snapshotRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
	}
}

// usageStore is a single per-store row in the ranked table.
type usageStore struct {
	Domain       string `json:"domain"`
	ShopName     string `json:"shopName"`
	UsageCents   int64  `json:"usageCents"`
	OneTimeCents int64  `json:"oneTimeCents"`
	ChargeCount  int    `json:"chargeCount"`
}

// usageTrendPoint is a single point in the usage-revenue time series.
type usageTrendPoint struct {
	Date       string `json:"date"`
	UsageCents int64  `json:"usageCents"`
}

// usageReport is the full JSON contract for the Usage & One-Time report.
type usageReport struct {
	Currency     string            `json:"currency"`
	UsageCents   int64             `json:"usageCents"`
	OneTimeCents int64             `json:"oneTimeCents"`
	ChargesCount int               `json:"chargesCount"`
	Trend        []usageTrendPoint `json:"trend"`
	// Interval is the trend granularity: day / week / month.
	Interval string       `json:"interval"`
	Stores   []usageStore `json:"stores"`
	// StoresTotal is the full store count before ?limit/?offset paging, so the
	// report preview and the dedicated page can show "N of M" / page correctly.
	StoresTotal int64 `json:"storesTotal"`
}

// GetUsageReport returns the Usage & One-Time Charges report for an app.
// GET /api/v1/apps/{appID}/reports/usage?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *UsageReportHandler) GetUsageReport(w http.ResponseWriter, r *http.Request) {
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
		writeUsageRepoError(w, "FindByAppID", err)
		return
	}

	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, to)
	if err != nil {
		writeUsageRepoError(w, "FindByAppIDRange", err)
		return
	}

	interval := resolveTrendInterval(from, to)
	report := buildUsageReport(txs)
	report.Interval = string(interval)
	// UsageRevenueCents is a rolling-12mo (as-of) figure, so last-in-bucket downsampling
	// is correct — a monthly point is the rolling value as of end of month, not a sum.
	report.Trend = buildUsageTrend(downsampleSnapshots(snapshots, interval))

	allStores := report.Stores
	report.StoresTotal = int64(len(allStores))

	// CSV exports the full table (all rows), regardless of paging.
	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeUsageStoresCSV(w, allStores)
		return
	}

	// Page only the JSON store rows; KPIs above already reflect the full set.
	limit, offset := parsePaging(r)
	report.Stores = pageSlice(allStores, offset, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("usage: encode report: %v", err)
	}
}

// writeUsageRepoError logs a repository failure and responds 503. Neither the
// transaction nor the snapshot repo has a not-found sentinel — every error is an
// infrastructure failure (ADR-042).
func writeUsageRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("usage: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// buildUsageReport sums net amounts for the two non-recurring streams and builds the
// per-store ranked table. Only USAGE and ONE_TIME charges are considered — RECURRING
// (which belongs to MRR) and REFUND are ignored entirely, keeping USAGE strictly
// separated from MRR per the Revenue Classification rule. Uses AmountCents() (net) for
// every sum, matching the other reports. Stores are grouped by MyshopifyDomain and
// sorted by usage revenue descending; an empty domain is kept as its own bucket
// (not special-cased), mirroring the retention/churn reports.
func buildUsageReport(txs []*entity.Transaction) usageReport {
	type agg struct {
		usageCents   int64
		oneTimeCents int64
		chargeCount  int
		shopName     string
	}
	byDomain := map[string]*agg{}
	order := make([]string, 0)

	var totalUsageCents, totalOneTimeCents int64
	var chargesCount int

	for _, tx := range txs {
		// Only the two non-recurring streams count; skip RECURRING/REFUND (and any
		// unrecognized type) so USAGE stays strictly separated from MRR.
		if tx.ChargeType != valueobject.ChargeTypeUsage && tx.ChargeType != valueobject.ChargeTypeOneTime {
			continue
		}

		a, ok := byDomain[tx.MyshopifyDomain]
		if !ok {
			a = &agg{}
			byDomain[tx.MyshopifyDomain] = a
			order = append(order, tx.MyshopifyDomain)
		}
		// First non-empty shop name wins for the bucket.
		if a.shopName == "" && tx.ShopName != "" {
			a.shopName = tx.ShopName
		}
		a.chargeCount++
		chargesCount++

		switch tx.ChargeType {
		case valueobject.ChargeTypeUsage:
			a.usageCents += tx.AmountCents()
			totalUsageCents += tx.AmountCents()
		case valueobject.ChargeTypeOneTime:
			a.oneTimeCents += tx.AmountCents()
			totalOneTimeCents += tx.AmountCents()
		}
	}

	stores := make([]usageStore, 0, len(byDomain))
	for _, domain := range order {
		a := byDomain[domain]
		stores = append(stores, usageStore{
			Domain:       domain,
			ShopName:     a.shopName,
			UsageCents:   a.usageCents,
			OneTimeCents: a.oneTimeCents,
			ChargeCount:  a.chargeCount,
		})
	}
	sort.SliceStable(stores, func(i, j int) bool {
		return stores[i].UsageCents > stores[j].UsageCents
	})

	return usageReport{
		Currency:     usageCurrency(txs),
		UsageCents:   totalUsageCents,
		OneTimeCents: totalOneTimeCents,
		ChargesCount: chargesCount,
		Trend:        []usageTrendPoint{},
		Stores:       stores,
	}
}

// usageCurrency returns the first transaction's non-empty currency, defaulting to
// "USD" when no transaction carries one.
func usageCurrency(txs []*entity.Transaction) string {
	for _, tx := range txs {
		if tx.Currency != "" {
			return tx.Currency
		}
	}
	return "USD"
}

// buildUsageTrend converts daily snapshots (ascending) to the usage-revenue time
// series, one point per snapshot from snap.UsageRevenueCents. NOTE: that snapshot
// field is USAGE-only and a rolling 12-month figure (see DailyMetricsSnapshot), so
// the trend is intentionally narrower than the headline (which is the windowed
// USAGE+ONE_TIME transaction sum) — its last point need not equal the KPI.
func buildUsageTrend(snapshots []*entity.DailyMetricsSnapshot) []usageTrendPoint {
	trend := make([]usageTrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		trend = append(trend, usageTrendPoint{
			Date:       snap.Date.Format(dateLayout),
			UsageCents: snap.UsageRevenueCents,
		})
	}
	return trend
}

// writeUsageStoresCSV writes the per-store ranked table as a CSV attachment. Uses
// encoding/csv so free-text domains/shop names with commas/quotes stay one column.
func writeUsageStoresCSV(w http.ResponseWriter, stores []usageStore) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="usage.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"domain", "shopName", "usageCents", "oneTimeCents", "chargeCount"})
	for _, s := range stores {
		_ = cw.Write([]string{
			s.Domain,
			s.ShopName,
			strconv.FormatInt(s.UsageCents, 10),
			strconv.FormatInt(s.OneTimeCents, 10),
			strconv.Itoa(s.ChargeCount),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("usage: write CSV: %v", err)
	}
}

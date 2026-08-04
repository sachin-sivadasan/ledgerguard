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

// NetNewSubsReportHandler serves the "Net-New Subscriptions" report (REPORTS.md — Growth,
// Archetype A): new vs churned subscriptions over a period with a daily net-new trend and
// a recent-new-subscriptions table. Derived entirely from subscriptions — "new" = subs
// whose StartDate() (ActivatedAt, the real business start — NOT the record-created
// CreatedAt) falls in range, "churned" = subs whose effective churn date (churnedDateOf,
// or UpdatedAt when no charge date) falls in range — so the KPIs, trend and table all
// reconcile. Mirrors the MRRReportHandler's new/churned movement logic (as counts).
type NetNewSubsReportHandler struct {
	subRepo     repository.SubscriptionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewNetNewSubsReportHandler constructs a NetNewSubsReportHandler.
func NewNetNewSubsReportHandler(
	subRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *NetNewSubsReportHandler {
	return &NetNewSubsReportHandler{
		subRepo:     subRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// netNewTrendPoint is a single day in the net-new time series.
type netNewTrendPoint struct {
	Date    string `json:"date"`
	New     int    `json:"new"`
	Churned int    `json:"churned"`
	Net     int    `json:"net"`
}

// newSubRow is a single row in the recent-new-subscriptions table.
type newSubRow struct {
	Domain   string `json:"domain"`
	ShopName string `json:"shopName"`
	PlanName string `json:"planName"`
	MrrCents int64  `json:"mrrCents"`
	Started  string `json:"started"`
}

// netNewSubsReport is the full JSON contract for the Net-New Subscriptions report.
type netNewSubsReport struct {
	Currency  string             `json:"currency"`
	NewSubs   int                `json:"newSubs"`
	Churned   int                `json:"churned"`
	Net       int                `json:"net"`
	Interval  string             `json:"interval"` // trend granularity: day / week / month
	Trend     []netNewTrendPoint `json:"trend"`
	NewStores []newSubRow        `json:"newStores"`
	// NewStoresTotal is the full new-stores count before ?limit/?offset paging, so
	// the report preview and the dedicated page can show "N of M" / page correctly.
	NewStoresTotal int64 `json:"newStoresTotal"`
}

// GetNetNewSubs returns the Net-New Subscriptions report for an app.
// GET /api/v1/apps/{appID}/reports/net-new-subscriptions?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *NetNewSubsReportHandler) GetNetNewSubs(w http.ResponseWriter, r *http.Request) {
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
		writeNetNewSubsRepoError(w, "FindByAppID", err)
		return
	}

	report := buildNetNewSubsReport(subs, from, to, newPlanLabeler(nil))
	allStores := report.NewStores
	report.NewStoresTotal = int64(len(allStores))

	// CSV exports the full table (all rows), regardless of paging.
	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeNewSubsCSV(w, allStores)
		return
	}

	// Page only the JSON store rows; KPIs and the trend already count all subs.
	limit, offset := parsePaging(r)
	report.NewStores = pageSlice(allStores, offset, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("net-new-subs: encode report: %v", err)
	}
}

// writeNetNewSubsRepoError logs a repository failure and responds 503. The subscription
// repo has no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeNetNewSubsRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("net-new-subs: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// buildNetNewSubsReport counts new subscriptions (StartDate() in [from,to]) and churned
// subscriptions (churnedDateOf in [from,to], whole `to` day inclusive), builds a daily
// net-new trend (only days with activity, ascending), and the full recent-new-subs table
// (newest first); the caller pages that table via parsePaging/pageSlice. Net = new −
// churned. A sub that both started and churned in the window counts in both (net 0 for
// it), which is the correct net-growth semantics.
func buildNetNewSubsReport(subs []*entity.Subscription, from, to time.Time, labeler planLabeler) netNewSubsReport {
	toExclusive := to.AddDate(0, 0, 1)
	inRange := func(t time.Time) bool { return !t.Before(from) && t.Before(toExclusive) }
	interval := resolveTrendInterval(from, to)

	// Currency: first non-empty wins (break avoids the "USD"-sentinel-collision bug).
	currency := "USD"
	for _, s := range subs {
		if s.Currency != "" {
			currency = s.Currency
			break
		}
	}

	type dayAgg struct{ newCount, churnedCount int }
	byDay := map[string]*dayAgg{}
	day := func(k string) *dayAgg {
		a, ok := byDay[k]
		if !ok {
			a = &dayAgg{}
			byDay[k] = a
		}
		return a
	}

	var newSubs, churned, noChurnDate, noStartDate int
	newList := make([]*entity.Subscription, 0)

	for _, s := range subs {
		start := s.StartDate() // business start (ActivatedAt); falls back to CreatedAt when nil
		if inRange(start) {
			newSubs++
			day(bucketKeyOf(start, interval)).newCount++
			newList = append(newList, s)
			if s.ActivatedAt == nil {
				// No real start date — dated by the CreatedAt (ingestion) fallback, which
				// can misdate a recurring-less (e.g. trial) sub. Count so it's diagnosable.
				noStartDate++
			}
		}
		if s.RiskState == valueobject.RiskStateChurned {
			cd := churnedDateOf(s)
			if cd == nil {
				// Churned before any charge date was recorded (e.g. cancelled during a
				// trial / before the first charge). Fall back to UpdatedAt so the churn
				// isn't silently dropped from the count (which would inflate Net).
				cd = &s.UpdatedAt
				noChurnDate++
			}
			if inRange(*cd) {
				churned++
				day(bucketKeyOf(*cd, interval)).churnedCount++
			}
		}
	}
	if noChurnDate > 0 {
		log.Printf("net-new-subs: %d churned subscription(s) had no charge date — used UpdatedAt as the churn date (cancelled before first charge?)", noChurnDate)
	}
	if noStartDate > 0 {
		log.Printf("net-new-subs: %d new subscription(s) had no activated_at — dated by the CreatedAt (ingestion) fallback, may be misdated", noStartDate)
	}

	// Trend: only days with activity, ascending (YYYY-MM-DD keys sort chronologically).
	days := make([]string, 0, len(byDay))
	for k := range byDay {
		days = append(days, k)
	}
	sort.Strings(days)
	trend := make([]netNewTrendPoint, 0, len(days))
	for _, k := range days {
		a := byDay[k]
		trend = append(trend, netNewTrendPoint{
			Date:    k,
			New:     a.newCount,
			Churned: a.churnedCount,
			Net:     a.newCount - a.churnedCount,
		})
	}

	// Recent new subscriptions: newest first (by StartDate — the business start). The
	// full list is returned; the handler pages it (KPIs and the trend count all subs).
	sort.SliceStable(newList, func(i, j int) bool {
		return newList[i].StartDate().After(newList[j].StartDate())
	})
	newStores := make([]newSubRow, 0, len(newList))
	for _, s := range newList {
		newStores = append(newStores, newSubRow{
			Domain:   s.MyshopifyDomain,
			ShopName: s.ShopName,
			PlanName: labeler.label(s),
			MrrCents: s.MRRCents(),
			Started:  s.StartDate().Format(dateLayout),
		})
	}

	return netNewSubsReport{
		Currency:  currency,
		NewSubs:   newSubs,
		Churned:   churned,
		Net:       newSubs - churned,
		Interval:  string(interval),
		Trend:     trend,
		NewStores: newStores,
	}
}

// writeNewSubsCSV writes the recent-new-subscriptions table as a CSV attachment. Uses
// encoding/csv so free-text domains/shop/plan names with commas/quotes stay one column.
func writeNewSubsCSV(w http.ResponseWriter, rows []newSubRow) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="net-new-subscriptions.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"domain", "shopName", "plan", "mrrCents", "started"})
	for _, row := range rows {
		_ = cw.Write([]string{
			row.Domain,
			row.ShopName,
			row.PlanName,
			strconv.FormatInt(row.MrrCents, 10),
			row.Started,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("net-new-subs: write CSV: %v", err)
	}
}

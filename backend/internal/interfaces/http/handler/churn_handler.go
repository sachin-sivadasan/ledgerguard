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

// ChurnHandler serves the "Churn" report (REPORTS.md — churned count + MRR lost,
// churn %). Mirrors the RevenueAtRiskHandler structure.
type ChurnHandler struct {
	subRepo      repository.SubscriptionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
}

// NewChurnHandler constructs a ChurnHandler.
func NewChurnHandler(
	subRepo repository.SubscriptionRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *ChurnHandler {
	return &ChurnHandler{
		subRepo:      subRepo,
		snapshotRepo: snapshotRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
	}
}

// churnStore is a single churned store row in the report.
type churnStore struct {
	Domain       string `json:"domain"`
	ShopName     string `json:"shopName"`
	MRRLostCents int64  `json:"mrrLostCents"`
	ChurnedDate  string `json:"churnedDate"`
	TenureDays   int    `json:"tenureDays"`
	PlanName     string `json:"planName"`
}

// churnTrendPoint is a single point in the churn-rate time series.
type churnTrendPoint struct {
	Date      string  `json:"date"`
	ChurnRate float64 `json:"churnRate"`
}

// churnReport is the full JSON contract for the Churn report.
type churnReport struct {
	Currency            string            `json:"currency"`
	ChurnRate           float64           `json:"churnRate"`
	ChurnedMrrLostCents int64             `json:"churnedMrrLostCents"`
	ChurnedCount        int               `json:"churnedCount"`
	Trend               []churnTrendPoint `json:"trend"`
	Stores              []churnStore      `json:"stores"`
}

// GetChurn returns the Churn report for an app.
// GET /api/v1/apps/{appID}/reports/churn?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *ChurnHandler) GetChurn(w http.ResponseWriter, r *http.Request) {
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

	churned, err := h.subRepo.FindByRiskState(r.Context(), app.ID, valueobject.RiskStateChurned)
	if err != nil {
		writeChurnRepoError(w, "FindByRiskState(churned)", err)
		return
	}

	snapshots, err := h.snapshotRepo.FindByAppIDRange(r.Context(), app.ID, from, to)
	if err != nil {
		writeChurnRepoError(w, "FindByAppIDRange", err)
		return
	}

	stores := buildChurnStores(churned, now)
	report := buildChurnReport(churned, stores, latestSnapshot(snapshots))
	report.Trend = buildChurnTrend(snapshots)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeChurnStoresCSV(w, stores)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("churn: encode report: %v", err)
	}
}

// writeChurnRepoError logs a repository failure and responds 503. These repos have
// no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeChurnRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("churn: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// latestSnapshot returns the newest snapshot in the range (the last element, as
// FindByAppIDRange returns snapshots ordered by date ascending), or nil if empty.
func latestSnapshot(snapshots []*entity.DailyMetricsSnapshot) *entity.DailyMetricsSnapshot {
	if len(snapshots) == 0 {
		return nil
	}
	return snapshots[len(snapshots)-1]
}

// churnedDateOf returns the effective churn date for a subscription, preferring
// ExpectedNextChargeDate, falling back to LastRecurringChargeDate (nil-safe).
func churnedDateOf(s *entity.Subscription) *time.Time {
	if s.ExpectedNextChargeDate != nil {
		return s.ExpectedNextChargeDate
	}
	return s.LastRecurringChargeDate
}

// tenureDays returns whole days from createdAt to the churn date (or now if unset),
// floored at 0.
func tenureDays(createdAt time.Time, churnedDate *time.Time, now time.Time) int {
	end := now
	if churnedDate != nil {
		end = *churnedDate
	}
	days := int(end.Sub(createdAt).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// buildChurnStores converts churned subscriptions to sorted report rows (MRR lost desc).
func buildChurnStores(subs []*entity.Subscription, now time.Time) []churnStore {
	stores := make([]churnStore, 0, len(subs))
	for _, s := range subs {
		churned := churnedDateOf(s)
		churnedStr := ""
		if churned != nil {
			churnedStr = churned.Format(dateLayout)
		}
		stores = append(stores, churnStore{
			Domain:       s.MyshopifyDomain,
			ShopName:     s.ShopName,
			MRRLostCents: s.MRRCents(),
			ChurnedDate:  churnedStr,
			TenureDays:   tenureDays(s.CreatedAt, churned, now),
			PlanName:     s.PlanName,
		})
	}
	sort.SliceStable(stores, func(i, j int) bool {
		return stores[i].MRRLostCents > stores[j].MRRLostCents
	})
	return stores
}

// churnRate returns churnedCount ÷ totalSubscriptions, guarding divide-by-zero.
func churnRate(churnedCount, totalSubscriptions int) float64 {
	if totalSubscriptions <= 0 {
		return 0
	}
	return float64(churnedCount) / float64(totalSubscriptions)
}

// buildChurnReport aggregates churned count, MRR lost, churn rate and currency.
func buildChurnReport(subs []*entity.Subscription, stores []churnStore, latest *entity.DailyMetricsSnapshot) churnReport {
	var mrrLostCents int64
	currency := "USD"
	for _, s := range subs {
		if currency == "USD" && s.Currency != "" {
			currency = s.Currency
		}
		mrrLostCents += s.MRRCents()
	}

	total := 0
	if latest != nil {
		total = latest.TotalSubscriptions
	}

	return churnReport{
		Currency:            currency,
		ChurnRate:           churnRate(len(subs), total),
		ChurnedMrrLostCents: mrrLostCents,
		ChurnedCount:        len(subs),
		Trend:               []churnTrendPoint{},
		Stores:              stores,
	}
}

// buildChurnTrend converts daily snapshots to the churn-rate time series.
func buildChurnTrend(snapshots []*entity.DailyMetricsSnapshot) []churnTrendPoint {
	trend := make([]churnTrendPoint, 0, len(snapshots))
	for _, snap := range snapshots {
		trend = append(trend, churnTrendPoint{
			Date:      snap.Date.Format(dateLayout),
			ChurnRate: churnRate(snap.ChurnedCount, snap.TotalSubscriptions),
		})
	}
	return trend
}

// writeChurnStoresCSV writes the ranked churned stores as a CSV attachment. Uses
// encoding/csv so free-text fields (shopName, planName) with commas/quotes/newlines
// are quoted.
func writeChurnStoresCSV(w http.ResponseWriter, stores []churnStore) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="churn.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{
		"domain", "shopName", "mrrLostCents", "churnedDate", "tenureDays", "planName",
	})
	for _, s := range stores {
		_ = cw.Write([]string{
			s.Domain,
			s.ShopName,
			strconv.FormatInt(s.MRRLostCents, 10),
			s.ChurnedDate,
			strconv.Itoa(s.TenureDays),
			s.PlanName,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("churn: write CSV: %v", err)
	}
}

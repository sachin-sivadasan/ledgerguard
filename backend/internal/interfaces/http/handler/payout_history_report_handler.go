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
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// periodLayout groups paid earnings by calendar month ("2006-01").
const periodLayout = "2006-01"

// PayoutHistoryReportHandler serves the "Payout History" report (REPORTS.md —
// Archetype D, Schedule/Timeline): the historical record of PAID_OUT earnings,
// aggregated by the calendar month the charge was billed on Shopify (the charge date,
// NOT the row-ingestion time). It is the completed-payout counterpart to Payout Schedule
// (which shows only the not-yet-paid PENDING/AVAILABLE earnings). Amounts are net (what
// the developer received). Mirrors the EarningsReportHandler structure (transactions +
// stored EarningsStatus, no snapshot).
type PayoutHistoryReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewPayoutHistoryReportHandler constructs a PayoutHistoryReportHandler.
func NewPayoutHistoryReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *PayoutHistoryReportHandler {
	return &PayoutHistoryReportHandler{
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// payoutHistoryRow is one historical payout period (a calendar month of paid earnings).
type payoutHistoryRow struct {
	Period      string `json:"period"`      // charge month, "YYYY-MM"
	AmountCents int64  `json:"amountCents"` // Σ net paid in the period
	ChargeCount int    `json:"chargeCount"`
	// AvailableDate is the latest estimated availability date among the period's charges
	// ("YYYY-MM-DD", or "" when unset) — EarningsCalculator's ~7-day estimate, NOT
	// Shopify's authoritative disbursement date.
	AvailableDate string `json:"availableDate"`
}

// payoutHistoryReport is the full JSON contract for the Payout History report.
type payoutHistoryReport struct {
	Currency       string             `json:"currency"`
	TotalPaidCents int64              `json:"totalPaidCents"`
	PayoutCount    int                `json:"payoutCount"`
	AvgPayoutCents int64              `json:"avgPayoutCents"`
	Rows           []payoutHistoryRow `json:"rows"`
	// RowsTotal is the full row count before ?limit/?offset paging, so the report
	// preview and the dedicated page can show "N of M" / page correctly.
	RowsTotal int64 `json:"rowsTotal"`
}

// GetPayoutHistory returns the Payout History report for an app.
// GET /api/v1/apps/{appID}/reports/payout-history?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *PayoutHistoryReportHandler) GetPayoutHistory(w http.ResponseWriter, r *http.Request) {
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
		writePayoutHistoryRepoError(w, "FindByAppID", err)
		return
	}

	report := buildPayoutHistoryReport(txs)
	allRows := report.Rows
	report.RowsTotal = int64(len(allRows))

	// CSV exports the full table (all rows), regardless of paging.
	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writePayoutHistoryCSV(w, allRows)
		return
	}

	// Page only the JSON rows; KPIs above already reflect the full set.
	limit, offset := parsePaging(r)
	report.Rows = pageSlice(allRows, offset, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("payout-history: encode report: %v", err)
	}
}

// writePayoutHistoryRepoError logs a repository failure and responds 503. The
// transaction repo has no not-found sentinel — every error is an infrastructure
// failure (ADR-042).
func writePayoutHistoryRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("payout-history: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// chargeDate returns the Shopify charge date to group by — CreatedDate, falling back to
// TransactionDate when unset — matching EarningsCalculator. NOT CreatedAt, which is the
// row-ingestion time (time.Now() at sync) and would collapse all historical charges into
// the sync month.
func chargeDate(tx *entity.Transaction) time.Time {
	if !tx.CreatedDate.IsZero() {
		return tx.CreatedDate
	}
	return tx.TransactionDate
}

// buildPayoutHistoryReport aggregates PAID_OUT earnings into per-calendar-month payout
// periods, keyed by the Shopify charge date (see chargeDate). Only PAID_OUT transactions
// count — PENDING/AVAILABLE (still upcoming) belong to Payout Schedule and are skipped
// silently by design. Amounts are net (AmountCents = NetAmountCents). Each period's
// AvailableDate is the LATEST availability date among its charges (EarningsCalculator's
// ~7-day estimate, not Shopify's actual disbursement date). Rows are sorted by period
// descending (most recent first). TotalPaid is the sum of all periods; PayoutCount is the
// number of periods; AvgPayout = TotalPaid ÷ PayoutCount (0 when none).
func buildPayoutHistoryReport(txs []*entity.Transaction) payoutHistoryReport {
	type agg struct {
		amount           int64
		count            int
		latestAvailable  time.Time
		hasAvailableDate bool
	}
	byPeriod := map[string]*agg{}

	var totalPaidCents int64
	var noChargeDate int
	for _, tx := range txs {
		if tx.EarningsStatus != entity.EarningsStatusPaidOut {
			continue // not yet paid → Payout Schedule, not history
		}

		cd := chargeDate(tx)
		if cd.IsZero() {
			// A PAID_OUT charge with no charge date is a data anomaly; it would bucket
			// under "0001-01". Count it so the situation is diagnosable rather than a
			// silent garbage row, but keep the money in (still added below).
			noChargeDate++
		}
		period := cd.Format(periodLayout)
		a, ok := byPeriod[period]
		if !ok {
			a = &agg{}
			byPeriod[period] = a
		}
		a.amount += tx.AmountCents()
		a.count++
		totalPaidCents += tx.AmountCents()
		// Track the latest availability date in the period as its representative date.
		if !tx.AvailableDate.IsZero() && (!a.hasAvailableDate || tx.AvailableDate.After(a.latestAvailable)) {
			a.latestAvailable = tx.AvailableDate
			a.hasAvailableDate = true
		}
	}

	if noChargeDate > 0 {
		log.Printf("payout-history: %d PAID_OUT transaction(s) had no charge date (CreatedDate/TransactionDate) — bucketed under 0001-01", noChargeDate)
	}

	rows := make([]payoutHistoryRow, 0, len(byPeriod))
	for period, a := range byPeriod {
		availableDate := ""
		if a.hasAvailableDate {
			availableDate = a.latestAvailable.Format(dateLayout)
		}
		rows = append(rows, payoutHistoryRow{
			Period:        period,
			AmountCents:   a.amount,
			ChargeCount:   a.count,
			AvailableDate: availableDate,
		})
	}
	// Most recent period first (period keys are "YYYY-MM", so lexical == chronological).
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].Period > rows[j].Period
	})

	payoutCount := len(rows)
	var avgPayoutCents int64
	if payoutCount > 0 {
		avgPayoutCents = totalPaidCents / int64(payoutCount)
	}

	return payoutHistoryReport{
		Currency:       earningsCurrency(txs), // default USD + first non-empty (shared helper)
		TotalPaidCents: totalPaidCents,
		PayoutCount:    payoutCount,
		AvgPayoutCents: avgPayoutCents,
		Rows:           rows,
	}
}

// writePayoutHistoryCSV writes the payout log as a CSV attachment.
func writePayoutHistoryCSV(w http.ResponseWriter, rows []payoutHistoryRow) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="payout-history.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"period", "amountCents", "chargeCount", "availableDate"})
	for _, row := range rows {
		_ = cw.Write([]string{
			row.Period,
			strconv.FormatInt(row.AmountCents, 10),
			strconv.Itoa(row.ChargeCount),
			row.AvailableDate,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("payout-history: write CSV: %v", err)
	}
}

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

// PayoutScheduleReportHandler serves the "Payout Schedule" report (REPORTS.md —
// Archetype D, Schedule/Timeline): the UPCOMING (not-yet-paid) earnings grouped by
// their available date into a forward-looking payout timeline. That date is
// LedgerGuard-computed (charge date + the earnings-delay estimate from
// EarningsCalculator, ~7 days in MVP), NOT Shopify's authoritative payout date. Only PENDING
// and AVAILABLE earnings are shown — PAID_OUT belongs to the Payout History report.
// Amounts are net (what the developer receives). Mirrors the EarningsReportHandler
// structure (transactions + stored EarningsStatus, no snapshot repo).
type PayoutScheduleReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewPayoutScheduleReportHandler constructs a PayoutScheduleReportHandler.
func NewPayoutScheduleReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *PayoutScheduleReportHandler {
	return &PayoutScheduleReportHandler{
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// payoutScheduleRow is a single upcoming-payout group in the timeline: all not-yet-paid
// earnings sharing an available date and status.
type payoutScheduleRow struct {
	AvailableDate string `json:"availableDate"` // YYYY-MM-DD, or "" when unscheduled
	AmountCents   int64  `json:"amountCents"`   // Σ net earnings in the group
	ChargeCount   int    `json:"chargeCount"`
	Status        string `json:"status"` // "Available" or "Pending"
}

// payoutScheduleReport is the full JSON contract for the Payout Schedule report.
type payoutScheduleReport struct {
	Currency            string              `json:"currency"`
	UpcomingPayoutCents int64               `json:"upcomingPayoutCents"` // Σ AVAILABLE net (scheduled to pay next)
	PendingCents        int64               `json:"pendingCents"`        // Σ PENDING net (still clearing)
	NextPayoutDate      string              `json:"nextPayoutDate"`      // earliest scheduled date, or ""
	Rows                []payoutScheduleRow `json:"rows"`
}

// GetPayoutSchedule returns the Payout Schedule report for an app.
// GET /api/v1/apps/{appID}/reports/payout-schedule?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *PayoutScheduleReportHandler) GetPayoutSchedule(w http.ResponseWriter, r *http.Request) {
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
		writePayoutScheduleRepoError(w, "FindByAppID", err)
		return
	}

	report := buildPayoutScheduleReport(txs)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writePayoutScheduleCSV(w, report.Rows)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("payout-schedule: encode report: %v", err)
	}
}

// writePayoutScheduleRepoError logs a repository failure and responds 503. The
// transaction repo has no not-found sentinel — every error is an infrastructure
// failure (ADR-042).
func writePayoutScheduleRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("payout-schedule: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// buildPayoutScheduleReport groups the upcoming (PENDING + AVAILABLE) earnings by
// (available date, status) into timeline rows and computes the KPIs. PAID_OUT is
// excluded by design (it belongs to Payout History) and is NOT logged; any other,
// unrecognized status IS logged and dropped so the rows always reconcile with the
// KPIs (upcomingPayoutCents + pendingCents == Σ row amounts). Amounts are net
// (AmountCents = NetAmountCents). NextPayoutDate is the earliest scheduled date across
// the rows (rows are sorted so scheduled dates come first, ascending).
func buildPayoutScheduleReport(txs []*entity.Transaction) payoutScheduleReport {
	type agg struct {
		amount int64
		count  int
	}
	byKey := map[string]*agg{}

	var upcomingPayoutCents, pendingCents int64
	var unknown int

	for _, tx := range txs {
		var status string
		switch tx.EarningsStatus {
		case entity.EarningsStatusAvailable:
			status = "Available"
			upcomingPayoutCents += tx.AmountCents()
		case entity.EarningsStatusPending:
			status = "Pending"
			pendingCents += tx.AmountCents()
		case entity.EarningsStatusPaidOut:
			continue // already disbursed — Payout History, not the schedule
		default:
			unknown++
			continue
		}

		date := formatReportDate(tx.AvailableDate)
		key := date + "|" + status
		a, ok := byKey[key]
		if !ok {
			a = &agg{}
			byKey[key] = a
		}
		a.amount += tx.AmountCents()
		a.count++
	}

	if unknown > 0 {
		log.Printf("payout-schedule: excluded %d transaction(s) with an unrecognized EarningsStatus (would not reconcile with KPIs)", unknown)
	}

	rows := make([]payoutScheduleRow, 0, len(byKey))
	for key, a := range byKey {
		date, status, _ := strings.Cut(key, "|")
		rows = append(rows, payoutScheduleRow{
			AvailableDate: date,
			AmountCents:   a.amount,
			ChargeCount:   a.count,
			Status:        status,
		})
	}
	sortPayoutScheduleRows(rows)

	// The sort puts the earliest scheduled date first and sinks unscheduled ("") rows
	// last, so rows[0] is the next payout date — or "" only when every row is
	// unscheduled (then rows[0].AvailableDate is itself "", so the != "" check is
	// belt-and-suspenders, kept for intent).
	nextPayoutDate := ""
	if len(rows) > 0 && rows[0].AvailableDate != "" {
		nextPayoutDate = rows[0].AvailableDate
	}

	return payoutScheduleReport{
		Currency:            earningsCurrency(txs), // default USD + first non-empty (shared helper)
		UpcomingPayoutCents: upcomingPayoutCents,
		PendingCents:        pendingCents,
		NextPayoutDate:      nextPayoutDate,
		Rows:                rows,
	}
}

// sortPayoutScheduleRows orders the timeline: scheduled (non-empty) dates first,
// ascending by date, then by status ("Available" before "Pending" so the sooner-paying
// bucket leads a shared date); unscheduled ("") rows sink to the bottom.
func sortPayoutScheduleRows(rows []payoutScheduleRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if (a.AvailableDate == "") != (b.AvailableDate == "") {
			return a.AvailableDate != "" // non-empty dates first
		}
		if a.AvailableDate != b.AvailableDate {
			return a.AvailableDate < b.AvailableDate
		}
		return a.Status < b.Status
	})
}

// writePayoutScheduleCSV writes the timeline as a CSV attachment.
func writePayoutScheduleCSV(w http.ResponseWriter, rows []payoutScheduleRow) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="payout-schedule.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"availableDate", "amountCents", "chargeCount", "status"})
	for _, row := range rows {
		_ = cw.Write([]string{
			row.AvailableDate,
			strconv.FormatInt(row.AmountCents, 10),
			strconv.Itoa(row.ChargeCount),
			row.Status,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("payout-schedule: write CSV: %v", err)
	}
}

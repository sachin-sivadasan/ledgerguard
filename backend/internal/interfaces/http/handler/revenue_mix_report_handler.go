package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// RevenueMixReportHandler serves the "Revenue Mix" report (REPORTS.md — Archetype B,
// Composition): the split of net revenue across the three positive charge streams
// (Recurring / Usage / One-time), plus refunds as a separate adjustment. Mirrors the
// EarningsReportHandler structure and reuses the same per-ChargeType summing that the
// internal GraphQL earnings resolver performs. RECURRING and USAGE are kept strictly
// separated (never combined) per the Revenue Classification rule.
type RevenueMixReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewRevenueMixReportHandler constructs a RevenueMixReportHandler.
func NewRevenueMixReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *RevenueMixReportHandler {
	return &RevenueMixReportHandler{
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// revenueMixSegment is one composition slice: a positive revenue stream with its
// share of gross revenue (pct in [0,1]).
type revenueMixSegment struct {
	Type        string  `json:"type"`
	AmountCents int64   `json:"amountCents"`
	Pct         float64 `json:"pct"`
}

// revenueMixReport is the full JSON contract for the Revenue Mix report.
type revenueMixReport struct {
	Currency       string              `json:"currency"`
	RecurringCents int64               `json:"recurringCents"`
	UsageCents     int64               `json:"usageCents"`
	OneTimeCents   int64               `json:"oneTimeCents"`
	RefundCents    int64               `json:"refundCents"`
	GrossCents     int64               `json:"grossCents"`
	NetCents       int64               `json:"netCents"`
	Segments       []revenueMixSegment `json:"segments"`
}

// GetRevenueMix returns the Revenue Mix report for an app.
// GET /api/v1/apps/{appID}/reports/revenue-mix?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *RevenueMixReportHandler) GetRevenueMix(w http.ResponseWriter, r *http.Request) {
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
		writeRevenueMixRepoError(w, "FindByAppID", err)
		return
	}

	report := buildRevenueMixReport(txs)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeRevenueMixCSV(w, report)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("revenue-mix: encode report: %v", err)
	}
}

// buildRevenueMixReport sums net amounts by ChargeType and composes the report. The
// per-ChargeType summing mirrors the GraphQL earnings resolver (RECURRING, USAGE, and
// ONE_TIME are the three positive streams, REFUND a separate negative adjustment); the
// gross-vs-net split and the composition segments are additions specific to this report.
func buildRevenueMixReport(txs []*entity.Transaction) revenueMixReport {
	var recurringCents, usageCents, oneTimeCents, refundCents, unknownCents int64
	var unknownCount int
	for _, tx := range txs {
		switch tx.ChargeType {
		case valueobject.ChargeTypeRecurring:
			recurringCents += tx.AmountCents()
		case valueobject.ChargeTypeUsage:
			usageCents += tx.AmountCents()
		case valueobject.ChargeTypeOneTime:
			oneTimeCents += tx.AmountCents()
		case valueobject.ChargeTypeRefund:
			refundCents += tx.AmountCents()
		default:
			// An unrecognized/empty ChargeType (e.g. a new Partner API charge type)
			// is excluded from gross so segments stay internally consistent — but log
			// it so the resulting gross under-count is diagnosable, not silent.
			unknownCount++
			unknownCents += tx.AmountCents()
		}
	}
	if unknownCount > 0 {
		log.Printf("revenue-mix: %d transaction(s) with unrecognized ChargeType (%d cents) excluded from gross — total under-reports true revenue", unknownCount, unknownCents)
	}

	// Gross is the sum of the three positive streams; net subtracts refunds.
	grossCents := recurringCents + usageCents + oneTimeCents
	netCents := grossCents - refundCents

	// Segments are a composition of the positive streams: each pct is a share of
	// gross, so the three sum to ~1.0. Refunds are NOT a segment. Always emit all 3.
	segments := []revenueMixSegment{
		{Type: "Recurring", AmountCents: recurringCents, Pct: segmentPct(recurringCents, grossCents)},
		{Type: "Usage", AmountCents: usageCents, Pct: segmentPct(usageCents, grossCents)},
		{Type: "One-time", AmountCents: oneTimeCents, Pct: segmentPct(oneTimeCents, grossCents)},
	}

	return revenueMixReport{
		Currency:       revenueMixCurrency(txs),
		RecurringCents: recurringCents,
		UsageCents:     usageCents,
		OneTimeCents:   oneTimeCents,
		RefundCents:    refundCents,
		GrossCents:     grossCents,
		NetCents:       netCents,
		Segments:       segments,
	}
}

// segmentPct returns amount/gross clamped to [0,1]. Guards gross<=0 (empty report or
// all-refund) so there is never a divide-by-zero — it returns 0 in that case.
func segmentPct(amountCents, grossCents int64) float64 {
	if grossCents <= 0 {
		return 0
	}
	pct := float64(amountCents) / float64(grossCents)
	if pct < 0 {
		return 0
	}
	if pct > 1 {
		return 1
	}
	return pct
}

// revenueMixCurrency returns the first transaction's non-empty currency, defaulting
// to "USD" when no transaction carries one.
func revenueMixCurrency(txs []*entity.Transaction) string {
	for _, tx := range txs {
		if tx.Currency != "" {
			return tx.Currency
		}
	}
	return "USD"
}

// writeRevenueMixRepoError logs a repository failure and responds 503. The transaction
// repo has no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeRevenueMixRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("revenue-mix: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// writeRevenueMixCSV writes the composition as a CSV attachment: the 3 segment rows,
// then a Refund row (only when refunds exist) and a Net row. Uses encoding/csv so any
// future free-text stays safely quoted. pct is formatted to 4 decimal places; the
// Refund/Net rows leave pct blank (they are adjustments/totals, not composition slices).
func writeRevenueMixCSV(w http.ResponseWriter, report revenueMixReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="revenue-mix.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"type", "amountCents", "pct"})
	for _, s := range report.Segments {
		_ = cw.Write([]string{
			s.Type,
			strconv.FormatInt(s.AmountCents, 10),
			strconv.FormatFloat(s.Pct, 'f', 4, 64),
		})
	}
	if report.RefundCents > 0 {
		_ = cw.Write([]string{"Refund", strconv.FormatInt(-report.RefundCents, 10), ""})
	}
	_ = cw.Write([]string{"Net", strconv.FormatInt(report.NetCents, 10), ""})
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("revenue-mix: write CSV: %v", err)
	}
}

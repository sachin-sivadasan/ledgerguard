package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// LedgerReconciliationReportHandler serves the "Ledger Reconciliation" report
// (REPORTS.md — Guard): does the money add up? For each month it recomputes the net
// payout from Shopify's own figures (gross − fee) and reconciles it against the recorded
// net. A non-zero residual means the ledger and Shopify disagree for that month — most
// often incomplete fee data (a pre-fee-sync month) or a genuine drift to investigate.
type LedgerReconciliationReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
	feeService  *service.FeeVerificationService
}

// NewLedgerReconciliationReportHandler constructs a LedgerReconciliationReportHandler.
func NewLedgerReconciliationReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	feeService *service.FeeVerificationService,
) *LedgerReconciliationReportHandler {
	return &LedgerReconciliationReportHandler{txRepo: txRepo, appRepo: appRepo, partnerRepo: partnerRepo, feeService: feeService}
}

type reconMonthJSON struct {
	Month            string `json:"month"`
	GrossCents       int64  `json:"gross_cents"`
	FeeCents         int64  `json:"fee_cents"`
	NetCents         int64  `json:"net_cents"`
	ExpectedNetCents int64  `json:"expected_net_cents"` // gross − fee
	// ResidualCents = recorded net − expected net. Non-zero ⇒ the identity gross = net
	// + fee doesn't hold (usually missing fee data for that month).
	ResidualCents int64 `json:"residual_cents"`
	TxCount       int   `json:"tx_count"`
	Reconciled    bool  `json:"reconciled"`
}

type reconReport struct {
	Currency         string           `json:"currency"`
	TotalGrossCents  int64            `json:"total_gross_cents"`
	TotalFeeCents    int64            `json:"total_fee_cents"`
	TotalNetCents    int64            `json:"total_net_cents"`
	ResidualCents    int64            `json:"residual_cents"`
	Reconciled       bool             `json:"reconciled"` // every month reconciled
	MonthsReconciled int              `json:"months_reconciled"`
	MonthsFlagged    int              `json:"months_flagged"`
	MonthsAudited    int              `json:"months_audited"`
	Months           []reconMonthJSON `json:"months"`
}

// GetLedgerReconciliation returns the Ledger Reconciliation report for an app.
// GET /api/v1/apps/{appID}/reports/ledger-reconciliation?months=6&format=csv
func (h *LedgerReconciliationReportHandler) GetLedgerReconciliation(w http.ResponseWriter, r *http.Request) {
	if user := middleware.UserFromContext(r.Context()); user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	months := 6
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 24 {
			months = parsed
		}
	}

	now := time.Now()
	report := reconReport{Currency: "USD", Months: make([]reconMonthJSON, 0, months)}
	for i := months - 1; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := time.Date(now.Year(), now.Month()-time.Month(i)+1, 0, 23, 59, 59, 0, time.UTC)

		txs, err := h.txRepo.FindByAppID(r.Context(), app.ID, monthStart, monthEnd)
		if err != nil {
			continue
		}
		summary := h.feeService.CalculateFeeSummary(txs)
		gross := summary.TotalGrossAmountCents
		fee := summary.TotalRevenueShareCents
		net := summary.TotalNetAmountCents
		expectedNet := gross - fee
		residual := net - expectedNet
		// A month reconciles when recorded net matches gross − fee within 1% of gross
		// (absorbs cent-level rounding); an empty month is trivially reconciled.
		reconciled := gross <= 0 || abs64(residual) <= gross/100

		report.TotalGrossCents += gross
		report.TotalFeeCents += fee
		report.TotalNetCents += net
		if reconciled {
			report.MonthsReconciled++
		} else {
			report.MonthsFlagged++
		}
		report.Months = append(report.Months, reconMonthJSON{
			Month:            monthStart.Format("Jan"),
			GrossCents:       gross,
			FeeCents:         fee,
			NetCents:         net,
			ExpectedNetCents: expectedNet,
			ResidualCents:    residual,
			TxCount:          summary.TransactionCount,
			Reconciled:       reconciled,
		})
	}
	report.MonthsAudited = len(report.Months)
	report.ResidualCents = report.TotalNetCents - (report.TotalGrossCents - report.TotalFeeCents)
	report.Reconciled = report.MonthsFlagged == 0

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeReconCSV(w, report)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("ledger-reconciliation: encode report: %v", err)
	}
}

func writeReconCSV(w http.ResponseWriter, report reconReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="ledger-reconciliation.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"month", "gross_cents", "fee_cents", "net_cents", "expected_net_cents", "residual_cents", "tx_count", "reconciled"})
	for _, m := range report.Months {
		_ = cw.Write([]string{
			m.Month,
			strconv.FormatInt(m.GrossCents, 10),
			strconv.FormatInt(m.FeeCents, 10),
			strconv.FormatInt(m.NetCents, 10),
			strconv.FormatInt(m.ExpectedNetCents, 10),
			strconv.FormatInt(m.ResidualCents, 10),
			strconv.Itoa(m.TxCount),
			strconv.FormatBool(m.Reconciled),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("ledger-reconciliation: write CSV: %v", err)
	}
}

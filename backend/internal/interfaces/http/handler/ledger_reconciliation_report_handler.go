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
// (REPORTS.md — Guard): does the money add up? For each month it decomposes the gross the
// merchant paid into where it went — your net payout, Shopify's revenue-share cut, and the
// derived payment-processing deduction — and checks the identity closes:
//
//	gross = net + revenue_share + processing
//
// A residual is what's LEFT after those three buckets. Because processing is recovered per
// sale at sync time (gross − shopifyFee − net), a normal sales month closes to ~0; a
// residual now isolates refund/credit rows whose fee reversal never synced — a real gap,
// not the processing fee the old version wrongly flagged on every month.
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

// processingSuspectPct is the ceiling for a plausible payment-processing rate. Shopify's
// processing runs ~3% of gross; a derived rate well above that means the revenue-share cut
// was folded into processing because shopifyFee didn't sync for the month — the buckets
// still sum to gross (so the residual is ~0) but the decomposition can't be trusted. Set a
// generous margin above real processing yet far below any revenue-share tier (15/20%), so a
// legitimately 0%-share, ~3%-processing month never trips while an absorbed tier always does.
const processingSuspectPct = 6.0

type reconMonthJSON struct {
	Month             string  `json:"month"`
	GrossCents        int64   `json:"gross_cents"`
	NetCents          int64   `json:"net_cents"`
	RevenueShareCents int64   `json:"revenue_share_cents"` // Shopify's cut (shopifyFee)
	ProcessingCents   int64   `json:"processing_cents"`    // derived payment-processing deduction
	AccountedCents    int64   `json:"accounted_cents"`     // net + revenue_share + processing
	ProcessingPct     float64 `json:"processing_pct"`      // processing ÷ gross
	// ProcessingSuspect ⇒ processing_pct exceeds a plausible rate, so revenue-share data
	// was likely absorbed (unsynced shopifyFee). The month is NOT reconciled even though
	// its buckets sum to gross — this is what stops a silently-missing fee reading as clean.
	ProcessingSuspect bool `json:"processing_suspect"`
	// ResidualCents = gross − accounted. Non-zero ⇒ money the three buckets don't
	// explain — most often a refund/credit whose fee reversal hasn't synced.
	ResidualCents int64 `json:"residual_cents"`
	TxCount       int   `json:"tx_count"`
	Reconciled    bool  `json:"reconciled"`
}

type reconReport struct {
	Currency               string           `json:"currency"`
	TotalGrossCents        int64            `json:"total_gross_cents"`
	TotalNetCents          int64            `json:"total_net_cents"`
	TotalRevenueShareCents int64            `json:"total_revenue_share_cents"`
	TotalProcessingCents   int64            `json:"total_processing_cents"`
	ResidualCents          int64            `json:"residual_cents"`
	Reconciled             bool             `json:"reconciled"` // every month reconciled
	MonthsReconciled       int              `json:"months_reconciled"`
	MonthsFlagged          int              `json:"months_flagged"`
	MonthsAudited          int              `json:"months_audited"`
	Months                 []reconMonthJSON `json:"months"`
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
		net := summary.TotalNetAmountCents
		revShare := summary.TotalRevenueShareCents
		processing := summary.TotalProcessingFeeCents
		accounted := net + revShare + processing
		residual := gross - accounted

		var processingPct float64
		if gross > 0 {
			processingPct = float64(processing) / float64(gross) * 100
		}
		// An implausibly high processing rate means shopifyFee didn't sync and its cut got
		// absorbed into processing — the buckets still sum to gross, so guard on it explicitly
		// or a silently-missing fee would read as reconciled.
		processingSuspect := gross > 0 && processingPct > processingSuspectPct
		// A month reconciles when the three buckets account for gross within 1% of it
		// (absorbs cent-level rounding) AND the decomposition is trustworthy; an empty month
		// is trivially reconciled.
		bucketsClose := abs64(residual) <= gross/100
		reconciled := gross <= 0 || (bucketsClose && !processingSuspect)

		report.TotalGrossCents += gross
		report.TotalNetCents += net
		report.TotalRevenueShareCents += revShare
		report.TotalProcessingCents += processing
		if reconciled {
			report.MonthsReconciled++
		} else {
			report.MonthsFlagged++
		}
		report.Months = append(report.Months, reconMonthJSON{
			Month:             monthStart.Format("Jan"),
			GrossCents:        gross,
			NetCents:          net,
			RevenueShareCents: revShare,
			ProcessingCents:   processing,
			AccountedCents:    accounted,
			ProcessingPct:     processingPct,
			ProcessingSuspect: processingSuspect,
			ResidualCents:     residual,
			TxCount:           summary.TransactionCount,
			Reconciled:        reconciled,
		})
	}
	report.MonthsAudited = len(report.Months)
	report.ResidualCents = report.TotalGrossCents -
		(report.TotalNetCents + report.TotalRevenueShareCents + report.TotalProcessingCents)
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
	_ = cw.Write([]string{"month", "gross_cents", "net_cents", "revenue_share_cents", "processing_cents", "accounted_cents", "residual_cents", "processing_suspect", "tx_count", "reconciled"})
	for _, m := range report.Months {
		_ = cw.Write([]string{
			m.Month,
			strconv.FormatInt(m.GrossCents, 10),
			strconv.FormatInt(m.NetCents, 10),
			strconv.FormatInt(m.RevenueShareCents, 10),
			strconv.FormatInt(m.ProcessingCents, 10),
			strconv.FormatInt(m.AccountedCents, 10),
			strconv.FormatInt(m.ResidualCents, 10),
			strconv.FormatBool(m.ProcessingSuspect),
			strconv.Itoa(m.TxCount),
			strconv.FormatBool(m.Reconciled),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("ledger-reconciliation: write CSV: %v", err)
	}
}

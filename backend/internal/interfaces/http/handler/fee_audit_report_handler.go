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
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// FeeAuditReportHandler serves the "Fee Audit" report (REPORTS.md — Guard): does what
// Shopify actually retained match what the revenue-share tier says it should? It reuses
// the Fee Guard engine (buildFeeAudit) that also powers Profit & Expense, and adds the
// audit framing: a pass/flag verdict per month and the savings vs the default 20% plan.
type FeeAuditReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
	feeService  *service.FeeVerificationService
}

// NewFeeAuditReportHandler constructs a FeeAuditReportHandler.
func NewFeeAuditReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	feeService *service.FeeVerificationService,
) *FeeAuditReportHandler {
	return &FeeAuditReportHandler{txRepo: txRepo, appRepo: appRepo, partnerRepo: partnerRepo, feeService: feeService}
}

type feeAuditRowJSON struct {
	Month            string  `json:"month"`
	GrossCents       int64   `json:"gross_cents"`
	ShopifyCutCents  int64   `json:"shopify_cut_cents"`
	EffectiveFeePct  float64 `json:"effective_fee_pct"`
	ExpectedCutCents int64   `json:"expected_cut_cents"`
	FeeVarianceCents int64   `json:"fee_variance_cents"`
	FeeGuardOk       bool    `json:"fee_guard_ok"`
}

type feeAuditReport struct {
	Currency         string  `json:"currency"`
	ConfiguredTier   string  `json:"configured_tier"`
	ConfiguredFeePct float64 `json:"configured_fee_pct"`
	// DetectedFeePct is the app's real revenue-share rate derived from Shopify's
	// shopifyFee (RPT-FEES-2); TierMatches is whether the configured tier agrees.
	DetectedFeePct        float64           `json:"detected_fee_pct"`
	TierMatches           bool              `json:"tier_matches"`
	TotalGrossCents       int64             `json:"total_gross_cents"`
	TotalCutCents         int64             `json:"total_cut_cents"`
	EffectiveFeePct       float64           `json:"effective_fee_pct"`
	FlaggedMonths         int               `json:"flagged_months"`
	MonthsAudited         int               `json:"months_audited"`
	SavingsVsDefaultCents int64             `json:"savings_vs_default_cents"` // vs the 20% default plan
	Months                []feeAuditRowJSON `json:"months"`
}

// GetFeeAudit returns the Fee Audit report for an app.
// GET /api/v1/apps/{appID}/reports/fee-audit?months=6&format=csv
func (h *FeeAuditReportHandler) GetFeeAudit(w http.ResponseWriter, r *http.Request) {
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

	tier := app.RevenueShareTier
	audit := buildFeeAudit(r.Context(), h.txRepo, h.feeService, app.ID, tier, months, time.Now())

	var effectivePct float64
	if audit.TotalGrossCents > 0 {
		effectivePct = float64(audit.TotalCutCents) / float64(audit.TotalGrossCents) * 100
	}
	// Savings vs the default 20% plan = what 20% would have taken − what was actually
	// retained. Positive = the reduced-share plan is saving the developer money.
	defaultCut := int64(float64(audit.TotalGrossCents) * valueobject.RevenueShareTierDefault.RevenueSharePercent() / 100)
	savings := defaultCut - audit.TotalCutCents

	report := feeAuditReport{
		Currency:              "USD",
		ConfiguredTier:        tier.String(),
		ConfiguredFeePct:      tier.RevenueSharePercent(),
		DetectedFeePct:        audit.DetectedFeePct,
		TierMatches:           audit.TierMatches,
		TotalGrossCents:       audit.TotalGrossCents,
		TotalCutCents:         audit.TotalCutCents,
		EffectiveFeePct:       effectivePct,
		FlaggedMonths:         audit.flaggedMonths(),
		MonthsAudited:         len(audit.Months),
		SavingsVsDefaultCents: savings,
		Months:                make([]feeAuditRowJSON, 0, len(audit.Months)),
	}
	for _, m := range audit.Months {
		report.Months = append(report.Months, feeAuditRowJSON{
			Month:            m.Month,
			GrossCents:       m.GrossCents,
			ShopifyCutCents:  m.ShopifyCutCents,
			EffectiveFeePct:  m.EffectiveFeePct,
			ExpectedCutCents: m.ExpectedCutCents,
			FeeVarianceCents: m.FeeVarianceCents,
			FeeGuardOk:       m.FeeGuardOk,
		})
	}

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeFeeAuditCSV(w, report)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("fee-audit: encode report: %v", err)
	}
}

// writeFeeAuditCSV writes the per-month audit as a CSV attachment.
func writeFeeAuditCSV(w http.ResponseWriter, report feeAuditReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="fee-audit.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"month", "gross_cents", "shopify_cut_cents", "effective_fee_pct", "expected_cut_cents", "fee_variance_cents", "fee_guard_ok"})
	for _, m := range report.Months {
		_ = cw.Write([]string{
			m.Month,
			strconv.FormatInt(m.GrossCents, 10),
			strconv.FormatInt(m.ShopifyCutCents, 10),
			strconv.FormatFloat(m.EffectiveFeePct, 'f', 2, 64),
			strconv.FormatInt(m.ExpectedCutCents, 10),
			strconv.FormatInt(m.FeeVarianceCents, 10),
			strconv.FormatBool(m.FeeGuardOk),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("fee-audit: write CSV: %v", err)
	}
}

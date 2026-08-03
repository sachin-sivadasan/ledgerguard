package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type FeeHandler struct {
	appRepo         repository.AppRepository
	partnerRepo     repository.PartnerAccountRepository
	transactionRepo repository.TransactionRepository
	feeService      *service.FeeVerificationService
}

func NewFeeHandler(
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	transactionRepo repository.TransactionRepository,
	feeService *service.FeeVerificationService,
) *FeeHandler {
	return &FeeHandler{
		appRepo:         appRepo,
		partnerRepo:     partnerRepo,
		transactionRepo: transactionRepo,
		feeService:      feeService,
	}
}

// GetFeeSummary returns aggregated fee information for an app
// GET /api/v1/apps/{appID}/fees/summary?start=YYYY-MM-DD&end=YYYY-MM-DD
// appID is numeric Shopify app ID (e.g., "4599915")
func (h *FeeHandler) GetFeeSummary(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeFeeError(w, lookupErr.statusCode, lookupErr.message)
		return
	}
	tier := app.RevenueShareTier

	// Parse date range (default to last 30 days)
	now := time.Now()
	start := now.AddDate(0, -1, 0) // 30 days ago
	end := now

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsed
		}
	}
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsed.Add(24*time.Hour - time.Second) // End of day
		}
	}

	// Get transactions
	transactions, err2 := h.transactionRepo.FindByAppID(r.Context(), app.ID, start, end)
	if err2 != nil {
		writeFeeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	// Calculate fee summary
	summary := h.feeService.CalculateFeeSummary(transactions)

	// Calculate tier savings
	savings := h.feeService.CalculateTierSavings(summary.TotalGrossAmountCents, tier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"period": map[string]string{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		},
		"tier": map[string]interface{}{
			"code":               tier.String(),
			"display_name":       tier.DisplayName(),
			"description":        tier.Description(),
			"revenue_share_pct":  tier.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":    tier.IsReducedPlan(),
		},
		"summary": map[string]interface{}{
			"transaction_count":          summary.TransactionCount,
			"total_gross_cents":          summary.TotalGrossAmountCents,
			"total_revenue_share_cents":  summary.TotalRevenueShareCents,
			"total_processing_fee_cents": summary.TotalProcessingFeeCents,
			"total_tax_on_fees_cents":    summary.TotalTaxOnFeesCents,
			"total_fees_cents":           summary.TotalFeesCents,
			"total_net_cents":            summary.TotalNetAmountCents,
			"avg_revenue_share_pct":      summary.AverageRevenueSharePct,
			"avg_processing_fee_pct":     summary.AverageProcessingFeePct,
			"effective_fee_pct":          summary.EffectiveFeePercent,
		},
		"savings": map[string]interface{}{
			"compared_to":        "DEFAULT_20",
			"default_fees_cents": savings.DefaultTierFeesCents,
			"current_fees_cents": savings.CurrentTierFeesCents,
			"savings_cents":      savings.SavingsCents,
			"savings_pct":        savings.SavingsPercent,
		},
	})
}

func writeFeeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    http.StatusText(statusCode),
			"message": message,
		},
	})
}

// GetTierBreakdown returns the fee breakdown for a hypothetical amount
// GET /api/v1/apps/{appID}/fees/breakdown?amount_cents=4900
// appID is numeric Shopify app ID (e.g., "4599915")
func (h *FeeHandler) GetTierBreakdown(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeFeeError(w, lookupErr.statusCode, lookupErr.message)
		return
	}
	currentTier := &app.RevenueShareTier

	// Parse amount (default to $49.00)
	amountCents := int64(4900)
	if amountStr := r.URL.Query().Get("amount_cents"); amountStr != "" {
		if _, err := json.Number(amountStr).Int64(); err == nil {
			amountCents, _ = json.Number(amountStr).Int64()
		}
	}

	// Tax rate (default 8%)
	taxRate := 0.08

	// Calculate breakdowns for all tiers
	tiers := []valueobject.RevenueShareTier{
		valueobject.RevenueShareTierDefault,
		valueobject.RevenueShareTierSmallDev0,
		valueobject.RevenueShareTierSmallDev15,
		valueobject.RevenueShareTierLargeDev,
	}

	breakdowns := make([]map[string]interface{}, len(tiers))
	for i, tier := range tiers {
		breakdown := tier.CalculateFeeBreakdown(amountCents, taxRate)
		breakdowns[i] = map[string]interface{}{
			"tier":                 tier.String(),
			"tier_display_name":    tier.DisplayName(),
			"is_current":           tier == *currentTier,
			"gross_cents":          breakdown.GrossAmountCents,
			"revenue_share_cents":  breakdown.RevenueShareCents,
			"processing_fee_cents": breakdown.ProcessingFeeCents,
			"tax_on_fees_cents":    breakdown.TaxOnFeesCents,
			"total_fees_cents":     breakdown.TotalFeesCents,
			"net_cents":            breakdown.NetAmountCents,
			"revenue_share_pct":    breakdown.RevenueSharePercent,
			"processing_fee_pct":   breakdown.ProcessingFeePercent,
		}
	}

	currentBreakdown := currentTier.CalculateFeeBreakdown(amountCents, taxRate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"amount_cents": amountCents,
		"tax_rate":     taxRate,
		"current_tier": map[string]interface{}{
			"code":                 currentTier.String(),
			"display_name":         currentTier.DisplayName(),
			"gross_cents":          currentBreakdown.GrossAmountCents,
			"revenue_share_cents":  currentBreakdown.RevenueShareCents,
			"processing_fee_cents": currentBreakdown.ProcessingFeeCents,
			"tax_on_fees_cents":    currentBreakdown.TaxOnFeesCents,
			"total_fees_cents":     currentBreakdown.TotalFeesCents,
			"net_cents":            currentBreakdown.NetAmountCents,
		},
		"all_tiers": breakdowns,
	})
}

// GetMonthlyProfitBreakdown returns monthly P&L breakdown for the app.
// GET /api/v1/apps/{appID}/fees/monthly?months=6
func (h *FeeHandler) GetMonthlyProfitBreakdown(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeFeeError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	months := 6
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := json.Number(m).Int64(); err == nil && parsed >= 1 && parsed <= 24 {
			months = int(parsed)
		}
	}

	now := time.Now()
	tier := app.RevenueShareTier
	type monthBreakdown struct {
		Month           string  `json:"month"`
		GrossCents      int64   `json:"gross_cents"`
		ShopifyCutCents int64   `json:"shopify_cut_cents"` // actual, from Shopify's shopifyFee
		NetCents        int64   `json:"net_cents"`
		ProfitMarginPct float64 `json:"profit_margin_pct"`
		EffectiveFeePct float64 `json:"effective_fee_pct"` // actual cut ÷ gross
		// Fee Guard: expected cut = gross × the app's revenue-share tier %; variance
		// = actual − expected. FeeGuardOk is false when |variance| exceeds 1% of gross
		// (Shopify may have charged the wrong rate — the "Guard" differentiator).
		ExpectedCutCents int64 `json:"expected_cut_cents"`
		FeeVarianceCents int64 `json:"fee_variance_cents"`
		FeeGuardOk       bool  `json:"fee_guard_ok"`
		// Deprecated: Shopify's Partner API does not report processing fee / tax
		// separately (shopifyFee is the total retained). Kept at 0 for compatibility.
		ProcessingFeeCents int64 `json:"processing_fee_cents"`
		TaxCents           int64 `json:"tax_cents"`
	}

	result := make([]monthBreakdown, 0, months)

	// Pass 1: per-month actuals + running totals (to derive the app's real fee rate).
	var totalGross, totalCut int64
	for i := months - 1; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := time.Date(now.Year(), now.Month()-time.Month(i)+1, 0, 23, 59, 59, 0, time.UTC)

		transactions, err := h.transactionRepo.FindByAppID(r.Context(), app.ID, monthStart, monthEnd)
		if err != nil {
			continue
		}

		summary := h.feeService.CalculateFeeSummary(transactions)
		gross := summary.TotalGrossAmountCents
		actualCut := summary.TotalRevenueShareCents
		totalGross += gross
		totalCut += actualCut

		var margin, effectivePct float64
		if gross > 0 {
			margin = float64(summary.TotalNetAmountCents) / float64(gross) * 100
			effectivePct = float64(actualCut) / float64(gross) * 100
		}
		result = append(result, monthBreakdown{
			Month:           monthStart.Format("Jan"),
			GrossCents:      gross,
			ShopifyCutCents: actualCut,
			NetCents:        summary.TotalNetAmountCents,
			ProfitMarginPct: margin,
			EffectiveFeePct: effectivePct,
		})
	}

	// Derive the app's actual revenue-share rate from the observed cut, snapped to the
	// nearest real Shopify tier (0/15/20%). The Fee Guard compares each month against
	// THIS — not the configured tier, which defaults to 0% and is wrong for apps past
	// $1M (RPT-FEES-2) — so it flags true anomalies (e.g. a tier-transition month or a
	// mischarge) rather than every month.
	detectedPct := tier.RevenueSharePercent()
	if totalGross > 0 {
		detectedPct = snapToTierRate(float64(totalCut) / float64(totalGross) * 100)
	}
	for idx := range result {
		gross := result[idx].GrossCents
		expected := int64(float64(gross) * detectedPct / 100)
		variance := result[idx].ShopifyCutCents - expected
		result[idx].ExpectedCutCents = expected
		result[idx].FeeVarianceCents = variance
		result[idx].FeeGuardOk = gross <= 0 || abs64(variance) <= gross/100
	}

	// tierMatches surfaces RPT-FEES-2 directly: is the app's CONFIGURED tier consistent
	// with what Shopify actually retains?
	tierMatches := abs64Float(tier.RevenueSharePercent()-detectedPct) < 1.0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"configured_tier":    tier.String(),
		"configured_fee_pct": tier.RevenueSharePercent(),
		"detected_fee_pct":   detectedPct, // derived from actual shopifyFee/gross
		"tier_matches":       tierMatches,
		"expected_fee_pct":   detectedPct, // what the Fee Guard compares against
		"months":             result,
	})
}

// knownTierRates are the Shopify app revenue-share rates a paying app can be on.
var knownTierRates = []float64{0, 15, 20}

// snapToTierRate returns the known Shopify tier rate nearest the observed effective rate,
// so a small rounding/mix in the data resolves to a real tier.
func snapToTierRate(effectivePct float64) float64 {
	best := knownTierRates[0]
	for _, rate := range knownTierRates[1:] {
		if abs64Float(effectivePct-rate) < abs64Float(effectivePct-best) {
			best = rate
		}
	}
	return best
}

// abs64Float returns the absolute value of a float64.
func abs64Float(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// abs64 returns the absolute value of an int64.
func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

// ListAvailableTiers returns all available revenue share tiers
// GET /api/v1/tiers
func (h *FeeHandler) ListAvailableTiers(w http.ResponseWriter, r *http.Request) {
	tiers := []map[string]interface{}{
		{
			"code":               valueobject.RevenueShareTierDefault.String(),
			"display_name":       valueobject.RevenueShareTierDefault.DisplayName(),
			"description":        valueobject.RevenueShareTierDefault.Description(),
			"revenue_share_pct":  valueobject.RevenueShareTierDefault.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":    valueobject.RevenueShareTierDefault.IsReducedPlan(),
		},
		{
			"code":               valueobject.RevenueShareTierSmallDev0.String(),
			"display_name":       valueobject.RevenueShareTierSmallDev0.DisplayName(),
			"description":        valueobject.RevenueShareTierSmallDev0.Description(),
			"revenue_share_pct":  valueobject.RevenueShareTierSmallDev0.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":    valueobject.RevenueShareTierSmallDev0.IsReducedPlan(),
		},
		{
			"code":               valueobject.RevenueShareTierSmallDev15.String(),
			"display_name":       valueobject.RevenueShareTierSmallDev15.DisplayName(),
			"description":        valueobject.RevenueShareTierSmallDev15.Description(),
			"revenue_share_pct":  valueobject.RevenueShareTierSmallDev15.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":    valueobject.RevenueShareTierSmallDev15.IsReducedPlan(),
		},
		{
			"code":               valueobject.RevenueShareTierLargeDev.String(),
			"display_name":       valueobject.RevenueShareTierLargeDev.DisplayName(),
			"description":        valueobject.RevenueShareTierLargeDev.Description(),
			"revenue_share_pct":  valueobject.RevenueShareTierLargeDev.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":    valueobject.RevenueShareTierLargeDev.IsReducedPlan(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
	})
}

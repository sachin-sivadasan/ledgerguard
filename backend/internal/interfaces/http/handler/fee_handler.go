package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type FeeHandler struct {
	appRepo         repository.AppRepository
	transactionRepo repository.TransactionRepository
	feeService      *service.FeeVerificationService
}

func NewFeeHandler(
	appRepo repository.AppRepository,
	transactionRepo repository.TransactionRepository,
	feeService *service.FeeVerificationService,
) *FeeHandler {
	return &FeeHandler{
		appRepo:         appRepo,
		transactionRepo: transactionRepo,
		feeService:      feeService,
	}
}

// GetFeeSummary returns aggregated fee information for an app
// GET /api/v1/apps/{appID}/fees/summary?start=YYYY-MM-DD&end=YYYY-MM-DD
func (h *FeeHandler) GetFeeSummary(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appIDStr := chi.URLParam(r, "appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid app_id format")
		return
	}

	// Get app to determine tier
	app, err := h.appRepo.FindByID(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

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
	transactions, err := h.transactionRepo.FindByAppID(r.Context(), appID, start, end)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	// Calculate fee summary
	summary := h.feeService.CalculateFeeSummary(transactions)

	// Calculate tier savings
	savings := h.feeService.CalculateTierSavings(summary.TotalGrossAmountCents, app.RevenueShareTier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"period": map[string]string{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		},
		"tier": map[string]interface{}{
			"code":                app.RevenueShareTier.String(),
			"display_name":        app.RevenueShareTier.DisplayName(),
			"description":         app.RevenueShareTier.Description(),
			"revenue_share_pct":   app.RevenueShareTier.RevenueSharePercent(),
			"processing_fee_pct":  valueobject.ProcessingFeePercent,
			"is_reduced_plan":     app.RevenueShareTier.IsReducedPlan(),
		},
		"summary": map[string]interface{}{
			"transaction_count":       summary.TransactionCount,
			"total_gross_cents":       summary.TotalGrossAmountCents,
			"total_revenue_share_cents": summary.TotalRevenueShareCents,
			"total_processing_fee_cents": summary.TotalProcessingFeeCents,
			"total_tax_on_fees_cents":   summary.TotalTaxOnFeesCents,
			"total_fees_cents":          summary.TotalFeesCents,
			"total_net_cents":           summary.TotalNetAmountCents,
			"avg_revenue_share_pct":     summary.AverageRevenueSharePct,
			"avg_processing_fee_pct":    summary.AverageProcessingFeePct,
			"effective_fee_pct":         summary.EffectiveFeePercent,
		},
		"savings": map[string]interface{}{
			"compared_to":            "DEFAULT_20",
			"default_fees_cents":     savings.DefaultTierFeesCents,
			"current_fees_cents":     savings.CurrentTierFeesCents,
			"savings_cents":          savings.SavingsCents,
			"savings_pct":            savings.SavingsPercent,
		},
	})
}

// GetTierBreakdown returns the fee breakdown for a hypothetical amount
// GET /api/v1/apps/{appID}/fees/breakdown?amount_cents=4900
func (h *FeeHandler) GetTierBreakdown(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	appIDStr := chi.URLParam(r, "appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid app_id format")
		return
	}

	// Get app to determine tier
	app, err := h.appRepo.FindByID(r.Context(), appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

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
			"is_current":           tier == app.RevenueShareTier,
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

	currentBreakdown := app.RevenueShareTier.CalculateFeeBreakdown(amountCents, taxRate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"amount_cents": amountCents,
		"tax_rate":     taxRate,
		"current_tier": map[string]interface{}{
			"code":                 app.RevenueShareTier.String(),
			"display_name":         app.RevenueShareTier.DisplayName(),
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

// ListAvailableTiers returns all available revenue share tiers
// GET /api/v1/tiers
func (h *FeeHandler) ListAvailableTiers(w http.ResponseWriter, r *http.Request) {
	tiers := []map[string]interface{}{
		{
			"code":              valueobject.RevenueShareTierDefault.String(),
			"display_name":      valueobject.RevenueShareTierDefault.DisplayName(),
			"description":       valueobject.RevenueShareTierDefault.Description(),
			"revenue_share_pct": valueobject.RevenueShareTierDefault.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":   valueobject.RevenueShareTierDefault.IsReducedPlan(),
		},
		{
			"code":              valueobject.RevenueShareTierSmallDev0.String(),
			"display_name":      valueobject.RevenueShareTierSmallDev0.DisplayName(),
			"description":       valueobject.RevenueShareTierSmallDev0.Description(),
			"revenue_share_pct": valueobject.RevenueShareTierSmallDev0.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":   valueobject.RevenueShareTierSmallDev0.IsReducedPlan(),
		},
		{
			"code":              valueobject.RevenueShareTierSmallDev15.String(),
			"display_name":      valueobject.RevenueShareTierSmallDev15.DisplayName(),
			"description":       valueobject.RevenueShareTierSmallDev15.Description(),
			"revenue_share_pct": valueobject.RevenueShareTierSmallDev15.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":   valueobject.RevenueShareTierSmallDev15.IsReducedPlan(),
		},
		{
			"code":              valueobject.RevenueShareTierLargeDev.String(),
			"display_name":      valueobject.RevenueShareTierLargeDev.DisplayName(),
			"description":       valueobject.RevenueShareTierLargeDev.Description(),
			"revenue_share_pct": valueobject.RevenueShareTierLargeDev.RevenueSharePercent(),
			"processing_fee_pct": valueobject.ProcessingFeePercent,
			"is_reduced_plan":   valueobject.RevenueShareTierLargeDev.IsReducedPlan(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
	})
}

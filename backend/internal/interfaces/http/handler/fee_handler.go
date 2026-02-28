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

const feeAppGIDPrefix = "gid://partners/App/"

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

// getAppFromRequest resolves app from numeric Shopify app ID
func (h *FeeHandler) getAppFromRequest(r *http.Request) (*uuid.UUID, *valueobject.RevenueShareTier, error) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		return nil, nil, &feeError{http.StatusUnauthorized, "authentication required"}
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		return nil, nil, &feeError{http.StatusNotFound, "no partner account found"}
	}

	// Get numeric appID from URL and construct full GID
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		return nil, nil, &feeError{http.StatusBadRequest, "app ID is required"}
	}
	fullAppGID := feeAppGIDPrefix + appIDStr

	// Find app by partner app ID (GID)
	app, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		return nil, nil, &feeError{http.StatusNotFound, "app not found"}
	}

	return &app.ID, &app.RevenueShareTier, nil
}

type feeError struct {
	statusCode int
	message    string
}

func (e *feeError) Error() string {
	return e.message
}

// GetFeeSummary returns aggregated fee information for an app
// GET /api/v1/apps/{appID}/fees/summary?start=YYYY-MM-DD&end=YYYY-MM-DD
// appID is numeric Shopify app ID (e.g., "4599915")
func (h *FeeHandler) GetFeeSummary(w http.ResponseWriter, r *http.Request) {
	appID, tier, err := h.getAppFromRequest(r)
	if err != nil {
		if fe, ok := err.(*feeError); ok {
			writeFeeError(w, fe.statusCode, fe.message)
		} else {
			writeFeeError(w, http.StatusInternalServerError, "internal error")
		}
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
	transactions, err2 := h.transactionRepo.FindByAppID(r.Context(), *appID, start, end)
	if err2 != nil {
		writeFeeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	// Calculate fee summary
	summary := h.feeService.CalculateFeeSummary(transactions)

	// Calculate tier savings
	savings := h.feeService.CalculateTierSavings(summary.TotalGrossAmountCents, *tier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"period": map[string]string{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		},
		"tier": map[string]interface{}{
			"code":                tier.String(),
			"display_name":        tier.DisplayName(),
			"description":         tier.Description(),
			"revenue_share_pct":   tier.RevenueSharePercent(),
			"processing_fee_pct":  valueobject.ProcessingFeePercent,
			"is_reduced_plan":     tier.IsReducedPlan(),
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
	_, currentTier, err := h.getAppFromRequest(r)
	if err != nil {
		if fe, ok := err.(*feeError); ok {
			writeFeeError(w, fe.statusCode, fe.message)
		} else {
			writeFeeError(w, http.StatusInternalServerError, "internal error")
		}
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

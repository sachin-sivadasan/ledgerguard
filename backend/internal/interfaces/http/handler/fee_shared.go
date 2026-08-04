package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// feeAuditMonth is one month of the fee audit: what Shopify actually retained vs
// what the app's (detected) revenue-share tier says it should have.
type feeAuditMonth struct {
	Month            string
	GrossCents       int64
	ShopifyCutCents  int64 // actual, from Shopify's shopifyFee
	NetCents         int64
	ProfitMarginPct  float64
	EffectiveFeePct  float64 // actual cut ÷ gross
	ExpectedCutCents int64   // gross × the DETECTED tier rate
	FeeVarianceCents int64   // actual − expected
	FeeGuardOk       bool    // |variance| ≤ 1% of gross
}

// feeAuditResult is the computed monthly fee audit + the derived tier verdict.
type feeAuditResult struct {
	DetectedFeePct  float64 // real rate from the data, snapped to a Shopify tier
	TierMatches     bool    // configured tier ≈ detected rate
	TotalGrossCents int64
	TotalCutCents   int64
	Months          []feeAuditMonth
}

// buildFeeAudit is the single source of truth behind both Profit & Expense
// (/fees/monthly) and the Fee Audit report (/reports/fee-audit): it buckets
// transactions by calendar month, derives the app's REAL revenue-share rate from
// the observed shopifyFee/gross (snapped to a Shopify tier — see RPT-FEES-2, since
// the configured tier defaults to 0% and is wrong past $1M), and flags each month
// whose actual cut deviates from that detected rate.
func buildFeeAudit(
	ctx context.Context,
	txRepo repository.TransactionRepository,
	feeService *service.FeeVerificationService,
	appID uuid.UUID,
	tier valueobject.RevenueShareTier,
	months int,
	now time.Time,
) feeAuditResult {
	var res feeAuditResult
	res.Months = make([]feeAuditMonth, 0, months)

	for i := months - 1; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := time.Date(now.Year(), now.Month()-time.Month(i)+1, 0, 23, 59, 59, 0, time.UTC)

		transactions, err := txRepo.FindByAppID(ctx, appID, monthStart, monthEnd)
		if err != nil {
			continue
		}

		summary := feeService.CalculateFeeSummary(transactions)
		gross := summary.TotalGrossAmountCents
		cut := summary.TotalRevenueShareCents
		res.TotalGrossCents += gross
		res.TotalCutCents += cut

		var margin, effectivePct float64
		if gross > 0 {
			margin = float64(summary.TotalNetAmountCents) / float64(gross) * 100
			effectivePct = float64(cut) / float64(gross) * 100
		}
		res.Months = append(res.Months, feeAuditMonth{
			Month:           monthStart.Format("Jan"),
			GrossCents:      gross,
			ShopifyCutCents: cut,
			NetCents:        summary.TotalNetAmountCents,
			ProfitMarginPct: margin,
			EffectiveFeePct: effectivePct,
		})
	}

	// Evaluate each month against ITS OWN nearest Shopify tier (0/15/20%), not one
	// window-wide rate. Shopify's reduced-share plan flips 0%→15% at $1M lifetime, so a
	// window spanning that crossing has a misleading blended rate (e.g. 5% → would snap
	// to 0% and wrongly flag every paying month). Per-month snapping flags only the
	// months that DON'T sit cleanly on a tier — a transition month or a real mischarge.
	for idx := range res.Months {
		gross := res.Months[idx].GrossCents
		expectedPct := snapToTierRate(res.Months[idx].EffectiveFeePct)
		expected := int64(float64(gross) * expectedPct / 100)
		variance := res.Months[idx].ShopifyCutCents - expected
		res.Months[idx].ExpectedCutCents = expected
		res.Months[idx].FeeVarianceCents = variance
		res.Months[idx].FeeGuardOk = gross <= 0 || abs64(variance) <= gross/100
	}

	// The app's CURRENT tier = the most recent month with revenue, snapped. Tiers only
	// ratchet UP with lifetime earnings, so the latest month reflects today's rate.
	res.DetectedFeePct = tier.RevenueSharePercent()
	for i := len(res.Months) - 1; i >= 0; i-- {
		if res.Months[i].GrossCents > 0 {
			res.DetectedFeePct = snapToTierRate(res.Months[i].EffectiveFeePct)
			break
		}
	}
	res.TierMatches = abs64Float(tier.RevenueSharePercent()-res.DetectedFeePct) < 1.0
	return res
}

// flaggedMonths counts the months whose Fee Guard tripped (a real fee mismatch).
func (r feeAuditResult) flaggedMonths() int {
	n := 0
	for _, m := range r.Months {
		if !m.FeeGuardOk {
			n++
		}
	}
	return n
}

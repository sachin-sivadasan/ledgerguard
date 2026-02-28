package service

import (
	"math"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// FeeVerificationService verifies transaction fees against expected tier-based calculations
type FeeVerificationService struct{}

// NewFeeVerificationService creates a new fee verification service
func NewFeeVerificationService() *FeeVerificationService {
	return &FeeVerificationService{}
}

// FeeVerificationResult contains the result of fee verification for a transaction
type FeeVerificationResult struct {
	Transaction *entity.Transaction
	Tier        valueobject.RevenueShareTier

	// Expected fees based on tier
	ExpectedRevenueShareCents  int64
	ExpectedProcessingFeeCents int64
	ExpectedTotalFeesCents     int64
	ExpectedNetAmountCents     int64

	// Actual fees from Shopify
	ActualRevenueShareCents  int64
	ActualProcessingFeeCents int64
	ActualTotalFeesCents     int64
	ActualNetAmountCents     int64

	// Discrepancies
	RevenueShareDiscrepancyCents  int64
	ProcessingFeeDiscrepancyCents int64
	TotalFeeDiscrepancyCents      int64
	NetAmountDiscrepancyCents     int64

	// Status
	IsVerified         bool // True if actual matches expected (within tolerance)
	DiscrepancyPercent float64
}

// VerifyTransaction verifies a single transaction's fees against expected tier-based calculations
func (s *FeeVerificationService) VerifyTransaction(
	tx *entity.Transaction,
	tier valueobject.RevenueShareTier,
	tolerancePercent float64, // e.g., 0.01 for 1% tolerance
) *FeeVerificationResult {
	// Calculate expected fees based on tier
	expected := tier.CalculateFeeBreakdown(tx.GrossAmountCents, 0) // Tax rate is variable, so we don't include it

	result := &FeeVerificationResult{
		Transaction: tx,
		Tier:        tier,

		// Expected (excluding tax, as tax is variable)
		ExpectedRevenueShareCents:  expected.RevenueShareCents,
		ExpectedProcessingFeeCents: expected.ProcessingFeeCents,
		ExpectedTotalFeesCents:     expected.RevenueShareCents + expected.ProcessingFeeCents,
		ExpectedNetAmountCents:     tx.GrossAmountCents - expected.RevenueShareCents - expected.ProcessingFeeCents,

		// Actual from Shopify
		ActualRevenueShareCents:  tx.ShopifyFeeCents,
		ActualProcessingFeeCents: tx.ProcessingFeeCents,
		ActualTotalFeesCents:     tx.TotalFeesCents(),
		ActualNetAmountCents:     tx.NetAmountCents,
	}

	// Calculate discrepancies
	result.RevenueShareDiscrepancyCents = result.ActualRevenueShareCents - result.ExpectedRevenueShareCents
	result.ProcessingFeeDiscrepancyCents = result.ActualProcessingFeeCents - result.ExpectedProcessingFeeCents
	result.TotalFeeDiscrepancyCents = result.ActualTotalFeesCents - result.ExpectedTotalFeesCents
	result.NetAmountDiscrepancyCents = result.ActualNetAmountCents - result.ExpectedNetAmountCents

	// Calculate discrepancy percentage (based on gross amount)
	if tx.GrossAmountCents > 0 {
		result.DiscrepancyPercent = math.Abs(float64(result.TotalFeeDiscrepancyCents)) / float64(tx.GrossAmountCents) * 100
	}

	// Determine if verified (within tolerance)
	// Note: We allow some tolerance because tax is variable and not included in expected
	toleranceAmount := int64(float64(tx.GrossAmountCents) * tolerancePercent)
	result.IsVerified = math.Abs(float64(result.RevenueShareDiscrepancyCents)) <= float64(toleranceAmount) &&
		math.Abs(float64(result.ProcessingFeeDiscrepancyCents)) <= float64(toleranceAmount)

	return result
}

// FeeSummary contains aggregated fee information
type FeeSummary struct {
	TotalGrossAmountCents     int64
	TotalRevenueShareCents    int64
	TotalProcessingFeeCents   int64
	TotalTaxOnFeesCents       int64
	TotalFeesCents            int64
	TotalNetAmountCents       int64
	TransactionCount          int
	AverageRevenueSharePct    float64
	AverageProcessingFeePct   float64
	EffectiveFeePercent       float64 // Total fees as % of gross
}

// CalculateFeeSummary calculates aggregated fee information for a list of transactions
func (s *FeeVerificationService) CalculateFeeSummary(transactions []*entity.Transaction) *FeeSummary {
	summary := &FeeSummary{
		TransactionCount: len(transactions),
	}

	for _, tx := range transactions {
		summary.TotalGrossAmountCents += tx.GrossAmountCents
		summary.TotalRevenueShareCents += tx.ShopifyFeeCents
		summary.TotalProcessingFeeCents += tx.ProcessingFeeCents
		summary.TotalTaxOnFeesCents += tx.TaxOnFeesCents
		summary.TotalNetAmountCents += tx.NetAmountCents
	}

	summary.TotalFeesCents = summary.TotalRevenueShareCents +
		summary.TotalProcessingFeeCents +
		summary.TotalTaxOnFeesCents

	// Calculate percentages
	if summary.TotalGrossAmountCents > 0 {
		summary.AverageRevenueSharePct = float64(summary.TotalRevenueShareCents) / float64(summary.TotalGrossAmountCents) * 100
		summary.AverageProcessingFeePct = float64(summary.TotalProcessingFeeCents) / float64(summary.TotalGrossAmountCents) * 100
		summary.EffectiveFeePercent = float64(summary.TotalFeesCents) / float64(summary.TotalGrossAmountCents) * 100
	}

	return summary
}

// TierSavingsResult contains the savings from using a reduced tier vs default
type TierSavingsResult struct {
	CurrentTier          valueobject.RevenueShareTier
	DefaultTierFeesCents int64 // What fees would be at 20%
	CurrentTierFeesCents int64 // What fees actually are
	SavingsCents         int64 // Difference (positive = savings)
	SavingsPercent       float64
}

// CalculateTierSavings calculates how much the developer is saving by using a reduced tier
func (s *FeeVerificationService) CalculateTierSavings(
	grossAmountCents int64,
	currentTier valueobject.RevenueShareTier,
) *TierSavingsResult {
	defaultFees := valueobject.RevenueShareTierDefault.CalculateFeeBreakdown(grossAmountCents, 0)
	currentFees := currentTier.CalculateFeeBreakdown(grossAmountCents, 0)

	result := &TierSavingsResult{
		CurrentTier:          currentTier,
		DefaultTierFeesCents: defaultFees.TotalFeesCents,
		CurrentTierFeesCents: currentFees.TotalFeesCents,
		SavingsCents:         defaultFees.TotalFeesCents - currentFees.TotalFeesCents,
	}

	if defaultFees.TotalFeesCents > 0 {
		result.SavingsPercent = float64(result.SavingsCents) / float64(defaultFees.TotalFeesCents) * 100
	}

	return result
}

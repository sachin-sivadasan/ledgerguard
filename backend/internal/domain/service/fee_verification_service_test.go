package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestFeeVerificationService_VerifyTransaction_Default20(t *testing.T) {
	svc := NewFeeVerificationService()

	// $49 gross, DEFAULT_20 tier
	// Expected: 20% revenue share = $9.80, 2.9% processing = $1.42
	tx := &entity.Transaction{
		ID:                 uuid.New(),
		GrossAmountCents:   4900,
		ShopifyFeeCents:    980,  // 20% of 4900
		ProcessingFeeCents: 142,  // 2.9% of 4900
		TaxOnFeesCents:     90,   // ~8% of fees
		NetAmountCents:     3688, // 4900 - 980 - 142 - 90
	}

	result := svc.VerifyTransaction(tx, valueobject.RevenueShareTierDefault, 0.01)

	if result.ExpectedRevenueShareCents != 980 {
		t.Errorf("ExpectedRevenueShareCents = %d, want 980", result.ExpectedRevenueShareCents)
	}

	if result.ExpectedProcessingFeeCents != 142 {
		t.Errorf("ExpectedProcessingFeeCents = %d, want 142", result.ExpectedProcessingFeeCents)
	}

	if result.RevenueShareDiscrepancyCents != 0 {
		t.Errorf("RevenueShareDiscrepancyCents = %d, want 0", result.RevenueShareDiscrepancyCents)
	}

	if !result.IsVerified {
		t.Error("Expected transaction to be verified")
	}
}

func TestFeeVerificationService_VerifyTransaction_SmallDev0(t *testing.T) {
	svc := NewFeeVerificationService()

	// $49 gross, SMALL_DEV_0 tier (0% revenue share)
	// Expected: 0% revenue share = $0, 2.9% processing = $1.42
	tx := &entity.Transaction{
		ID:                 uuid.New(),
		GrossAmountCents:   4900,
		ShopifyFeeCents:    0,    // 0% of 4900
		ProcessingFeeCents: 142,  // 2.9% of 4900
		TaxOnFeesCents:     11,   // ~8% of fees
		NetAmountCents:     4747, // 4900 - 0 - 142 - 11
	}

	result := svc.VerifyTransaction(tx, valueobject.RevenueShareTierSmallDev0, 0.01)

	if result.ExpectedRevenueShareCents != 0 {
		t.Errorf("ExpectedRevenueShareCents = %d, want 0", result.ExpectedRevenueShareCents)
	}

	if !result.IsVerified {
		t.Error("Expected transaction to be verified")
	}
}

func TestFeeVerificationService_VerifyTransaction_Discrepancy(t *testing.T) {
	svc := NewFeeVerificationService()

	// Transaction with wrong fees (claiming 0% but should be 20%)
	tx := &entity.Transaction{
		ID:                 uuid.New(),
		GrossAmountCents:   4900,
		ShopifyFeeCents:    0,   // Should be 980 for DEFAULT_20
		ProcessingFeeCents: 142,
		TaxOnFeesCents:     0,
		NetAmountCents:     4758,
	}

	result := svc.VerifyTransaction(tx, valueobject.RevenueShareTierDefault, 0.01)

	// Should not be verified because revenue share doesn't match
	if result.IsVerified {
		t.Error("Expected transaction to NOT be verified due to fee discrepancy")
	}

	// Should have negative discrepancy (actual < expected)
	if result.RevenueShareDiscrepancyCents != -980 {
		t.Errorf("RevenueShareDiscrepancyCents = %d, want -980", result.RevenueShareDiscrepancyCents)
	}
}

func TestFeeVerificationService_CalculateFeeSummary(t *testing.T) {
	svc := NewFeeVerificationService()

	transactions := []*entity.Transaction{
		{
			GrossAmountCents:   4900,
			ShopifyFeeCents:    980,
			ProcessingFeeCents: 142,
			TaxOnFeesCents:     90,
			NetAmountCents:     3688,
		},
		{
			GrossAmountCents:   2900,
			ShopifyFeeCents:    580,
			ProcessingFeeCents: 84,
			TaxOnFeesCents:     53,
			NetAmountCents:     2183,
		},
	}

	summary := svc.CalculateFeeSummary(transactions)

	if summary.TransactionCount != 2 {
		t.Errorf("TransactionCount = %d, want 2", summary.TransactionCount)
	}

	expectedGross := int64(7800) // 4900 + 2900
	if summary.TotalGrossAmountCents != expectedGross {
		t.Errorf("TotalGrossAmountCents = %d, want %d", summary.TotalGrossAmountCents, expectedGross)
	}

	expectedRevenueShare := int64(1560) // 980 + 580
	if summary.TotalRevenueShareCents != expectedRevenueShare {
		t.Errorf("TotalRevenueShareCents = %d, want %d", summary.TotalRevenueShareCents, expectedRevenueShare)
	}

	expectedNetAmount := int64(5871) // 3688 + 2183
	if summary.TotalNetAmountCents != expectedNetAmount {
		t.Errorf("TotalNetAmountCents = %d, want %d", summary.TotalNetAmountCents, expectedNetAmount)
	}

	// Revenue share should be 20%
	expectedPct := 20.0
	if summary.AverageRevenueSharePct != expectedPct {
		t.Errorf("AverageRevenueSharePct = %.2f, want %.2f", summary.AverageRevenueSharePct, expectedPct)
	}
}

func TestFeeVerificationService_CalculateTierSavings(t *testing.T) {
	svc := NewFeeVerificationService()

	// $10,000 gross revenue
	grossAmountCents := int64(1000000)

	// Test SMALL_DEV_0 savings vs DEFAULT_20
	result := svc.CalculateTierSavings(grossAmountCents, valueobject.RevenueShareTierSmallDev0)

	// Default (20%): $2000 revenue share + $290 processing = $2290
	// Small Dev 0 (0%): $0 revenue share + $290 processing = $290
	// Savings: $2290 - $290 = $2000

	expectedDefaultFees := int64(229000) // 20% + 2.9% of 1000000
	if result.DefaultTierFeesCents != expectedDefaultFees {
		t.Errorf("DefaultTierFeesCents = %d, want %d", result.DefaultTierFeesCents, expectedDefaultFees)
	}

	expectedCurrentFees := int64(29000) // 0% + 2.9% of 1000000
	if result.CurrentTierFeesCents != expectedCurrentFees {
		t.Errorf("CurrentTierFeesCents = %d, want %d", result.CurrentTierFeesCents, expectedCurrentFees)
	}

	expectedSavings := int64(200000) // $2000
	if result.SavingsCents != expectedSavings {
		t.Errorf("SavingsCents = %d, want %d", result.SavingsCents, expectedSavings)
	}
}

func TestFeeVerificationService_CalculateTierSavings_LargeDev(t *testing.T) {
	svc := NewFeeVerificationService()

	// $10,000 gross revenue
	grossAmountCents := int64(1000000)

	// Test LARGE_DEV_15 savings vs DEFAULT_20
	result := svc.CalculateTierSavings(grossAmountCents, valueobject.RevenueShareTierLargeDev)

	// Default (20%): $2000 revenue share + $290 processing = $2290
	// Large Dev 15 (15%): $1500 revenue share + $290 processing = $1790
	// Savings: $2290 - $1790 = $500

	expectedCurrentFees := int64(179000) // 15% + 2.9% of 1000000
	if result.CurrentTierFeesCents != expectedCurrentFees {
		t.Errorf("CurrentTierFeesCents = %d, want %d", result.CurrentTierFeesCents, expectedCurrentFees)
	}

	expectedSavings := int64(50000) // $500
	if result.SavingsCents != expectedSavings {
		t.Errorf("SavingsCents = %d, want %d", result.SavingsCents, expectedSavings)
	}
}

// Helper to create transaction with date
func createTransaction(grossCents, shopifyFeeCents, processingFeeCents, taxCents, netCents int64, date time.Time) *entity.Transaction {
	return &entity.Transaction{
		ID:                 uuid.New(),
		GrossAmountCents:   grossCents,
		ShopifyFeeCents:    shopifyFeeCents,
		ProcessingFeeCents: processingFeeCents,
		TaxOnFeesCents:     taxCents,
		NetAmountCents:     netCents,
		TransactionDate:    date,
	}
}

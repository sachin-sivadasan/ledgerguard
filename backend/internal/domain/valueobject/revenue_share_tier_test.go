package valueobject

import (
	"testing"
)

func TestRevenueShareTier_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		tier     RevenueShareTier
		expected bool
	}{
		{"DEFAULT_20 is valid", RevenueShareTierDefault, true},
		{"SMALL_DEV_0 is valid", RevenueShareTierSmallDev0, true},
		{"SMALL_DEV_15 is valid", RevenueShareTierSmallDev15, true},
		{"LARGE_DEV_15 is valid", RevenueShareTierLargeDev, true},
		{"empty string is invalid", RevenueShareTier(""), false},
		{"random string is invalid", RevenueShareTier("RANDOM"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.IsValid(); got != tt.expected {
				t.Errorf("RevenueShareTier.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRevenueShareTier_RevenueSharePercent(t *testing.T) {
	tests := []struct {
		name     string
		tier     RevenueShareTier
		expected float64
	}{
		{"DEFAULT_20 returns 20%", RevenueShareTierDefault, 20.0},
		{"SMALL_DEV_0 returns 0%", RevenueShareTierSmallDev0, 0.0},
		{"SMALL_DEV_15 returns 15%", RevenueShareTierSmallDev15, 15.0},
		{"LARGE_DEV_15 returns 15%", RevenueShareTierLargeDev, 15.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.RevenueSharePercent(); got != tt.expected {
				t.Errorf("RevenueShareTier.RevenueSharePercent() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRevenueShareTier_CalculateRevenueShareCents(t *testing.T) {
	tests := []struct {
		name             string
		tier             RevenueShareTier
		grossAmountCents int64
		expected         int64
	}{
		{"DEFAULT_20 on $49.00 = $9.80", RevenueShareTierDefault, 4900, 980},
		{"SMALL_DEV_0 on $49.00 = $0.00", RevenueShareTierSmallDev0, 4900, 0},
		{"SMALL_DEV_15 on $49.00 = $7.35", RevenueShareTierSmallDev15, 4900, 735},
		{"LARGE_DEV_15 on $49.00 = $7.35", RevenueShareTierLargeDev, 4900, 735},
		{"DEFAULT_20 on $100.00 = $20.00", RevenueShareTierDefault, 10000, 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.CalculateRevenueShareCents(tt.grossAmountCents); got != tt.expected {
				t.Errorf("CalculateRevenueShareCents(%d) = %d, want %d", tt.grossAmountCents, got, tt.expected)
			}
		})
	}
}

func TestCalculateProcessingFeeCents(t *testing.T) {
	tests := []struct {
		name             string
		grossAmountCents int64
		expected         int64
	}{
		{"$49.00 = $1.42 (2.9%)", 4900, 142},
		{"$100.00 = $2.90 (2.9%)", 10000, 290},
		{"$29.00 = $0.84 (2.9%)", 2900, 84},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateProcessingFeeCents(tt.grossAmountCents); got != tt.expected {
				t.Errorf("CalculateProcessingFeeCents(%d) = %d, want %d", tt.grossAmountCents, got, tt.expected)
			}
		})
	}
}

func TestRevenueShareTier_CalculateFeeBreakdown(t *testing.T) {
	// Test case: $49/month subscription with DEFAULT_20 tier and 8% tax on fees
	// From Shopify docs example:
	// grossAmount: $49.00
	// Revenue Share (20%): -$9.80
	// Processing Fee (2.9%): -$1.42
	// Tax on fees (8%): -$0.90 (on $11.22 fees)
	// netAmount: $36.88

	tier := RevenueShareTierDefault
	grossAmountCents := int64(4900)
	taxRate := 0.08

	breakdown := tier.CalculateFeeBreakdown(grossAmountCents, taxRate)

	if breakdown.GrossAmountCents != 4900 {
		t.Errorf("GrossAmountCents = %d, want 4900", breakdown.GrossAmountCents)
	}

	if breakdown.RevenueShareCents != 980 {
		t.Errorf("RevenueShareCents = %d, want 980", breakdown.RevenueShareCents)
	}

	if breakdown.ProcessingFeeCents != 142 {
		t.Errorf("ProcessingFeeCents = %d, want 142", breakdown.ProcessingFeeCents)
	}

	// Tax on fees: 8% of (980 + 142) = 8% of 1122 = 89.76 ≈ 89
	expectedTax := int64(89)
	if breakdown.TaxOnFeesCents != expectedTax {
		t.Errorf("TaxOnFeesCents = %d, want %d", breakdown.TaxOnFeesCents, expectedTax)
	}

	// Total fees: 980 + 142 + 89 = 1211
	expectedTotalFees := int64(1211)
	if breakdown.TotalFeesCents != expectedTotalFees {
		t.Errorf("TotalFeesCents = %d, want %d", breakdown.TotalFeesCents, expectedTotalFees)
	}

	// Net amount: 4900 - 1211 = 3689
	expectedNetAmount := int64(3689)
	if breakdown.NetAmountCents != expectedNetAmount {
		t.Errorf("NetAmountCents = %d, want %d", breakdown.NetAmountCents, expectedNetAmount)
	}
}

func TestRevenueShareTier_CalculateFeeBreakdown_SmallDev0(t *testing.T) {
	// Test case: $49/month subscription with SMALL_DEV_0 tier (0% revenue share)
	// grossAmount: $49.00
	// Revenue Share (0%): $0.00
	// Processing Fee (2.9%): -$1.42
	// Tax on fees (8%): -$0.11 (on $1.42 fees)
	// netAmount: $47.47

	tier := RevenueShareTierSmallDev0
	grossAmountCents := int64(4900)
	taxRate := 0.08

	breakdown := tier.CalculateFeeBreakdown(grossAmountCents, taxRate)

	if breakdown.RevenueShareCents != 0 {
		t.Errorf("RevenueShareCents = %d, want 0 for SMALL_DEV_0", breakdown.RevenueShareCents)
	}

	if breakdown.ProcessingFeeCents != 142 {
		t.Errorf("ProcessingFeeCents = %d, want 142", breakdown.ProcessingFeeCents)
	}

	// Tax on fees: 8% of 142 = 11.36 ≈ 11
	expectedTax := int64(11)
	if breakdown.TaxOnFeesCents != expectedTax {
		t.Errorf("TaxOnFeesCents = %d, want %d", breakdown.TaxOnFeesCents, expectedTax)
	}

	// Total fees: 0 + 142 + 11 = 153
	expectedTotalFees := int64(153)
	if breakdown.TotalFeesCents != expectedTotalFees {
		t.Errorf("TotalFeesCents = %d, want %d", breakdown.TotalFeesCents, expectedTotalFees)
	}

	// Net amount: 4900 - 153 = 4747
	expectedNetAmount := int64(4747)
	if breakdown.NetAmountCents != expectedNetAmount {
		t.Errorf("NetAmountCents = %d, want %d", breakdown.NetAmountCents, expectedNetAmount)
	}
}

func TestRevenueShareTier_IsReducedPlan(t *testing.T) {
	tests := []struct {
		name     string
		tier     RevenueShareTier
		expected bool
	}{
		{"DEFAULT_20 is not reduced plan", RevenueShareTierDefault, false},
		{"SMALL_DEV_0 is reduced plan", RevenueShareTierSmallDev0, true},
		{"SMALL_DEV_15 is reduced plan", RevenueShareTierSmallDev15, true},
		{"LARGE_DEV_15 is reduced plan", RevenueShareTierLargeDev, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tier.IsReducedPlan(); got != tt.expected {
				t.Errorf("RevenueShareTier.IsReducedPlan() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseRevenueShareTier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected RevenueShareTier
	}{
		{"valid DEFAULT_20", "DEFAULT_20", RevenueShareTierDefault},
		{"valid SMALL_DEV_0", "SMALL_DEV_0", RevenueShareTierSmallDev0},
		{"valid SMALL_DEV_15", "SMALL_DEV_15", RevenueShareTierSmallDev15},
		{"valid LARGE_DEV_15", "LARGE_DEV_15", RevenueShareTierLargeDev},
		{"invalid returns DEFAULT", "INVALID", RevenueShareTierDefault},
		{"empty returns DEFAULT", "", RevenueShareTierDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseRevenueShareTier(tt.input); got != tt.expected {
				t.Errorf("ParseRevenueShareTier(%s) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

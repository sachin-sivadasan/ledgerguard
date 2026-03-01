package valueobject

// RevenueShareTier represents the Shopify revenue share tier for an app
// Based on Shopify's Reduced Revenue Share Plan:
// - DEFAULT: 20% revenue share (not registered for reduced plan)
// - SMALL_DEV_0: 0% on first $1M lifetime (registered, <$20M app / <$100M company)
// - SMALL_DEV_15: 15% after $1M lifetime (registered, <$20M app / <$100M company)
// - LARGE_DEV: 15% on all revenue (registered, ≥$20M app OR ≥$100M company)
type RevenueShareTier string

const (
	// RevenueShareTierDefault is 20% revenue share (not registered for reduced plan)
	RevenueShareTierDefault RevenueShareTier = "DEFAULT_20"

	// RevenueShareTierSmallDev0 is 0% revenue share on first $1M lifetime
	// Eligible: <$20M prior-year app revenue AND <$100M company revenue
	RevenueShareTierSmallDev0 RevenueShareTier = "SMALL_DEV_0"

	// RevenueShareTierSmallDev15 is 15% revenue share after $1M lifetime
	// Eligible: <$20M prior-year app revenue AND <$100M company revenue
	RevenueShareTierSmallDev15 RevenueShareTier = "SMALL_DEV_15"

	// RevenueShareTierLargeDev is 15% revenue share on all revenue
	// Eligible: ≥$20M prior-year app revenue OR ≥$100M company revenue
	RevenueShareTierLargeDev RevenueShareTier = "LARGE_DEV_15"
)

// Processing fee is always 2.9% regardless of tier
const ProcessingFeePercent = 2.9

func (t RevenueShareTier) String() string {
	return string(t)
}

func (t RevenueShareTier) IsValid() bool {
	switch t {
	case RevenueShareTierDefault, RevenueShareTierSmallDev0, RevenueShareTierSmallDev15, RevenueShareTierLargeDev:
		return true
	}
	return false
}

// RevenueSharePercent returns the revenue share percentage for this tier
func (t RevenueShareTier) RevenueSharePercent() float64 {
	switch t {
	case RevenueShareTierDefault:
		return 20.0
	case RevenueShareTierSmallDev0:
		return 0.0
	case RevenueShareTierSmallDev15, RevenueShareTierLargeDev:
		return 15.0
	default:
		return 20.0 // Default to 20% if unknown
	}
}

// DisplayName returns a human-readable name for the tier
func (t RevenueShareTier) DisplayName() string {
	switch t {
	case RevenueShareTierDefault:
		return "Default (20%)"
	case RevenueShareTierSmallDev0:
		return "Small Developer (0%)"
	case RevenueShareTierSmallDev15:
		return "Small Developer (15%)"
	case RevenueShareTierLargeDev:
		return "Large Developer (15%)"
	default:
		return "Unknown"
	}
}

// Description returns a description of the tier eligibility
func (t RevenueShareTier) Description() string {
	switch t {
	case RevenueShareTierDefault:
		return "Not registered for reduced revenue share plan"
	case RevenueShareTierSmallDev0:
		return "0% on first $1M lifetime (under $1M earned)"
	case RevenueShareTierSmallDev15:
		return "15% after $1M lifetime earnings"
	case RevenueShareTierLargeDev:
		return "15% on all revenue (large developer)"
	default:
		return "Unknown tier"
	}
}

// IsReducedPlan returns true if registered for reduced revenue share plan
func (t RevenueShareTier) IsReducedPlan() bool {
	return t != RevenueShareTierDefault
}

// CalculateRevenueShareCents calculates the revenue share fee in cents
func (t RevenueShareTier) CalculateRevenueShareCents(grossAmountCents int64) int64 {
	percent := t.RevenueSharePercent()
	return int64(float64(grossAmountCents) * percent / 100.0)
}

// CalculateProcessingFeeCents calculates the 2.9% processing fee in cents
func CalculateProcessingFeeCents(grossAmountCents int64) int64 {
	return int64(float64(grossAmountCents) * ProcessingFeePercent / 100.0)
}

// FeeBreakdown contains the calculated fee breakdown for a transaction
type FeeBreakdown struct {
	GrossAmountCents     int64
	RevenueShareCents    int64
	ProcessingFeeCents   int64
	TaxOnFeesCents       int64 // Estimated, actual from Shopify
	TotalFeesCents       int64
	NetAmountCents       int64
	RevenueSharePercent  float64
	ProcessingFeePercent float64
}

// CalculateFeeBreakdown calculates the complete fee breakdown for a gross amount
// taxRate is the estimated tax rate on fees (e.g., 0.08 for 8%)
func (t RevenueShareTier) CalculateFeeBreakdown(grossAmountCents int64, taxRate float64) FeeBreakdown {
	revenueShare := t.CalculateRevenueShareCents(grossAmountCents)
	processingFee := CalculateProcessingFeeCents(grossAmountCents)
	taxOnFees := int64(float64(revenueShare+processingFee) * taxRate)
	totalFees := revenueShare + processingFee + taxOnFees
	netAmount := grossAmountCents - totalFees

	return FeeBreakdown{
		GrossAmountCents:     grossAmountCents,
		RevenueShareCents:    revenueShare,
		ProcessingFeeCents:   processingFee,
		TaxOnFeesCents:       taxOnFees,
		TotalFeesCents:       totalFees,
		NetAmountCents:       netAmount,
		RevenueSharePercent:  t.RevenueSharePercent(),
		ProcessingFeePercent: ProcessingFeePercent,
	}
}

// ParseRevenueShareTier parses a string into a RevenueShareTier
func ParseRevenueShareTier(s string) RevenueShareTier {
	tier := RevenueShareTier(s)
	if tier.IsValid() {
		return tier
	}
	return RevenueShareTierSmallDev0 // Default to 0% for most indie devs
}

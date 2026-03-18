package valueobject

// BillingPlan represents a LedgerSpear subscription plan for B2B billing.
type BillingPlan string

const (
	BillingPlanStarter BillingPlan = "STARTER"
	BillingPlanPro     BillingPlan = "PRO"
)

func (b BillingPlan) String() string {
	return string(b)
}

func (b BillingPlan) IsValid() bool {
	switch b {
	case BillingPlanStarter, BillingPlanPro:
		return true
	default:
		return false
	}
}

// PriceUSDCents returns the monthly price in USD cents.
func (b BillingPlan) PriceUSDCents() int {
	switch b {
	case BillingPlanStarter:
		return 24900 // $249/mo
	case BillingPlanPro:
		return 49900 // $499/mo
	default:
		return 0
	}
}

// ParseBillingPlan converts a string to BillingPlan.
func ParseBillingPlan(s string) BillingPlan {
	switch s {
	case "STARTER":
		return BillingPlanStarter
	case "PRO":
		return BillingPlanPro
	default:
		return BillingPlan(s)
	}
}

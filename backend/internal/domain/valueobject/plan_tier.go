package valueobject

type PlanTier string

const (
	PlanTierFree PlanTier = "FREE"
	PlanTierPro  PlanTier = "PRO"
)

func (p PlanTier) String() string {
	return string(p)
}

func (p PlanTier) IsValid() bool {
	switch p {
	case PlanTierFree, PlanTierPro:
		return true
	default:
		return false
	}
}

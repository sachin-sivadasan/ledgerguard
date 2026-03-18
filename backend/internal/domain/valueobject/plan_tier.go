package valueobject

type PlanTier string

const (
	PlanTierFree    PlanTier = "FREE"
	PlanTierStarter PlanTier = "STARTER"
	PlanTierPro     PlanTier = "PRO"
)

func (p PlanTier) String() string {
	return string(p)
}

func (p PlanTier) IsValid() bool {
	switch p {
	case PlanTierFree, PlanTierStarter, PlanTierPro:
		return true
	default:
		return false
	}
}

package valueobject

type RiskState string

const (
	RiskStateSafe            RiskState = "SAFE"
	RiskStateOneCycleMissed  RiskState = "ONE_CYCLE_MISSED"
	RiskStateTwoCyclesMissed RiskState = "TWO_CYCLES_MISSED"
	RiskStateChurned         RiskState = "CHURNED"
)

func (r RiskState) String() string {
	return string(r)
}

func (r RiskState) IsValid() bool {
	switch r {
	case RiskStateSafe, RiskStateOneCycleMissed, RiskStateTwoCyclesMissed, RiskStateChurned:
		return true
	}
	return false
}

// IsAtRisk returns true if the subscription is at risk of churning
func (r RiskState) IsAtRisk() bool {
	return r == RiskStateOneCycleMissed || r == RiskStateTwoCyclesMissed
}

// IsChurned returns true if the subscription has churned
func (r RiskState) IsChurned() bool {
	return r == RiskStateChurned
}

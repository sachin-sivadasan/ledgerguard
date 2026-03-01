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

// ParseRiskState parses a string into a RiskState
func ParseRiskState(s string) RiskState {
	switch s {
	case "SAFE":
		return RiskStateSafe
	case "ONE_CYCLE_MISSED":
		return RiskStateOneCycleMissed
	case "TWO_CYCLES_MISSED":
		return RiskStateTwoCyclesMissed
	case "CHURNED":
		return RiskStateChurned
	default:
		return RiskStateSafe // Default to safe if unknown
	}
}

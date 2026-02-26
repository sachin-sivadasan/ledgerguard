package valueobject

type ChargeType string

const (
	ChargeTypeRecurring ChargeType = "RECURRING"
	ChargeTypeUsage     ChargeType = "USAGE"
	ChargeTypeOneTime   ChargeType = "ONE_TIME"
	ChargeTypeRefund    ChargeType = "REFUND"
)

func (c ChargeType) String() string {
	return string(c)
}

func (c ChargeType) IsValid() bool {
	switch c {
	case ChargeTypeRecurring, ChargeTypeUsage, ChargeTypeOneTime, ChargeTypeRefund:
		return true
	}
	return false
}

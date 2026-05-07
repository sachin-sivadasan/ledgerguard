package valueobject

type MemberStatus string

const (
	MemberStatusActive    MemberStatus = "ACTIVE"
	MemberStatusSuspended MemberStatus = "SUSPENDED"
)

func (s MemberStatus) String() string {
	return string(s)
}

func (s MemberStatus) IsValid() bool {
	switch s {
	case MemberStatusActive, MemberStatusSuspended:
		return true
	default:
		return false
	}
}

// IsActive returns true if the member can access org data.
func (s MemberStatus) IsActive() bool {
	return s == MemberStatusActive
}

// ParseMemberStatus converts a string to MemberStatus, returns empty string if invalid.
func ParseMemberStatus(str string) MemberStatus {
	s := MemberStatus(str)
	if s.IsValid() {
		return s
	}
	return ""
}

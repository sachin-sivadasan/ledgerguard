package valueobject

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
	InvitationStatusRevoked  InvitationStatus = "REVOKED"
)

func (s InvitationStatus) String() string {
	return string(s)
}

func (s InvitationStatus) IsValid() bool {
	switch s {
	case InvitationStatusPending, InvitationStatusAccepted, InvitationStatusExpired, InvitationStatusRevoked:
		return true
	default:
		return false
	}
}

// IsPending returns true if the invitation is still waiting to be accepted.
func (s InvitationStatus) IsPending() bool {
	return s == InvitationStatusPending
}

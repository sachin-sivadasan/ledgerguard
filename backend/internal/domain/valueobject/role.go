package valueobject

type Role string

const (
	RoleOwner Role = "OWNER"
	RoleAdmin Role = "ADMIN"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin:
		return true
	default:
		return false
	}
}

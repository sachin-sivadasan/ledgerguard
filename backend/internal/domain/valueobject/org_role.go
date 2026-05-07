package valueobject

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "OWNER"
	OrgRoleAdmin  OrgRole = "ADMIN"
	OrgRoleViewer OrgRole = "VIEWER"
)

func (r OrgRole) String() string {
	return string(r)
}

func (r OrgRole) IsValid() bool {
	switch r {
	case OrgRoleOwner, OrgRoleAdmin, OrgRoleViewer:
		return true
	default:
		return false
	}
}

// CanManageMembers returns true if this role can invite/remove members.
func (r OrgRole) CanManageMembers() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// CanManageAdmins returns true if this role can change admin roles.
func (r OrgRole) CanManageAdmins() bool {
	return r == OrgRoleOwner
}

// CanTriggerSync returns true if this role can trigger data syncs.
func (r OrgRole) CanTriggerSync() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// CanManageOrg returns true if this role can change org settings.
func (r OrgRole) CanManageOrg() bool {
	return r == OrgRoleOwner
}

// CanViewAuditLog returns true if this role can view the audit log.
func (r OrgRole) CanViewAuditLog() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// ParseOrgRole converts a string to OrgRole, returns empty string if invalid.
func ParseOrgRole(s string) OrgRole {
	r := OrgRole(s)
	if r.IsValid() {
		return r
	}
	return ""
}

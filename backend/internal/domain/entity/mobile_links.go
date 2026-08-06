package entity

import "github.com/google/uuid"

// MobileLinks holds the public store identifiers for an app's companion mobile apps, used
// to pull ratings + reviews from the public App Store / Google Play endpoints (no creds).
type MobileLinks struct {
	AppID       uuid.UUID
	IosAppID    string // Apple numeric app id, e.g. "310633997"
	PlayPackage string // Android package, e.g. "com.whatsapp"
}

// HasAny reports whether at least one store is linked.
func (m *MobileLinks) HasAny() bool {
	return m != nil && (m.IosAppID != "" || m.PlayPackage != "")
}

package valueobject

// SubscriptionStatus represents the Shopify subscription status
type SubscriptionStatus string

const (
	// SubscriptionStatusActive - The subscription is active and charging
	SubscriptionStatusActive SubscriptionStatus = "ACTIVE"
	// SubscriptionStatusCancelled - The subscription has been cancelled
	SubscriptionStatusCancelled SubscriptionStatus = "CANCELLED"
	// SubscriptionStatusFrozen - The subscription is frozen due to payment failure
	SubscriptionStatusFrozen SubscriptionStatus = "FROZEN"
	// SubscriptionStatusExpired - The subscription has expired
	SubscriptionStatusExpired SubscriptionStatus = "EXPIRED"
	// SubscriptionStatusPending - The subscription is pending activation
	SubscriptionStatusPending SubscriptionStatus = "PENDING"
)

// String returns the string representation
func (s SubscriptionStatus) String() string {
	return string(s)
}

// IsTerminal returns true if the subscription has ended
func (s SubscriptionStatus) IsTerminal() bool {
	return s == SubscriptionStatusCancelled || s == SubscriptionStatusExpired
}

// IsActive returns true if the subscription is active
func (s SubscriptionStatus) IsActive() bool {
	return s == SubscriptionStatusActive
}

// IsFrozen returns true if the subscription is frozen
func (s SubscriptionStatus) IsFrozen() bool {
	return s == SubscriptionStatusFrozen
}

// IsPending returns true if the subscription is pending
func (s SubscriptionStatus) IsPending() bool {
	return s == SubscriptionStatusPending
}

// ParseSubscriptionStatus parses a string to SubscriptionStatus
func ParseSubscriptionStatus(s string) SubscriptionStatus {
	switch s {
	case "ACTIVE":
		return SubscriptionStatusActive
	case "CANCELLED":
		return SubscriptionStatusCancelled
	case "FROZEN":
		return SubscriptionStatusFrozen
	case "EXPIRED":
		return SubscriptionStatusExpired
	case "PENDING":
		return SubscriptionStatusPending
	default:
		return SubscriptionStatusActive // Default to ACTIVE for unknown
	}
}

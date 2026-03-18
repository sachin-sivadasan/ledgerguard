package valueobject

// BillingSubscriptionStatus represents the status of a Razorpay subscription.
type BillingSubscriptionStatus string

const (
	BillingSubscriptionStatusCreated   BillingSubscriptionStatus = "CREATED"
	BillingSubscriptionStatusActive    BillingSubscriptionStatus = "ACTIVE"
	BillingSubscriptionStatusPending   BillingSubscriptionStatus = "PENDING"
	BillingSubscriptionStatusHalted    BillingSubscriptionStatus = "HALTED"
	BillingSubscriptionStatusCancelled BillingSubscriptionStatus = "CANCELLED"
	BillingSubscriptionStatusCompleted BillingSubscriptionStatus = "COMPLETED"
)

func (s BillingSubscriptionStatus) String() string {
	return string(s)
}

func (s BillingSubscriptionStatus) IsValid() bool {
	switch s {
	case BillingSubscriptionStatusCreated,
		BillingSubscriptionStatusActive,
		BillingSubscriptionStatusPending,
		BillingSubscriptionStatusHalted,
		BillingSubscriptionStatusCancelled,
		BillingSubscriptionStatusCompleted:
		return true
	default:
		return false
	}
}

// IsActive returns true if the subscription is currently active.
func (s BillingSubscriptionStatus) IsActive() bool {
	return s == BillingSubscriptionStatusActive
}

// IsTerminal returns true if the subscription has ended.
func (s BillingSubscriptionStatus) IsTerminal() bool {
	return s == BillingSubscriptionStatusCancelled || s == BillingSubscriptionStatusCompleted
}

// ParseBillingSubscriptionStatus converts a string to BillingSubscriptionStatus.
func ParseBillingSubscriptionStatus(s string) BillingSubscriptionStatus {
	switch s {
	case "CREATED":
		return BillingSubscriptionStatusCreated
	case "ACTIVE":
		return BillingSubscriptionStatusActive
	case "PENDING":
		return BillingSubscriptionStatusPending
	case "HALTED":
		return BillingSubscriptionStatusHalted
	case "CANCELLED":
		return BillingSubscriptionStatusCancelled
	case "COMPLETED":
		return BillingSubscriptionStatusCompleted
	default:
		return BillingSubscriptionStatus(s)
	}
}

package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// SubscriptionEvent represents a subscription lifecycle event
// Used for tracking state transitions and understanding churn patterns
type SubscriptionEvent struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	FromStatus     string                  // Previous status (ACTIVE, CANCELLED, FROZEN, etc.)
	ToStatus       string                  // New status
	FromRiskState  valueobject.RiskState   // Previous risk state
	ToRiskState    valueobject.RiskState   // New risk state
	EventType      string                  // webhook, sync, manual, billing_failure, app_uninstalled
	Reason         string                  // Human-readable reason for the change
	OccurredAt     time.Time               // When the event occurred
	CreatedAt      time.Time               // When we recorded the event
}

// NewSubscriptionEvent creates a new subscription lifecycle event
func NewSubscriptionEvent(
	subscriptionID uuid.UUID,
	fromStatus, toStatus string,
	fromRiskState, toRiskState valueobject.RiskState,
	eventType string,
	reason string,
) *SubscriptionEvent {
	now := time.Now().UTC()
	return &SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: subscriptionID,
		FromStatus:     fromStatus,
		ToStatus:       toStatus,
		FromRiskState:  fromRiskState,
		ToRiskState:    toRiskState,
		EventType:      eventType,
		Reason:         reason,
		OccurredAt:     now,
		CreatedAt:      now,
	}
}

// IsChurnEvent returns true if this event represents a transition to churned state
func (e *SubscriptionEvent) IsChurnEvent() bool {
	return e.ToRiskState == valueobject.RiskStateChurned &&
		e.FromRiskState != valueobject.RiskStateChurned
}

// IsReactivationEvent returns true if this event represents a reactivation
func (e *SubscriptionEvent) IsReactivationEvent() bool {
	return e.FromRiskState == valueobject.RiskStateChurned &&
		e.ToRiskState == valueobject.RiskStateSafe
}

// IsVoluntaryChurn returns true if churn was initiated by the user
func (e *SubscriptionEvent) IsVoluntaryChurn() bool {
	return e.IsChurnEvent() &&
		(e.EventType == "app_uninstalled" || e.ToStatus == "CANCELLED")
}

// IsInvoluntaryChurn returns true if churn was due to payment failure
func (e *SubscriptionEvent) IsInvoluntaryChurn() bool {
	return e.IsChurnEvent() && e.EventType == "billing_failure"
}

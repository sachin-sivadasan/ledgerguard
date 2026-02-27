package graphql

import (
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/application/service"
)

// Resolver is the root resolver for the Revenue API GraphQL schema
type Resolver struct {
	subscriptionService *service.SubscriptionStatusService
	usageService        *service.UsageStatusService
}

// NewResolver creates a new Resolver with the required services
func NewResolver(
	subscriptionService *service.SubscriptionStatusService,
	usageService *service.UsageStatusService,
) *Resolver {
	return &Resolver{
		subscriptionService: subscriptionService,
		usageService:        usageService,
	}
}

// RiskState enum values
type RiskState string

const (
	RiskStateSafe            RiskState = "SAFE"
	RiskStateOneCycleMissed  RiskState = "ONE_CYCLE_MISSED"
	RiskStateTwoCyclesMissed RiskState = "TWO_CYCLES_MISSED"
	RiskStateChurned         RiskState = "CHURNED"
)

// SubscriptionStatusEnum values
type SubscriptionStatusEnum string

const (
	SubscriptionStatusActive    SubscriptionStatusEnum = "ACTIVE"
	SubscriptionStatusCancelled SubscriptionStatusEnum = "CANCELLED"
	SubscriptionStatusFrozen    SubscriptionStatusEnum = "FROZEN"
	SubscriptionStatusPending   SubscriptionStatusEnum = "PENDING"
)

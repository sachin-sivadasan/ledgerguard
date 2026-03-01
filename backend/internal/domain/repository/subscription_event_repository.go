package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// SubscriptionEventRepository handles persistence of subscription lifecycle events
type SubscriptionEventRepository interface {
	// Create stores a new subscription event
	Create(ctx context.Context, event *entity.SubscriptionEvent) error

	// FindBySubscriptionID retrieves all events for a subscription
	FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.SubscriptionEvent, error)

	// FindByAppID retrieves all events for subscriptions of an app within a date range
	FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.SubscriptionEvent, error)

	// FindChurnEvents retrieves churn events within a date range
	FindChurnEvents(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.SubscriptionEvent, error)

	// CountByEventType counts events by type within a date range
	CountByEventType(ctx context.Context, appID uuid.UUID, from, to time.Time) (map[string]int, error)
}

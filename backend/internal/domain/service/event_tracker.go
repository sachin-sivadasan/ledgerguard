package service

import "context"

// EventProperties is a map of event properties for analytics tracking.
type EventProperties map[string]interface{}

// EventTracker tracks user lifecycle events for analytics.
// Implementations must be fire-and-forget (non-blocking).
type EventTracker interface {
	Track(ctx context.Context, userID string, event string, properties EventProperties)
	SetUserProperties(ctx context.Context, userID string, properties EventProperties)
}

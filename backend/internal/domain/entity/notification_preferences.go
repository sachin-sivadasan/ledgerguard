package entity

import (
	"time"

	"github.com/google/uuid"
)

// NotificationPreferences represents a user's notification settings
type NotificationPreferences struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	CriticalEnabled      bool      // Risk state change alerts
	DailySummaryEnabled  bool      // Daily summary notifications
	DailySummaryTime     time.Time // Preferred time for daily summary (hour/minute only)
	SlackWebhookURL      string    // Slack integration (Pro tier)
	EmailEnabled         bool      // Email delivery for all alert types
	SlackEnabled         bool      // Slack delivery toggle
	ChurnAlertsEnabled   bool      // Alert on subscription risk transitions
	RevenueAlertsEnabled bool      // Daily revenue summary alerts
	ReviewAlertsEnabled  bool      // New app store review alerts
	RiskThresholdDays    int       // Alert after N days overdue (7–90)
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewNotificationPreferences creates default notification preferences
func NewNotificationPreferences(userID uuid.UUID) *NotificationPreferences {
	now := time.Now().UTC()
	// Default daily summary time: 8:00 AM UTC
	defaultTime := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)

	return &NotificationPreferences{
		ID:                   uuid.New(),
		UserID:               userID,
		CriticalEnabled:      true,
		DailySummaryEnabled:  true,
		DailySummaryTime:     defaultTime,
		EmailEnabled:         true,
		SlackEnabled:         false,
		ChurnAlertsEnabled:   true,
		RevenueAlertsEnabled: true,
		ReviewAlertsEnabled:  true,
		RiskThresholdDays:    30,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// ShouldSendCritical returns true if critical alerts are enabled
func (p *NotificationPreferences) ShouldSendCritical() bool {
	return p.CriticalEnabled
}

// ShouldSendDailySummary returns true if daily summaries are enabled
func (p *NotificationPreferences) ShouldSendDailySummary() bool {
	return p.DailySummaryEnabled
}

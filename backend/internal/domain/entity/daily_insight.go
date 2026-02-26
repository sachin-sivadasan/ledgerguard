package entity

import (
	"time"

	"github.com/google/uuid"
)

// DailyInsight represents an AI-generated daily summary (Pro tier only)
// One insight per app per day
type DailyInsight struct {
	ID          uuid.UUID
	AppID       uuid.UUID
	Date        time.Time // Date of insight (truncated to day)
	InsightText string    // AI-generated summary (80-120 words)
	CreatedAt   time.Time
}

// NewDailyInsight creates a new daily insight
func NewDailyInsight(appID uuid.UUID, date time.Time, insightText string) *DailyInsight {
	now := time.Now().UTC()
	// Truncate date to start of day
	truncatedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	return &DailyInsight{
		ID:          uuid.New(),
		AppID:       appID,
		Date:        truncatedDate,
		InsightText: insightText,
		CreatedAt:   now,
	}
}

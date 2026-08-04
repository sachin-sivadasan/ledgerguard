package entity

import (
	"time"

	"github.com/google/uuid"
)

// ForecastPoint represents a single projected data point.
type ForecastPoint struct {
	Date             time.Time
	ExpectedCents    int64
	OptimisticCents  int64
	PessimisticCents int64
}

// ForecastResult holds the complete forecast for an app.
type ForecastResult struct {
	AppID          uuid.UUID
	GeneratedAt    time.Time
	Model          string
	DataPointsUsed int
	Points         []ForecastPoint
}

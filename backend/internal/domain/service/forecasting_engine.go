package service

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

const (
	// MinDataPointsForForecast is the minimum number of snapshots needed for forecasting.
	MinDataPointsForForecast = 90
	// ForecastModelLinear is the linear regression model identifier.
	ForecastModelLinear = "linear"
	// ForecastModelExponential is the exponential smoothing model identifier.
	ForecastModelExponential = "exponential"
)

var ErrInsufficientData = errors.New("insufficient data points for forecasting")

// ForecastingEngine provides revenue forecasting capabilities.
type ForecastingEngine struct{}

// NewForecastingEngine creates a new ForecastingEngine.
func NewForecastingEngine() *ForecastingEngine {
	return &ForecastingEngine{}
}

// LinearRegressionForecast fits an OLS line through daily MRR values and projects
// forward with confidence bands (±15% for optimistic/pessimistic).
func (e *ForecastingEngine) LinearRegressionForecast(
	snapshots []*entity.DailyMetricsSnapshot,
	months int,
	appID uuid.UUID,
) (*entity.ForecastResult, error) {
	if len(snapshots) < MinDataPointsForForecast {
		return nil, ErrInsufficientData
	}

	n := float64(len(snapshots))
	// Use day index as x, MRR cents as y
	var sumX, sumY, sumXY, sumX2 float64
	for i, s := range snapshots {
		x := float64(i)
		y := float64(s.ActiveMRRCents)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// OLS coefficients: y = a + b*x
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		// All x values are the same (shouldn't happen with daily data)
		denominator = 1
	}
	b := (n*sumXY - sumX*sumY) / denominator
	a := (sumY - b*sumX) / n

	// Project forward
	lastDate := snapshots[len(snapshots)-1].Date
	lastIndex := float64(len(snapshots) - 1)

	points := make([]entity.ForecastPoint, 0, months)
	for m := 1; m <= months; m++ {
		futureDate := lastDate.AddDate(0, m, 0)
		futureDays := lastIndex + float64(m*30)
		expected := a + b*futureDays
		if expected < 0 {
			expected = 0
		}
		expectedCents := int64(math.Round(expected))
		optimistic := int64(math.Round(expected * 1.15))
		pessimistic := int64(math.Round(expected * 0.85))
		if pessimistic < 0 {
			pessimistic = 0
		}
		points = append(points, entity.ForecastPoint{
			Date:             futureDate,
			ExpectedCents:    expectedCents,
			OptimisticCents:  optimistic,
			PessimisticCents: pessimistic,
		})
	}

	return &entity.ForecastResult{
		AppID:          appID,
		GeneratedAt:    time.Now(),
		Model:          ForecastModelLinear,
		DataPointsUsed: len(snapshots),
		Points:         points,
	}, nil
}

// ExponentialSmoothingForecast uses Holt's method (double exponential smoothing)
// with trend to project MRR forward.
func (e *ForecastingEngine) ExponentialSmoothingForecast(
	snapshots []*entity.DailyMetricsSnapshot,
	months int,
	alpha float64,
	appID uuid.UUID,
) (*entity.ForecastResult, error) {
	if len(snapshots) < MinDataPointsForForecast {
		return nil, ErrInsufficientData
	}

	// Clamp alpha
	if alpha < 0.1 {
		alpha = 0.1
	}
	if alpha > 0.9 {
		alpha = 0.9
	}

	// Holt's method: level + trend
	beta := alpha * 0.5 // trend smoothing factor
	level := float64(snapshots[0].ActiveMRRCents)
	trend := 0.0

	// Initialize trend from first few data points
	if len(snapshots) > 30 {
		trend = (float64(snapshots[30].ActiveMRRCents) - float64(snapshots[0].ActiveMRRCents)) / 30.0
	}

	for i := 1; i < len(snapshots); i++ {
		y := float64(snapshots[i].ActiveMRRCents)
		prevLevel := level
		level = alpha*y + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
	}

	// Project forward
	lastDate := snapshots[len(snapshots)-1].Date
	points := make([]entity.ForecastPoint, 0, months)
	for m := 1; m <= months; m++ {
		futureDate := lastDate.AddDate(0, m, 0)
		stepsAhead := float64(m * 30) // daily steps
		expected := level + trend*stepsAhead
		if expected < 0 {
			expected = 0
		}
		expectedCents := int64(math.Round(expected))
		optimistic := int64(math.Round(expected * 1.15))
		pessimistic := int64(math.Round(expected * 0.85))
		if pessimistic < 0 {
			pessimistic = 0
		}
		points = append(points, entity.ForecastPoint{
			Date:             futureDate,
			ExpectedCents:    expectedCents,
			OptimisticCents:  optimistic,
			PessimisticCents: pessimistic,
		})
	}

	return &entity.ForecastResult{
		AppID:          appID,
		GeneratedAt:    time.Now(),
		Model:          ForecastModelExponential,
		DataPointsUsed: len(snapshots),
		Points:         points,
	}, nil
}

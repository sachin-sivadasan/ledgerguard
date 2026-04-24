# F05. Revenue Forecasting (ML)

## What It Will Do
Predict future MRR, churn rate, and revenue at risk using machine learning models trained on historical daily snapshots. Display forecasts on the Analytics Forecasting tab with confidence intervals. Alert users when predicted metrics cross threshold values.

## Why It Matters
Proactive revenue forecasting helps developers plan for seasonal dips, detect early warning signs of accelerating churn, and make informed decisions about pricing changes or marketing spend. Moves LedgerGuard from reactive monitoring to predictive intelligence.

## Dependencies
- Daily metrics snapshots (implemented — 12+ months of historical data needed)
- Analytics Forecasting tab (prototype exists with mock data)
- Notification service (implemented — for threshold alerts)

## Integration Points
- Backend: New domain service `ForecastingEngine` with time-series prediction
- Database: New `forecasts` table for storing predictions
- API: New endpoint `GET /api/v1/forecasts/{appId}` returning predictions with confidence intervals
- Flutter: Wire Forecasting tab to real API (replace mock data)
- Notifications: Alert when forecast predicts MRR drop > 10% or churn spike

## Estimated Scope
- Research and select time-series model: 2 days
- Implement forecasting engine in Go: 3-5 days
- Database schema + API endpoint: 1 day
- Flutter integration: 1 day
- Threshold alerting: 1 day
- Total: ~8-10 days

## Open Questions
- Which model? Options: simple exponential smoothing, ARIMA, Prophet (via Python sidecar), or linear regression (simplest, pure Go)
- How much historical data needed? (Minimum 3 months, ideal 12+ months)
- Should we forecast per-app or aggregate across all apps?
- Forecast horizon: 1 month? 3 months? 6 months?
- Is a Python sidecar acceptable, or must it be pure Go?

## Suggested Approach
1. Start with simple exponential smoothing (pure Go, no external dependencies)
2. Train on daily snapshot MRR values (minimum 90 data points)
3. Generate 30/60/90-day forecasts with upper/lower confidence bounds
4. Store in `forecasts` table: `(app_id, date, metric, predicted_value, lower_bound, upper_bound)`
5. Run forecast generation daily after sync completes
6. Expose via REST API for Flutter consumption
7. Add notification threshold: alert if predicted MRR drops > 10% vs current
8. Later: evaluate Prophet or ARIMA for improved accuracy (requires Python sidecar)

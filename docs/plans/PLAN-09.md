# PLAN-09: Flutter Executive Dashboard

**Date:** 2024-01-XX
**Status:** Completed

## Scope
- Dashboard layout with Primary KPIs (Renewal Rate, Active MRR, Revenue at Risk, Churn)
- Secondary sections with charts (revenue mix, risk distribution)
- Connect to backend `/api/v1/apps/{appId}/metrics/latest`
- Dashboard configuration (choose/reorder KPIs, toggle widgets)
- Time range selector (Last 7 Days, Last Month, Last Quarter, etc.)
- Period-over-period delta badges (green/red, Play Store-style analytics)
- KPI cards scale down for narrow widths

## Blocs
- `DashboardBloc` — LoadDashboardRequested, metrics state
- `PreferencesBloc` — Dashboard KPI/widget configuration

## Widgets
- `KpiCard` — Metric display with delta badge
- `TimeRangeSelector` — Period filter
- `DashboardConfigDialog` — Configure visible KPIs/widgets

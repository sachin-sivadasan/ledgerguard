# 26. Flutter Analytics Suite

## What It Does
A 5-tab analytics dashboard providing deep revenue analysis: Revenue (monthly/annual breakdown by charge type), Forecasting (ML-based revenue predictions), Profit (fee analysis and net revenue), Cohort (subscription retention by signup month), and Multi-App (cross-app comparison). Uses fl_chart for visualizations.

## Architecture
Presentation layer. `AnalyticsProvider` supplies mock data for all 5 tabs. The `AnalyticsScreen` uses a `TabBarView` with dedicated tab widgets for each analytics view. Charts use `fl_chart` (LineChart, BarChart, PieChart).

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/providers/analytics_provider.dart` | ~60 | Analytics data provider |
| `ledgerguard-flutter/lib/screens/analytics/analytics_screen.dart` | ~100 | TabBar container for 5 tabs |
| `ledgerguard-flutter/lib/screens/analytics/revenue_tab.dart` | ~200 | Revenue breakdown charts |
| `ledgerguard-flutter/lib/screens/analytics/forecasting_tab.dart` | ~200 | Revenue prediction charts |
| `ledgerguard-flutter/lib/screens/analytics/profit_tab.dart` | ~200 | Fee analysis and net revenue |
| `ledgerguard-flutter/lib/screens/analytics/cohort_tab.dart` | ~200 | Retention cohort analysis |
| `ledgerguard-flutter/lib/screens/analytics/multi_app_tab.dart` | ~200 | Cross-app comparison |
| `ledgerguard-flutter/lib/models/analytics_model.dart` | ~100 | Analytics data models |
| `ledgerguard-flutter/lib/mock_data/mock_analytics.dart` | ~80 | Mock chart data |

## Data Flow
```
MockAnalytics → AnalyticsProvider → AnalyticsScreen
                                        │
                                   TabBarView
                                        ├── RevenueTab (monthly revenue by type)
                                        ├── ForecastingTab (predicted MRR)
                                        ├── ProfitTab (gross - fees = net)
                                        ├── CohortTab (retention matrix)
                                        └── MultiAppTab (app comparison)
```

## Configuration
None — mock data.

## Widget Tree
```
AnalyticsScreen
├── LgPage (title: "Analytics")
│   ├── TabBar (5 tabs)
│   └── TabBarView
│       ├── RevenueTab
│       │   ├── LgCard: Monthly Revenue BarChart
│       │   │   (stacked: Recurring, Usage, OneTime, Refund)
│       │   └── LgCard: Revenue Mix PieChart
│       ├── ForecastingTab
│       │   ├── LgCard: MRR Forecast LineChart (actual + predicted)
│       │   └── LgCard: Confidence interval bands
│       ├── ProfitTab
│       │   ├── LgCard: Gross vs Net Revenue BarChart
│       │   └── LgCard: Fee Breakdown by tier
│       ├── CohortTab
│       │   └── LgCard: Retention heatmap (month × cohort)
│       └── MultiAppTab
│           ├── LgCard: App MRR comparison BarChart
│           └── LgCard: App risk distribution comparison
```

## State Machine
```
AnalyticsProvider (ChangeNotifier)
  State:
    _selectedAppId: String?
    _timeRange: AnalyticsTimeRange

  Events:
    setSelectedApp(), setTimeRange()

  Computed:
    mrrSnapshots      → monthly MRR data points
    revenueMix        → {recurring, usage, oneTime, refund}
    riskDistribution  → {safe, oneCycle, twoCycles, churned}
    forecastPoints    → predicted future MRR
    cohortData        → retention matrix
```

## Gotchas
- All 5 tabs use mock data — no backend analytics API yet
- Forecasting tab shows predictions but has no real ML model behind it
- Cohort tab uses a simplified retention matrix (not a full survival analysis)
- Multi-app tab only works when multiple mock apps are configured
- `fl_chart` renders differently on web vs mobile — test both

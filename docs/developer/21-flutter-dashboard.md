# 21. Flutter Dashboard

## What It Does
The executive dashboard displays key revenue metrics (MRR, Renewal Rate, Revenue at Risk, Usage Revenue), activity summaries, MRR trend charts, risk distribution, and revenue mix. Supports multi-app filtering and time range selection. Shows an onboarding checklist for new users without connected apps.

## Architecture
Presentation layer in the Flutter prototype (`ledgerguard-flutter/`). Uses Provider for state management with mock data. `DashboardProvider` computes KPIs from mock subscriptions and events. `DashboardScreen` renders a responsive layout with `LgMetricCard` widgets, fl_chart charts, and activity summary tables.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/screens/dashboard/dashboard_screen.dart` | ~300 | Dashboard UI with KPI cards, charts, activity table |
| `ledgerguard-flutter/lib/providers/dashboard_provider.dart` | ~114 | KPI computation, time range filtering, mock data |
| `ledgerguard-flutter/lib/widgets/lg_metric_card.dart` | ~60 | Reusable KPI card with icon, value, label |
| `ledgerguard-flutter/lib/widgets/lg_onboarding_checklist.dart` | ~80 | Setup checklist for new users |
| `ledgerguard-flutter/lib/mock_data/mock_subscriptions.dart` | ~100 | Mock subscription data |
| `ledgerguard-flutter/lib/mock_data/mock_analytics.dart` | ~80 | Mock MRR snapshots, risk distribution, revenue mix |

## Data Flow
```
MockData (subscriptions, events, analytics)
    │
    ▼
DashboardProvider
    ├── mrrCents (SAFE subs only)
    ├── renewalRate (safe / total × 100)
    ├── revenueAtRiskCents (at-risk subs)
    ├── activity (event counts by type + time range)
    ├── mrrTrend (chart data)
    └── riskDistribution / revenueMix
    │
    ▼
DashboardScreen
    ├── App selector dropdown (if >1 app)
    ├── Time range selector
    ├── Row 1: 4x LgMetricCard (MRR, Rate, Risk, Usage)
    ├── Row 2: MRR Trend chart + Risk Funnel
    ├── Row 3: Activity Summary table
    └── Row 4: Recent AI Alerts
```

## Configuration
None — uses mock data. In production, DashboardProvider would call the backend API.

## Widget Tree
```
DashboardScreen
├── LgPage (title: "Dashboard")
│   ├── LgOnboardingChecklist (if no apps)
│   ├── Row: App selector + Time range dropdown
│   ├── Row: 4x LgMetricCard
│   │   ├── LgMetricCard (MRR, attach_money icon)
│   │   ├── LgMetricCard (Renewal Rate, autorenew icon)
│   │   ├── LgMetricCard (Revenue at Risk, warning icon)
│   │   └── LgMetricCard (Usage Revenue, trending_up icon)
│   ├── Row: LgCard (MRR Trend LineChart) + LgCard (Risk Funnel)
│   ├── LgCard: Activity Summary DataTable
│   └── LgCard: Recent Alerts (top 3 insights)
```

## State Machine
```
DashboardProvider (ChangeNotifier)
  State:
    _selectedAppId: String?       → null = all apps
    _timeRange: DashboardTimeRange → thisWeek (default)

  Events:
    setSelectedApp(appId) → filter subscriptions/events by app
    setTimeRange(range)   → filter activity by time window

  Computed (from mock data):
    mrrCents, mrrFormatted        → sum of SAFE subs
    renewalRate                    → safe / total × 100
    revenueAtRiskCents             → sum of at-risk subs
    activity                       → event counts by type
    mrrTrend, riskDistribution     → chart data
    recentAlerts                   → top 3 insights
```

## Gotchas
- All data is from mock — no API calls in the prototype
- MRR only includes SAFE subscriptions (matches backend logic)
- Activity summary uses `_activityCutoff` based on time range enum
- `LgOnboardingChecklist` shows only when `AppsProvider.apps` is empty
- App selector dropdown only appears when multiple mock apps exist

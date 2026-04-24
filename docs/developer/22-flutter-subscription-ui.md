# 22. Flutter Subscription UI

## What It Does
Displays subscription list with multi-filter support (search, status, risk state, plan, app) and a detail view with payment history, risk timeline, and subscription metadata. Powers the core CRM workflow for tracking individual merchant subscriptions.

## Architecture
Presentation layer. `SubscriptionProvider` (ChangeNotifier) manages filter state and computes filtered lists from mock data. `SubscriptionListScreen` renders the list; `SubscriptionDetailScreen` renders the detail view with tabs for payment history and risk timeline.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/providers/subscription_provider.dart` | ~100 | Filter state, computed lists, detail lookup |
| `ledgerguard-flutter/lib/screens/subscriptions/subscription_list_screen.dart` | ~200 | Filtered subscription list |
| `ledgerguard-flutter/lib/screens/subscriptions/subscription_detail_screen.dart` | ~250 | Detail view with payment history + risk timeline |
| `ledgerguard-flutter/lib/models/subscription_model.dart` | ~80 | Subscription data model |
| `ledgerguard-flutter/lib/mock_data/mock_subscriptions.dart` | ~100 | Mock subscription data |
| `ledgerguard-flutter/lib/widgets/lg_risk_badge.dart` | ~50 | Color-coded risk state badge |

## Data Flow
```
MockSubscriptions → SubscriptionProvider (filters) → SubscriptionListScreen
                                                          │
                                                    tap subscription
                                                          │
                                                          ▼
                                                 SubscriptionDetailScreen
                                                    ├── Subscription info
                                                    ├── Payment history tab
                                                    └── Risk timeline tab
```

## Configuration
None — mock data.

## Widget Tree
```
SubscriptionListScreen
├── LgPage (title: "Subscriptions")
│   ├── LgSearchField (search by domain/plan)
│   ├── Row: Filter chips (Status, Risk, Plan, App)
│   └── ListView.builder
│       └── ListTile per subscription
│           ├── Shop domain + plan name
│           ├── Price (formatted)
│           └── LgRiskBadge (risk state)

SubscriptionDetailScreen
├── LgPage (title: shop domain)
│   ├── LgCard: Subscription info (plan, price, status, dates)
│   ├── LgCard: Risk Assessment (state, days past due)
│   ├── LgCard: Payment History (list of charges)
│   └── LgCard: Risk Timeline (state changes over time)
```

## State Machine
```
SubscriptionProvider (ChangeNotifier)
  State:
    _searchQuery: String           → '' (default)
    _statusFilter: SubscriptionStatus?  → null (all)
    _riskFilter: RiskState?        → null (all)
    _planFilter: String?           → null (all)
    _appFilter: String?            → null (all)

  Events:
    setSearch(query)    → filter by domain/plan
    setStatusFilter()   → filter by ACTIVE/CANCELLED/etc
    setRiskFilter()     → filter by SAFE/ONE_CYCLE/etc
    setPlanFilter()     → filter by plan name
    clearFilters()      → reset all filters

  Computed:
    subscriptions → filtered mock list
    getById(id)   → single subscription lookup
    paymentHistory(id) → mock payment records
    riskTimeline(id)   → mock risk state changes
```

## Gotchas
- Search is case-insensitive, matches on `shopDomain` and `planName`
- Filters are AND-combined (all active filters must match)
- `getById()` catches exceptions silently, returns null on not found
- Payment history and risk timeline come from separate mock data files
- RiskState enum: `safe`, `oneCycleMissed`, `twoCyclesMissed`, `churned`

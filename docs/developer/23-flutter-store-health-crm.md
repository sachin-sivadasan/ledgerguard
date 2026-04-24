# 23. Flutter Store Health & CRM

## What It Does
Provides a CRM-style view of Shopify stores using your app. The store list shows all merchants with their subscription status, revenue, and brand data (logo, country). The detail view shows a store's subscription, transaction timeline, earnings breakdown, and CRM notes. Supports search and app filtering.

## Architecture
Presentation layer. `StoreProvider` manages store list filtering. `StoreListScreen` renders searchable store cards. `StoreDetailScreen` shows per-store health with subscription, transactions, and earnings sections. Store data comes from mock data including brand info (logos, country, description).

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/providers/store_provider.dart` | ~60 | Store filtering and lookup |
| `ledgerguard-flutter/lib/screens/stores/store_list_screen.dart` | ~150 | Searchable store list |
| `ledgerguard-flutter/lib/screens/stores/store_detail_screen.dart` | ~300 | Store health detail view |
| `ledgerguard-flutter/lib/models/store_model.dart` | ~50 | Store model (domain, logo, country, revenue) |
| `ledgerguard-flutter/lib/mock_data/mock_stores.dart` | ~80 | Mock store data with brand info |

## Data Flow
```
MockStores → StoreProvider (search + app filter) → StoreListScreen
                                                        │
                                                   tap store
                                                        │
                                                        ▼
                                                StoreDetailScreen
                                                  ├── Brand header (logo, name, country)
                                                  ├── Subscription status card
                                                  ├── Transaction timeline
                                                  └── Earnings breakdown
```

## Configuration
None — mock data.

## Widget Tree
```
StoreListScreen
├── LgPage (title: "Stores")
│   ├── LgSearchField
│   ├── App filter dropdown
│   └── ListView.builder
│       └── LgCard per store
│           ├── CircleAvatar (store logo)
│           ├── Shop domain + country flag
│           ├── Subscription count
│           └── LgRiskBadge

StoreDetailScreen
├── LgPage (title: store domain)
│   ├── LgCard: Brand Header (logo, name, country, description)
│   ├── LgCard: Active Subscription (plan, price, risk state)
│   ├── LgCard: Transaction Timeline (list of charges)
│   └── LgCard: Earnings Summary (pending, available, paid out)
```

## State Machine
```
StoreProvider (ChangeNotifier)
  State:
    _searchQuery: String   → '' (default)
    _selectedAppId: String? → null (all apps)

  Events:
    setSearch(query)       → filter stores by domain
    setSelectedApp(appId)  → filter by installed app

  Computed:
    stores    → filtered list (by search + app)
    getById() → single store lookup
    subscriptionsForStore() → subscriptions linked to store
```

## Gotchas
- Stores are linked to apps via `installedAppIds` list (a store can have multiple apps)
- Store search matches on `shopDomain` only (case-insensitive)
- Brand data (logo, country) comes from the `shops` table in production
- Transaction timeline on detail screen pulls from mock transactions filtered by domain

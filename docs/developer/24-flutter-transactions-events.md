# 24. Flutter Transactions & Events

## What It Does
Displays financial transactions (charges, refunds), app lifecycle events (installs, uninstalls, plan changes), and webhook delivery logs. Three separate screens cover the three data types, each with filtering, search, and detail views. Provides the audit trail view for revenue tracking.

## Architecture
Presentation layer. Three providers (`TransactionProvider`, `EventsProvider`, `WebhookProvider`) each manage their own filter state and mock data. Screens render filterable lists with type badges and amount formatting.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `ledgerguard-flutter/lib/providers/transaction_provider.dart` | ~60 | Transaction filtering (type, search) |
| `ledgerguard-flutter/lib/providers/events_provider.dart` | ~60 | Event filtering (type, date range) |
| `ledgerguard-flutter/lib/providers/webhook_provider.dart` | ~50 | Webhook log filtering |
| `ledgerguard-flutter/lib/screens/transactions/transactions_screen.dart` | ~200 | Transaction list with type filter |
| `ledgerguard-flutter/lib/screens/events/events_screen.dart` | ~200 | Events timeline |
| `ledgerguard-flutter/lib/screens/webhooks/webhooks_screen.dart` | ~150 | Webhook delivery log |
| `ledgerguard-flutter/lib/models/transaction_model.dart` | ~50 | Transaction model |
| `ledgerguard-flutter/lib/models/event_model.dart` | ~40 | Event model |
| `ledgerguard-flutter/lib/models/webhook_model.dart` | ~30 | Webhook model |

## Data Flow
```
MockTransactions → TransactionProvider → TransactionsScreen
MockEvents       → EventsProvider      → EventsScreen
MockWebhooks     → WebhookProvider     → WebhooksScreen
```

## Configuration
None — mock data.

## Widget Tree
```
TransactionsScreen
├── LgPage (title: "Transactions")
│   ├── LgSearchField
│   ├── Filter chips: ChargeType (Recurring, Usage, OneTime, Refund)
│   └── ListView.builder
│       └── ListTile per transaction
│           ├── Shop domain
│           ├── Amount (color-coded: green=credit, red=refund)
│           ├── LgBadge (charge type)
│           └── Date

EventsScreen
├── LgPage (title: "Events")
│   ├── Filter chips: EventType (Install, Uninstall, Cancel, etc.)
│   └── ListView.builder
│       └── ListTile per event
│           ��── Event description
│           ├── Shop domain
│           ├── LgBadge (event type)
│           └── Date
```

## State Machine
```
TransactionProvider (ChangeNotifier)
  State:
    _searchQuery: String
    _typeFilter: ChargeType?
  Events:
    setSearch(), setTypeFilter(), clearFilters()
  Computed:
    transactions → filtered mock list

EventsProvider (ChangeNotifier)
  State:
    _typeFilter: EventType?
    _appFilter: String?
  Events:
    setTypeFilter(), setAppFilter(), clearFilters()
  Computed:
    events → filtered mock list

WebhookProvider (ChangeNotifier)
  State:
    _statusFilter: WebhookStatus?
  Events:
    setStatusFilter()
  Computed:
    webhooks → filtered mock list
```

## Gotchas
- ChargeType enum: `recurring`, `usage`, `oneTime`, `refund`
- EventType enum: `appInstall`, `appUninstall`, `subscriptionCancelled`, `billingFailure`, `planUpgrade`, `planDowngrade`
- Amounts displayed in dollars (cents / 100), refunds shown as negative with red color
- Events screen doubles as the activity log referenced from the dashboard

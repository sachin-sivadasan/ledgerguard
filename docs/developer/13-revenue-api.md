# 13. Revenue API

## What It Does
Provides revenue timeline data and earnings availability status for frontend charts and dashboards. The service aggregates transaction data by date and charge type, supporting two display modes: combined (single total per day) and split (broken down by subscription vs. usage revenue). It also exposes an earnings status endpoint that summarizes how much revenue is pending, available, or already paid out, including per-date breakdowns of pending amounts and upcoming availability over the next 30 days.

## Architecture
Two layers collaborate to serve revenue data:

- **Application layer** — `RevenueMetricsService` (`internal/application/service/`) handles business logic: date validation, mode selection, response formatting, and earnings status aggregation.
- **HTTP handler layer** — `RevenueHandler` (`internal/interfaces/http/handler/`) handles authentication, parameter parsing, app ID resolution (numeric Shopify GID to internal UUID), and JSON serialization.

The service depends on two repository interfaces:
- `RevenueRepository` — aggregates transactions by date range, returning `RevenueAggregation` records with total, subscription, and usage amounts.
- `TransactionRepository` — provides earnings summary (pending/available/paid out totals), pending-by-date breakdowns, and upcoming availability projections.

Two constructors exist: `NewRevenueMetricsService` (revenue timeline only) and `NewRevenueMetricsServiceWithTransactions` (adds earnings status support).

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/interfaces/http/handler/revenue_handler.go` | RevenueHandler: GetEarnings, GetEarningsStatus, app ID lookup, JSON error responses |
| `backend/internal/application/service/revenue_metrics_service.go` | RevenueMetricsService: GetEarningsByDateRange, GetEarningsStatus, response DTOs |
| `backend/internal/domain/repository/revenue_repository.go` | RevenueRepository interface: GetRevenueByDateRange, RevenueAggregation struct |

## Data Flow

### Earnings Timeline
```
GET /api/v1/apps/{appID}/earnings?start=2024-01-01&end=2024-12-31&mode=split
│
├── Authenticate user via Firebase middleware
│
├── Parse appID from URL path
│     └── lookupAppByNumericID(userID, numericID)
│           ├── Find partner account by userID
│           ├── Find all apps for partner account
│           └── Match by extracting numeric ID from Shopify GID
│                 (e.g., "gid://partners/App/12345" → "12345")
│
├── Parse query params: start, end (YYYY-MM-DD), mode (combined|split)
│     └── Missing or invalid → 400 Bad Request
│
├── RevenueMetricsService.GetEarningsByDateRange(appID, start, end, mode)
│     ├── Clamp end date to today (no future dates)
│     ├── Validate start <= end → else ErrInvalidDateRange
│     ├── revenueRepo.GetRevenueByDateRange(appID, start, end)
│     │     └── Returns []RevenueAggregation (date, total, subscription, usage cents)
│     └── Convert to EarningsTimelineResponse
│           ├── mode=combined → TotalAmountCents only
│           └── mode=split → TotalAmountCents + SubscriptionAmountCents + UsageAmountCents
│
└── Return JSON response
```

### Earnings Status
```
GET /api/v1/apps/{appID}/earnings/status
│
├── Authenticate + resolve app ID (same as above)
│
├── RevenueMetricsService.GetEarningsStatus(appID)
│     ├── Requires transactionRepo (else ErrTransactionRepoRequired)
│     ├── transactionRepo.GetEarningsSummary(appID)
│     │     └── Returns: PendingCents, AvailableCents, PaidOutCents
│     ├── transactionRepo.GetPendingByAvailableDate(appID)
│     │     └── Returns: [{date, amountCents}, ...] grouped by available date
│     └── transactionRepo.GetUpcomingAvailability(appID, 30 days)
│           └── Returns: [{date, amountCents}, ...] for next 30 days
│
└── Return EarningsStatusResponse JSON
```

## Configuration
| Setting | Value | Notes |
|---------|-------|-------|
| Date format | `YYYY-MM-DD` | Used for both input parsing and output formatting |
| Default mode | `combined` | If `mode` query param is omitted or not "split" |
| Upcoming availability window | 30 days | Hardcoded as `upcomingAvailabilityDays` constant |
| Future date clamping | End date clamped to today | Prevents queries for dates that have no data |

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/apps/{appID}/earnings` | Firebase | Revenue timeline for date range. Query params: `start` (required), `end` (required), `mode` (optional: combined/split) |
| GET | `/api/v1/apps/{appID}/earnings/status` | Firebase | Earnings availability breakdown: pending, available, paid out, with per-date detail |

### Response: Earnings Timeline
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "earnings": [
    {
      "date": "2024-01-15",
      "total_amount_cents": 125000,
      "subscription_amount_cents": 100000,
      "usage_amount_cents": 25000
    }
  ]
}
```
Note: `subscription_amount_cents` and `usage_amount_cents` are only present when `mode=split`.

### Response: Earnings Status
```json
{
  "total_pending_cents": 50000,
  "total_available_cents": 200000,
  "total_paid_out_cents": 1500000,
  "pending_by_date": [
    { "date": "2024-04-20", "amount_cents": 25000 },
    { "date": "2024-04-22", "amount_cents": 25000 }
  ],
  "upcoming_availability": [
    { "date": "2024-04-20", "amount_cents": 25000 }
  ]
}
```

## Read Model Population

The Revenue API's read model tables (`api_subscription_status`, `api_usage_status`) are populated by `ReadModelBuilder` after every ledger rebuild. This runs as a non-fatal post-sync step in both sync paths:

- **Queue path:** `TransactionProcessor` calls `ReadModelBuilder.RebuildForApp()` after ledger rebuild
- **Direct path:** `SyncService` calls `ReadModelBuilder.RebuildForApp()` after ledger rebuild

If read model population fails, the sync still succeeds — errors are logged but not propagated.

**Admin rebuild endpoint:** `POST /api/v1/admin/apps/{appID}/rebuild-read-model` allows manual triggering without re-running the full sync. Requires ADMIN role.

## External Revenue API (API Key Authenticated)

The Revenue API provides external, API-key authenticated endpoints for querying subscription and usage status. All endpoints that accept a Shopify GID also accept a plain numeric ID (suffix match). For example, `28262727727` matches `gid://partners/AppSubscription/28262727727`.

### Subscription Endpoints
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/subscriptions/{id}` | API Key | Get subscription by numeric ID or full GID |
| GET | `/api/v1/subscriptions/status?domain={domain}` | API Key | Get subscription by myshopify domain |
| POST | `/api/v1/subscriptions/batch` | API Key | Batch lookup. Body: `{"ids": ["12345", "67890"]}` — accepts numeric IDs or full GIDs |

### Usage Endpoints
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/usages/{id}` | API Key | Get single usage by numeric ID or full GID |
| GET | `/api/v1/usages?subscription_id={id}` | API Key | Get all usages for a subscription (numeric ID or full GID) |
| POST | `/api/v1/usages/batch` | API Key | Batch lookup. Body: `{"ids": ["12345", "67890"]}` — accepts numeric IDs or full GIDs |

### ID Format Rule
Anywhere a GID is accepted, both formats work:
- **Numeric:** `28262727727` (recommended — no URL encoding needed)
- **Full GID:** `gid://partners/AppSubscription/28262727727` (requires URL encoding in path params)

## Extension Points
- **Granularity options** — add `granularity` query param (daily, weekly, monthly, yearly) to control aggregation level. Currently the repository returns per-date records.
- **Charge type filtering** — add a `charge_type` query param to filter to RECURRING, USAGE, ONE_TIME, or REFUND only.
- **Currency conversion** — the service currently returns cents in the transaction's original currency. Multi-currency apps would need conversion to a base currency.
- **Caching** — revenue data for past dates is immutable (ledger rebuild is deterministic). Past-date queries could be cached indefinitely; only today's data changes.
- **CSV/PDF export** — add content negotiation or a separate endpoint to export earnings timelines in non-JSON formats.

## Gotchas
- **App ID is numeric, not UUID.** The `{appID}` path parameter is the numeric suffix of the Shopify GID (e.g., `12345` from `gid://partners/App/12345`), not the internal UUID. The handler performs a lookup to resolve it, which involves loading all apps for the partner account and iterating.
- **EarningsStatus requires TransactionRepository.** If the service was constructed with `NewRevenueMetricsService` (without transactions), calling `GetEarningsStatus` returns `ErrTransactionRepoRequired`. The handler returns a 500 error in this case, which is misleading.
- **No pagination.** The earnings timeline returns all records in the date range. A query spanning multiple years could return a large response. There is no limit or cursor-based pagination.
- **Future dates are silently clamped.** If `end` is in the future, it is quietly set to today. The response's `end_date` field reflects the clamped date, but there is no explicit warning.
- **Error responses return 200 for some internal errors.** The `writeJSONErrorResponse` helper sets the correct status code, but any error from `json.NewEncoder(w).Encode()` on the success path is silently dropped.
- **extractNumericID scans from the end.** The GID parser finds the last `/` and returns everything after it. This works for standard GIDs but would fail for a GID with a trailing slash.

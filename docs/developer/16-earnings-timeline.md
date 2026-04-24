# 16. Earnings Timeline

## What It Does
Tracks when transaction earnings become available for payout, based on Shopify's hold period. Every transaction goes through a lifecycle: PENDING (earnings held by Shopify), AVAILABLE (released for payout), and PAID_OUT (disbursed to the developer). The `EarningsCalculator` domain service determines the available date for each transaction, updates statuses as time passes, and provides aggregate summaries of earnings by status.

The key business rule: Shopify holds earnings for 7 days after the charge is created (minimum hold period). Refunds are an exception and are processed immediately with no hold period.

## Architecture
Domain service layer (`internal/domain/service/`). The `EarningsCalculator` is a pure, stateless struct with zero external dependencies. It operates entirely on `Transaction` entities, reading and mutating their earnings tracking fields (`CreatedDate`, `AvailableDate`, `EarningsStatus`).

The `Transaction` entity (`internal/domain/entity/`) carries the earnings tracking fields and provides convenience methods (`IsPending`, `IsAvailable`, `IsPaidOut`, `SetEarningsTracking`, `UpdateEarningsStatus`). The `EarningsStatus` type is defined as a string constant in the entity package (PENDING, AVAILABLE, PAID_OUT).

The `Transaction` entity is intentionally rich: it holds both the financial data (gross/net amounts, fees) and the earnings lifecycle state. `SetEarningsTracking` is the single method for initializing all three earnings fields atomically.

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/domain/service/earnings_calculator.go` | EarningsCalculator: ProcessTransaction, ProcessTransactions, CalculateAvailableDate, DetermineEarningsStatus, SummarizeEarnings, UpdateStatuses |
| `backend/internal/domain/entity/transaction.go` | Transaction entity: EarningsStatus type (PENDING, AVAILABLE, PAID_OUT), SetEarningsTracking, UpdateEarningsStatus, IsPending, IsAvailable, IsPaidOut |
| `backend/internal/application/service/revenue_metrics_service.go` | RevenueMetricsService.GetEarningsStatus: serves earnings availability data via the API |
| `backend/internal/interfaces/http/handler/revenue_handler.go` | RevenueHandler.GetEarningsStatus: HTTP handler for GET /api/v1/apps/{appID}/earnings/status |

## Data Flow

### Processing a Single Transaction
```
EarningsCalculator.ProcessTransaction(tx, createdDate, now)
│
├── CalculateAvailableDate(tx.ChargeType, createdDate)
│     ├── ChargeType == REFUND → availableDate = createdDate (immediate)
│     └── All others (RECURRING, USAGE, ONE_TIME) → availableDate = createdDate + 7 days
│
├── DetermineEarningsStatus(availableDate, now)
│     ├── now < availableDate → PENDING
│     └── now >= availableDate → AVAILABLE
│
└── tx.SetEarningsTracking(createdDate, availableDate, status)
      ├── tx.CreatedDate = createdDate
      ├── tx.AvailableDate = availableDate
      └── tx.EarningsStatus = status
```

### Batch Processing
```
EarningsCalculator.ProcessTransactions(txs, now)
│
└── For each transaction:
      ├── Determine createdDate:
      │     ├── tx.CreatedDate is set → use it
      │     └── tx.CreatedDate is zero → fallback to tx.TransactionDate
      │
      └── ProcessTransaction(tx, createdDate, now)
```

### Summarize Earnings
```
EarningsCalculator.SummarizeEarnings(txs)
│
└── For each transaction, group by EarningsStatus:
      ├── PENDING   → PendingCents += tx.NetAmountCents,   PendingCount++
      ├── AVAILABLE → AvailableCents += tx.NetAmountCents, AvailableCount++
      └── PAID_OUT  → PaidOutCents += tx.NetAmountCents,   PaidOutCount++

Returns EarningsSummary with TotalCents() and TotalCount() convenience methods
```

### Update Statuses (Batch Time-Based)
```
EarningsCalculator.UpdateStatuses(txs, now)
│
└── For each transaction:
      tx.UpdateEarningsStatus(now)
      │
      ├── EarningsStatus == PAID_OUT → skip (immutable once paid)
      ├── now < AvailableDate → set PENDING
      └── now >= AvailableDate → set AVAILABLE
```

## Configuration
| Setting | Value | Notes |
|---------|-------|-------|
| `DefaultEarningsDelayDays` | 7 | Minimum Shopify hold period. Shopify actually holds 7-37 days depending on charge type and billing cycle, but 7 is used for MVP simplicity. |
| Refund delay | 0 days | Refunds are immediately available (same day as creation) |

### Earnings Status Lifecycle
```
PENDING ──[7 days pass]──> AVAILABLE ──[Shopify disburses]──> PAID_OUT
                                                                  │
                                                          (immutable, never
                                                           transitions back)
```

Note: The PENDING to AVAILABLE transition is time-based and handled by `UpdateStatuses`. The AVAILABLE to PAID_OUT transition is not automated -- it must be set externally when Shopify confirms disbursement.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/apps/{appID}/earnings/status` | Firebase | Earnings availability: totals (pending/available/paid out) + per-date breakdowns + 30-day upcoming availability |

### Response Format
```json
{
  "total_pending_cents": 50000,
  "total_available_cents": 200000,
  "total_paid_out_cents": 1500000,
  "pending_by_date": [
    { "date": "2024-04-20", "amount_cents": 25000 }
  ],
  "upcoming_availability": [
    { "date": "2024-04-20", "amount_cents": 25000 }
  ]
}
```

The `EarningsCalculator` itself is not exposed via HTTP. It is called during the sync/rebuild pipeline to set earnings fields on transactions before they are persisted. The API layer reads the pre-computed fields from the database.

## Extension Points
- **Variable hold periods per charge type.** Replace the flat 7-day delay with charge-type-specific delays (e.g., 7 days for RECURRING, 14 days for ONE_TIME, 30 days for USAGE). The `CalculateAvailableDate` switch already branches on `ChargeType`.
- **PAID_OUT transition automation.** Integrate with Shopify's payout API or Partner API to detect when earnings are actually disbursed and automatically transition from AVAILABLE to PAID_OUT.
- **Earnings forecasting.** Using the `upcoming_availability` data pattern, extend to forecast future earnings based on active subscription renewal dates.
- **Hold period per region.** Shopify's hold periods can vary by developer region. The calculator could accept a hold-period configuration rather than using a global constant.
- **Earnings notifications.** Trigger a notification when a large batch of earnings transitions from PENDING to AVAILABLE (e.g., "Your $5,000 in earnings are now available for payout").

## Gotchas
- **7-day hold is an approximation.** Shopify actually holds earnings for 7 to 37 days depending on the charge type, the app's payout schedule, and the developer's billing cycle. The `DefaultEarningsDelayDays = 7` is the optimistic minimum. Earnings may show as AVAILABLE in LedgerGuard before Shopify actually releases them.
- **PAID_OUT is a terminal state.** `UpdateEarningsStatus` explicitly skips transactions with `EarningsStatus == PAID_OUT`. Once a transaction is marked PAID_OUT, it cannot transition back to PENDING or AVAILABLE, even if `UpdateStatuses` is called with an earlier time.
- **CreatedDate fallback to TransactionDate.** In `ProcessTransactions`, if `tx.CreatedDate` is the zero value, the code falls back to `tx.TransactionDate`. These dates may differ (CreatedDate is when the charge was created in Shopify; TransactionDate is when the transaction was recorded). Using the wrong date shifts the available date by the difference.
- **SummarizeEarnings uses NetAmountCents.** The summary counts net amounts (after Shopify fees), not gross amounts. This is correct for "what the developer receives" but may be confusing if compared against gross revenue metrics elsewhere in the dashboard.
- **No persistence in the calculator.** `EarningsCalculator` mutates Transaction entities in memory but does not save them. The caller (typically the sync service) is responsible for persisting the updated transactions after processing.
- **UpdateStatuses is not idempotent for PAID_OUT.** While the method correctly skips PAID_OUT transactions, if called on a transaction that was externally marked PAID_OUT and then the available date has not passed yet, the status remains PAID_OUT (correct). But there is no validation that PAID_OUT should only be set for transactions where `now >= AvailableDate`.
- **Refund amounts are typically negative.** When a refund transaction is summarized, its `NetAmountCents` is negative, which reduces the total for whatever status it is in. A $-50 refund marked as AVAILABLE will decrease `AvailableCents` by 50, which is correct accounting but may look odd in isolation.

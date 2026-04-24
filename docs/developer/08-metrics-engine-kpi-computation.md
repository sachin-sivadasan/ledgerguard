# 08. Metrics Engine & KPI Computation

## What It Does
Computes all key performance indicators (KPIs) from subscription and transaction data. Calculates Active MRR, Revenue at Risk, Usage Revenue, Total Revenue, and Renewal Success Rate. Produces a `DailyMetricsSnapshot` that serves as the immutable audit record for each day.

## Architecture
Domain service layer (`internal/domain/service/`). Pure Go with zero external dependencies. Called by `LedgerService` at the end of every ledger rebuild. The MetricsEngine is stateless — all computation is from inputs passed in, not from stored state.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/domain/service/metrics_engine.go` | ~129 | KPI computation logic |
| `backend/internal/domain/service/metrics_engine_test.go` | ~100 | Unit tests for all KPI formulas |
| `backend/internal/domain/entity/daily_metrics_snapshot.go` | ~68 | Snapshot entity with SetMetrics() |
| `backend/internal/application/service/metrics_aggregation_service.go` | ~200 | Aggregates snapshots for API responses |

## Data Flow
```
┌──────────────┐     ┌──────────────┐
│ Subscriptions│     │ Transactions │
│ (rebuilt)    │     │ (12-month)   │
└──────┬───────┘     └──────┬───────┘
       │                    │
       ▼                    ▼
┌──────────────────────────────────┐
│         MetricsEngine            │
│                                  │
│ 1. CalculateActiveMRR()          │  ← SAFE subs only
│ 2. CalculateRevenueAtRisk()      │  ← at-risk subs
│ 3. CalculateUsageRevenue()       │  ← USAGE tx only
│ 4. CalculateTotalRevenue()       │  ← all tx types
│ 5. CalculateRenewalSuccessRate() │  ← SAFE / total
│ 6. Count by risk state           │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│    DailyMetricsSnapshot          │
│    (upsert to PostgreSQL)        │
└──────────────────────────────────┘
```

1. LedgerService calls `ComputeAllMetrics(appID, subscriptions, transactions, now)`
2. MetricsEngine calculates each KPI independently
3. Results stored in `DailyMetricsSnapshot` entity via `SetMetrics()`
4. Snapshot upserted to `daily_metrics_snapshots` table (ON CONFLICT update)

## Configuration
No configuration needed. All computation is deterministic from inputs.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/v1/metrics/{appId} | Firebase | Current and historical KPIs |
| GET | /api/v1/metrics/{appId}?range=30d | Firebase | Metrics for time range |

## Extension Points
- Add new KPI methods to MetricsEngine (e.g., Net Revenue Retention, LTV)
- MetricsAggregationService can add new aggregation windows (weekly, quarterly)
- DailyMetricsSnapshot entity can be extended with new fields (add migration)

## Gotchas
- **MRR includes only SAFE subscriptions** — at-risk revenue is tracked separately as Revenue at Risk
- **Annual subscriptions**: MRR = BasePriceCents / 12 (monthly equivalent)
- **Renewal Success Rate** returns 0 (not NaN) when there are no subscriptions
- **Total Revenue** formula: RECURRING + USAGE + ONE_TIME - REFUNDS (refunds are subtracted)
- **Usage Revenue** is tracked separately from MRR — never mixed
- Snapshot `TotalSubscriptions` is computed as sum of all risk state counts

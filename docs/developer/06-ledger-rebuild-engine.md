# 06. Ledger Rebuild Engine

## What It Does
Deterministically rebuilds all subscription state from raw transactions. Given the same set of transactions, the engine always produces the same subscriptions, risk classifications, and metrics snapshot. This is the heart of LedgerGuard's financial integrity guarantee: there are no incremental updates, no delta calculations, and no mutable state carried forward between syncs. Every rebuild starts from scratch.

The engine groups transactions by myshopify domain, constructs one subscription per domain from recurring charge history, classifies risk for each subscription, computes all KPIs, and stores a daily metrics snapshot.

## Architecture
Domain service layer (`internal/domain/service/`). The `LedgerService` struct depends only on repository interfaces (TransactionRepository, SubscriptionRepository) and the `MetricsEngine` domain service. An optional `DailyMetricsSnapshotRepository` can be attached via `WithSnapshotRepository()` for metrics persistence.

The rebuild follows the strategy defined in ADR-002:
1. Fetch all transactions for the app (12-month window)
2. Group by domain (one subscription per store)
3. Delete all existing subscriptions for the app
4. Insert freshly rebuilt subscriptions
5. Compute and upsert daily metrics snapshot

This delete-and-reinsert approach ensures no stale subscriptions survive across rebuilds.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/domain/service/ledger_service.go` | ~376 | Core rebuild logic, subscription construction, revenue separation |
| `backend/internal/domain/service/ledger_service_test.go` | ~514 | Determinism tests, edge cases, billing interval detection |
| `backend/internal/domain/service/metrics_engine.go` | ~129 | KPI computation called at end of rebuild |
| `backend/internal/domain/entity/subscription.go` | ~137 | Subscription entity with ClassifyRisk(), MRRCents() |
| `backend/internal/domain/valueobject/billing_interval.go` | — | MONTHLY vs ANNUAL interval with NextChargeDate() |

## Data Flow
```
RebuildFromTransactions(appID, now)
│
├── 1. txRepo.FindByAppID(appID, now-12months, now)
│       └── Returns all stored transactions
│
├── 2. groupTransactionsByDomain(transactions)
│       └── map[myshopify_domain] → []*Transaction
│
├── 3. rebuildSubscriptions(appID, byDomain, now)
│       │
│       └── for each domain:
│             buildSubscriptionFromTransactions()
│             │
│             ├── Sort transactions by date (oldest first)
│             ├── Filter to RECURRING charges only
│             ├── Get most recent recurring transaction
│             ├── Detect billing interval (MONTHLY vs ANNUAL)
│             ├── Assign GID (Shopify real or synthetic lg_sub_*)
│             ├── Set base price from GrossAmountCents (or NetAmountCents fallback)
│             ├── Set status from transaction (default ACTIVE)
│             ├── UpdateFromRecurringCharge() → sets ExpectedNextChargeDate
│             ├── Override with SubscriptionPeriodEnd if available
│             └── ClassifyRisk(now) → sets RiskState
│
│       Sort subscriptions by domain (deterministic order)
│
├── 4. subRepo.DeleteByAppID(appID)         ← Wipe existing
├── 5. subRepo.Upsert(sub) for each         ← Insert rebuilt
│       └── Accumulate MRR (ACTIVE only) and risk counts
│
├── 6. sumUsageRevenue(transactions)         ← USAGE charges only
│
└── 7. metrics.ComputeAllMetrics() → snapshotRepo.Upsert()
        └── Daily snapshot with all KPIs

Returns LedgerRebuildResult {
  SubscriptionsUpdated, TotalMRRCents, TotalUsageCents,
  RiskSummary, Snapshot
}
```

### BackfillHistoricalSnapshots Flow
```
BackfillHistoricalSnapshots(appID, transactions)
│
├── Find earliest and latest transaction dates
│
└── For each month in range:
      ├── Filter transactions up to end-of-month
      ├── Rebuild subscriptions at that point in time
      ├── Filter transactions within that month (for revenue)
      ├── ComputeAllMetrics() for that month's snapshot date
      └── snapshotRepo.Upsert()   (ON CONFLICT update)
```

## Configuration
No environment variables needed. The engine is purely computational — all behavior is determined by the transactions and the current timestamp passed in.

| Parameter | Source | Description |
|-----------|--------|-------------|
| `appID` | Caller | Which app to rebuild |
| `now` | Caller | Current time for risk classification and snapshot dating |
| 12-month window | Hardcoded | `now.AddDate(-1, 0, 0)` — always fetches one full year |

## API Surface
The LedgerService is not exposed directly via HTTP. It is called internally by:

| Caller | Method | When |
|--------|--------|------|
| `SyncService.SyncApp()` | `RebuildFromTransactions()` | After every transaction sync |
| `SyncService.SyncApp()` | `BackfillHistoricalSnapshots()` | After rebuild, for historical data |

The rebuild results surface through the sync API response and the metrics endpoints:

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/metrics/{appId}` | Firebase | Current snapshot from latest rebuild |
| POST | `/api/v1/sync` | Firebase | Triggers rebuild (returns risk summary and MRR) |

## Extension Points
- **WithSnapshotRepository()** — attach snapshot persistence. Without it, metrics are computed but not stored.
- **Add new subscription fields** — extend `buildSubscriptionFromTransactions()` to extract additional data from transactions (e.g., plan tier, discount codes).
- **SeparateRevenue()** — public method that splits transactions into RECURRING and USAGE streams. Can be used by reporting services for revenue breakdowns.
- **Custom billing interval detection** — `detectBillingInterval()` uses average days between charges. Override with explicit `BillingInterval` from transaction metadata when available.
- **Additional snapshot types** — the monthly backfill pattern could be extended to weekly or quarterly snapshots.

## Gotchas
- **Delete-and-reinsert on every rebuild.** `subRepo.DeleteByAppID()` wipes all subscriptions before inserting rebuilt ones. This means subscription UUIDs change on every sync. Do not store foreign keys to subscription IDs from external systems.
- **Domains without RECURRING transactions produce no subscription.** If a store only has USAGE or ONE_TIME charges, `buildSubscriptionFromTransactions()` returns nil and no subscription is created for that domain.
- **Billing interval detection threshold is 180 days.** If the average days between recurring charges exceeds 180, the subscription is classified as ANNUAL. With fewer than 2 recurring transactions, it defaults to MONTHLY.
- **GrossAmountCents vs NetAmountCents.** The base price uses `GrossAmountCents` (what the customer pays) with a fallback to `NetAmountCents` (after Shopify's cut). If neither is set, the subscription price will be 0.
- **Synthetic subscription GIDs** use the format `lg_sub_<SHA1-UUID>` derived from the domain name. These are deterministic (same domain always produces the same synthetic GID) but clearly distinguishable from real Shopify GIDs.
- **MRR accumulates only from ACTIVE subscriptions** during the rebuild loop. At-risk subscriptions contribute to the RiskSummary counts but not to TotalMRRCents.
- **BackfillHistoricalSnapshots does not create future snapshots.** Snapshot dates are capped at `time.Now().UTC()` even if transaction data extends to a future period end date.
- **Deterministic output order.** Subscriptions are sorted by `MyshopifyDomain` before insertion to ensure the same input always produces the same output order, which is critical for reproducible testing.

# 09. Daily Snapshots

## What It Does
Stores an immutable daily record of all KPIs for each tracked app. Snapshots are upserted (one per app per day) and never deleted, forming a permanent audit trail. Used for trend analysis, AI insights, and revenue reconciliation.

## Architecture
Domain entity (`internal/domain/entity/`) with persistence adapter (`internal/infrastructure/persistence/`). Snapshots are created by MetricsEngine during ledger rebuild and stored via the DailyMetricsSnapshotRepository interface (port/adapter pattern). Historical backfill creates monthly snapshots from past transactions.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/domain/entity/daily_metrics_snapshot.go` | ~68 | Snapshot entity with SetMetrics() |
| `backend/internal/domain/repository/daily_metrics_snapshot_repository.go` | ~15 | Repository interface (port) |
| `backend/internal/infrastructure/persistence/daily_metrics_snapshot_repository.go` | ~100 | PostgreSQL implementation (adapter) |
| `backend/internal/domain/service/ledger_service.go` | ~376 | Creates snapshots during rebuild + backfill |
| `backend/migrations/000006_create_daily_metrics_snapshot_table.up.sql` | ~20 | Table creation with UNIQUE constraint |

## Data Flow
```
┌────────────────┐
│ LedgerService  │
│ RebuildFrom-   │
│ Transactions() │
└───────┬────────┘
        │ ComputeAllMetrics()
        ▼
┌────────────────┐
│ MetricsEngine  │
│                │
│ Returns:       │
│ DailyMetrics-  │
│ Snapshot       │
└───────┬────────┘
        │
        ▼
┌────────────────────────────────────┐
│ SnapshotRepository.Upsert()        │
│                                    │
│ INSERT INTO daily_metrics_snapshots │
│ ON CONFLICT (app_id, date)         │
│ DO UPDATE SET ...                  │
└────────────────────────────────────┘
```

1. LedgerService completes subscription rebuild
2. Calls `MetricsEngine.ComputeAllMetrics()` to create snapshot
3. Snapshot upserted via repository — same day overwrites, new day inserts
4. BackfillHistoricalSnapshots() creates one snapshot per month for historical data

## Configuration
No configuration needed. Snapshots are created automatically during every sync.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/v1/metrics/{appId} | Firebase | Latest snapshot + trend data |
| GET | /api/v1/metrics/{appId}?from=2025-01&to=2025-12 | Firebase | Historical snapshots |

## Extension Points
- Add new metric fields to DailyMetricsSnapshot entity (requires migration)
- Create weekly/monthly rollup snapshots for faster trend queries
- Add snapshot comparison endpoint (period-over-period)

## Gotchas
- **Never delete snapshots** — they are permanent audit records
- **Upsert semantics**: `ON CONFLICT (app_id, date) DO UPDATE` means re-running sync for the same day updates the existing snapshot
- **Date truncation**: Snapshot date is always truncated to start of day (midnight UTC)
- **Backfill creates monthly snapshots** (end-of-month), not daily, to avoid excessive records for historical data
- **Backfill uses all transactions up to snapshot date** for subscription state, but only current month transactions for revenue calculation
- Backfill errors are silently ignored during sync — non-critical operation

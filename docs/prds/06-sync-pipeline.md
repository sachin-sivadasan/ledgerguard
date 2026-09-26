# PRD: Sync & Ledger Rebuild (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/application/service/sync_service.go`, `infrastructure/queue/` (worker, lock,
> recovery, progress, processors/), `application/scheduler/`, `handler/queue_sync.go`,
> `domain/service/ledger_service.go`, `frontend-flutter/lib/services/sync_status_service.dart`.
> Reference: `docs/blueprints/SYNC_PIPELINE_BLUEPRINT.md`, CLAUDE.md §8.

## Problem & evidence

LedgerGuard's entire value depends on an accurate, trustworthy mirror of a partner's Shopify Partner-API financials. Raw transactions must be pulled reliably, and every derived number (MRR, risk, KPIs) must be **reproducible from those transactions** — surviving crashes, retries, and re-runs without double-counting or drift.

- **FACT:** the pipeline is one of the most engineered parts of the system (7 job types, wave orchestration, distributed locking, two-tier crash recovery) with broad test coverage — it's treated as foundational.
- **UNKNOWN — original demand evidence not recorded**, but the determinism/idempotency mandate is explicit product policy (CLAUDE.md §8).

## Target users

**INFERRED:** indirectly every user — sync feeds all other capabilities. Directly, the **partner** who connects an app and watches it populate, and **operators** who need the sync to self-heal.
- **Not the target:** real-time/streaming financial feeds; users expecting sub-minute freshness (sync is batch — see Non-goals).

## Proposed solution (as-built behavior)

**FACT — triggering & status.** A sync is enqueued on app-connect or via a manual action: `POST /api/v1/sync/enqueue/{appID}?type=full` → **202** with `job_id` (`queue_sync.go:39-117`); a duplicate active job → **409** (`queue_sync_service.go:67-72`). The Flutter `SyncStatusService` polls `/api/v1/sync/jobs?status=processing` and shows **queued → running → done** with a progress fraction `completed_items/total_items`; for a `full_sync` the furthest-along child drives the bar (`sync_status_service.dart:50-87`).

**FACT — wave orchestration** (`full_sync_processor.go`): **Wave 1** (parallel) `transaction_sync` + `review_sync`; wait for transactions; **Wave 2** (parallel) `event_sync` + `status_sync` + `store_sync`; **Wave 3** `snapshot_sync` — run **after** status reconciliation so Dashboard KPIs match Risk/Subscriptions (RISK-1). If any child fails, the parent is `partial_failure`, not `completed` (`:186-201`).

**FACT — ledger rebuild (CLAUDE.md §8 verified).** Full syncs fetch **entire history from `SyncHistoryStart` = 2010-01-01** (`ledger_service.go:16-24`); catch-up syncs fetch a **`LookbackDays` delta** (`transaction_processor.go:84-90`). Raw transactions are stored **immutably** (`UpsertBatch`). `RebuildFromTransactions` **deletes all old subscriptions and rebuilds the set de novo** from the full transaction history (`ledger_service.go:71-101`) — no incremental updates. Same transactions → same ledger (deterministic + idempotent).

**FACT — reliability.** Redis distributed lock (SETNX, 2h TTL, ownership-aware Lua for release/extend/steal, `lock.go`); heartbeats; **two-tier recovery** (startup re-enqueues heartbeat-less `processing` jobs; periodic re-enqueues jobs stale > lock TTL, with a 2-min grace, `recovery.go`); re-enqueue with 5s backoff on lock contention; guarded status transitions for idempotency.

**FACT — scheduling.** Periodic full sync every **12h** (`sync_scheduler.go:30`); daily **catch-up at UTC hour 3** with `lookbackDays=2` (`daily_catchup_scheduler.go:37-141`).

### Key screens

**No new wireframe** (as-built). Real surface: sync status/progress in the stores/apps screens via `frontend-flutter/lib/services/sync_status_service.dart` (polling `/api/v1/sync/jobs`).

## Platform & policy constraints

**FACT/INFERRED.** Bound to the **Shopify Partner API** (version support-window, [[shopify-partner-api-gotchas]]) and its rate limits. **Requires Redis** for queue/lock/heartbeat/progress — a hard runtime dependency (no in-process fallback). Runs in Docker Compose on the shared Hetzner box ([[hetzner-migration]]). CLAUDE.md §8 mandates deterministic + idempotent rebuild.

## Pricing-tier impact

**UNKNOWN — no tiered sync limits found** (e.g. frequency by plan). Open question: should sync frequency/history depth be tiered? — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Rebuild is idempotent; re-running a sync reproduces the ledger. A logic change reflects on the next sync with no migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Determinism:** two consecutive full syncs on unchanged upstream data produce identical ledger + snapshots (byte-identical KPIs) in 100% of checks.
2. **Self-healing:** ≥ 99% of jobs interrupted by a worker crash reach a terminal state (completed/partial_failure) without manual intervention, within one recovery interval.
3. **Freshness SLA:** ≥ 95% of apps have data no older than the catch-up window (≤ ~27h given hour-3 + 2-day lookback) — ties dashboard trust to sync.

## Non-goals (deliberately absent in code)

1. **No incremental ledger updates** — every sync rebuilds from scratch (CLAUDE.md §8). **FACT.**
2. **No real-time/streaming sync** — batch (12h periodic + daily catch-up + manual). **FACT.**
3. **No automatic retry of `partial_failure`** — it's one-shot; recovery re-runs only crash-interrupted jobs, not failed-child parents. **FACT.**
4. **No Redis-less operation** — Redis is required, not optional. **FACT.**

## Open questions

- **`partial_failure` has no auto-retry** — should a failed child trigger a bounded retry, or stay manual? Owner: Eng/Product.
- **Catch-up gap:** an outage > 2 days leaves a transaction gap until a manual full sync — should catch-up widen its lookback adaptively? Owner: Eng.
- **Shopify 429/backpressure:** no queue-level rate-limit backpressure visible; relies on the client. Is that sufficient at scale? Owner: Eng.
- **Parent `0/0` progress edge:** if all children fail to enqueue, the bar shows 0% until timeout. Owner: Eng.
- Tiered sync frequency/history by plan? Owner: Product.
- Read-model (Revenue API) rebuild has no rollback if it fails mid-sync — acceptable divergence window? Owner: Eng.

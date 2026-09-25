# Tech Design: Sync & Ledger Rebuild (as-built)

**PRD:** docs/prds/06-sync-pipeline.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** This is the system's data-integrity core. The job-state machine (`sync_jobs`), the immutable `transactions` store, and the delete-and-rebuild of `subscriptions` are all hard to undo if wrong — a rebuild bug corrupts every downstream number. The public sync API + Redis contract are also load-bearing. Highest-stakes design in the backfill.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE. Reference: `docs/blueprints/SYNC_PIPELINE_BLUEPRINT.md`, CLAUDE.md §8.

## Current state

**FACT — enqueue & jobs.** `POST /api/v1/sync/enqueue/{appID}?type=full` inserts a `sync_jobs` row (pending) and LPUSHes to Redis, returning **202** with `job_id`; a duplicate active job → **409** (`queue_sync.go:39-117`, `queue_sync_service.go:67-72`). Status/progress/list/cancel endpoints exist (`queue_sync.go`). Two Redis queues: regular + orchestrator (full_sync only).

**FACT — worker.** `WorkerPool` BRPOPs, acquires a distributed lock, runs a heartbeat loop, invokes the mapped processor, and centralizes terminal state transitions (`worker.go:76-199`). Lock contention → re-enqueue with 5s backoff.

**FACT — wave orchestration** (`full_sync_processor.go`): Wave 1 (parallel) `transaction_sync` + `review_sync`; wait for transactions; Wave 2 (parallel) `event_sync` + `status_sync` + `store_sync`; Wave 3 `snapshot_sync` (after status reconciliation, RISK-1). Any failed child → parent set to `partial_failure` via `UpdateStatus` directly (bypassing the worker's MarkCompleted, `:186-201`).

**FACT — ledger rebuild** (`ledger_service.go:71-109`): reads the app's **entire** transaction history from `SyncHistoryStart` (2010-01-01), groups by domain, rebuilds subscriptions de novo, **`DeleteByAppID` then `Upsert`** the rebuilt set. Deterministic (same transactions → same subscriptions). Catch-up path fetches a `LookbackDays` delta (`transaction_processor.go:84-90`).

**FACT — reliability.** Lock = Redis SETNX, 2h TTL, ownership-aware Lua for release/extend/steal (`lock.go`). Recovery (`recovery.go`) = startup + periodic re-enqueue of heartbeat-less `processing` jobs with a 2-min grace and a conditional `MarkPendingIfProcessing` to avoid racing live workers. Schedulers: 12h periodic full sync (`sync_scheduler.go:30`), daily catch-up at UTC hour 3, lookback 2 days (`daily_catchup_scheduler.go:37-141`).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **orchestrator + fan-out child jobs over a Redis-backed queue, with classify/rebuild done as a full deterministic recompute, and two-tier crash recovery.** End-to-end happy path + recovery: see `docs/designs/06-sync-pipeline-sequence.puml` (validated `plantuml -checkonly`). Components: handler → enqueue service → Redis → worker/lock → full-sync processor → child processors → ledger service → Postgres.

### Data model

**FACT — reads/writes existing tables, no new schema for this backfill.** `sync_jobs` (job state machine: pending/processing/completed/failed/cancelled/partial_failure, parent/child, progress counters); `transactions` (immutable raw, `UpsertBatch`); `subscriptions` (deleted + rebuilt each sync); `daily_metrics_snapshot` (written Wave 3). Redis holds queues, locks, heartbeats, and a progress overlay. **Migration:** none — rebuild is idempotent.

### API & events

**FACT.** `POST /sync/enqueue/{appID}` → 202 `{job_id,job_type,status,message}` / 409. `GET /sync/jobs/{jobID}` → job entity. `GET /sync/jobs/{jobID}/progress` → `{Job,Total,Completed,Message,Children[]}` with a Redis overlay while processing. `GET /sync/jobs?app_id&status&job_type&limit&offset`. `POST /sync/jobs/{jobID}/cancel` → sets a Redis cancel flag, cascades to children. No external webhooks emitted.
**Breaking-change check:** the `sync_jobs.status` enum + progress shape are consumed by the Flutter `SyncStatusService`; renames break polling.

## Alternatives considered

- **Full rebuild vs incremental updates.** As-built rebuilds from scratch every sync (CLAUDE.md §8, `ledger_service.go:71`). Incremental was **explicitly rejected** by policy (determinism/idempotency, auditability). **Choose incremental only if** rebuild cost becomes prohibitive at scale (accepting drift risk).
- **Redis queue + custom worker vs a managed job system (e.g. Cloud Tasks / Temporal).** As-built is a hand-rolled Redis queue with Lua locks + recovery (blueprint). Given the Hetzner move off GCP ([[hetzner-migration]]), a self-hosted Redis fits. **Choose a managed system if** operational burden of recovery/locking outweighs the control.
- **Wave ordering (snapshot last).** Snapshot deliberately runs after status reconciliation (RISK-1) so KPIs converge — a recorded rationale, not arbitrary. Alternative (snapshot inline) would reintroduce the convergence bug.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Duplicate concurrent sync | active job check | **FACT** 409 (`queue_sync_service.go:67-72`) | client waits |
| Lock contention | SETNX fails | **FACT** re-enqueue + 5s backoff (`worker.go:202-216`) | retried |
| Worker dies mid-job | no heartbeat | **FACT** startup+periodic recovery re-enqueues (2-min grace) (`recovery.go`) | auto self-heal |
| Child job fails | parent scans children | **FACT** parent → `partial_failure` (`full_sync_processor.go:186-201`) | **manual/next sync only** |
| Stale lock after crash | Lua ownership check | **FACT** steal/extend guarded atomically (`lock.go`) | safe |
| Transient upstream fetch error | processor returns err | **FACT** job → failed; parent → partial_failure | manual re-run |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. `partial_failure` has no auto-retry.** A single failed child leaves the parent partial; recovery only re-runs crash-interrupted jobs, not failed ones. → open question (bounded retry?).
- **D2. Redis is a hard dependency.** No in-process fallback for queue/lock/heartbeat/progress; a Redis outage stalls all sync until it returns (recovery then catches interrupted jobs). → operational risk to document/monitor.
- **D3. Rebuild is not transactionally atomic.** `DeleteByAppID` then a per-sub `Upsert` loop (`ledger_service.go:98-109`) is **not wrapped in one DB transaction** — a crash between delete and full re-insert leaves a partially-rebuilt subscription set until the next sync. Idempotent eventually, but a transient-inconsistency window that reads (dashboard/risk) can observe. → proposed test + open question (wrap in a txn?).
- **D4. Catch-up gap.** An outage > 2 days exceeds the lookback; transactions in the gap aren't fetched until a manual full sync. → open question (adaptive lookback).
- **D5. No visible Shopify 429 backpressure** at the queue level; relies on the client. → verify client behavior under rate limits.
- **D6. Parent `0/0` progress edge.** If all children fail to enqueue, the parent progress shows 0% until timeout. → minor UX/robustness fix.
- **D7. Read-model (Revenue API) rebuild has no rollback** if it fails mid-sync → ledger vs read-model divergence window. → open question.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Determinism: run two full syncs on a fixed transaction fixture → identical rebuilt subscriptions + snapshots (metric #1). Extend `sync_service_test.go`.
2. Self-healing: kill a worker mid-job → recovery drives it to terminal within one interval (metric #2). Extend `recovery_test.go`.

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- Lock acquire/release/ownership — `lock_test.go`. processJob sequence/backoff/heartbeat — `worker_test.go`. Startup+periodic recovery, parent-child skip — `recovery_test.go`. Wave orchestration + partial-failure detection — `processors_test.go`. Progress dual-write throttle — `progress_test.go`. Duplicate detection — covered in queue tests.

**Adversarial — DIVERGENCE TODO (none exist today):**
- **D3 crash-between-delete-and-rebuild** (assert reads never see empty subs, or wrap in txn) — **highest priority**. D1 partial_failure retry policy · D4 >2-day gap · D5 429 backpressure · D7 read-model rollback. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/infrastructure/queue/... -v`
2. `curl -X POST -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/sync/enqueue/<appID>?type=full"` → expect 202 + job_id; repeat immediately → expect 409.
3. Poll `GET /api/v1/sync/jobs/<jobID>/progress | jq` through waves; confirm terminal state.
4. Kill the worker container mid-sync; restart; confirm recovery re-enqueues and the job completes.

**External-platform reality:** the Shopify Partner API is the external dependency; unit tests fake the fetcher below the client. Rate-limit/429 behavior and true full-history volume MUST be smoke-tested against a real Partner account — never fully reproducible in unit tests. Redis is real in integration tests.

## Pricing & policy touchpoints

**Platform:** Shopify Partner API version support-window + rate limits ([[shopify-partner-api-gotchas]]). **UNKNOWN** whether sync frequency/history is tiered by plan (no code). → Open question (Product).

## Rollout

**N/A (as-built, deployed on Hetzner via Docker Compose, [[hetzner-migration]]).** Deploy = SSH box, git pull, `docker compose ... up -d --build`, then prune build cache/images. Because rebuild is idempotent, a logic change simply reflects on the next sync; rollback = revert + redeploy; next sync restores correct state. No data migration.

## Observability

**FACT (partial):** rich `log.Printf` through rebuild (transaction counts, charge-type breakdown, domain/subscription counts, `ledger_service.go:81-96`) and recovery; job state + progress persisted in `sync_jobs` + Redis overlay. **DIVERGENCE:** no metric/alert on partial_failure rate, recovery frequency, per-app sync latency, or catch-up gaps. On-call greps Hetzner container logs (no sync dashboard/index). → proposed observability TODO tied to metrics #2/#3.

## Open questions

- **D3 make rebuild atomic** (wrap delete+reinsert in a txn?) — Owner: Eng. **(highest priority)**
- D1 auto-retry policy for partial_failure. Owner: Eng/Product.
- D4 adaptive catch-up lookback after long outages. Owner: Eng.
- D5 confirm Shopify 429 backpressure in the client. Owner: Eng.
- D2/D7 Redis-outage + read-model divergence runbooks. Owner: Eng.
- Tiered sync by plan? Owner: Product.

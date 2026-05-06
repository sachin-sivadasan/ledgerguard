# 31. Queue-Based Async Sync System

## What It Does
Provides an asynchronous, Redis-backed job queue for running sync operations in the background. Instead of blocking the HTTP response while fetching 12 months of Shopify data, the system immediately enqueues a job (HTTP 202) and processes it via worker pools. A `full_sync` orchestrator decomposes the work into independent child jobs (transactions, events, reviews) and dependent child jobs (snapshots, status, stores), enabling granular progress tracking visible in the Flutter frontend.

Also auto-triggers a `full_sync` on app selection during onboarding so users see data within minutes.

## Architecture
The queue system sits between the HTTP layer and the existing domain services. It does not replace the synchronous sync (`POST /sync/{appID}`) — both coexist. When `cfg.Queue.Enabled=true`, the `SyncScheduler` (12h interval) is disabled in favor of queue-based sync.

**Dual Storage:** Job metadata lives in PostgreSQL (permanent audit trail); live progress counters live in Redis hashes (ephemeral, high-frequency). The frontend queries the DB for job status and overlays Redis progress for any job in `processing` state.

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/infrastructure/queue/client.go` | Redis client factory |
| `backend/internal/infrastructure/queue/queue.go` | Enqueue (LPUSH) / Dequeue (BRPOP) |
| `backend/internal/infrastructure/queue/lock.go` | Distributed locks + heartbeats (SETNX, TTL) |
| `backend/internal/infrastructure/queue/progress.go` | Dual-write progress tracker (Redis hash + periodic DB flush) |
| `backend/internal/infrastructure/queue/recovery.go` | Startup + periodic recovery of stuck jobs |
| `backend/internal/infrastructure/queue/processor.go` | SyncProcessor interface + ProcessorRegistry |
| `backend/internal/infrastructure/queue/processor_context.go` | Shared preamble (lock, heartbeat, progress init) |
| `backend/internal/infrastructure/queue/worker.go` | WorkerPool (goroutine pool with BRPOP loop) |
| `backend/internal/infrastructure/queue/processors/transaction_processor.go` | Fetches transactions + ledger rebuild |
| `backend/internal/infrastructure/queue/processors/snapshot_processor.go` | Backfills historical monthly snapshots |
| `backend/internal/infrastructure/queue/processors/event_processor.go` | Fetches app lifecycle events |
| `backend/internal/infrastructure/queue/processors/status_processor.go` | Enriches subscription status from events |
| `backend/internal/infrastructure/queue/processors/store_processor.go` | Fetches shop brand/logo data |
| `backend/internal/infrastructure/queue/processors/review_processor.go` | Scrapes app store reviews |
| `backend/internal/infrastructure/queue/processors/full_sync_processor.go` | Orchestrator — dispatches child jobs in waves |
| `backend/internal/domain/entity/sync_job.go` | SyncJob entity + status constants |
| `backend/internal/domain/repository/sync_job_repository.go` | Repository interface |
| `backend/internal/infrastructure/persistence/sync_job_repository.go` | PostgreSQL implementation |
| `backend/internal/application/service/queue_sync_service.go` | Application service (enqueue, progress, cancel, TriggerSync) |
| `backend/internal/interfaces/http/handler/queue_sync.go` | HTTP handler (5 endpoints) |
| `backend/internal/interfaces/http/handler/app.go` | SyncTrigger interface + auto-sync on app selection |
| `backend/cmd/server/main.go` | Full wiring: Redis, processors, worker pools, recovery, SetSyncTrigger |

## Data Flow

### Enqueue Flow
```
POST /api/v1/sync/enqueue/{appID}?type=full_sync
  │
  ▼
QueueSyncHandler.EnqueueSync()
  │
  ├── Validate tenant ownership (appID belongs to user's partner account)
  ├── Check for duplicate active job (same appID + type)
  ├── INSERT INTO sync_jobs (status='pending', priority=N)
  ├── LPUSH job JSON to Redis queue (full_sync → :queue:full, others → :queue)
  └── Return HTTP 202 { job_id: "..." }
```

### Worker Processing Flow
```
WorkerPool goroutine (BRPOP loop)
  │
  ├── BRPOP from Redis queue
  ├── Deserialize SyncJobPayload
  ├── AcquireLock(appID, jobType) — Redis SETNX with workerID
  │     └── If lock fails: check holder's heartbeat → StealLock if dead, else backoff + re-enqueue
  ├── Write initial heartbeat (prevents steal race)
  ├── MarkStarted (WHERE status='pending' — conditional)
  ├── Start heartbeat goroutine (renew every 10 min, extend lock every 1h)
  ├���─ Look up processor from ProcessorRegistry
  ├── Processor.Process(ctx, payload)
  │     └── Update Redis progress every ~2s
  │     └── Flush to DB every ~30s
  │
  ├── On success: MarkCompleted (WHERE status='processing')
  ├── On error + ctx cancelled (shutdown): leave in 'processing' for recovery
  ├── On error + cancelled flag: skip MarkFailed (already cancelled)
  ├── On error (genuine): MarkFailed
  └── Cleanup: ReleaseLockIfOwner + DeleteHeartbeat + CleanupCancellation
```

### Full Sync Orchestration
```
full_sync_processor starts
  │
  │  Wave 1 — No subscription dependency
  ├── transaction_sync  → regular queue (fetches transactions + rebuilds ledger + rebuilds read model)
  ├── review_sync       → regular queue (scrapes app store reviews)
  │
  │  Poll transaction_sync until complete (5s interval)
  │
  │  Wave 2 — Depends on subscriptions from ledger rebuild
  ├── event_sync        → regular queue (fetches events per shop GID)
  ├── snapshot_sync     → regular queue (backfills monthly snapshots)
  ├── status_sync       → regular queue (enriches subscription status)
  ├── store_sync        → regular queue (fetches shop brand/logo)
  │
  │  Poll all children until all complete/fail
  └── Mark parent completed (or partial_failure)
```

### Auto-Sync on App Selection
```
POST /api/v1/apps/select
  │
  ├── Create app entity
  ├── appRepo.Create()
  │
  ▼
SyncTrigger.TriggerSync(appID, userID, partnerAccountID)
  │
  ├── Queue mode: EnqueueSync(type=full_sync, priority=1)
  └── Direct mode: go SyncApp(appID) (fire-and-forget)
```

## Job Types
| Job Type | Queue | What It Does | Progress Unit |
|----------|-------|-------------|---------------|
| `full_sync` | `:queue:full` | Orchestrator — dispatches 6 child jobs | children completed / 6 |
| `transaction_sync` | `:queue` | Fetch 12-month transactions + ledger rebuild + read model rebuild | transactions fetched |
| `event_sync` | `:queue` | Fetch app lifecycle events + store raw | events fetched |
| `snapshot_sync` | `:queue` | Backfill historical monthly snapshots | months / 12 |
| `status_sync` | `:queue` | Enrich subscription status from events | subscriptions enriched |
| `store_sync` | `:queue` | Fetch shop brand/logo for new domains | shops fetched |
| `review_sync` | `:queue` | Scrape App Store reviews (2 pages max) | pages scraped |

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `QUEUE_ENABLED` | `false` | No | Enable Redis queue system (disables SyncScheduler) |
| `REDIS_ADDR` | `localhost:6379` | When queue enabled | Redis connection address |
| `REDIS_PASSWORD` | — | No | Redis password |
| `REDIS_DB` | `0` | No | Redis database number |
| `QUEUE_NUM_WORKERS` | `3` | No | Regular queue worker pool size |
| `QUEUE_FULL_SYNC_WORKERS` | `1` | No | Full sync queue worker pool size |

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/sync/enqueue/{appID}?type=full&priority=normal` | Firebase | Enqueue sync job → 202 with job_id |
| GET | `/api/v1/sync/jobs/{jobID}` | Firebase | Job status from DB |
| GET | `/api/v1/sync/jobs/{jobID}/progress` | Firebase | Progress with Redis overlay + children |
| GET | `/api/v1/sync/jobs?app_id={appID}&status=&type=&limit=&offset=` | Firebase | Paginated job history |
| POST | `/api/v1/sync/jobs/{jobID}/cancel` | Firebase | Cooperative cancellation |

## State Machine
```
pending → processing → completed
                    → failed
                    → cancelled (cooperative)
                    → partial_failure (full_sync with some child failures)
```

Terminal states: `completed`, `failed`, `cancelled`, `partial_failure`.

## Worker Processing Flow
1. **BRPOP** from queue (blocking dequeue)
2. **AcquireLock** — SETNX with workerID as value (ownership tracking)
3. If lock fails: check heartbeat of holder → StealLock if dead, else backoff + re-enqueue
4. **Initial heartbeat** — written immediately after lock (prevents steal race before goroutine starts)
5. **MarkStarted** — conditional: `WHERE status='pending'` (prevents double-start)
6. **Heartbeat goroutine** — renews heartbeat every 10 min + `ExtendLockIfOwner` every 1h
7. **Process** — delegate to processor (returns nil on success)
8. **State transition** — centralized in worker:
   - Success → `MarkCompleted`
   - Error + ctx cancelled (shutdown) → leave in `processing` (recovery re-enqueues on restart)
   - Error + `IsCancelled` flag → skip (already cancelled by user)
   - Error (genuine) → `MarkFailed`
9. **Cleanup** — `ReleaseLockIfOwner` + delete heartbeat + delete cancellation flag

## Recovery
- **Startup recovery:** On server boot, any jobs stuck in `processing` (no heartbeat) are re-enqueued. Pending jobs are pushed to Redis without touching locks. **Child job dedup:** if a child's parent is also being recovered, the child is skipped and marked failed — the parent will recreate it.
- **Graceful shutdown:** Worker detects `ctx.Err() != nil` → skips MarkFailed → leaves job in `processing`. On next boot, startup recovery re-enqueues immediately (no heartbeat = stale).
- **Periodic recovery:** Every 10 minutes, check for orphaned processing jobs. **Grace period:** skip jobs started less than 2 minutes ago. **Threshold:** jobs older than `lockTTL` (2h) with no heartbeat are re-enqueued.
- **Conditional re-enqueue:** `MarkPendingIfProcessing` uses `WHERE status='processing'` — if a worker already completed the job, recovery does nothing.
- **Heartbeat:** Workers renew a Redis key every 10 min (TTL 20 min). If a worker crashes, the TTL expires and recovery picks up the job.
- **Lock cleanup:** Recovery uses `ForceReleaseLock` (unconditional DEL) since it's definitively clearing stale state from dead workers.

## Extension Points
- **New processor:** Create a file in `processors/`, implement `SyncProcessor` interface, register in `main.go` — the worker pool picks it up automatically.
- **New queue priority:** Add a new Redis list key and a dedicated `WorkerPool` with configurable concurrency.
- **SyncTrigger interface:** Any service implementing `TriggerSync(ctx, appID, userID, partnerAccountID)` can be wired into `AppHandler` for auto-sync.

## Multi-Node Deployment

The queue system is designed for multiple server instances sharing Redis + PostgreSQL:

- **Lock ownership:** Each lock stores the workerID as its Redis value. Only the owner can release/extend.
- **Fan-out:** Any worker on any node can BRPOP from the shared queue. Redis handles fair distribution.
- **Cross-node recovery:** Recovery on Node A can detect and clean up stale locks from crashed Node B (using `ForceReleaseLock`).
- **No single-node assumptions:** heartbeats, locks, and queues are all in shared Redis.

**Key safety guarantees:**
- `ReleaseLockIfOwner` prevents Worker A from releasing Worker B's lock after slow processing
- `ExtendLockIfOwner` prevents zombie goroutines from extending locks they no longer own
- `StealLock` is atomic (Lua script) — no race window between check and acquire
- DB status guards prevent concurrent workers from double-transitioning job state

## Gotchas
- **Dual-write progress:** Redis is the source of truth for live progress during processing; DB is flushed every ~30s. If Redis is flushed mid-job, progress display resets but the job continues.
- **Duplicate detection:** `EnqueueSync` rejects requests if an active job with the same `appID + jobType` already exists. The auto-sync trigger silently swallows this error.
- **Lock contention:** Only one worker can process a given `appID + jobType` at a time. Other workers skip and re-enqueue with backoff.
- **Lock ownership:** All lock operations (release, extend) verify the caller is the owner. Only recovery uses `ForceReleaseLock`.
- **Heartbeat race prevention:** Initial heartbeat is written synchronously after lock acquisition (before goroutine). Without this, another worker could check `HasHeartbeat` → false → steal the lock from a live worker.
- **Centralized MarkCompleted:** Processors must NOT call MarkCompleted themselves. Return nil for success; the worker handles state transitions.
- **Shutdown recovery:** On Ctrl+C, jobs are NOT marked failed — they stay `processing` so startup recovery re-enqueues them. This avoids permanent data loss from transient shutdowns.
- **Recovery child dedup:** When both parent (full_sync) and child are stale, only the parent is re-enqueued. Children are marked failed — the parent recreates them fresh.
- **Wave ordering:** event_sync is in Wave 2 (not Wave 1) because it iterates subscriptions which are created by transaction_sync's ledger rebuild.
- **SyncScheduler disabled:** When `cfg.Queue.Enabled=true`, the 12h scheduler does not start. Queue-based syncs must be triggered externally (API, frontend, or auto-trigger).
- **Priority:** `priority=1` is highest (used by auto-sync on app selection). Default is `priority=5`. The queue is sorted by priority on dequeue.
- **full_sync worker pool is separate** (1 worker by default) to prevent orchestrator jobs from blocking regular entity processors.
- **Cancellation check:** Before marking failed, worker checks `IsCancelled` — cancelled jobs retain their cancelled status.
- **worker_id column:** `NOT NULL DEFAULT ''` — MarkPendingIfProcessing resets to empty string, not NULL.

## Related
- [05-transaction-sync-engine.md](05-transaction-sync-engine.md) — Synchronous sync (still active, untouched)
- [QUEUE_SYNC_IMPLEMENTATION_PLAN.md](QUEUE_SYNC_IMPLEMENTATION_PLAN.md) — Original implementation plan (v3)
- [ADR-021](../../DECISIONS.md) — Separate snapshot_sync from transaction_sync
- [ADR-023](../../DECISIONS.md) — Cooperative cancellation over hard kill
- [ADR-024](../../DECISIONS.md) — SyncTrigger interface for auto-sync on app selection
- [ADR-026](../../DECISIONS.md) — Ownership-aware distributed locks (Lua scripts)
- [ADR-027](../../DECISIONS.md) — Centralized job state transitions in worker
- [33-queue-multi-node-deployment.puml](../diagrams/puml/33-queue-multi-node-deployment.puml) — Multi-node deployment diagram

---

## Daily Catch-Up Sync

### What It Does
A lightweight daily scheduler that re-syncs recent transactions and events for all active apps using a short lookback window (default 2 days). It fills gaps left by missed or failed full syncs — if the server was down or a sync errored out, the catch-up ensures no recent transaction data is lost.

Unlike a `full_sync` (which rebuilds the entire 12-month ledger and dispatches 6 child jobs), the daily catch-up only enqueues `transaction_sync` + `event_sync` with a narrow date window. This makes it fast and cheap — typically a few API pages per app instead of hundreds.

### How It Works

```
DailyCatchupScheduler (goroutine)
  │
  ├── Tick every 15 minutes
  ├── Check: current UTC hour == configured hour (default 3 AM)?
  ├── Check: lastRunDate != today?
  │
  │  If both true:
  ├── Fetch all active apps from AppRepository
  ├── For each app:
  │     ├── EnqueueCatchupSync(appID, userID, partnerAccountID, lookbackDays=2)
  │     │     ├── Enqueue transaction_sync with LookbackDays=2
  │     │     └── Enqueue event_sync with LookbackDays=2
  │     └── Skip silently if duplicate job already active
  └── Set lastRunDate = today (prevents double-run)
```

### LookbackDays Payload Field
`SyncJobPayload.LookbackDays` controls the transaction fetch window:
- `0` (default) — uses the standard 1-month window (backward-compatible)
- `> 0` — `TransactionProcessor` sets `since = now - LookbackDays` instead of the default window

This field is backward-compatible: existing jobs without it behave exactly as before.

### Key Files
| File | Purpose |
|------|---------|
| `backend/internal/application/scheduler/daily_catchup_scheduler.go` | Scheduler implementation (check interval, run logic, `RunOnce`) |
| `backend/internal/application/service/queue_sync_service.go` | `EnqueueCatchupSync()` — enqueues transaction + event jobs with lookback |
| `backend/internal/infrastructure/queue/queue.go` | `SyncJobPayload.LookbackDays` field |
| `backend/internal/infrastructure/queue/processors/transaction_processor.go` | Lookback-aware date window logic |
| `backend/internal/interfaces/http/handler/admin.go` | `TriggerDailyCatchup` admin handler |
| `backend/internal/interfaces/http/router/router.go` | Admin + internal route registration |

### API Endpoints
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/admin/sync/daily-catchup?lookback_days=2` | Firebase (admin) | Manual trigger with optional lookback override |
| POST | `/api/v1/internal/sync/daily-catchup` | Internal | Cloud Scheduler trigger (no auth, internal network only) |

### Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `DAILY_CATCHUP_HOUR` | `3` | UTC hour to run the daily catch-up (0-23) |
| `DAILY_CATCHUP_LOOKBACK_DAYS` | `2` | Default lookback window in days |

### Design Decisions
- **Follows `NotificationScheduler` pattern** — same 15-minute check interval, same `lastRunDate` guard, same `Start()`/`Stop()` lifecycle.
- **Only `transaction_sync` + `event_sync`** — snapshots, status, store, and review syncs are unnecessary for a 2-day window. Snapshots are daily aggregates computed from the full ledger; status/store/review data doesn't change frequently enough to warrant daily re-sync.
- **Duplicate-safe** — uses existing `EnqueueSync` duplicate detection. If an app already has an active `transaction_sync`, the catch-up silently skips it.
- **`RunOnce(ctx, lookbackDays)`** — enables admin-triggered catch-ups with custom windows (e.g., 7-day lookback after a prolonged outage).
- **Internal endpoint for Cloud Scheduler** — on Cloud Run, the in-process scheduler may not run reliably (scale-to-zero). The internal endpoint allows GCP Cloud Scheduler to trigger the catch-up via HTTP cron.

### Related
- [ADR-032](../../DECISIONS.md) — Daily catch-up sync with configurable lookback

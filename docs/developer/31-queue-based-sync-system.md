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
  ├── Look up processor from ProcessorRegistry
  │
  ▼
Processor.Process(ctx, payload)
  │
  ├── UPDATE sync_jobs SET status='processing', started_at=now
  ├── AcquireLock(appID, jobType) — Redis SETNX with TTL
  ├── Start heartbeat goroutine (renew lock every 10s)
  ├── Execute domain logic (fetch, transform, store)
  │     └── Update Redis progress every ~2s
  │     └── Flush to DB every ~30s
  ├── ReleaseLock + stop heartbeat
  └── UPDATE sync_jobs SET status='completed', completed_at=now
```

### Full Sync Orchestration
```
full_sync_processor starts
  │
  │  Wave 1 — Independent (enqueue immediately)
  ├── transaction_sync  → regular queue
  ├── event_sync        → regular queue
  ├── review_sync       → regular queue
  │
  │  Poll transaction_sync until complete (5s interval)
  │
  │  Wave 2 — Dependent (enqueue after transaction_sync)
  ├── snapshot_sync     → regular queue
  ├── status_sync       → regular queue
  ├── store_sync        → regular queue
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
| `transaction_sync` | `:queue` | Fetch 12-month transactions + ledger rebuild | transactions fetched |
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

## Recovery
- **Startup recovery:** On server boot, any jobs stuck in `processing` (stale heartbeat) are re-enqueued.
- **Periodic recovery:** Every 10 minutes, check for orphaned processing jobs and re-enqueue.
- **Heartbeat:** Workers renew a Redis key every 10s. If a worker crashes, the TTL expires and recovery picks up the job.

## Extension Points
- **New processor:** Create a file in `processors/`, implement `SyncProcessor` interface, register in `main.go` — the worker pool picks it up automatically.
- **New queue priority:** Add a new Redis list key and a dedicated `WorkerPool` with configurable concurrency.
- **SyncTrigger interface:** Any service implementing `TriggerSync(ctx, appID, userID, partnerAccountID)` can be wired into `AppHandler` for auto-sync.

## Gotchas
- **Dual-write progress:** Redis is the source of truth for live progress during processing; DB is flushed every ~30s. If Redis is flushed mid-job, progress display resets but the job continues.
- **Duplicate detection:** `EnqueueSync` rejects requests if an active job with the same `appID + jobType` already exists. The auto-sync trigger silently swallows this error.
- **Lock contention:** Only one worker can process a given `appID + jobType` at a time. Other workers skip and re-enqueue with backoff.
- **SyncScheduler disabled:** When `cfg.Queue.Enabled=true`, the 12h scheduler does not start. Queue-based syncs must be triggered externally (API, frontend, or auto-trigger).
- **Priority:** `priority=1` is highest (used by auto-sync on app selection). Default is `priority=5`. The queue is sorted by priority on dequeue.
- **full_sync worker pool is separate** (1 worker by default) to prevent orchestrator jobs from blocking regular entity processors.

## Related
- [05-transaction-sync-engine.md](05-transaction-sync-engine.md) — Synchronous sync (still active, untouched)
- [QUEUE_SYNC_IMPLEMENTATION_PLAN.md](QUEUE_SYNC_IMPLEMENTATION_PLAN.md) — Original implementation plan (v3)
- [ADR-021](../../DECISIONS.md) — Separate snapshot_sync from transaction_sync
- [ADR-023](../../DECISIONS.md) — Cooperative cancellation over hard kill
- [ADR-024](../../DECISIONS.md) — SyncTrigger interface for auto-sync on app selection

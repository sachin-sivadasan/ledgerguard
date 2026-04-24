# Queue-Based Sync System — Implementation Plan

> **Status:** Final Draft v3
> **Date:** 2026-04-21
> **Relates to:** [05-transaction-sync-engine.md](05-transaction-sync-engine.md)

---

## 1. Overview

Introduce an **async, queue-based sync system** alongside the existing synchronous sync. The existing `POST /api/v1/sync/{appID}` remains unchanged. A new set of endpoints, Redis queues, worker pools, and a `sync_jobs` PostgreSQL table enables non-blocking syncs with granular per-entity-type progress visible in the Flutter frontend.

### Design Principles

| Principle | How |
|-----------|-----|
| **Non-destructive** | Existing sync untouched; new system calls the same domain services |
| **Composable** | `full_sync` dispatches independent child jobs — does not duplicate logic |
| **Observable** | Job metadata + progress in PostgreSQL (permanent); live counters in Redis (ephemeral) |
| **Extensible** | Adding a new sync type = 1 processor file + register in main.go |
| **Scalable** | Distributed locks, Redis-based rate limiting, cluster-ready from day one |

### Storage Philosophy (Dual Storage)

| Data | Storage | Why |
|------|---------|-----|
| Job metadata (type, status, app_id, timestamps, error) | **PostgreSQL** | Permanent audit trail, queryable history, survives Redis flush |
| Live progress counters ("450/1000", "Fetching page 5") | **Redis hash** | High-frequency updates every ~2s, ephemeral, avoids DB write pressure |
| Queue (pending jobs waiting to be picked up) | **Redis list** | LPUSH/BRPOP for reliable FIFO delivery |
| Distributed locks + heartbeats | **Redis** | TTL-based, auto-expiring, cross-instance safe |

**Read path for frontend:** Query `sync_jobs` from DB → for any job with `status='processing'`, overlay live counters from Redis → return merged response.

---

## 2. Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                           HTTP Layer                                  │
│                                                                       │
│  POST /api/v1/sync/enqueue/{appID}?type=full      → Enqueue job      │
│  POST /api/v1/sync/enqueue/{appID}?type=transaction→ Enqueue          │
│  GET  /api/v1/sync/jobs/{jobID}                    → Job status (DB)  │
│  GET  /api/v1/sync/jobs/{jobID}/progress           → Progress (DB+R)  │
│  GET  /api/v1/sync/jobs?app_id={appID}             → List/history     │
│  POST /api/v1/sync/jobs/{jobID}/cancel             → Cancel job       │
│                                                                       │
│  (Existing endpoints unchanged)                                       │
│  POST /api/v1/sync/{appID}                         → Sync (old)       │
│  POST /api/v1/sync                                 → Sync all (old)   │
└───────────────────────────┬──────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────────┐
│                       Queue Sync Service                              │
│                                                                       │
│  EnqueueSync(appID, jobType, userID, partnerAccID)                    │
│    1. Validate tenant ownership                                       │
│    2. INSERT INTO sync_jobs (status='pending')                        │
│    3. LPUSH job JSON to Redis queue                                   │
│    4. Return jobID immediately (HTTP 202)                             │
└───────────────────────────┬──────────────────────────────────────────┘
                            │
             ┌──────────────┼───────────────────┐
             ▼              ▼                    ▼
    ┌──────────────┐ ┌──────────────┐  ┌────────────────┐
    │ Regular Queue │ │ Regular Queue│  │  Full Sync     │
    │  Worker 1     │ │  Worker 2    │  │  Queue Worker  │
    │ (transaction, │ │ (store,      │  │ (dispatches    │
    │  event, etc.) │ │  review,etc.)│  │  child jobs)   │
    └──────┬───────┘ └──────┬───────┘  └──────┬─────────┘
           │                │                  │
           ▼                ▼                  ▼
    ┌──────────────────────────────────────────────────────┐
    │                  Sync Processor                       │
    │                                                       │
    │  1. UPDATE sync_jobs SET status='processing'          │
    │  2. AcquireLock(appID, syncType) — Redis SETNX        │
    │  3. Start heartbeat goroutine (Redis, every 10min)    │
    │  4. Process — calls existing domain services           │
    │  5. Update Redis live progress every ~2s               │
    │  6. UPDATE sync_jobs SET completed_items (every ~30s)  │
    │  7. ReleaseLock + DeleteHeartbeat                      │
    │  8. UPDATE sync_jobs SET status='completed'            │
    └──────────────────────┬───────────────────────────────┘
                           │
                           ▼
    ┌──────────────────────────────────────────────────────┐
    │           Existing Domain Services                    │
    │  SyncService, LedgerService, RiskEngine — unchanged  │
    └──────────────────────────────────────────────────────┘
```

---

## 3. Job Types

| Job Type | Queue | What It Does | Progress Unit | Depends On |
|----------|-------|-------------|---------------|------------|
| `full_sync` | `:queue:full` | Orchestrator — dispatches 6 child jobs, tracks aggregate | children completed / total | — |
| `transaction_sync` | `:queue` | Fetch 12-month transactions + upsert + ledger rebuild | transactions fetched | — |
| `event_sync` | `:queue` | Fetch app lifecycle events (install/uninstall/charge) + store raw in `app_events` | events fetched | — |
| `snapshot_sync` | `:queue` | Backfill historical monthly snapshots from stored transactions | months backfilled / total | `transaction_sync` (needs rebuilt subscriptions + transactions) |
| `status_sync` | `:queue` | Enrich subscription status from lifecycle events | subscriptions enriched / total | `transaction_sync` (needs rebuilt subscriptions) |
| `store_sync` | `:queue` | Fetch shop brand/logo data for new domains | shops fetched / new domains | `transaction_sync` (needs subscriptions for domain list) |
| `review_sync` | `:queue` | Scrape Shopify App Store reviews (max 2 pages during sync) | pages scraped / max | — |

**Note:** `subscription_sync` is intentionally omitted — subscriptions are rebuilt as part of `transaction_sync` via `LedgerService.RebuildFromTransactions()`. There is no use case for rebuilding subscriptions without fetching transactions.

**Note:** `snapshot_sync` is intentionally **separate** from `transaction_sync` (see [ADR-021](../../DECISIONS.md#adr-021)). Previously snapshot backfill was hidden inside the transaction sync with no progress visibility. Separating it enables: independent recompute after metrics bug fixes, granular progress tracking ("months 3/12"), and cleaner single-responsibility processors.

### Full Sync Orchestration

```
full_sync job starts
  │
  │  Wave 1 — Independent jobs (enqueue immediately)
  ├─ 1. Enqueue transaction_sync  ──→ regular queue
  ├─ 2. Enqueue event_sync        ──→ regular queue
  ├─ 3. Enqueue review_sync       ──→ regular queue
  │
  │  (poll until transaction_sync completes — 5s interval)
  │
  │  Wave 2 — Dependent jobs (enqueue after transaction_sync done)
  ├─ 4. Enqueue snapshot_sync     ──→ regular queue
  ├─ 5. Enqueue status_sync       ──→ regular queue
  ├─ 6. Enqueue store_sync        ──→ regular queue
  │
  │  Poll all children until all complete/fail (5s interval)
  │  Update parent sync_jobs row: completed_children count
  │
  └─ Mark parent as completed (or partial_failure if any child failed)
```

**Dependency rationale:**
- `snapshot_sync` needs rebuilt subscriptions + transactions for monthly metric computation — only available after `transaction_sync`
- `status_sync` needs subscriptions with `ShopifyShopGID` — only available after ledger rebuild in `transaction_sync`
- `store_sync` needs subscription domain list — only available after ledger rebuild
- `event_sync` fetches raw Partner API events — independent of transactions
- `review_sync` scrapes App Store HTML — completely independent

---

## 4. Data Model

### 4.1 sync_jobs Table (PostgreSQL — Source of Truth)

```sql
CREATE TABLE sync_jobs (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id             UUID NOT NULL REFERENCES apps(id),
    user_id            UUID NOT NULL REFERENCES users(id),
    partner_account_id UUID NOT NULL REFERENCES partner_accounts(id),
    job_type           TEXT NOT NULL,             -- full_sync, transaction_sync, etc.
    parent_job_id      UUID REFERENCES sync_jobs(id), -- NULL if top-level, set for children
    status             TEXT NOT NULL DEFAULT 'pending',
                       -- pending | processing | completed | failed | cancelled | partial_failure
    priority           INT NOT NULL DEFAULT 0,    -- 0=normal, 1=high (force sync)
    total_items        INT NOT NULL DEFAULT 0,    -- total entities to process
    completed_items    INT NOT NULL DEFAULT 0,    -- entities processed so far
    entity_type        TEXT,                      -- transaction, store, subscription, review, event, snapshot
    error_message      TEXT,
    worker_id          TEXT,                      -- which worker picked this up
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_sync_jobs_app_id ON sync_jobs(app_id);
CREATE INDEX idx_sync_jobs_parent ON sync_jobs(parent_job_id);
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status) WHERE status IN ('pending', 'processing');
CREATE INDEX idx_sync_jobs_created ON sync_jobs(created_at DESC);

-- Composite: find active syncs for an app quickly
CREATE INDEX idx_sync_jobs_app_active ON sync_jobs(app_id, status)
    WHERE status IN ('pending', 'processing');
```

**Why `sync_jobs` in DB (not Redis)?**
- Queryable history: "show me last 20 syncs for this app"
- Analytics: "average sync duration by type"
- Survives Redis flush/restart
- Tenant isolation via SQL WHERE clause
- Parent/child relationships via foreign key
- Frontend can show sync history page without Redis

### 4.2 app_events Table (PostgreSQL — New Entity for event_sync)

```sql
CREATE TABLE app_events (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id           UUID NOT NULL REFERENCES apps(id),
    shop_domain      TEXT NOT NULL,
    shopify_shop_gid TEXT,
    event_type       TEXT NOT NULL,
                     -- RELATIONSHIP_INSTALLED, RELATIONSHIP_UNINSTALLED,
                     -- SUBSCRIPTION_CHARGE_ACCEPTED, SUBSCRIPTION_CHARGE_CANCELED
    occurred_at      TIMESTAMPTZ NOT NULL,
    raw_payload      JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(app_id, shop_domain, event_type, occurred_at)
);

CREATE INDEX idx_app_events_app_id ON app_events(app_id);
CREATE INDEX idx_app_events_type ON app_events(event_type);
CREATE INDEX idx_app_events_occurred ON app_events(occurred_at DESC);
```

### 4.3 Redis Keys (Ephemeral Only)

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `lg:sync:queue` | List | — | Regular job queue (LPUSH/BRPOP) |
| `lg:sync:queue:full` | List | — | Full sync job queue |
| `lg:sync:progress:{jobID}` | Hash | 1 hour | Live progress counters (total, completed, message) |
| `lg:sync:lock:{appID}:{syncType}` | String | 2 hours | Distributed lock (SETNX) |
| `lg:sync:heartbeat:{appID}:{syncType}` | String | 20 min | Worker liveness indicator |
| `lg:sync:cancel:{jobID}` | String | 1 hour | Cancellation flag (cooperative) |

**Redis is NOT the source of truth for job state.** If Redis is flushed:
- Queued jobs are lost → recovery reads `sync_jobs WHERE status='pending'` and re-enqueues
- Live progress lost → frontend shows DB progress (slightly stale but accurate)
- Locks lost → workers re-acquire on next heartbeat cycle

### 4.4 SyncJob Struct (Go — used for queue payload)

```go
type SyncJob struct {
    JobID          string `json:"job_id"`            // UUID, matches sync_jobs.id
    AppID          string `json:"app_id"`
    UserID         string `json:"user_id"`
    PartnerAccID   string `json:"partner_acc_id"`
    JobType        string `json:"job_type"`          // "full_sync", "transaction_sync", etc.
    ParentJobID    string `json:"parent_job_id"`     // empty if top-level
    Priority       int    `json:"priority"`          // 0=normal, 1=high
}
```

Minimal payload — workers read additional context from DB/config as needed.

---

## 5. API Endpoints

### 5.1 Enqueue Sync

```
POST /api/v1/sync/enqueue/{appID}
Auth: Firebase JWT
Query params:
  type     = full | transaction | status | store | review | event | snapshot  (default: full)
  priority = normal | high  (default: normal)

Response 202 Accepted:
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "job_type": "full_sync",
  "status": "pending",
  "message": "Sync job enqueued"
}

Error 409 Conflict (if same sync type already running for this app):
{
  "error": "transaction_sync already in progress for this app",
  "existing_job_id": "existing-uuid"
}
```

**Handler logic:**
1. Authenticate user via Firebase JWT
2. Verify user owns the app (tenant isolation)
3. Check `sync_jobs` for existing `pending`/`processing` job with same `app_id` + `job_type` → 409 if found
4. INSERT into `sync_jobs` with `status='pending'`
5. LPUSH to Redis queue
6. Return 202 with job_id

### 5.2 Job Status (from DB)

```
GET /api/v1/sync/jobs/{jobID}
Auth: Firebase JWT (tenant-scoped — user can only see their own jobs)

Response 200:
{
  "job_id": "550e8400-...",
  "app_id": "...",
  "job_type": "full_sync",
  "status": "processing",
  "priority": 0,
  "total_items": 0,
  "completed_items": 0,
  "entity_type": null,
  "error_message": null,
  "worker_id": "worker-regular-2",
  "created_at": "2026-04-21T10:00:00Z",
  "started_at": "2026-04-21T10:00:01Z",
  "completed_at": null
}
```

### 5.3 Job Progress (DB + Redis overlay — Frontend polls this)

```
GET /api/v1/sync/jobs/{jobID}/progress
Auth: Firebase JWT (tenant-scoped)
```

**Response for `full_sync`:**

```json
{
  "job_id": "parent-uuid",
  "job_type": "full_sync",
  "status": "processing",
  "started_at": "2026-04-21T10:00:01Z",
  "children": [
    {
      "job_id": "tx-001",
      "job_type": "transaction_sync",
      "status": "completed",
      "entity_type": "transaction",
      "total_items": 1200,
      "completed_items": 1200,
      "message": null,
      "started_at": "2026-04-21T10:00:02Z",
      "completed_at": "2026-04-21T10:00:35Z",
      "duration_seconds": 33
    },
    {
      "job_id": "ev-001",
      "job_type": "event_sync",
      "status": "completed",
      "entity_type": "event",
      "total_items": 85,
      "completed_items": 85,
      "message": null,
      "started_at": "2026-04-21T10:00:02Z",
      "completed_at": "2026-04-21T10:00:12Z",
      "duration_seconds": 10
    },
    {
      "job_id": "st-001",
      "job_type": "status_sync",
      "status": "processing",
      "entity_type": "subscription",
      "total_items": 200,
      "completed_items": 142,
      "message": "Enriching subscription 142/200",
      "started_at": "2026-04-21T10:00:36Z",
      "completed_at": null,
      "duration_seconds": null
    },
    {
      "job_id": "sh-001",
      "job_type": "store_sync",
      "status": "processing",
      "entity_type": "store",
      "total_items": 50,
      "completed_items": 30,
      "message": "Fetching brand for cool-store.myshopify.com",
      "started_at": "2026-04-21T10:00:36Z",
      "completed_at": null,
      "duration_seconds": null
    },
    {
      "job_id": "rv-001",
      "job_type": "review_sync",
      "status": "completed",
      "entity_type": "review",
      "total_items": 18,
      "completed_items": 18,
      "message": null,
      "started_at": "2026-04-21T10:00:02Z",
      "completed_at": "2026-04-21T10:00:08Z",
      "duration_seconds": 6
    },
    {
      "job_id": "sn-001",
      "job_type": "snapshot_sync",
      "status": "processing",
      "entity_type": "snapshot",
      "total_items": 12,
      "completed_items": 7,
      "message": "Backfilling snapshot month 7/12",
      "started_at": "2026-04-21T10:00:36Z",
      "completed_at": null,
      "duration_seconds": null
    }
  ],
  "summary": {
    "total_children": 6,
    "completed": 3,
    "processing": 3,
    "failed": 0
  }
}
```

**Response for single job (e.g., `transaction_sync`):**

```json
{
  "job_id": "tx-001",
  "job_type": "transaction_sync",
  "status": "processing",
  "entity_type": "transaction",
  "total_items": 1200,
  "completed_items": 780,
  "message": "Fetching transactions page 8/12",
  "started_at": "2026-04-21T10:00:02Z",
  "completed_at": null,
  "duration_seconds": null
}
```

**Handler logic:**
1. Read `sync_jobs` row from DB (+ children if `full_sync`)
2. For any job with `status='processing'`: read `lg:sync:progress:{jobID}` from Redis
3. Overlay Redis `completed`/`total`/`message` onto DB values (Redis is more current)
4. If Redis key missing (Redis flushed), fall back to DB `completed_items` (stale by up to 30s)
5. Compute `duration_seconds` for completed jobs

### 5.4 List Jobs / Sync History

```
GET /api/v1/sync/jobs?app_id={appID}&status=completed&limit=20&offset=0
Auth: Firebase JWT (tenant-scoped)

Response 200:
{
  "jobs": [
    {
      "job_id": "abc",
      "job_type": "full_sync",
      "status": "completed",
      "total_items": 0,
      "completed_items": 0,
      "created_at": "2026-04-21T10:00:00Z",
      "started_at": "2026-04-21T10:00:01Z",
      "completed_at": "2026-04-21T10:01:15Z",
      "duration_seconds": 74
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0
}

Filters:
  app_id   = required (UUID)
  status   = optional (pending, processing, completed, failed, partial_failure)
  job_type = optional (full_sync, transaction_sync, etc.)
  limit    = optional (default 20, max 100)
  offset   = optional (default 0)
```

### 5.5 Cancel Job

```
POST /api/v1/sync/jobs/{jobID}/cancel
Auth: Firebase JWT (tenant-scoped)

Response 200:
{"message": "Job cancellation requested", "job_id": "abc"}

Response 409 (already completed/cancelled):
{"error": "Job is already completed"}
```

**Logic:**
1. SET `lg:sync:cancel:{jobID}` in Redis (TTL 1 hour)
2. UPDATE `sync_jobs SET status='cancelled'` (for pending jobs — immediate)
3. For processing jobs: worker checks cancel flag between steps → sets status='cancelled' when detected

---

## 6. Progress Reporting — Dual Write Strategy

### 6.1 Write Path (Worker → Redis + DB)

```
Worker processing a job:
  │
  ├─ Every ~2 seconds (or after meaningful batch):
  │   → Redis HSET lg:sync:progress:{jobID} total=1000 completed=450 message="..."
  │   (cheap, fire-and-forget, ephemeral)
  │
  ├─ Every ~30 seconds (or at milestones: 25%, 50%, 75%, 100%):
  │   → UPDATE sync_jobs SET completed_items=450, updated_at=NOW() WHERE id=$1
  │   (durable checkpoint — survives Redis loss)
  │
  └─ On completion:
      → UPDATE sync_jobs SET status='completed', completed_items=total_items, completed_at=NOW()
      → DEL lg:sync:progress:{jobID}  (cleanup ephemeral data)
```

### 6.2 Read Path (Frontend → API → DB + Redis)

```
GET /sync/jobs/{jobID}/progress
  │
  ├─ 1. SELECT * FROM sync_jobs WHERE id=$1  (always)
  ├─ 2. If status='processing':
  │      HGETALL lg:sync:progress:{jobID}    (overlay live counters)
  │      Use Redis values if available, else DB values
  └─ 3. Return merged response
```

### 6.3 Throttling

```go
type ProgressTracker struct {
    redisClient  *RedisClient
    syncJobRepo  repository.SyncJobRepository
    minInterval  time.Duration  // 2 seconds for Redis
    dbInterval   time.Duration  // 30 seconds for DB
    lastRedis    map[string]time.Time
    lastDB       map[string]time.Time
}

func (pt *ProgressTracker) Update(ctx context.Context, jobID string, p Progress) {
    now := time.Now()

    // Always update Redis (throttled to 2s)
    if now.Sub(pt.lastRedis[jobID]) >= pt.minInterval {
        pt.redisClient.HSet(ctx, "lg:sync:progress:"+jobID, map[string]interface{}{
            "total":     p.Total,
            "completed": p.Completed,
            "message":   p.Message,
        })
        pt.redisClient.Expire(ctx, "lg:sync:progress:"+jobID, time.Hour)
        pt.lastRedis[jobID] = now
    }

    // Checkpoint to DB (throttled to 30s or at milestones)
    if now.Sub(pt.lastDB[jobID]) >= pt.dbInterval || pt.isMilestone(p) {
        pt.syncJobRepo.UpdateProgress(ctx, jobID, p.Total, p.Completed)
        pt.lastDB[jobID] = now
    }
}

func (pt *ProgressTracker) isMilestone(p Progress) bool {
    if p.Total == 0 { return false }
    pct := float64(p.Completed) / float64(p.Total) * 100
    return pct == 25 || pct == 50 || pct == 75 || pct == 100
}
```

### 6.4 Frontend Polling Strategy

```
Flutter SyncProvider:
  1. POST /sync/enqueue/{appID}?type=full → get jobID
  2. Poll GET /sync/jobs/{jobID}/progress every 3 seconds
  3. Render per-entity progress bars from children array
  4. Stop polling when parent status == "completed" or "failed" or "partial_failure"
  5. On "partial_failure": show which children failed with error messages
```

---

## 7. Event Sync — New Job Type

### What It Does

Fetches **all app lifecycle events** from Shopify Partner API and stores them as raw records in the `app_events` table. This creates a permanent historical audit trail separate from the derived subscription status.

### Current State

Currently `enrichSubscriptionStatus()` in `sync_service.go` calls `FetchAppEvents()` per-subscription, processes events in memory, and discards them after deriving status. No raw events are stored.

### What Changes

| Before | After |
|--------|-------|
| Events fetched per-subscription, used once, discarded | Events fetched once per app, stored permanently |
| No install/uninstall history | Full install/uninstall timeline in DB |
| `status_sync` fetches events + derives status | `event_sync` fetches + stores; `status_sync` reads from `app_events` or re-fetches |

### Why Separate from status_sync?

| | event_sync | status_sync |
|---|-----------|-------------|
| **Purpose** | Capture + store raw events | Derive subscription status from events |
| **Output** | `app_events` rows | `subscriptions.status` + `subscriptions.risk_state` |
| **Dependencies** | None — fetches directly from Partner API | Needs subscriptions (from `transaction_sync`) |
| **Frequency** | Can run independently | Must run after `transaction_sync` |
| **Value** | Historical audit trail, install count, churn analysis | Current subscription state for risk engine |

### Future Uses of Stored Events

- Install/uninstall timeline per shop (CRM feature)
- Churn rate calculation from actual uninstall events
- Webhook reconciliation (compare webhook events vs Partner API events)
- Event-driven notifications (alert on uninstall)

---

## 8. Distributed Locking Strategy

### Lock Granularity

```
Lock key:      lg:sync:lock:{appID}:{syncType}
Heartbeat key: lg:sync:heartbeat:{appID}:{syncType}
```

| Sync Type | Lock Prevents | Lock Duration |
|-----------|--------------|---------------|
| `transaction` | Concurrent transaction fetches for same app | ~30-60s typical |
| `status` | Concurrent status enrichment for same app | ~10-30s typical |
| `store` | Concurrent brand fetches for same app | ~5-15s typical |
| `review` | Concurrent review scrapes for same app | ~5-10s typical |
| `event` | Concurrent event fetches for same app | ~5-15s typical |
| `snapshot` | Concurrent snapshot backfills for same app | ~10-30s typical |
| `full` | Concurrent full syncs for same app | ~60-120s typical |

**Same app, different sync types CAN run concurrently.** For example, `store_sync` + `review_sync` for the same app can run in parallel (different lock keys).

### Lock Lifecycle

```
Worker dequeues job from Redis
  │
  ├─ UPDATE sync_jobs SET status='processing', worker_id=X, started_at=NOW()
  │
  ├─ AcquireLock(appID, syncType)  — Redis SETNX, TTL=2hr
  │   │
  │   ├─ ACQUIRED:
  │   │   ├─ Start heartbeat goroutine (SET every 10min, TTL=20min)
  │   │   ├─ Start lock extension goroutine (EXPIRE every 1hr, TTL=2hr)
  │   │   ├─ Process job...
  │   │   ├─ ReleaseLock() — DEL lock key
  │   │   └─ DeleteHeartbeat() — DEL heartbeat key
  │   │
  │   └─ NOT ACQUIRED:
  │       ├─ Check IsWorkerAlive() — GET heartbeat key
  │       │   ├─ Alive → re-enqueue job, skip
  │       │   └─ Dead (no heartbeat or expired) → StealLock() + process
  │       │
  │       └─ UPDATE sync_jobs SET status='pending' (if re-enqueued)
  │
  └─ On completion:
      UPDATE sync_jobs SET status='completed', completed_items=X, completed_at=NOW()
```

### Lock Constants

```go
const (
    lockTTL              = 2 * time.Hour
    lockExtensionEvery   = 1 * time.Hour
    heartbeatRefreshEvery = 10 * time.Minute
    heartbeatTTL         = 20 * time.Minute
)
```

---

## 9. Recovery System

### 9.1 On Startup

When backend starts (or a new instance joins):

```go
func (rs *RecoveryService) RecoverOnStartup(ctx context.Context) error {
    // Step 1: Re-enqueue interrupted jobs
    // Find sync_jobs with status='processing' that have no live heartbeat
    jobs := rs.syncJobRepo.FindByStatus(ctx, "processing")
    for _, job := range jobs {
        alive := rs.redisClient.IsWorkerAlive(ctx, job.AppID, getSyncType(job.JobType))
        if !alive {
            rs.syncJobRepo.UpdateStatus(ctx, job.ID, "pending")
            rs.queue.Enqueue(ctx, jobToSyncJob(job))
            log.Info("Re-enqueued interrupted job", "job_id", job.ID)
        }
    }

    // Step 2: Re-enqueue pending jobs not in Redis queue
    // (handles Redis flush scenario)
    pendingJobs := rs.syncJobRepo.FindByStatus(ctx, "pending")
    for _, job := range pendingJobs {
        rs.queue.Enqueue(ctx, jobToSyncJob(job))
    }

    // Step 3: Clean stale locks (locks without heartbeats)
    rs.cleanupStaleLocks(ctx)

    return nil
}
```

### 9.2 Periodic Recovery (Every 10 Minutes)

```go
func (rs *RecoveryService) StartPeriodicRecovery(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Minute)
    for {
        select {
        case <-ticker.C:
            rs.cleanupStaleLocks(ctx)
            rs.reEnqueueStuckJobs(ctx) // Jobs in 'processing' for > 2 hours
        case <-ctx.Done():
            return
        }
    }
}
```

### 9.3 Redis Flush Recovery

If Redis is completely flushed while backend is running:
1. **Queued jobs lost** → Periodic recovery detects `pending` jobs in DB with no Redis queue entry → re-enqueues
2. **Locks lost** → Workers re-acquire on next job (may cause brief duplicate, but idempotent processing handles it)
3. **Progress lost** → Frontend falls back to DB `completed_items` (stale by up to 30s)
4. **Heartbeats lost** → Recovery treats all processing jobs as potentially dead → re-enqueues after verification

---

## 10. Scaling Analysis

### 10.1 Single Instance (MVP Target)

```
┌──────────────────────────────────┐
│       Go Backend Server           │
│                                   │
│  HTTP handlers                    │
│  + 3 regular queue workers        │
│  + 1 full sync queue worker       │
│  + 1 recovery goroutine           │
│  + existing sync scheduler (12hr) │
│                                   │
│  All in-process, no sidecar       │
└──────────┬───────────────────────┘
           │               │
           ▼               ▼
┌─────────────────┐  ┌────────────┐
│  Redis (single)  │  │ PostgreSQL │
│  Queues + locks  │  │ sync_jobs  │
│  + progress      │  │ app_events │
└─────────────────┘  └────────────┘
```

**Good for:** Up to ~50 apps, ~5 concurrent syncs.

### 10.2 Horizontal Scaling (Multiple Instances)

```
┌────────────┐ ┌────────────┐ ┌────────────┐
│ Instance 1  │ │ Instance 2  │ │ Instance 3  │
│ 3+1 workers │ │ 3+1 workers │ │ 3+1 workers │
│ HTTP + queue│ │ HTTP + queue│ │ HTTP + queue│
└──────┬──────┘ └──────┬──────┘ └──────┬──────┘
       │               │               │
       └───────────────┼───────────────┘
                       │
          ┌────────────┼────────────┐
          ▼                         ▼
┌──────────────────┐     ┌─────────────────┐
│ Redis Sentinel    │     │   PostgreSQL     │
│ or Redis Cluster  │     │  (Cloud SQL)     │
│ (HA)              │     │                  │
└──────────────────┘     └─────────────────┘
```

**What works automatically across instances:**

| Feature | Why It Works |
|---------|-------------|
| Job dequeue | Redis BRPOP is atomic — each job consumed by exactly one worker |
| Distributed locks | Redis SETNX — only one worker acquires per app+type |
| Heartbeat detection | Any instance can detect dead workers from other instances |
| Recovery | Any instance's recovery goroutine re-enqueues globally |
| Job state | PostgreSQL — shared across all instances |
| Progress | Redis + DB — both shared |
| Tenant isolation | SQL WHERE clause — always correct |

**What needs explicit handling:**

| Concern | Solution | Priority |
|---------|----------|----------|
| **Shopify rate limiting** | Replace in-memory token bucket with Redis sliding window (see 10.3) | **Must have** — breaks without it |
| **Sync scheduler** | Only one instance should schedule. Use `SETNX` leader election: `lg:scheduler:leader` with 15min TTL, refresh every 5min | Medium — current scheduler is in-memory ticker |
| **In-memory cache** | Replace with Redis cache or accept cache misses (cache is for performance, not correctness) | Low — stale cache = slightly slow reads |
| **Redis SPOF** | Redis Sentinel (min 3 nodes) or Redis Cluster | Medium — needed for production |

### 10.3 Rate Limiting Across Workers (Critical for Scaling)

**Problem:** Current `ShopifyPartnerClient` uses in-memory per-partner token bucket (4 RPS). With multiple workers hitting the Partner API for different apps under the **same partner account**, they will collectively exceed 4 RPS.

**Solution: Redis-based sliding window rate limiter**

```go
// backend/internal/infrastructure/queue/ratelimit.go

type RedisRateLimiter struct {
    client     redis.Cmdable
    rps        int           // requests per second limit
    windowSize time.Duration // 1 second
}

func (r *RedisRateLimiter) Wait(ctx context.Context, partnerID string) error {
    key := fmt.Sprintf("lg:ratelimit:%s", partnerID)
    for {
        allowed, err := r.allow(ctx, key)
        if err != nil {
            return err // Redis error — fail open or fail closed (configurable)
        }
        if allowed {
            return nil
        }
        // Wait and retry
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(250 * time.Millisecond): // Retry 4x per second
            continue
        }
    }
}

func (r *RedisRateLimiter) allow(ctx context.Context, key string) (bool, error) {
    now := time.Now().UnixMicro()
    windowStart := now - r.windowSize.Microseconds()

    // Atomic pipeline: clean old entries, count current, add new
    pipe := r.client.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
    countCmd := pipe.ZCard(ctx, key)
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
    pipe.Expire(ctx, key, 2*time.Second)
    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    count := countCmd.Val()
    if count >= int64(r.rps) {
        // Over limit — remove the optimistic add
        r.client.ZRem(ctx, key, fmt.Sprintf("%d", now))
        return false, nil
    }
    return true, nil
}
```

**Integration:** Workers call `rateLimiter.Wait(ctx, partnerID)` before every Shopify Partner API call. This replaces the in-memory token bucket when queue is enabled.

### 10.4 Capacity Planning

| Scenario | Apps | Workers | Rate Limit Bottleneck | Queue Drain |
|----------|------|---------|----------------------|-------------|
| Solo developer | 1-5 | 3 (1 instance) | 1 partner → 4 RPS max | < 1 min |
| Small agency | 10-30 | 3 (1 instance) | 1-3 partners → 4-12 RPS | ~5 min |
| Growing platform | 50-200 | 9 (3 instances) | 10+ partners → 40+ RPS | ~10 min |
| Enterprise | 500-1000 | 30 (10 instances) | 50+ partners → 200+ RPS | ~17 min |

**Key insight:** The bottleneck is **Shopify's 4 RPS per partner account**, not worker count. Adding workers only helps when you have many partner accounts (each with its own rate limit pool). For a single partner account, even 100 workers would be throttled to 4 RPS.

### 10.5 DB Write Pressure Analysis

| Operation | Frequency | Impact |
|-----------|-----------|--------|
| INSERT sync_job (enqueue) | Once per job | Negligible |
| UPDATE status='processing' | Once per job | Negligible |
| UPDATE completed_items (progress checkpoint) | Every 30s per active job | ~10 writes/min at 20 concurrent syncs |
| UPDATE status='completed' | Once per job | Negligible |
| Redis progress HSET | Every 2s per active job | ~600 writes/min at 20 concurrent syncs (Redis handles this easily) |

**Conclusion:** DB write pressure from progress checkpoints is minimal. Even at 100 concurrent syncs, it's ~33 writes/min — negligible for PostgreSQL.

---

## 11. Configuration

### 11.1 Config File Additions

```yaml
# config.local.yaml
redis:
  addr: "localhost:6379"       # Comma-separated for cluster: "r1:6379,r2:6379,r3:6379"
  password: ""
  db: 0                        # Ignored in cluster mode

queue:
  enabled: true                # Feature flag — false = queue system disabled, existing sync only
  num_workers: 3               # Regular queue worker goroutines
  full_sync_workers: 1         # Full sync queue worker goroutines
  recovery_interval: "10m"     # Periodic stale lock cleanup
  progress_redis_interval: "2s"   # Live progress write throttle
  progress_db_interval: "30s"     # DB checkpoint throttle
```

### 11.2 Environment Variable Overrides

```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=secret
QUEUE_ENABLED=true
QUEUE_NUM_WORKERS=3
QUEUE_FULL_SYNC_WORKERS=1
```

### 11.3 Config Struct

```go
type RedisConfig struct {
    Addr     string `yaml:"addr"`     // Comma-separated for cluster
    Password string `yaml:"password"`
    DB       int    `yaml:"db"`
}

type QueueConfig struct {
    Enabled              bool          `yaml:"enabled"`
    NumWorkers           int           `yaml:"num_workers"`
    FullSyncWorkers      int           `yaml:"full_sync_workers"`
    RecoveryInterval     time.Duration `yaml:"recovery_interval"`
    ProgressRedisInterval time.Duration `yaml:"progress_redis_interval"`
    ProgressDBInterval   time.Duration `yaml:"progress_db_interval"`
}
```

---

## 12. File Structure

```
backend/
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   └── app_event.go                    — AppEvent entity (NEW)
│   │   └── repository/
│   │       ├── app_event_repository.go          — AppEventRepository interface (NEW)
│   │       └── sync_job_repository.go           — SyncJobRepository interface (NEW)
│   │
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── app_event_repository.go          — PostgreSQL impl (NEW)
│   │   │   └── sync_job_repository.go           — PostgreSQL impl (NEW)
│   │   │
│   │   └── queue/                               — NEW PACKAGE
│   │       ├── config.go                        — RedisConfig, QueueConfig
│   │       ├── client.go                        — NewRedisClient (single + cluster)
│   │       ├── queue.go                         — SyncJob, Enqueue/Dequeue, queue names
│   │       ├── lock.go                          — AcquireLock, ReleaseLock, Heartbeat
│   │       ├── progress.go                      — ProgressTracker (dual write)
│   │       ├── recovery.go                      — RecoverOnStartup, periodic cleanup
│   │       ├── worker.go                        — WorkerPool, worker loop, graceful shutdown
│   │       ├── processor.go                     — SyncProcessor interface + registry
│   │       ├── ratelimit.go                     — Redis sliding window rate limiter
│   │       └── processors/
│   │           ├── transaction_processor.go     — Reuses fetcher + txRepo + ledger
│   │           ├── snapshot_processor.go        — Reuses LedgerService.BackfillHistoricalSnapshots
│   │           ├── status_processor.go          — Reuses event enrichment logic
│   │           ├── store_processor.go           — Reuses brand fetch logic
│   │           ├── review_processor.go          — Reuses review scraper logic
│   │           ├── event_processor.go           — NEW: fetch + store raw events
│   │           └── full_sync_processor.go       — Orchestrator: dispatch + poll
│   │
│   └── interfaces/http/handler/
│       └── queue_sync.go                        — HTTP handlers (NEW)
│
├── migrations/
│   ├── 000033_create_sync_jobs_table.up.sql     — NEW
│   ├── 000033_create_sync_jobs_table.down.sql   — NEW
│   ├── 000034_create_app_events_table.up.sql    — NEW
│   └── 000034_create_app_events_table.down.sql  — NEW
│
docs/developer/
├── 31-queue-sync-system.md                      — Feature doc (8-section format)
└── GUIDE-adding-new-sync-type.md                — Extensibility tutorial

docs/diagrams/puml/
└── 31-queue-sync-sequence.puml                  — PlantUML diagram
```

**Total new files: ~21** (10 queue package + 7 processors + 2 repos + 2 migrations + handler + docs)

---

## 13. Implementation Phases

### Phase A: Foundation — Redis + DB Schema (Est: 1 session)

| # | Task | Files |
|---|------|-------|
| 1 | Add `github.com/redis/go-redis/v9` dependency | `go.mod`, `go.sum` |
| 2 | RedisConfig + QueueConfig structs | `queue/config.go`, update `config/config.go` |
| 3 | Redis client init (single + cluster auto-detect) | `queue/client.go` |
| 4 | `sync_jobs` migration | `migrations/000033_*` |
| 5 | `app_events` migration | `migrations/000034_*` |
| 6 | SyncJobRepository interface + Postgres impl | `repository/sync_job_repository.go`, `persistence/sync_job_repository.go` |
| 7 | AppEventRepository interface + Postgres impl | `repository/app_event_repository.go`, `persistence/app_event_repository.go` |
| 8 | AppEvent entity | `entity/app_event.go` |
| 9 | Wire Redis client in `main.go` (optional, graceful) | `cmd/server/main.go` |
| 10 | Tests: repo CRUD, Redis connection | unit + integration |

### Phase B: Queue Core — Enqueue/Dequeue + Locks (Est: 1 session)

| # | Task | Files |
|---|------|-------|
| 1 | SyncJob struct, Enqueue/Dequeue, queue names | `queue/queue.go` |
| 2 | AcquireLock, ReleaseLock, ExtendLock, Heartbeat | `queue/lock.go` |
| 3 | ProgressTracker (dual write: Redis + DB) | `queue/progress.go` |
| 4 | RecoverOnStartup, periodic cleanup | `queue/recovery.go` |
| 5 | Tests: enqueue/dequeue, lock contention, progress, recovery | unit (miniredis) |

### Phase C: Worker + Processor Framework (Est: 1 session)

| # | Task | Files |
|---|------|-------|
| 1 | SyncProcessor interface, ProcessorRegistry | `queue/processor.go` |
| 2 | WorkerPool, worker loop, graceful shutdown | `queue/worker.go` |
| 3 | Wire workers in `main.go`, start on boot | `cmd/server/main.go` |
| 4 | Tests: worker lifecycle, shutdown, processor dispatch | unit |

### Phase D: Processors (Est: 2 sessions)

| # | Task | Files |
|---|------|-------|
| 1 | TransactionProcessor — calls fetcher + txRepo + ledger | `processors/transaction_processor.go` |
| 2 | SnapshotProcessor — calls LedgerService.BackfillHistoricalSnapshots | `processors/snapshot_processor.go` |
| 3 | EventProcessor — fetches + stores raw events | `processors/event_processor.go` |
| 4 | StatusProcessor — enriches subscription status | `processors/status_processor.go` |
| 5 | StoreProcessor — fetches brand data | `processors/store_processor.go` |
| 6 | ReviewProcessor — scrapes reviews | `processors/review_processor.go` |
| 7 | FullSyncProcessor — orchestrator (6 children) | `processors/full_sync_processor.go` |
| 8 | Register all processors in `main.go` | `cmd/server/main.go` |
| 9 | Tests: each processor with mocked dependencies | unit |

### Phase E: HTTP Layer (Est: 1 session)

| # | Task | Files |
|---|------|-------|
| 1 | Enqueue, JobStatus, JobProgress, ListJobs, Cancel handlers | `handler/queue_sync.go` |
| 2 | Wire routes in router | `router/router.go` |
| 3 | Wire handler in main.go | `cmd/server/main.go` |
| 4 | Tests: HTTP handler tests | unit |
| 5 | E2E test: enqueue → poll progress → verify completion | integration |

### Phase F: Documentation (Est: 1 session)

| # | Task | Files |
|---|------|-------|
| 1 | Feature doc (8-section format) | `docs/developer/31-queue-sync-system.md` |
| 2 | Extensibility tutorial | `docs/developer/GUIDE-adding-new-sync-type.md` |
| 3 | PlantUML diagram | `docs/diagrams/puml/31-queue-sync-sequence.puml` |
| 4 | Update `00-index.md` with new entries | `docs/developer/00-index.md` |
| 5 | Update `DATABASE_SCHEMA.md` | `DATABASE_SCHEMA.md` |
| 6 | Update `IMPLEMENTATION_LOG.md` | `IMPLEMENTATION_LOG.md` |

### Phase G: Scaling Prep (Future — when multi-instance needed)

| # | Task | Files |
|---|------|-------|
| 1 | Redis sliding window rate limiter | `queue/ratelimit.go` |
| 2 | Scheduler leader election via Redis | `scheduler/leader.go` |
| 3 | Redis Sentinel/Cluster config | `queue/config.go` |
| 4 | Migrate in-memory cache to Redis | `cache/redis_cache.go` |

---

## 14. Extensibility Guide (Summary)

> Full tutorial: [GUIDE-adding-new-sync-type.md](GUIDE-adding-new-sync-type.md)

### Adding a New Sync Type: `payout_sync` Example

**Step 1: Create processor** (1 new file)

```go
// backend/internal/infrastructure/queue/processors/payout_processor.go

package processors

type PayoutProcessor struct {
    partnerClient *external.ShopifyPartnerClient
    payoutRepo    repository.PayoutRepository
    progress      *queue.ProgressTracker
    syncJobRepo   repository.SyncJobRepository
}

func (p *PayoutProcessor) Type() string { return "payout_sync" }

func (p *PayoutProcessor) Process(ctx context.Context, job *queue.SyncJob) error {
    // 1. Load app + partner account (same pattern as other processors)
    // 2. Fetch payouts from Shopify Partner API
    // 3. Update progress: p.progress.Update(ctx, job.JobID, ...)
    // 4. Store payouts via payoutRepo
    // 5. Return nil on success
    return nil
}
```

**Step 2: Register in main.go** (1 line)

```go
registry.Register(&processors.PayoutProcessor{...})
```

**Step 3: Add to full_sync children** (optional, 1 line)

```go
// In full_sync_processor.go
independentJobs := []string{"transaction_sync", "event_sync", "review_sync", "payout_sync"}
// or dependentJobs for jobs that need transaction_sync first:
dependentJobs := []string{"snapshot_sync", "status_sync", "store_sync", "payout_sync"}
```

**Step 4: Add lock type** (1 line in lock.go)

```go
case "payout_sync": return "payout"
```

**That's it.** 1 file + 3 one-line edits. The queue, workers, progress tracking, locking, recovery, HTTP endpoints, and frontend progress display all work automatically for the new sync type.

---

## 15. Migration from Existing Sync

### Coexistence Strategy

```
                    ┌──────────────────────────┐
                    │   Existing Sync (kept)    │
                    │  POST /api/v1/sync/{appID}│
                    │  Synchronous, blocking    │
                    │  For: quick test, backward│
                    │        compat, scheduler  │
                    └──────────────────────────┘

                    ┌──────────────────────────┐
                    │   Queue Sync (new)        │
                    │  POST /api/v1/sync/       │
                    │       enqueue/{appID}     │
                    │  Async, non-blocking      │
                    │  For: Flutter app, API    │
                    └──────────────────────────┘

Both call the same domain services.
Zero code duplication.
Feature-flagged: QUEUE_ENABLED=false → queue system fully disabled.
```

### Migration Path

| Phase | What Changes | Risk |
|-------|-------------|------|
| **Phase 1 (Now)** | Ship queue sync alongside existing sync. Flutter uses new endpoints. Old endpoints untouched. | Zero — additive only |
| **Phase 2 (Later)** | Migrate scheduler from `SyncService.SyncAllApps()` to `QueueService.EnqueueSync(type=full)` | Low — scheduler uses new path |
| **Phase 3 (Later)** | Deprecate `POST /api/v1/sync/{appID}` or keep for health checks / quick manual sync | Low — optional |

---

## 16. Gotchas & Edge Cases

| # | Gotcha | Mitigation |
|---|--------|-----------|
| 1 | **Redis unavailable at startup** | Queue system disabled gracefully via feature flag. Existing sync unaffected. Log warning. |
| 2 | **Redis flushed while running** | Recovery goroutine re-enqueues `pending` + `processing` jobs from DB. Progress falls back to DB checkpoints. |
| 3 | **Full sync child fails** | Parent marks child as failed, other children continue. Parent final status = `partial_failure`. Error message preserved in child's `sync_jobs` row. |
| 4 | **App deleted during sync** | Processor checks app exists before processing. If deleted → mark job `cancelled`. |
| 5 | **Duplicate full_sync** | Enqueue handler checks `sync_jobs` for existing active job with same `app_id + job_type` → returns 409. Distributed lock is second safety net. |
| 6 | **Worker dies mid-sync** | Heartbeat expires (20min) → lock stealable → recovery re-enqueues from DB → another worker picks up. |
| 7 | **Shopify rate limit exceeded** | Redis rate limiter blocks worker. Other workers for different partners continue. 250ms retry loop with context cancellation. |
| 8 | **Job stuck > 2 hours** | Periodic recovery (every 10min) detects jobs in `processing` longer than lock TTL → re-enqueues. |
| 9 | **Progress write pressure** | Redis: throttled to every 2s. DB: throttled to every 30s or at milestones (25/50/75/100%). |
| 10 | **Cancellation lag** | Cooperative: worker checks `lg:sync:cancel:{jobID}` between major steps. Up to ~30s lag (one API page). |
| 11 | **transaction_sync fails in full_sync** | snapshot_sync, status_sync, and store_sync never enqueued (they depend on it). Parent retains event_sync + review_sync results. Status = `partial_failure`. |
| 12 | **Same partner, many apps** | All apps share 4 RPS rate limit. Workers block on rate limiter fairly. Queue drains slower but correctly. |
| 13 | **DB connection exhaustion** | Workers share the same pgx pool as HTTP handlers. Pool size must account for worker count. Default pgx pool = 4; increase to `max_workers + http_concurrency`. |

---

## 17. Testing Strategy

| Level | What | Tool | Coverage |
|-------|------|------|----------|
| **Unit** | Processor logic (each of 7) | `go test`, mock repos | Each processor in isolation |
| **Unit** | Queue enqueue/dequeue | `miniredis` | FIFO ordering, full sync routing |
| **Unit** | Lock acquire/release/steal | `miniredis` | Contention, heartbeat expiry |
| **Unit** | ProgressTracker throttling | `miniredis` + mock repo | Redis vs DB write frequency |
| **Unit** | Recovery logic | `miniredis` + mock repo | Re-enqueue interrupted, clean stale |
| **Unit** | Rate limiter | `miniredis` | Sliding window correctness |
| **Integration** | Full sync orchestration | Real Redis + test DB | Enqueue → workers → children → completion |
| **Integration** | Recovery after crash | Kill worker mid-job | Verify re-enqueue + completion |
| **Integration** | HTTP endpoints | httptest + real Redis | Enqueue → poll progress → verify |
| **E2E** | Frontend flow | Manual | Trigger from Flutter → watch progress bars → verify |

**Test dependency:** `github.com/alicebob/miniredis/v2` for fast, in-memory Redis in unit tests.

---

## 18. Verification Checklist

### Functional

- [ ] `POST /api/v1/sync/{appID}` still works (existing sync completely unchanged)
- [ ] `QUEUE_ENABLED=false` → queue system fully disabled, no Redis needed
- [ ] `POST /api/v1/sync/enqueue/{appID}?type=full` returns 202 with `job_id`
- [ ] `POST /api/v1/sync/enqueue/{appID}?type=transaction` enqueues single job
- [ ] `GET /api/v1/sync/jobs/{jobID}/progress` returns DB + Redis merged progress
- [ ] `GET /api/v1/sync/jobs?app_id=X` returns sync history (paginated)
- [ ] Full sync dispatches 6 child jobs in correct dependency order
- [ ] `snapshot_sync`, `status_sync`, and `store_sync` wait for `transaction_sync` completion
- [ ] Each child job reports progress with correct `entity_type` and `total/completed`
- [ ] Duplicate enqueue for same app+type returns 409 with existing job_id
- [ ] Cancellation sets status to `cancelled` and stops worker
- [ ] Tenant isolation: user cannot see/cancel jobs for apps they don't own

### Reliability

- [ ] Distributed lock prevents duplicate processing across workers
- [ ] Worker heartbeat detects dead workers (heartbeat TTL=20min)
- [ ] Recovery re-enqueues interrupted jobs on startup
- [ ] Recovery cleans stale locks every 10 minutes
- [ ] Redis unavailable → graceful degradation, existing sync works
- [ ] Redis flush → recovery re-enqueues from DB, progress falls back to DB
- [ ] `partial_failure` status when some children fail but others succeed

### Performance

- [ ] Progress Redis writes throttled to every 2s (not per-record)
- [ ] Progress DB writes throttled to every 30s (or milestones)
- [ ] Rate limiter respects 4 RPS per partner across all workers
- [ ] pgx pool size increased to accommodate workers
- [ ] `go test ./internal/infrastructure/queue/... -v` passes
- [ ] `go test ./internal/infrastructure/persistence/... -v` passes (new repos)

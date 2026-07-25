# Sync Pipeline Blueprint

> A reusable, language-agnostic blueprint for building production-grade async job processing pipelines.
> Extracted from a system with 14+ production bug fixes covering distributed locking, multi-node
> coordination, crash recovery, graceful shutdown, and wave-based orchestration.

---

## How to Use This Blueprint

**Three usage modes:**

1. **Include in CLAUDE.md** (or equivalent AI agent config):
   ```
   ## Sync Pipeline
   Follow the blueprint at `docs/SYNC_PIPELINE_BLUEPRINT.md` for all sync/job-processing implementation.
   ```

2. **Give to an AI agent as a prompt** — paste the relevant sections as context when implementing sync features.

3. **Follow as an implementation checklist** — use Section 15 as a step-by-step guide.

**Prerequisites:**
- A KV store with atomic operations (Redis recommended, or DynamoDB/etcd/Valkey)
- A relational database (Postgres, MySQL, SQLite)
- Background worker support (goroutines, threads, async tasks, Celery, Sidekiq, etc.)

**Extension points:** Search for `[YOUR PROJECT]` to find the ~11 places you must customize for your domain.

---

## Table of Contents

0. [Header & Usage Instructions](#how-to-use-this-blueprint)
1. [Architecture Overview](#1-architecture-overview)
2. [Job Entity & State Machine](#2-job-entity--state-machine)
3. [Queue](#3-queue)
4. [Lock Manager (Lua Scripts)](#4-lock-manager)
5. [Worker Pool](#5-worker-pool)
6. [Processor Registry & Interface](#6-processor-registry--interface)
7. [Processor Context (Common Preamble)](#7-processor-context)
8. [Orchestrator Pattern (Parent/Child with Waves)](#8-orchestrator-pattern)
9. [Progress Tracking (Dual-Write)](#9-progress-tracking)
10. [Cooperative Cancellation](#10-cooperative-cancellation)
11. [Recovery Service](#11-recovery-service)
12. [Enqueue Service (API Layer)](#12-enqueue-service)
13. [Schedulers](#13-schedulers)
14. [Pitfall Catalog](#14-pitfall-catalog)
15. [Implementation Checklist](#15-implementation-checklist)
16. [Extension Point Registry](#16-extension-point-registry)
17. [Database Schema (Reference)](#17-database-schema)
18. [Redis Key Reference](#18-redis-key-reference)

---

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        API / Scheduler                          │
│  ┌──────────────┐  ┌──────────────────┐  ┌─────────────────┐   │
│  │ EnqueueService│  │ PeriodicScheduler│  │ CatchupScheduler│   │
│  └──────┬───────┘  └────────┬─────────┘  └────────┬────────┘   │
│         │                   │                     │             │
│         ▼                   ▼                     ▼             │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Redis Queues                          │    │
│  │  ┌─────────────────┐    ┌──────────────────────────┐    │    │
│  │  │  regular queue   │    │  orchestrator queue       │    │    │
│  │  └─────────────────┘    └──────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────┘    │
│         │                                     │                 │
│         ▼                                     ▼                 │
│  ┌──────────────┐                     ┌──────────────┐          │
│  │ WorkerPool   │                     │ WorkerPool   │          │
│  │ (N workers)  │                     │ (1 worker)   │          │
│  └──────┬───────┘                     └──────┬───────┘          │
│         │                                     │                 │
│         ▼                                     ▼                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                 ProcessorRegistry                        │    │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────┐ │    │
│  │  │Processor A│ │Processor B│ │Processor C│ │Orchestr.│ │    │
│  │  └───────────┘ └───────────┘ └───────────┘ └─────────┘ │    │
│  └─────────────────────────────────────────────────────────┘    │
│         │              │              │              │           │
│         ▼              ▼              ▼              ▼           │
│  ┌─────────────┐ ┌──────────┐ ┌──────────────┐ ┌───────────┐   │
│  │ LockManager │ │ Progress │ │RecoveryService│ │ Database  │   │
│  │ (Redis Lua) │ │ Tracker  │ │(startup+loop)│ │(sync_jobs)│   │
│  └─────────────┘ └──────────┘ └──────────────┘ └───────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

**Component Responsibilities:**

| Component | Responsibility |
|-----------|---------------|
| **Job Entity** | Domain object with status state machine and terminal-state helper |
| **Queue** | Redis FIFO (LPUSH/BRPOP) with separate keys for priority separation |
| **WorkerPool** | Dequeue loop → lock → process → state transition → cleanup |
| **LockManager** | Distributed locks + heartbeat + cancellation via atomic Lua scripts |
| **ProcessorRegistry** | Maps job types to processor implementations |
| **ProgressTracker** | Dual-write progress (Redis for fast UI, DB for durability) |
| **RecoveryService** | Finds and re-enqueues stale/orphaned jobs on startup and periodically |
| **Orchestrator** | Parent processor that creates and monitors child jobs in waves |
| **Scheduler** | Periodic and daily catch-up job creation |
| **EnqueueService** | API layer: duplicate detection, atomic create+enqueue, cancellation |

---

## 2. Job Entity & State Machine

### Entity Definition

```
Job {
    id               : UUID
    resource_id      : UUID          // [YOUR PROJECT] The entity this job operates on
    owner_id         : UUID          // [YOUR PROJECT] The user/tenant who owns this job
    job_type         : string        // [YOUR PROJECT] e.g., "full_sync", "export", "import"
    parent_job_id    : UUID | null   // Link to parent (for orchestrated child jobs)
    status           : JobStatus     // pending | processing | completed | failed | cancelled | partial_failure
    priority         : int           // 0 = normal, 1 = high
    total_items      : int           // Progress numerator
    completed_items  : int           // Progress denominator
    entity_type      : string        // Subtype for display/routing (optional)
    error_message    : string
    worker_id        : string        // Which worker is processing this job
    started_at       : timestamp | null
    completed_at     : timestamp | null
    created_at       : timestamp
    updated_at       : timestamp
}
```

### State Machine

```
                    ┌──────────┐
                    │ pending  │
                    └────┬─────┘
                         │ MarkStarted (requires status='pending')
                         ▼
                    ┌──────────────┐
                    │  processing  │
                    └──┬───┬───┬───┘
           MarkCompleted│  │   │MarkFailed (requires status IN ('pending','processing'))
       (requires        │  │   │
       status=          │  │   │
       'processing')    │  │   │
                        ▼  │   ▼
              ┌──────────┐ │ ┌────────┐
              │completed │ │ │ failed │
              └──────────┘ │ └────────┘
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
        ┌───────────┐          ┌─────────────────┐
        │ cancelled │          │ partial_failure  │
        └───────────┘          └─────────────────┘
```

**Critical: DB status guards on every transition.** Each `Mark*` operation includes a `WHERE status = <expected>` clause. This prevents concurrent workers from double-transitioning a job.

### Key Functions

```
IsTerminal(job) -> bool:
    return job.status IN (completed, failed, cancelled, partial_failure)

NewJob(resource_id, owner_id, job_type, priority) -> Job:
    return Job {
        id:          generate_uuid(),
        resource_id: resource_id,
        owner_id:    owner_id,
        job_type:    job_type,
        status:      "pending",
        priority:    priority,
        created_at:  now_utc(),
        updated_at:  now_utc(),
    }

NewChildJob(parent_job, job_type, entity_type) -> Job:
    child = NewJob(parent_job.resource_id, parent_job.owner_id, job_type, parent_job.priority)
    child.parent_job_id = parent_job.id
    child.entity_type   = entity_type
    return child
```

---

## 3. Queue

### Redis FIFO Pattern

Use LPUSH (enqueue) and BRPOP (dequeue) for reliable FIFO ordering.

### Queue Keys

```
REGULAR_QUEUE_KEY      = "app:sync:queue"              // All normal job types
ORCHESTRATOR_QUEUE_KEY = "app:sync:queue:orchestrator"  // Orchestrator jobs only

// [YOUR PROJECT] Add more queue keys if you need priority lanes
```

Separate queues ensure orchestrator jobs (which poll for children) don't block regular workers.

### Payload Structure

The payload contains everything needed to start processing without a DB lookup:

```
QueuePayload {
    job_id            : UUID
    resource_id       : UUID      // [YOUR PROJECT] e.g., app_id, account_id
    owner_id          : UUID      // [YOUR PROJECT] e.g., user_id, tenant_id
    job_type          : string
    parent_job_id     : UUID | null
    priority          : int
    entity_type       : string    // Optional subtype
    lookback_days     : int       // 0 = default window (for catch-up syncs)
    enqueued_at       : timestamp
}
```

### Enqueue / Dequeue

```
function Enqueue(kv_client, payload):
    data = json_encode(payload)
    queue_key = QueueKeyForJobType(payload.job_type)
    kv_client.LPUSH(queue_key, data)

function Dequeue(kv_client, queue_key, timeout):
    result = kv_client.BRPOP(queue_key, timeout)
    if result is null:
        return null  // Timeout, no item
    return json_decode(result)

function QueueKeyForJobType(job_type):
    if job_type == "full_sync":    // [YOUR PROJECT] your orchestrator type
        return ORCHESTRATOR_QUEUE_KEY
    return REGULAR_QUEUE_KEY
```

---

## 4. Lock Manager

All lock operations use **atomic Lua scripts** to prevent race conditions in multi-node deployments.

### Key Format

```
Lock key:        "app:sync:lock:{resource_id}:{job_type}"
Heartbeat key:   "app:sync:heartbeat:{job_id}"
Cancellation key: "app:sync:cancel:{job_id}"
```

### Timing Constants

| Constant | Value | Rationale |
|----------|-------|-----------|
| `LOCK_TTL` | 2 hours | Must exceed max job duration. Auto-expires if worker dies without cleanup. |
| `HEARTBEAT_TTL` | 20 minutes | Heartbeat key expires if worker stops writing. Used by recovery to detect stale jobs. |
| `HEARTBEAT_INTERVAL` | 10 minutes | How often workers renew heartbeat. Must be < `HEARTBEAT_TTL`. |
| `LOCK_EXTENSION_INTERVAL` | 1 hour | How often workers extend lock TTL. Must be < `LOCK_TTL`. |

### 5 Lock Operations

**1. AcquireLock** — Try to acquire via SETNX (set-if-not-exists):
```
function AcquireLock(kv, resource_id, job_type, worker_id):
    key = lock_key(resource_id, job_type)
    return kv.SET_NX(key, worker_id, ttl=LOCK_TTL)
```

**2. ReleaseLockIfOwner** — Atomic release only if caller owns the lock:
```lua
-- Lua script: ReleaseLockIfOwner
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
```

**3. ExtendLockIfOwner** — Atomic TTL extension only if caller owns the lock:
```lua
-- Lua script: ExtendLockIfOwner
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
```

**4. StealLock** — Atomic steal: verify current holder, delete, re-acquire:
```lua
-- Lua script: StealLock
local current = redis.call("GET", KEYS[1])
if current == ARGV[1] then
    redis.call("DEL", KEYS[1])
    return redis.call("SET", KEYS[1], ARGV[2], "NX", "PX", ARGV[3]) and 1 or 0
end
return 0
```
Used when a worker detects the lock holder has no heartbeat (dead worker).

**5. ForceReleaseLock** — Unconditional delete (recovery only):
```
function ForceReleaseLock(kv, resource_id, job_type):
    kv.DEL(lock_key(resource_id, job_type))
```
Only used by RecoveryService for legitimately clearing stale locks from crashed nodes.

### Heartbeat & Cancellation

```
function Heartbeat(kv, job_id):
    kv.SET(heartbeat_key(job_id), now_utc(), ttl=HEARTBEAT_TTL)

function HasHeartbeat(kv, job_id):
    return kv.EXISTS(heartbeat_key(job_id))

function DeleteHeartbeat(kv, job_id):
    kv.DEL(heartbeat_key(job_id))

function RequestCancellation(kv, job_id):
    kv.SET(cancel_key(job_id), "1", ttl=LOCK_TTL)

function IsCancelled(kv, job_id):
    return kv.EXISTS(cancel_key(job_id))

function CleanupCancellation(kv, job_id):
    kv.DEL(cancel_key(job_id))
```

---

## 5. Worker Pool

The worker pool is the most critical component. The `processJob` sequence below encodes fixes for bugs #1-14 — **the order of operations matters**.

### Worker Loop

```
function WorkerLoop(ctx, worker_id, queue_key):
    loop:
        if ctx.is_cancelled():
            return

        payload = Dequeue(kv, queue_key, timeout=5s)
        if payload is null:
            continue  // Timeout, try again

        processJob(ctx, worker_id, payload)
```

### The 15-Step processJob Sequence

This sequence is the result of 14 production bug fixes. Every step is intentionally ordered.

```
function processJob(ctx, worker_id, payload):
    job_id = payload.job_id

    // ── STEP 1: Acquire lock BEFORE MarkStarted ──
    // (Bug 2 fix: prevents processing→pending bounce)
    locked = LockManager.AcquireLock(payload.resource_id, payload.job_type, worker_id)

    // ── STEP 2: If lock failed, attempt steal from dead holder ──
    if not locked:
        existing_holder = LockManager.GetLockHolder(payload.resource_id, payload.job_type)
        if existing_holder != "":
            existing_job = JobRepo.FindActiveByResourceAndType(payload.resource_id, payload.job_type)
            if existing_job != null AND existing_job.id != job_id:
                has_hb = LockManager.HasHeartbeat(existing_job.id)
                if not has_hb:
                    // Bug 4 fix: Atomic steal instead of separate release+acquire
                    locked = LockManager.StealLock(payload.resource_id, payload.job_type,
                                                   existing_holder, worker_id)

    // ── STEP 3: If still no lock, re-enqueue with backoff ──
    if not locked:
        reEnqueueWithBackoff(ctx, worker_id, job_id, payload)
        return

    // ── STEP 4: Write initial heartbeat immediately ──
    // Prevents steal-lock race: another worker checking HasHeartbeat
    // between lock acquisition and heartbeat goroutine starting.
    LockManager.Heartbeat(job_id)

    // ── STEP 5: Mark job as started (after lock acquired) ──
    error = JobRepo.MarkStarted(job_id, worker_id)
    if error:
        LockManager.ReleaseLockIfOwner(payload.resource_id, payload.job_type, worker_id)
        LockManager.DeleteHeartbeat(job_id)
        return

    // ── STEP 6: Start heartbeat background task ──
    heartbeat_handle = start_background(heartbeatLoop, job_id, payload, worker_id)

    // ── STEP 7: Look up processor ──
    processor = ProcessorRegistry.Get(payload.job_type)
    if processor is null:
        stop(heartbeat_handle)
        JobRepo.MarkFailed(job_id, "no processor for job type")
        cleanup(job_id, payload, worker_id)
        return

    // ── STEP 8: Execute processor ──
    error = processor.Process(ctx, payload)

    // ── STEP 9: Stop heartbeat ──
    stop(heartbeat_handle)

    // ── STEP 10: Use background context for final state transitions ──
    // Parent ctx may be cancelled (server shutdown), but we MUST still
    // persist the final job state and release locks.
    cleanup_ctx = new_context_with_timeout(10s)

    // ── STEP 11: Handle shutdown interruption ──
    if error AND ctx.is_cancelled():
        // Leave in 'processing' state — recovery re-enqueues on next startup
        cleanup(cleanup_ctx, job_id, payload, worker_id)
        return

    // ── STEP 12: Handle processor failure ──
    if error:
        // Bug 13 fix: Check cancellation before MarkFailed
        if LockManager.IsCancelled(job_id):
            // Don't overwrite cancelled status with "failed"
            skip
        else:
            JobRepo.MarkFailed(cleanup_ctx, job_id, error.message)

    // ── STEP 13: Handle processor success ──
    if not error:
        JobRepo.MarkCompleted(cleanup_ctx, job_id)

    // ── STEP 14: Cleanup ──
    cleanup(cleanup_ctx, job_id, payload, worker_id)

// ── STEP 15: Cleanup function ──
function cleanup(ctx, job_id, payload, worker_id):
    LockManager.ReleaseLockIfOwner(payload.resource_id, payload.job_type, worker_id)
    LockManager.DeleteHeartbeat(job_id)
    LockManager.CleanupCancellation(job_id)
    ProgressTracker.Cleanup(job_id)
```

### Heartbeat Background Task

Runs two periodic actions: heartbeat renewal and lock TTL extension.

```
function heartbeatLoop(ctx, job_id, resource_id, job_type, worker_id):
    hb_ticker  = new_ticker(HEARTBEAT_INTERVAL)
    ext_ticker = new_ticker(LOCK_EXTENSION_INTERVAL)

    // Initial heartbeat
    LockManager.Heartbeat(job_id)

    loop:
        select:
            case ctx.done():
                return
            case hb_ticker.tick():
                LockManager.Heartbeat(job_id)
            case ext_ticker.tick():
                // Bug 12 fix: Ownership-aware lock extension
                LockManager.ExtendLockIfOwner(resource_id, job_type, worker_id)
```

### Re-enqueue with Backoff

```
function reEnqueueWithBackoff(ctx, worker_id, job_id, payload):
    wait 5 seconds OR until ctx.cancelled:
        if ctx.cancelled:
            // Bug 7 fix: Best-effort enqueue before returning on shutdown
            Enqueue(background_ctx, kv, payload)
            return
    Enqueue(ctx, kv, payload)
```

### Graceful Shutdown

```
function WorkerPool.Stop():
    cancel(ctx)        // Signal all workers to stop
    wait_group.wait()  // Wait for all workers to finish current job
```

Key behavior: On shutdown, jobs are left in `processing` state (not marked `failed`). The heartbeat and lock are cleaned up. RecoveryService finds these on next startup and re-enqueues them.

---

## 6. Processor Registry & Interface

### Interface

```
interface Processor:
    Type() -> string
    Process(ctx, payload) -> error
```

**Critical rule:** Processors NEVER call `MarkCompleted` or `MarkFailed`. They return:
- `nil/null` on success — worker calls `MarkCompleted`
- `error` on failure — worker calls `MarkFailed`

This centralizes state transitions in the worker (Bug 3 fix).

### Registry

```
ProcessorRegistry:
    processors: map<string, Processor>

    Register(processor):
        processors[processor.Type()] = processor

    Get(job_type) -> Processor | error:
        if job_type not in processors:
            return error("no processor for job type")
        return processors[job_type]
```

---

## 7. Processor Context

Many processors share the same setup: look up the resource, load credentials, decrypt tokens. Extract this into a reusable preamble.

```
ProcessorContext {
    resource     : Resource         // [YOUR PROJECT] Your domain resource
    credentials  : Credentials      // [YOUR PROJECT] Decrypted API tokens, etc.
    // ... any shared setup data
}

function PrepareProcessorContext(ctx, payload, resource_repo, credential_repo, decryptor):
    // Step 1: Look up the resource
    resource = resource_repo.FindByID(payload.resource_id)
    if resource is null:
        return error("resource not found")

    // Step 2: Look up credentials
    credentials = credential_repo.FindByID(payload.credential_id)
    if credentials is null:
        return error("credentials not found")

    // Step 3: Decrypt token
    // [YOUR PROJECT] Replace with your token/secret management
    token = decryptor.Decrypt(credentials.encrypted_token)

    return ProcessorContext { resource, credentials, token }
```

Usage in a processor:

```
function MyProcessor.Process(ctx, payload):
    pctx = PrepareProcessorContext(ctx, payload, ...)
    if pctx is error:
        return pctx.error

    // Check cancellation
    if LockManager.IsCancelled(payload.job_id):
        return error("job cancelled")

    // [YOUR PROJECT] Your business logic here
    data = fetch_data(pctx.token, ...)
    process(data)
    store(data)

    // Check cancellation again between major steps
    if LockManager.IsCancelled(payload.job_id):
        return error("job cancelled")

    // More business logic...
    return nil  // Success — worker will call MarkCompleted
```

---

## 8. Orchestrator Pattern

The orchestrator is a special processor that creates child jobs, enqueues them, and waits for completion. It implements wave-based execution where later waves depend on earlier waves.

### Wave Definitions

```
// [YOUR PROJECT] Define your waves based on data dependencies

WAVE_1 = [                          // Independent jobs (run in parallel)
    { type: "data_fetch",   entity: "record" },
    { type: "review_sync",  entity: "review" },
]

WAVE_2 = [                          // Depends on Wave 1 completing
    { type: "event_sync",   entity: "event" },
    { type: "snapshot",     entity: "snapshot" },
    { type: "status_check", entity: "status" },
]
```

### Orchestrator Processor

```
function OrchestratorProcessor.Process(ctx, payload):
    parent_job = JobRepo.FindByID(payload.job_id)

    total_children = len(WAVE_1) + len(WAVE_2)
    ProgressTracker.Update(payload.job_id, { total: total_children, message: "Starting Wave 1..." })

    // ── Dispatch Wave 1 ──
    wave1_ids = []
    for each job_def in WAVE_1:
        child = NewChildJob(parent_job, job_def.type, job_def.entity)
        JobRepo.Create(child)
        child_payload = build_payload(child)
        error = Enqueue(kv, child_payload)
        if error:
            JobRepo.MarkFailed(child.id, error.message)
            continue
        wave1_ids.append(child.id)

    // Bug 8 fix: Guard against empty wave (all enqueues failed)
    if len(wave1_ids) == 0:
        return error("failed to enqueue any Wave 1 jobs")

    // ── Wait for Wave 1 ──
    error = waitForChildren(ctx, payload, wave1_ids, "Waiting for Wave 1...")
    if error:
        return error

    ProgressTracker.Update(payload.job_id, { completed: len(WAVE_1), message: "Starting Wave 2..." })

    // ── Dispatch Wave 2 ──
    wave2_ids = []
    for each job_def in WAVE_2:
        child = NewChildJob(parent_job, job_def.type, job_def.entity)
        JobRepo.Create(child)
        child_payload = build_payload(child)
        error = Enqueue(kv, child_payload)
        if error:
            JobRepo.MarkFailed(child.id, error.message)
            continue
        wave2_ids.append(child.id)

    // ── Wait for all remaining ──
    all_remaining = wave1_ids[1:] + wave2_ids   // Skip first Wave 1 (already waited)
    error = waitForChildren(ctx, payload, all_remaining, "Waiting for all jobs...")
    if error:
        return error

    // ── Check final status ──
    children = JobRepo.FindByParentJobID(payload.job_id)
    has_failure = any(child.status == "failed" for child in children)

    if has_failure:
        ProgressTracker.ForceUpdate(payload.job_id, { total: total_children, completed: total_children,
            message: "Completed with some failures" })
        // Use direct status update for partial_failure — worker won't call MarkCompleted for this
        JobRepo.UpdateStatus(payload.job_id, "partial_failure")
        return nil  // Return nil so worker doesn't also call MarkFailed

    ProgressTracker.ForceUpdate(payload.job_id, { total: total_children, completed: total_children,
        message: "Complete" })
    return nil
```

### waitForChildren

```
function waitForChildren(ctx, payload, child_ids, progress_msg):
    POLL_INTERVAL = 5 seconds

    loop:
        if ctx.is_cancelled():
            return ctx.error()

        sleep(POLL_INTERVAL)

        // Check cancellation
        if LockManager.IsCancelled(payload.job_id):
            return error("job cancelled")

        children = JobRepo.FindByParentJobID(payload.job_id)
        child_set = set(child_ids)

        all_done = true
        for each child in children:
            if child.id in child_set AND not IsTerminal(child):
                all_done = false
                break

        if all_done:
            return nil

        ProgressTracker.Update(payload.job_id, { message: progress_msg })
```

---

## 9. Progress Tracking

Progress uses a **dual-write** strategy: Redis for fast UI polling, DB for durability.

### Timing

| Target | Interval | Purpose |
|--------|----------|---------|
| Redis | Every 2 seconds | Fast UI updates (poll from frontend) |
| DB | Every 30 seconds | Durable (survives Redis flush) |

### ProgressTracker

```
Progress {
    total     : int
    completed : int
    message   : string
}

ProgressTracker:
    kv_client       : KV
    job_repo        : JobRepository
    redis_interval  : duration = 2s
    db_interval     : duration = 30s
    last_redis_write: map<UUID, timestamp>
    last_db_write   : map<UUID, timestamp>

function Update(ctx, job_id, progress):
    now = now()

    // Throttled write to Redis
    if now - last_redis_write[job_id] >= redis_interval:
        kv.HSET("app:sync:progress:{job_id}", {
            total:     progress.total,
            completed: progress.completed,
            message:   progress.message,
        })
        kv.EXPIRE("app:sync:progress:{job_id}", 2h)
        last_redis_write[job_id] = now

    // Throttled write to DB
    if now - last_db_write[job_id] >= db_interval:
        job_repo.UpdateProgress(job_id, progress.total, progress.completed)
        last_db_write[job_id] = now

function ForceUpdate(ctx, job_id, progress):
    // Write immediately to both (for milestones like completion)
    kv.HSET(...)
    job_repo.UpdateProgress(...)
    last_redis_write[job_id] = now
    last_db_write[job_id] = now

function GetProgress(ctx, job_id) -> Progress | null:
    result = kv.HGETALL("app:sync:progress:{job_id}")
    if empty(result):
        return null
    return Progress { total: result.total, completed: result.completed, message: result.message }

function Cleanup(ctx, job_id):
    kv.DEL("app:sync:progress:{job_id}")
    delete(last_redis_write, job_id)
    delete(last_db_write, job_id)
```

---

## 10. Cooperative Cancellation

Cancellation uses a **Redis flag pattern**, not context cancellation. This ensures clean shutdown — processors finish their current atomic step before stopping.

### Why Not Context Cancellation?

Context cancellation kills the operation mid-stream. If a processor is halfway through writing to the database, context cancel can leave partial writes. Cooperative cancellation lets the processor reach a safe checkpoint before stopping.

### Pattern

**Setting the flag (API layer):**
```
function CancelJob(job_id):
    job = JobRepo.FindByID(job_id)
    if IsTerminal(job):
        return error("job already in terminal state")

    // Set cancellation flag in KV store
    LockManager.RequestCancellation(job_id)

    // For orchestrator jobs, cancel all children
    if job.job_type == "full_sync":   // [YOUR PROJECT] your orchestrator type
        children = JobRepo.FindByParentJobID(job_id)
        for each child in children:
            if not IsTerminal(child):
                LockManager.RequestCancellation(child.id)

    JobRepo.UpdateStatus(job_id, "cancelled")
```

**Checking in processors (between major steps):**
```
function MyProcessor.Process(ctx, payload):
    // Step 1: Fetch data
    data = fetch_external_data(...)

    // Checkpoint
    if LockManager.IsCancelled(payload.job_id):
        return error("job cancelled")

    // Step 2: Process data
    results = process(data)

    // Checkpoint
    if LockManager.IsCancelled(payload.job_id):
        return error("job cancelled")

    // Step 3: Store results
    store(results)
    return nil
```

**Worker integration (Bug 13 fix):**
```
// In processJob, before MarkFailed:
if error:
    if LockManager.IsCancelled(job_id):
        // Don't overwrite cancelled status with "failed"
        skip MarkFailed
    else:
        JobRepo.MarkFailed(job_id, error.message)
```

---

## 11. Recovery Service

Recovery handles two scenarios: startup recovery (after crash/restart) and periodic recovery (for jobs that slip through).

### Startup Recovery

```
function RecoverOnStartup(ctx):
    // 1. Find all jobs stuck in 'processing'
    processing_jobs = JobRepo.FindByStatus("processing")

    stale_jobs = []
    stale_ids  = set()

    // First pass: identify stale jobs (no heartbeat = dead worker)
    for each job in processing_jobs:
        has_hb = LockManager.HasHeartbeat(job.id)
        if has_hb:
            continue  // Worker is alive
        stale_jobs.append(job)
        stale_ids.add(job.id)

    // Second pass: re-enqueue only parents and orphans
    // Skip children whose parent is also being recovered — parent will recreate them
    for each job in stale_jobs:
        if job.parent_job_id != null AND job.parent_job_id in stale_ids:
            // Parent will recreate this child on re-run
            JobRepo.MarkFailed(job.id, "parent recovered — will be recreated")
            LockManager.ForceReleaseLock(job.resource_id, job.job_type)
            LockManager.DeleteHeartbeat(job.id)
            continue

        LockManager.ForceReleaseLock(job.resource_id, job.job_type)
        LockManager.DeleteHeartbeat(job.id)
        reEnqueueJob(job)

    // 3. Re-enqueue orphaned pending jobs (handles Redis flush)
    // Bug 14 fix: Do NOT release locks for pending jobs — they don't hold locks.
    pending_jobs = JobRepo.FindByStatus("pending")
    for each job in pending_jobs:
        if job.id in stale_ids:
            continue  // Already handled
        if job.parent_job_id != null AND job.parent_job_id in stale_ids:
            JobRepo.MarkFailed(job.id, "parent recovered — will be recreated")
            continue
        // Just re-enqueue to Redis (already pending in DB)
        Enqueue(kv, build_payload(job))
```

### Periodic Recovery

```
RECOVERY_GRACE_PERIOD = 2 minutes   // Bug 5 fix: don't recover jobs that just started

function StartPeriodicRecovery(ctx, interval):
    every interval:
        processing_jobs = JobRepo.FindByStatus("processing")
        for each job in processing_jobs:
            // Grace period — skip recently started jobs
            if job.started_at != null AND time_since(job.started_at) < RECOVERY_GRACE_PERIOD:
                continue

            // Only recover jobs older than lock TTL without heartbeat
            if job.started_at != null AND time_since(job.started_at) > LOCK_TTL:
                has_hb = LockManager.HasHeartbeat(job.id)
                if has_hb:
                    continue
                LockManager.ForceReleaseLock(job.resource_id, job.job_type)
                LockManager.DeleteHeartbeat(job.id)
                reEnqueueJob(job)
```

### Conditional Re-enqueue

```
// Bug 10 fix: Use conditional status update to avoid racing with workers
function reEnqueueJob(job):
    if job.status == "processing":
        // Only reset to pending if CURRENTLY processing (WHERE status='processing')
        error = JobRepo.MarkPendingIfProcessing(job.id)
        if error:
            return  // Job already moved to another state — skip

    Enqueue(kv, build_payload(job))
```

---

## 12. Enqueue Service

The API layer that creates jobs and enqueues them. Three variants handle different use cases.

### Standard Enqueue (User-Triggered)

```
function EnqueueSync(ctx, resource_id, owner_id, job_type, priority) -> Job | error:
    // Duplicate detection
    existing = JobRepo.FindActiveByResourceAndType(resource_id, job_type)
    if existing != null:
        return error("DUPLICATE: active job already exists")  // Return 409 to API

    // Create job in DB
    job = NewJob(resource_id, owner_id, job_type, priority)
    JobRepo.Create(job)

    // Enqueue to KV store
    payload = build_payload(job)
    error = Enqueue(kv, payload)
    if error:
        // Atomic: mark DB row as failed if KV enqueue fails
        JobRepo.MarkFailed(job.id, "enqueue failed: " + error.message)
        return error

    return job
```

### Catchup Enqueue (Scheduler-Triggered)

```
function EnqueueCatchupSync(ctx, resource_id, owner_id, job_type, lookback_days) -> Job | null:
    // Duplicate detection — silently skip (no error)
    existing = JobRepo.FindActiveByResourceAndType(resource_id, job_type)
    if existing != null:
        return null  // Already running — silent skip

    job = NewJob(resource_id, owner_id, job_type, priority=0)
    JobRepo.Create(job)

    payload = build_payload(job)
    payload.lookback_days = lookback_days
    error = Enqueue(kv, payload)
    if error:
        JobRepo.MarkFailed(job.id, "enqueue failed")
        return error

    return job
```

### Trigger Sync (Fire-and-Forget)

```
function TriggerSync(ctx, resource_id, owner_id) -> error:
    // High priority, swallow duplicate errors
    _, error = EnqueueSync(ctx, resource_id, owner_id, "full_sync", priority=1)
    if error is DUPLICATE:
        return nil  // Already running — not an error
    return error
```

### Job Progress Query

```
function GetJobProgress(ctx, job_id) -> JobProgress:
    job = JobRepo.FindByID(job_id)
    progress = { job: job, total: job.total_items, completed: job.completed_items }

    // Overlay Redis progress for active jobs (more current than DB)
    if job.status == "processing":
        redis_progress = ProgressTracker.GetProgress(job_id)
        if redis_progress != null:
            progress.total     = redis_progress.total
            progress.completed = redis_progress.completed
            progress.message   = redis_progress.message

    // Include children for orchestrator jobs
    if job.job_type == "full_sync":   // [YOUR PROJECT] your orchestrator type
        children = JobRepo.FindByParentJobID(job_id)
        for each child in children:
            child_progress = { job: child, ... }
            // Same Redis overlay logic
            progress.children.append(child_progress)

    return progress
```

---

## 13. Schedulers

### Periodic Full Sync

```
PeriodicScheduler:
    interval  : duration = 12 hours
    stop_ch   : channel
    done_ch   : channel

function Start(ctx):
    start_background:
        syncAll(ctx)  // Run immediately on start
        every interval:
            syncAll(ctx)

function syncAll(ctx):
    // [YOUR PROJECT] Iterate over all resources that need periodic syncing
    resource_ids = ResourceRepo.GetAllIDs(ctx)
    for each id in resource_ids:
        TriggerSync(ctx, id, ...)  // Duplicate-safe

function Stop():
    signal(stop_ch)
    wait(done_ch)
```

### Daily Catch-Up Sync

```
DailyCatchupScheduler:
    target_hour    : int = 3           // UTC hour to run (off-peak)
    lookback_days  : int = 2           // Days to look back
    check_interval : duration = 15min  // How often to check if it's time
    last_run_date  : string = ""       // "YYYY-MM-DD" to prevent double-runs

function Start(ctx):
    start_background:
        check(ctx)  // Check immediately on start
        every check_interval:
            check(ctx)

function check(ctx):
    now = now_utc()
    if now.hour != target_hour:
        return
    today = format(now, "YYYY-MM-DD")
    if today == last_run_date:
        return  // Already ran today
    last_run_date = today
    enqueueAll(ctx, lookback_days)

function enqueueAll(ctx, lookback_days):
    // [YOUR PROJECT] Iterate all resources, enqueue lightweight sync jobs
    resources = ResourceRepo.GetAll(ctx)
    for each resource in resources:
        EnqueueCatchupSync(ctx, resource.id, resource.owner_id,
                          "data_fetch", lookback_days)  // Duplicate-safe

function Stop():
    signal(stop_ch)
    wait(done_ch)
```

---

## 14. Pitfall Catalog

### Critical Severity

**Pitfall 1: Lock Release Without Ownership Check**
- **Symptom:** Worker A finishes slowly and calls `DEL lock_key`, deleting Worker B's lock. Worker C acquires the now-free lock. Two workers process the same resource simultaneously.
- **Root cause:** Simple `DEL` doesn't verify who holds the lock.
- **Fix:** Use Lua script `ReleaseLockIfOwner` — only delete if value matches caller's worker ID.
- **Pseudocode:**
  ```lua
  if redis.call("GET", KEYS[1]) == ARGV[1] then
      return redis.call("DEL", KEYS[1])
  end
  return 0
  ```

**Pitfall 2: MarkStarted Before Lock Acquisition**
- **Symptom:** Job bounces between `processing` and `pending`. Worker marks job `processing`, fails to acquire lock, recovery resets to `pending`.
- **Root cause:** `MarkStarted` called before `AcquireLock`. If lock fails, job is already in `processing` state but no worker owns it.
- **Fix:** Always acquire lock FIRST, then `MarkStarted`. If lock fails, job stays `pending` — no state corruption.

**Pitfall 3: Processors Calling MarkCompleted/MarkFailed**
- **Symptom:** Successfully processed job is marked `failed`. Job completes, processor calls `MarkCompleted`, DB write times out, worker falls through to `MarkFailed`.
- **Root cause:** Dual state management — both processor and worker try to set final status.
- **Fix:** Processors return `nil` (success) or `error` (failure). Only the worker calls `MarkCompleted`/`MarkFailed`. Single point of control.

**Pitfall 4: Non-Atomic Lock Stealing**
- **Symptom:** Two recovery workers both detect a dead lock holder. Both delete and re-acquire. One gets the lock, the other overwrites it.
- **Root cause:** Separate `DEL` + `SET NX` operations have a race window.
- **Fix:** Single Lua script `StealLock` — check holder, delete, set NX atomically.

### High Severity

**Pitfall 5: Recovery Stealing From Slow-Starting Workers**
- **Symptom:** A just-started job is recovered and re-enqueued. The original worker hasn't written its first heartbeat yet.
- **Root cause:** Recovery checks for heartbeat immediately. Worker takes a moment to start the heartbeat goroutine.
- **Fix:** Grace period (2 minutes) in periodic recovery. Also: write initial heartbeat synchronously right after lock acquisition, before starting the heartbeat goroutine.

**Pitfall 6: Shutdown Marks Jobs as Failed**
- **Symptom:** After Ctrl+C/deployment restart, all in-flight jobs are marked `failed` (a terminal state). Recovery ignores terminal jobs. Jobs never run again.
- **Root cause:** `processor.Process()` returns `context.Canceled` on shutdown. Worker treats this as failure.
- **Fix:** Check `ctx.is_cancelled()` before `MarkFailed`. If shutdown, leave job in `processing` state. Delete heartbeat + release lock. Recovery finds it on next boot (no heartbeat = stale) and re-enqueues.

**Pitfall 7: Lost Jobs During Backoff on Shutdown**
- **Symptom:** Worker is waiting in the 5-second backoff sleep when shutdown signal arrives. Job is never re-enqueued — lost forever.
- **Root cause:** `sleep(5s)` ignores context cancellation.
- **Fix:** Use `select { case sleep(5s); case ctx.cancelled }`. On cancellation, best-effort enqueue with a background context before returning.

**Pitfall 8: Empty Wave in Orchestrator**
- **Symptom:** Orchestrator dispatches Wave 1, all enqueues fail, then calls `waitForChildren` with an empty list. `waitForChildren` returns immediately (all done). Wave 2 dispatches without Wave 1 completing.
- **Root cause:** No guard against empty `wave_ids` list.
- **Fix:** After dispatching a wave, check `len(wave_ids) == 0`. If zero, return error immediately.

**Pitfall 9: Lock Extension Without Ownership**
- **Symptom:** Worker A's heartbeat goroutine extends a lock that Worker B now owns. Worker B's lock TTL gets reset unpredictably.
- **Root cause:** Simple `EXPIRE` doesn't check who holds the lock.
- **Fix:** Use Lua script `ExtendLockIfOwner` — only extend TTL if value matches caller's worker ID.

### Medium Severity

**Pitfall 10: Recovery Race With Active Workers**
- **Symptom:** Recovery resets a job to `pending` while a worker is actively processing it. Job runs twice simultaneously.
- **Root cause:** Recovery does unconditional `UPDATE status='pending'` without checking current state.
- **Fix:** Use `MarkPendingIfProcessing` — conditional update with `WHERE status='processing'`. If worker has already moved job to `completed`/`failed`, the update affects 0 rows and recovery skips it.

**Pitfall 11: Cancelled Job Marked as Failed**
- **Symptom:** User cancels a job, but it shows as `failed` instead of `cancelled`. Cancellation flag is set, processor returns error "job cancelled", worker calls `MarkFailed`.
- **Root cause:** Worker doesn't check cancellation status before `MarkFailed`.
- **Fix:** Before `MarkFailed`, check `IsCancelled()`. If cancelled, skip `MarkFailed` — the cancel API already set `cancelled` status.

**Pitfall 12: Heartbeat-Lock Race Window**
- **Symptom:** Worker A acquires lock. Before heartbeat goroutine starts, Worker B checks `HasHeartbeat` for Worker A — returns false. Worker B steals the lock.
- **Root cause:** Gap between lock acquisition and first heartbeat write.
- **Fix:** Write initial heartbeat synchronously immediately after lock acquisition, before starting the heartbeat goroutine.

**Pitfall 13: Recovery Releases Locks for Pending Jobs**
- **Symptom:** Recovery finds pending jobs and calls `ForceReleaseLock` for each. This releases locks belonging to active workers processing the same resource+type.
- **Root cause:** Pending jobs shouldn't hold locks, but recovery was treating them the same as stale processing jobs.
- **Fix:** Only call `ForceReleaseLock` for `processing` jobs. For `pending` jobs, just re-enqueue to Redis without touching locks.

**Pitfall 14: Recovery Re-enqueues Children of Recovered Parents**
- **Symptom:** Parent orchestrator job and its children are both stale. Recovery re-enqueues all of them. Parent recreates fresh children. Old children also run. Duplicate work, possible conflicts.
- **Root cause:** Recovery doesn't check parent-child relationships.
- **Fix:** Two-pass recovery. First pass: identify all stale job IDs. Second pass: skip children whose parent is also in the stale set. Mark those children as `failed` with message "parent recovered — will be recreated".

---

## 15. Implementation Checklist

### Quick Start (MVP) — Steps 1-8

These steps give you a working async job system with basic locking.

**Step 1: Job entity + state machine**
- Define the Job entity with all fields from Section 2
- Implement `IsTerminal()`, `NewJob()`, `NewChildJob()`
- Define your job types (`[YOUR PROJECT]`)

**Step 2: Job repository (DB CRUD + guarded transitions)**
- `Create(job)` — INSERT
- `FindByID(id)` — SELECT
- `FindByStatus(status)` — SELECT WHERE status=?
- `FindActiveByResourceAndType(resource_id, type)` — SELECT WHERE status IN ('pending','processing')
- `MarkStarted(id, worker_id)` — UPDATE SET status='processing' WHERE id=? AND status='pending'
- `MarkCompleted(id)` — UPDATE SET status='completed' WHERE id=? AND status='processing'
- `MarkFailed(id, message)` — UPDATE SET status='failed' WHERE id=? AND status IN ('pending','processing')
- `UpdateProgress(id, total, completed)` — UPDATE
- **Critical:** Every `Mark*` function has a `WHERE status=<expected>` guard

**Step 3: Queue (enqueue/dequeue with Redis)**
- Define queue key constants
- Implement `Enqueue()` (LPUSH) and `Dequeue()` (BRPOP)
- Implement `QueueKeyForJobType()` for routing

**Step 4: Basic lock manager (acquire + force-release)**
- `AcquireLock()` via SETNX
- `ForceReleaseLock()` via DEL
- Define key format and timing constants

**Step 5: Processor interface + registry**
- Define the `Processor` interface (Type, Process)
- Implement `ProcessorRegistry` (Register, Get)

**Step 6: Worker pool (single queue, single worker)**
- Implement worker loop (dequeue → process)
- Basic `processJob`: acquire lock → mark started → process → mark completed/failed → cleanup
- No heartbeat yet, no steal logic yet

**Step 7: First processor implementation**
- Implement one processor for your primary job type
- Return nil on success, error on failure
- Do NOT call MarkCompleted/MarkFailed from the processor

**Step 8: Enqueue API endpoint**
- Implement `EnqueueSync()` with duplicate detection
- Wire up HTTP endpoint to create + enqueue jobs

### Production Hardening — Steps 9-18

Add these incrementally after the MVP is working.

**Step 9: Ownership-aware locks (Lua scripts)**
- Replace ForceReleaseLock calls in worker with `ReleaseLockIfOwner` Lua script
- Add `ExtendLockIfOwner` Lua script
- Add `StealLock` Lua script
- Keep `ForceReleaseLock` for recovery only

**Step 10: Heartbeat system**
- Add `Heartbeat()`, `HasHeartbeat()`, `DeleteHeartbeat()`
- Write initial heartbeat synchronously after lock acquisition

**Step 11: Lock-before-MarkStarted worker refactor**
- Reorder processJob: acquire lock → heartbeat → MarkStarted (was: MarkStarted → lock)
- Add steal-lock logic for dead holders (check HasHeartbeat → StealLock)
- Add re-enqueue with backoff on lock contention

**Step 12: Heartbeat goroutine in worker**
- Start background heartbeat loop after MarkStarted
- Two tickers: heartbeat renewal + lock extension
- Stop heartbeat before final state transition

**Step 13: Cooperative cancellation**
- Add `RequestCancellation()`, `IsCancelled()`, `CleanupCancellation()`
- Add cancellation checks in processors (between major steps)
- Add cancellation check in worker before MarkFailed
- Implement CancelJob API with child propagation

**Step 14: Progress tracker (dual-write)**
- Implement ProgressTracker with throttled Redis + DB writes
- Add `Update()`, `ForceUpdate()`, `GetProgress()`, `Cleanup()`
- Add progress calls to processors
- Implement GetJobProgress API (Redis overlay on DB data)

**Step 15: Recovery service (startup + periodic)**
- Implement `RecoverOnStartup()` — find stale processing + orphaned pending
- Implement two-pass recovery (skip children of recovered parents)
- Implement `StartPeriodicRecovery()` with grace period
- Use `MarkPendingIfProcessing()` for race-safe re-enqueue

**Step 16: Orchestrator processor (waves)**
- Implement orchestrator processor with wave definitions
- Implement `waitForChildren()` with cancellation checks
- Handle partial_failure detection
- Empty wave guard

**Step 17: Schedulers with deduplication**
- Periodic full sync scheduler (configurable interval)
- Daily catch-up scheduler (specific UTC hour, lookback window)
- Both use duplicate-safe enqueue variants

**Step 18: Graceful shutdown handling**
- Worker: check ctx.cancelled before MarkFailed, leave in processing
- Backoff: best-effort re-enqueue on shutdown during backoff
- Use background context for final state transitions + cleanup
- WorkerPool.Stop() cancels context and waits for all workers

---

## 16. Extension Point Registry

| # | Extension Point | Section | What to Customize |
|---|----------------|---------|-------------------|
| 1 | `resource_id` field | 2 | The entity your jobs operate on (e.g., `app_id`, `account_id`, `project_id`) |
| 2 | `owner_id` field | 2 | The user/tenant who owns the job (e.g., `user_id`, `org_id`) |
| 3 | Job types | 2 | Your domain-specific job types (e.g., `"data_export"`, `"report_gen"`, `"etl_pipeline"`) |
| 4 | Additional queue keys | 3 | If you need more than two priority lanes |
| 5 | Queue payload fields | 3 | Domain-specific fields on the payload (e.g., `lookback_days`, `filter_params`) |
| 6 | Queue routing logic | 3 | `QueueKeyForJobType()` — which jobs go to which queues |
| 7 | Processor context setup | 7 | Resource lookup, credential decryption, external service clients |
| 8 | Processor implementations | 6 | Your business logic for each job type |
| 9 | Wave definitions | 8 | Data dependency ordering for orchestrated jobs |
| 10 | Orchestrator job type | 8, 10, 12 | The job type that triggers orchestration (e.g., `"full_sync"`) |
| 11 | Lock/heartbeat TTL values | 4 | Tune for your typical and maximum job durations |

---

## 17. Database Schema

```sql
CREATE TABLE sync_jobs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id       UUID NOT NULL,          -- [YOUR PROJECT] FK to your resource table
    owner_id          UUID NOT NULL,          -- [YOUR PROJECT] FK to your user/tenant table
    job_type          VARCHAR(50) NOT NULL,   -- [YOUR PROJECT] your job type enum
    parent_job_id     UUID REFERENCES sync_jobs(id),
    status            VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority          INT NOT NULL DEFAULT 0,
    total_items       INT NOT NULL DEFAULT 0,
    completed_items   INT NOT NULL DEFAULT 0,
    entity_type       VARCHAR(50),
    error_message     TEXT,
    worker_id         VARCHAR(100),
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_status CHECK (
        status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'partial_failure')
    )
);

-- Index for duplicate detection (find active jobs for a resource+type)
CREATE INDEX idx_sync_jobs_active
    ON sync_jobs (resource_id, job_type)
    WHERE status IN ('pending', 'processing');

-- Index for recovery (find all jobs by status)
CREATE INDEX idx_sync_jobs_status ON sync_jobs (status);

-- Index for parent-child lookups
CREATE INDEX idx_sync_jobs_parent ON sync_jobs (parent_job_id)
    WHERE parent_job_id IS NOT NULL;

-- Index for listing jobs by resource (with sorting)
CREATE INDEX idx_sync_jobs_resource_created
    ON sync_jobs (resource_id, created_at DESC);
```

### Guarded Status Transition Queries

```sql
-- MarkStarted: only from pending
UPDATE sync_jobs
SET status = 'processing', worker_id = $2, started_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status = 'pending';

-- MarkCompleted: only from processing
UPDATE sync_jobs
SET status = 'completed', completed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status = 'processing';

-- MarkFailed: from pending or processing
UPDATE sync_jobs
SET status = 'failed', error_message = $2, completed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status IN ('pending', 'processing');

-- MarkPendingIfProcessing: conditional reset for recovery
UPDATE sync_jobs
SET status = 'pending', worker_id = NULL, started_at = NULL, updated_at = NOW()
WHERE id = $1 AND status = 'processing';

-- UpdateProgress
UPDATE sync_jobs
SET total_items = $2, completed_items = $3, updated_at = NOW()
WHERE id = $1;
```

---

## 18. Redis Key Reference

| Key Pattern | TTL | Purpose |
|-------------|-----|---------|
| `app:sync:queue` | None (persistent) | Regular job queue (LPUSH/BRPOP) |
| `app:sync:queue:orchestrator` | None (persistent) | Orchestrator job queue (separate pool) |
| `app:sync:lock:{resource_id}:{job_type}` | 2 hours | Distributed lock preventing duplicate processing |
| `app:sync:heartbeat:{job_id}` | 20 minutes | Liveness signal from active worker |
| `app:sync:cancel:{job_id}` | 2 hours | Cooperative cancellation flag |
| `app:sync:progress:{job_id}` | 2 hours | HSET with total/completed/message for fast UI polling |

---

## Appendix: Production Validation

This blueprint was extracted from a system that processed:
- 7 distinct job types orchestrated in 2 dependency-ordered waves
- Distributed locking across multi-node deployments
- Crash recovery handling Redis flush and worker death scenarios
- Cooperative cancellation with parent-to-child propagation
- Dual-write progress tracking (2s Redis + 30s DB)
- 14 production bug fixes, each encoding a specific lesson in the processJob sequence

The pitfall catalog (Section 14) is not theoretical — every entry represents a bug found and fixed in production.

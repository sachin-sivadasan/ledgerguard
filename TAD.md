# TAD – LedgerGuard

## 1. Architecture Style

### Backend
- **Language:** Go 1.22+
- **Pattern:** Domain-Driven Design (DDD)
- **Structure:** Modular Monolith

```
backend/
├── cmd/
│   └── server/main.go              # Entry point
├── internal/
│   ├── domain/                     # Core business logic (no external dependencies)
│   │   ├── entity/                 # Domain entities (User, Subscription, Transaction)
│   │   ├── valueobject/            # Value objects (Money, RiskState, ChargeType)
│   │   ├── service/                # Domain services (RiskEngine, MetricsEngine)
│   │   └── repository/             # Repository interfaces (ports)
│   ├── application/                # Application layer (orchestration)
│   │   ├── service/                # Application services (use cases)
│   │   └── dto/                    # Data transfer objects
│   ├── infrastructure/             # External concerns (adapters)
│   │   ├── config/                 # Environment configuration
│   │   ├── persistence/            # Repository implementations (PostgreSQL)
│   │   └── external/               # External service clients (Firebase, Shopify, OpenAI)
│   └── interfaces/                 # Entry points
│       └── http/
│           ├── handler/            # HTTP handlers
│           ├── middleware/         # Auth, CORS, logging
│           └── router/             # Route definitions
├── pkg/                            # Shared utilities
└── migrations/                     # SQL migrations
```

### DDD Layers

| Layer | Purpose | Dependencies |
|-------|---------|--------------|
| **Domain** | Core business logic, entities, rules | None (pure Go) |
| **Application** | Use cases, orchestration, DTOs | Domain |
| **Infrastructure** | DB, external APIs, config | Domain, Application |
| **Interfaces** | HTTP handlers, CLI, gRPC | Application |

**Dependency Rule:** Outer layers depend on inner layers. Domain has zero external dependencies.

### Frontend
- **Framework:** Flutter 3.x
- **Platforms:** Web + iOS + Android (unified codebase)
- **State:** Flutter Bloc
- **Pattern:** Clean Architecture

```
frontend/
├── lib/
│   ├── domain/                  # Entities, repositories interfaces
│   ├── data/                    # API client, repository implementations
│   ├── presentation/
│   │   ├── blocs/              # State management (Bloc pattern)
│   │   ├── pages/              # Screen widgets
│   │   └── widgets/            # Reusable components
│   └── core/
│       ├── pages/               # Screens
│       └── widgets/             # Reusable components
└── test/
```

### Marketing Site
- **Framework:** Next.js 14
- **Styling:** TailwindCSS
- **Hosting:** Vercel
- **Repo:** Separate (`ledgerguard-web`)

### Authentication
- **Provider:** Firebase Authentication
- **Flow:** Frontend gets Firebase ID token → Backend verifies via Firebase Admin SDK
- **Session:** Stateless JWT verification per request

---

## 2. Core Services

| Service | Layer | Responsibility |
|---------|-------|----------------|
| `FirebaseAuthAdapter` | Infrastructure/External | Verify ID tokens, extract user claims |
| `PartnerIntegrationService` | Application | OAuth flow, token storage, app selection |
| `PartnerSyncService` | Application | Fetch transactions, coordinate sync |
| `LedgerRebuilder` | Domain/Service | Deterministic ledger rebuild from transactions |
| `RiskEngine` | Domain/Service | Classify subscription risk states |
| `MetricsEngine` | Domain/Service | Compute KPIs from ledger state |
| `SnapshotService` | Application | Store/retrieve daily snapshots |
| `AIInsightService` | Infrastructure/External | Generate daily briefs via OpenAI |
| `NotificationService` | Infrastructure/External | Email, Slack, in-app alerts |

### Dependency Graph (DDD)

```
Interfaces (HTTP Handlers)
    ↓
Application (PartnerIntegrationService, PartnerSyncService, SnapshotService)
    ↓
Domain (Entities, Value Objects, Domain Services, Repository Interfaces)
    ↓
Infrastructure (PostgreSQL, Firebase, Shopify API implementations)
```

---

## 3. Data Flow

### Sync Pipeline

```
┌─────────────────┐
│ Shopify Partner │
│      API        │
└────────┬────────┘
         │ Fetch transactions (paginated)
         ▼
┌─────────────────┐
│   Transaction   │
│     Store       │ (Immutable, append-only)
└────────┬────────┘
         │ Read 12-month window
         ▼
┌─────────────────┐
│ LedgerRebuilder │ (Deterministic)
│                 │
│ - Classify type │
│ - Link to subs  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   RiskEngine    │
│                 │
│ - SAFE          │
│ - ONE_CYCLE     │
│ - TWO_CYCLE     │
│ - CHURNED       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  MetricsEngine  │
│                 │
│ - MRR           │
│ - At Risk       │
│ - Renewal Rate  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ SnapshotService │ (Daily, immutable)
│                 │
│ - Backfill 365d │
│ - UpsertBatch   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Trend API     │
│                 │
│ - Downsample    │
│ - DAILY/WEEKLY/ │
│   MONTHLY       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ AIInsightService│ (Pro tier)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Notifications   │
│                 │
│ - Email         │
│ - Slack         │
│ - In-app        │
└─────────────────┘
```

### Request Flow (Dashboard)

```
Client → HTTP Handler → Usecase → Repository → PostgreSQL
                                      ↓
Client ← JSON Response ← Usecase ← Domain Models
```

---

## 4. Sync Strategy

### Dual Sync Modes

**Queue-Based Async (Primary — Redis + Worker Pools)**
- Enabled via `cfg.Queue.Enabled=true`
- Job metadata in PostgreSQL (permanent audit trail)
- Live progress in Redis hashes (ephemeral, high-frequency)
- Worker pools: configurable regular (default 3) + full_sync (default 1)
- Auto-trigger on app selection (`POST /api/v1/apps/select` → priority=1 full_sync)
- Returns HTTP 202 immediately with job ID
- `full_sync` creates parent + 6 child jobs in two waves:
  - Wave 1 (independent): transaction_sync, event_sync, review_sync
  - Wave 2 (dependent on transaction_sync): snapshot_sync, status_sync, store_sync
- SyncScheduler disabled when queue enabled

**Synchronous (Legacy — still active)**
- `POST /api/v1/sync` and `/api/v1/sync/{appID}` unchanged
- Returns when sync completes (blocking)
- No job tracking, single-threaded per app
- SyncScheduler runs every 12h when queue disabled

### Schedule
| Event | Timing |
|-------|--------|
| Queue Sync | On-demand (API, frontend, auto-trigger on app selection) |
| Scheduled Sync (Legacy) | Every 12 hours (if queue disabled) |
| Auto-Trigger | On app selection during onboarding |

### Behavior
- **Window:** 12-month rolling
- **Method:** Full recalculation (not incremental)
- **Idempotency:** Same transactions → Same ledger state
- **Concurrency:** Distributed lock per appID+jobType (Redis SETNX)

### Job State Machine
```
pending → processing → completed
                    → failed
                    → cancelled (cooperative)
                    → partial_failure (full_sync with child failures)
```

### Recovery & Reliability
- Workers renew Redis heartbeat every 10s (lock with TTL)
- Startup recovery: re-enqueue jobs stuck in `processing` with stale heartbeat
- Periodic recovery: every 10 minutes, check for orphaned jobs
- Duplicate detection: reject enqueue if active job exists for same appID+type

---

## 5. Security

### Authentication
| Layer | Mechanism |
|-------|-----------|
| Frontend → Backend | Firebase ID Token (Bearer) |
| Backend Verification | Firebase Admin SDK |
| Token Refresh | Handled by Firebase SDK |

### Authorization
```go
// Middleware chain
router.Use(
    middleware.FirebaseAuth(firebaseApp),
    middleware.WorkspaceAccess(workspaceRepo),
    middleware.RoleRequired(role.ADMIN), // per-route
)
```

### Role Permissions
| Action | OWNER | ADMIN | VIEWER |
|--------|-------|-------|--------|
| View dashboard | ✓ | ✓ | ✓ |
| Trigger sync | ✓ | ✓ | ✗ |
| Manage integrations | ✓ | ✓ | ✗ |
| Add manual token | ✓ | ✓ | ✗ |
| Invite members | ✓ | ✓ | ✗ |
| Delete workspace | ✓ | ✗ | ✗ |
| Billing | ✓ | ✗ | ✗ |

### Token Encryption
```go
// Encrypt before storage
encrypted := crypto.EncryptAES256GCM(partnerToken, masterKey)
db.Store(workspaceID, encrypted)

// Decrypt on use
token := crypto.DecryptAES256GCM(encrypted, masterKey)
```

- **Algorithm:** AES-256-GCM
- **Master Key:** Environment variable (rotatable)
- **At Rest:** Encrypted in PostgreSQL

### Input Validation
- All inputs validated at HTTP layer
- SQL injection prevented via parameterized queries
- XSS prevented via JSON-only API

---

## 6. Deployment Environments

### Production — Hetzner Cloud
- Single VPS (CX31: 4 vCPU, 8GB RAM)
- Self-hosted PostgreSQL 16 (local)
- Caddy reverse proxy (auto-SSL)
- systemd services, ~$15/month
- Deploy: `main` branch → GitHub Actions → SSH

### Staging — GCP Cloud Run (Backend)
- Serverless containers (scale 0-2)
- Cloud SQL PostgreSQL 14 (db-f1-micro, private IP)
- Artifact Registry for Docker images
- Secret Manager for credentials
- VPC Connector for private networking
- Deploy: `staging` branch → GitHub Actions → Docker build → Cloud Run
- Cost: $0 (free credits)

### Frontend — Firebase Hosting
- Static Flutter web builds served via Firebase CDN
- Entry points: `main_dev.dart` (localhost), `main_staging.dart` (Cloud Run), `main_prod.dart` (Hetzner)
- SPA routing with rewrites to `index.html`
- Deploy: GitHub Actions → `flutter build web` → Firebase Hosting deploy
- Project: `ledgerguard-c7557`
- Cost: $0 (free tier: 10GB bandwidth/month)
- CORS: Backend allowlists `ledgerguard-c7557.web.app` and `ledgerguard-c7557.firebaseapp.com`

---

## 7. Scalability Strategy

### Current (MVP)
- Single instance
- PostgreSQL for all data
- Full rebuild each sync

### Indexed Tables
```sql
-- Transactions (high volume)
CREATE INDEX idx_transactions_workspace_date
ON transactions(workspace_id, transaction_date DESC);

CREATE INDEX idx_transactions_shop
ON transactions(workspace_id, shop_domain);

-- Subscriptions
CREATE INDEX idx_subscriptions_workspace_status
ON subscriptions(workspace_id, status);

CREATE INDEX idx_subscriptions_risk
ON subscriptions(workspace_id, risk_state);

-- Snapshots
CREATE INDEX idx_snapshots_workspace_date
ON daily_metrics_snapshot(workspace_id, snapshot_date DESC);
```

### Storage Retention
| Data | Retention |
|------|-----------|
| Transactions | 12 months (rolling) |
| Subscriptions | Current state |
| Snapshots | Permanent |
| Audit logs | 2 years |

### Future: Hybrid Incremental Mode
```
Phase 1 (MVP): Full rebuild every sync
Phase 2: Incremental for recent transactions + periodic full rebuild
Phase 3: Event-driven updates via webhooks
```

### Horizontal Scaling (Future)
- Stateless backend → multiple instances
- PostgreSQL read replicas
- Redis for caching/sessions
- Background jobs via task queue

---

## 8. Database Schema

```sql
-- Core tables
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    owner_id VARCHAR(128) NOT NULL, -- Firebase UID
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE workspace_members (
    workspace_id UUID REFERENCES workspaces(id),
    user_id VARCHAR(128) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('OWNER', 'ADMIN', 'VIEWER')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    provider VARCHAR(50) NOT NULL, -- 'shopify_partner'
    org_id VARCHAR(100),
    encrypted_token BYTEA NOT NULL,
    app_ids TEXT[], -- Selected apps to track
    last_sync_at TIMESTAMPTZ,
    sync_status VARCHAR(20) DEFAULT 'IDLE',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    shopify_gid VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL, -- RECURRING, USAGE, ONE_TIME, REFUND
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    app_name VARCHAR(255),
    shop_domain VARCHAR(255),
    transaction_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    shopify_gid VARCHAR(255) UNIQUE NOT NULL,
    shop_domain VARCHAR(255) NOT NULL,
    plan_name VARCHAR(255),
    price_cents BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) NOT NULL, -- ACTIVE, CANCELLED, FROZEN, PENDING
    risk_state VARCHAR(30) NOT NULL, -- SAFE, ONE_CYCLE_MISSED, TWO_CYCLE_MISSED, CHURNED
    current_period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE daily_metrics_snapshot (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    snapshot_date DATE NOT NULL,
    renewal_success_rate DECIMAL(5,2),
    active_mrr_cents BIGINT,
    at_risk_mrr_cents BIGINT,
    usage_revenue_cents BIGINT,
    total_revenue_cents BIGINT,
    safe_count INT,
    one_cycle_missed_count INT,
    two_cycle_missed_count INT,
    churned_count INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, snapshot_date)
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL, -- email, slack, in_app
    payload JSONB,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 9. External Integrations

### Shopify Partner API
- **Endpoint:** `https://partners.shopify.com/{org_id}/api/2025-07/graphql.json`
- **Auth:** `X-Shopify-Access-Token` header
- **Rate Limit:** 4 requests/second (handle with backoff)

### Firebase Auth
- **Admin SDK:** Token verification
- **Client SDK:** Login, token refresh

### OpenAI (Pro)
- **Model:** GPT-4o-mini (insights), GPT-4o (chat)
- **Purpose:** Daily insight generation + AI Chat function calling
- **Input:** Structured JSON snapshot (insights) / natural language + tools (chat)
- **Output:** 80-120 word summary (insights) / tool calls + response (chat)

### Email (Notifications)
- **Provider:** SendGrid / AWS SES
- **Templates:** Transactional (alerts, summaries)

### Slack (Pro)
- **Webhook:** Incoming webhook URL per workspace
- **Format:** Block Kit messages

---

## 10. AI Chat + Internal GraphQL

### Architecture

```
Flutter Chat UI ──WebSocket──► Chat Handler
                                    │
                                    ├── Module Registry (6 modules, 16 tools)
                                    │       │
                                    │       └── Tool Execution → GraphQL Executor
                                    │                                │
                                    ├── AIClient Interface ◄─────────┘
                                    │       │
                                    │       ├── OpenAIClient (gpt-4o, function calling)
                                    │       └── ClaudeClient (future)
                                    │
                                    └── GraphQL Resolvers → Domain Services → PostgreSQL
```

### Components

| Component | Purpose |
|-----------|---------|
| `AIClient` interface | Provider-agnostic AI abstraction (OpenAI today, Claude later) |
| `Module` interface | Self-contained plugin: `Name()`, `Tools()`, `ExecuteTool()` |
| `Registry` | Registers modules, routes tool calls, builds system prompts |
| `GraphQL Executor` | Thread-safe gqlgen wrapper for tool execution |
| `Chat Handler` | WebSocket endpoint, tool call loop (max 5), state extraction |

### Modules

| Module | Tools | Description |
|--------|-------|-------------|
| `risk` | 3 | Risk summary, at-risk list, risk timeline |
| `subscriptions` | 4 | List, detail, search, summary |
| `metrics` | 3 | Current, trend, compare |
| `store_health` | 2 | Health check, overview |
| `earnings` | 2 | Summary, timeline |
| `sync` | 2 | Trigger sync, sync status |

### WebSocket Protocol

```json
// Client → Server
{ "type": "message", "content": "What's my MRR?", "context": [] }

// Server → Client (streaming)
{ "type": "thinking", "content": "Fetching metrics..." }
{ "type": "tool_call", "tool": "metrics__get_current", "args": {...} }
{ "type": "message", "content": "Your MRR is $4,250...", "state": {...}, "suggestions": [...] }
```

---

## 11. Error Handling

### API Errors
```json
{
  "error": {
    "code": "SYNC_IN_PROGRESS",
    "message": "A sync is already running for this workspace",
    "retry_after": 300
  }
}
```

### Error Codes
| Code | HTTP | Meaning |
|------|------|---------|
| `UNAUTHORIZED` | 401 | Invalid/expired token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid input |
| `SYNC_IN_PROGRESS` | 409 | Sync already running |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

### Retry Strategy
```go
// Exponential backoff for external APIs
retrier := retry.New(
    retry.MaxAttempts(3),
    retry.InitialDelay(1*time.Second),
    retry.MaxDelay(30*time.Second),
    retry.Multiplier(2),
)
```

---

## 12. Observability

### Logging
- **Format:** Structured JSON
- **Levels:** DEBUG, INFO, WARN, ERROR
- **Fields:** request_id, workspace_id, user_id, duration

### Metrics
- Request latency (p50, p95, p99)
- Sync duration
- Error rates by type
- Active workspaces

### Alerting
- Sync failures > 3 consecutive
- Error rate > 5%
- API latency p95 > 2s

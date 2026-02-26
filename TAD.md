# TAD – LedgerGuard

## 1. Architecture Style

### Backend
- **Language:** Go 1.22+
- **Pattern:** Clean Architecture
- **Structure:** Modular Monolith

```
backend/
├── cmd/
│   └── server/main.go           # Entry point
├── internal/
│   ├── domain/                  # Entities, business rules (no dependencies)
│   ├── usecase/                 # Application logic
│   ├── repository/              # Data access interfaces
│   │   ├── postgres/            # PostgreSQL implementation
│   │   └── memory/              # In-memory (testing)
│   ├── delivery/
│   │   └── http/                # HTTP handlers
│   ├── service/                 # External service adapters
│   │   ├── firebase/            # Firebase Auth
│   │   ├── shopify/             # Partner API client
│   │   └── openai/              # AI insights
│   └── infrastructure/
│       ├── config/              # Environment config
│       ├── middleware/          # Auth, CORS, logging
│       └── scheduler/           # Cron jobs
├── pkg/                         # Shared utilities
└── migrations/                  # SQL migrations
```

### Frontend
- **Framework:** Flutter 3.x
- **Platforms:** Web + iOS + Android (unified codebase)
- **State:** Riverpod
- **Pattern:** Clean Architecture

```
frontend/
├── lib/
│   ├── domain/                  # Entities, enums
│   ├── data/                    # API client, repositories
│   ├── application/             # Providers, state
│   └── presentation/
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
| `FirebaseAuthAdapter` | Service | Verify ID tokens, extract user claims |
| `PartnerIntegrationService` | Usecase | OAuth flow, token storage, app selection |
| `PartnerSyncService` | Usecase | Fetch transactions, coordinate sync |
| `LedgerRebuilder` | Domain | Deterministic ledger rebuild from transactions |
| `RiskEngine` | Domain | Classify subscription risk states |
| `MetricsEngine` | Domain | Compute KPIs from ledger state |
| `SnapshotService` | Usecase | Store/retrieve daily snapshots |
| `AIInsightService` | Service | Generate daily briefs via OpenAI |
| `NotificationService` | Service | Email, Slack, in-app alerts |

### Dependency Graph

```
HTTP Handlers
    ↓
Usecases (PartnerIntegrationService, PartnerSyncService, SnapshotService)
    ↓
Domain (LedgerRebuilder, RiskEngine, MetricsEngine)
    ↓
Repository Interfaces
    ↓
Implementations (PostgreSQL, Firebase, Shopify API)
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

### Schedule
| Event | Timing |
|-------|--------|
| Scheduled Sync | Every 12 hours (00:00, 12:00 UTC) |
| On-Demand Sync | User-triggered via UI/API |
| Initial Backfill | On first connection (background) |

### Behavior
- **Window:** 12-month rolling
- **Method:** Full recalculation (not incremental)
- **Idempotency:** Same transactions → Same ledger state
- **Concurrency:** One sync per workspace at a time (mutex)

### Sync States
```
IDLE → SYNCING → PROCESSING → COMPLETE
                     ↓
                  FAILED (retry with backoff)
```

### Backfill Strategy
```go
// First sync: fetch all historical data
if workspace.LastSyncAt.IsZero() {
    fetchAllTransactions(ctx, workspace, 12*months)
} else {
    fetchTransactionsSince(ctx, workspace, lastSyncAt)
}
// Always rebuild full 12-month ledger
rebuildLedger(ctx, workspace, 12*months)
```

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

## 6. Scalability Strategy

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

## 7. Database Schema

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

## 8. External Integrations

### Shopify Partner API
- **Endpoint:** `https://partners.shopify.com/{org_id}/api/2025-07/graphql.json`
- **Auth:** `X-Shopify-Access-Token` header
- **Rate Limit:** 4 requests/second (handle with backoff)

### Firebase Auth
- **Admin SDK:** Token verification
- **Client SDK:** Login, token refresh

### OpenAI (Pro)
- **Model:** GPT-4o-mini
- **Purpose:** Daily insight generation
- **Input:** Structured JSON snapshot
- **Output:** 80-120 word summary

### Email (Notifications)
- **Provider:** SendGrid / AWS SES
- **Templates:** Transactional (alerts, summaries)

### Slack (Pro)
- **Webhook:** Incoming webhook URL per workspace
- **Format:** Block Kit messages

---

## 9. Error Handling

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

## 10. Observability

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

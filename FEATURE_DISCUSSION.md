# Feature Discussion – LedgerGuard

Track feature discussions and planning before implementation.

---

## Current Status

### Implemented
- Authentication (Firebase Auth, Google Sign-In)
- Partner Integration (OAuth, Manual Token)
- App Selection
- Executive Dashboard (KPIs, Time Filtering, Delta Comparison)
- Risk Breakdown Page
- Subscription List & Detail
- AI Insight Card (Pro tier)
- Notification Settings
- Profile Page
- Mobile Responsive Layouts

### Pending from Plan
- Dashboard "View All" button (risk breakdown → subscription list)

---

## Feature Discussion Log

### Session: 2026-02-27

**Topic:** Revenue API (External Access Layer)

**Overview:**
New module allowing Shopify app developers to query subscription payment status and usage billing status via API.

**Phased Approach:**
- **Phase 1:** REST API (MVP) - Implement fully
- **Phase 2:** GraphQL (future-ready) - Structure only, no implementation

---

## Revenue API Specification

### Architecture Rules

1. Follow Clean Architecture (Go backend)
2. Do NOT mix core ledger DB with external API read logic
3. Implement CQRS-style read model
4. REST now, GraphQL-ready later
5. Update all documentation (PRD, TAD, DATABASE_SCHEMA, DECISIONS, TEST_PLAN, future.md, PlantUML diagrams)

### Phase 1 Implementation Steps

#### Step 1: Database Extension

**New Tables:**

```sql
-- API Keys for external access
api_keys (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    key_hash VARCHAR(64) NOT NULL,  -- SHA-256 hash
    name VARCHAR(100),              -- User-friendly name
    created_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    rate_limit_per_minute INT DEFAULT 60
)

-- Read model: Subscription status (CQRS)
api_subscription_status (
    subscription_id UUID PRIMARY KEY,
    app_id UUID NOT NULL,
    myshopify_domain VARCHAR(255) NOT NULL,
    risk_state VARCHAR(50) NOT NULL,
    is_paid_current_cycle BOOLEAN NOT NULL,
    expected_next_charge_date TIMESTAMP,
    last_synced_at TIMESTAMP NOT NULL
)

-- Read model: Usage billing status (CQRS)
api_usage_status (
    app_usage_record_id UUID PRIMARY KEY,
    subscription_id UUID NOT NULL,
    billed BOOLEAN NOT NULL,
    billing_date TIMESTAMP,
    amount_cents INT NOT NULL,
    last_synced_at TIMESTAMP NOT NULL
)

-- API Audit Log (separate from core ledger)
api_audit_log (
    id UUID PRIMARY KEY,
    api_key_id UUID NOT NULL REFERENCES api_keys(id),
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    request_params JSONB,
    response_status INT NOT NULL,
    response_time_ms INT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
)
```

#### Step 2: Read Model Population

**Service:** `RevenueReadModelBuilder`

- Triggered after every ledger rebuild
- Populates `api_subscription_status` and `api_usage_status`
- Uses existing ledger state (NO Shopify re-sync)
- Idempotent operation

#### Step 3: API Key System

**Endpoints:**
```
POST   /api/v1/api-keys         (OWNER only) - Create key
GET    /api/v1/api-keys         (OWNER only) - List keys
DELETE /api/v1/api-keys/{id}    (OWNER only) - Revoke key
```

**Security:**
- Store only hashed keys (SHA-256)
- Return raw key ONLY at creation (one-time)
- Implement `X-API-KEY` header validation middleware
- Attach user context from key
- Enforce per-key rate limiting

#### Step 4: REST Endpoints

**Subscription Status:**
```
GET /v1/subscription/{subscription_id}/status

Response:
{
    "subscription_id": "uuid",
    "risk_state": "SAFE|ONE_CYCLE_MISSED|TWO_CYCLES_MISSED|CHURNED",
    "is_paid_current_cycle": true,
    "expected_next_charge_date": "2024-03-15T00:00:00Z"
}
```

**Usage Status:**
```
GET /v1/usage/{usage_id}/status

Response:
{
    "usage_id": "uuid",
    "billed": true,
    "billing_date": "2024-02-28T00:00:00Z",
    "amount_cents": 1500
}
```

**Batch Subscription Status:**
```
POST /v1/subscriptions/status/batch

Request:
{
    "subscription_ids": ["uuid1", "uuid2", "uuid3"]
}

Response:
{
    "results": [
        {
            "subscription_id": "uuid1",
            "risk_state": "SAFE",
            "is_paid_current_cycle": true,
            "expected_next_charge_date": "2024-03-15T00:00:00Z"
        },
        ...
    ],
    "not_found": ["uuid3"]
}
```

**Batch Usage Status:**
```
POST /v1/usage/status/batch

Request:
{
    "usage_ids": ["uuid1", "uuid2"]
}

Response:
{
    "results": [
        {
            "usage_id": "uuid1",
            "billed": true,
            "billing_date": "2024-02-28T00:00:00Z",
            "amount_cents": 1500
        },
        ...
    ],
    "not_found": []
}
```

**Rules:**
- Tenant isolation (user can only access their own app data)
- Return 404 if not found or unauthorized (single queries)
- Batch queries return `not_found` array for missing IDs
- Batch limit: 100 IDs per request
- Rate limiting enforced

#### Step 5: Security

- Per-key rate limiting middleware
- Audit log for API usage
- Mask sensitive internal fields
- Integration tests for all scenarios

#### Step 6: Documentation Updates

| Document | Updates |
|----------|---------|
| PRD.md | Add Revenue API module (Pro/Enterprise tier) |
| TAD.md | Add Revenue API Layer + CQRS read model |
| DATABASE_SCHEMA.md | Add new tables |
| DECISIONS.md | ADR for CQRS pattern, API key security |
| TEST_PLAN.md | Add Revenue API test scenarios |
| future.md | GraphQL Phase 2 plan |
| C4.puml | Add Revenue API component |
| ER.puml | Add api_keys, api_subscription_status, api_usage_status |
| SEQUENCE.puml | Add Developer → Revenue API → Read Model flow |

### Phase 2: GraphQL (Future - DO NOT IMPLEMENT)

**Add to future.md:**
- GraphQL endpoint `/graphql`
- Uses read-model tables only
- Schema:
  - `subscription(id)` - Single subscription status
  - `usage(id)` - Single usage status
  - `subscriptions(filter)` - Filtered list
  - `usageRecords(filter)` - Filtered list

**Prepare structure:**
- Create placeholder: `internal/revenue_api/graphql/` (empty)
- REST handlers call application layer services
- GraphQL resolvers will reuse same services later

### Test Requirements

| Test Category | Scenarios |
|---------------|-----------|
| API Key | Creation, validation, revocation |
| Rate Limiting | Enforcement, per-key limits |
| Subscription Status | Success, not found, unauthorized |
| Usage Status | Success, not found, unauthorized |
| Tenant Isolation | Cross-user access denied |
| Read Model | Population after ledger rebuild |

### Directory Structure (Proposed)

```
internal/
├── revenue_api/
│   ├── application/
│   │   └── service/
│   │       ├── subscription_status_service.go
│   │       ├── usage_status_service.go
│   │       └── api_key_service.go
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── api_key.go
│   │   │   ├── subscription_status.go
│   │   │   └── usage_status.go
│   │   └── repository/
│   │       ├── api_key_repository.go
│   │       ├── subscription_status_repository.go
│   │       └── usage_status_repository.go
│   ├── infrastructure/
│   │   └── persistence/
│   │       ├── api_key_repository.go
│   │       ├── subscription_status_repository.go
│   │       └── usage_status_repository.go
│   ├── interfaces/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── api_key_handler.go
│   │       │   ├── subscription_status_handler.go
│   │       │   └── usage_status_handler.go
│   │       └── middleware/
│   │           ├── api_key_auth.go
│   │           └── rate_limiter.go
│   └── graphql/              # Phase 2 placeholder (empty)
```

---

## Questions Before Implementation

1. **Rate Limiting Strategy:** In-memory (Redis) or database-backed? → **Redis**
2. **Audit Log Table:** Separate table or append to existing logs? → **Separate table** (better for CQRS separation, easier analytics)
3. **API Versioning:** Path-based (`/v1/`) confirmed? → **Yes**
4. **Key Rotation:** Should we support key rotation or just revoke+create? → **Revoke + create new**
5. **Batch Queries:** Allow batch subscription/usage lookups in Phase 1? → **Yes, allow**

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-27 | CQRS read model | Separate read concerns from write ledger |
| 2026-02-27 | Phased approach (REST → GraphQL) | MVP fast, future-ready |
| 2026-02-27 | Hashed API keys only | Security best practice |
| 2026-02-27 | Redis for rate limiting | Distributed, fast, supports TTL natively |
| 2026-02-27 | Separate api_audit_log table | CQRS separation, easier API analytics |
| 2026-02-27 | No key rotation | Simpler MVP, revoke+create sufficient |
| 2026-02-27 | Batch queries in Phase 1 | Developer convenience, reduce API calls |

---

## Implementation Checklist

### Phase 1: REST Revenue API

- [ ] **Step 1:** Database migration (api_keys, api_subscription_status, api_usage_status)
- [ ] **Step 2:** RevenueReadModelBuilder service
- [ ] **Step 3:** API key management (create, list, revoke)
- [ ] **Step 4:** API key auth middleware + rate limiter
- [ ] **Step 5:** Subscription status endpoint
- [ ] **Step 6:** Usage status endpoint
- [ ] **Step 7:** Integration tests
- [ ] **Step 8:** Documentation updates
- [ ] **Step 9:** Phase 2 placeholder structure

---

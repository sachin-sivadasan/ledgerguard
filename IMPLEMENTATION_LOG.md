# Implementation Log – LedgerGuard

A chronological record of all features implemented with detailed summaries.

---

## [2026-02-26] Initial Setup

**Commit:** Initial commit

**Summary:**
- Initialized git repository
- Created project documentation structure
- Added remote origin

**Files Created:**
- PRD.md - Product Requirements Document
- TAD.md - Technical Architecture Document
- DATABASE_SCHEMA.md - Database schema documentation
- CLAUDE.md - Development directives
- .gitignore

---

## [2026-02-26] Initialize Go Backend

**Commit:** Initialize Go backend with DDD structure

**Summary:**
Initialized Go 1.22+ backend project with Clean Architecture folder structure, PostgreSQL connection, migration setup, and basic health endpoint.

**Implemented:**

1. **Project Structure:**
   - `cmd/server/main.go` - Entry point
   - `internal/domain/` - Domain layer (entities, value objects, repository interfaces)
   - `internal/application/` - Application layer (services)
   - `internal/infrastructure/` - Infrastructure layer (config, persistence, external)
   - `internal/interfaces/` - Interfaces layer (HTTP handlers, middleware, router)
   - `pkg/` - Shared utilities

2. **Dependencies:**
   - Chi router for HTTP
   - pgx v5 for PostgreSQL
   - golang-migrate for migrations

3. **Configuration:**
   - Environment variable loading
   - YAML config file support
   - Config precedence: defaults → file → env vars

4. **Health Endpoint:**
   - `GET /health` - Returns server and database status
   - 3 tests (healthy, unhealthy DB, no DB)

**Tests:** 3 passing

---

## [2026-02-26] Firebase Auth Middleware

**Commit:** Implement Firebase authentication middleware

**Summary:**
Implemented Firebase ID token verification middleware with auto-user creation on first login.

**Implemented:**

1. **Domain Layer:**
   - `User` entity (ID, FirebaseUID, Email, Role, PlanTier, CreatedAt)
   - `Role` value object (OWNER, ADMIN)
   - `PlanTier` value object (FREE, PRO)
   - `UserRepository` interface
   - `AuthTokenVerifier` interface

2. **Infrastructure Layer:**
   - `PostgresUserRepository` implementation
   - `FirebaseAuthService` for token verification
   - Migration `000001_create_users_table`

3. **Interfaces Layer:**
   - `AuthMiddleware` - Verifies Firebase tokens, loads/creates users
   - Sets user in request context for downstream handlers

4. **Behavior:**
   - Validates `Authorization: Bearer <token>` header
   - Verifies token with Firebase Admin SDK
   - Auto-creates user with OWNER role on first login
   - Returns 401 for invalid/missing tokens

**Tests:** 6 passing (total: 9)

---

## [2026-02-26] Config File Support

**Commit:** Add YAML config file support

**Summary:**
Added YAML configuration file support with environment variable overrides.

**Implemented:**

1. **Config Loading:**
   - YAML file support (`config.yaml`)
   - `-config` CLI flag or `CONFIG_PATH` env var
   - Load order: defaults → config file → env vars (env wins)

2. **Config Sections:**
   - Server (port)
   - Database (host, port, user, password, name, sslmode)
   - Firebase (credentials path)
   - Shopify (client ID, secret, redirect URI)
   - Encryption (master key)

3. **Files:**
   - `config.example.yaml` - Template with all options
   - Updated `.gitignore` for local configs

**Tests:** 5 passing (total: 14)

---

## [2026-02-26] Role-Based Access Middleware

**Commit:** Implement role-based access control middleware

**Summary:**
Implemented RBAC middleware for protecting routes by user role.

**Implemented:**

1. **Middleware:**
   - `RequireRoles(roles ...Role)` - Restricts access to specified roles
   - OWNER has access to all routes (superset of ADMIN)
   - ADMIN only has access to ADMIN-allowed routes

2. **Behavior:**
   - Returns 401 if no user in context
   - Returns 403 if user lacks required role
   - Passes through if user has any allowed role

**Tests:** 5 passing (total: 19)

---

## [2026-02-26] Shopify Partner OAuth Flow

**Commit:** Implement Shopify Partner OAuth flow

**Summary:**
Implemented complete OAuth 2.0 flow for Shopify Partner API integration with encrypted token storage.

**Implemented:**

1. **Domain Layer:**
   - `IntegrationType` value object (OAUTH, MANUAL)
   - `PartnerAccount` entity
   - `PartnerAccountRepository` interface

2. **Infrastructure Layer:**
   - `PostgresPartnerAccountRepository` implementation
   - `ShopifyOAuthService` for token exchange
   - Migration `000002_create_partner_accounts_table`

3. **Security:**
   - `pkg/crypto/aes.go` - AES-256-GCM encryption
   - Tokens encrypted before storage
   - Random IV for non-deterministic ciphertext

4. **Endpoints:**
   - `GET /api/v1/integrations/shopify/oauth` - Start OAuth flow
   - `GET /api/v1/integrations/shopify/callback` - Handle callback

**Tests:** 12 passing (total: 31)

---

## [2026-02-26] Manual Partner Token Integration

**Commit:** Implement manual partner token integration (ADMIN only)

**Summary:**
Implemented manual token entry for development/testing, restricted to ADMIN users.

**Implemented:**

1. **Endpoints (ADMIN only):**
   - `POST /api/v1/integrations/shopify/token` - Add manual token
   - `GET /api/v1/integrations/shopify/token` - Get token info (masked)
   - `DELETE /api/v1/integrations/shopify/token` - Revoke token

2. **Security:**
   - AES-256-GCM encryption for token storage
   - Token masking in responses (`***...xxxx`)
   - RequireRoles(ADMIN) middleware

3. **Repository:**
   - Added `Delete` method to `PartnerAccountRepository`

**Tests:** 12 passing (total: 36)

---

## [2026-02-26] App Fetching and Selection

**Commit:** Implement app fetching and selection from Partner API

**Summary:**
Implemented fetching apps from Shopify Partner GraphQL API and storing selected apps for tracking.

**Implemented:**

1. **Domain Layer:**
   - `App` entity (ID, PartnerAccountID, PartnerAppID, Name, TrackingEnabled)
   - `AppRepository` interface

2. **Infrastructure Layer:**
   - `PostgresAppRepository` implementation
   - `ShopifyPartnerClient` for GraphQL API calls
   - Migration `000003_create_apps_table`

3. **Endpoints:**
   - `GET /api/v1/apps/available` - Fetch apps from Partner API
   - `POST /api/v1/apps/select` - Select and store an app
   - `GET /api/v1/apps` - List user's tracked apps

4. **Behavior:**
   - Decrypts partner token to call Shopify API
   - Prevents duplicate app tracking (409 Conflict)
   - Stores app with tracking enabled by default

**Tests:** 14 passing (total: 50)

---

## [2026-02-26] PartnerSyncService

**Commit:** `59ce04c` - Implement PartnerSyncService for transaction synchronization

**Summary:**
Implemented transaction synchronization service with scheduled 12-hour sync and on-demand triggers.

**Implemented:**

1. **Domain Layer:**
   - `ChargeType` value object (RECURRING, USAGE, ONE_TIME, REFUND)
   - `Transaction` entity (immutable ledger record)
   - `TransactionRepository` interface with batch upsert
   - Added `FindByID` to `PartnerAccountRepository`

2. **Infrastructure Layer:**
   - `PostgresTransactionRepository` with batch upsert
   - Idempotent storage via `ON CONFLICT (shopify_gid) DO UPDATE`
   - Migration `000004_create_transactions_table`

3. **Application Layer:**
   - `SyncService` with:
     - `SyncApp(appID)` - Sync single app
     - `SyncAllApps(partnerAccountID)` - Sync all apps
   - `SyncScheduler` - 12-hour interval (00:00, 12:00 UTC)
   - `TransactionFetcher` interface (mock-ready)

4. **Endpoints:**
   - `POST /api/v1/sync` - Sync all user's apps
   - `POST /api/v1/sync/{appID}` - Sync specific app

5. **Indexes:**
   - `idx_transactions_app_date` - For time-range queries
   - `idx_transactions_domain` - For store lookups
   - `idx_transactions_type` - For charge type filtering

**Tests:** 11 passing (total: 58)

---

## [2026-02-26] Deterministic Ledger Rebuild

**Commit:** `a098db8` - Implement deterministic ledger rebuild service

**Summary:**
Implemented deterministic ledger rebuild that reconstructs subscription state from transactions, separates revenue streams, and classifies risk.

**Implemented:**

1. **Value Objects:**
   - `RiskState` (SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED)
   - `BillingInterval` (MONTHLY, ANNUAL) with `NextChargeDate()` calculation

2. **Domain Layer:**
   - `Subscription` entity with:
     - `last_recurring_charge_date` - Most recent RECURRING transaction
     - `expected_next_charge_date` - Computed from last charge + interval
     - `risk_state` - Classified based on days past due
     - `MRRCents()` - Monthly recurring revenue (annual ÷ 12)
   - `SubscriptionRepository` interface
   - `LedgerService` with:
     - `RebuildFromTransactions(appID, now)` - Full rebuild
     - `SeparateRevenue(transactions)` - Split RECURRING/USAGE

3. **Infrastructure Layer:**
   - `PostgresSubscriptionRepository` implementation
   - Migration `000005_create_subscriptions_table`

4. **Risk Classification (per CLAUDE.md):**
   - 0-30 days past due: **SAFE** (grace period)
   - 31-60 days past due: **ONE_CYCLE_MISSED**
   - 61-90 days past due: **TWO_CYCLES_MISSED**
   - >90 days past due: **CHURNED**

5. **Revenue Separation:**
   - **MRR** = Sum of RECURRING charges only
   - **Usage Revenue** = Sum of USAGE charges only
   - ONE_TIME and REFUND tracked separately

6. **Deterministic Guarantee:**
   - Same transactions → Same subscription state
   - Sorted output for consistent ordering
   - Full rebuild (delete + insert) for idempotency

7. **Billing Interval Detection:**
   - Analyzes transaction spacing
   - >180 days average = ANNUAL
   - ≤180 days average = MONTHLY

8. **Indexes:**
   - `idx_subscriptions_app_status` - For status queries
   - `idx_subscriptions_app_risk` - For risk state queries
   - `idx_subscriptions_domain` - For store lookups
   - `idx_subscriptions_expected_charge` - For at-risk detection

**Tests:** 8 passing (total: 66)

---

## [2026-02-26] RiskEngine Integration

**Commit:** Implement RiskEngine with sync integration

**Summary:**
Created dedicated RiskEngine domain service and integrated it with the sync flow to recalculate risk states after each synchronization.

**Implemented:**

1. **Domain Service - RiskEngine:**
   - `ClassifyRisk(subscription, now)` - Determines risk state based on payment history
   - `DaysPastDue(subscription, now)` - Calculates days past expected charge
   - `RiskStateFromDaysPastDue(days)` - Converts days to risk state
   - `ClassifyAll(subscriptions, now)` - Batch classification
   - `CalculateRiskSummary(subscriptions)` - Risk distribution counts
   - `CalculateRevenueAtRisk(subscriptions)` - MRR at risk (ONE_CYCLE + TWO_CYCLES)
   - `IsAtRisk(subscription)` - Helper for at-risk detection
   - `IsChurned(subscription)` - Helper for churn detection

2. **Risk State Classification (per CLAUDE.md):**
   - **SAFE:** Active subscription or ≤30 days past due (grace period)
   - **ONE_CYCLE_MISSED:** 31-60 days past due
   - **TWO_CYCLES_MISSED:** 61-90 days past due
   - **CHURNED:** >90 days past due

3. **Sync Integration:**
   - Added `LedgerRebuilder` interface to SyncService
   - SyncService triggers `RebuildFromTransactions` after storing transactions
   - `SyncResult` now includes:
     - `RiskSummary` - Distribution of subscriptions by risk state
     - `RevenueAtRisk` - Total MRR from at-risk subscriptions
     - `TotalMRRCents` - Total MRR from active subscriptions

4. **Tests:**
   - Risk classification tests (all states)
   - Days past due calculation tests
   - Batch classification tests
   - Risk summary tests
   - Revenue at risk tests
   - Updated SyncService tests with mock LedgerRebuilder
   - Updated SyncHandler tests with mock LedgerRebuilder

**Tests:** 22 new tests (total: 88)

---

## [2026-02-26] MetricsEngine Implementation

**Commit:** Implement MetricsEngine with daily snapshots

**Summary:**
Created MetricsEngine domain service for KPI computation and daily metrics snapshots storage.

**Implemented:**

1. **Domain Entity - DailyMetricsSnapshot:**
   - Immutable daily KPI records (one per app per day)
   - Stores: ActiveMRR, RevenueAtRisk, UsageRevenue, TotalRevenue
   - Stores: RenewalSuccessRate, subscription counts by risk state
   - Never deleted - permanent audit trail

2. **Domain Service - MetricsEngine:**
   - `CalculateActiveMRR(subscriptions)` - Sum MRR from SAFE subscriptions
   - `CalculateRevenueAtRisk(subscriptions)` - MRR from ONE_CYCLE + TWO_CYCLES
   - `CalculateUsageRevenue(transactions)` - Sum of USAGE transactions
   - `CalculateTotalRevenue(transactions)` - RECURRING + USAGE + ONE_TIME - REFUNDS
   - `CalculateRenewalSuccessRate(subscriptions)` - SAFE / Total as decimal
   - `ComputeAllMetrics(appID, subscriptions, transactions, now)` - Creates complete snapshot

3. **Repository:**
   - `DailyMetricsSnapshotRepository` interface
   - PostgreSQL implementation with `Upsert` (ON CONFLICT DO UPDATE)
   - `FindByAppIDAndDate`, `FindByAppIDRange`, `FindLatestByAppID`

4. **Integration:**
   - LedgerService now accepts optional snapshot repository
   - `WithSnapshotRepository(repo)` builder method
   - Stores daily snapshot after each ledger rebuild

5. **Migration:**
   - `000006_create_daily_metrics_snapshot_table`
   - UNIQUE constraint on (app_id, date)
   - Indexes for time-series queries

**Tests:** 10 new tests (total: 98)

---

## [2026-02-26] AIInsightService Implementation

**Commit:** Implement AIInsightService with plan tier gating

**Summary:**
Created AIInsightService for generating AI-powered daily executive briefs (Pro tier only).

**Implemented:**

1. **Domain Entity - DailyInsight:**
   - AI-generated daily summary (80-120 words)
   - One insight per app per day
   - Stored for audit trail

2. **Application Service - AIInsightService:**
   - `GenerateInsight(userID, appID, snapshot, now)` - Generate AI brief
   - `BuildPrompt(snapshot)` - Construct LLM prompt from metrics
   - Plan tier gating (returns `ErrProTierRequired` for FREE users)
   - Uses `AIProvider` interface for mockable LLM calls

3. **Interfaces:**
   - `AIProvider` - Interface for LLM API (OpenAI, Claude, etc.)
   - `DailyInsightRepository` - Repository interface

4. **Repository:**
   - `PostgresDailyInsightRepository` implementation
   - Upsert with ON CONFLICT for idempotency

5. **Migration:**
   - `000007_create_daily_insight_table`
   - UNIQUE constraint on (app_id, date)

6. **UserRepository Enhancement:**
   - Added `FindByID(id uuid.UUID)` method
   - Updated PostgreSQL implementation

**Tests:** 5 new tests (total: 103)

---

## Test Summary

| Package | Tests |
|---------|-------|
| infrastructure/config | 5 |
| infrastructure/external | 7 |
| interfaces/http/handler | 42 |
| interfaces/http/middleware | 11 |
| application/service | 10 |
| domain/service | 30 |
| pkg/crypto | 5 |
| **Total** | **103** |

---

## Migration Summary

| Migration | Description | Status |
|-----------|-------------|--------|
| 000001_create_users_table | Users with Firebase UID, role, plan tier | ✓ |
| 000002_create_partner_accounts_table | Partner accounts with encrypted tokens | ✓ |
| 000003_create_apps_table | Tracked Shopify apps | ✓ |
| 000004_create_transactions_table | Immutable transaction ledger | ✓ |
| 000005_create_subscriptions_table | Subscription state with risk tracking | ✓ |
| 000006_create_daily_metrics_snapshot_table | Daily KPI snapshots | ✓ |
| 000007_create_daily_insight_table | AI-generated daily insights (Pro only) | ✓ |

---

## Architecture

```
cmd/server/main.go              → Entry point only
internal/domain/
  ├── entity/                   → User, PartnerAccount, App, Transaction, Subscription, DailyMetricsSnapshot, DailyInsight
  ├── valueobject/              → Role, PlanTier, IntegrationType, ChargeType, RiskState, BillingInterval
  ├── repository/               → Interfaces (UserRepo, PartnerAccountRepo, AppRepo, TransactionRepo, SubscriptionRepo, DailyMetricsSnapshotRepo, DailyInsightRepo)
  └── service/                  → LedgerService, RiskEngine, MetricsEngine
internal/application/
  ├── service/                  → SyncService, AIInsightService
  └── scheduler/                → SyncScheduler
internal/infrastructure/
  ├── config/                   → YAML + env config loading
  ├── persistence/              → PostgreSQL implementations
  └── external/                 → Firebase, Shopify OAuth, Shopify Partner Client
internal/interfaces/http/
  ├── handler/                  → Health, OAuth, ManualToken, App, Sync
  ├── middleware/               → Auth, Role
  └── router/                   → Chi router setup
pkg/crypto/                     → AES-256-GCM encryption
migrations/                     → SQL migrations
```

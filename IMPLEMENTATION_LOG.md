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

## [2026-02-27] NotificationService Implementation

**Commit:** Implement NotificationService with push notifications

**Summary:**
Created NotificationService for push notifications with device token management, critical alerts for risk state changes, and daily summary notifications.

**Implemented:**

1. **Domain Entities:**
   - `DeviceToken` entity (ID, UserID, DeviceToken, Platform)
   - `NotificationPreferences` entity (CriticalEnabled, DailySummaryEnabled, DailySummaryTime, SlackWebhookURL)
   - `Platform` value object (ios, android, web)

2. **Repository Interfaces:**
   - `DeviceTokenRepository` - CRUD for device tokens
   - `NotificationPreferencesRepository` - CRUD with upsert for preferences

3. **Application Service - NotificationService:**
   - `RegisterDevice(userID, deviceToken, platform)` - Register push token
   - `UnregisterDevice(userID, deviceToken)` - Remove push token
   - `SendCriticalAlert(userID, appName, storeDomain, oldState, newState)` - Risk change alert
   - `SendDailySummary(userID, appName, snapshot)` - Daily metrics summary
   - `GetPreferences(userID)` - Get notification settings
   - `UpdatePreferences(prefs)` - Update notification settings

4. **Interfaces:**
   - `PushNotificationProvider` - Interface for FCM/APNs (mockable)

5. **Infrastructure Layer:**
   - `PostgresDeviceTokenRepository` implementation
   - `PostgresNotificationPreferencesRepository` implementation

6. **Migrations:**
   - `000008_create_device_tokens_table` - Device tokens with platform
   - `000009_create_notification_preferences_table` - User notification settings

7. **Behavior:**
   - Respects user preferences before sending notifications
   - Handles device token transfer between users
   - Creates default preferences on first device registration
   - Sends to all registered devices for a user

**Tests:** 15 new tests (total: 109 top-level, 128 with subtests)

---

## [2026-02-27] SlackNotificationProvider Implementation

**Commit:** Implement SlackNotificationProvider for Slack webhooks

**Summary:**
Created SlackNotificationProvider for sending notifications to Slack via webhooks, integrated with NotificationService.

**Implemented:**

1. **Infrastructure - SlackNotificationProvider:**
   - `SendSlack(ctx, webhookURL, title, body, color)` - Send Slack webhook message
   - Slack payload with attachments for rich formatting
   - Color constants (danger, warning, success, info)
   - Configurable HTTP client (for testing)

2. **Application Service Updates:**
   - Added `SlackNotifier` interface to NotificationService
   - Added `WithSlackNotifier(notifier)` builder method
   - `SendCriticalAlert` now sends to Slack (danger color) when webhook configured
   - `SendDailySummary` now sends to Slack (info color) when webhook configured
   - Continues sending push notifications even if Slack fails

3. **Tests:**
   - SlackNotificationProvider tests (5 test cases)
   - Slack integration tests in NotificationService (5 test cases)

**Tests:** 10 new tests (total: 112)

---

## [2026-02-27] Marketing Site Implementation

**Commit:** Create Next.js marketing site for LedgerGuard

**Summary:**
Created a public-facing marketing landing page for LedgerGuard using Next.js 14+ with TailwindCSS.

**Implemented:**

1. **Documentation:**
   - `marketing/REQUIREMENTS.md` - Site requirements, copy, and design specs

2. **Next.js Site (`marketing/site/`):**
   - Next.js 14+ with App Router
   - TailwindCSS for styling
   - Inter font from Google Fonts
   - Responsive, mobile-first design

3. **Components:**
   - `Header.tsx` - Fixed navigation with logo and CTA
   - `Hero.tsx` - Main headline, subheadline, dual CTAs
   - `Problem.tsx` - 3 pain point cards
   - `RenewalRate.tsx` - Key metric explanation with visual
   - `RevenueAtRisk.tsx` - Key metric with breakdown visual
   - `AIBrief.tsx` - Pro feature showcase with example brief
   - `Pricing.tsx` - Free vs Pro tier comparison
   - `FinalCTA.tsx` - Conversion-focused closing section
   - `Footer.tsx` - Simple footer with copyright

4. **SEO:**
   - Meta title and description
   - Semantic HTML structure
   - Smooth scroll behavior

**Files Created:**
- `marketing/REQUIREMENTS.md`
- `marketing/site/` - Full Next.js project
- 9 React components

---

## Test Summary

| Package | Tests |
|---------|-------|
| infrastructure/config | 5 |
| infrastructure/external | 12 |
| interfaces/http/handler | 42 |
| interfaces/http/middleware | 11 |
| application/service | 21 |
| domain/service | 16 |
| pkg/crypto | 5 |
| **Total** | **112** |

---

## Related Logs

- **Frontend (Flutter):** See [`frontend/IMPLEMENTATION_LOG.md`](frontend/IMPLEMENTATION_LOG.md)
- **Marketing Site:** See [`marketing/REQUIREMENTS.md`](marketing/REQUIREMENTS.md)

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
| 000008_create_device_tokens_table | Push notification device tokens | ✓ |
| 000009_create_notification_preferences_table | User notification preferences | ✓ |
| 000010_add_shop_name_to_subscriptions | Add shop_name to subscriptions | ✓ |
| 000011_add_shop_name_and_gross_amount_to_transactions | Add shop_name, gross_amount_cents to transactions | ✓ |

---

## Architecture

```
cmd/server/main.go              → Entry point only
internal/domain/
  ├── entity/                   → User, PartnerAccount, App, Transaction, Subscription, DailyMetricsSnapshot, DailyInsight, DeviceToken, NotificationPreferences
  ├── valueobject/              → Role, PlanTier, IntegrationType, ChargeType, RiskState, BillingInterval, Platform
  ├── repository/               → Interfaces (UserRepo, PartnerAccountRepo, AppRepo, TransactionRepo, SubscriptionRepo, DailyMetricsSnapshotRepo, DailyInsightRepo, DeviceTokenRepo, NotificationPreferencesRepo)
  └── service/                  → LedgerService, RiskEngine, MetricsEngine
internal/application/
  ├── service/                  → SyncService, AIInsightService, NotificationService
  └── scheduler/                → SyncScheduler
internal/infrastructure/
  ├── config/                   → YAML + env config loading
  ├── persistence/              → PostgreSQL implementations
  └── external/                 → Firebase, Shopify OAuth, Shopify Partner Client, SlackNotificationProvider
internal/interfaces/http/
  ├── handler/                  → Health, OAuth, ManualToken, App, Sync
  ├── middleware/               → Auth, Role
  └── router/                   → Chi router setup
pkg/crypto/                     → AES-256-GCM encryption
migrations/                     → SQL migrations
```

---

## [2026-02-27] KPI Dashboard Upgrade: Time Filtering and Delta Comparison

**Commit:** feat: implement KPI dashboard time filtering and delta comparison

**Summary:**
Implemented Play Store-style analytics with time-based filtering and period-over-period delta comparisons for the executive dashboard.

**Implemented:**

1. **Domain Layer:**
   - `TimeRangePreset` value object (THIS_MONTH, LAST_MONTH, LAST_30_DAYS, LAST_90_DAYS, CUSTOM)
   - `DateRange` helper with factory methods for each preset
   - `PeriodMetrics` entity with current, previous, and delta
   - `MetricsSummary` with all KPI fields and period dates
   - `MetricsDelta` with percentage changes and good/bad semantics

2. **Application Layer:**
   - `MetricsAggregationService` with `GetPeriodMetrics(ctx, appID, start, end)`
   - Aggregates daily snapshots into period summaries
   - Point-in-time metrics: End-of-period snapshot (MRR, risk counts)
   - Cumulative metrics: Sum across period (revenue totals)
   - Delta calculation with divide-by-zero protection

3. **Interfaces Layer:**
   - `GetMetricsByPeriod` handler with start/end query params
   - Route: `GET /api/v1/apps/{appID}/metrics`
   - Backward compatible with existing `/metrics/latest` endpoint

4. **Delta Semantics:**
   - Higher is good: Renewal Success Rate, Active MRR, Usage Revenue
   - Lower is good: Revenue at Risk, Churn Count
   - Colors: Green for good change, Red for bad change

5. **API Response Structure:**
   ```json
   {
     "period": { "start": "2024-02-01", "end": "2024-02-27" },
     "current": { ... },
     "previous": { ... },
     "delta": {
       "active_mrr_percent": 5.93,
       "revenue_at_risk_percent": -8.5,
       "renewal_success_rate_percent": 2.1
     }
   }
   ```

**Tests:** 5 new tests in MetricsAggregationService (total backend: 117)

---

## [2026-02-27] Live FetchTransactions from Shopify Partner API

**Commit:** feat: implement live FetchTransactions from Shopify Partner API

**Summary:**
Implemented live transaction fetching from Shopify Partner API with GraphQL pagination, replacing the mock fetcher.

**Implemented:**

1. **Infrastructure Layer - ShopifyPartnerClient:**
   - `FetchTransactions(ctx, accessToken, appID, from, to)` method
   - GraphQL query with pagination (100 per page)
   - Supported transaction types (Shopify Partner API only supports these):
     - AppSubscriptionSale → RECURRING
     - AppUsageSale → RECURRING
     - AppOneTimeSale → ONE_TIME
   - NOTE: AppCredit, ServiceSale, ReferralTransaction are NOT supported in transactions query
   - Context-based organization ID passing via `WithOrganizationID`
   - Amount parsing from decimal strings to cents

2. **Application Layer - SyncService:**
   - Updated to pass organization ID via context
   - Uses `external.WithOrganizationID(ctx, partnerAccount.PartnerID)`

3. **Main Integration:**
   - Wired `ShopifyPartnerClient` as `TransactionFetcher` in main.go
   - Configured ledger service with snapshot repository: `ledgerService.WithSnapshotRepository(snapshotRepo)`
   - This enables daily snapshots to be saved after each sync

4. **Debug Logging:**
   - Added token verification error logging to auth middleware
   - Added metrics fetch error logging to metrics handler

5. **Tests:**
   - `TestFetchTransactions_Success` - Basic transaction fetching
   - `TestFetchTransactions_Pagination` - Multi-page fetching
   - `TestFetchTransactions_NoOrganizationID` - Error handling
   - `TestFetchTransactions_GraphQLError` - GraphQL error handling
   - `TestFetchTransactions_HTTPError` - HTTP error handling
   - `TestFetchTransactions_EmptyTransactions` - Empty result handling
   - Fixed `TestFetchApps_Success` to match new implementation

**Database Notes:**
- `daily_metrics_snapshot` table requires columns: `total_revenue_cents`, `total_subscriptions`, `updated_at`
- If table exists without these columns, run ALTER TABLE to add them

**Tests:** 6 new tests (total backend: 123)

---

## [2026-02-27] Shop Name, Gross Amount, and Period-Based Usage Revenue

**Commit:** feat: add shop name, gross amount, and fix period-based usage revenue

**Summary:**
Added shop name and gross amount fields to transactions, fixed charge type inference using __typename, and fixed usage revenue to be calculated per period from transactions instead of cumulative snapshots.

**Implemented:**

1. **Transaction Entity Updates:**
   - Added `ShopName` field - Store display name from Shopify
   - Added `GrossAmountCents` field - Subscription price (what customer pays)
   - Renamed `AmountCents` to `NetAmountCents` - Revenue after Shopify's cut
   - Updated `NewTransaction` factory with new fields

2. **Subscription Entity Updates:**
   - Added `ShopName` field for store display name

3. **ShopifyPartnerClient Updates:**
   - Added `__typename` to GraphQL query for proper type identification
   - Fixed `inferChargeType` to use typename:
     - `AppSubscriptionSale` → RECURRING
     - `AppUsageSale` → USAGE
     - `AppOneTimeSale` → ONE_TIME
     - `AppCredit` → REFUND
   - Added `shop { name }` to GraphQL query
   - Added `grossAmount { amount currencyCode }` to query
   - New `parseAmounts` function returns both gross and net amounts

4. **Transaction Repository Updates:**
   - Updated INSERT/UPDATE queries for new fields
   - Added `charge_type = EXCLUDED.charge_type` to ON CONFLICT clause
   - Fixed SELECT queries to handle nullable shop_name and gross_amount_cents

5. **Subscription Repository Updates:**
   - Added shop_name to INSERT/SELECT queries

6. **MetricsAggregationService Refactor:**
   - Added `TransactionRepository` dependency
   - Added `MetricsEngine` dependency
   - `GetPeriodMetrics` now fetches transactions for specific date range
   - Usage and total revenue calculated from transactions (not snapshots)
   - Point-in-time metrics (MRR, risk states) still from snapshots
   - **Fix:** Usage revenue now varies by time filter (was same for all periods)

7. **Frontend Fixes:**
   - Fixed `subscription_tile.dart` index out of range errors
   - Added defensive string handling in `_getInitials` and `_formatDisplayName`

8. **Migrations:**
   - `000010_add_shop_name_to_subscriptions` - Add shop_name column
   - `000011_add_shop_name_and_gross_amount_to_transactions` - Add shop_name, gross_amount_cents columns

**Tests:** All tests passing (124 backend)

---

## [2026-02-27] Revenue API Implementation

**Commit:** feat: implement Revenue API with REST and GraphQL endpoints

**Summary:**
Implemented external Revenue API for Shopify app developers to query subscription payment status and usage billing status via REST and GraphQL endpoints. Uses CQRS pattern with a dedicated read model built from the ledger.

**Implemented:**

1. **Database Migrations (4 tables):**
   - `000012_create_api_keys_table` - API keys with SHA-256 hash storage
   - `000013_create_api_subscription_status_table` - CQRS read model for subscriptions
   - `000014_create_api_usage_status_table` - Usage billing status
   - `000015_create_api_audit_log_table` - Request audit logging

2. **Domain Layer (`internal/revenue_api/domain/`):**
   - **Entities:**
     - `APIKey` - API key with NewAPIKey(), HashKey() using SHA-256
     - `SubscriptionStatus` - Read model with risk state, payment status
     - `UsageStatus` - Usage billing with parent subscription reference
     - `AuditLog` - Request audit entry
   - **Repository Interfaces:**
     - `APIKeyRepository` - CRUD with FindByKeyHash
     - `SubscriptionStatusRepository` - Read model queries
     - `UsageStatusRepository` - Usage queries
     - `AuditLogRepository` - Async audit logging

3. **Infrastructure Layer (`internal/revenue_api/infrastructure/persistence/`):**
   - PostgreSQL implementations for all repositories
   - Async audit logging with background goroutine

4. **Application Layer (`internal/revenue_api/application/service/`):**
   - `APIKeyService` - Create, List, Revoke, ValidateKey
   - `SubscriptionStatusService` - GetByShopifyGID, GetByDomain, GetByShopifyGIDs
   - `UsageStatusService` - GetByShopifyGID, GetByShopifyGIDs
   - `RevenueReadModelBuilder` - Rebuilds read model from ledger

5. **HTTP Layer (`internal/revenue_api/interfaces/http/`):**
   - **Middleware:**
     - `APIKeyAuth` - X-API-Key header validation
     - `RateLimiter` - In-memory token bucket (Redis interface ready)
     - `AuditLogger` - Async request logging
   - **Handlers:**
     - `APIKeyHandler` - POST/GET/DELETE /api-keys
     - `SubscriptionStatusHandler` - REST endpoints
     - `UsageStatusHandler` - REST endpoints

6. **GraphQL Layer (`internal/revenue_api/interfaces/graphql/`):**
   - **Schema (`schema.graphql`):**
     ```graphql
     type Query {
       subscription(shopifyGid: ID!): SubscriptionStatus
       subscriptionByDomain(domain: String!): SubscriptionStatus
       subscriptions(shopifyGids: [ID!]!): SubscriptionBatchResult!
       usage(shopifyGid: ID!): UsageStatus
       usages(shopifyGids: [ID!]!): UsageBatchResult!
     }
     ```
   - `gqlgen.yml` - gqlgen configuration
   - `resolver.go` - Root resolver with enums
   - `schema.resolvers.go` - Query resolvers
   - `handler.go` - HTTP handler for /graphql

7. **Router (`internal/revenue_api/interfaces/http/router/router.go`):**
   - Separate router for Revenue API
   - API key management routes (Firebase auth)
   - Public API routes (API key auth)

8. **API Endpoints:**
   - **API Key Management (requires Firebase auth):**
     - `POST /v1/api-keys` - Create new API key
     - `GET /v1/api-keys` - List user's API keys
     - `DELETE /v1/api-keys/{keyID}` - Revoke API key
   - **REST Endpoints (requires API key):**
     - `GET /v1/subscriptions/{shopify_gid}` - Get subscription by GID
     - `GET /v1/subscriptions/status?domain={domain}` - Get subscription by domain
     - `POST /v1/subscriptions/batch` - Batch lookup (max 100)
     - `GET /v1/usages/{shopify_gid}` - Get usage by GID
     - `POST /v1/usages/batch` - Batch lookup (max 100)
   - **GraphQL Endpoint (requires API key):**
     - `POST /v1/graphql` - GraphQL queries

9. **Security Features:**
   - API keys hashed with SHA-256 (raw key never stored)
   - Tenant isolation (users can only access their own apps)
   - Rate limiting per API key (default: 60 requests/minute)
   - Request audit logging

10. **CQRS Pattern:**
    - Read model separate from write ledger
    - `RevenueReadModelBuilder` rebuilds from transactions
    - Eventually consistent (rebuilt on sync)

**Files Created:**
- 4 migrations
- 4 domain entities
- 4 repository interfaces
- 4 PostgreSQL repositories
- 4 application services
- 3 HTTP middleware
- 3 HTTP handlers
- 4 GraphQL files
- 1 router

**Tests:** Pending (Task #10)

---

## [2026-02-27] Revenue API Documentation Site

**Commits:** `33686f6`, `bc4f61e`, `d1b8189`, `100f3ef`, `200af3c`

**Summary:**
Created comprehensive API documentation for clients using two platforms: Mintlify (hosted) and custom Next.js (self-hosted).

**Implemented:**

1. **Mintlify Documentation (`docs/api/`):**
   - `mint.json` - Mintlify configuration with branding
   - `openapi.yaml` - OpenAPI 3.1 specification
   - MDX pages for all endpoints and concepts
   - Endpoint pages: subscriptions (single, batch, by-domain), usage (single, batch)
   - GraphQL documentation pages

2. **Custom Next.js Documentation (`docs/site/`):**
   - Next.js 14 with App Router
   - TailwindCSS for styling
   - Dark/light mode toggle
   - Responsive sidebar navigation
   - Code blocks with copy button and language tabs
   - Components: Header, Sidebar, CodeBlock, CodeTabs, Callout, Endpoint

3. **Documentation Pages:**
   - Getting Started: Introduction, Quick Start, Authentication
   - Core Concepts: Subscription Status, Risk States, Usage Billing
   - REST API: Overview, 5 endpoint pages
   - GraphQL: Overview, Schema, Queries, Examples
   - Resources: Error Codes, Rate Limits, Best Practices, Changelog

4. **Deployment:**
   - `DEPLOYMENT.md` - Vercel deployment guide
   - Custom domain setup instructions

**Files Created:**
- Mintlify: 25 files in `docs/api/`
- Custom Next.js: 30+ files in `docs/site/`

---

## [2026-02-28] Subscriptions Page Premium Analytics

**Commit:** feat: upgrade subscriptions page with advanced filtering, sorting, and pagination

**Summary:**
Transformed the basic subscriptions list into a premium SaaS-level reporting interface with server-side filtering, sorting, pagination, and distinct price-based filtering.

**Implemented:**

1. **New Endpoints:**
   - `GET /api/v1/apps/{appID}/subscriptions/summary` - Returns aggregate statistics
   - `GET /api/v1/apps/{appID}/subscriptions/price-stats` - Returns distinct prices with counts

2. **Enhanced List Endpoint:**
   - Multi-select status filter (comma-separated risk states)
   - Price range filter (priceMin, priceMax in cents)
   - Billing interval filter (MONTHLY, ANNUAL)
   - Search filter (shop_name or myshopify_domain)
   - Sort parameters (sortBy, sortOrder)
   - Pagination (page, pageSize)

3. **Domain Layer - Repository:**
   - `SubscriptionFilters` struct with RiskStates, PriceMinCents, PriceMaxCents, BillingInterval, SearchTerm, SortBy, SortOrder, Page, PageSize
   - `SubscriptionPage` struct with Subscriptions, Total, Page, PageSize, TotalPages
   - `SubscriptionSummary` struct with ActiveCount, AtRiskCount, ChurnedCount, AvgPriceCents, TotalCount
   - `PriceStats` struct with MinCents, MaxCents, AvgCents, Prices (distinct prices with counts)
   - `PricePoint` struct with PriceCents and Count

4. **Infrastructure Layer - Persistence:**
   - `FindWithFilters(ctx, appID, filters)` - Dynamic SQL with WHERE conditions
   - `GetSummary(ctx, appID)` - Aggregate statistics query
   - `GetPriceStats(ctx, appID)` - Distinct prices with counts, sorted ascending

5. **Handler:**
   - `Summary()` - Returns subscription statistics
   - `PriceStats()` - Returns distinct prices with counts for filter dropdown
   - Enhanced `List()` with all new query parameters
   - Legacy support for `risk_state` and `limit/offset` params

6. **Database Indexes:**
   - `idx_subscriptions_filters` - Composite index on (app_id, risk_state, base_price_cents, billing_interval)
   - `idx_subscriptions_search` - Index on (app_id, lower(shop_name), lower(myshopify_domain))

**API Response Examples:**

Summary endpoint:
```json
{
  "activeCount": 847,
  "atRiskCount": 23,
  "churnedCount": 12,
  "avgPriceCents": 4999,
  "totalCount": 882
}
```

Price stats endpoint:
```json
{
  "minCents": 499,
  "maxCents": 40499,
  "avgCents": 4389,
  "prices": [
    { "priceCents": 499, "count": 281 },
    { "priceCents": 1283, "count": 1 },
    { "priceCents": 4999, "count": 156 }
  ]
}
```

**Files Modified:**
- `internal/domain/repository/subscription_repository.go` - Added filter types, PriceStats, PricePoint
- `internal/infrastructure/persistence/subscription_repository.go` - Implemented FindWithFilters, GetSummary, GetPriceStats
- `internal/interfaces/http/handler/subscription.go` - Added Summary, PriceStats handlers
- `internal/interfaces/http/router/router.go` - Added new routes
- `migrations/000016_add_subscription_filter_indexes` - New indexes

**Tests:** Updated subscription handler tests

---

## [2026-03-01] Revenue Share Tier Tracking (Phase 1)

**Commit:** feat: add revenue share tier tracking with fee breakdown

**Summary:**
Implemented accurate Shopify revenue share tier tracking based on Shopify Partner API documentation. Revenue share is NOT always 20% - it varies by tier (0%, 15%, or 20%). Processing fee of 2.9% always applies. Tax is calculated on Shopify's fees, not gross revenue.

**Key Features:**
- 4 revenue share tiers: DEFAULT_20 (20%), SMALL_DEV_0 (0%), SMALL_DEV_15 (15%), LARGE_DEV_15 (15%)
- Fee breakdown calculation: gross → revenue share → processing fee (2.9%) → tax on fees → net
- Fee verification service to compare expected vs actual fees from Shopify
- Tier savings comparison (how much saved vs default 20%)

**Backend Files Created:**
- `internal/domain/valueobject/revenue_share_tier.go` - RevenueShareTier value object with 4 tiers, FeeBreakdown struct, calculation methods
- `internal/domain/valueobject/revenue_share_tier_test.go` - Unit tests
- `internal/domain/service/fee_verification_service.go` - Service for verifying fees and calculating summaries
- `internal/domain/service/fee_verification_service_test.go` - Unit tests
- `internal/interfaces/http/handler/fee_handler.go` - HTTP handlers for fee endpoints
- `migrations/000017_add_revenue_share_tier.up.sql` - Database migration
- `migrations/000017_add_revenue_share_tier.down.sql` - Rollback migration

**Backend Files Modified:**
- `internal/domain/entity/app.go` - Added RevenueShareTier field, SetRevenueShareTier method
- `internal/domain/entity/transaction.go` - Added fee breakdown fields (ShopifyFeeCents, ProcessingFeeCents, TaxOnFeesCents)
- `internal/infrastructure/persistence/app_repository.go` - Updated queries for tier field
- `internal/infrastructure/persistence/transaction_repository.go` - Updated queries for fee breakdown
- `internal/interfaces/http/handler/app.go` - Added UpdateAppTier handler
- `internal/interfaces/http/router/router.go` - Added routes for tier and fee endpoints

**New API Endpoints:**
- `PATCH /api/v1/apps/{appID}/tier` - Update app's revenue share tier
- `GET /api/v1/apps/{appID}/fees/summary` - Get fee summary with savings
- `GET /api/v1/apps/{appID}/fees/breakdown` - Get fee breakdown for amount

**Database Changes:**
```sql
ALTER TABLE apps ADD COLUMN revenue_share_tier VARCHAR(20) DEFAULT 'DEFAULT_20';
ALTER TABLE apps ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE transactions ADD COLUMN gross_amount_cents BIGINT;
ALTER TABLE transactions ADD COLUMN shopify_fee_cents BIGINT;
ALTER TABLE transactions ADD COLUMN processing_fee_cents BIGINT;
ALTER TABLE transactions ADD COLUMN tax_on_fees_cents BIGINT;
ALTER TABLE transactions ADD COLUMN net_amount_cents BIGINT;
```

**Tests:** All backend tests pass

---

## [2026-03-01] KPI Metrics Visualization Component

**Commit:** feat(marketing): add KPI metrics visualization component

**Summary:**
Created interactive KPI metrics guide component for the marketing site, showing how LedgerGuard calculates and presents revenue KPIs for Shopify app developers.

**Implemented:**

1. **KPI Cards (6 metrics):**
   - Active MRR - MRR from SAFE subscriptions only
   - Revenue at Risk - MRR from ONE_CYCLE_MISSED + TWO_CYCLES_MISSED subscriptions
   - Renewal Success Rate - (Safe Count / Total) × 100
   - Usage Revenue - SUM where ChargeType = USAGE
   - Total Revenue - RECURRING + USAGE + ONE_TIME - REFUNDS
   - Churned Revenue - MRR from CHURNED subscriptions

2. **View Modes:**
   - **Overview** - Formula display + risk timeline + risk distribution
   - **Detail** - Data flow animation + subscription list with MRR highlighting
   - **Comparison** - Period-over-period delta with semantic coloring

3. **Visualizations:**
   - Risk Classification Timeline (30/60/90 day thresholds with animated cursor)
   - Risk Distribution Bar Chart (Safe, At Risk, Critical, Churned)
   - Data Flow Animation (Partner API → Ledger Rebuild → Metrics Engine → Dashboard)
   - Subscription List with risk state filtering and MRR display
   - Period Comparison with delta calculation and good/bad semantics

4. **Delta Semantics:**
   - Higher is good: Active MRR, Renewal Rate, Usage Revenue, Total Revenue (green ↑)
   - Lower is good: Revenue at Risk, Churned Revenue (green ↓)

5. **Animation Features:**
   - Continuous animation loop with progress tracking
   - Risk state highlighting based on selected KPI
   - Formula reveal animations
   - Data flow step highlighting

**Files Created:**
- `marketing/site/components/KPIMetricsGuide.tsx` - Main component (1264 lines)
- `marketing/site/app/kpi-guide/page.tsx` - Page route

**Page URL:** http://localhost:3000/kpi-guide

**Related Prompt Document:** `docs/prompts/kpi-metrics-visualization.md`

---

## [2026-03-01] Earnings Timeline Tracking (Phase 2)

**Commit:** feat: add earnings timeline with pending/available status tracking

**Summary:**
Implemented Shopify earnings availability tracking. Shopify holds payments for a "cool-down" period before making them available for payout. This feature tracks when earnings become PENDING → AVAILABLE → PAID_OUT.

**Key Features:**
- Earnings availability calculation based on Shopify rules:
  - RECURRING charges: 7-37 days (depending on transaction date in billing cycle)
  - ONE_TIME charges: 7 days
- Earnings status tracking: PENDING, AVAILABLE, PAID_OUT
- API endpoint for earnings status summary
- Dashboard widget showing pending vs available earnings

**Backend Files Created:**
- `internal/domain/service/earnings_calculator.go` - EarningsCalculator with:
  - `CalculateAvailableDate(chargeType, createdDate)` - Calculates availability date
  - `DetermineEarningsStatus(availableDate, now)` - Returns PENDING/AVAILABLE status
  - `ProcessTransaction(tx, createdDate, now)` - Sets earnings fields on transaction
  - `ProcessTransactions(txs, now)` - Batch processing
  - `SummarizeEarnings(txs)` - Aggregates by status
- `internal/domain/service/earnings_calculator_test.go` - TDD unit tests
- `migrations/000018_add_earnings_tracking.up.sql` - Add earnings fields to transactions
- `migrations/000018_add_earnings_tracking.down.sql` - Rollback migration

**Backend Files Modified:**
- `internal/domain/entity/transaction.go` - Added EarningsStatus type, CreatedDate, AvailableDate, EarningsStatus fields
- `internal/domain/repository/transaction_repository.go` - Added EarningsSummary, EarningsByDate structs, GetEarningsSummary, GetPendingByAvailableDate, GetUpcomingAvailability methods
- `internal/infrastructure/persistence/transaction_repository.go` - Implemented new repository methods with SQL queries
- `internal/application/service/revenue_metrics_service.go` - Added GetEarningsStatus method
- `internal/application/service/sync_service.go` - Integrated EarningsCalculator before storing transactions (critical fix for constraint violation)
- `internal/interfaces/http/handler/revenue_handler.go` - Added GetEarningsStatus handler
- `internal/interfaces/http/router/router.go` - Added route for earnings/status endpoint

**New API Endpoint:**
- `GET /api/v1/apps/{appID}/earnings/status` - Returns earnings summary
  ```json
  {
    "total_pending_cents": 12500,
    "total_available_cents": 45000,
    "total_paid_out_cents": 120000,
    "pending_by_date": [
      { "date": "2026-03-08", "amount_cents": 5000 },
      { "date": "2026-03-15", "amount_cents": 7500 }
    ],
    "upcoming_availability": [
      { "date": "2026-03-08", "amount_cents": 5000 }
    ]
  }
  ```

**Database Changes:**
```sql
ALTER TABLE transactions ADD COLUMN created_date TIMESTAMPTZ;
ALTER TABLE transactions ADD COLUMN available_date TIMESTAMPTZ;
ALTER TABLE transactions ADD COLUMN earnings_status VARCHAR(20) DEFAULT 'PENDING';
ALTER TABLE transactions ADD CONSTRAINT transactions_earnings_status_check
  CHECK (earnings_status IN ('PENDING', 'AVAILABLE', 'PAID_OUT'));
CREATE INDEX idx_transactions_earnings ON transactions(app_id, earnings_status, available_date);
```

**Configuration Change:**
- Changed history sync window from 12 months to 3 months (sync_service.go and ledger_service.go)

**Tests:** All backend tests pass

---


---

## [2026-03-01] Phase 1: Data Integrity Fixes

**Commit:** Phase 1 comprehensive review fixes

**Summary:**
Implemented Phase 1 fixes from comprehensive app review, focusing on data integrity issues.

**Changes:**

### 1. Risk Classification Logic (P0-1)
- Added SubscriptionStatus value object with ACTIVE, CANCELLED, FROZEN, EXPIRED, PENDING states
- Updated RiskEngine to check subscription status before days-based classification:
  - CANCELLED/EXPIRED → CHURNED (terminal states)
  - FROZEN → ONE_CYCLE_MISSED (payment failure)
  - PENDING → SAFE (not yet active)
  - ACTIVE → Check days past due
- Added 10+ new boundary tests (exactly 30/31/60/61/90/91 days)
- Added status-specific tests (CANCELLED ignores timing, FROZEN overrides days)

**Files:**
- `internal/domain/valueobject/subscription_status.go` (new)
- `internal/domain/service/risk_engine.go` (updated)
- `internal/domain/service/risk_engine_test.go` (updated)

### 2. MRR Annual Normalization (P0-3)
- Verified already implemented in `Subscription.MRRCents()` method
- Annual subscriptions correctly divide by 12

### 3. GraphQL Query Fields (P0-7)
- Added shop.id, shop.plan.displayName to transaction query
- Added appSubscription fields: id, name, status, currentPeriodEnd, lineItems.plan.pricingDetails
- Added appUsageRecord fields: id, description
- Updated transactionNode struct with new fields
- Updated Transaction entity with new fields:
  - ShopifyShopGID, ShopPlan (shop details)
  - SubscriptionGID, SubscriptionStatus, SubscriptionPeriodEnd, BillingInterval (subscription reference)
- Updated transaction repository (Upsert, UpsertBatch, FindByAppID, FindByShopifyGID, FindByDomain)

**Files:**
- `internal/infrastructure/external/shopify_partner_client.go` (updated)
- `internal/domain/entity/transaction.go` (updated)
- `internal/infrastructure/persistence/transaction_repository.go` (updated)
- `migrations/000020_add_transaction_subscription_details.up.sql` (new)
- `migrations/000020_add_transaction_subscription_details.down.sql` (new)

### 4. Critical Domain Service Tests (P0-6)
- Added boundary tests for risk classification thresholds
- Added status-override tests
- Verified existing test coverage for metrics engine and ledger service

**Test Count:** 55+ domain service tests passing

---

## [2026-03-01] Phase 3: Real-time Capabilities - Webhook Integration

**Commit:** feat: implement webhook integration for real-time subscription updates

**Summary:**
Implemented Phase 3 Task #1 - Webhook Integration (P1-8). Added support for Shopify webhooks to receive real-time updates for subscription status changes, app uninstalls, and billing failures. This enables immediate risk state updates instead of waiting for the next sync.

**Implemented:**

### 1. Subscription Event Entity
- Created `SubscriptionEvent` entity for tracking lifecycle events
- Fields: ID, SubscriptionID, FromStatus/ToStatus, FromRiskState/ToRiskState, EventType, Reason, OccurredAt, CreatedAt
- Helper methods: `IsChurnEvent()`, `IsVoluntaryChurn()`, `IsInvoluntaryChurn()`, `IsReactivationEvent()`

### 2. Webhook Service
- HMAC signature validation for webhook security
- Event processing for three webhook types:
  - `app_subscriptions/update` - Status changes (ACTIVE, CANCELLED, FROZEN, EXPIRED)
  - `app/uninstalled` - App uninstallation (soft delete subscription)
  - `subscription_billing_attempts/failure` - Payment failures (escalate risk state)
- Risk state escalation: SAFE → ONE_CYCLE_MISSED → TWO_CYCLES_MISSED → CHURNED
- Lifecycle event recording for audit trail

### 3. Webhook Handler
- HTTP handler for `/webhooks/shopify/*` routes
- Extracts headers: X-Shopify-Topic, X-Shopify-Shop-Domain, X-Shopify-Hmac-Sha256
- Returns 200 OK even on processing errors (prevents Shopify retries)
- Separate endpoints for each webhook type

### 4. Repository Layer
- `SubscriptionEventRepository` interface
- PostgreSQL implementation with queries:
  - FindBySubscriptionID, FindByAppID, FindChurnEvents, CountByEventType
- Added `FindAllByPartnerAppID` to AppRepository for webhook lookups

### 5. Database Migration
- Migration 000023: subscription_events table
- Indexes for subscription_id, event_type+occurred_at, churn events

### 6. Value Object Updates
- Added `ParseRiskState()` function for database deserialization

**Files Created:**
- `internal/domain/entity/subscription_event.go`
- `internal/application/service/webhook_service.go`
- `internal/interfaces/http/handler/webhook.go`
- `internal/domain/repository/subscription_event_repository.go`
- `internal/infrastructure/persistence/subscription_event_repository.go`
- `migrations/000023_create_subscription_events_table.up.sql`
- `migrations/000023_create_subscription_events_table.down.sql`

**Files Updated:**
- `internal/domain/repository/app_repository.go` - Added FindAllByPartnerAppID
- `internal/infrastructure/persistence/app_repository.go` - Implemented FindAllByPartnerAppID
- `internal/domain/valueobject/risk_state.go` - Added ParseRiskState
- `internal/interfaces/http/router/router.go` - Added webhook routes
- `DATABASE_SCHEMA.md` - Documented subscription_events table
- Test mocks updated for new interface method

---

## [2026-03-01] Phase 4: Architecture & Documentation

**Commit:** docs: update ER diagrams, sequence diagrams, and API documentation

**Summary:**
Completed Phase 4 - Architecture documentation updates. Updated all PlantUML diagrams to match current implementation and extended OpenAPI specification with webhook endpoints.

**Implemented:**

### 1. ER Diagram Updates (P1-10)
Updated `docs/ER_current.puml` to match actual database schema:
- Added `subscription_events` table (Phase 3 addition)
- Added `user_preferences` table (migration 000019)
- Added `api_subscription_status` table (CQRS read model)
- Added `api_usage_status` table (CQRS read model)
- Added `api_audit_log` table
- Updated `subscriptions` with `shopify_shop_gid`, `deleted_at` columns
- Updated `transactions` with shop and subscription reference columns
- Added all relationships and index documentation

### 2. Manual Charge Handling Verification (P1-11)
Verified correct implementation:
- `ChargeTypeOneTime` defined in `charge_type.go`
- ONE_TIME charges excluded from MRR (only RECURRING creates subscriptions)
- ONE_TIME charges included in total revenue (`metrics_engine.go:58`)
- Test coverage in `ledger_service_test.go`, `metrics_engine_test.go`

### 3. Sequence Diagrams Updates (P2-17)
Added missing flows to `docs/SEQUENCE_current.puml`:
- **Webhook Processing Flow** (section 10): subscription updates, app uninstalls, billing failures
- **AI Daily Brief Generation Flow** (section 11): scheduled generation and on-demand retrieval
- **Revenue API Flow** (section 12): external API access with authentication and rate limiting

### 4. API Documentation Updates (P2-16)
Extended `docs/api/openapi.yaml`:
- Added Webhooks tag and documentation
- Added `/webhooks/shopify` endpoint (generic handler)
- Added `/webhooks/shopify/subscriptions` endpoint
- Added `/webhooks/shopify/uninstalled` endpoint
- Added `/webhooks/shopify/billing-failure` endpoint
- Added webhook payload schemas: `SubscriptionUpdatePayload`, `AppUninstalledPayload`, `BillingFailurePayload`

**Files Updated:**
- `docs/ER_current.puml` - Complete schema with all 14 tables
- `docs/SEQUENCE_current.puml` - Added 3 new sequence diagrams
- `docs/api/openapi.yaml` - Added webhook endpoint documentation


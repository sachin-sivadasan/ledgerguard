# Prompts Log – LedgerGuard

## Format
```
### [YYYY-MM-DD] Prompt Title
**Original:**
> User's original prompt

**Improved:**
> Enhanced version executed

**Result:**
- Files changed
- Commit hash
```

---

## Log

### [2026-05-18] Fix Silent $0 Dashboard When Postgres Is Down
**Original:**
> Implement the following plan: Fix: Silent $0 Dashboard When Postgres Is Down — distinguish DB errors from not-found in backend, add service-unavailable dashboard state with auto-retry in frontend.

**Improved:**
> Fix silent $0 dashboard and false "No data yet" states when Postgres is unreachable. (1) Backend auth middleware: return 503 instead of 500 for DB connection errors; (2) Backend app_lookup.go: distinguish sentinel not-found errors (→ 404) from DB errors (→ 503) via `isNotFoundError()` helper; (3) Frontend Bloc app (`frontend/app/`): add `ServiceUnavailableException`, `DashboardServiceUnavailable` state, auto-retry (3×, 15s); (4) Frontend Provider app (`frontend-flutter/`): rethrow 503 from `OrgService.listOrganizations()`, add `isServiceUnavailable` to OrganizationProvider/AppsProvider/DashboardProvider, intercept at `AppShell` level to show cloud_off UI on all screens; (5) Fix existing test mocks using generic `errors.New("not found")` to use sentinel errors.

**Result:**
- 22 files changed (2 new, 20 modified)
- Commit `f57914b`

---

### [2026-05-09] Persist Org + Default App Selection via Backend Preferences
**Original:**
> Wire backend user_preferences endpoints (selected-org, default-app) to frontend OrganizationProvider and AppsProvider so selection persists across sessions.

**Improved:**
> Wire frontend providers to backend `user_preferences` table for org/app persistence. (1) Create `UserPreferencesService` with GET/PUT for selected-org and default-app; (2) `OrganizationProvider.loadOrganizations()` resolves saved org before auto-selecting first; (3) `AppsProvider.loadApps()` resolves saved app, adds `selectedAppId` getter and `setSelectedApp()` method; (4) `DataLoadingMixin` uses `selectedAppId` instead of `apps.first.id`; (5) Wire service in `main.dart`; (6) PlantUML sequence diagram.

**Result:**
- 7 files changed (1 new service, 1 new diagram, 5 modified)

---

### [2026-05-08] Migrate from Numeric Shopify App IDs to Internal UUIDs
**Original:**
> Implement the UUID migration plan: Backend shared resolveAppFromRequest(), frontend UUID as canonical ID, backend UUID-only enforcement, documentation.

**Improved:**
> Migrate from numeric Shopify app IDs to internal UUIDs across full stack. (1) Create unified `resolveAppFromRequest()` replacing 5 duplicate helpers and 4 duplicate GID constants across 13 handlers; (2) Frontend `ShopifyApp.id` = UUID, add `shopifyId` for display, remove `_idToUuid` mapping hack in SyncStatusProvider; (3) Enforce UUID-only in backend, reject numeric IDs with 400; (4) Update all handler tests to use UUIDs in URL params.

**Result:**
- ~18 files changed across 3 commits
- Commits `ee6f1bc`, `9ede232`, `c3b4daa`

### [2026-05-08] Fix Flutter Provider Data-Loading Bugs
**Original:**
> Implement the following plan: Fix Flutter Provider Data-Loading Bugs (DataLoadingMixin, CancelToken, LgErrorState, DemoModeCoordinator, bootstrap chain fix, screen migration)

**Improved:**
> Fix 6 recurring data-loading bugs across the Flutter Provider frontend: (1) Replace side effects in build() with listener-based DataLoadingMixin; (2) Remove fragile _wasDemoMode (7 screens) and hasAttemptedLoad (2 providers) one-shot booleans; (3) Add LgErrorState widget with retry for error recovery; (4) Add CancelToken to ApiClient, all 8 services, and all 9 providers for request cancellation; (5) Fix bootstrap chain in app.dart by removing _orgLoaded/_appsLoaded booleans; (6) Centralize 9 individual setDemoMode() calls with DemoModeCoordinator.

**Result:**
- 3 new files: `data_loading_mixin.dart`, `lg_error_state.dart`, `demo_mode_coordinator.dart`
- 30 modified files across api_client, 8 services, 9 providers, 8 screens, app.dart, main.dart, settings_screen
- Commit `11c0c7b`

### [2026-05-08] Whale Persona (~1M Transactions) + Usage Charges Admin UI
**Original:**
> Implement the plan: Add Whale Persona (~1M Transactions) + Usage Charges Admin UI

**Improved:**
> Add a "Whale Partner" persona (org 1005) to mock-shopify-api for load testing: 5 apps, 50K shops, ~1M transactions generated programmatically from a compact YAML `generated` block. Implement DataStore.expand_generated() with deterministic seeded RNG. Add per-org transaction caching to TransactionResolver. Add usage charges table + form to admin UI (persona.erb). Cap subscriptions table at 200 rows. Add cache invalidation to all data-modifying routes.

**Result:**
- `mock-shopify-api/data/whale.yml` (NEW) — whale persona definition
- `mock-shopify-api/data/personas.yml` — org 1005 entry
- `mock-shopify-api/lib/data_store.rb` — expand_generated(), stats() update
- `mock-shopify-api/lib/transaction_resolver.rb` — transaction caching
- `mock-shopify-api/lib/graphql_handler.rb` — invalidate_transaction_cache()
- `mock-shopify-api/views/persona.erb` — usage charges UI + subscription cap
- `mock-shopify-api/app.rb` — usage charges route + cache invalidation
- `IMPLEMENTATION_LOG.md` — entry added
- `prompts.md` — prompt logged

### [2026-05-08] Fix Flutter: Org Header Not Sent + Disable Demo Mode by Default
**Original:**
> Implement the plan: Fix Flutter org header not sent + disable demo mode by default

**Improved:**
> Fix two Flutter frontend issues: (1) X-Org-Id header never sent because loadOrganizations() is never called on app startup — add auth listener in app.dart that triggers the org load chain on login; (2) All 10 data providers default to demoMode=true — change to false so new users get live mode by default. Update settings toggle subtitle.

**Result:**
- `frontend-flutter/lib/app.dart` — auth listener + org load on startup
- 10 provider files — `_demoMode = false`
- `frontend-flutter/lib/screens/settings/settings_screen.dart` — updated subtitle
- `IMPLEMENTATION_LOG.md` — entry added
- `prompts.md` — prompt logged

### [2026-05-08] Fix 500 on POST /integrations/shopify/token
**Original:**
> 500 86B on POST /api/v1/integrations/shopify/token

**Improved:**
> Investigate and fix the 500 error on POST /api/v1/integrations/shopify/token. The error is caused by NewPartnerAccount() not setting OrgID, violating the NOT NULL FK constraint from migration 000037. Fix the entity constructor, add OrgContextMW to the token routes, resolve org in the OAuth callback via memberRepo, and update tests.

**Result:**
- `backend/internal/domain/entity/partner_account.go` — added orgID param
- `backend/internal/interfaces/http/handler/manual_token.go` — org from context
- `backend/internal/interfaces/http/handler/oauth.go` — memberRepo + org resolution
- `backend/internal/interfaces/http/router/router.go` — OrgContextMW on token routes
- `backend/internal/interfaces/http/middleware/org_context.go` — SetOrgContext helper
- `backend/cmd/server/main.go` — wire memberRepo into OAuthHandler
- Tests updated, all passing

### [2025-02-26] Initial Setup
**Original:**
> init local first / origin git@github.com:sachin-sivadasan/ledgerguard.git

**Improved:**
> Initialize local git repository, create .gitignore, add remote origin, commit documentation files (PRD.md, TAD.md, DATABASE_SCHEMA.md, CLAUDE.md)

**Result:**
- Initialized git repo
- Created .gitignore
- Added remote origin
- Initial commit with docs

### [2026-02-26] Initialize Go Backend
**Original:**
> Initialize Go backend project. Requirements: Clean Architecture folder structure, PostgreSQL connection, Migration setup, Basic health endpoint, Initialize TEST_PLAN.md. Do NOT implement business logic yet.

**Improved:**
> Initialize Go backend project with:
> 1. Create Clean Architecture folder structure (cmd/server, internal/domain, internal/usecase, internal/repository, internal/delivery/http, internal/service, pkg, migrations)
> 2. Initialize Go module (github.com/sachin-sivadasan/ledgerguard)
> 3. Add dependencies: chi router, pgx (PostgreSQL driver), golang-migrate
> 4. Create config loader from environment variables
> 5. PostgreSQL connection pool with health check
> 6. Migration setup with golang-migrate (infrastructure only)
> 7. Basic HTTP server with /health endpoint
> 8. Create TEST_PLAN.md with test scenarios
> 9. Write tests for health endpoint (TDD)
> 10. No business logic, domain entities, or usecases

**Result:**
- Created backend/ folder structure
- go.mod with chi, pgx, golang-migrate
- internal/infrastructure/config/config.go
- internal/infrastructure/database/postgres.go
- internal/infrastructure/database/migrate.go
- internal/delivery/http/health.go + health_test.go
- internal/delivery/http/router.go
- cmd/server/main.go
- TEST_PLAN.md
- All tests passing

### [2026-02-26] Refactor to DDD Architecture
**Original:**
> backend should be in domain driven architecture

**Improved:**
> Refactor backend from Clean Architecture to Domain-Driven Design (DDD) folder structure:
> 1. Restructure to: cmd/, internal/domain/, internal/application/, internal/infrastructure/, internal/interfaces/
> 2. Domain layer: entity/, valueobject/, service/, repository/ (interfaces)
> 3. Application layer: service/ (use cases), dto/
> 4. Infrastructure layer: config/, persistence/, external/
> 5. Interfaces layer: http/handler/, http/middleware/, http/router/
> 6. Move existing health endpoint code to new structure
> 7. Update TAD.md and DECISIONS.md with architecture change
> 8. Keep all tests passing

**Result:**
- Restructured to DDD: domain/, application/, infrastructure/, interfaces/
- Moved health handler to interfaces/http/handler/
- Moved router to interfaces/http/router/
- Moved database to infrastructure/persistence/
- Updated TAD.md with DDD structure
- Added ADR-005 for DDD decision
- Updated CLAUDE.md with new folder structure
- All tests passing (3/3)

### [2026-02-26] Firebase Auth Middleware
**Original:**
> Implement Firebase ID token verification middleware. Verify Firebase token, Extract firebase_uid, Auto-create user if first login, Assign OWNER role by default. Write tests first. Update DATABASE_SCHEMA.md.

**Improved:**
> Implement Firebase Authentication middleware:
> 1. Create User entity in domain/entity (ID, FirebaseUID, Email, Role, PlanTier, CreatedAt)
> 2. Create Role, PlanTier value objects in domain/valueobject
> 3. Create UserRepository interface in domain/repository
> 4. Create PostgresUserRepository in infrastructure/persistence
> 5. Create FirebaseAuthService in infrastructure/external
> 6. Create AuthMiddleware in interfaces/http/middleware
> 7. Write tests first (TDD): mock Firebase client, test token extraction, user creation
> 8. Add first SQL migration (000001_create_users_table)
> 9. Update DATABASE_SCHEMA.md and TEST_PLAN.md

**Result:**
- domain/entity/user.go - User entity
- domain/valueobject/role.go - Role (OWNER, ADMIN)
- domain/valueobject/plan_tier.go - PlanTier (FREE, PRO)
- domain/service/auth_service.go - AuthTokenVerifier interface
- domain/repository/user_repository.go - UserRepository interface
- infrastructure/persistence/user_repository.go - PostgreSQL implementation
- infrastructure/external/firebase_auth.go - Firebase Admin SDK integration
- interfaces/http/middleware/auth.go + auth_test.go - Auth middleware (6 tests)
- migrations/000001_create_users_table.up.sql / .down.sql
- Updated DATABASE_SCHEMA.md with migrations section
- Updated TEST_PLAN.md with auth test scenarios
- All tests passing (9/9)

### [2026-02-26] Config File Support
**Original:**
> read it from config file

**Improved:**
> Add config file support to configuration loader:
> 1. Support YAML config file (config.yaml)
> 2. Load order: defaults → config file → environment variables (env vars override)
> 3. Add config file path via -config flag or CONFIG_PATH env var
> 4. Add Firebase credentials path to config
> 5. Write tests for config loading (TDD)
> 6. Create config.example.yaml template
> 7. Update .gitignore for local config files

**Result:**
- infrastructure/config/config.go - YAML file loading with env override
- infrastructure/config/config_test.go - 5 tests
- config.example.yaml - Template with all options
- Updated main.go with -config flag and CONFIG_PATH env var
- Updated .gitignore for config.yaml, config.local.yaml
- All tests passing (14/14)

### [2026-02-26] Role-Based Access Middleware
**Original:**
> Implement role-based access middleware. Roles: OWNER, ADMIN. Restrict manual partner integration to ADMIN only. Add tests.

**Improved:**
> Implement role-based access control (RBAC) middleware:
> 1. Create RoleMiddleware in interfaces/http/middleware
> 2. Support multiple allowed roles per route
> 3. Return 403 Forbidden if user lacks required role
> 4. OWNER has all permissions (superset of ADMIN)
> 5. Create helper function RequireRoles(roles ...Role)
> 6. Write tests first (TDD)

**Result:**
- interfaces/http/middleware/role.go - RequireRoles middleware
- interfaces/http/middleware/role_test.go - 5 tests
- OWNER can access all routes (superset)
- ADMIN can only access ADMIN-allowed routes
- Updated TEST_PLAN.md with role middleware tests
- All tests passing (19/19)

### [2026-02-26] Shopify Partner OAuth Flow
**Original:**
> Implement Shopify Partner OAuth flow. - Redirect endpoint - Callback endpoint - Exchange code for access_token - Encrypt token before saving - Store partner_account. Write tests.

**Improved:**
> Implement Shopify Partner OAuth flow:
> 1. Create IntegrationType value object (OAUTH, MANUAL)
> 2. Create PartnerAccount entity
> 3. Create PartnerAccountRepository interface
> 4. Create PostgresPartnerAccountRepository implementation
> 5. Create AES-256-GCM encryption in pkg/crypto with tests
> 6. Create ShopifyOAuthService for token exchange with tests
> 7. Create OAuthHandler with StartOAuth and Callback endpoints
> 8. Add migration 000002_create_partner_accounts_table
> 9. Update config with Shopify and Encryption settings
> 10. Update router with OAuth routes
> 11. Write tests first (TDD)

**Result:**
- domain/valueobject/integration_type.go - IntegrationType (OAUTH, MANUAL)
- domain/entity/partner_account.go - PartnerAccount entity
- domain/repository/partner_account_repository.go - Repository interface
- infrastructure/persistence/partner_account_repository.go - PostgreSQL implementation
- infrastructure/external/shopify_oauth.go + shopify_oauth_test.go - OAuth service (3 tests)
- interfaces/http/handler/oauth.go + oauth_test.go - OAuth handlers (4 tests)
- pkg/crypto/aes.go + aes_test.go - AES-256-GCM encryption (5 tests)
- migrations/000002_create_partner_accounts_table.up.sql / .down.sql
- Updated config with Shopify and Encryption settings
- Updated router with OAuth routes (/api/v1/integrations/shopify/oauth, /callback)
- Updated TEST_PLAN.md with OAuth and encryption test scenarios
- Updated DATABASE_SCHEMA.md with partner_accounts migration
- All tests passing (24/24)

### [2026-02-26] Manual Partner Token Integration
**Original:**
> Implement manual partner token integration. - ADMIN only - Encrypt token - Mask token in API responses - Allow revoke

**Improved:**
> Implement manual Shopify Partner token integration (for development/testing):
> 1. Create ManualTokenHandler with endpoints:
>    - POST /api/v1/integrations/shopify/token - Add manual token (ADMIN only)
>    - GET /api/v1/integrations/shopify/token - Get token info (masked)
>    - DELETE /api/v1/integrations/shopify/token - Revoke token
> 2. Use existing AES-256-GCM encryption for token storage
> 3. Use existing PartnerAccount entity with IntegrationType = MANUAL
> 4. Mask token in responses (show only last 4 chars: `***...xxxx`)
> 5. Apply RequireRoles(ADMIN) middleware
> 6. Add Delete method to PartnerAccountRepository
> 7. Write tests first (TDD)
> 8. Update router, diagrams, documentation

**Result:**
- domain/repository/partner_account_repository.go - Added Delete method
- infrastructure/persistence/partner_account_repository.go - Added Delete implementation
- interfaces/http/handler/manual_token.go - ManualTokenHandler (AddToken, GetToken, RevokeToken)
- interfaces/http/handler/manual_token_test.go - 12 tests
- Updated router with /token routes (POST, GET, DELETE) with ADMIN middleware
- Updated TEST_PLAN.md with manual token test scenarios
- All tests passing (36/36)

### [2026-02-26] Fetch Apps from Partner API
**Original:**
> Fetch apps from Partner API. Allow user to select one app. Store selected app in apps table.

**Improved:**
> Implement Shopify Partner API app fetching and selection:
> 1. Create App entity in domain/entity
> 2. Create AppRepository interface and PostgreSQL implementation
> 3. Create migration 000003_create_apps_table
> 4. Create ShopifyPartnerClient in infrastructure/external for GraphQL API calls
> 5. Create AppHandler with endpoints:
>    - GET /api/v1/apps/available - Fetch apps from Partner API
>    - POST /api/v1/apps/select - Select and store an app
>    - GET /api/v1/apps - List user's tracked apps
> 6. Use decrypted partner token to call Shopify Partner API
> 7. Write tests first (TDD)
> 8. Update router, diagrams, documentation

**Result:**
- domain/entity/app.go - App entity
- domain/repository/app_repository.go - AppRepository interface
- infrastructure/persistence/app_repository.go - PostgreSQL implementation
- infrastructure/external/shopify_partner_client.go + tests - GraphQL client (4 tests)
- interfaces/http/handler/app.go + app_test.go - AppHandler (10 tests)
- migrations/000003_create_apps_table.up.sql / .down.sql
- Updated router with /apps routes
- Updated TEST_PLAN.md with app test scenarios
- Updated DATABASE_SCHEMA.md with apps migration
- All tests passing (50/50)

### [2026-02-26] Implement PartnerSyncService
**Original:**
> Implement PartnerSyncService. - Pull transactions (mock first) - Store transactions - Add 12-hour scheduler

**Improved:**
> Implement PartnerSyncService for transaction synchronization:
> 1. Create Transaction entity in domain/entity
> 2. Create ChargeType value object (RECURRING, USAGE, ONE_TIME, REFUND)
> 3. Create TransactionRepository interface and PostgreSQL implementation
> 4. Create migration 000004_create_transactions_table
> 5. Create SyncService in application/service with:
>    - SyncApp(appID) - Sync single app
>    - SyncAllApps(partnerAccountID) - Sync all apps for account
> 6. Create TransactionFetcher interface (mock for now)
> 7. Create SyncScheduler with 12-hour interval (00:00, 12:00 UTC)
> 8. Create SyncHandler for on-demand sync endpoints:
>    - POST /api/v1/sync - Sync all apps
>    - POST /api/v1/sync/{appID} - Sync specific app
> 9. Add FindByID to PartnerAccountRepository
> 10. Write tests first (TDD)
> 11. Update router, diagrams, documentation

**Result:**
- domain/valueobject/charge_type.go - ChargeType (RECURRING, USAGE, ONE_TIME, REFUND)
- domain/entity/transaction.go - Transaction entity
- domain/repository/transaction_repository.go - TransactionRepository interface
- domain/repository/partner_account_repository.go - Added FindByID method
- infrastructure/persistence/transaction_repository.go - PostgreSQL implementation with batch upsert
- infrastructure/persistence/partner_account_repository.go - Added FindByID implementation
- application/service/sync_service.go + sync_service_test.go - SyncService (5 tests)
- application/scheduler/sync_scheduler.go - 12-hour interval scheduler
- interfaces/http/handler/sync.go + sync_test.go - SyncHandler (6 tests)
- migrations/000004_create_transactions_table.up.sql / .down.sql
- Updated router with /sync routes
- Updated all mocks with FindByID method
- Updated TEST_PLAN.md with sync test scenarios
- Updated DATABASE_SCHEMA.md with transactions migration
- Updated ER_current.puml with transactions entity
- Updated SEQUENCE_current.puml with sync flow
- All tests passing (58/58)

### [2026-02-26] Implement Deterministic Ledger Rebuild
**Original:**
> Implement deterministic ledger rebuild.
> - Separate RECURRING and USAGE
> - Compute expected renewal date
> - Store last_recurring_charge_date

**Improved:**
> Implement deterministic ledger rebuild service:
> 1. Create Subscription entity in domain/entity (with last_recurring_charge_date, expected_next_charge_date, risk_state)
> 2. Create RiskState and BillingInterval value objects
> 3. Create SubscriptionRepository interface and PostgreSQL implementation
> 4. Create migration 000005_create_subscriptions_table
> 5. Create LedgerService in domain/service with:
>    - RebuildFromTransactions(appID) - Rebuild subscription state from transactions
>    - Separate RECURRING revenue (for MRR) from USAGE revenue
>    - Compute expected_next_charge_date based on billing_interval (MONTHLY +1 month, ANNUAL +1 year)
>    - Update last_recurring_charge_date from most recent RECURRING transaction
>    - Classify risk state based on days past due
> 6. Ensure deterministic: same input → same output
> 7. Write tests first (TDD)
> 8. Update diagrams, documentation

**Result:**
- domain/valueobject/risk_state.go - RiskState (SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED)
- domain/valueobject/billing_interval.go - BillingInterval (MONTHLY, ANNUAL) with NextChargeDate()
- domain/entity/subscription.go - Subscription entity with risk classification
- domain/repository/subscription_repository.go - SubscriptionRepository interface
- infrastructure/persistence/subscription_repository.go - PostgreSQL implementation
- domain/service/ledger_service.go + ledger_service_test.go - LedgerService (8 tests)
- migrations/000005_create_subscriptions_table.up.sql / .down.sql
- Updated TEST_PLAN.md with ledger and risk tests
- Updated DATABASE_SCHEMA.md with subscriptions migration
- Updated ER_current.puml with subscriptions entity
- Updated SEQUENCE_current.puml with ledger rebuild flow
- All tests passing (66/66)

### [2026-02-26] Implement RiskEngine
**Original:**
> Implement RiskEngine.
> States: SAFE ONE_CYCLE_MISSED TWO_CYCLE_MISSED CHURNED
> Recalculate each sync.

**Improved:**
> Implement RiskEngine integration with sync flow:
> 1. Create RiskEngine in domain/service that encapsulates risk classification logic
> 2. Integrate with SyncService to recalculate risk after each sync
> 3. Risk states: SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED
> 4. Trigger LedgerService.RebuildFromTransactions after each successful sync
> 5. Return risk summary in sync results
> 6. Write tests first (TDD)
> 7. Update documentation

**Result:**
- domain/service/risk_engine.go - RiskEngine service with classification methods
- domain/service/risk_engine_test.go - Comprehensive tests (12 test cases)
- application/service/sync_service.go - Added LedgerRebuilder interface, triggers rebuild after sync
- application/service/sync_service_test.go - Updated with mock LedgerRebuilder
- interfaces/http/handler/sync_test.go - Updated with mock LedgerRebuilder
- SyncResult now includes RiskSummary, RevenueAtRisk, TotalMRRCents
- Updated TEST_PLAN.md with RiskEngine test scenarios
- Updated IMPLEMENTATION_LOG.md with RiskEngine implementation
- All tests passing (88/88)

### [2026-02-26] Implement MetricsEngine
**Original:**
> Implement MetricsEngine.
> Compute: Renewal Success Rate, Active MRR, Revenue at Risk, Usage Revenue, Total Revenue
> Store daily snapshot.

**Improved:**
> Implement MetricsEngine for KPI computation and daily snapshots:
> 1. Create MetricsEngine in domain/service that computes:
>    - Renewal Success Rate = SAFE subscriptions / Total active subscriptions
>    - Active MRR = Sum of MRR from SAFE subscriptions only
>    - Revenue at Risk = MRR from ONE_CYCLE_MISSED + TWO_CYCLES_MISSED
>    - Usage Revenue = Sum of USAGE transactions (12-month window)
>    - Total Revenue = RECURRING + USAGE + ONE_TIME - REFUNDS
> 2. Create DailyMetricsSnapshot entity
> 3. Create DailyMetricsSnapshotRepository interface and PostgreSQL implementation
> 4. Create migration 000006_create_daily_metrics_snapshot_table with ALL columns:
>    - id, app_id, date
>    - active_mrr_cents, revenue_at_risk_cents, usage_revenue_cents, total_revenue_cents
>    - renewal_success_rate
>    - safe_count, one_cycle_missed_count, two_cycles_missed_count, churned_count, total_subscriptions
>    - created_at, updated_at
> 5. Integrate with LedgerService to store snapshot after rebuild
> 6. In main.go, configure ledger service: `ledgerService.WithSnapshotRepository(snapshotRepo)`
> 7. Write tests first (TDD)
> 8. Update documentation

**Result:**
- domain/entity/daily_metrics_snapshot.go - DailyMetricsSnapshot entity
- domain/repository/daily_metrics_snapshot_repository.go - Repository interface
- domain/service/metrics_engine.go - MetricsEngine with KPI calculations
- domain/service/metrics_engine_test.go - Comprehensive tests (10 test cases)
- infrastructure/persistence/daily_metrics_snapshot_repository.go - PostgreSQL implementation
- domain/service/ledger_service.go - Added WithSnapshotRepository, stores snapshot after rebuild
- migrations/000006_create_daily_metrics_snapshot_table.up.sql / .down.sql
- Updated TEST_PLAN.md with MetricsEngine test scenarios
- Updated DATABASE_SCHEMA.md with migration 000006
- Updated IMPLEMENTATION_LOG.md with MetricsEngine implementation
- All tests passing (98/98)

### [2026-02-26] Implement AIInsightService
**Original:**
> Implement AIInsightService.
> - Input structured snapshot JSON
> - Output 80–120 word executive brief
> - Gate by plan_tier
> - Store daily_insight

**Improved:**
> Implement AIInsightService for AI-generated daily summaries:
> 1. Create DailyInsight entity in domain/entity
> 2. Create DailyInsightRepository interface and PostgreSQL implementation
> 3. Create AIInsightService in application/service with:
>    - GenerateInsight(userID, appID, snapshot, now) - Generate 80-120 word brief
>    - Uses AIProvider interface for LLM calls (mockable)
>    - Gate by user's plan_tier (PRO only, return ErrProTierRequired for FREE)
> 4. Create migration 000007_create_daily_insight_table
> 5. Write tests first (TDD)
> 6. Update documentation (with Pre-Commit Checklist)

**Result:**
- domain/entity/daily_insight.go - DailyInsight entity
- domain/repository/daily_insight_repository.go - Repository interface
- application/service/ai_insight_service.go - AIInsightService with plan tier gating
- application/service/ai_insight_service_test.go - Tests (5 test cases)
- infrastructure/persistence/daily_insight_repository.go - PostgreSQL implementation
- domain/repository/user_repository.go - Added FindByID method
- infrastructure/persistence/user_repository.go - Added FindByID implementation
- migrations/000007_create_daily_insight_table.up.sql / .down.sql
- Updated TEST_PLAN.md with AIInsightService test scenarios
- Updated DATABASE_SCHEMA.md with migration 000007
- Updated IMPLEMENTATION_LOG.md with AIInsightService implementation
- Updated docs/ER_current.puml with daily_insight entity
- All tests passing (103/103)

### [2026-02-27] Implement NotificationService
**Original:**
> Implement NotificationService.
> - Critical alerts (state change)
> - Daily summary
> - Device token support

**Improved:**
> Implement NotificationService for push notifications:
> 1. Create DeviceToken entity (id, user_id, device_token, platform)
> 2. Create NotificationPreferences entity (critical_enabled, daily_summary_enabled, daily_summary_time)
> 3. Create DeviceTokenRepository and NotificationPreferencesRepository interfaces
> 4. Create NotificationService in application/service with:
>    - SendCriticalAlert(userID, appID, oldState, newState)
>    - SendDailySummary(userID, snapshot)
>    - RegisterDevice(userID, token, platform)
>    - UnregisterDevice(userID, token)
> 5. Create PushNotificationProvider interface for FCM/APNs (mockable)
> 6. Create migrations 000008_create_device_tokens_table, 000009_create_notification_preferences_table
> 7. Respect user preferences before sending
> 8. Write tests first (TDD)
> 9. Update documentation (Pre-Commit Checklist)

**Result:**
- domain/entity/device_token.go - DeviceToken entity with Platform value object
- domain/entity/notification_preferences.go - NotificationPreferences entity
- domain/repository/device_token_repository.go - Repository interface
- domain/repository/notification_preferences_repository.go - Repository interface
- application/service/notification_service.go - NotificationService with all methods
- application/service/notification_service_test.go - Tests (15 test cases)
- infrastructure/persistence/device_token_repository.go - PostgreSQL implementation
- infrastructure/persistence/notification_preferences_repository.go - PostgreSQL implementation
- migrations/000008_create_device_tokens_table.up.sql / .down.sql
- migrations/000009_create_notification_preferences_table.up.sql / .down.sql
- Updated TEST_PLAN.md with NotificationService test scenarios
- Updated DATABASE_SCHEMA.md with migrations 000008, 000009
- Updated IMPLEMENTATION_LOG.md with NotificationService implementation
- Updated docs/ER_current.puml with device_tokens and notification_preferences entities
- All tests passing (109/109)

### [2026-02-27] Implement SlackNotificationProvider
**Original:**
> Implement SlackNotificationProvider

**Improved:**
> Implement SlackNotificationProvider for Slack webhook notifications:
> 1. Create SlackNotifier interface in application/service
> 2. Create SlackNotificationProvider in infrastructure/external
> 3. Implement SendSlack(ctx, webhookURL, title, body, color) method
> 4. Use Slack webhook payload with attachments for rich formatting
> 5. Integrate with NotificationService to send to Slack when webhook is configured
> 6. Write tests first (TDD)
> 7. Update documentation

**Result:**
- infrastructure/external/slack_provider.go - SlackNotificationProvider with SendSlack
- infrastructure/external/slack_provider_test.go - Tests (6 test cases)
- application/service/notification_service.go - Added SlackNotifier interface, WithSlackNotifier builder
- application/service/notification_service_test.go - Added Slack integration tests (5 test cases)
- Updated TEST_PLAN.md with Slack test scenarios
- Updated IMPLEMENTATION_LOG.md with SlackNotificationProvider implementation
- All tests passing (112/112)

### [2026-02-27] Create Marketing Site
**Original:**
> Create a Next.js marketing site for LedgerGuard.
> Sections: Hero section, Problem statement, Renewal Success Rate explanation, Revenue at Risk explanation, AI Daily Revenue Brief section, Pricing tiers, CTA: Connect Shopify Partner
> Use TailwindCSS. Keep minimal professional design. No Firebase.

**Improved:**
> Create a Next.js marketing site for LedgerGuard:
> 1. Create marketing/ folder with REQUIREMENTS.md
> 2. Initialize Next.js 14+ with App Router and TailwindCSS
> 3. Create landing page with sections:
>    - Hero (headline, subheadline, CTA)
>    - Problem statement (Shopify app developer pain points)
>    - Renewal Success Rate explanation
>    - Revenue at Risk explanation
>    - AI Daily Revenue Brief (Pro feature)
>    - Pricing tiers (FREE vs PRO)
>    - Final CTA: Connect Shopify Partner
> 4. Minimal professional design, responsive
> 5. No Firebase
> 6. Update CLAUDE.md with marketing site structure

**Result:**
- Updated CLAUDE.md with marketing site documentation and structure
- marketing/REQUIREMENTS.md - Site requirements, copy, design specs
- marketing/site/ - Next.js 14+ project with TailwindCSS
- Components: Header, Hero, Problem, RenewalRate, RevenueAtRisk, AIBrief, Pricing, FinalCTA, Footer
- SEO metadata configured
- Build successful

### [2026-02-27] Initialize Flutter Web Project
**Original:**
> Initialize Flutter Web project for LedgerGuard. Requirements: Clean Architecture folder structure, Bloc for state management, GoRouter for navigation, Firebase core setup (no auth screens yet), Separate environments (dev/prod ready structure). Do NOT implement UI yet. Just project structure and configuration.

**Improved:**
> Initialize Flutter Web project for LedgerGuard:
> 1. Create frontend/REQUIREMENTS.md with tech stack and architecture specs
> 2. Create frontend/prompts.md for frontend-specific prompt tracking
> 3. Initialize Flutter project with web support (flutter create --platforms web)
> 4. Set up Clean Architecture folders:
>    - lib/core/ (config, constants, theme, utils, di)
>    - lib/data/ (datasources, models, repositories)
>    - lib/domain/ (entities, repositories, usecases)
>    - lib/presentation/ (blocs, pages, widgets, router)
> 5. Add dependencies: flutter_bloc, go_router, get_it, injectable, firebase_core, dio, freezed
> 6. Create environment config (EnvConfig, AppConfig) for dev/prod
> 7. Set up dependency injection with get_it + injectable
> 8. Create GoRouter configuration with placeholder pages
> 9. Create app theme matching marketing site colors
> 10. Update TEST_PLAN.md with frontend test scenarios
> 11. No UI implementation

**Result:**
- frontend/REQUIREMENTS.md - Tech stack, architecture specs, dependencies
- frontend/prompts.md - Frontend prompt tracking
- frontend/app/ - Flutter Web project
- lib/core/config/ - EnvConfig, AppConfig for dev/prod environments
- lib/core/di/ - Dependency injection setup with get_it + injectable
- lib/core/theme/app_theme.dart - App theme matching marketing colors
- lib/presentation/router/app_router.dart - GoRouter configuration
- lib/presentation/pages/placeholder_page.dart - Placeholder for routes
- lib/main.dart, main_dev.dart, main_prod.dart - Entry points
- pubspec.yaml - Dependencies: flutter_bloc, go_router, get_it, firebase_core, dio, freezed, etc.
- test/widget_test.dart - Basic app render test
- Updated TEST_PLAN.md with frontend test scenarios (4 sections, 20+ test cases)
- All tests passing (1/1 Flutter)

### [2026-02-27] Implement Firebase Authentication
**Original:**
> Implement Firebase Authentication integration. Requirements: Email/Password login, Google login, Firebase initialization, Auth state listener, Basic loading state. Create: AuthRepository, AuthController, AuthState. Write widget tests for login logic. Do not build dashboard yet.

**Improved:**
> Implement Firebase Authentication with Clean Architecture and TDD:
> 1. Add dependencies: firebase_auth, google_sign_in
> 2. Create UserEntity in domain/entities
> 3. Create AuthRepository interface in domain/repositories with exceptions
> 4. Create FirebaseAuthRepository in data/repositories implementing Firebase Auth
> 5. Create AuthBloc with events (AuthCheckRequested, SignInWithEmail, SignInWithGoogle, SignOut)
> 6. Create AuthState (AuthInitial, AuthLoading, Authenticated, Unauthenticated, AuthError)
> 7. Register dependencies in injection.config.dart
> 8. Write tests first (TDD) for AuthBloc
> 9. Update TEST_PLAN.md and documentation

**Result:**
- pubspec.yaml - Added firebase_auth, google_sign_in dependencies
- domain/entities/user_entity.dart - UserEntity with Equatable
- domain/repositories/auth_repository.dart - AuthRepository interface + exception classes
- data/repositories/firebase_auth_repository.dart - Firebase implementation
- presentation/blocs/auth/auth_bloc.dart - AuthBloc with all event handlers
- presentation/blocs/auth/auth_event.dart - Auth events
- presentation/blocs/auth/auth_state.dart - Auth states
- presentation/blocs/auth/auth.dart - Barrel export
- core/di/injection.config.dart - Registered AuthRepository and AuthBloc
- test/presentation/blocs/auth_bloc_test.dart - 11 test cases (TDD)
- Updated TEST_PLAN.md with AuthBloc test scenarios
- Updated frontend/prompts.md with prompt entry
- All tests passing (12/12 Flutter)

### [2026-02-27] Create Login and Signup Screens
**Original:**
> Create login and signup screens. Requirements: Email field, Password field, Google login button, Loading state, Error display, Clean minimal UI. Navigation: If logged in → redirect to dashboard route. If not logged in → show login. Write widget tests.

**Improved:**
> Create login and signup screens with auth navigation:
> 1. Create LoginPage with email/password fields, Sign In button, Google Sign In button
> 2. Create SignupPage with email/password fields, Create Account button, Google Sign In button
> 3. Add loading state (CircularProgressIndicator, disabled buttons)
> 4. Add error display (red container with message)
> 5. Update AppRouter with auth-aware redirects using GoRouterRefreshStream
> 6. Update LedgerGuardApp to provide AuthBloc and trigger AuthCheckRequested
> 7. Write widget tests for both pages (TDD)
> 8. Update documentation

**Result:**
- presentation/pages/login_page.dart - Login screen with form, loading, error states
- presentation/pages/signup_page.dart - Signup screen with form, loading, error states
- presentation/router/app_router.dart - Auth redirects with GoRouterRefreshStream
- app.dart - BlocProvider setup, AuthBloc initialization
- test/presentation/pages/login_page_test.dart - 9 test cases
- test/presentation/pages/signup_page_test.dart - 8 test cases
- Updated TEST_PLAN.md with page widget tests
- Updated frontend/IMPLEMENTATION_LOG.md
- Updated frontend/prompts.md
- All tests passing (29/29 Flutter)

### [2026-02-27] KPI Dashboard Upgrade: Time Filtering and Delta Comparison
**Original:**
> (Plan file provided) KPI Dashboard Upgrade with time filtering and delta comparison. Backend: time range value objects, period metrics, aggregation service, API endpoint. Frontend: time range selector, delta indicators on KPI cards.

**Improved:**
> Implement Play Store-style analytics upgrade for KPI dashboard:
> **Backend:**
> 1. Create TimeRangePreset value object and DateRange helpers
> 2. Create PeriodMetrics entity with current, previous, and delta
> 3. Create MetricsAggregationService for period aggregation
> 4. Add GetMetricsByPeriod handler with start/end query params
> 5. Delta calculation with good/bad semantics
> **Frontend:**
> 6. Create TimeRange entity and TimeRangeSelector widget
> 7. Add TimeRangeChanged event to DashboardBloc
> 8. Add MetricsDelta and DeltaIndicator to dashboard_metrics.dart
> 9. Update KpiCard with delta badges (green/red based on semantics)
> 10. Wire TimeRangeSelector to dashboard app bar
> 11. Update tests for new timeRange parameter

**Result:**
- Backend:
  - internal/domain/valueobject/time_range.go - TimeRangePreset, DateRange
  - internal/domain/entity/period_metrics.go - PeriodMetrics, MetricsDelta
  - internal/application/service/metrics_aggregation_service.go + tests
  - internal/interfaces/http/handler/metrics.go - GetMetricsByPeriod
  - internal/interfaces/http/router/router.go - New route
- Frontend:
  - lib/domain/entities/time_range.dart - TimeRange, TimeRangePreset
  - lib/domain/entities/dashboard_metrics.dart - MetricsDelta, DeltaIndicator
  - lib/presentation/widgets/time_range_selector.dart - Time range dropdown
  - lib/presentation/widgets/kpi_card.dart - Delta badges
  - lib/presentation/blocs/dashboard/* - TimeRangeChanged event
  - lib/presentation/pages/dashboard_page.dart - Wired together
  - Tests updated for timeRange parameter
- All backend tests passing (124/124)
- All frontend dashboard tests passing (32/32)

### [2026-02-27] Live FetchTransactions from Shopify Partner API
**Original:**
> is FetchTransactions from live data? → yes, implement it

**Improved:**
> Implement live FetchTransactions in ShopifyPartnerClient:
> 1. Add FetchTransactions method with GraphQL pagination
> 2. Support only Shopify-supported transaction types: AppSubscriptionSale, AppUsageSale, AppOneTimeSale
>    - NOTE: AppCredit, ServiceSale, ReferralTransaction are NOT supported in transactions query
> 3. Add context-based organization ID passing via WithOrganizationID
> 4. Update SyncService to pass organization ID via context
> 5. Wire ShopifyPartnerClient as TransactionFetcher in main.go
> 6. Configure ledger service with snapshot repository: `ledgerService.WithSnapshotRepository(snapshotRepo)`
> 7. Add comprehensive tests for FetchTransactions
> 8. Add debug logging to auth middleware and metrics handler for troubleshooting

**Result:**
- infrastructure/external/shopify_partner_client.go - FetchTransactions, WithOrganizationID
- infrastructure/external/shopify_partner_client_test.go - 6 new tests
- application/service/sync_service.go - Context with organization ID
- cmd/server/main.go - Wired ShopifyPartnerClient + WithSnapshotRepository
- interfaces/http/handler/metrics.go - Added error logging
- interfaces/http/middleware/auth.go - Added token verification error logging
- All backend tests passing (123/123)

### [2026-02-27] Subscription List and Detail Implementation
**Original:**
> Implement subscription list and detail views for backend and frontend

**Improved:**
> Implement subscription list and detail views:
> 1. Backend handler with List and GetByID endpoints
> 2. API: GET /api/v1/apps/{appID}/subscriptions with risk_state filter
> 3. API: GET /api/v1/apps/{appID}/subscriptions/{subscriptionID}
> 4. Frontend Subscription entity and repository
> 5. SubscriptionListBloc and SubscriptionDetailBloc
> 6. Subscription list page with filter dropdown
> 7. Subscription detail page with risk badge
> 8. RiskBadge and SubscriptionTile widgets

**Result:**
- Backend subscription handler + tests
- Frontend subscription pages and blocs
- All tests passing

### [2026-02-28] Revenue API Documentation Site
**Original:**
> this is need some documention site for my client. right? → what feels premium. since this is paid service → lets go with mintlify first and 2 for backup → now my own version → document stepts to deploy it to vercel

**Improved:**
> Create premium API documentation for Revenue API:
> 1. Create Mintlify documentation site with OpenAPI spec
> 2. Create custom Next.js documentation site as backup
> 3. Include all REST and GraphQL endpoints
> 4. Add code examples in cURL, Node.js, Python
> 5. Document authentication, error codes, rate limits
> 6. Add Vercel deployment instructions

**Result:**
- Mintlify docs: `docs/api/` with mint.json, openapi.yaml, MDX pages
- Custom Next.js: `docs/site/` with 24 pages
- DEPLOYMENT.md for Vercel deployment
- Commits: 33686f6, bc4f61e, d1b8189, 100f3ef, 200af3c

### [2026-02-27] Shop Name, Gross Amount, and Period-Based Usage Revenue
**Original:**
> - show shop.name in subscription list instead of domain
> - why usage charge is 0
> - usage charge is same for all filters

**Improved:**
> Fix transaction data quality and period-based metrics:
> 1. Add shop_name to transactions and subscriptions (display name, not domain)
> 2. Add gross_amount_cents to transactions (subscription price pre-Shopify cut)
> 3. Add __typename to GraphQL query for proper charge type inference (USAGE vs RECURRING)
> 4. Fix MetricsAggregationService to calculate revenue from transactions for specific date range
> 5. Fix frontend subscription_tile.dart index out of range errors
> 6. Add migrations 000010 and 000011

**Result:**
- internal/domain/entity/transaction.go - ShopName, GrossAmountCents, NetAmountCents
- internal/domain/entity/subscription.go - ShopName
- internal/infrastructure/external/shopify_partner_client.go - __typename, shop.name, grossAmount
- internal/application/service/metrics_aggregation_service.go - Calculate from transactions
- migrations/000010, 000011 - New columns
- frontend/app/lib/presentation/widgets/subscription_tile.dart - Defensive string handling
- All tests passing (124 backend)

### [2026-02-27] Revenue API Implementation (REST + GraphQL)
**Original:**
> Implement external Revenue API for Shopify app developers to query subscription payment status

**Improved:**
> Implement Revenue API with REST and GraphQL endpoints for external clients:
>
> **Database (4 migrations):**
> 1. `000012_create_api_keys_table` - API keys with SHA-256 hash storage
> 2. `000013_create_api_subscription_status_table` - CQRS read model
> 3. `000014_create_api_usage_status_table` - Usage billing status
> 4. `000015_create_api_audit_log_table` - Request audit logging
>
> **Domain Layer (`internal/revenue_api/domain/`):**
> 5. `APIKey` entity with NewAPIKey(), HashKey() using SHA-256
> 6. `SubscriptionStatus` read model with risk state, payment status
> 7. `UsageStatus` with parent subscription reference
> 8. `AuditLog` request audit entry
> 9. Repository interfaces for all entities
>
> **Infrastructure Layer:**
> 10. PostgreSQL implementations for all repositories
> 11. Async audit logging with background goroutine
>
> **Application Layer:**
> 12. `APIKeyService` - Create, List, Revoke, ValidateKey
> 13. `SubscriptionStatusService` - GetByShopifyGID, GetByDomain, batch
> 14. `UsageStatusService` - GetByShopifyGID, batch
> 15. `RevenueReadModelBuilder` - Rebuilds read model from ledger
>
> **HTTP Layer:**
> 16. `APIKeyAuth` middleware - X-API-Key header validation
> 17. `RateLimiter` middleware - In-memory token bucket
> 18. `AuditLogger` middleware - Async request logging
> 19. `APIKeyHandler` - POST/GET/DELETE /api-keys
> 20. `SubscriptionStatusHandler` - REST endpoints
> 21. `UsageStatusHandler` - REST endpoints
>
> **GraphQL Layer:**
> 22. `schema.graphql` with Query type, SubscriptionStatus, UsageStatus types
> 23. `resolver.go` - Root resolver with enums
> 24. `schema.resolvers.go` - Query resolvers
> 25. `handler.go` - HTTP handler for /graphql endpoint
>
> **Router:**
> 26. Separate router for Revenue API at `/v1/`
> 27. API key management routes (Firebase auth protected)
> 28. Public API routes (API key auth protected)
>
> **API Endpoints:**
> - `POST /v1/api-keys` - Create new API key
> - `GET /v1/api-keys` - List user's API keys
> - `DELETE /v1/api-keys/{keyID}` - Revoke API key
> - `GET /v1/subscriptions/{shopify_gid}` - Get subscription by GID
> - `GET /v1/subscriptions/by-domain?domain={domain}` - Get by domain
> - `POST /v1/subscriptions/batch` - Batch lookup (max 100)
> - `GET /v1/usage/{shopify_gid}` - Get usage by GID
> - `POST /v1/usage/batch` - Batch lookup (max 100)
> - `POST /v1/graphql` - GraphQL endpoint

**Result:**
- 4 migrations (000012-000015)
- 4 domain entities + 4 repository interfaces
- 4 PostgreSQL repository implementations
- 4 application services
- 3 HTTP middleware (APIKeyAuth, RateLimiter, AuditLogger)
- 3 HTTP handlers (APIKey, SubscriptionStatus, UsageStatus)
- GraphQL schema + resolvers + handler
- Revenue API router
- Files in `internal/revenue_api/` directory

### [2026-02-28] API Key Management Backend Integration
**Original:**
> Integrate Revenue API's API key management endpoints into main router

**Improved:**
> Integrate API key handler from `internal/revenue_api/` into main backend router:
>
> **Router Integration:**
> 1. Import `apikeyhandler` package from `internal/revenue_api/interfaces/http/handler`
> 2. Add `APIKeyHandler *apikeyhandler.APIKeyHandler` to router Config struct
> 3. Add `/api/v1/api-keys` routes with Firebase auth middleware:
>    - `GET /api/v1/api-keys` - List user's API keys
>    - `POST /api/v1/api-keys` - Create new API key
>    - `DELETE /api/v1/api-keys/{id}` - Revoke API key
>
> **Main.go Integration:**
> 4. Import API key service and repository packages
> 5. Initialize PostgresAPIKeyRepository with db.Pool
> 6. Initialize APIKeyService with repository
> 7. Initialize APIKeyHandler with service
> 8. Add apiKeyHandler to router config
>
> **Handler Updates:**
> 9. Update Create response to match frontend format:
>    - Return `api_key` object with id, name, key_prefix, created_at, last_used_at
>    - Return `full_key` with the one-time visible raw key
> 10. Update List response to return formatted APIKeyResponse array

**Result:**
- cmd/server/main.go - Import and initialize API key handler
- internal/interfaces/http/router/router.go - Add APIKeyHandler to Config, add routes
- internal/revenue_api/interfaces/http/handler/api_key_handler.go - Updated response format
- Endpoint: GET/POST/DELETE /api/v1/api-keys

### [2026-02-28] API Key Management Frontend
**Original:**
> Create Flutter screens for managing API keys for Revenue API access

**Improved:**
> Implement API Key Management frontend screens using Clean Architecture + BLoC:
>
> **Domain Layer:**
> 1. `ApiKey` entity with id, name, keyPrefix, createdAt, lastUsedAt
> 2. `ApiKeyCreationResult` for returning full key (shown only once after creation)
> 3. `ApiKeyRepository` interface with getApiKeys, createApiKey, revokeApiKey
> 4. Exception classes: ApiKeyException, ApiKeyLimitException, ApiKeyNotFoundException, ApiKeyUnauthorizedException
>
> **Data Layer:**
> 5. `ApiApiKeyRepository` - API implementation calling `/api/v1/api-keys` endpoints
> 6. Uses Dio with Bearer token authentication
> 7. Handles error responses with proper exception mapping
>
> **Presentation Layer (BLoC):**
> 8. `ApiKeyBloc` - Manages API key state
> 9. Events: LoadApiKeysRequested, CreateApiKeyRequested, RevokeApiKeyRequested, DismissKeyCreatedRequested
> 10. States: ApiKeyInitial, ApiKeyLoading, ApiKeyLoaded, ApiKeyCreated, ApiKeyEmpty, ApiKeyError
>
> **Widgets:**
> 11. `ApiKeyTile` - Card showing key name, masked prefix, created date, last used, revoke button
>
> **Pages:**
> 12. `ApiKeyListPage` - List of API keys with:
>     - Create button in app bar
>     - Create dialog with name validation
>     - Key created dialog shows full key once with copy button and warning
>     - Revoke confirmation dialog with warning
>     - Pull-to-refresh functionality
>     - Empty state with create button
>     - Error state with retry button
>
> **Integration:**
> 13. Add route `/settings/api-keys` to app_router.dart
> 14. Register ApiKeyRepository and ApiKeyBloc in injection.config.dart
> 15. Add "API Keys" navigation tile to Profile page Settings section
>
> **Tests (TDD):**
> 16. ApiKeyBloc tests (14 tests): initial state, load, create, revoke, dismiss

**Result:**
- frontend/app/lib/domain/entities/api_key.dart
- frontend/app/lib/domain/repositories/api_key_repository.dart
- frontend/app/lib/data/repositories/api_api_key_repository.dart
- frontend/app/lib/presentation/blocs/api_key/ (bloc, events, states, barrel)
- frontend/app/lib/presentation/widgets/api_key_tile.dart
- frontend/app/lib/presentation/pages/api_key_list_page.dart
- frontend/app/lib/presentation/pages/profile_page.dart (modified)
- frontend/app/lib/presentation/router/app_router.dart (modified)
- frontend/app/lib/core/di/injection.config.dart (modified)
- test/presentation/blocs/api_key_bloc_test.dart
- 14 passing tests

---

## [2026-03-01] Dashboard Preferences Not Applied

**Original Prompt:**
> dashboard not considering the preferences

**Improved Prompt:**
> Fix dashboard preferences not being applied to the dashboard display:
> 1. Dashboard page hardcodes KPIs and secondary widgets instead of reading user preferences
> 2. PreferencesBloc saves/loads correctly but DashboardPage ignores the state
> 3. Fix DashboardPage to:
>    - Load preferences on init
>    - Use BlocBuilder<PreferencesBloc, PreferencesState> to get current preferences
>    - Filter/order primary KPIs based on preferences.primaryKpis
>    - Filter secondary widgets based on preferences.enabledSecondaryWidgets
> 4. Fix go.mod invalid version (1.25.4 → 1.23)

**Result:**
- frontend/app/lib/presentation/pages/dashboard_page.dart - Major refactor:
  - Converted to StatefulWidget to load preferences on init
  - Added BlocBuilder<PreferencesBloc> wrapping dashboard body
  - New `_buildKpiCard()` method maps KpiType to KpiCard widget
  - `_buildPrimaryKpis()` now iterates preferences.primaryKpis
  - `_buildSecondarySection()` now checks preferences.isSecondaryWidgetEnabled()
  - `_hasSecondaryWidgets()` helper to conditionally show section
- backend/go.mod - Fixed invalid Go version 1.25.4 → 1.23
- Flutter analyze: Only pre-existing deprecation warning (not from this change)

---

## [2026-03-01] Revenue Share Tier Tracking Implementation

**Original Prompt:**
> is the new finding from shoify helpfult to improve any areas in the app we currently build. find anything by review and prepare stepby step backend/fronend changes promt and implement one by one. mainly logics and all

**Improved Prompt:**
> Review the Shopify Partner API documentation findings (shopify-money-flow-diagram.md) and identify gaps between the actual Shopify fee structure and our current implementation. Prepare a phased implementation plan with backend and frontend changes. Focus on:
> 1. Revenue share tiers (0%, 15%, 20%) instead of hardcoded 20%
> 2. Processing fee (2.9%) always applies
> 3. Tax calculation on Shopify's fees, not gross revenue
> 4. Earnings availability timeline (7-37 days)

**Implementation:**
- Phase 1: Revenue Share Tier Tracking (COMPLETED)
  - Backend: RevenueShareTier value object, FeeVerificationService, FeeHandler, migration 000017
  - Frontend: TierSelector, FeeInsightsCard, AppSettingsPage, updated repository/bloc

**Follow-up Prompts:**
1. "yes proceed with phase 1" - Confirmed starting Phase 1
2. "yes continue with frontend" - Confirmed frontend implementation
3. "commit and push" - Committed changes
4. "run the tests" - Ran Go and Flutter tests
5. "fix the failing profile page tests" - Fixed 10 failing tests

---

## [2026-03-01] Shopify Money Flow Visualization

**Original Prompt:**
> (Based on shopify-money-flow-diagram.md prompt) Create animated visualization showing Shopify Partner fee structure

**Improved Prompt:**
> Create an animated money flow visualization for the marketing site showing:
> 1. Revenue flow: Customer → Shopify → Developer (with fee breakdowns)
> 2. Recurring subscription flow with revenue share tiers (0%, 15%, 20%)
> 3. Usage-based billing flow with different charge types
> 4. Processing fees (2.9%) and tax calculations
> 5. Interactive toggles for subscription tiers and usage types
> 6. Real-world example calculations
> 7. Route at /money-flow

**Prompt File:** `docs/prompts/shopify-money-flow-diagram.md`

**Result:**
- marketing/site/app/money-flow/page.tsx
- Animated flow diagram with entity boxes and flowing currency symbols
- Revenue share tier selector (Default 20%, Small Dev 0%, Large Dev 15%)
- Usage type selector (Orders, SMS, API, AI, Storage)
- Fee breakdown visualization
- Commit: f46a758 feat(marketing): expand usage charge types in money-flow visualization

---

## [2026-03-01] KPI Metrics Visualization Component

**Original Prompt:**
> implement the kpi visualization component

**Improved Prompt:**
> Implement the KPIMetricsGuide React component for the marketing site based on the prompt document at docs/prompts/kpi-metrics-visualization.md. The component should:
> 1. Display 6 KPI cards with current values, previous values, and period-over-period deltas
> 2. Three view modes: Overview (formula + timeline), Detail (data flow + subscription list), Comparison (period vs period)
> 3. Animated risk classification timeline (30/60/90 day thresholds)
> 4. Risk distribution bar chart visualization
> 5. Data flow animation showing Partner API → Ledger Rebuild → Metrics Engine → Dashboard
> 6. Semantic delta coloring (higher is good vs lower is good)
> 7. Interactive KPI selection and animation controls (play/pause/restart)
> 8. Page route at /kpi-guide
> 9. Commit and push changes

**Prompt File:** `docs/prompts/kpi-metrics-visualization.md`

**Result:**
- marketing/site/components/KPIMetricsGuide.tsx (1264 lines)
- marketing/site/app/kpi-guide/page.tsx
- Build verified: `npm run build` successful
- Commit: f91d064 feat(marketing): add KPI metrics visualization component
- Pushed to origin/main

---

## [2026-03-01] API Integration Guide Visualization

**Original Prompt:**
> Create API integration guide visualization similar to KPI guide

**Improved Prompt:**
> Create an animated API Integration Guide visualization showing:
> 1. Data flow: Your App → LedgerGuard API → Response
> 2. Integration patterns: Checkout, Dashboard, Alerting, Feature Gating
> 3. Code snippets in JavaScript, cURL, Python with tabs
> 4. Single vs Batch request toggle with separate examples
> 5. Risk state simulator with sample store data
> 6. Request/response animation
> 7. Route at /api-guide

**Prompt File:** `docs/prompts/ledgerguard-api-integration.md`

**Result:**
- marketing/site/app/api-guide/page.tsx
- marketing/site/components/APIIntegrationGuide.tsx
- Interactive flow diagram with integration pattern selector
- Code snippets with language tabs
- Risk state simulator
- Commit: 4043c8e feat(marketing): add API Integration Guide and visualization prompt template

---

## [2026-03-01] Customer Pitch UI

**Original Prompt:**
> o. now based on visualization prompt. i want a customer pitch ui. like before keep separate promt. ask me questions to ask shopify.

**Improved Prompt:**
> Create an interactive, scroll-based customer pitch page for LedgerGuard targeting Shopify app developers ($10K-$100K MRR). Include:
> 1. Hero section with animated dashboard preview (MRR counter, risk alerts)
> 2. Problem section highlighting: Blind to Churn, Spreadsheet Hell, Wrong Tools
> 3. Interactive solution preview with tabbed dashboard (MRR Health, Risk Radar, Cohort Retention)
> 4. How It Works flow diagram (Connect → Sync → Analyze → Act)
> 5. Features grid (6 features: Renewal Rate, Revenue at Risk, Churn Prediction, Usage Revenue, AI Brief, Revenue API)
> 6. Before/After comparison table
> 7. Social proof section with stats and testimonial placeholder
> 8. Pricing preview (Free vs Pro $49/mo)
> 9. Final CTA with trust badges
> 10. Create prompt document at docs/prompts/customer-pitch-ui.md
> 11. Route at /pitch

**Result:**
- docs/prompts/customer-pitch-ui.md - Detailed implementation spec
- marketing/site/components/CustomerPitch.tsx (~700 lines) - Full pitch page component with 9 sections
- marketing/site/app/pitch/page.tsx - Page route with SEO metadata
- Build verified: `npm run build` successful
- All sections: Hero, Problem, Solution Preview, How It Works, Features, Comparison, Social Proof, Pricing, Final CTA

---

## [2026-03-01] Internal Architecture Flow Visualization

**Original Prompt:**
> but need another money-flow like page for internal purpoese with these details from shopify and my app. animated like visual

**Improved Prompt:**
> Create an animated internal architecture visualization page showing LedgerGuard's data flow:
> 1. Data Ingestion: Shopify Partner Account → Partner API → Sync Engine
> 2. Data Processing: Transaction Repository → Ledger Rebuild → Subscription Builder
> 3. Risk Classification: Days overdue → Risk state (Safe/1-Cycle/2-Cycles/Churned)
> 4. Metrics Computation: Renewal Rate, Active MRR, At Risk, Usage Revenue
> 5. Output Layer: Dashboard UI, Push Alerts, AI Daily Brief, Revenue API
> 6. Interactive controls: View modes (Full/Ingestion/Processing/Risk/Output)
> 7. Animation controls: Play/Pause, Speed (0.5x/1x/2x), Show details toggle
> 8. Code snippets: GraphQL query, Go risk engine, metrics computation, SQL snapshot
> 9. Phase-by-phase animation with auto-advance
> 10. Route at /architecture (noindex for internal use)

**Result:**
- docs/prompts/internal-architecture-flow.md - Implementation spec
- marketing/site/components/ArchitectureFlow.tsx (~850 lines)
- marketing/site/app/architecture/page.tsx - Page route with noindex
- Features:
  - 5 pipeline stages with animated entity boxes
  - Risk timeline visualization with moving day marker
  - Animated number counters for metrics
  - Code snippets for GraphQL, Go, SQL
  - View mode switching and playback controls
- Build verified: `npm run build` successful

---

## [2026-03-01] Comprehensive App Review

**Original Prompt:**
> from the answers from shopify's and all. review my app entirely and suggest things like must have and wrong implementations and prompt improvements, promt orderdering, puml corrections. improve this prompt before implementation and save it in prompts for implementation

**Improved Prompt:**
> Conduct a comprehensive review of LedgerGuard against Shopify Partner API documentation and best practices:
> 1. Identify incorrect implementations (risk classification, MRR calculation, fee structure)
> 2. Find missing must-have features (webhooks, soft deletes, audit logging)
> 3. Review architecture gaps (CQRS read models, caching, error handling)
> 4. Check PUML diagram accuracy (ER, Sequence, C4)
> 5. Prioritize issues by severity (P0-Critical, P1-High, P2-Medium, P3-Low)
> 6. Create phased implementation plan with prompt ordering
> 7. Save comprehensive findings as implementation prompt document

**Result:**
- docs/prompts/comprehensive-review-fixes.md - Detailed implementation prompt with:
  - 7 P0-Critical issues (risk classification, synthetic GIDs, MRR normalization, revenue share, rate limiting, test coverage, GraphQL fields)
  - 6 P1-High issues (webhooks, soft deletes, ER diagrams, manual charges, lifecycle events, payment failures)
  - 7 P2-Medium issues (mobile responsiveness, caching, API docs, sequence diagrams, audit logging, error boundaries, data export)
  - 4 P3-Low issues (dark mode, localization, observability, CI/CD)
  - Phased implementation plan (5 phases over 9+ weeks)
  - PUML correction specifications
  - Validation checklist

---

## [2026-03-01] Phase 1: Data Integrity Fixes

**Original Prompt:**
> start with phase 1

**Improved Prompt:**
> Implement Phase 1 (Data Integrity) from comprehensive-review-fixes.md:
> 1. P0-1: Fix Risk Classification Logic - Add SubscriptionStatus value object, update RiskEngine to handle CANCELLED/FROZEN/EXPIRED/PENDING statuses
> 2. P0-3: Verify MRR Annual Normalization - Check Subscription.MRRCents() divides annual by 12
> 3. P0-7: Update GraphQL Query Fields - Add shop.id, shop.plan, appSubscription details, appUsageRecord details
> 4. P0-6: Add Critical Domain Service Tests - Boundary tests (30/60/90 days), status-based tests
> 5. Create migration for new transaction columns
> 6. Update transaction repository with new fields

**Result:**
- Created SubscriptionStatus value object (ACTIVE, CANCELLED, FROZEN, EXPIRED, PENDING)
- Updated RiskEngine with status-first classification logic
- Added 10+ boundary tests for risk classification
- Added 3 status-override tests
- Updated GraphQL query with shop.id, shop.plan, appSubscription, appUsageRecord fields
- Updated Transaction entity with ShopifyShopGID, ShopPlan, SubscriptionGID, SubscriptionStatus, SubscriptionPeriodEnd, BillingInterval
- Created migration 000020_add_transaction_subscription_details
- Updated transaction repository (Upsert, UpsertBatch, Find methods)
- Verified MRR annual normalization already implemented
- All 55+ domain service tests passing

---

## [2026-03-02] Notification Engine Visualization Page

**Original Prompt:**
> the notification engine flow needed. check it can included in any of these. or implement separately with refering propt PROMPT-interactive-visualization-chat.md

**Improved Prompt:**
> Check if notification engine flow visualization can be included in existing marketing site pages (kpi-guide, money-flow, architecture, api-guide, affiliate-program). If not suitable for existing pages, create a dedicated /notifications visualization page following the pattern in PROMPT-interactive-visualization-chat.md with:
> 1. Interactive flow diagrams for: Critical Alert flow, Daily Summary flow, Device Registration flow, Multi-channel delivery
> 2. Animated step-by-step visualization with play/pause controls
> 3. Risk state reference cards (SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED)
> 4. Notification types documentation (Critical Alert, Daily Summary, Billing Failure, App Uninstalled)
> 5. Architecture overview (WebhookService, NotificationService, NotificationScheduler, Push Providers)
> 6. Webhook events documentation with examples
> 7. API endpoints reference
> 8. User preferences section

**Prompt File:** `docs/prompts/notification-engine-flow.md`

**Result:**
- Determined notification flow is distinct topic, needs dedicated page
- Created `/marketing/site/components/NotificationFlowVisualization.tsx` - Animated flow diagram with 4 flow types
- Created `/marketing/site/app/notifications/page.tsx` - Full documentation page
- Features: Flow selector, animated SVG diagram, step details, risk states reference, notification types, architecture cards, webhook events, API endpoints

---

## [2026-03-02] Affiliate Program Flow Visualization

**Original Prompt:**
> Create visualization for affiliate program flow

**Improved Prompt:**
> Create an interactive animated visualization showing common affiliate and referral program patterns:
> 1. Attribution Flow: Cookie-based tracking with configurable windows (30/60/90/lifetime)
> 2. Commission Calculation: One-time vs recurring vs hybrid models with comparisons
> 3. Multi-Tier Structure: Direct referral and sub-affiliate override commissions
> 4. Affiliate Lifecycle: Application → Approval → Share → Earn → Tier progression
> 5. Real-world examples from Shopify, ConvertKit, HubSpot, Webflow
> 6. Program types: Referral links, Coupon codes, Partner programs
> 7. Implementation considerations: Attribution window, fraud prevention, payout logistics, platform options
> 8. Route at /affiliate-program

**Prompt File:** `docs/prompts/affiliate-program-flow.md`

**Result:**
- Created `/marketing/site/components/AffiliateFlowVisualization.tsx` - Animated flow visualization
- Created `/marketing/site/app/affiliate-program/page.tsx` - Full documentation page
- Features: 4 flow types (Attribution, Commission, Multi-Tier, Lifecycle), animated SVG diagrams, real-world examples, implementation considerations
- Commit: 5716527 feat: add affiliate program flow visualization page

---

## [2026-03-02] Per-Partner Shopify API Rate Limiting

**Original Prompt:**
> implement rate limit for api call using partner token. keep it configurable and default is 3/sec. so it should be app level because partner may have different app

**Improved Prompt:**
> Implement per-partner rate limiting for outgoing Shopify Partner API calls:
> 1. Scope: Per-partner-account level (keyed by organization ID), since Shopify rate limits are per access token
> 2. Default: 3 requests/second, configurable via `SHOPIFY_RATE_LIMIT_RPS` environment variable
> 3. Implementation: Use token bucket rate limiter with per-partner limiter map
> 4. Storage: Map of limiters keyed by organization ID in ShopifyPartnerClient
> 5. Behavior: Wait/block when limit reached, log when throttling occurs
> 6. Thread-safe: RWMutex for concurrent access to limiter map
> 7. Apply to: All Partner API GraphQL calls (FetchTransactions, FetchApps, FetchAppEvents)

**Result:**
- Updated `internal/infrastructure/config/config.go` - Added RateLimitRPS field with env override
- Updated `internal/infrastructure/external/shopify_partner_client.go` - Per-partner rate limiter map, WithRequestsPerSecond option
- Updated `cmd/server/main.go` - Wired rate limit config to partner client
- Each partner gets their own token bucket limiter (prevents one partner's sync from affecting another)
- Commit: 2abf64c feat: implement per-partner rate limiting for Shopify Partner API

---

## [2026-03-02] Voice AI Assistant Visualization Page

**Original Prompt:**
> in future i want an option in app to ask voice like want details of store xxxx. so ai with handle this and navigate to the subscription details/health page. or ask list of subscription with risk. so it naviages to list with rist tab is cliced. so prepare o visualizaton page. so it will update daily until implementation starts

**Improved Prompt:**
> Create an interactive visualization page for the Voice AI Assistant feature (future implementation):
> 1. Voice Commands Flow: Show how voice input is processed (Capture → Transcript → Intent → Navigate)
> 2. Supported Intents: STORE_DETAILS, STORE_HEALTH, LIST_FILTER, METRIC_QUERY, ALERT_QUERY, NAVIGATE
> 3. Demo flow with selectable example commands
> 4. Architecture overview: Flutter speech_to_text → Claude API → Entity resolver → GoRouter
> 5. Fallback behavior: Show suggestions if intent unclear (confidence < 0.7)
> 6. Route at `/voice-assistant` with noindex (future feature spec)

**Prompt File:** `docs/prompts/voice-assistant-flow.md`

**Result:**
- Created `docs/prompts/voice-assistant-flow.md` - Full specification with Flutter code examples
- Created `marketing/site/components/VoiceAssistantVisualization.tsx` - Interactive demo with stage progression
- Created `marketing/site/app/voice-assistant/page.tsx` - Page with use cases, roadmap, tech stack
- Features: Voice waveform animation, typing effect transcript, intent/entity cards, fallback suggestions
- Fallback: Text response with suggested commands when intent unclear

---

## [2026-03-03] Hetzner Infrastructure Visualization Page

**Original Prompt:**
> need a visualization page of how hetzner operates. from enduser perspective and hetzner perspective. their farm and all.

**Improved Prompt:**
> Create an interactive visualization page showing how Hetzner operates as a company:
> 1. End-User Journey flow: Sign Up → Cloud Console → Order Server → Configure → Deploy
> 2. Data Center Operations flow: Procure → Assemble → Rack & Stack → Network → Provision → Live
> 3. Network Architecture flow: End User → IX/Peering → Backbone → DC Router → Server
> 4. Server Lifecycle flow: Ordered → Provisioned → Active → Maintained → Decommission → Recycle
> 5. Static sections: DC locations (6 sites), differentiators, product lineup, network infra, server auction, two-perspectives comparison
> 6. Route at `/hetzner-infrastructure`, orange/red color theme (distinct from existing `/deployment` page)

**Prompt File:** `docs/prompts/hetzner-infrastructure-visualization.md`

**Result:**
- Created `docs/prompts/hetzner-infrastructure-visualization.md` - Full specification
- Created `marketing/site/components/HetznerInfrastructureVisualization.tsx` - Interactive 4-flow SVG animation
- Created `marketing/site/app/hetzner-infrastructure/page.tsx` - Full page with DC locations, products, network, auction, perspectives
- Features: 4 animated flows, 6 DC location cards, product lineup, network layers, server auction explainer, user vs Hetzner perspective rows

---

## [2026-03-03] GCP Staging Infrastructure + Visualization Page

**Original Prompt:**
> GCP provides free credit. proceed with that (staging env + visualization)

**Improved Prompt:**
> Set up GCP Cloud Run as a staging environment using free credits, alongside the existing Hetzner production:
> 1. Terraform configs (deploy/gcp/): Cloud Run, Cloud SQL, VPC, Artifact Registry, Secret Manager
> 2. CI/CD: staging branch → Docker build → Cloud Run deploy (alongside main → Hetzner)
> 3. Scripts: gcp-setup.sh (one-time), gcp-deploy.sh (manual deploy)
> 4. Visualization page at /gcp-staging with 4 flows: Dual Architecture, GCP Topology, CI/CD Branching, Request Flow
> 5. Static sections: GCP services, environment comparison, CI/CD strategy, secrets, cost comparison
> 6. Update Makefile, .gitignore, DECISIONS.md, TAD.md, future.md

**Prompt File:** `docs/prompts/gcp-staging-visualization.md`

**Result:**
- Created `deploy/gcp/` — 9 Terraform files (main, variables, network, database, cloudrun, secrets, artifact_registry, outputs, tfvars.example)
- Created `scripts/gcp-setup.sh` — One-time GCP project setup
- Created `scripts/gcp-deploy.sh` — Manual deployment script
- Updated `.github/workflows/deploy.yml` — Added staging branch + deploy-staging job
- Updated `.gitignore` — Terraform state, tfvars, GCP credentials
- Updated `Makefile` — Added gcp-setup, gcp-deploy, gcp-logs targets
- Created `marketing/site/components/GCPStagingVisualization.tsx` — 4-flow interactive SVG component
- Created `marketing/site/app/gcp-staging/page.tsx` — Full page with services, comparison, CI/CD, costs
- Created `docs/prompts/gcp-staging-visualization.md` — Prompt specification
- Updated `DECISIONS.md` — ADR-009: GCP Cloud Run for Staging
- Updated `TAD.md` — Added Deployment Environments section
- Updated `future.md` — Deferred: custom staging domain, marketing site on Cloud Run

---

### [2026-03-04] Deploy Flutter Frontend to Firebase Hosting
**Original:**
> frond end deploy

**Improved:**
> Deploy the Flutter web frontend to Firebase Hosting. Set up production entry point (main_prod.dart with Firebase init), update web branding (title, manifest), configure Firebase Hosting (firebase.json + .firebaserc), add CI/CD (flutter build web in CI, deploy-frontend job in deploy.yml using FirebaseExtended/action-hosting-deploy). Choose Firebase Hosting over Vercel ($0 vs $20/mo commercial) and Cloud Run (overkill for static files).

**Result:**
- Updated `main_prod.dart` with Firebase initialization
- Updated `web/index.html` and `manifest.json` with LedgerGuard branding
- Created `.firebaserc`, added hosting to `firebase.json`
- Added web build to CI, frontend deploy job to deploy workflow
- Build verified: `flutter build web --release -t lib/main_prod.dart`

---

### [2026-03-05] Staging Environment End-to-End Setup
**Original:**
> Connect the Flutter web frontend (Firebase Hosting) to the Go backend (GCP Cloud Run) for staging. Fix all issues needed for end-to-end connectivity: staging environment config, CORS, Docker builds, database migrations, Firebase Auth.

**Improved:**
> Set up staging environment end-to-end connectivity:
> 1. Add `Environment.staging` to Flutter EnvConfig pointing to Cloud Run API
> 2. Create `main_staging.dart` entry point for Firebase Hosting builds
> 3. Fix `apiBaseUrl` getter — staging falling through to localhost
> 4. Fix `gcp-deploy.sh` — add `--platform linux/amd64` for Apple Silicon
> 5. Add Firebase Hosting domains to CORS AllowedOrigins in router.go
> 6. Fix dirty database migration 17, run migrations 18-27
> 7. Enable Firebase Auth Email/Password provider, update API key restrictions (Identity Toolkit + Token Service APIs)
> 8. Change deploy workflow from `main_prod.dart` to `main_staging.dart`

**Result:**
- Created `main_staging.dart`, added `Environment.staging` to EnvConfig
- Fixed `apiBaseUrl` getter for staging
- Added `--platform linux/amd64` to gcp-deploy.sh
- Added Firebase Hosting domains to backend CORS
- Fixed dirty migration 17, ran migrations 18-27
- Enabled Firebase Auth Email/Password, updated API key restrictions
- Changed deploy workflow to use staging entry point

---


### [2026-03-07] Create Excalidraw Architecture Diagrams
**Original:**
> Implement the plan to create Excalidraw diagrams for LedgerGuard

**Improved:**
> Create 7 Excalidraw JSON diagram files in docs/diagrams/ for visual architectural documentation: system architecture (C4-style), auth & onboarding flow, sync & ledger rebuild pipeline, risk engine decision tree, database ER diagram, frontend screen flow, and snapshot backfill lifecycle. Use consistent color coding (blue=backend, green=frontend, orange=external, purple=database, red/yellow=risk states).

**Result:**
- Created docs/diagrams/ directory with 7 .excalidraw files
- All files valid JSON, openable in Excalidraw or VS Code extension
- Consistent color scheme across all diagrams

---

### [2026-03-07] Add API Interface Diagram
**Original:**
> update about graphql — i mean api interface from this app

**Improved:**
> Create an Excalidraw diagram documenting all REST API endpoints exposed by the LedgerGuard Go backend, organized by resource group (Auth, Apps, Subscriptions, Revenue, Sync, etc.), showing middleware stack and endpoint categories (authenticated, external, internal).

**Result:**
- Created `docs/diagrams/api-interface.excalidraw` (38 elements)
- 12 endpoint group cards covering all routes from router.go
- Shows middleware stack (CORS, Firebase Auth, Admin, Logging)
- Color-coded: blue=authenticated, orange=external, gray=internal

---

### [2026-03-07] AI Chat + GraphQL Developer API (Future Feature)
**Original:**
> also thinking to add a chat window (ai) as anything. So ai tool can introduce so query anything if graphql. suggest anything. improve the prompt. visualization and diagram first. add it to future

**Improved:**
> Add two synergistic future features: (1) GraphQL Developer API (P2) — gqlgen-based schema-driven API for external Shopify app developers, replacing/supplementing REST Revenue API, authenticated via API keys. (2) AI Chat Assistant (P3) — Flutter chat widget where developers ask natural language questions, Claude API translates to GraphQL queries, returns conversational responses with data tables and follow-ups. Create prompt file, Excalidraw diagram, and detailed future.md entries.

**Result:**
- Created `docs/prompts/ai-graphql-chat-flow.md` (full spec with schema, examples, pipeline)
- Created `docs/diagrams/ai-graphql-chat.excalidraw` (53 elements)
- Updated `future.md` with GraphQL API (P2) and AI Chat (P3) detailed specs

---

### [2026-03-07] Adapt Chat Builder System Prompt for LedgerGuard
**Original:**
> Refer to OpenAI Explorer's chat-builder-system-prompt.md for AI chat. Convert it for this app removing template and broadcast module. Improve the prompt.

**Improved:**
> Adapt the Chat Builder System Prompt from OpenAI Explorer for LedgerGuard's AI Chat Assistant. Keep core architecture patterns (module interface, registry, tool call loop, state extraction, context-aware input, dynamic system prompt). Remove WhatsApp-specific modules (Templates, Broadcast, CSV, Catalog). Replace with LedgerGuard domain modules: Subscriptions (4 tools), Metrics (3), Risk (3), Store Health (2), Earnings (2), Sync (2). Swap OpenAI for Claude API tool use. Adapt frontend from React to Flutter Bloc. Use internal GraphQL (gqlgen) as query execution layer. Preserve applicable gotchas.

**Result:**
- Created `docs/prompts/chat-builder-system-prompt.md` — full implementation prompt
- 8 architecture patterns adapted from OpenAI Explorer
- 6 LedgerGuard modules defined with 16 total tools
- API contracts (WebSocket messages), wiring pattern, implementation order (38 steps)
- Gotchas table with 12 LedgerGuard-specific issues and fixes

---

### [2026-03-07] Switch AI Chat to OpenAI-First Architecture
**Original:**
> make OpenAI function calling now & make the architecture add claude api later

**Improved:**
> Update the AI Chat Builder prompt to use OpenAI function calling (gpt-4o) as the default AI provider — the same proven pattern from OpenAI Explorer. Add a provider-agnostic `AIClient` interface so Claude API can be swapped in later without changing the chat handler, modules, or frontend. Update all references from Claude-specific to OpenAI-first with Claude as future option. Add Phase 9 (Claude Migration) to implementation order.

**Result:**
- Updated `docs/prompts/chat-builder-system-prompt.md`:
  - Added `AIClient` interface with `OpenAIClient` (default) and `ClaudeClient` (future) implementations
  - All Claude API references → OpenAI function calling
  - Architecture diagram shows AIClient with OpenAI default
  - Wiring shows one-line swap from OpenAI to Claude
  - Added Phase 9: Claude API Migration to implementation order
  - Gotchas reflect OpenAI-specific errors
- Updated `future.md` Phase 1 description to reflect OpenAI-first approach

---

### [2026-03-07] AI Chat + Internal GraphQL — 12-Commit Implementation
**Original:**
> implement [the 13-commit AI Chat plan from `docs/prompts/chat-builder-system-prompt.md`]

**Improved:**
> Implement the AI Chat + Internal GraphQL feature as a 12-commit micro-phase plan. Each commit is a self-contained, testable unit: (1) docs, (2) config+deps, (3) GraphQL schema, (4) gqlgen+resolvers, (5) GraphQL endpoint+tests, (6) module framework, (7) risk module, (8) remaining 5 modules, (9) AIClient+OpenAI, (10) chat handler+SSE, (11) DB migration, (12) Flutter chat UI. Use SSE streaming instead of WebSocket (user decision). All commits follow TDD with tests before implementation.

**Result:**
- 12 commits implemented across backend and frontend
- Backend: `internal/chat/` package with GraphQL, 6 modules (16 tools), AIClient interface, SSE handler
- Frontend: ChatBloc, ChatPage, DataPanel, SSE client, `/chat` route
- 47 Go tests + 5 Flutter tests passing
- Architectural change: SSE chosen over WebSocket for chat streaming

---

### [2026-03-09] Material 3 Theme Standardization (Frontend)
**Original:**
> Migrate hardcoded TextStyle to M3 theme styles. Fix AppBar visibility, scaffold backgrounds, popup tints, button visibility.

**Improved:**
> Comprehensive Material 3 theme standardization across Flutter frontend: (1) Define custom textTheme in app_theme.dart and migrate ~120 hardcoded TextStyle(fontSize:) to Theme.of(context).textTheme.* across 21 files. (2) Fix dashboard initial load race condition — auto-select first app, defer load via BlocListener. (3) Switch AppBar to primary blue with white foreground. (4) Centralize scaffoldBackgroundColor grey-50, remove 9 hardcoded overrides. (5) Fix preferences Save button blue-on-blue visibility. (6) Add popupMenuTheme/dialogTheme with white bg + transparent surfaceTintColor.

**Result:**
- 8 commits: typography (ce6002d), auto-select app (b9aa36e), deferred dashboard load (e0123e9), blank flash fix, AppBar theme (20aa855), scaffold bg (b311d44), Save button (c5d31b9), popup/dialog theme (3be9e1c)
- All 405 Flutter tests passing
- Frontend docs updated: REQUIREMENTS.md (theming standards), IMPLEMENTATION_LOG.md, prompts.md

---

### [2026-03-09] Welcome & Onboarding Flow — Hybrid Design (Documentation)
**Original:**
> Now I am thinking about how to welcome new signups. Via email, via a flow (maybe call the webhook URL of that flow, maybe an external app). Visualization flow needed, Excalidraw if necessary. I need custom option along with this — calling an external API (that may be a third-party flow builder with their own mailer).

**Improved:**
> Design a Hybrid Welcome & Onboarding Flow for LedgerGuard combining: (1) Webhook-triggered email drip campaign via n8n (self-hosted) + Postmark (transactional email), (2) In-app onboarding checklist with step tracking and progress banner, (3) Custom webhook escape hatch for third-party flow builders (Customer.io, Brevo, ActiveCampaign) with HMAC-signed payloads. Deliverables: visualization prompt file, PlantUML flow diagram, ADR-015 (n8n) + ADR-016 (Postmark), future.md entry. Documentation only — no code implementation.

**Result:**
- Created `docs/prompts/welcome-onboarding-flow.md` — full visualization prompt with 4 flows + 2 decision comparison tables
- Created `docs/welcome-onboarding-flow.puml` — PlantUML activity diagram of hybrid flow
- Added ADR-015 (n8n) + ADR-016 (Postmark) to DECISIONS.md
- Added P1 entry to future.md
- No code changes — documentation/visualization deliverable

---

### [2026-03-09] Billing System Design — Stripe + Trial-Freemium (Documentation)
**Original:**
> Implement my billing. How to introduce plans table? Feature list as config in table? Payment gateway as Stripe or Shopify. If we're providing manual app connection we can't direct him to Shopify right because of no access token? Or we can ask them to install during billing setup. I need custom option along with this — calling an external API (that may be a third-party flow builder with their own mailer).

**Improved:**
> Design LedgerGuard's billing system using Stripe (confirmed — Shopify Billing is merchant-only, Partner API has no billing endpoints, no Partner Tools marketplace exists). Model: 14-day free trial (all features) → FREE tier (limited: 1 app, no AI Chat/API Keys/Slack/export) → PRO (paid, unlimited) → ENTERPRISE (custom). Database-driven feature gating via plans + plan_features tables. Deliverables: visualization prompt, PlantUML diagram, ADR-017 (Stripe) + ADR-018 (Trial-Freemium), plan doc PLAN-20.

**Result:**
- Created `docs/prompts/billing-system-flow.md` — full visualization prompt with 4 flows + Stripe decision table + data model
- Created `docs/billing-system-flow.puml` — PlantUML activity diagram of billing lifecycle
- Added ADR-017 (Stripe) + ADR-018 (Trial-Freemium) to DECISIONS.md
- Created `docs/plans/PLAN-20.md` + updated PLAN_INDEX.md
- Added P1 entry to future.md
- No code changes — documentation/visualization deliverable

---

### [2026-03-09] Billing Visualization — PlantUML + Marketing Page
**Original:**
> Implement the plan: Stripe Billing Visualization — PlantUML + Marketing Page. Add payment money flow (USD → Stripe → fees → INR → bank) and Stripe Customer creation strategies to PlantUML. Create interactive marketing page at /billing-flow with 6 tabs (Lifecycle, Payment, Checkout, Webhooks, Gating, Upgrade). Update prompt file with Flow 5 + Flow 6.

**Improved:**
> Extend billing system documentation with two missing flows: (1) Payment Money Flow showing customer USD → Stripe processing → fee deduction (~2.9% + $0.30) → USD→INR conversion → developer bank payout (T+2 to T+7), and (2) Stripe Customer Creation Strategies comparing 3 options (At Signup [chosen], At First Checkout, At Trial End) with trade-off analysis. Create interactive marketing visualization page at `/billing-flow` following existing pattern (ShopifyMoneyFlow.tsx) with 6 selectable tabs, SVG-based animated flow diagrams, play/pause controls, and reference cards. Update `docs/prompts/billing-system-flow.md` with new flows and component structure.

**Result:**
- Updated `docs/billing-system-flow.puml` — added PaymentMoneyFlow + StripeCustomerCreation diagrams
- Created `marketing/site/app/billing-flow/page.tsx` — page wrapper with explanation sections
- Created `marketing/site/components/BillingFlowVisualization.tsx` — 6-tab interactive visualization
- Updated `docs/prompts/billing-system-flow.md` — added Flow 5 (Payment Money Flow) + Flow 6 (Stripe Customer Creation)
- Logged to `prompts.md`

---

### [2026-03-15] Shop Logos on Subscription Pages
**Original:**
> Implement the plan: Add Shop Logo to Subscription Pages. Subscription list and detail pages currently show letter initials. Display actual shop logo from Shopify Storefront API, falling back to letter initials when unavailable.

**Improved:**
> Create `shops` table (migration 000029) to store Shopify store brand data (logo, square logo, cover image) fetched from the public Storefront API (`https://{domain}/api/2026-01/graphql.json`, no auth required). Add `entity.Shop`, `repository.ShopRepository`, `PostgresShopRepository`, `ShopifyStorefrontClient.FetchBrand()`. Integrate brand fetch into sync service (only for new domains). Enrich subscription list/detail API responses with `shop_logo_url` and `shop_square_logo_url`. Update Flutter `Subscription` entity with logo URL fields. Replace letter-initial avatars in `SubscriptionTile` and `SubscriptionDetailPage` with `CachedNetworkImage` + letter fallback.

**Result:**
- Created migration 000029 (shops table with domain unique index)
- Created `entity.Shop`, `repository.ShopRepository`, `PostgresShopRepository`
- Created `ShopifyStorefrontClient` for public Storefront API brand fetching (including `shop { id }` for GID)
- Integrated brand fetch into `SyncService` (only fetches new domains)
- Updated subscription handler to include `shop_name`, logo URLs in list/detail responses
- Updated store health handler with shopRepo for logo + name enrichment
- Updated Flutter `Subscription` entity with `shopName`, `shopLogoUrl`, `shopSquareLogoUrl`, `displayName` getter
- Updated `SubscriptionTile`, `SubscriptionDetailPage`, and `StoreHealthPage` with `CachedNetworkImage` + letter fallback
- Replaced domain-parsing name logic with `displayName` (prefers real shop name)
- Added `cached_network_image` dependency to pubspec.yaml
- Updated `DATABASE_SCHEMA.md` and `IMPLEMENTATION_LOG.md`
- All backend tests pass, all Flutter tests pass

---

### [2026-03-18] Razorpay Subscriptions Integration (Test Mode)
**Original:**
> Implement the Razorpay Subscriptions Integration plan (8 incremental commits)

**Improved:**
> Integrate Razorpay Subscriptions into the Go backend for LedgerSpear B2B SaaS billing (Starter $249/mo, Pro $499/mo). 8 incremental commits: domain value objects, entity, repository interface + migration, PostgreSQL repo, Razorpay HTTP client, BillingService, HTTP handlers, wire everything. Test mode only.

**Result:**
- 8 commits implementing full Razorpay billing integration
- Domain: BillingPlan, BillingSubscriptionStatus, BillingSubscription entity
- Infrastructure: RazorpayClient (Basic Auth, HMAC webhook verification)
- Application: BillingService (checkout, status, webhook event routing)
- Handlers: POST /billing/checkout, GET /billing/status, POST /webhooks/razorpay
- Migration 000030: billing_subscriptions table
- Config: RazorpayConfig with env var overrides
- All tests pass, server starts without Razorpay config (graceful skip)

---

### [2026-03-18] Frontend Billing UI (Razorpay Integration)
**Original:**
> frontend changes

**Improved:**
> Add Billing/Subscription UI to Flutter app for the Razorpay billing flow. Follow existing Bloc + Clean Architecture pattern. Create BillingStatus entity, BillingRepository interface + API impl, BillingBloc, BillingPage with plan cards and Razorpay hosted checkout via url_launcher. Wire into DI, GoRouter, and settings page.

**Result:**
- Created `BillingStatus` entity with `fromJson`, display helpers
- Created `BillingRepository` interface + `ApiBillingRepository` (ApiClient)
- Created `BillingBloc` (LoadBillingStatus, StartCheckout events)
- Created `BillingPage` with current plan card, Starter/Pro plan cards, subscribe CTA
- Opens Razorpay `short_url` via `url_launcher` (no SDK needed)
- Registered in get_it DI, added `/settings/billing` GoRoute
- Added "Plan & Billing" tile to settings page
- Flutter analyze: 0 errors

---

### [2026-05-07] Finalize Revenue API Format
**Original:**
> Implement the following plan: Finalize Revenue API Format — Accept Numeric IDs + Add Usages-by-Subscription

**Improved:**
> Fix Revenue API DX: (1) batch endpoints accept numeric IDs via suffix matching, (2) add GET /usages?subscription_id={id} endpoint, (3) fix ReadModelBuilder to use chargeId as usage shopify_gid, (4) add admin rebuild-read-model endpoint, (5) update Postman + docs.

**Result:**
- 20 files changed across repos, services, handlers, router, Postman, docs
- Squashed into commit: `8acfa59`

---

### [2026-04-21] Queue-Based Async Sync System
**Original:**
> Implement the following plan: Queue-Based Sync System — Implementation Plan (5 phases A-E)

**Improved:**
> Implement a complete async, queue-based sync system per `docs/developer/QUEUE_SYNC_IMPLEMENTATION_PLAN.md` (v3). Five phases: (A) Foundation — Redis client, migrations, entities, repo interfaces/impls, config. (B) Queue core — enqueue/dequeue, distributed locks, dual-write progress, recovery. (C) Worker framework — processor interface, registry, context, worker pool. (D) 7 processors — transaction, snapshot, event, status, store, review, full_sync orchestrator. (E) HTTP layer — QueueSyncService, QueueSyncHandler with 5 endpoints, router + main.go wiring. Each phase must compile before proceeding. Feature-flagged via QUEUE_ENABLED.

**Result:**
- Implemented all 5 phases (29 new files, 4 modified)
- All existing tests pass
- Commit: `38bf1c2`

---

### [2026-05-07] Daily Catch-Up Sync (Transactions + Events)
**Original:**
> Implement the following plan: Daily Catch-Up Sync — Transactions + Events (Last 2 Days) [detailed plan with 8 files to modify]

**Improved:**
> Implement DailyCatchupScheduler (3 AM UTC, 2-day lookback) that enqueues transaction_sync + event_sync for all active apps. Add LookbackDays to SyncJobPayload (backward-compatible). Update TransactionProcessor to use lookback window. Add EnqueueCatchupSync to QueueSyncService. Admin endpoint: POST /api/v1/admin/sync/daily-catchup. Internal endpoint: POST /api/v1/internal/sync/daily-catchup for Cloud Scheduler. Update Postman collection.

**Result:**
- 11 files changed (1 new: `daily_catchup_scheduler.go`)
- `queue.go`, `transaction_processor.go`, `queue_sync_service.go`, `admin.go`, `router.go`, `main.go`, Postman collection, IMPLEMENTATION_LOG, DECISIONS (ADR-032), docs/developer/31
- Commit: `1f4b76c`

---

### [2026-05-07] User Personas PlantUML Diagram (9 initial)
**Original:**
> want puml whatever diagram which shows for example: client has no website or anything they just had a embeded shopify app. so they will use my app to see installations, subscriptions. example 2: client has non embeded app... [6 examples of different user types]

**Improved:**
> Create a PlantUML use-case diagram mapping every distinct user persona for LedgerGuard, showing what features each persona uses and why they'd pay. Include 9+ personas: Embedded-Only Dev, Non-Embedded Dev, API-First Dev, Notification-Only, Mobile-First Dev, AI Power User, Agency/Multi-App, Finance/Compliance, Marketplace Veteran. Group features into packages. Add pricing tier legend.

**Result:**
- Created initial `docs/diagrams/puml/USER_PERSONAS.puml` with 9 personas
- Identified 6 additional personas needed (Freemium Dev, Investor, CS Manager, Support Lead, Side-Project Dev, Growth/Marketing)
- Expanded to 15 personas with `docs/USER_PERSONAS.md` (feature coverage matrix, pricing tiers, gap analysis)
- Restructured into mindmap overview + 3 focused group diagrams after original was visually unreadable
- Files: `USER_PERSONAS.puml`, `USER_PERSONAS_1_dashboard.puml`, `USER_PERSONAS_2_api_ops.puml`, `USER_PERSONAS_3_power.puml`, `USER_PERSONAS.md`
- Commit: `9b1aa89` (squashed from 4 commits)

---

### [2026-05-07] Missing UI Features Gap Analysis
**Original:**
> now note missing ui feature or menus from whatever we discussed and add it to future

**Improved:**
> Review all 15 user personas and their required features from USER_PERSONAS.md. Cross-reference with existing Flutter app screens/menus in both frontend/app (Bloc) and frontend-flutter (Provider). Identify every UI feature, screen, or menu item referenced in personas but not yet built. Append as actionable items to future.md grouped by feature area with priority, description, backend readiness, and persona references.

**Result:**
- 18 missing UI features added to `future.md` across 8 areas: Conversion & Growth, Reporting & Export, Notifications, Risk & CS, Dashboard, Reviews, Mobile, Sync & Data
- Key gaps: conversion funnel, PDF/CSV export, weekly digest, at-risk outreach list, global domain search, AI brief card, revenue concentration, sync jobs history
- Commit: `9b1aa89` (included in squash)

---

## Prompt: Mock Shopify Partner API Server + Admin UI + Webhook Tester

**Date:** 2026-05-07

**Original Prompt:** Implement the following plan: Mock Shopify Partner API Server + Admin UI + Webhook Tester

**Improved Prompt:** Build a Ruby/Sinatra mock server at `mock-shopify-api/` that serves Shopify Partner API GraphQL responses based on 4 YAML-defined personas (Solo/Growing/Power/Churning). Include admin UI for viewing/editing persona data and webhook tester panel. Modify Go backend to support `partner_api_url` config override for routing API calls to the mock server.

**Result:**
- Mock server with 4 personas, GraphQL endpoint, admin UI, and webhook tester
- Backend `WithBaseURL()` option + `partner_api_url` config field
- `config.mock.yaml` for running backend against mock server
- All Go tests pass, mock server verified with curl

---

### [2026-05-07] Fix App Filter API Calls + Transaction Upsert + Disconnect Cleanup
**Original:**
> Fix the providers so that when an app is selected, it triggers API calls in live mode. Also fix the transaction upsert to include app_id in ON CONFLICT. And make sure disconnect flow resets all providers to demo mode.

**Improved:**
> Audit and fix all 8 data providers (Dashboard, Store, Subscription, Transaction, Earnings, Events, Risk, Analytics) so `setSelectedApp()`/`setAppFilter()` triggers `load*()` when in live mode. Fix `transaction_repository.go` ON CONFLICT to include `app_id = EXCLUDED.app_id`. Add debug logging to `ledger_service.go` (charge type breakdown, domain count, subscription count). Ensure `connect_shopify_screen.dart` disconnect calls `setDemoMode(true)` on all 10 providers.

**Result:**
- 8 providers fixed with live-mode API load triggers
- Transaction upsert ON CONFLICT now includes app_id
- Ledger service has 4 new debug log statements
- Disconnect flow resets all 10 providers to demo mode

---

### [2026-05-07] Multi-User Team Access Model Design
**Original:**
> Design the multi-user / team access model for LedgerGuard. Allow team members (co-founders, VAs, bookkeepers) to access the same data. Include organizations, roles, invitations, audit log, per-member notifications, API key scoping, SSO/SAML (PRO), org-level webhooks.

**Improved:**
> Design a comprehensive multi-user team access model with: (1) Organizations as top-level data-owning entity replacing user-level ownership; (2) Multi-org support (user belongs to multiple orgs); (3) Role-based access (OWNER/ADMIN/VIEWER) with plan-based member limits (FREE=1, STARTER=3, PRO=10); (4) Member lifecycle (invite→active→suspend→remove); (5) Org audit log; (6) Per-member notification preferences; (7) API key scoping to org; (8) SSO/SAML for PRO; (9) Org-level webhooks; (10) Zero-downtime migration strategy with backfill. Include new DB tables, API endpoints, middleware, Flutter UI changes, and documentation artifacts.

**Result:**
- Full design document with 4 new tables (organizations, org_members, org_invitations, org_audit_log)
- 3 modified tables (partner_accounts, billing_subscriptions, api_keys)
- 20+ new API endpoints with role-based access
- OrgContextMiddleware design
- Member lifecycle state machine
- Plan-based team limits
- 10-phase implementation plan
- Zero-downtime migration strategy

---

### [2026-05-07] Multi-User Team Access Model — Full Implementation
**Original:**
> Implement the following plan: Part 1 (documentation audit) + Part 2 (multi-user team access model) across all phases. Then: frontend changes done? postman collection api done? → yes do both.

**Improved:**
> Execute the full multi-user org model implementation plan across 8 phases: (A) DB migration 000036 + domain entities + value objects + repository interfaces; (B) PostgreSQL repository implementations + OrgService + OrgAuditService; (C) 24 unit tests for org service with mock repos; (D) OrgContextMiddleware for org resolution + role/status checking; (E) OrgHandler (15 endpoints) + OrgAuditHandler + router wiring with X-Org-Id CORS; (F) Flutter organization UI — models, service, provider, team screen, audit log screen, org switcher widget, wiring into main.dart/app.dart/app_shell/settings; (G) Postman collection with 16 org endpoints; (H) Documentation — DATABASE_SCHEMA.md, DECISIONS.md (ADR-033, ADR-034), IMPLEMENTATION_LOG.md, all PlantUML diagrams (ER, C4, SEQUENCE, webhook, new 36-org-management).

**Result:**
- Migration 000036: 4 new tables (organizations, org_members, org_invitations, org_audit_log) + org_id on partner_accounts/api_keys
- Domain: organization.go entity, 3 value objects (OrgRole, MemberStatus, InvitationStatus)
- 4 repository interfaces + 4 PostgreSQL implementations
- OrgService: create org, invite (plan limits), accept, suspend, unsuspend, change role, webhooks
- OrgAuditService: separate from existing user-level AuditService
- 24 unit tests — all passing
- OrgContextMiddleware: resolves org from URL/{orgId} or X-Org-Id header, auto-selects single org
- OrgHandler: 15 HTTP endpoints wired in router
- Flutter: organization_model.dart, organization_service.dart, organization_provider.dart (with ApiClient.setOrgId wiring)
- Flutter UI: TeamScreen (members + invite dialog + suspend/remove), AuditLogScreen, OrgSwitcher (desktop rail + tablet drawer)
- Flutter wiring: main.dart (provider registration), app.dart (routes), settings_screen.dart (Organization card), app_shell.dart (OrgSwitcher)
- flutter analyze: 0 issues
- Postman: "19 - Organizations" folder with 16 requests + 4 collection variables
- ER_current.puml: +4 entities, +10 relationships, +4 notes
- C4_current.puml: +OrgService, +OrgHandler, +OrgContextMW, +4 new REST handlers (Transaction/Store/Event/Risk)
- SEQUENCE_current.puml: +4 flows (invite, accept, suspend/unsuspend, org switch)
- New: 36-org-management-sequence.puml (6-page detailed org flows)
- Updated: 14-webhook-processing-sequence.puml (app installed + reinstall flow)
- ADR-033: ON CONFLICT includes app_id
- ADR-034: Organizations as top-level data-owning entity
- Commit: `5bd4f29`

---

### [2026-05-08] Wire Org Context into Existing Data Endpoints
**Original:**
> Implement the plan: Wire Org Context into Existing Data Endpoints

**Improved:**
> Complete the org model wiring by: (1) adding OrgID to PartnerAccount entity and FindByOrgID to repository interface, (2) creating resolvePartnerAccount helper that prefers org-based lookup with user-fallback, (3) replacing FindByUserID in all 26 handler call sites across 14 files, (4) adding OrgContextMW to /apps, /sync, /metrics, /integrations route groups, (5) adding selected_org_id to user_preferences for backend persistence, (6) updating documentation. Maintain backward compatibility — single-org users auto-select, multi-org users use X-Org-Id header.

**Result:**
- Phase 1: Entity + repo interface + persistence (org_id on all queries, scanOne DRY refactor)
- Phase 2: resolvePartnerAccount helper + lookupAppID rewrite
- Phase 3: 26 call sites in 14 handler files updated
- Phase 4: OrgContextMW on 4 route groups (guarded)
- Phase 5: Migration 000038 + GET/PUT /user/preferences/selected-org (backend persistence instead of Flutter SharedPreferences, per user request)
- Phase 6: IMPLEMENTATION_LOG, DECISIONS (ADR-035), DATABASE_SCHEMA, all PlantUML/Excalidraw/developer docs/Postman
- 10 test mocks updated with FindByOrgID stubs
- 52 files changed, `go build ./...` clean, `go test ./...` all pass
- Commit: `4021029`

---

### [2026-05-09] Events & Webhooks Page Improvements
**Original:**
> Implement the plan: Events & Webhooks Page Improvements

**Improved:**
> Fix Events page: add 4 missing EventType enum values (appReactivated, appDeactivated, subscriptionFrozen, subscriptionUnfrozen) with Shopify API aliases, add icons/colors/badges for new types, fix KPI cards to count from unfiltered events, wire eventType filter to API. Fix Webhooks page: change mockWebhooks from final to getter for fresh timestamps, add demo mode toggle wired to DemoModeCoordinator, add app filter and time range toggle (Today/Week/Month), make KPI cards respect selected time range.

**Result:**
- Commit 1: Events page fixes (5 files)
- Commit 2: Webhooks page fixes (5 files)
- Commit 3: Documentation (4 files)
- `flutter analyze` clean

---

### [2026-05-09] Analytics & Earnings — Complete Wiring & Missing Features

**Original:**
> Implement the following plan: [Analytics & Earnings — Complete Wiring & Missing Features Plan with 6 phases, 15 commits]

**Improved:**
> Implement the 6-phase plan to (1) fix the P1 mock-data-leaking bug in all 5 analytics tabs, (2) wire existing backend APIs to earnings page tiers/fee-calculator/status and to MRR movements, (3) build 4 new backend endpoints (forecasting engine with linear+exponential models, cohort retention, monthly profit P&L, revenue concentration), (4) wire all new endpoints to their frontend analytics tabs with model selector/empty states, (5) add 3 enhancements (revenue concentration, historical snapshot browser, 30-day earnings outlook), (6) update documentation. Execute in dependency order: Phase 1 → 2a → 2b → 3a-3d (parallel) → 4a-4d → 5a-5c → 6.

**Result:**
- 15 commits across all 6 phases
- ~50 files changed (backend + frontend)
- All Go tests pass, `flutter analyze` clean

### [2026-05-09] Snapshot Strategy — Daily Backfill, Query-Time Aggregation & Trend API
**Original:**
> Implement the following plan: [Snapshot Strategy — Daily Backfill, Query-Time Aggregation & Trend API with 5 phases, 8 commits]

**Improved:**
> Implement ADR-041: single daily table with query-time downsampling. (1) Fix DEV LIMIT in snapshot_processor.go (1m → 12m). (2) Optimize `BackfillHistoricalSnapshots` from O(365×n) to O(n log n) with sort-once + pointer advancement + UpsertBatch. (3) Add Granularity value object (DAILY/WEEKLY/MONTHLY) + DownsampleSnapshots (picks last snapshot per period). (4) Create REST `GET /apps/{appID}/metrics/trend?granularity=weekly` + enhance GraphQL `metricsTrend` with granularity enum. (5) Wire Flutter frontend to trend API with weekly granularity for Revenue tab chart. (6) Update docs: ADR-041, DATABASE_SCHEMA, TAD, IMPLEMENTATION_LOG, future.md, verification, prompts.

**Result:**
- 19 files changed across backend + frontend
- 9 new tests (4 backfill optimization + 5 downsampling)
- All Go tests pass, flutter analyze clean
- ADR-041 documented in DECISIONS.md

---

### [2026-05-09] Wire Dashboard Preferences — Dynamic KPIs + Widgets via Backend API
**Original:**
> Implement the plan: Wire Dashboard Preferences — Dynamic KPIs + Widgets via Backend API

**Improved:**
> Implement end-to-end dashboard customization: (1) Fix backend `defaultPreferences()` to match current dashboard layout, (2) Add `getDashboardPreferences()` / `saveDashboardPreferences()` to `UserPreferencesService`, (3) Create a KPI + widget registry mapping IDs to rendering config, (4) Make `DashboardProvider` load/save prefs and expose `primaryKpis`/`secondaryWidgets`, (5) Replace hardcoded dashboard cards with dynamic registry-based rendering including new `_EarningsOverviewCard`, (6) Add Dashboard Customization section in Settings with checkbox toggles (max 4 KPIs enforced), (7) Wire `userPreferencesService` into `DashboardProvider` in `main.dart`.

**Result:**
- 7 files changed (1 new registry, 6 modified)
- Backend defaults aligned to actual dashboard
- Dashboard renders KPIs and widgets dynamically from user preferences
- Settings screen allows toggling KPIs (max 4) and secondary widgets
- flutter analyze: 0 issues, backend builds clean

---

## Prompt 39: Wire Settings Preferences — Notification, Sync & Workspace via Backend API

**Date:** 2026-05-09

**Original prompt:** Wire Settings Preferences — Notification, Sync & Workspace via Backend API (detailed plan with 10 phases, gap analysis, migration SQL, endpoint specs, frontend wiring)

**Improved prompt:** Implement full-stack wiring of Settings page preferences to backend API: extend notification_preferences table with 6 new columns (migration 000039), add sync/workspace columns to user_preferences (migration 000040), update backend entity/repo/handler, add GET/PUT settings endpoints, wire frontend SettingsProvider to API with optimistic updates, create PlantUML sequence diagram, update all documentation.

**Result:**
- 18 files changed (6 new, 12 modified)
- 2 migrations (000039, 000040) adding 11 new columns across 2 tables
- Backend: entity + repo + handler updated, 2 new endpoints registered
- Frontend: service methods + SettingsProvider wired to API + screen loads on open
- PlantUML sequence diagram created
- go build: clean, flutter analyze: 0 issues

---

## Prompt 40: Diagram Audit & Update

**Date:** 2026-05-09

**Original prompt:** run diagram audit

**Improved prompt:** Execute `docs/prompts/diagram-audit-prompt.md` — audit all architecture diagrams (PlantUML, Excalidraw, sequence), update outdated ones, create missing diagrams for backend and frontend-flutter.

**Audit findings:**
- `C4_current.puml` — CURRENT (393 lines)
- `ER_current.puml` — OUTDATED (missing migrations 000039/000040 columns)
- `SEQUENCE_current.puml` — CURRENT (22 flows, 1563 lines)
- 35 feature diagrams (01–39) — ALL CURRENT
- 12 excalidraw files — UNKNOWN (binary format)
- `frontend-flutter/docs/MOBILE_NAVIGATION.puml` — CURRENT
- `frontend/docs/SCREENS.puml` — OUTDATED (Bloc version, not active)

**Actions taken:**
1. **Updated** `docs/ER_current.puml` — Added 6 notification_preferences columns + 5 user_preferences columns from migrations 000039/000040
2. **Created** `frontend-flutter/docs/SCREENS.puml` — Full screen flow diagram (auth screens, AppShell, 13 navigation branches, sub-routes)
3. **Created** `docs/diagrams/puml/40-flutter-provider-service-graph.puml` — Provider → Service → API dependency graph (16 providers, 13 services, DemoModeCoordinator)
4. **Created** `docs/diagrams/puml/41-backend-api-endpoint-map.puml` — All ~85 REST endpoints grouped by domain, color-coded by auth scheme

**Result:**
- 1 diagram updated, 3 new diagrams created
- All current diagrams verified against codebase
- No deletions, no code changes

---

## Prompt 41: Diagram Audit (second run — Postman focus)

**Date:** 2026-05-09

**Original prompt:** run diagram audit

**Improved prompt:** Re-run diagram audit with new Postman collection step added. Diagrams already current from prior run — focus on Postman gap analysis.

**Postman audit findings:**
- 2 endpoints missing (GET/PUT /user/preferences/settings)
- 4 request bodies outdated (dashboard, default-app, selected-org, notification-preferences)
- 1 ghost route (PUT /integrations/shopify/token — doesn't exist in router)

**Actions taken:**
1. Added GET + PUT `/user/preferences/settings` requests
2. Fixed dashboard prefs body (primary_kpis + secondary_widgets)
3. Fixed default-app body (field renamed to default_app_id)
4. Fixed selected-org body (field renamed to selected_org_id)
5. Fixed notification-preferences body (added 6 new fields)
6. Fixed ghost PUT token → changed to POST (actual method)

**Result:**
- Postman collection now matches router.go (107 routes, all covered)
- Valid JSON verified

---

### [2026-05-09] Move Dashboard Customization to Settings Sub-Page

**Original:**
> Move the Dashboard card (KPI + widget checkboxes) out of the main settings page into a /settings/dashboard sub-page, matching the Team/AuditLog/ConnectShopify pattern. Log rejected alternatives to future.md.

**Improved:**
> Extract the ~50-line Dashboard customization card from `settings_screen.dart` into a new `DashboardSettingsScreen` at `/settings/dashboard`. Replace inline checkboxes with a navigation ListTile. Register GoRoute. Update SCREENS.puml. Log two rejected alternatives (all sub-pages, collapsible sections) to `future.md`.

**Result:**
- 1 new file: `dashboard_settings_screen.dart`
- 4 modified: `settings_screen.dart`, `app.dart`, `SCREENS.puml`, `future.md`
- `flutter analyze` — no issues

---

### [2026-05-09] Diagram Audit

**Original:**
> run diagram audit

**Improved:**
> Execute full diagram audit per `docs/prompts/diagram-audit-prompt.md`. Audit all PlantUML, Excalidraw, and sequence diagrams in `docs/`. Update outdated diagrams, verify Postman collection against `router.go`, check frontend-flutter diagrams, and ensure consistency.

**Result:**
- **37 numbered diagrams** (01–40): all CURRENT, no updates needed
- **Core diagrams** (ER_current, C4_current, SEQUENCE_current): all CURRENT
- **Frontend diagrams** (SCREENS.puml, MOBILE_NAVIGATION.puml): CURRENT
- **41-backend-api-endpoint-map.puml**: OUTDATED → fixed 5 issues (missing `GET /metrics/aggregate`, wrong paths for refresh-install-count / reviews/scrape / internal sync, missing invitation ID param), updated count ~85 → ~107
- **Postman collection**: Fixed 9 parameter name mismatches (apiKeyId→id, memberId→userId, invitationToken→token, invitationId→id), added GET /graphql endpoint, replaced hardcoded IDs with variables

---

### [2026-05-10] Shopify Ecosystem Diagram

**Original:**
> I also want a Shopify ecosystem diagram (any one or more puml + excalidraw) includes Shopify App Store, embedded app, non-embedded app, website theme, shops, admin account, partner account, dropshippers, external apps for shops & admins & partners, any external apps who help Shopify & other ecosystem like Amazon. Also highlight LedgerGuard's position.

**Improved:**
> Create a comprehensive Shopify Ecosystem diagram set with all actors, platforms, APIs, extension surfaces, and external integrations. Produce 3 files: (1) full ecosystem PlantUML component diagram, (2) LedgerGuard-focused PlantUML showing data/money flows through the ecosystem to LedgerGuard, (3) Excalidraw visual map. Enriched with Shopify AI assistant's response covering all account types, API surfaces, extension types, and ecosystem categories.

**Result:**
- `docs/diagrams/puml/42-shopify-ecosystem.puml` — Full ecosystem (~80 components, 13 actor types, 8 external categories)
- `docs/diagrams/puml/43-ledgerguard-in-shopify-ecosystem.puml` — LedgerGuard-focused: money flow, data ingestion, processing pipeline, user types
- `docs/diagrams/shopify-ecosystem.excalidraw` — Visual map (138 elements, 5 color-coded zones)

### [2026-05-18] Diagram Audit

**Original:**
> run diagram audit

**Improved:**
> Execute the full diagram audit prompt at `docs/prompts/diagram-audit-prompt.md`. Audit all 63 diagrams (PlantUML, Excalidraw, sequence) across backend + frontend-flutter. Identify outdated diagrams, update them to match current codebase, create missing ones, verify Postman collection completeness, and run a consistency pass across all diagrams.

**Result:**
- **ER_current.puml** — Added `billing_subscriptions` entity, updated `shops` (6 new fields), added `app_store_slug` to `apps`, added relationship + notes
- **C4_current.puml** — Added 6 missing handlers (ForecastHandler, CohortHandler, ReviewHandler, NotificationPreferencesHandler, DeviceHandler, BillingHandler), 7 missing repos + impls (Shop, AppReview, BillingSubscription, DeviceToken, NotificationPreferences, Revenue, AppEvent), 3 domain services (EarningsCalculator, ForecastingEngine, FeeVerificationService), 3 value objects (BillingPlan, BillingSubscriptionStatus, Granularity), Apps chat module, fixed ChatHandler WebSocket→SSE
- **41-backend-api-endpoint-map.puml** — Updated title from ~107 to ~127 endpoints, removed phantom `GET /api/v1/usages` endpoint
- **40-flutter-provider-service-graph.puml** — Added ApiKeyProvider→API and WebhookProvider→API arrows
- **SEQUENCE_current.puml** — Updated Flow 14 from WebSocket to SSE (title, note, protocol)
- **18-ai-chat-tool-loop-sequence.puml** — Fixed WebSocket→SSE references (participant, message arrow)
- **18-ai-chat-modules-component.puml** — Fixed WebSocket→SSE references (component, protocol note)
- **Postman collection** — Verified 110 entries cover all 127 router endpoints (GET+PUT pairs consolidated); minor issues noted (duplicate token endpoint, phantom usages path)
- **No new diagrams needed** — All candidate diagrams from audit prompt already exist

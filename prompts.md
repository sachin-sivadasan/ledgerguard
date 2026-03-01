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

**Result:**
- marketing/site/components/KPIMetricsGuide.tsx (1264 lines)
- marketing/site/app/kpi-guide/page.tsx
- Build verified: `npm run build` successful
- Commit: f91d064 feat(marketing): add KPI metrics visualization component
- Pushed to origin/main

---

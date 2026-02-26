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

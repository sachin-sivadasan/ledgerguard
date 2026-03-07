# Architecture Decisions – LedgerGuard

## Format
```
### ADR-XXX: Title
**Date:** YYYY-MM-DD
**Status:** Accepted / Superseded / Deprecated

**Context:**
Why this decision was needed.

**Decision:**
What we decided.

**Consequences:**
Trade-offs and implications.
```

---

## Decisions

### ADR-001: Modular Monolith over Microservices
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need to choose architecture style for MVP. Team is small, rapid iteration needed.

**Decision:**
Build as a modular monolith in Go with clean architecture. Modules communicate via interfaces, not network calls.

**Consequences:**
- Faster development
- Simpler deployment
- Easy refactoring
- Can extract to microservices later if needed

---

### ADR-002: Full Ledger Rebuild over Incremental Updates
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need to decide how to sync transactions and compute metrics.

**Decision:**
Rebuild entire 12-month ledger on every sync instead of incremental updates.

**Consequences:**
- Deterministic: same input always produces same output
- Simpler to debug and audit
- Higher compute cost (acceptable at MVP scale)
- Can optimize later with hybrid approach

---

### ADR-003: Firebase Authentication
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need authentication system with Google OAuth support.

**Decision:**
Use Firebase Authentication. Frontend gets ID token, backend verifies via Admin SDK.

**Consequences:**
- Fast to implement
- Google OAuth included
- Stateless verification
- Vendor lock-in (acceptable trade-off)

---

### ADR-004: PostgreSQL as Primary Database
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need a database for transactions, subscriptions, snapshots.

**Decision:**
Use PostgreSQL with pgcrypto for UUID generation.

**Consequences:**
- ACID compliance
- JSON support if needed
- Well-known, easy to hire for
- Requires managed instance in production

---

### ADR-005: Domain-Driven Design over Clean Architecture
**Date:** 2026-02-26
**Status:** Accepted

**Context:**
Initial implementation used Clean Architecture folder structure. Need clearer separation between business logic and infrastructure with explicit domain modeling.

**Decision:**
Refactor to Domain-Driven Design (DDD) structure:
- `domain/` - Entities, value objects, domain services, repository interfaces
- `application/` - Use cases, DTOs, orchestration
- `infrastructure/` - Database, external services, config
- `interfaces/` - HTTP handlers, middleware, routing

**Consequences:**
- Better domain isolation (domain layer has zero external dependencies)
- Clearer boundaries between layers
- Repository interfaces defined in domain (ports), implementations in infrastructure (adapters)
- More explicit modeling of business concepts
- Slightly more directories, but clearer responsibilities

---

### ADR-006: OAuth State Validation for CSRF Protection
**Date:** 2026-02-27
**Status:** Accepted

**Context:**
OAuth callback endpoint was missing state parameter validation, creating a CSRF vulnerability where an attacker could complete OAuth flow with their own credentials and link to victim's account.

**Decision:**
Implement in-memory state store with:
- State stored with user ID when StartOAuth called
- State validated and consumed (one-time use) in Callback
- 10-minute TTL for expiration
- State lookup returns associated user ID

**Consequences:**
- CSRF protection for OAuth flow
- No external dependencies (in-memory store)
- Needs Redis/distributed cache for multi-instance deployment
- Tests added for state validation

---

### ADR-007: Tenant Isolation in Sync Handler
**Date:** 2026-02-27
**Status:** Accepted

**Context:**
SyncApp endpoint allowed users to sync any app by ID without verifying ownership, creating a tenant isolation vulnerability.

**Decision:**
Add ownership verification before sync:
1. Get user's partner account from context
2. Lookup requested app by ID
3. Verify app.PartnerAccountID matches user's partner account
4. Return 403 Forbidden if mismatch

**Consequences:**
- Users can only sync their own apps
- Additional database lookup per request (acceptable)
- Tests added for forbidden case

---

### ADR-008: Default Revenue Share Tier Changed to 0%
**Date:** 2026-03-01
**Status:** Accepted

**Context:**
The default revenue share tier was set to 20% (DEFAULT_20), but the majority of Shopify app developers (especially indie developers) are on the reduced revenue share plan with 0% on their first $1M lifetime earnings.

**Decision:**
Change the default revenue share tier from DEFAULT_20 (20%) to SMALL_DEV_0 (0%):
- Backend: `entity.NewApp()` defaults to `RevenueShareTierSmallDev0`
- Backend: `ParseRevenueShareTier()` returns `SMALL_DEV_0` for invalid/empty input
- Frontend: `RevenueShareTier.fromCode()` defaults to `smallDev0`
- Users can change their tier in App Settings if they're on a different plan

**Consequences:**
- More accurate default for majority of indie developers
- Reduces initial confusion about fee calculations
- Users on 20% tier need to manually update their setting
- Existing apps in database retain their current tier (no data migration needed)

---

### ADR-009: GCP Cloud Run for Staging Environment
**Date:** 2026-03-03
**Status:** Accepted

**Context:**
GCP offers $300 in free credits for 90 days. We need a staging environment to test changes before deploying to production on Hetzner. Options considered: GCE (VM), GKE (Kubernetes), Cloud Run (serverless containers).

**Decision:**
Use GCP Cloud Run for staging because:
- Scales to zero when idle (preserves free credits)
- No infrastructure management (serverless)
- Existing Dockerfile works without modification
- Auto-HTTPS included (no Caddy needed)
- Branch-based CI/CD: main → Hetzner, staging → Cloud Run

Supporting services: Cloud SQL (PostgreSQL 14, db-f1-micro), Artifact Registry, Secret Manager, VPC Connector for private networking.

**Consequences:**
- Free staging environment for 90 days
- After credits expire, estimated ~$20-30/month (can be shut down)
- Cold starts when scaling from zero (~2s for Go binary, acceptable for staging)
- No custom domain for staging (uses auto-generated Cloud Run URL)
- Same codebase, different config — validates production readiness

---

### ADR-010: Staging Frontend Deployment via Firebase Hosting + main_staging.dart
**Date:** 2026-03-05
**Status:** Accepted

**Context:**
The Flutter web frontend needed to be deployed for staging, pointing to the GCP Cloud Run backend rather than the production Hetzner backend. The existing entry points were `main_dev.dart` (localhost) and `main_prod.dart` (production API). Firebase Hosting was already configured from ADR for production deployment.

**Decision:**
- Create a separate `main_staging.dart` entry point with `Environment.staging` pointing to Cloud Run URL (`https://ledgerspear-api-ineifpjrdq-uc.a.run.app`)
- Deploy to the same Firebase Hosting project (`ledgerguard-c7557`) using the staging entry point
- Add Firebase Hosting domains to backend CORS allowlist
- Use `--platform linux/amd64` in Docker builds for Cloud Run compatibility with Apple Silicon

**Consequences:**
- Clean separation of dev/staging/prod environments in frontend config
- Single Firebase Hosting project serves staging for now (can split later)
- CORS explicitly lists allowed origins rather than using wildcards (more secure)
- Apple Silicon developers must use platform flag for Cloud Run builds
- Firebase Auth requires Identity Toolkit API + Token Service API enabled on API key restrictions

---

### ADR-011: gqlgen for Internal GraphQL Layer
**Date:** 2026-03-07
**Status:** Accepted

**Context:**
Need a GraphQL layer for AI Chat to query revenue data. Options: gqlgen (code-first Go), graphql-go, or custom JSON-based query engine.

**Decision:**
Use gqlgen — already a transitive dependency via Revenue API. Schema-first approach with code generation. Internal-only (not a public API contract), protected by Firebase Auth.

**Consequences:**
- Type-safe resolvers generated from schema
- Resolvers delegate to existing domain services (no new business logic)
- GraphQL Playground available for testing at `/graphql` (GET)
- Schema can iterate freely since it's internal
- `go generate` step required after schema changes

---

### ADR-012: OpenAI-First with AIClient Interface
**Date:** 2026-03-07
**Status:** Accepted

**Context:**
Need an LLM for chat function calling. OpenAI gpt-4o has mature function calling. Want to add Claude later.

**Decision:**
Create `AIClient` interface with OpenAI as first implementation. Interface methods: `ChatCompletion()` with tool definitions. Provider selected per-user via `ai_provider` column in `user_preferences`.

**Consequences:**
- OpenAI ships first — proven function calling pattern
- Claude API can be added as parallel provider without changing chat handler or modules
- Users can switch providers in settings
- Each provider handles its own tool format conversion

---

### ADR-013: WebSocket for Chat Communication
**Date:** 2026-03-07
**Status:** Accepted

**Context:**
AI chat needs real-time communication. Options: REST polling, SSE, WebSocket.

**Decision:**
Use WebSocket (`gorilla/websocket`) for bidirectional streaming. Supports typing indicators, tool call progress, and streaming responses.

**Consequences:**
- Real-time UX (no polling delay)
- Slightly more complex than REST
- Needs WebSocket upgrade support in reverse proxy (Caddy handles this)
- Firebase Auth token sent as query parameter on connection

---

### ADR-014: Module Plugin Architecture for Chat Tools
**Date:** 2026-03-07
**Status:** Accepted

**Context:**
AI chat needs 16+ tools across 6 domains (risk, subscriptions, metrics, etc.). Need organized, extensible structure.

**Decision:**
Module plugin pattern: each module implements `Module` interface with `Name()`, `Description()`, `PromptFragment()`, `Tools()`, `ExecuteTool()`. Registry manages registration and routing. Tool names use `module__tool_name` format (double underscore for OpenAI compliance).

**Consequences:**
- Modules are self-contained and independently testable
- New modules can be added without touching existing code
- System prompt assembled dynamically from registered modules
- Clear ownership: each module owns its tools and GraphQL queries

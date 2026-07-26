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

---

### ADR-015: n8n for Automation Platform (Welcome Flow)
**Date:** 2026-03-09
**Status:** Accepted

**Context:**
Need an automation platform to orchestrate the post-signup welcome email drip campaign. The platform receives webhook events from the backend (user.created, onboarding.step_completed, onboarding.completed) and triggers email sequences via Postmark. Options evaluated: n8n, Make (Integromat), Zapier, GCP Cloud Functions.

**Decision:**
Use n8n (self-hosted) as the automation platform:
- Free and open-source (no per-operation costs)
- Self-hosted on existing Hetzner infrastructure (Docker deployment)
- Visual workflow builder for non-technical iteration on drip campaigns
- Full data control — user data never leaves our infrastructure (GDPR-friendly)
- Built-in webhook receiver for bidirectional communication
- 400+ integrations including Postmark, Slack, and custom HTTP

Architecture also supports a **custom webhook escape hatch**: admins can configure any external API URL (Customer.io, Brevo, ActiveCampaign, etc.) instead of n8n, with the same HMAC-signed event payloads.

**Consequences:**
- Zero recurring cost for automation (vs $20+/mo for Zapier)
- Requires Docker hosting and maintenance on Hetzner
- Visual builder enables rapid drip campaign iteration without code changes
- Custom webhook option prevents vendor lock-in
- Need to set up n8n backup/monitoring alongside existing infra

---

### ADR-016: Postmark for Transactional Email
**Date:** 2026-03-09
**Status:** Accepted

**Context:**
Need a transactional email provider for the welcome drip campaign and future notification emails. Options evaluated: SendGrid, Postmark, Resend, Firebase Extensions.

**Decision:**
Use Postmark as the transactional email provider:
- Best-in-class delivery speed (<1 second to inbox)
- Pure transactional focus — no marketing email contamination of sender reputation
- Dedicated IP reputation (not shared with bulk mailers like SendGrid's free tier)
- B2B SaaS industry standard for welcome/notification/alert emails
- Clean analytics: open rate, bounce rate, spam complaints
- $15/month for 10,000 emails (sufficient for early-stage B2B SaaS)
- Mustache-based templates with visual preview

Postmark will be called from n8n workflows (or directly from custom webhook endpoints for third-party integrations).

**Consequences:**
- $15/month cost (vs free tiers of SendGrid/Resend) — justified by delivery quality
- Superior inbox placement due to dedicated transactional focus
- Need to set up Postmark account, verify sending domain, create email templates
- Templates managed in Postmark dashboard (not in codebase)
- If email volume grows significantly, can negotiate volume pricing

---

### ADR-017: Stripe for Billing (Not Shopify Billing)
**Date:** 2026-03-09
**Status:** Accepted

**Context:**
LedgerGuard needs a billing system for its own SaaS subscription plans (FREE, PRO, ENTERPRISE). Evaluated Stripe vs Shopify Billing API.

Shopify Billing was ruled out after confirming with Shopify documentation:
1. Shopify Billing API is exclusively for merchant-facing apps installed in shops — it cannot bill Partner organizations
2. Partner API exposes no billing endpoints
3. No Partner Tools marketplace exists
4. Manual token users (admin flow) have no OAuth/app installation — Shopify Billing impossible for them
5. All comparable partner-facing tools (Baremetrics, Ship) use external billing

**Decision:**
Use Stripe as the billing provider:
- Stripe Checkout (hosted) for payment collection — no PCI compliance needed
- Stripe Customer Portal for self-service plan management
- Webhook-driven state sync (checkout.session.completed, invoice.paid, etc.)
- Subscription lifecycle: Trial (14 days) → Free (limited) → Pro (paid) → Enterprise (custom)
- `billing_subscriptions` table (not `subscriptions` to avoid Shopify data collision)

**Consequences:**
- Works for all users regardless of Shopify connection method (OAuth or manual token)
- Stripe fees (~2.9% + 30c per transaction) vs 0% Shopify rev share avoided
- Full control over subscription lifecycle, trials, proration, coupons
- Need Stripe account setup, webhook endpoint, PCI-compliant checkout (handled by hosted page)
- Database-driven `plans` + `plan_features` tables for admin-editable feature gating

---

### ADR-018: Trial-Freemium Billing Model
**Date:** 2026-03-09
**Status:** Accepted

**Context:**
Need to decide on pricing model for LedgerGuard SaaS. Options: pure freemium, trial-only, or trial + freemium hybrid.

**Decision:**
All-paid model with trial on the base plan. No free tier.
- **Signup:** 14-day free trial with all Starter features unlocked
- **Trial expires + paid:** plan_tier = STARTER (or PRO if user chose Pro during checkout)
- **Trial expires + no payment:** plan_tier = EXPIRED, read-only mode (view last-synced dashboard, no sync/chat/export)
- **STARTER:** Dashboard, risk analytics, sync, notifications, 1 app
- **PRO:** + AI Chat, API Keys, Slack, data export, unlimited apps
- **ENTERPRISE:** + custom risk rules, priority support, SLA

Feature configuration stored in database (`plans` + `plan_features` tables) for runtime admin editability.

**Consequences:**
- Trial shows value before requiring payment — higher conversion than paywall
- No free tier simplifies the business model and ensures revenue from all active users
- Read-only mode preserves data and keeps users engaged (not locked out entirely)
- Database-driven features allow plan changes without code deploy
- Need daily cron to check trial expiry and downgrade to EXPIRED
- Frontend needs `BillingBloc` to gate UI elements per plan tier

### ADR-019: Auto-Deduct Billing Behaviors
**Date:** 2026-03-09
**Status:** Accepted

**Context:**
Need to document how recurring charges, payment failures, trial conversions, and RBI compliance are handled automatically.

**Decision:**
Four auto-deduct behaviors in the billing system:

1. **Auto-renewal:** Stripe auto-charges saved payment method each billing cycle (monthly/annual). No user action or backend cron needed — Stripe handles invoicing and charging.
2. **Auto-downgrade on failure:** After Stripe exhausts smart retries (~2 weeks), daily cron detects `past_due > 7 days`, cancels subscription, sets `plan_tier = EXPIRED`. User enters read-only mode automatically.
3. **Auto trial-to-paid:** If user adds payment method during trial, Stripe auto-creates subscription at trial expiry. Seamless transition with no interruption. If no payment method → `plan_tier = EXPIRED`.
4. **RBI auto-debit mandate (India):** Stripe India handles e-mandate registration. Charges ≤ ₹15,000 auto-debit without extra auth. Charges > ₹15,000 require customer approval via bank/UPI (24h pre-debit notification). LedgerGuard Pro ($29 ≈ ₹2,400) is well under threshold.

**Consequences:**
- Minimal backend logic for recurring billing — Stripe does the heavy lifting
- Daily cron only needed for: trial expiry, past-due cleanup, scheduled downgrades
- RBI compliance handled by Stripe India entity — no custom mandate code needed
- Enterprise plans with custom pricing may exceed ₹15,000 — Stripe's mandate flow handles this

---

### ADR-020: Razorpay as Primary Payment Provider (India)
**Date:** 2026-03-18
**Status:** Accepted

**Context:**
Stripe India is invite-only and we don't have access yet. We need a payment provider for B2B SaaS billing (Starter $249/mo, Pro $499/mo).

**Decision:**
Use Razorpay Subscriptions as the primary payment provider. Key choices:
1. **BillingSubscription is a separate entity** from Shopify Subscription — different domain concepts
2. **Domain layer has zero Razorpay deps** — RazorpayClient lives in infrastructure/external
3. **Razorpay plan IDs stored in config** — created via Razorpay dashboard first, referenced by env vars
4. **Webhook handler returns 200 always** — logs processing errors but prevents Razorpay retry storms
5. **Billing routes are optional** — server starts gracefully without Razorpay config
6. **HMAC-SHA256 for webhook verification** — standard Razorpay signature validation
7. **Test mode only** — no GST calculations or live mode config needed yet

**Consequences:**
- Can start accepting payments immediately via Razorpay test mode
- Clean separation from Stripe code (future migration path preserved)
- No vendor lock-in at domain layer — only infrastructure layer knows about Razorpay
- Will need to add GST/tax handling before going live in India

---

### ADR-021: Snapshot Sync as Separate Job Type
**Date:** 2026-04-21
**Status:** Accepted

**Context:**
In the queue-based sync system design, historical snapshot backfill (`LedgerService.BackfillHistoricalSnapshots()`) was initially embedded inside `transaction_sync`. This meant snapshot computation had no independent progress visibility — it was hidden as a silent sub-step of transaction processing. During plan review, we identified three problems:
1. **No progress visibility** — frontend shows "transactions: 1200/1200 done" but snapshot backfill (iterating 12 months) runs invisibly afterward
2. **No independent rerun** — if a metrics bug is fixed, you'd have to re-fetch all transactions just to recompute snapshots
3. **Mixed responsibilities** — `transaction_sync` processor was doing two distinct things: fetching transactions + rebuilding ledger, then backfilling snapshots

**Decision:**
Extract snapshot backfill into a separate `snapshot_sync` job type. Full sync now dispatches 6 children (not 5):
- **Wave 1 (independent):** `transaction_sync`, `event_sync`, `review_sync`
- **Wave 2 (depends on transaction_sync):** `snapshot_sync`, `status_sync`, `store_sync`

`snapshot_sync` calls `LedgerService.BackfillHistoricalSnapshots()` with progress reporting: "month 3/12". It depends on `transaction_sync` because it needs rebuilt subscriptions and transactions to compute monthly metrics.

**Consequences:**
- Granular progress: frontend shows "Snapshots: 7/12 months" as a separate progress bar
- Independent recompute: can re-run `snapshot_sync` alone after a metrics formula change without re-fetching transactions
- Cleaner processors: each processor has a single responsibility
- One additional child job per full sync (6 instead of 5) — negligible overhead
- `transaction_sync` is now faster (no snapshot backfill step)

---

### ADR-022: Redis Queue-Based Async Sync System
**Date:** 2026-04-21
**Status:** Accepted

**Context:**
The synchronous `POST /api/v1/sync/{appID}` endpoint blocks the HTTP request for the entire sync duration (transaction fetch, ledger rebuild, snapshot backfill, status enrichment, brand fetch, review scrape). This prevents showing granular progress in the Flutter frontend, can time out for large apps, and doesn't scale for many concurrent syncs.

**Decision:**
Implement an async, queue-based sync system using Redis + PostgreSQL:
- **Redis LPUSH/BRPOP queues** for job distribution (two queues: regular + full_sync)
- **PostgreSQL `sync_jobs` table** for durable job tracking and status
- **Worker pools** with configurable concurrency (default 3 regular + 1 full sync)
- **Distributed locks** (SETNX) prevent duplicate processing per app+type
- **Dual-write progress** (Redis 2s for fast polling, DB 30s for durability)
- **Recovery service** re-enqueues stuck jobs on startup and periodically (10min)
- **Feature flag** `QUEUE_ENABLED=false` keeps the system fully disabled by default
- **Existing sync endpoints untouched** — new endpoints live alongside at `/api/v1/sync/enqueue/`

**Consequences:**
- Frontend can poll `/sync/jobs/{jobID}/progress` for granular, real-time progress
- Full sync orchestrates 6 child jobs in 2 waves (parallel where possible)
- Requires Redis in production (adds operational complexity)
- Feature-flagged — zero impact when disabled, no Redis dependency
- Cooperative cancellation via Redis flag checked between major steps
- Recovery handles Redis flush and worker crashes gracefully

**Reusable Blueprint:** The patterns from this ADR (and ADR-023 through ADR-028) were extracted into `docs/blueprints/SYNC_PIPELINE_BLUEPRINT.md` — a language-agnostic blueprint covering all 14 production bug fixes, suitable for dropping into any new project.

---

### ADR-023: Cooperative Cancellation over Hard Kill
**Date:** 2026-04-21
**Status:** Accepted

**Context:**
When a user cancels a running sync job, we need to decide between hard kill (context cancel) and cooperative cancellation (flag check).

**Decision:**
Use cooperative cancellation: set a Redis flag (`lg:sync:cancel:{jobID}`), and each processor checks `IsCancelled()` between major steps. The worker does not forcefully terminate the goroutine.

**Consequences:**
- Clean shutdown — processors finish their current atomic step before stopping
- No partial writes or corrupted state from mid-operation cancellation
- Slightly delayed cancellation (up to the time between checkpoint checks)
- Simpler error handling — no need to distinguish "cancelled" from "crashed"

### ADR-024: SyncTrigger Interface for Auto-Sync on App Selection
**Date:** 2026-04-23
**Status:** Accepted

**Context:**
After onboarding (OAuth → app selection), no sync runs automatically. The user sees an empty dashboard until the 12h scheduler fires or they manually trigger a sync. This creates a poor first experience.

**Decision:**
Define a `SyncTrigger` interface in the handler layer with a single `TriggerSync(ctx, appID, userID, partnerAccountID)` method. Both `QueueSyncService` (enqueue with priority=1) and `SyncService` (fire-and-forget goroutine) implement it. The handler uses setter injection (`SetSyncTrigger`) to avoid constructor changes. Duplicate job errors from the queue are silently swallowed (already running = not an error).

**Consequences:**
- New users see data within minutes of selecting an app
- Fire-and-forget: `SelectApp` returns immediately, sync runs in background
- Interface in handler layer keeps domain clean (no circular dependency)
- Setter injection follows existing pattern (`SetShopRepo`)
- Queue mode gets a high-priority job; direct mode starts a background goroutine

---

### ADR-025: EventTracker Interface for Mixpanel Analytics
**Date:** 2026-05-04
**Status:** Accepted

**Context:**
The platform lacks visibility into user lifecycle events (signup funnel, sync health, billing conversions). Data exists in audit tables but is not pushed to any external analytics tool. We need server-side event tracking that works with Mixpanel but doesn't couple the domain layer to a specific vendor.

**Decision:**
Define an `EventTracker` interface in the domain service layer with `Track()` and `SetUserProperties()` methods. Provide two implementations: `MixpanelClient` (uses Mixpanel HTTP API directly with fire-and-forget goroutines) and `NoopTracker` (silent no-op when `MIXPANEL_TOKEN` is empty). No external SDK dependency — just `net/http` + JSON. Services receive the tracker via setter injection (`SetTracker`), following the existing `SetShopRepo`/`SetSyncTrigger` pattern.

**Consequences:**
- Domain layer has zero vendor dependency (interface only)
- Fire-and-forget: analytics calls never block request handling
- NoopTracker enables clean dev/test without external services
- No new Go dependencies — uses stdlib `net/http` and `encoding/json`
- Setter injection avoids constructor signature changes in existing services
- Adding new events is trivial: one `tracker.Track()` call at the trigger point

---

### ADR-026: Ownership-Aware Distributed Locks via Lua Scripts
**Date:** 2026-05-05
**Status:** Accepted

**Context:**
The queue system used simple `DEL` for lock release and `EXPIRE` for lock extension, without verifying who holds the lock. In a multi-node deployment, Worker A could release Worker B's lock after slow processing. Lock steal was also non-atomic (separate release + acquire = race window). Recovery could release locks belonging to active workers on other nodes.

**Decision:**
All lock operations now use Lua scripts for atomicity:
- `ReleaseLockIfOwner`: Only DEL if Redis value matches the requesting workerID
- `ExtendLockIfOwner`: Only PEXPIRE if value matches ownerID
- `StealLock`: Atomic check-current-holder → DEL → SET NX in a single script
- `ForceReleaseLock`: Unconditional DEL retained for recovery (legitimately clearing stale locks from crashed nodes)
- `GetLockHolder`: Read current lock value for diagnostics

**Consequences:**
- Eliminates lock-release-without-ownership bugs across all multi-node scenarios
- Lock steal is now race-free (single Lua script = atomic Redis operation)
- Recovery can only force-release; workers always use ownership-aware release
- Minor performance overhead from Lua eval (negligible for lock operations at this frequency)

---

### ADR-027: Centralized Job State Transitions in Worker
**Date:** 2026-05-05
**Status:** Accepted

**Context:**
All 7 processors called `MarkCompleted` themselves after successful processing. If the DB write for `MarkCompleted` failed (network blip, timeout), the worker would then call `MarkFailed` — marking a successfully-processed job as failed. Additionally, `MarkStarted` was called before lock acquisition, causing jobs to bounce processing→pending on lock failure. Cancelled jobs were being overwritten to "failed" status.

**Decision:**
Centralize all job state transitions in the worker:
1. Lock acquired BEFORE `MarkStarted` (avoids processing→pending bounce)
2. Worker calls `MarkCompleted` on processor success (nil return)
3. Worker calls `MarkFailed` on processor failure (non-nil return)
4. Before `MarkFailed`, check `IsCancelled` — skip if already cancelled
5. Processors return nil on success instead of calling MarkCompleted
6. DB status guards: `MarkStarted` requires `status='pending'`, `MarkCompleted` requires `status='processing'`, `MarkFailed` requires `status IN ('pending','processing')`

**Consequences:**
- Single point of control for state machine transitions
- DB failures on completion don't corrupt job state
- Cancelled jobs are never misclassified as "failed"
- Processors are simpler (no repo dependency for status changes)
- Status guard prevents concurrent workers from double-transitioning a job

### ADR-028: Graceful Shutdown Recovery (Leave Processing, Don't Fail)
**Date:** 2026-05-05
**Status:** Accepted

**Context:**
On Ctrl+C, `processor.Process(ctx, payload)` returns `context.Canceled`. The worker was marking these jobs `failed` — a terminal state that recovery ignores. Jobs interrupted by shutdown never restarted.

**Decision:**
When `ctx.Err() != nil` (shutdown signal), skip `MarkFailed`. Leave the job in `processing` state with heartbeat deleted and lock released. On next boot, `RecoverOnStartup` finds it (no heartbeat = stale) and re-enqueues immediately.

**Consequences:**
- Graceful shutdown → instant recovery on next boot
- No data loss from transient restarts/deployments
- Initial heartbeat written synchronously after lock (prevents steal race during the window)
- Recovery skips child jobs if parent is also recovered (parent recreates them fresh)

### ADR-029: event_sync in Wave 2 (Subscription Dependency)
**Date:** 2026-05-05
**Status:** Accepted

**Context:**
`EventProcessor` iterates subscriptions to get shop GIDs and calls `FetchAppEvents` per shop. Subscriptions are only created during `transaction_sync` (ledger rebuild). Running event_sync in Wave 1 (parallel with transaction_sync) always produces 0 events.

**Decision:**
Move event_sync from Wave 1 to Wave 2. Wave 1 = [transaction_sync, review_sync]. Wave 2 = [event_sync, snapshot_sync, status_sync, store_sync].

**Consequences:**
- event_sync correctly finds subscriptions after ledger rebuild
- Wave 1 is smaller (2 jobs) but transaction_sync dominates runtime anyway
- Future: can move event_sync back to Wave 1 by removing subscription dependency (fetch all events without shop filter)

---

### ADR-030: CQRS Read Models for Revenue API
**Date:** 2026-05-06
**Status:** Accepted

**Context:**
The Revenue API exposes subscription and usage status via external API key auth. Initially, read model tables (`api_subscription_status`, `api_usage_status`) were created but never populated — `ReadModelBuilder.RebuildForApp()` existed but was never called from the sync pipeline.

**Decision:**
Wire `ReadModelBuilder` as a non-fatal post-sync step in both sync paths:
- **Queue path:** `TransactionProcessor` calls `ReadModelBuilder.RebuildForApp()` after ledger rebuild via `WithReadModelBuilder()` setter
- **Direct path:** `SyncService` calls it after ledger rebuild via `WithReadModelBuilder()` setter
- **Admin safety net:** `POST /api/v1/admin/apps/{appID}/rebuild-read-model` allows manual triggering
- Read model errors are logged but do not fail the sync

**Consequences:**
- Revenue API endpoints now return real data after every sync
- Non-fatal: sync succeeds even if read model population fails
- Admin endpoint provides manual recovery without re-running full sync
- Two wiring points in main.go (sync service + admin handler)

---

### ADR-031: Dual Subscription Identity (ShopifyGID + StableDomainKey)
**Date:** 2026-05-06
**Status:** Accepted

**Context:**
Shopify subscriptions get a new GID on every reinstall. When a store uninstalls and reinstalls an app, the old subscription is cancelled and a new one is created with a different `shopify_gid`. This makes it impossible to track churn-return patterns across reinstalls using the GID alone.

**Decision:**
Introduce dual identity for subscriptions:
- `shopify_gid`: The real Shopify subscription GID, mapped from Partner API `chargeId` field on `AppSubscriptionSale` transactions. Changes on reinstall.
- `stable_domain_key`: Deterministic key `lg_sub_` + SHA1(`myshopify_domain`). Survives reinstalls because it's derived from the store's domain, which is permanent.

Migration 000035 adds the `stable_domain_key` column to the `subscriptions` table.

**Consequences:**
- Enables future churn-return analysis (same domain, different GID = reinstall)
- `shopify_gid` now contains the real Shopify GID (previously it was a synthetic key)
- `stable_domain_key` is computed deterministically — no external lookup needed
- SHA1 is used for compactness, not security (domain names are not secret)

---

### ADR-033: ON CONFLICT Includes app_id for Transaction Upsert
**Date:** 2026-05-07
**Status:** Accepted

**Context:**
The transaction upsert (`ON CONFLICT (shopify_gid)`) did not update `app_id` when a conflicting row existed. If a Shopify app is deleted and re-created, its internal `app_id` changes but the underlying transaction GIDs remain the same. The upsert would keep the stale `app_id`, causing orphaned transactions that don't appear under the new app.

**Decision:**
Include `app_id = EXCLUDED.app_id` in the `ON CONFLICT ... DO UPDATE SET` clause for both `Upsert()` and `UpsertBatch()` in `transaction_repository.go`. This ensures the transaction is always associated with the currently-syncing app.

**Consequences:**
- Transactions are correctly re-parented when an app is re-created with the same Shopify GID space
- No data loss — all other fields were already being updated on conflict
- Idempotent: re-running the same sync produces the same result
- No migration needed — query-only change

---

### ADR-032: Daily Catch-Up Sync with Configurable Lookback
**Date:** 2026-05-07
**Status:** Accepted

**Context:**
The queue-based sync system runs full syncs on a 12-hour schedule, but if a sync fails, the server is restarted, or the scheduler misses a window, recent transactions can go unsynced until the next full cycle. For a revenue intelligence platform, even a few hours of missing data is problematic — risk states and MRR calculations become stale. A lightweight mechanism is needed to fill these gaps without the overhead of a full 12-month ledger rebuild.

**Decision:**
Add a daily catch-up sync that enqueues `transaction_sync` + `event_sync` jobs with a short lookback window (default 2 days) for all active apps. Key design choices:

1. **`LookbackDays` on `SyncJobPayload`** — backward-compatible field (0 = default 1-month window). `TransactionProcessor` uses it when > 0 to narrow the fetch window.
2. **`DailyCatchupScheduler`** — follows the `NotificationScheduler` pattern: checks every 15 minutes, fires at a configurable UTC hour (default 3 AM), tracks `lastRunDate` to prevent double-runs.
3. **Selective job types** — only `transaction_sync` + `event_sync`, not a `full_sync`. Snapshots, status, store, and review syncs are unnecessary for a 2-day window.
4. **Duplicate-safe** — `EnqueueCatchupSync` silently skips apps that already have an active job of the same type (uses existing duplicate detection).
5. **Admin + internal endpoints** — `POST /api/v1/admin/sync/daily-catchup?lookback_days=N` for manual trigger; `POST /api/v1/internal/sync/daily-catchup` for Cloud Scheduler.

**Consequences:**
- Fills data gaps from missed or failed syncs without full 12-month rebuild
- 2-day lookback is cheap (few API pages vs hundreds for full sync)
- Runs at 3 AM UTC by default — off-peak for most Shopify partners
- Does not replace the full sync scheduler — complementary mechanism
- `RunOnce(ctx, lookbackDays)` enables manual/admin-triggered catch-ups with custom windows
- Internal endpoint enables Cloud Scheduler integration for serverless environments (Cloud Run)

---

### ADR-034: Organizations as Top-Level Data-Owning Entity
**Date:** 2026-05-07
**Status:** Accepted

**Context:**
LedgerGuard currently uses a single-user model where partner accounts, billing, and API keys are owned by individual users. App developers often work in teams — co-founders, VAs, bookkeepers need access to the same data. A multi-user model is needed without breaking the existing single-user flow.

**Decision:**
Introduce `organizations` as the top-level data-owning entity. Key design choices:

1. **Organizations own data, not users** — `partner_accounts`, `billing_subscriptions`, and `api_keys` get `org_id` columns. Data is accessed through org membership.
2. **Multi-org support** — A user can belong to multiple organizations (e.g., freelancer managing apps for different companies). The `X-Org-Id` header or auto-selection (when user has exactly 1 org) determines the active org.
3. **Three roles** — OWNER (full control), ADMIN (manage members/apps/sync), VIEWER (read-only). Role hierarchy: OWNER > ADMIN > VIEWER.
4. **Member lifecycle** — INVITED → ACTIVE → SUSPENDED → REMOVED. Suspended members cannot access data but still count toward plan limits.
5. **Plan-based limits** — FREE=1 member, STARTER=3, PRO=10. Pending invitations count toward limits.
6. **Zero-downtime migration** — Add nullable `org_id` first, backfill existing users into personal orgs, then make NOT NULL.
7. **Org-scoped audit log** — Separate from user-level audit log. Records member lifecycle events, sync triggers, setting changes.
8. **Org-level webhooks** — STARTER+ plans can configure a webhook URL with HMAC-SHA256 signing.
9. **SSO/SAML** — PRO only, stored as JSONB config on the organization.

**Consequences:**
- Existing single-user flow works unchanged (auto-created personal org, auto-selected)
- New middleware (`OrgContextMiddleware`) resolves org context on every request
- All data access routes become org-scoped via membership check
- Migration 000036 creates 4 new tables and adds 3 `org_id` columns
- Plan tier moves from `users` to `organizations` table (after backfill)
- API keys become org-scoped: any org member with ADMIN+ can manage

---

### ADR-035: Org-Scoped Data Access via resolvePartnerAccount + FindByOrgID
**Date:** 2026-05-08
**Status:** Accepted

**Context:**
ADR-034 introduced organizations but all 26 data endpoint call sites still resolved partner accounts via `FindByUserID`. Multi-org data isolation was not enforced.

**Decision:**
1. **`resolvePartnerAccount(r, partnerRepo)` helper** — centralized function that tries org-based lookup first (`FindByOrgID`), falls back to user-based (`FindByUserID`). Located in `app_lookup.go`.
2. **`FindByOrgID` on repository** — new method `WHERE org_id = $1` for org-scoped queries.
3. **OrgContextMW on data routes** — added to `/apps/*`, `/sync/*`, `/metrics/aggregate`, `/integrations/shopify/status`. Guarded with `if cfg.OrgContextMW != nil` for backward compatibility.
4. **Backend org persistence** — `selected_org_id` stored in `user_preferences` table (migration 000038), accessed via `GET/PUT /api/v1/user/preferences/selected-org`. Persists across devices/logins.
5. **Not changed** — background jobs (`sync_service.go`, schedulers) continue using `GetAllIDs`/`FindByID` since they have no HTTP context.

**Consequences:**
- Existing single-user flow works unchanged (OrgContextMW auto-selects single org)
- Multi-org users must pass `X-Org-Id` header or use URL param
- All data queries now flow through org membership verification
- Selected org persists server-side (no client-only storage needed)

### ADR-036: UUID-First App Identification (Deprecate Numeric Shopify IDs)
**Date:** 2026-05-08
**Status:** Accepted

**Context:**
Every API request paid a "translation tax": frontend sent numeric Shopify app ID (e.g., "5001") → backend constructed GID string (`gid://partners/App/5001`) → `FindByPartnerAppID()` (compound index) → finally got UUID for actual queries. Additionally: 5 duplicate lookup helpers, 4 duplicate GID prefix constants, numeric IDs are sequential/guessable, and `SyncStatusProvider` maintained a separate `_idToUuid` mapping hack.

**Decision:**
1. Frontend uses `ShopifyApp.id` = UUID (from `json['uuid']`), `shopifyId` for display only.
2. Backend `resolveAppFromRequest()` accepts UUID only — direct `FindByID()` (primary key lookup).
3. Numeric Shopify IDs kept only for: display labels, Shopify Partner API calls, Revenue API external consumers.
4. Auth/ownership checks handled by router-level middleware, not in `resolveAppFromRequest()`.

**Consequences:**
- Every app-scoped endpoint is one PK lookup instead of GID construction + compound index query
- ~200 lines of duplicate lookup code deleted
- Non-UUID app IDs in URL now return 400 "invalid app ID format"
- External Revenue API consumers unaffected (separate handler with its own GID lookup)

### ADR-037: Plan-Gated App Limits (STARTER=1, PRO=Unlimited)
**Date:** 2026-05-08
**Status:** Accepted

**Context:**
With organizations and plan tiers (FREE, STARTER, PRO) established in ADR-034, needed to enforce app limits per plan tier. Multi-app portfolio features ("All Apps" aggregate views) are only valuable for users with multiple apps.

**Decision:**
1. `Organization.MaxApps()` — domain entity method returns 1 for FREE/STARTER, 0 (unlimited) for PRO.
2. `SelectApp` handler checks app count vs limit, returns 403 if exceeded.
3. Frontend mirrors logic: `OrganizationModel.maxApps`, `isPro`, `canViewAllApps` helpers.
4. "All Apps" aggregate filter removed from 6 per-entity screens (subscriptions, stores, risk, earnings, transactions, events) — single-app context always.
5. "All Apps" gated behind PRO on Dashboard and Analytics screens only.
6. Partner Connect screen disables excess checkboxes for STARTER users with upgrade prompt.

**Consequences:**
- STARTER users see app-switcher dropdown, never "All Apps" option
- PRO users get portfolio-level KPIs on Dashboard/Analytics
- Partner Connect enforces limit at selection time, not just backend

---

### ADR-039: Forecasting Model Selection (Linear + Exponential)
**Date:** 2026-05-09
**Status:** Accepted

**Context:**
Revenue forecasting needed for the Analytics Forecasting tab. Must handle various MRR trajectories (growing, flat, declining) and provide confidence bands.

**Decision:**
Implemented two forecasting models in `ForecastingEngine`:
1. **Linear Regression (OLS)**: Fits ordinary least-squares line through daily MRR values. Simple, interpretable, works well for steady growth. Confidence bands: ±15%.
2. **Holt's Exponential Smoothing**: Double exponential smoothing with level + trend components. More responsive to recent changes. Alpha parameter (0.1–0.9, default 0.3) controls smoothing factor. Confidence bands widen over time (5% + 2%/month).

Both require minimum 90 data points. Frontend provides a toggle (SegmentedButton) to switch between models, which re-fetches from the API.

**Consequences:**
- Users can choose the model that best fits their revenue pattern
- Linear is the default (more conservative, easier to understand)
- Exponential responds faster to trend changes but can overfit to noise
- Confidence bands give visual indication of forecast uncertainty
- No database storage needed — forecasts computed on-demand from daily snapshots

---

### ADR-040: AppComparison Model for Multi-App Analytics
**Date:** 2026-05-09
**Status:** Accepted

**Context:**
The Multi-App analytics tab used `ShopifyApp` model (with installCount, avgRating) for display, but the backend aggregate API returns different fields (mrrCents, atRiskCents, subscriptionCount, renewalRate). Needed a clean separation.

**Decision:**
Created `AppComparison` model in `analytics_model.dart` matching the backend `AppMetricsSummary` response shape. In demo mode, mock `ShopifyApp` data is converted to `AppComparison` objects. This removed the dependency on `mockApps` and `mockSubscriptions` imports for the live data path.

**Consequences:**
- Clean type separation: `ShopifyApp` for app management, `AppComparison` for analytics comparison
- Multi-app tab now shows relevant metrics (MRR, at-risk, subscriptions) instead of install counts/ratings
- Demo mode still works by converting mock data to the same model

---

### ADR-041: Single Daily Table with Query-Time Downsampling
**Date:** 2026-05-09
**Status:** Accepted

**Context:**
`BackfillHistoricalSnapshots` was changed from monthly to daily granularity so users get ~365 daily snapshots on first sync (forecasting needs 90+). This raised questions: should we have separate daily/weekly/monthly tables? How to optimize the O(365*n) backfill? How should the UI consume different granularities?

**Decision:**
**No new tables.** 365 rows/year/app is trivial for PostgreSQL. Weekly and monthly views are derived at query time using `DownsampleSnapshots()` which picks the last snapshot in each period (end-of-week or end-of-month). A new `GET /api/v1/apps/{appID}/metrics/trend?granularity=weekly` endpoint and GraphQL `metricsTrend(granularity: WEEKLY)` parameter expose this to consumers.

**Why NOT separate tables:**
- Dual-write complexity (sync must write to 2-3 tables, failures create inconsistencies)
- CLAUDE.md mandates daily as the atomic unit ("Store daily snapshot")
- No storage pressure (365 rows/year is negligible)
- Schema changes require migrations for minimal benefit

**Backfill optimization:** Sorted transactions once O(n log n), then pointer-based advancement — each transaction visited exactly once. Combined with `UpsertBatch` for batch DB writes (365 → ~7 round-trips of 50).

**Future extensibility:** If query-time downsampling becomes a bottleneck at scale (1000+ apps, 3+ years), add a `frequency` column and pre-compute weekly/monthly rollup rows in the same table. Backward compatible — all existing rows get `frequency = 'daily'`.

**Consequences:**
- Single source of truth — daily snapshots serve all granularities
- No migration needed — same table, same schema
- Backfill performance: O(n log n + 365*k) instead of O(365*n)
- Frontend Revenue tab uses weekly granularity for cleaner chart (~26 points vs ~180)

---

### ADR-042: Three-State HTTP Error Model for DB Outages
**Date:** 2026-05-18
**Status:** Accepted

**Context:**
When Postgres was unreachable, the auth middleware returned HTTP 500 and `app_lookup.go` returned HTTP 404 for all errors. The frontend treated 404 as "no data" and 500 as a generic error, causing silent $0 dashboards and false onboarding wizards instead of informing the user about a service outage.

**Decision:**
Implement a three-state HTTP error model:
- **404** — Resource genuinely not found (sentinel errors `ErrPartnerAccountNotFound`, `ErrAppNotFound`)
- **503** — Backend service unavailable (DB connection errors, timeouts)
- **500** — Application bugs / unexpected errors

Auth middleware and `app_lookup.go` use `errors.Is()` against sentinel errors to distinguish 404 from 503. Frontend intercepts 503 at the **AppShell level** (not per-screen) to show a global "Service Temporarily Unavailable" UI with auto-retry.

**Alternatives rejected:**
- Per-screen 503 handling — required changes to every screen; shell-level is DRY
- Retry at API client interceptor level — too aggressive, retries every request silently
- Health-check polling — adds complexity; auto-retry on user action is simpler for MVP

**Consequences:**
- Clear user feedback when backend is down instead of misleading empty states
- Auto-retry (3× at 15s) recovers automatically when DB comes back
- All existing test mocks updated to use sentinel errors — prevents future regressions
- Shell-level intercept means zero changes needed when adding new screens
- Forecasting still uses raw daily data (unchanged)

---

### ADR-043: Migrate Backend off GCP to Hetzner (Co-Host on checkoutmate VPS)

**Date:** 2026-07-26
**Status:** Deployed — co-host POC live at https://api.ledgerspear.com (2026-07-26); GCP decommissioned

**Context:**
GCP staging (`ledgerspear`) cost ~₹2000/mo, dominated by an always-on Serverless VPC Access connector (min 2× e2-micro, ~₹1,100) plus an always-on Cloud SQL `db-f1-micro` (~₹800). The VPC connector existed only because Cloud SQL was private-IP-only — a cost with no analog on a single-box setup. Cloud Run itself (min-scale 0) was ~₹0.

**Decision:**
Move the Go backend to a Hetzner VPS running Docker Compose (postgres + redis + api + nginx + certbot), modeled on the proven checkoutmate deployment. **Roll out as a POC by co-hosting on the existing, near-idle checkoutmate VPS** (`46.224.203.174`), then graduate to a dedicated CX22.

Backend domain is **api.ledgerspear.com** (the owned domain; `ledgerguard.com` is parked on Afternic — not owned). Frontend stays on Firebase Hosting. The Go backend already reads all config from env vars, so no code changes were needed.

**Isolation for the co-host:** own containers (`ledgerguard-*`), volumes (`lg_*`), and internal network; joins checkoutmate's docker network only so its nginx can proxy; memory-limited (api 768m / pg 512m / redis 256m); host given a 2 GB swapfile (previously none). The only change to checkoutmate is two additive nginx server blocks + one cert.

**Alternatives rejected:**
- Dedicated Hetzner box now — cleaner isolation, but +€3.79/mo and slower to prove out; deferred to post-POC.
- Hetzner Managed Postgres (~€16/mo) — self-hosted Docker Postgres is cheaper and matches checkoutmate.
- Stay on GCP but drop the VPC connector (Direct VPC egress) — still pays for Cloud SQL + Cloud Run; Hetzner is ~80% cheaper overall.

**Consequences:**
- Hosting drops ~₹2000/mo → ~₹0 extra during POC (co-hosted), ~₹360/mo on a dedicated box later.
- Shared blast radius with checkoutmate production during POC — mitigated by mem limits + swap.
- Cost is env-based + Dockerized, so graduating to a dedicated box is copy-`.env` + `up -d` + DNS repoint (no lock-in).
- Scaffolding: `docs/HETZNER_MIGRATION_PLAN.md`, `docker-compose.prod.yml` (standalone), `deploy/cohost/` (co-host variant).
- Follow-ups: rotate the OpenAI key; repoint frontend prod to api.ledgerspear.com; update Shopify Partner app redirect URI; tear down GCP once verified.

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

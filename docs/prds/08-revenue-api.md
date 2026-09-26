# PRD: Revenue API (external) (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/revenue_api/` (handlers, services, middleware, read-model builder),
> `entity/api_key.go`, migrations 000012-000015, `frontend-flutter/lib/screens/api_keys/`.

## Problem & evidence

A Shopify app partner wants to **embed LedgerGuard's revenue/risk truth in their own systems** — support dashboards, billing portals, internal tools — by querying a subscription's or usage record's status programmatically, rather than logging into LedgerGuard.

- **FACT:** a complete external API subtree exists (`revenue_api/`) with API-key auth, CQRS read models, rate limiting, and audit — it's a deliberate product surface, not a side effect.
- **UNKNOWN — original demand evidence not recorded.** No tickets/usage metrics stored. Open question: validate demand + who's integrating — owner: Product.

## Target users

**INFERRED:** technical staff at a **Shopify app partner** (engineers/ops) integrating revenue status into their own apps/sites; keys are self-managed by the workspace **OWNER**.
- **Not the target:** merchants; end-customers; write/mutation use cases (the API is read-only status lookup — see Non-goals).

## Proposed solution (as-built behavior)

**FACT — endpoints** (`/v1/` prefix, `router.go:527-549`): subscription status by Shopify GID, by `?domain=`, and batch (≤100); usage status by GID, by `?subscription_id=`, and batch (≤100); plus a `POST /graphql` (skeleton only). Responses expose `risk_state`, `is_paid_current_cycle`, `months_overdue`, charge dates, `status` (`subscription_status.go:67-95`).

**FACT — key management** (`frontend-flutter/lib/screens/api_keys/`, `api_key_handler.go`): the OWNER creates/lists/revokes keys in the dashboard. Keys are `lgk_<32 hex>` (256-bit random), shown **once at creation**, stored as a **SHA-256 hash** (`api_key.go:36-49`); auth via `X-API-Key` or `Authorization: Bearer` (`api_key_auth.go:55-61`); revocation is a soft-delete (`revoked_at`) checked at validation.

**FACT — tenant isolation (the good news).** Unlike the internal app endpoints, this API performs an **explicit per-request ownership check**: key → user → PartnerAccount → apps, verifying `app.PartnerAccountID == the key owner's` (`subscription_status_service.go:147-167`); inaccessible IDs fall into a `not_found` array rather than leaking. *(Caveat: scoping is per-**user**, not per-**key** — see Open questions.)*

**FACT — rate limiting & audit.** Per-key sliding-window limit (default 60/min, configurable 1–1000) with `X-RateLimit-*` + `Retry-After` headers (`rate_limiter.go`); every request is asynchronously audited (key, endpoint, method, sanitized params, status, latency, IP, UA) to `api_audit_log` (`audit_logger.go`).

**FACT — data source.** Served from denormalized **CQRS read models** (`api_subscription_status`, `api_usage_status`) rebuilt per app after each ledger sync (`read_model_builder.go:43-63`), carrying a `last_synced_at`.

### Key screens

**No new wireframe** (as-built). Real surface: `frontend-flutter/lib/screens/api_keys/` (create/list-with-prefix-mask/revoke). The API itself is machine-facing (no UI).

## Platform & policy constraints

**FACT.** Read-only status over data already synced from Shopify; freshness is bounded by sync cadence (read models are stale between syncs, `last_synced_at` exposes it). Runs single-instance in Docker Compose on Hetzner — which currently masks the in-memory rate-limiter's multi-instance limitation (see Non-goals/Open questions).

## Pricing-tier impact

**UNKNOWN — no plan-gating found** for API access or rate tiers beyond the per-key configurable limit. Open question: is programmatic access a paid capability, and should rate limits be tiered? — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Additive external surface; read models rebuild from existing data. No migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Isolation guarantee:** 100% of cross-org access attempts return `not_found`/403 (never another org's data) — enforced by an automated test suite (which doesn't exist yet — see Non-goals).
2. **Freshness transparency:** every response's `last_synced_at` is ≤ the sync SLA; ≥95% within one catch-up window.
3. **Reliability:** rate limiting behaves correctly under the actual deployment topology (today single-instance) — 429s issued at the configured limit, 0 cross-instance leakage once multi-instance.

## Non-goals (deliberately absent in code)

1. **No write/mutation API** — read-only status lookup. **FACT.**
2. **No real-time push** — polling only; data is sync-stale, no webhooks/streaming to the partner. **FACT.**
3. **No multi-instance-safe rate limiting today** — in-memory store only (Redis stub commented out). **FACT.**
4. **No test coverage** — the entire `revenue_api/` subtree has **no `_test.go`** (notable for an external, security-sensitive API). **FACT.**

## Open questions

- **Per-user vs per-key scoping:** a user in multiple PartnerAccounts would have every key access all of them. Add explicit per-key `scope: [app_ids]`? Owner: Eng/Sec.
- **Rate limiter is in-memory + fails open** (allows the request on store error, `rate_limiter.go:59-63`) — wire the Redis store before horizontal scaling; decide fail-open vs fail-closed. Owner: Eng.
- **Zero test coverage** on auth/isolation/rate-limit/read-model — highest-priority gap for an external API. Owner: Eng.
- **GraphQL endpoint is a skeleton** — ship it or remove it (it's live but placeholder). Owner: Eng.
- **Audit log has no retention/TTL** — grows unbounded. Owner: Eng.
- **No key rotation** (must create-new + revoke-old, no atomic swap). Owner: Eng.
- API **versioning**: `/v1/` path but no version in responses — schema drift could silently break partners. Owner: Eng.

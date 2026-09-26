# Tech Design: Revenue API (external) (as-built)

**PRD:** docs/prds/08-revenue-api.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** This is an **externally-consumed** contract — partners' systems bind to the `/v1/` endpoints, response shapes, and `lgk_` key format. Breaking any of them breaks live integrations. The API-key auth + tenant-isolation logic is the highest-stakes code in the whole backfill (a bug here leaks another org's financials). New tables (`api_keys`, `api_subscription_status`, `api_usage_status`, `api_audit_log`).

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — request pipeline** (`router.go:527-549`, middleware order): `api_key_auth` → `rate_limiter` → `audit_logger` → handler. Auth (`api_key_auth.go:52-91`) reads `X-API-Key`/`Bearer`, SHA-256-hashes the raw key, looks it up by hash (`api_key_service.go:135-155`), rejects missing/revoked (soft-delete `revoked_at`). Keys are `lgk_<32 hex>` (256-bit), shown once at creation (`api_key.go:36-49`).

**FACT — endpoints** (`subscription_status_handler.go`, `usage_status_handler.go`): subscription status by GID / `?domain=` / batch(≤100); usage by GID / `?subscription_id=` / batch(≤100); plus `POST /graphql` (skeleton, `graphql/handler.go:76-91`). Responses (`subscription_status.go:67-95`) expose `risk_state`, `is_paid_current_cycle`, `months_overdue`, charge dates, `status`.

**FACT — tenant isolation (real).** `verifyAppAccess` (`subscription_status_service.go:147-167`) loads the app, loads the key-owner's PartnerAccount, and returns `ErrAppAccessDenied` unless `app.PartnerAccountID == partnerAccount.ID`. Batch endpoints route inaccessible IDs to `not_found` (no leak, no error). `getUserApps` is explicitly **single-partner-account-per-user** (`:127`).

**FACT — rate limiting** (`rate_limiter.go:20-79`): per-key sliding window (default 60/min, per-key configurable 1–1000), `X-RateLimit-*` + `Retry-After` headers, 429 over limit. Store is **in-memory** (`persistence/rate_limiter.go`; Redis stub commented). On store error it **allows the request** (fail-open, documented at `:59-63`).

**FACT — read models & audit.** CQRS: `api_subscription_status`/`api_usage_status` rebuilt per app after ledger sync (`read_model_builder.go:43-63`), carrying `last_synced_at`. Every request is asynchronously audited (key, endpoint, sanitized params, status, latency, IP, UA) to `api_audit_log` via a buffered worker (`audit_logger.go`, `audit_log_repository.go:64-83`).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **hashed-key auth → per-key rate limit → ownership-checked read-model lookup → async audit**, over a CQRS projection rebuilt at sync time. Happy path: see `docs/designs/08-revenue-api-sequence.puml` (validated `plantuml -checkonly`). Multiple middleware + a services layer + a projector interact — diagrammed.

### Data model

**FACT — new tables** (migrations 000012-000015): `api_keys` (SHA-256 `key_hash` unique, `user_id`, `rate_limit_per_minute`, `revoked_at`); `api_subscription_status` (`shopify_gid` unique, `app_id`, `risk_state`, `is_paid_current_cycle`, `months_overdue`, `status`, `last_synced_at`); `api_usage_status` (usage GID, subscription linkage, `billed`, `amount_cents`, `last_synced_at`); `api_audit_log` (`api_key_id`, endpoint, method, params JSONB, status, `response_time_ms`, ip, ua, `created_at`). **Migration:** additive; read models backfill on first sync. **No TTL** on audit log (unbounded growth).

### API & events

**FACT.** `/v1/` REST + `/v1/graphql` (skeleton). Auth header `X-API-Key` or `Bearer`. Error shape `{"error":{"code","message"}}`. Rate headers on all responses. No response-body version field (path-only `/v1/`). No outbound webhooks/events.
**Breaking-change check:** every field in the status responses + the `lgk_` prefix + the `/v1/` paths are an external contract; any rename breaks partners silently (no version negotiation).

## Alternatives considered

- **In-memory rate-limit store vs Redis.** As-built ships in-memory with a **commented Redis stub** — a recorded, deliberate "not yet" (`persistence/rate_limiter.go`). Correct only while single-instance (current Hetzner deploy). **Choose Redis before** any horizontal scaling (counters otherwise leak per-instance).
- **Per-user key scoping vs per-key `scope:[app_ids]`.** As-built scopes by the owning user's PartnerAccount (`getUserApps:127`), acknowledged as "currently single partner account per user." **Choose per-key scopes if** a user ever spans multiple PartnerAccounts, or to grant least-privilege keys.
- **CQRS read models vs querying the domain tables directly.** As-built denormalizes into `api_*_status` for fast external lookups + isolation from the domain model. **Choose direct queries if** staleness between syncs becomes unacceptable (accepting coupling + load).
- **Fail-open vs fail-closed rate limiter.** Code chose fail-open with an inline note that production may want fail-closed — a recorded, contestable decision.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Missing/invalid/revoked key | hash lookup / `revoked_at` | **FACT** 401 (`api_key_auth.go:52-91`) | issue/rotate key |
| Cross-org access attempt | `verifyAppAccess` | **FACT** 403 (single) / `not_found[]` (batch), no leak (`:147-167`) | — |
| Over rate limit | window count > limit | **FACT** 429 + Retry-After (`rate_limiter.go:72-76`) | back off |
| Batch > 100 IDs | size guard | **FACT** rejected (`subscription_status_handler.go:110`) | split batch |
| Read model missing GID | lookup miss | **FACT** `not_found`, not error | wait for sync |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. Per-user, not per-key, scoping.** A user in >1 PartnerAccount would have every key reach all of them; no least-privilege per-key `scope`. → open question (security).
- **D2. Rate limiter is in-memory + fail-open.** Multi-instance deploy → per-instance counters (limit effectively N×); on store error the request is allowed. Safe only single-instance. → wire Redis + decide fail-closed before scaling.
- **D3. Read-model staleness.** Status is only as fresh as the last sync (`last_synced_at`); a partner polling sees stale `risk_state`/`months_overdue` between syncs. → document SLA; consider on-write projection.
- **D4. Zero test coverage.** No `_test.go` anywhere in `revenue_api/` — auth, isolation, rate limit, audit, and read-model correctness are all unverified by CI. **Highest-priority gap** for an external, security-sensitive API. → build the suite.
- **D5. GraphQL is a live skeleton** (`graphql/handler.go:76-91`) — an active endpoint returning a placeholder; ship or remove.
- **D6. Unbounded audit retention** — `api_audit_log` grows forever (no TTL/rollup). → retention policy.
- **D7. No key rotation** — only create-new + revoke-old; no atomic swap or usage-based expiry. → open question.
- **D8. No response versioning** — path is `/v1/` but responses carry no version; schema drift breaks partners silently. → versioning policy.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. **Isolation suite (D4, highest priority):** with two orgs A/B, a key for A must return `not_found`/403 for every one of B's GIDs/domains — single and batch (metric #1). **TODO — does not exist.**
2. Freshness: response `last_synced_at` present and within SLA (metric #2). **TODO.**

**Adversarial (one per Failure-modes row — ALL currently TODO, no tests exist):**
- Invalid/revoked-key → 401 · cross-org → 403/not_found · over-limit → 429+headers · batch>100 rejected · missing GID → not_found. **TODO.**

**Adversarial — DIVERGENCE TODO:**
- D1 multi-partner-account key scope · D2 multi-instance limit leakage + fail-open behavior · D3 staleness window · D8 version drift. **TODO.**

**Manual verification (copy-paste):**
1. Create a key in the dashboard `/api-keys` (OWNER); copy the `lgk_...` once.
2. `curl -H "X-API-Key: lgk_..." "https://api.ledgerspear.com/v1/subscriptions/status?domain=<yourshop>.myshopify.com" | jq .`
3. Cross-org check: try a GID you don't own → expect 403/`not_found`.
4. Hammer past the per-minute limit → expect 429 + `Retry-After`; inspect `X-RateLimit-*`.
5. Revoke the key → confirm subsequent calls 401.

**External-platform reality:** the API IS the external surface. Auth/isolation/rate-limit are pure-local and MUST be unit-tested (currently aren't). End-to-end must be smoke-tested against the real deployment with a real key; multi-instance rate limiting can only be validated once >1 instance runs (not today).

## Pricing & policy touchpoints

**UNKNOWN — no tier gating** for API access or rate tiers (only per-key numeric limit). → Open question (Product). No Shopify-policy interaction (data already synced).

## Rollout

**N/A for the surface (shipped).** Deployed single-instance on Hetzner ([[hetzner-migration]]) — which is what makes the in-memory limiter currently acceptable. **Before scaling horizontally:** wire the Redis rate-limit store (D2) — this is a rollout blocker, not optional. Rollback = revert; keys/read-models persist.

## Observability

**FACT:** `api_audit_log` is a rich per-request record (status, latency, IP, UA), queryable by key/time/errors (`audit_log_repository.go:86-161`). **DIVERGENCE:** no aggregated dashboard/alert on 401/403/429 rates, per-key volume, or read-model staleness; no retention. On-call would query `api_audit_log` directly. → proposed observability + retention TODO tied to metrics #1/#3.

## Open questions

- **Build the test suite (D4)** — isolation first. Owner: Eng. **(highest priority)**
- **Wire Redis rate-limit store + decide fail-open/closed (D2)** before scaling. Owner: Eng.
- Per-key `scope:[app_ids]` (D1). Owner: Eng/Sec.
- GraphQL: finish or remove (D5). Owner: Eng.
- Audit retention/rollup (D6) + response versioning (D8) + key rotation (D7). Owner: Eng.
- Pricing/tier gating + rate tiers. Owner: Product.

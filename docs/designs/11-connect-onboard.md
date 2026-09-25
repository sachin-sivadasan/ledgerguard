# Tech Design: Connect & Onboard (as-built)

**PRD:** docs/prds/11-connect-onboard.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Handles **secrets** (Partner API tokens) — the encryption scheme + key management are hard to change once tokens are stored (a key change orphans existing ciphertext). Public connect/token/app API shapes bind the Flutter client. The `partner_accounts` credential store is the sensitive data model.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — connect flow** (`manual_token.go`, `app.go`): `POST /api/v1/integrations/shopify/token` (ADMIN + OrgContextMW) encrypts the token and upserts a `partner_account`; `GET /api/v1/apps/available` decrypts it and calls `shopify_partner_client.FetchApps(orgID, token)` (Partner GraphQL); `POST /api/v1/apps/select` enforces the plan app-limit (`app.go:157-172`), 409s on an already-tracked app, `appRepo.Create`s the app, and **synchronously enqueues** the first sync via `syncTrigger.TriggerSync`, returning `sync_triggered` (`app.go:198-219`). `DELETE .../token` revokes; `GET .../status` reports connected state.

**FACT — encryption** (`crypto/aes.go`): AES-256-GCM, 32-byte key required, random nonce per encrypt prepended to ciphertext, `gcm.Open` on decrypt. Token stored as `partner_account.EncryptedAccessToken []byte` (`:17`), decrypted only when calling the Partner API.

**FACT — demo mode** (`apps_provider.dart`, `demo_mode_coordinator.dart`): a `SharedPreferences` flag broadcast to all providers; mock data from `lib/mock_data/`. Backend-agnostic.

## Proposed design

**No redesign — documents the shipped design.** Pattern: **encrypt-at-rest credential store → decrypt-on-use Partner API calls → app upsert → synchronous sync enqueue**, with a frontend-only demo toggle. Happy path: see `docs/designs/11-connect-onboard-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — existing tables.** `partner_accounts` (`org_id`, `partner_id`, `encrypted_access_token []byte`, name, timestamps — one per org); `apps` (created on select, `partner_app_id`, name, `revenue_share_tier`, tracking flags). No new migration. **Key management:** the 32-byte AES key comes from config/env (`NewAESEncryptor`) — rotating it invalidates all stored tokens (no re-encryption path observed).

### API & events

**FACT.** Token: `POST/GET/DELETE /api/v1/integrations/shopify/token` (ADMIN+Org), responses carry a **masked** token (last 4). Status: `GET .../status`. Apps: `GET /apps/available`, `POST /apps/select` (→ `sync_triggered`), `GET /apps`. No events.
**Breaking-change check:** the select response (`uuid`, `revenue_share_tier`, `sync_triggered`) + the masked-token shape bind the client.

## Alternatives considered

- **Manual token entry vs Shopify OAuth.** As-built takes a pasted Partner API token (simplest path to data). **Choose OAuth if** self-serve, revocable, scoped install becomes a requirement — manual tokens are long-lived and broadly scoped. No OAuth alternative recorded.
- **No-probe-on-save vs validate-on-save.** As-built accepts and encrypts without a Partner API probe; validity surfaces on first `FetchApps` (502). **Choose validate-on-save** to fail fast (PRD metric #2). No rationale recorded for deferring.
- **Env-key AES vs KMS/secret-manager.** As-built uses a config-provided 32-byte key. **Choose a KMS if** key rotation / per-tenant keys / auditability are needed.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Missing token/partner_id | validation | **FACT** 400 (`manual_token_test.go`) | fix input |
| Non-admin / no org | ADMIN + OrgContextMW | **FACT** rejected by middleware | use admin |
| Invalid/expired token | Partner API call fails | **FACT** 502 on `GET /apps/available` (late), logged (`app.go:98-102`) | re-enter token |
| App already tracked | duplicate check | **FACT** 409 (`app.go:176-180`) | — |
| Plan app-limit reached | limit check | **FACT** rejected at select (`app.go:157-172`) | upgrade (nav = TODO) |
| Sync enqueue fails | `TriggerSync` err | **FACT** `sync_triggered=false`, WARN logged, select still 201 (`app.go:200-206`) | manual sync |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. No token probe on save** — an invalid token is stored and only fails on the first app fetch (502), a late/confusing signal. → validate-on-save (metric #2).
- **D2. First-sync outcome invisible** — the client learns the *enqueue* result (`sync_triggered`) but never the sync's success/failure; no completion callback/poll wired into onboarding. → surface first-sync progress (metric #3; sync status API exists per PRD #6).
- **D3. Upgrade/billing nav is a TODO** — hitting the app-limit has no working upsell path (`connect_shopify_screen.dart`). → wire it.
- **D4. Demo-off + no partner account → 404s** — demo is frontend-only, so a mis-toggled client shows raw 404s instead of a "not connected" state. → graceful empty state.
- **D5. AES key rotation orphans tokens** — no re-encryption path; rotating the env key breaks all stored tokens. → document/handle rotation (or KMS).
- **D6. Long-lived broadly-scoped tokens** — manual tokens aren't scoped/expiring like OAuth; a leak is high-impact until manually revoked. → open question (OAuth?).

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Activation happy path: token → available apps → select → `sync_triggered=true` → live data (metric #1). Partly covered by handler tests; end-to-end is TODO.
2. Fast-fail on bad token (metric #2): **TODO** — requires probe-on-save (doesn't exist).

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- AddToken success/missing-field/no-user/update, GetToken, RevokeToken, maskToken — `manual_token_test.go`. GetAvailableApps / SelectApp / ListApps — `app_test.go`. FetchApps success/GraphQL-error/HTTP-error/empty — `shopify_partner_client_test.go`. Encrypt/Decrypt round-trip — should exist for `crypto/aes.go` (verify).

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 invalid-token-on-save behavior · D2 first-sync outcome surfaced · D5 key-rotation handling. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/interfaces/http/handler/ -run 'Token|App' -v && go test ./pkg/crypto/ -v`
2. In the app: enter a **valid** Partner token → confirm apps list, select one → response `sync_triggered:true` → live data appears.
3. Enter an **invalid** token → observe it's accepted at save, then 502 on the apps step (documents D1).
4. Disconnect → confirm providers revert to demo.

**External-platform reality:** the Shopify Partner API is faked below `shopify_partner_client` in unit tests; a real token + org must smoke-test discovery + first sync end-to-end (the only way to validate token acceptance + app discovery for real).

## Pricing & policy touchpoints

**FACT:** plan app-limit enforced at select; **upsell nav is a TODO** (D3). Partner API version support-window applies to discovery ([[shopify-partner-api-gotchas]]). → Open question (Product): the upgrade path.

## Rollout

**N/A (as-built, shipped).** Additive. **Sensitive:** the AES key must be present + stable in the Hetzner env ([[hetzner-migration]]); rotating it strands tokens (D5) — treat as an operational runbook item. Rollback = revert; encrypted tokens persist.

## Observability

**FACT (partial):** token save/revoke and Partner API failures logged; `sync_triggered` returned. **DIVERGENCE:** no metric on connect success/failure rate, invalid-token rate, or first-sync completion. On-call greps Hetzner logs. → proposed observability TODO tied to metric #1.

## Open questions

- **Probe-validate token on save (D1)** — Owner: Eng/Product.
- **Surface first-sync outcome in onboarding (D2)** — Owner: Eng.
- **Wire upgrade/billing nav (D3)** — Owner: Product/Eng.
- **Graceful not-connected state (D4)** — Owner: Eng.
- **AES key rotation / KMS (D5)** — Owner: Eng/Sec.
- **OAuth vs manual long-lived tokens (D6)** — Owner: Product/Sec.

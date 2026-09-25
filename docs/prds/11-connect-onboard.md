# PRD: Connect & Onboard (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/interfaces/http/handler/{manual_token,app,integration_status}.go`,
> `entity/partner_account.go`, `pkg/crypto/aes.go`, `application/service/queue_sync_service.go`,
> `frontend-flutter/lib/screens/settings/connect_shopify_screen.dart`, `providers/apps_provider.dart`.

## Problem & evidence

Before LedgerGuard can show a partner anything, it must securely obtain their **Shopify Partner API credentials**, discover their apps, and pull the first data. This is the entry gate for the whole product.

- **FACT:** a complete connect → discover → select → auto-sync flow exists, plus a demo mode so the app is explorable before connecting.
- **UNKNOWN — original demand evidence not recorded.** Open question: is a guided onboarding wanted, or is the settings-screen flow sufficient? — owner: Product.

## Target users

**INFERRED:** the workspace **ADMIN** at a Shopify app partner performing initial setup (token endpoints are ADMIN-gated).
- **Not the target:** merchants; non-admin org members (they can view but not connect).

## Proposed solution (as-built behavior)

**FACT — flow** (`connect_shopify_screen.dart`): the admin enters a **Partner API token + partner/org ID** (+ optional name) → `POST /api/v1/integrations/shopify/token` encrypts & stores it → `GET /api/v1/apps/available` fetches the org's apps from the Partner API → the admin **multi-selects** apps → `POST /api/v1/apps/select` creates each App and **auto-triggers the first sync** (fire-and-forget) → the UI flips from demo to live via `DemoModeCoordinator`. **Disconnect** (`DELETE .../token`) revokes the token and reverts all providers to demo.

**FACT — security.** The token is stored **AES-256-GCM encrypted** (`EncryptedAccessToken []byte`, `partner_account.go:17`; `crypto/aes.go` with nonce-prepended ciphertext), decrypted on-demand only when calling the Partner API. Scope is **per-org** (org context, user-ID fallback); token endpoints require **ADMIN + OrgContextMW**.

**FACT — demo mode.** Frontend-only: a `SharedPreferences` flag (`apps_provider.dart`) broadcast by `DemoModeCoordinator` swaps every provider to hardcoded mock data (`lib/mock_data/`). The backend has no demo concept.

### Key screens

**No new wireframe** (as-built). Real surface: `frontend-flutter/lib/screens/settings/connect_shopify_screen.dart` (token entry → app multi-select → connected state + demo toggle).

## Platform & policy constraints

**FACT.** Depends on a valid Shopify **Partner API** token + organization ID; app discovery is a Partner GraphQL call (`shopify_partner_client.FetchApps`), subject to the Partner API version support-window ([[shopify-partner-api-gotchas]]). Plan **app-limit is enforced at selection** (`app.go:157-172`). Token is transmitted over TLS then encrypted at rest.

## Pricing-tier impact

**FACT:** the number of connectable apps is plan-limited (enforced at select), but the **upgrade/billing navigation is a TODO** (`connect_shopify_screen.dart` — hitting the limit has no working upsell path yet). Open question — owner: Product.

## Migration for existing users

**INFERRED — N/A.** New connection is additive; reconnect re-encrypts a fresh token. No migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Activation:** ≥ 80% of admins who enter a token successfully select ≥ 1 app and see live data within the session.
2. **Fast failure:** an invalid/expired token is surfaced to the user **at entry** (not silently later) in 100% of cases — requires probe validation (which doesn't exist yet).
3. **First-sync visibility:** ≥ 95% of first syncs report a completion/failure state back to the user (today it's fire-and-forget, server-logged only).

## Non-goals (deliberately absent in code)

1. **No token probe/validation on save** — validity is discovered on the first app fetch (502 on failure). **FACT.**
2. **No OAuth flow** — manual token entry only (no Shopify OAuth handshake). **FACT.**
3. **No multi-partner-account per org** — one PartnerAccount per org. **FACT.**
4. **No backend demo mode** — demo is client-only; a live-mode client with no partner account gets 404s. **FACT.**

## Open questions

- **No probe validation** (metric #2) — validate the token against the Partner API on save so bad tokens fail fast? Owner: Eng/Product.
- **Fire-and-forget first sync** (metric #3) — surface first-sync progress/failure in the UI? Owner: Eng.
- **Upgrade/billing TODO** — wire the app-limit upsell path. Owner: Product/Eng.
- **Demo-off + no partner account → 404s** — add a graceful "not connected" state instead? Owner: Eng.
- **Token mask shows last 4 chars** — minor info-leak consideration. Owner: Eng/Sec.
- Guided onboarding vs settings-screen flow? Owner: Product.

# Tech Design: Authentication (as-built)

**PRD:** docs/prds/15-auth.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Auth is the trust root — every request's identity flows from here. Swapping the identity provider (Firebase) is a major change; the `users.firebase_uid` mapping is load-bearing. The verification + route-guarding logic is security-critical.

> **As-built.** "Current state IS the design." Every claim cites a real path:line; the security-critical ones were verified against code. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — client** (`auth_provider.dart`): Firebase email/password signup/login/reset/logout; the SDK persists the session (`authStateChanges`). A Dio interceptor (`api_client.dart`) attaches `Authorization: Bearer <getIdToken()>` per request (SDK silently refreshes the ~1h token) and calls `signOut()` on a 401.

**FACT — backend** (`middleware/auth.go`, `external/firebase_auth.go`): `AuthMiddleware` extracts the Bearer token (401 if missing/malformed), verifies via `FirebaseAuthService.VerifyIDToken` → `client.VerifyIDToken(ctx, token)` (401 on failure), extracts `UID`+`email`. **JIT provisioning:** `FindByFirebaseUID` → if absent, `NewUser(uid, email)` with Role=OWNER, PlanTier=FREE (`entity/user.go`), plus non-blocking org auto-provision (#14). `*User` → context; `GET /api/v1/me` returns it.

**FACT (verified) — route protection is fail-closed.** Every route group registers *inside* `if cfg.XHandler != nil && cfg.AuthMW != nil { ... }` (`router.go:167,176,188,197,...`). If Firebase/`AuthMW` is nil (unconfigured), the routes are **not registered at all** (404) — never served without auth. `AuthMiddleware` is only built when `firebaseAuth != nil && userRepo != nil` (`main.go:856`).

**FACT (verified) — no revocation check.** `VerifyIDToken` uses `client.VerifyIDToken`, **not** `VerifyIDTokenAndCheckRevoked` (`firebase_auth.go:41`) — a revoked/signed-out token remains valid until natural expiry (~1h).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **Firebase SDK client auth → per-request Bearer attachment → Admin-SDK verification + JIT user provisioning**, with fail-closed route registration. Flow: see `docs/designs/15-auth-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — `users`** (migration 000001): `id, firebase_uid (UNIQUE, indexed), email, role, plan_tier, created_at, onboarding_completed_at`. No new migration. `firebase_uid` is the external-identity join key.

### API & events

**FACT.** Auth is header-based (no login endpoint on the backend — the client talks to Firebase directly). Backend exposes `GET /api/v1/me` (current user) behind `AuthMW`. All `/api/v1` app routes sit behind `AuthMW` (+ `OrgContextMW` when wired, #14). No backend logout endpoint (client `signOut()`).
**Breaking-change check:** the `me` response + Bearer scheme bind the client.

## Alternatives considered

- **Firebase Auth vs self-managed auth.** As-built delegates identity to Firebase (token verification, refresh, password reset) — less code, proven security. **Choose self-managed only if** leaving Firebase; the `firebase_uid` coupling is the cost.
- **Verify-only vs verify-and-check-revoked.** As-built verifies without revocation check (cheaper, no Firebase round-trip for revocation). **Choose check-revoked if** immediate session invalidation matters (compromised account, forced logout). A recorded trade-off by omission.
- **Enforce email verification vs not.** As-built lets unverified emails act immediately (lower friction). **Choose enforcement if** the invite-hijack (#14 D2) / spam-signup risk matters.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Missing/malformed Bearer | header parse | **FACT** 401 (tested `auth_test.go:66,85`) | re-auth |
| Invalid/expired token | `VerifyIDToken` err | **FACT** 401, logged (tested `:105`) | client refresh/re-login |
| New user | `FindByFirebaseUID` miss | **FACT** JIT create (tested `:170`) | — |
| Org provisioning fails | error path | **FACT** login proceeds, user org-less; backfill 000042 (tested `:268`) | migration |
| DB lookup error (non-notfound) | repo err | **FACT** 503, logged only (tested `:343`) | retry |
| Firebase unconfigured | `AuthMW == nil` | **FACT (verified)** routes not registered → 404 (fail-closed) | configure Firebase |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1 (security). No token-revocation check** — `VerifyIDToken` (not `...AndCheckRevoked`), so a signed-out/compromised session's token works until ~1h expiry. → enable `CheckRevoked` on sensitive paths.
- **D2 (security). No email verification** — unverified emails act immediately; compounds the #14 invite-hijack risk. → gate on `email_verified`.
- **D3. DB errors are logged-only (503)** — no alerting on auth-path DB failures. → add metric/alert.
- **D4. No backend logout / session revocation** — logout is client-only; combined with D1, there's no server-side "kill this session now." → optional revocation endpoint.
- **D5. `email` claim best-effort** (`token.Claims["email"].(string)`, ignores the ok) — a token without an email yields an empty-email user. → validate presence.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Reliability: valid token → user in context; invalid/expired → 401 (never 500) — **exist** (`auth_test.go`).
2. Session hygiene: a revoked session denied within TTL (metric #2) — **TODO** (needs D1).

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- Missing header, bad format, invalid token, existing user, new-user auto-create, org-provision (+ failure-doesn't-block), create-user error, DB-lookup error, default-org-name — **all exist** (`auth_test.go:66-343`, 11 tests).

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 revoked-token still-accepted · D2 unverified-email access · D5 email-less token. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/interfaces/http/middleware/ -run Auth -v`
2. `curl -H "Authorization: Bearer <validFirebaseIdToken>" http://localhost:8080/api/v1/me | jq .`
3. `curl http://localhost:8080/api/v1/me` (no header) → expect 401.
4. Sign out in Firebase, immediately reuse the old token within the hour → today it **still works** (documents D1).

**External-platform reality:** Firebase is the identity provider; unit tests fake `FirebaseAuthService` below the middleware (11 tests). Real token verification, refresh, and revocation behavior can only be smoke-tested against a real Firebase project + a real ID token.

## Pricing & policy touchpoints

**FACT:** users default to `PlanTier=FREE`. **UNKNOWN** how plan tier gates auth. GDPR: email + Firebase UID are PII (note for DPA/retention). → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Firebase config (client `firebase_options.dart` + server credentials) must be present in the Hetzner env ([[hetzner-migration]]) — absence fails **closed** (routes 404), which is safe but total. Security changes (D1/D2) would deny some currently-accepted sessions → stage with comms. Rollback = revert.

## Observability

**FACT (partial):** token-verification failures + provisioning errors logged. **DIVERGENCE:** no metric on 401 rate, verification-failure spikes, or provisioning-failure rate. On-call greps Hetzner logs. → proposed observability TODO tied to metric #1.

## Open questions

- **Enable token-revocation check (D1)** on sensitive paths. Owner: Eng/Sec.
- **Enforce email verification (D2).** Owner: Product/Sec.
- **Server-side logout/revocation endpoint (D4)?** Owner: Eng/Sec.
- **Alerting on auth-path DB errors (D3).** Owner: Eng.
- **Validate email-claim presence (D5).** Owner: Eng.
- Plan-tier's role in auth gating + PII/DPA review. Owner: Product.

# PRD: Authentication (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** / **INFERRED** / **UNKNOWN**.
> Success metrics are **PROPOSED** (future), not historical. Sources:
> `middleware/auth.go`, `external/firebase_auth.go`, `entity/user.go`, `handler/me.go`,
> `frontend-flutter/lib/providers/auth_provider.dart`, `data/.../api_client.dart`, `firebase_options.dart`.
> (Org auto-provisioning + OrgContextMW: PRD #14.)

## Problem & evidence

Every LedgerGuard user needs a secure, low-friction way to sign in, and the backend needs to trust who's calling. This is the identity layer beneath everything.

- **FACT:** Firebase email/password auth is fully implemented end-to-end, with a JIT-provisioning backend middleware and comprehensive middleware tests.
- **UNKNOWN — original demand evidence not recorded.** Open question — owner: Product.

## Target users

**INFERRED:** any LedgerGuard user (partner staff). First-time sign-in creates a User (Role=OWNER, PlanTier=FREE) and a personal org (#14).
- **Not the target:** merchants; anonymous/guest usage.

## Proposed solution (as-built behavior)

**FACT — frontend** (`auth_provider.dart`): email/password **signup** (`createUserWithEmailAndPassword` + display-name), **login** (`signInWithEmailAndPassword`), **password reset** (`sendPasswordResetEmail`), **logout** (`signOut` + clears Mixpanel). Session is managed by the Firebase SDK (`authStateChanges` stream; persists across restarts). **Email/password only** — no social providers. **No email-verification gate.**

**FACT — token attachment** (`api_client.dart`): a Dio interceptor calls `getIdToken()` on the current Firebase user and injects `Authorization: Bearer <token>` on every request; the SDK silently refreshes the ~1-hour token. On a **401**, the error interceptor calls `signOut()`.

**FACT — backend verification** (`middleware/auth.go`, `external/firebase_auth.go`): `AuthMiddleware` extracts the Bearer token (401 if missing/malformed), verifies it via the Firebase Admin SDK `VerifyIDToken` (401 on failure), extracts `UID`+`email`, then **JIT-provisions**: `FindByFirebaseUID` → if absent, `NewUser(uid, email)` (Role=OWNER, PlanTier=FREE) + auto-org (non-blocking, #14). The `*User` is put in context; `GET /api/v1/me` returns it.

### Key screens

**No new wireframe** (as-built). Real surfaces: `login_screen.dart`, `sign_up_screen.dart`, `forgot_password_screen.dart`.

## Platform & policy constraints

**FACT.** Auth is entirely **Firebase Authentication** (Admin SDK server-side, FlutterFire client-side); Firebase config in `firebase_options.dart` + a server credentials file. ID tokens are ~1h with SDK-managed refresh. The backend `AuthMiddleware` is **only applied when Firebase + the user repo are both initialized** (`main.go:856`) — a config dependency worth verifying doesn't leave routes open when unconfigured.

## Pricing-tier impact

**FACT (partial):** new users default to `PlanTier=FREE` on the user entity. **UNKNOWN** how plan tier drives auth-level gating. Open question — owner: Product.

## Migration for existing users

**INFERRED — N/A.** JIT provisioning handles new users; migration 000042 backfills org-less users (#14). No auth migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Auth reliability:** ≥ 99.9% of valid tokens verify successfully; invalid/expired → 401 (never 500) in 100% of cases.
2. **Session hygiene:** a signed-out/revoked user is denied within ≤ token TTL (today up to ~1h, since revocation isn't checked).
3. **Onboarding friction:** ≥ X% of signups reach first live data (ties to Connect, #11).

## Non-goals (deliberately absent in code)

1. **No social / SSO login** — email/password only. **FACT.**
2. **No enforced email verification** — users act immediately post-signup. **FACT.**
3. **No server-side token revocation check** — `VerifyIDToken` doesn't check revocation, so a revoked session's token is valid until expiry. **FACT.**
4. **No backend logout endpoint** — logout is client-side Firebase `signOut()` only. **FACT.**

## Open questions

- **No email verification** — enforce it before granting access (esp. given the invite-hijack risk in #14)? Owner: Product/Sec.
- **No token revocation check** (`CheckRevoked`) — a signed-out/compromised session stays valid up to ~1h; enable revocation checks? Owner: Eng/Sec.
- **Auth-middleware-conditional-on-config** — confirm that if Firebase isn't configured, routes are **closed**, not open (deployment-safety). Owner: Eng/Sec.
- **Non-notfound DB errors return 503** and are only logged — acceptable, or need alerting? Owner: Eng.
- Plan-tier's role in auth gating. Owner: Product.

# Security Findings — authorization & session hardening

**Type:** Epic / security · **Priority:** High · **Date:** 2026-09-26
**Source:** as-built PRD/design backfill (`docs/prds/`, `docs/designs/`)
**Confidence:** ✅ verified in code · ⚠️ verified-gap, exploitability needs one confirmation

## Summary
A capability-by-capability read of the backend surfaced a cluster of authorization
and session-management gaps. **The hard part is already right** — org-level tenant
isolation verifies membership (`OrgContextMW`), the external Revenue API checks app
ownership (`verifyAppAccess`), partner tokens are AES-256-GCM, Razorpay webhooks are
HMAC-verified, and route registration fails closed. The findings below are targeted
fixes, not a rearchitecture.

---

## S1 — App-scoped endpoints don't verify `appID` belongs to the caller's org  ✅ **[High — confirmed authenticated cross-tenant IDOR] — ✅ FIXED**
**Status:** **FIXED** — `resolveAppFromRequest` now resolves the caller's partner account and returns 404 when `app.PartnerAccountID != account.ID`, closing the leak for all callers (reports/forecast/dashboard/subscriptions/stores) at one chokepoint. Regression tests added: `TestResolveAppFromRequest_CrossOrg_Returns404` (leak guard) + `TestResolveAppFromRequest_SameOrg_Succeeds` (legit-flow guard), `app_lookup_test.go`. Branch `fix/s1-app-org-ownership`.

**Component:** `internal/interfaces/http/handler/app_lookup.go:54` (`resolveAppFromRequest`) — used by reports, forecast, dashboard, subscriptions, stores.

**Confirmed evidence chain (all verified):**
1. `resolveAppFromRequest` (`app_lookup.go:54-77`) parses `appID`, calls `appRepo.FindByID(appID)`, returns the app. It receives a `partnerRepo` param **but never uses it** — no comparison of `app.PartnerAccountID` to the caller's org/partner account.
2. `appRepo.FindByID` (`app_repository.go:46-54`) — `WHERE id = $1` only; no org/partner filter.
3. `subRepo.FindByAppID` (`subscription_repository.go:91-98`) — `WHERE app_id = $1 AND deleted_at IS NULL` only; no org filter (same shape for transactions/snapshots).

`OrgContextMW` only proves the caller is a member of *their own* org (auto-selected if they have one); it never ties the requested `appID` to that org.

**Impact:** any authenticated user who belongs to any org can read **another org's** subscriptions / transactions / metrics / reports / forecasts / stores via `/api/v1/apps/{victim_appID}/...`. Broken object-level authorization → cross-tenant financial-data exposure.

**Exploitability caveat:** `appID` is an unguessable UUIDv4 (~122 bits) — this is **not** blind enumeration; an attacker must obtain a target org's app UUID (leaked via logs, shared links/screenshots, `Referer`, support tickets, a former teammate). Authenticated IDOR gated by UUID knowledge, not a mass-scrapable hole. Still High — isolation must not rely on UUID secrecy.

**PoC (to demonstrate/close):**
```
curl -H "Authorization: Bearer <orgA_token>" -H "X-Org-Id: <orgA_id>" \
  "http://localhost:8080/api/v1/apps/<orgB_app_uuid>/subscriptions"
# VULNERABLE if this returns Org B's data instead of 403/404.
```

**Fix (single chokepoint):** use the already-passed `partnerRepo` in `resolveAppFromRequest` — after `FindByID`, resolve the caller's partner account for the context org and assert `app.PartnerAccountID == callerPartnerAccount.ID` (else **404**, to avoid confirming existence). Mirrors the Revenue API's `verifyAppAccess`; fixes all callers at once. Add a cross-org regression test.

## S2 — `RequireOrgRole` middleware exists but is unused → privileged org ops ungated  ✅ **[High]**
**Component:** `middleware/org_context.go:96` (defined) vs `router.go` (never applied); `handler/audit_handler.go:22`, `handler/org_handler.go` (UpdateOrg/DeleteOrg/RevokeInvitation/UpdateNotificationPrefs).
**Evidence:** grep of `router.go` shows `OrgContextMW` wired but **zero `RequireOrgRole`** usages. `ListAuditLog` has no role check despite `CanViewAuditLog` being ADMIN/OWNER-only; `UpdateNotificationPrefs` trusts a `memberID` URL param.
**Impact:** any org member (incl. VIEWER) may perform admin/owner-only actions or read another member's data via guessed UUIDs — privilege escalation.
**Fix:** apply `RequireOrgRole(OWNER)` / `RequireOrgRole(ADMIN,OWNER)` to privileged routes; gate `UpdateNotificationPrefs` to self-or-admin. Add per-route RBAC tests.

## S3 — Invitation acceptance doesn't verify the accepting user's email  ✅ **[High]**
**Component:** `application/service/org_service.go` (`AcceptInvitation`).
**Evidence:** email referenced only in `InviteMember`, never compared in `AcceptInvitation`; the token is **returned in the API response** (no backend email delivery).
**Impact:** anyone who obtains the invite token can join the org **as the invited role** (invite-hijack), regardless of their email.
**Fix:** on accept, require the authenticated user's (verified) email == `invitation.email`; deliver the token via server-side email rather than returning it.

## S4 — No Firebase token-revocation check; no server-side logout  ✅ **[Medium]**
**Component:** `infrastructure/external/firebase_auth.go:41`.
**Evidence:** uses `client.VerifyIDToken`, **not** `VerifyIDTokenAndCheckRevoked`; no revocation endpoint.
**Impact:** a signed-out / compromised session's token stays valid until natural expiry (~1h); no immediate session kill.
**Fix:** use `VerifyIDTokenAndCheckRevoked` on sensitive paths (or globally); optionally add a revoke endpoint.

## S5 — No email-verification gate  ✅ **[Medium]**
**Component:** `middleware/auth.go` (JIT provisioning) / frontend signup.
**Evidence:** users authenticate immediately post-signup; `email_verified` isn't checked.
**Impact:** spam/abuse signups; **compounds S3** (unverified emails can accept invites).
**Fix:** require `email_verified` before granting access (or before privileged actions).

## S6 — External Revenue API: no tests + in-memory, fail-open rate limiter  ✅ **[Medium]**
**Component:** `internal/revenue_api/...` (no `_test.go`); `revenue_api/.../middleware/rate_limiter.go:59-63`.
**Evidence:** the entire external subtree is untested; the limiter **allows the request on store error** (fail-open) and is in-memory (per-instance on scale).
**Impact:** auth/isolation/limit behavior on a public, credentialed API is unverified by CI; a store error disables rate limiting. (Isolation itself *is* implemented via `verifyAppAccess` — this is about verification + limiter robustness.)
**Fix:** add an isolation-first test suite (cross-org key → `not_found`/403); move the limiter to the shared Redis store; decide fail-closed.

## S7 — Audit logs unbounded + store IP (PII)  ✅ **[Low]**
**Component:** `org_audit_log`, `api_audit_log` (no TTL); IP captured.
**Fix:** retention/rollup policy; document IP capture for DPA/GDPR.

---

## Suggested remediation order
1. **S1** (single chokepoint, broadest data-exposure surface) → 2. **S2** → 3. **S3 + S5** (invite/identity, related) → 4. **S6** (before any horizontal scaling) → 5. **S4** → 6. **S7**.

## Already solid (do not regress)
Org membership isolation (`OrgContextMW` `FindByOrgAndUser`) · Revenue API `verifyAppAccess` · AES-256-GCM partner tokens · HMAC-SHA256 (constant-time) Razorpay webhooks · fail-closed route registration · 256-bit `crypto/rand` invite tokens.

## References
Per-finding detail: `docs/designs/14-org-and-team.md` (S1/S2/S3/S7), `docs/designs/15-auth.md` (S4/S5), `docs/designs/08-revenue-api.md` (S6); S1 cross-refs in `02/03/04/10/13`.

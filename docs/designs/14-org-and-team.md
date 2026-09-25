# Tech Design: Org & Team (multi-tenancy) (as-built)

**PRD:** docs/prds/14-org-and-team.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** This is the **authorization boundary** for the whole product — the org/member/role model and the org-context middleware gate every tenant's data. Getting authz wrong is the highest-impact failure in the system. Data model (`organizations`, `members`, `invitations`, `org_audit_log`) is load-bearing.

> **As-built.** "Current state IS the design." Every claim cites a real path:line and, for the security-critical ones, was verified against code. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — auth & provisioning.** `AuthMiddleware` verifies the Firebase token; on first login it creates the User and **auto-provisions a personal org** with the creator as OWNER (`auth.go:69-99`; org-less fallback recoverable via migration 000042).

**FACT — org-context / isolation (verified).** `OrgContextMiddleware` (`org_context.go:36-91`) resolves the org from `:orgId`/`X-Org-Id`, then `FindByOrgAndUser(org, user)` → **403 "not a member"** if the caller isn't a member (`:77-80`), and **403 account_suspended** if inactive (`:83-85`). It injects org+member into context. **Membership isolation is enforced at the middleware, not header-trust.**

**FACT — roles.** `OrgRole` = OWNER / ADMIN / VIEWER with a permission matrix (`org_role.go`). A `RequireOrgRole(minRoles...)` middleware exists (`org_context.go:96-121`, OWNER always passes).

**FACT (verified) — RBAC middleware is unused.** `router.go` wires `OrgContextMW` on org routes but **never applies `RequireOrgRole`** (grep: zero references). So role enforcement falls to handlers/services, which is **inconsistent**: role-management ops enforce at the service layer (owner-protection tested), but `UpdateOrg`/`DeleteOrg`/`RevokeInvitation`/`ListAuditLog`/`UpdateNotificationPrefs` have no verified role/org-context gate (`org_handler.go` shows only two 403 mappings; `audit_handler.go:22` has none).

**FACT — invitations.** `InviteMember` enforces plan member-limits, generates a **256-bit crypto/rand** token, 7-day expiry, PENDING status, and audits (`org_service.go`). `AcceptInvitation` checks PENDING+not-expired+not-already-member — **but does not compare the accepting user's email to the invite email** (verified: email is referenced only in `InviteMember`, not accept). Token is **returned in the API response** (no backend email delivery).

**FACT — audit.** `org_audit_log` captures org/member/role/invitation/webhook actions (actor, action, target, metadata, IP); indexed by (org, created_at) and (actor, created_at); no TTL.

## Proposed design

**No redesign — documents the shipped design.** Pattern: **Firebase auth → membership-verified org context → (intended-but-absent) role gate → handler/service**, with an audited invitation lifecycle. Flow: see `docs/designs/14-org-and-team-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — tables** (migration 000036): `organizations` (name, slug, plan_tier), `org_members` (org_id, user_id, role, status, joined_at, suspended_at), `org_invitations` (org_id, email, role, token unique, status, expires_at, invited_by), `org_audit_log` (actor, action, target_type/id, metadata JSONB, ip, created_at). Migration 000042 backfills org-less users. No new migration.

### API & events

**FACT.** Orgs: `POST/GET /orgs`, `GET/PUT/DELETE /orgs/:orgId`. Members: `GET /orgs/:orgId/members`, `PUT .../:userId/role`, `DELETE .../:userId`, `PUT .../:userId/suspend|unsuspend`. Invitations: `POST /orgs/:orgId/invitations`, `DELETE .../invitations/:id`, `POST /invitations/:token/accept`. Audit: `GET /orgs/:orgId/audit-log`. No events.
**Breaking-change check:** role enum values + membership shapes bind `OrganizationProvider`.

## Alternatives considered

- **Membership-verified middleware vs header-trust.** As-built verifies membership (`org_context.go:77`) — the correct choice; header-trust would be a tenant-isolation hole. No alternative recorded (this is simply right).
- **Middleware role-gating (`RequireOrgRole`) vs handler/service checks.** The middleware was **built but not wired**; the codebase de-facto relies on scattered service checks. **Choose middleware gating** (the intended design) to make privileged ops uniformly safe — this is a gap, not a considered trade-off.
- **Token-in-response vs backend email delivery.** As-built returns the token (frontend delivers). **Choose backend delivery** to avoid token exposure in client logs/proxies and to enable email-bound acceptance.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Non-member requests org | `FindByOrgAndUser` fails | **FACT** 403 (verified `org_context.go:77-80`) | join/switch org |
| Suspended member | `!IsActive()` | **FACT** 403 account_suspended (`:83-85`) | unsuspend |
| Remove/suspend the OWNER | service guard | **FACT** `ErrCannotRemoveOwner` → 403 (tested `org_service_test.go`) | — |
| Expired/used invitation | status/expiry check | **FACT** rejected (`org_service.go`, tested) | re-invite |
| Provisioning fails on signup | error path | **FACT** user created org-less; backfill 000042 | migration |
| Invalid role value | validation | **FACT** 400 (`org_handler.go:204,336`) | fix input |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed) — several are SECURITY

- **D1 (security). `RequireOrgRole` is unused** — privileged ops (`UpdateOrg`/`DeleteOrg`/`RevokeInvitation`/`ListAuditLog`/`UpdateNotificationPrefs`) lack a verified role gate; any org member could invoke some of them (e.g., read the audit log despite `CanViewAuditLog` being ADMIN/OWNER-only). → apply `RequireOrgRole` to these routes.
- **D2 (security). Invitation acceptance doesn't verify email** — any logged-in user with the token can accept an invite meant for someone else (invite-hijack). → bind acceptance to the invited email.
- **D3 (security). No backend invite email delivery** — token returned in the response; risk of exposure in client logs/proxies. → server-side send.
- **D4 (security). Residual app-scoping gap** — `OrgContextMW` proves org membership, but app-scoped endpoints resolve `appID` via `appRepo.FindByID` **without** cross-checking the app belongs to the caller's org. A member of org A could pass org A's `X-Org-Id` + another org's `appID`. → cross-check `app.PartnerAccount/org` against context (cross-ref `02-ai-chat.md` D1; this is the refined systemic finding).
- **D5. `UpdateNotificationPrefs` trusts a `memberID` param** with no org-context/ownership check → potential cross-user pref read/write. → gate.
- **D6. Audit-log retention unbounded** — no TTL/rollup. → retention policy.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. **Cross-org denial (must-hold):** member of A requesting B → 403 for every org/member/invitation/audit route (metric #1). **TODO — no such test exists.**
2. **RBAC enforcement:** VIEWER/ADMIN rejected on owner-only ops; member rejected on audit-log read (metric #2). **TODO.**
3. **Invite integrity:** acceptance rejected when the logged-in email ≠ invite email (metric #3). **TODO** (needs D2 first).

**Adversarial (one per Failure-modes row — mostly EXISTING at service layer):**
- Owner-protection, expired/already-member invitation, role change, permission matrix — **exist** (`org_service_test.go`). Membership/suspension middleware behavior — **missing** (no `org_context_test.go`).

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 role-gating per privileged route · D2 email-bound acceptance · D4 cross-org appID · D5 cross-user prefs. **Highest priority; all security.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/application/service/ -run Org -v`
2. As a **VIEWER** of an org, call `PUT /orgs/{id}/members/{other}/role` → expect 403 (today: **verify** — may succeed, D1).
3. As a member of org A, call `GET /orgs/{B}/audit-log` → expect 403 (membership check should block; confirms D-none for this route).
4. Log in as user C (email ≠ invite), `POST /invitations/{token}/accept` → today likely **succeeds** (documents D2).

**External-platform reality:** Firebase auth is the external dependency (faked below `AuthMW` in tests). Authz logic is pure-local and MUST be unit-tested — the missing middleware + cross-org tests are the priority gap for a multi-tenant system.

## Pricing & policy touchpoints

**FACT:** member count is plan-limited at invite. **UNKNOWN** whether roles/audit are tier-gated. GDPR: audit log stores actor IP (note for retention/DPA review). → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Additive; migration 000042 already handles org-less users. **Security fixes (D1/D2/D4)** would be behavior-changing (some currently-allowed calls become 403) — stage with care + comms. Rollback = revert.

## Observability

**FACT:** `org_audit_log` is a rich actor/action/IP trail, queryable by org/actor/time. **DIVERGENCE:** no metric/alert on 403 rates, failed-authz spikes, or provisioning failures; no retention. On-call reads `org_audit_log` + Hetzner logs. → proposed observability + retention TODO.

## Open questions (security-weighted)

- **Wire `RequireOrgRole` on privileged routes (D1).** Owner: Eng/Sec. **(highest priority)**
- **Email-bind invitation acceptance (D2).** Owner: Eng/Sec.
- **Cross-check appID against caller's org (D4).** Owner: Eng/Sec.
- **Gate `UpdateNotificationPrefs` (D5).** Owner: Eng/Sec.
- **Backend invite email delivery (D3).** Owner: Eng/Product.
- **Add cross-org-denial + middleware + handler-RBAC tests.** Owner: Eng.
- **Audit retention + IP/DPA review (D6).** Owner: Eng/Product.

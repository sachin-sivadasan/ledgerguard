# PRD: Org & Team (multi-tenancy) (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** / **INFERRED** / **UNKNOWN**.
> Success metrics are **PROPOSED** (future), not historical. Sources:
> `handler/{org_handler,audit_handler}.go`, `application/service/{org_service,org_audit_service}.go`,
> `middleware/org_context.go`, `middleware/auth.go`, `valueobject/org_role.go`,
> `frontend-flutter/lib/screens/settings/{team,audit_log}`, `providers/organization_provider.dart`.
> ⚠️ Security-sensitive; the RBAC/isolation findings below are verified against code in the design doc.

## Problem & evidence

LedgerGuard is multi-tenant: a partner's data belongs to an **organization**, and teammates collaborate with different privilege levels. The product needs orgs, membership, roles, invitations, and an audit trail — with hard isolation between tenants.

- **FACT:** a full org/member/invitation/role/audit system exists, with membership-verified org-context middleware and a comprehensive `org_service` test suite.
- **UNKNOWN — original demand evidence not recorded.** Open question — owner: Product.

## Target users

**INFERRED:** the workspace **OWNER/ADMIN** managing a team, and **VIEWER** teammates with read-only access. One user may belong to multiple orgs.
- **Not the target:** merchants; external collaborators without an account.

## Proposed solution (as-built behavior)

**FACT — orgs & membership.** On first login, a **personal org is auto-provisioned** (name from the email local-part), creator added as **OWNER** (`auth.go:93-99`, `org_service.go`). Users can create/list/switch orgs; the frontend has an org switcher + role-gated UI (`organization_provider.dart`).

**FACT — org-context & isolation.** `OrgContextMiddleware` resolves the org from `:orgId`/`X-Org-Id`, **verifies the caller is a member** (`FindByOrgAndUser` → 403 if not), rejects suspended members, and injects org+member into context (`org_context.go:36-91`). **This is real tenant isolation at the org layer** — not header trust.

**FACT — roles (3):** OWNER / ADMIN / VIEWER (`org_role.go`), with a permission matrix (OWNER: everything; ADMIN: manage members/sync/audit, but not admins/org-settings; VIEWER: read-only). A `RequireOrgRole` middleware exists.

**FACT — invitations.** Invite by email+role, **256-bit crypto/rand token**, 7-day expiry (`org_service.go`); accept while logged in via `/invitations/:token/accept`; owner-protection on remove/suspend/role-change; plan member-limits enforced. Actions audited.

**FACT — audit log.** Org-level `org_audit_log` captures org.created/updated/deleted, member.invited/joined/removed/suspended, role.changed, invitation.revoked, webhook.configured (actor, action, target, metadata, IP).

### Key screens

**No new wireframe** (as-built). Real surfaces: `settings/team_screen.dart` (members, invite, role change), audit-log screen, org switcher.

## Platform & policy constraints

**FACT.** Auth is Firebase (`auth.go`); org membership is the authorization boundary. Plan **member-limits** enforced at invite. Invitation **email delivery is not implemented** (the token is returned in the API response for the frontend to deliver).

## Pricing-tier impact

**FACT:** member count is plan-limited (enforced at invite). **UNKNOWN** whether roles/audit are tier-gated. Open question — owner: Product.

## Migration for existing users

**FACT/INFERRED:** a backfill migration (000042) recovers users created without an org (provisioning-failure path, `auth.go:98`). Otherwise additive.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Isolation (must-hold):** 100% of cross-org access attempts (member of A requesting B) return 403 — verified by an automated suite (which is missing today).
2. **RBAC correctness:** 100% of privileged actions (org update/delete, role change, audit read, invitation revoke) are rejected for insufficient roles — also unverified today.
3. **Invite integrity:** an invitation can only be accepted by the invited email — currently not enforced.

## Non-goals (deliberately absent in code)

1. **No SSO/SAML** — Firebase email/password auth only. **INFERRED.**
2. **No custom roles/granular permissions** — three fixed roles. **FACT.**
3. **No backend invitation email delivery** — token returned to the client. **FACT.**
4. **No audit-log retention/rollup** — unbounded. **FACT.**

## Open questions (several are security items — see the design doc's DIVERGENCE list)

- **RBAC middleware is largely unused** — privileged handlers (UpdateOrg, DeleteOrg, ChangeRole, RevokeInvitation, ListAuditLog, UpdateNotificationPrefs) appear to lack role/org-context gating, risking self-escalation or cross-org mutation via guessed UUIDs. **Verify + gate.** Owner: Eng/Sec.
- **Invitation acceptance doesn't verify the accepting user's email** matches the invite → invite-hijack. Owner: Eng/Sec.
- **Residual app-scoping gap:** org membership is verified, but app-scoped endpoints don't cross-check `appID` belongs to the caller's org (the systemic finding, refined). Owner: Eng/Sec.
- **No email delivery** for invitations — implement backend send? Owner: Eng/Product.
- **Missing tests:** no cross-org-denial or handler-RBAC tests. Owner: Eng.
- **Audit retention** unbounded. Owner: Eng.
- Role/audit tiering by plan? Owner: Product.

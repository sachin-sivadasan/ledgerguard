# PLAN-12: Security Fixes (CSRF, Tenant Isolation)

**Date:** 2026-02-27
**Status:** Completed

## Scope
- Wire Firebase Auth middleware to all protected routes
- OAuth state validation for CSRF protection (in-memory state store, 10-min TTL)
- Tenant isolation in SyncApp handler (verify app ownership before sync)
- Fix critical security blockers identified during review

## Key Decisions
- ADR-006: OAuth State Validation for CSRF Protection
- ADR-007: Tenant Isolation in Sync Handler

## Security Measures
- State parameter stored with user ID on StartOAuth
- State validated and consumed (one-time use) in Callback
- App ownership verified: `app.PartnerAccountID == user's partner account`
- 403 Forbidden returned on mismatch

# Architecture Decisions – LedgerGuard

## Format
```
### ADR-XXX: Title
**Date:** YYYY-MM-DD
**Status:** Accepted / Superseded / Deprecated

**Context:**
Why this decision was needed.

**Decision:**
What we decided.

**Consequences:**
Trade-offs and implications.
```

---

## Decisions

### ADR-001: Modular Monolith over Microservices
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need to choose architecture style for MVP. Team is small, rapid iteration needed.

**Decision:**
Build as a modular monolith in Go with clean architecture. Modules communicate via interfaces, not network calls.

**Consequences:**
- Faster development
- Simpler deployment
- Easy refactoring
- Can extract to microservices later if needed

---

### ADR-002: Full Ledger Rebuild over Incremental Updates
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need to decide how to sync transactions and compute metrics.

**Decision:**
Rebuild entire 12-month ledger on every sync instead of incremental updates.

**Consequences:**
- Deterministic: same input always produces same output
- Simpler to debug and audit
- Higher compute cost (acceptable at MVP scale)
- Can optimize later with hybrid approach

---

### ADR-003: Firebase Authentication
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need authentication system with Google OAuth support.

**Decision:**
Use Firebase Authentication. Frontend gets ID token, backend verifies via Admin SDK.

**Consequences:**
- Fast to implement
- Google OAuth included
- Stateless verification
- Vendor lock-in (acceptable trade-off)

---

### ADR-004: PostgreSQL as Primary Database
**Date:** 2025-02-26
**Status:** Accepted

**Context:**
Need a database for transactions, subscriptions, snapshots.

**Decision:**
Use PostgreSQL with pgcrypto for UUID generation.

**Consequences:**
- ACID compliance
- JSON support if needed
- Well-known, easy to hire for
- Requires managed instance in production

---

### ADR-005: Domain-Driven Design over Clean Architecture
**Date:** 2026-02-26
**Status:** Accepted

**Context:**
Initial implementation used Clean Architecture folder structure. Need clearer separation between business logic and infrastructure with explicit domain modeling.

**Decision:**
Refactor to Domain-Driven Design (DDD) structure:
- `domain/` - Entities, value objects, domain services, repository interfaces
- `application/` - Use cases, DTOs, orchestration
- `infrastructure/` - Database, external services, config
- `interfaces/` - HTTP handlers, middleware, routing

**Consequences:**
- Better domain isolation (domain layer has zero external dependencies)
- Clearer boundaries between layers
- Repository interfaces defined in domain (ports), implementations in infrastructure (adapters)
- More explicit modeling of business concepts
- Slightly more directories, but clearer responsibilities

---

### ADR-006: OAuth State Validation for CSRF Protection
**Date:** 2026-02-27
**Status:** Accepted

**Context:**
OAuth callback endpoint was missing state parameter validation, creating a CSRF vulnerability where an attacker could complete OAuth flow with their own credentials and link to victim's account.

**Decision:**
Implement in-memory state store with:
- State stored with user ID when StartOAuth called
- State validated and consumed (one-time use) in Callback
- 10-minute TTL for expiration
- State lookup returns associated user ID

**Consequences:**
- CSRF protection for OAuth flow
- No external dependencies (in-memory store)
- Needs Redis/distributed cache for multi-instance deployment
- Tests added for state validation

---

### ADR-007: Tenant Isolation in Sync Handler
**Date:** 2026-02-27
**Status:** Accepted

**Context:**
SyncApp endpoint allowed users to sync any app by ID without verifying ownership, creating a tenant isolation vulnerability.

**Decision:**
Add ownership verification before sync:
1. Get user's partner account from context
2. Lookup requested app by ID
3. Verify app.PartnerAccountID matches user's partner account
4. Return 403 Forbidden if mismatch

**Consequences:**
- Users can only sync their own apps
- Additional database lookup per request (acceptable)
- Tests added for forbidden case

---

### ADR-008: Default Revenue Share Tier Changed to 0%
**Date:** 2026-03-01
**Status:** Accepted

**Context:**
The default revenue share tier was set to 20% (DEFAULT_20), but the majority of Shopify app developers (especially indie developers) are on the reduced revenue share plan with 0% on their first $1M lifetime earnings.

**Decision:**
Change the default revenue share tier from DEFAULT_20 (20%) to SMALL_DEV_0 (0%):
- Backend: `entity.NewApp()` defaults to `RevenueShareTierSmallDev0`
- Backend: `ParseRevenueShareTier()` returns `SMALL_DEV_0` for invalid/empty input
- Frontend: `RevenueShareTier.fromCode()` defaults to `smallDev0`
- Users can change their tier in App Settings if they're on a different plan

**Consequences:**
- More accurate default for majority of indie developers
- Reduces initial confusion about fee calculations
- Users on 20% tier need to manually update their setting
- Existing apps in database retain their current tier (no data migration needed)

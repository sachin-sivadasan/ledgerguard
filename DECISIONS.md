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

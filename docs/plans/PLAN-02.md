# PLAN-02: DDD Architecture & Auth Middleware

**Date:** 2026-02-26
**Status:** Completed

## Scope
- Refactor from Clean Architecture to Domain-Driven Design (DDD) structure
- Implement Firebase Auth middleware (token verification, user context injection)
- Add role-based access middleware (OWNER/ADMIN)
- Config file support (YAML-based, environment-specific)

## Key Decisions
- ADR-003: Firebase Authentication
- ADR-005: Domain-Driven Design over Clean Architecture

## Architecture
```
internal/
├── domain/          → Entities, value objects, domain services, repository interfaces
├── application/     → Use cases, DTOs, orchestration
├── infrastructure/  → Database, external services, config
└── interfaces/      → HTTP handlers, middleware, routing
```

## Files Modified
- `backend/internal/interfaces/http/middleware/auth.go`
- `backend/internal/interfaces/http/middleware/role.go`
- `backend/internal/infrastructure/config/`
- All domain layer files reorganized

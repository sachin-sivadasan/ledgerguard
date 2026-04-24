# 01. Project Initialization & DDD Structure

## What It Does
Establishes LedgerGuard as a DDD modular monolith in Go 1.22+, organizing the entire backend into four strict architectural layers. This structure enforces the dependency rule where outer layers depend on inner layers, and the domain layer has zero external dependencies. The result is a codebase where business logic is isolated from infrastructure concerns, making the system testable, maintainable, and adaptable to future changes.

## Architecture
Domain-Driven Design modular monolith (ADR-001, ADR-005). Four concentric layers enforce unidirectional dependency flow:

- **Domain** (innermost) — Entities, value objects, domain services, repository interfaces. Zero external imports; depends only on Go stdlib and `github.com/google/uuid`.
- **Application** — Use-case orchestration services (SyncService, BillingService, MetricsAggregationService), DTOs, and schedulers. Depends on domain layer only.
- **Infrastructure** (adapters) — PostgreSQL repository implementations, external API clients (Shopify, Firebase, Razorpay, OpenAI, Slack), encryption, caching, and configuration loading. Implements domain repository interfaces (port/adapter pattern).
- **Interfaces** (outermost) — HTTP handlers, middleware (auth, RBAC, internal key), and router definitions. Depends on all inner layers.

The entry point (`cmd/server/main.go`) acts as the composition root, performing all dependency injection manually without a DI framework. This 539-line file wires every repository, service, handler, and middleware together, then starts the HTTP server and schedulers.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/cmd/server/main.go` | ~539 | Composition root: DB, Firebase, encryption, repos, services, handlers, schedulers, server startup |
| `backend/internal/domain/entity/` | 17 files | Domain entities (User, App, Transaction, Subscription, PartnerAccount, DailyMetricsSnapshot, etc.) |
| `backend/internal/domain/service/` | 6 source files | Domain services (RiskEngine, MetricsEngine, LedgerService, FeeVerificationService, EarningsCalculator, AuthService interface) |
| `backend/internal/domain/valueobject/` | 10 source files | Value objects (ChargeType, RiskState, Role, PlanTier, RevenueShareTier, SubscriptionStatus, BillingInterval, etc.) |
| `backend/internal/domain/repository/` | 15 files | Repository interfaces (ports) for all domain entities |
| `backend/internal/application/service/` | dir | Application services: SyncService, BillingService, MetricsAggregationService, RevenueMetricsService, NotificationService |
| `backend/internal/infrastructure/persistence/` | dir | PostgreSQL adapter implementations for all repository interfaces |
| `backend/internal/infrastructure/external/` | dir | External clients: Shopify Partner, Shopify OAuth, Firebase Auth/Messaging, Razorpay, OpenAI, Slack |
| `backend/internal/interfaces/http/handler/` | dir | HTTP handlers for all API endpoints |
| `backend/internal/interfaces/http/middleware/` | dir | Auth middleware (Firebase), role-based access, internal key middleware |
| `backend/internal/interfaces/http/router/router.go` | ~270 | Chi router with all route definitions, CORS, and middleware wiring |
| `backend/migrations/` | 32 up/down pairs | Database migrations managed by golang-migrate |

## Data Flow
```
                    ┌─────────────────────────────────────────────────┐
                    │             INTERFACES LAYER                     │
                    │  ┌─────────┐  ┌────────────┐  ┌─────────────┐  │
                    │  │ Handlers│  │ Middleware  │  │   Router    │  │
                    │  └────┬────┘  └─────┬──────┘  └──────┬──────┘  │
                    │       │             │                 │          │
                    └───────┼─────────────┼─────────────────┼──────────┘
                            │ depends on  │                 │
                    ┌───────▼─────────────▼─────────────────▼──────────┐
                    │           APPLICATION LAYER                       │
                    │  ┌────────────┐  ┌───────────┐  ┌────────────┐  │
                    │  │SyncService │  │BillingServ│  │MetricsAggr │  │
                    │  └─────┬──────┘  └─────┬─────┘  └─────┬──────┘  │
                    │        │               │               │         │
                    └────────┼───────────────┼───────────────┼─────────┘
                             │ depends on    │               │
                    ┌────────▼───────────────▼───────────────▼─────────┐
                    │              DOMAIN LAYER (zero deps)             │
                    │  ┌──────────┐ ┌───────────┐ ┌──────────────────┐│
                    │  │ Entities │ │  Value     │ │Domain Services   ││
                    │  │          │ │  Objects   │ │(RiskEngine, etc.)││
                    │  └──────────┘ └───────────┘ └──────────────────┘│
                    │  ┌────────────────────────────────────────────┐  │
                    │  │   Repository Interfaces (ports)            │  │
                    │  └────────────────────┬───────────────────────┘  │
                    └───────────────────────┼──────────────────────────┘
                                            │ implemented by
                    ┌───────────────────────▼──────────────────────────┐
                    │           INFRASTRUCTURE LAYER                    │
                    │  ┌─────────────┐  ┌──────────────┐  ┌────────┐  │
                    │  │ PostgreSQL   │  │ Shopify      │  │Firebase│  │
                    │  │ Repositories │  │ Partner API  │  │ Auth   │  │
                    │  └─────────────┘  └──────────────┘  └────────┘  │
                    └──────────────────────────────────────────────────┘
```

Startup sequence in `main.go`:
1. Load config (file or environment)
2. Connect PostgreSQL, run golang-migrate migrations
3. Initialize Firebase Auth + Messaging (optional, graceful degradation)
4. Initialize AES-256-GCM encryption
5. Create all repository instances (nil-safe if DB unavailable)
6. Create OAuth state store (in-memory, 10-min TTL)
7. Create Shopify Partner client with rate limiting
8. Wire all handlers with their dependencies
9. Create sync and notification schedulers, start background jobs
10. Build Chi router with CORS, middleware, and route definitions
11. Start HTTP server with graceful shutdown on SIGINT/SIGTERM

## Configuration
No feature-specific configuration for the DDD structure itself. The composition root reads from a YAML config file or environment variables. See individual feature docs for feature-specific configuration.

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `CONFIG_PATH` | — | No | Path to YAML config file (also via `-config` CLI flag) |
| `DATABASE_DSN` | — | Yes | PostgreSQL connection string |
| `DATABASE_MIGRATIONS_PATH` | — | No | Path to migrations directory (enables auto-migration on startup) |
| `SERVER_PORT` | `8080` | No | HTTP server listen port |

## API Surface
This is a structural document; there is no API surface specific to the DDD structure. See individual feature docs for endpoint details. The full route map is defined in `backend/internal/interfaces/http/router/router.go`.

Public (no auth):
- `GET /health` — Health check

All other routes are organized under `/api/v1/` with Firebase auth middleware. See [02. Firebase Auth](02-firebase-auth-user-management.md) for auth details.

## Extension Points
- **New entity** — Add to `internal/domain/entity/`, create repository interface in `internal/domain/repository/`, implement adapter in `internal/infrastructure/persistence/`, wire in `main.go`.
- **New domain service** — Add to `internal/domain/service/`. Must have zero external dependencies; receives repository interfaces via constructor injection.
- **New value object** — Add to `internal/domain/valueobject/`. Pure Go types with validation methods.
- **New HTTP handler** — Add to `internal/interfaces/http/handler/`, register route in `router.go`, wire dependencies in `main.go`.
- **New external integration** — Add client to `internal/infrastructure/external/`, define interface in domain or application layer, inject via `main.go`.
- **New migration** — Add numbered `.up.sql` and `.down.sql` files to `backend/migrations/`. Auto-applied on server startup if migrations path is configured.

## Gotchas
- **Domain must never import infrastructure.** The Go compiler does not enforce layer boundaries by default. Developers must ensure `internal/domain/` packages never import from `internal/infrastructure/` or `internal/interfaces/`. A misplaced import silently violates the architecture.
- **Nil-safe handler wiring.** The composition root checks for nil dependencies before creating handlers. If a dependency is not configured (e.g., no Firebase credentials), the handler is nil and the route is not registered. This allows the server to start in degraded mode.
- **Migrations are auto-applied.** On every server startup, `golang-migrate` runs pending migrations. There are currently 32 migration pairs. A dirty migration state will prevent startup; see `docs/GCP_SETUP_LOG.md` section 12 for recovery steps.
- **Manual DI.** There is no DI framework. All wiring happens in `main.go`. Adding a new service requires updating the composition root manually. This is intentional to keep the dependency graph explicit and debuggable.
- **Graceful degradation.** The server can start without a database connection (health endpoint only), without Firebase (no auth), or without Shopify credentials (no sync). This is useful for local development and testing.

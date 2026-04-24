# 02. Firebase Auth & User Management

## What It Does
Handles all authentication and user management for LedgerGuard using Firebase Authentication as the identity provider. The auth middleware extracts Bearer tokens from incoming requests, verifies them against Firebase Admin SDK, and automatically provisions user records on first access. This gives the Flutter app a seamless login experience while the backend maintains its own user entity with roles, plan tiers, and onboarding state. A separate internal key middleware supports service-to-service calls without Firebase tokens.

## Architecture
Spans three DDD layers following the port/adapter pattern (ADR-003):

- **Domain layer** — `AuthTokenVerifier` interface (port) in `domain/service/`, `User` entity in `domain/entity/`, `UserRepository` interface in `domain/repository/`. These define what authentication means to the business logic without knowing about Firebase.
- **Infrastructure layer** — `FirebaseAuthService` (adapter) implements `AuthTokenVerifier` using Firebase Admin SDK. The `PostgresUserRepository` persists user records.
- **Interfaces layer** — `AuthMiddleware` orchestrates the flow: extract token, verify via `AuthTokenVerifier`, look up or create user, inject into request context. `InternalKeyMiddleware` provides an alternative auth path for server-to-server calls. `RequireRoles` middleware enforces role-based access control.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/interfaces/http/middleware/auth.go` | ~135 | AuthMiddleware (Firebase token verify + auto-provision), InternalKeyMiddleware (X-Internal-Key header), UserFromContext() helper |
| `backend/internal/interfaces/http/middleware/role.go` | ~54 | RequireRoles middleware, OWNER has superset access to all roles |
| `backend/internal/infrastructure/external/firebase_auth.go` | ~53 | FirebaseAuthService: adapts Firebase Admin SDK to AuthTokenVerifier interface |
| `backend/internal/domain/entity/user.go` | ~41 | User entity: UUID, firebase_uid, email, role, plan_tier, onboarding state |
| `backend/internal/domain/service/auth_service.go` | ~13 | AuthTokenVerifier interface (port) and TokenClaims struct |
| `backend/internal/domain/valueobject/role.go` | — | Role value object: OWNER, ADMIN |
| `backend/internal/domain/valueobject/plan_tier.go` | — | PlanTier value object: FREE, STARTER, PRO |
| `backend/internal/interfaces/http/handler/me.go` | ~34 | GET /api/v1/me handler returning current user profile |
| `backend/internal/interfaces/http/router/router.go` | ~270 | Route definitions with auth and role middleware wiring |

## Data Flow
```
┌──────────────┐     Bearer token      ┌─────────────────────┐
│  Flutter App │ ──────────────────────▶│   AuthMiddleware     │
│  (Firebase   │                        │  .Authenticate()     │
│   Auth SDK)  │                        └──────────┬──────────┘
└──────────────┘                                   │
                                          1. extractBearerToken(r)
                                                   │
                                                   ▼
                                        ┌─────────────────────┐
                                        │ AuthTokenVerifier    │
                                        │ .VerifyIDToken()     │
                                        │ (Firebase Admin SDK) │
                                        └──────────┬──────────┘
                                                   │
                                          2. Returns TokenClaims
                                             {UID, Email}
                                                   │
                                                   ▼
                                        ┌─────────────────────┐
                                        │ UserRepository       │
                                        │ .FindByFirebaseUID() │
                                        └──────────┬──────────┘
                                                   │
                                      ┌────────────┴────────────┐
                                      │                         │
                                 3a. Found                 3b. Not Found
                                      │                         │
                                      │              ┌──────────▼──────────┐
                                      │              │ entity.NewUser()     │
                                      │              │ → Role: OWNER        │
                                      │              │ → PlanTier: FREE     │
                                      │              │ userRepo.Create()    │
                                      │              └──────────┬──────────┘
                                      │                         │
                                      └────────────┬────────────┘
                                                   │
                                          4. context.WithValue(user)
                                                   │
                                                   ▼
                                        ┌─────────────────────┐
                                        │    HTTP Handler      │
                                        │ middleware.UserFrom  │
                                        │   Context(ctx)       │
                                        └─────────────────────┘
```

### Internal Key Auth Flow (service-to-service)
```
┌──────────────┐   X-Internal-Key     ┌───────────────────────┐
│  Internal    │ ─────────────────────▶│ InternalKeyMiddleware │
│  Service     │   header              │ .Authenticate()       │
└──────────────┘                       └──────────┬────────────┘
                                                  │
                                         1. Compare header value
                                            with configured key
                                                  │
                                       ┌──────────┴──────────┐
                                       │                      │
                                  Match                  No Match
                                       │                      │
                                       ▼                      ▼
                                  next.ServeHTTP()     401 Unauthorized
```

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `FIREBASE_CREDENTIALS_FILE` | — | Yes (for auth) | Path to Firebase service account JSON. If empty, Firebase initializes with Application Default Credentials. |
| `SERVER_INTERNAL_KEY` | — | No | Shared secret for service-to-service calls via X-Internal-Key header. If empty, internal routes are not registered. |

Firebase Auth initialization is optional. If credentials are not provided, the server starts without auth middleware and all authenticated routes are unavailable. The health endpoint remains accessible.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/me` | Firebase | Returns current user profile (id, email, role, plan_tier, created_at) |
| `GET` | `/health` | None | Health check, always available |
| `POST` | `/api/v1/internal/*` | Internal Key | Internal routes authenticated via X-Internal-Key header |

### Role-Based Access Control
The `RequireRoles` middleware restricts specific endpoints by role:

| Role | Access Level |
|------|-------------|
| `OWNER` | Superset: access to all routes including admin-only endpoints |
| `ADMIN` | Access to admin endpoints (manual token management) |

Currently, manual token routes (`POST/GET/DELETE /api/v1/integrations/shopify/token`) require ADMIN or OWNER role.

### User Entity Fields
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Internal primary key |
| `firebase_uid` | string | Firebase Authentication UID (unique) |
| `email` | string | Email from Firebase token claims |
| `role` | Role | OWNER (default for new users) or ADMIN |
| `plan_tier` | PlanTier | FREE (default), STARTER, or PRO |
| `onboarding_completed_at` | *time.Time | Nil until onboarding is complete |
| `created_at` | time.Time | Account creation timestamp |

## Extension Points
- **New auth provider** — Implement the `AuthTokenVerifier` interface in `internal/domain/service/auth_service.go`. The interface requires a single method: `VerifyIDToken(ctx, idToken) (*TokenClaims, error)`. Register the new provider in `main.go` instead of FirebaseAuthService.
- **New roles** — Add to `internal/domain/valueobject/role.go`. Update `RequireRoles` usage in `router.go` if new routes need the role.
- **New plan tiers** — Add to `internal/domain/valueobject/plan_tier.go`. Tier-gated features can check `user.PlanTier` in handlers.
- **WebSocket auth** — For the chat WebSocket, the token is passed as a query parameter (`?token=`). The chat handler extracts and verifies it outside the standard middleware chain.

## Gotchas
- **Auto-provisioning creates users on first API call.** Any valid Firebase token that reaches the backend creates a user record if one does not exist. There is no separate registration flow. New users default to `OWNER` role and `FREE` plan tier.
- **Token in WebSocket is a query param.** The standard `Authorization: Bearer` header pattern does not work for WebSocket upgrades in all browsers. The chat endpoint accepts the token as `?token=` in the connection URL.
- **OWNER is a superset of ADMIN.** The `RequireRoles` middleware always grants access to OWNER regardless of which roles are specified. This is intentional but means you cannot restrict an endpoint to ADMIN-only without OWNER access.
- **No token refresh on the backend.** Token refresh is handled entirely by the Flutter app's Firebase SDK. The backend only validates tokens; it never issues or refreshes them. Expired tokens result in 401 responses.
- **Graceful degradation without Firebase.** If Firebase credentials are missing, `firebaseAuth` is nil, `authMW` is nil, and all authenticated routes are simply not registered. The server starts but only the health endpoint works.

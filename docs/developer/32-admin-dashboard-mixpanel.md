# 32 — Admin Dashboard API + Mixpanel Event Tracking

> Cross-tenant admin visibility and analytics event pipeline for the LedgerGuard platform.

---

## Overview

Two features delivered together:

1. **Admin Dashboard API** — 4 read-only endpoints providing cross-tenant visibility into users, onboarding, sync jobs, and billing. Protected by `AuthMW + AdminMW`.
2. **Mixpanel Server-Side Tracking** — Fire-and-forget event tracking at key lifecycle points (signup, app selection, sync, billing), powering funnel analysis in Mixpanel.

---

## Admin Dashboard API

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/users` | All users with onboarding status, plan tier, app count |
| GET | `/api/v1/admin/onboarding` | Onboarding funnel breakdown |
| GET | `/api/v1/admin/sync?limit=50` | Recent sync jobs across all apps |
| GET | `/api/v1/admin/billing` | All billing subscriptions with user info |

All endpoints require `Authorization: Bearer <token>` and the user must have `ADMIN` or `OWNER` role.

### Architecture

```
AdminHandler → AdminRepository (interface) → PostgresAdminRepository
```

No new tables or migrations. Queries join existing tables:
- `users`, `partner_accounts`, `apps` → ListUsers, OnboardingFunnel
- `sync_jobs`, `apps`, `users` → ListSyncJobs
- `billing_subscriptions`, `users` → ListBillingSubscriptions

### Response Shapes

**GET /api/v1/admin/users**
```json
{
  "users": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "role": "OWNER",
      "plan_tier": "FREE",
      "created_at": "2026-01-15T...",
      "onboarding_completed_at": null,
      "app_count": 2,
      "partner_connected": true
    }
  ],
  "total": 1
}
```

**GET /api/v1/admin/onboarding**
```json
{
  "total_users": 100,
  "partner_connected": 75,
  "app_selected": 50,
  "onboarding_complete": 30
}
```

**GET /api/v1/admin/sync?limit=10**
```json
{
  "jobs": [
    {
      "id": "uuid",
      "app_id": "uuid",
      "app_name": "MyApp",
      "user_email": "user@example.com",
      "job_type": "full_sync",
      "status": "completed",
      "total_items": 150,
      "completed_items": 150,
      "created_at": "2026-05-01T..."
    }
  ],
  "total": 1
}
```

**GET /api/v1/admin/billing**
```json
{
  "subscriptions": [
    {
      "id": "uuid",
      "user_email": "user@example.com",
      "plan": "STARTER",
      "status": "ACTIVE",
      "amount_cents": 999,
      "currency": "INR",
      "created_at": "2026-04-01T..."
    }
  ],
  "total": 1
}
```

---

## Mixpanel Event Tracking

### Architecture

```
Handler/Service → EventTracker interface → MixpanelClient → Mixpanel HTTP API
                                         → NoopTracker (when disabled/dev)
```

**Domain layer** defines `EventTracker` interface (no Mixpanel dependency). Infrastructure layer provides two implementations:
- `MixpanelClient` — sends events via Mixpanel HTTP API (`/track`, `/engage`), fire-and-forget (async goroutine)
- `NoopTracker` — silent no-op when `MIXPANEL_TOKEN` is empty

### Events Tracked

| Event Name | Trigger Point | Properties |
|-----------|---------------|------------|
| `user_signup` | AuthMiddleware (user auto-created) | email, role |
| `app_selected` | AppHandler.SelectApp | app_name, app_id |
| `sync_started` | QueueSyncService.EnqueueSync | job_type, app_id, job_id |
| `dashboard_viewed` | MetricsHandler.GetLatestMetrics | app_id |
| `billing_subscription_created` | BillingService.CreateCheckout | plan, amount |
| `billing_activated` | BillingService.handleActivated | plan |
| `billing_cancelled` | BillingService.handleCancelled | plan |
| `billing_payment_failed` | BillingService.handleHalted | plan |

### Configuration

```yaml
# config.local.yaml
mixpanel:
  token: "your-mixpanel-project-token"
```

Or via environment variable:
```bash
MIXPANEL_TOKEN=your-token
```

When the token is empty, `NoopTracker` is used automatically (no errors, no external calls).

### Wiring

Each service that emits events receives the tracker via setter injection (`SetTracker`):
- `AuthMiddleware.SetTracker(tracker)`
- `AppHandler.SetTracker(tracker)`
- `BillingService.SetTracker(tracker)`
- `QueueSyncService.SetTracker(tracker)`
- `MetricsHandler.SetTracker(tracker)`

---

## Flutter Client-Side Tracking

### Architecture

```
Provider/Screen → MixpanelService → mixpanel_flutter SDK → Mixpanel
                                   → No-op (when token empty)
```

`MixpanelService` wraps the `mixpanel_flutter` SDK. Token is passed at build time via `--dart-define=MIXPANEL_TOKEN=...`. When the token is empty (dev), all calls are silent no-ops.

### Events Tracked (Client)

| Event | Trigger | Properties |
|-------|---------|------------|
| `login` | AuthProvider.signIn | method |
| `signup` | AuthProvider.signUp | method |
| `logout` | AuthProvider.signOut | — |
| `dashboard_viewed` | DashboardScreen.initState | app_id |
| `app_selected` | (future) AppsProvider | app_id, app_name |
| `page_view` | (future) GoRouter observer | page |

### Key Files (Flutter)

| File | Role |
|------|------|
| `frontend-flutter/lib/services/mixpanel_service.dart` | Mixpanel SDK wrapper + convenience methods |
| `frontend-flutter/lib/main.dart` | Service init + Provider registration |
| `frontend-flutter/lib/app.dart` | Wires MixpanelService into AuthProvider |
| `frontend-flutter/lib/providers/auth_provider.dart` | Tracks login, signup, logout |
| `frontend-flutter/lib/screens/dashboard/dashboard_screen.dart` | Tracks dashboard_viewed |

### Build with Token

```bash
# Dev (no token — tracking disabled)
cd frontend-flutter && flutter run

# Staging / Prod
cd frontend-flutter && flutter build web --release \
  --dart-define=MIXPANEL_TOKEN=your-mixpanel-token
```

---

## Key Files (Backend)

| File | Role |
|------|------|
| `internal/domain/repository/admin_repository.go` | AdminRepository interface + DTOs |
| `internal/infrastructure/persistence/admin_repository.go` | PostgreSQL implementation |
| `internal/interfaces/http/handler/admin.go` | HTTP handlers (4 endpoints) |
| `internal/domain/service/event_tracker.go` | EventTracker interface |
| `internal/infrastructure/external/mixpanel_client.go` | Mixpanel HTTP API client |
| `internal/infrastructure/external/noop_tracker.go` | No-op tracker for dev/test |
| `internal/infrastructure/config/config.go` | MixpanelConfig struct |
| `internal/interfaces/http/router/router.go` | Admin route group |
| `cmd/server/main.go` | Dependency wiring |

---

## Testing

```bash
# Run admin handler tests
cd backend && go test ./internal/interfaces/http/handler/ -run TestAdmin -v

# Run event tracker tests
cd backend && go test ./internal/infrastructure/external/ -run TestNoop -v
cd backend && go test ./internal/infrastructure/external/ -run TestMixpanel -v

# Full suite
cd backend && go test ./...
```

---

## Mixpanel Dashboard (Manual Setup)

After deploying with a valid `MIXPANEL_TOKEN`, create these reports in Mixpanel:

1. **Onboarding Funnel:** user_signup → app_selected → sync_started
2. **Billing Funnel:** billing_subscription_created → billing_activated → billing_cancelled
3. **Payment Failures:** billing_payment_failed count over time
4. **Sync Volume:** sync_started events over time by job_type

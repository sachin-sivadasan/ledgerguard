# 14. Webhook Integration

## What It Does
Processes incoming Shopify Partner webhooks to keep subscription data in near-real-time sync. Four webhook topics are handled:

- **`app/installed`** — a merchant installs the app. Logs the event; if a previous subscription exists (reinstall), reactivates it from UNINSTALLED/CANCELLED to PENDING and restores from soft-delete.
- **`app_subscriptions/update`** — subscription status changes (ACTIVE, CANCELLED, FROZEN, EXPIRED). Updates the subscription record and recalculates risk state.
- **`app/uninstalled`** — a merchant uninstalls the app. Soft-deletes the subscription and marks it as CHURNED.
- **`subscription_billing_attempts/failure`** — a billing attempt failed. Escalates the subscription's risk state by one level (SAFE to ONE_CYCLE_MISSED, ONE_CYCLE_MISSED to TWO_CYCLES_MISSED, etc.).

Each event is recorded as a `SubscriptionEvent` for lifecycle auditing, and critical risk state changes trigger notifications to the app developer via the NotificationService.

## Architecture
Two layers:

- **HTTP handler** (`internal/interfaces/http/handler/`) — `WebhookHandler` reads the raw request body, extracts Shopify headers (`X-Shopify-Topic`, `X-Shopify-Shop-Domain`, `X-Shopify-Hmac-Sha256`), and delegates to the service layer. Always returns HTTP 200 to prevent Shopify from retrying, even on processing errors.
- **Application service** (`internal/application/service/`) — `WebhookService` contains all business logic: payload parsing, subscription lookup, status/risk updates, lifecycle event recording, and notification dispatch.

The `WebhookService` uses the builder pattern for optional dependencies:
- `WithSubscriptionEventRepo()` — enables lifecycle event recording.
- `WithAppEventRepo()` — enables install/uninstall event recording in `app_events` table.
- `WithNotificationService()` — enables risk change alerts (requires both a `PartnerAccountRepository` and a `NotificationService`).

HMAC validation is implemented in the service (`ValidateHMAC`) using `crypto/hmac` with SHA-256 and base64 encoding, but is not enforced in the current handler. The handler logs a warning when the HMAC header is missing.

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/interfaces/http/handler/webhook.go` | WebhookHandler: HandleWebhook (generic router), HandleAppInstalled, HandleSubscriptionUpdate, HandleAppUninstalled, HandleBillingFailure, GetStats |
| `backend/internal/application/service/webhook_service.go` | WebhookService: ProcessEvent (router), ProcessAppInstalled, ProcessSubscriptionUpdate, ProcessAppUninstalled, ProcessBillingFailure, ValidateHMAC |
| `backend/internal/domain/entity/subscription_event.go` | SubscriptionEvent entity: lifecycle tracking with from/to status and risk state, event type classification (IsChurnEvent, IsVoluntaryChurn, IsInvoluntaryChurn) |
| `backend/internal/domain/repository/subscription_event_repository.go` | SubscriptionEventRepository interface: Create, FindBySubscriptionID, FindByAppID, FindChurnEvents, CountByEventType |

## Data Flow

### App Installed
```
POST /webhooks/shopify/installed
│
├── WebhookService.ProcessAppInstalled(event)
│     ├── Parse payload (same shape as AppUninstalledPayload: ID, Domain, MyshopifyDomain)
│     ├── appRepo.FindAllByPartnerAppID(event.AppID)
│     │     └── No matching app → log and return nil
│     └── For each app:
│           ├── Record RELATIONSHIP_INSTALLED in app_events (always — audit trail)
│           ├── subRepo.FindByAppIDAndDomain(appID, myshopifyDomain)
│           │     └── Not found → log "new install, subscription created on first sync"
│           ├── If previous status was UNINSTALLED or CANCELLED (reinstall):
│           │     ├── Set status = "PENDING", riskState = SAFE
│           │     ├── sub.Restore() — clear soft-delete marker
│           │     └── subRepo.Upsert(sub)
│           └── Record SubscriptionEvent (type="app_installed", reason="Shop installed the app")
│
└── Return HTTP 200
```

### Subscription Update
```
POST /webhooks/shopify/subscriptions
│
├── Read body, extract headers (Topic, ShopDomain, HMAC, WebhookId)
│
├── WebhookService.ProcessSubscriptionUpdate(event)
│     ├── Parse SubscriptionUpdatePayload (ID, Name, Status, BillingOn, etc.)
│     ├── subRepo.FindByShopifyGID(payload.ID)
│     │     └── Not found → log and return nil (new sub, not yet synced)
│     ├── Save old status and risk state
│     ├── Update status from payload
│     ├── Map status to risk state:
│     │     ACTIVE     → SAFE
│     │     CANCELLED  → CHURNED
│     │     EXPIRED    → CHURNED
│     │     FROZEN     → TWO_CYCLES_MISSED
│     ├── subRepo.Upsert(sub)
│     ├── If status changed AND subEventRepo configured:
│     │     └── Create SubscriptionEvent (oldStatus→newStatus, oldRisk→newRisk, type="webhook")
│     └── If risk state changed:
│           └── Resolve app → partner account → user ID
│                 └── notificationSvc.SendCriticalAlert(userID, appName, domain, oldRisk, newRisk)
│
└── Return HTTP 200
```

### App Uninstalled
```
POST /webhooks/shopify/uninstalled
│
├── WebhookService.ProcessAppUninstalled(event)
│     ├── Parse AppUninstalledPayload (ID, Name, Domain, MyshopifyDomain)
│     ├── appRepo.FindAllByPartnerAppID(event.AppID)
│     │     └── Iterate all apps matching the Shopify app GID
│     └── For each app:
│           ├── subRepo.FindByAppIDAndDomain(appID, myshopifyDomain)
│           ├── Set status = "UNINSTALLED", riskState = CHURNED
│           ├── sub.SoftDelete()
│           ├── subRepo.Upsert(sub)
│           ├── Record SubscriptionEvent (type="app_uninstalled", reason="Shop uninstalled the app")
│           └── If risk was not already CHURNED → SendCriticalAlert
│
└── Return HTTP 200
```

### Billing Failure
```
POST /webhooks/shopify/billing-failure
│
├── WebhookService.ProcessBillingFailure(event)
│     ├── Parse payload (SubscriptionID, ErrorCode, ErrorMessage)
│     ├── subRepo.FindByShopifyGID(subscriptionID)
│     ├── Escalate risk state by one level:
│     │     SAFE                → ONE_CYCLE_MISSED
│     │     ONE_CYCLE_MISSED    → TWO_CYCLES_MISSED
│     │     TWO_CYCLES_MISSED   → CHURNED
│     │     CHURNED             → (no change)
│     ├── subRepo.Upsert(sub)
│     ├── Record SubscriptionEvent (type="billing_failure", reason includes error code and message)
│     └── If risk state changed → SendCriticalAlert
│
└── Return HTTP 200
```

### Generic Router
```
POST /webhooks/shopify
│
└── WebhookService.ProcessEvent(event)
      ├── topic="app/installed"                         → ProcessAppInstalled
      ├── topic="app_subscriptions/update"              → ProcessSubscriptionUpdate
      ├── topic="app/uninstalled"                       → ProcessAppUninstalled
      ├── topic="subscription_billing_attempts/failure"  → ProcessBillingFailure
      └── default                                        → log "Unhandled webhook topic"
```

## Configuration
| Setting | Notes |
|---------|-------|
| Webhook secrets | Registered per app via `RegisterWebhookSecret(appID, secret)`. Stored in an in-memory map. |
| HMAC algorithm | SHA-256 with base64 encoding. Matches Shopify's `X-Shopify-Hmac-Sha256` header format. |
| HMAC enforcement | **Not enforced in current handler.** Validation method exists but the handler only logs a warning for missing signatures. |

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/webhooks/shopify` | HMAC (not enforced) | Generic webhook router, dispatches by `X-Shopify-Topic` header |
| POST | `/webhooks/shopify/installed` | HMAC (not enforced) | App installation events (new install or reinstall) |
| POST | `/webhooks/shopify/subscriptions` | HMAC (not enforced) | Subscription status update events |
| POST | `/webhooks/shopify/uninstalled` | HMAC (not enforced) | App uninstallation events |
| POST | `/webhooks/shopify/billing-failure` | HMAC (not enforced) | Billing attempt failure events |
| GET | `/api/v1/webhooks/stats` | Firebase (admin) | Webhook processing stats (placeholder) |

Note: Webhook endpoints do NOT use Firebase authentication. They are public endpoints intended to be validated via HMAC signature from Shopify.

## Extension Points
- **Enforce HMAC validation.** The `ValidateHMAC()` method is implemented and ready. Wire it into the handler middleware to reject unsigned requests. This is a critical security hardening step for production.
- **New webhook topics.** Add a case to `ProcessEvent()` and create a corresponding `Process*` method. Follow the same pattern: parse payload, find subscription, update state, record event, send notification.
- **Webhook retry handling.** Shopify retries webhooks on non-2xx responses. The current handler always returns 200 to prevent retries. Adding idempotency keys (using Shopify's `X-Shopify-Webhook-Id` header) would allow safe retries.
- **Async processing.** Move webhook processing to a background queue (e.g., database-backed job queue, Redis, or pub/sub) for decoupled, retriable processing. The handler would enqueue and return 200 immediately.
- **GetStats endpoint.** Currently returns a placeholder. Wire it to `SubscriptionEventRepository.CountByEventType` for actual webhook processing statistics.

## Gotchas
- **HMAC is not enforced.** The `ValidateHMAC()` method exists but is never called by the handler. In development this is fine, but in production any actor can POST to the webhook endpoints. This is the most critical security gap in the webhook system.
- **Always returns HTTP 200.** Even if processing fails, the handler returns 200 and logs the error. This prevents Shopify from retrying, which means failed events are silently dropped. Check application logs to detect processing failures.
- **FROZEN maps to TWO_CYCLES_MISSED in webhooks but ONE_CYCLE_MISSED in the Risk Engine.** The `ProcessSubscriptionUpdate` method maps FROZEN status to `RiskStateTwoCyclesMissed`, while the domain RiskEngine maps FROZEN to `RiskStateOneCycleMissed`. This inconsistency means webhook-driven updates produce different risk states than sync-driven updates for the same subscription status.
- **WebhookHandler not wired in main.go.** The handler and service exist but `WebhookHandler` is not set in the router config. Routes are guarded by `if cfg.WebhookHandler != nil` so they're currently inactive. Wire in main.go when deploying to a publicly reachable server.
- **Unknown subscriptions are silently ignored.** If a webhook arrives for a subscription that has not been synced yet (`FindByShopifyGID` returns not found), the event is logged and discarded. There is no queuing mechanism to process it once the subscription appears.
- **App uninstall iterates all matching apps.** `FindAllByPartnerAppID` can return multiple apps if the same Shopify app GID appears under different partner accounts. The uninstall handler processes subscriptions across all of them.
- **Billing failure escalation is one-directional.** Each failure bumps the risk state up by exactly one level. There is no de-escalation on successful retry; that only happens through `ProcessSubscriptionUpdate` when status returns to ACTIVE.
- **Webhook secrets are in-memory only.** The `webhookSecrets` map is not persisted. If the server restarts, all registered secrets are lost and must be re-registered.
- **SubscriptionEvent records are best-effort.** If recording the lifecycle event fails, the error is logged but does not cause the webhook processing to fail. This means the subscription update succeeds but the audit trail may have gaps.

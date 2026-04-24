# 11. Notification Service

## What It Does
Sends multi-channel notifications to users when important events occur in their Shopify app portfolio. Two notification types are supported:

- **Critical alerts** — triggered when a subscription's risk state changes (e.g., SAFE to ONE_CYCLE_MISSED, or any state to CHURNED). These include the app name, store domain, and the old/new risk states.
- **Daily summary digests** — sent once per day with a snapshot of MRR, revenue at risk, and renewal success rate.

Delivery channels:
- **Firebase Cloud Messaging (FCM)** — push notifications to iOS, Android, and web devices. Always available.
- **Slack webhooks** — rich-formatted messages with color-coded attachments. Pro tier feature, added via builder pattern.

Users control what they receive through `NotificationPreferences`, which default to both critical and daily summary enabled.

## Architecture
Application layer service (`internal/application/service/`). The `NotificationService` depends on interfaces rather than concrete implementations:

- `PushNotificationProvider` — abstraction over FCM with `SendPush(ctx, token, platform, title, body)`.
- `SlackNotifier` — abstraction over Slack webhook delivery with `SendSlack(ctx, webhookURL, title, body, color)`.
- `DeviceTokenRepository` — CRUD operations for device tokens.
- `NotificationPreferencesRepository` — CRUD and upsert for user notification settings.

The Slack notifier is optional, added via the builder method `WithSlackNotifier()`. This means the service can be constructed without Slack support and upgraded later without changing the constructor signature.

Infrastructure implementations:
- `FirebaseMessagingService` (external/) — wraps the Firebase Admin SDK's messaging client. Handles platform-specific payloads (APNS for iOS, high-priority Android with `FLUTTER_NOTIFICATION_CLICK` action).
- `SlackNotificationProvider` (external/) — sends JSON payloads to Slack incoming webhook URLs with colored attachments and a "LedgerGuard" footer.

## Key Files
| File | Purpose |
|------|---------|
| `backend/internal/application/service/notification_service.go` | NotificationService: RegisterDevice, UnregisterDevice, SendCriticalAlert, SendDailySummary, GetPreferences, UpdatePreferences |
| `backend/internal/infrastructure/external/firebase_messaging.go` | FirebaseMessagingService: FCM push with platform-specific config (APNS, Android) |
| `backend/internal/infrastructure/external/slack_provider.go` | SlackNotificationProvider: Slack webhook delivery with colored attachments |
| `backend/internal/domain/entity/notification_preferences.go` | NotificationPreferences entity: CriticalEnabled, DailySummaryEnabled, DailySummaryTime, SlackWebhookURL |
| `backend/internal/domain/entity/device_token.go` | DeviceToken entity: UserID, DeviceToken, Platform (ios/android/web) |

## Data Flow

### Device Registration
```
RegisterDevice(ctx, userID, deviceToken, platform)
│
├── Validate platform (ios, android, web)
│     └── Invalid → return ErrInvalidPlatform
│
├── Check if token already exists in DB
│     ├── Exists, same user → no-op, return nil
│     ├── Exists, different user → delete old, create new
│     └── Not found → create new
│
└── Ensure NotificationPreferences exist
      ├── Found → done
      └── Not found → create with defaults (critical=true, dailySummary=true, time=8:00 UTC)
            └── Handles duplicate key race condition gracefully (PostgreSQL error 23505)
```

### Critical Alert Delivery
```
SendCriticalAlert(ctx, userID, appName, storeDomain, oldState, newState)
│
├── Load user's NotificationPreferences
│     └── Not found → use defaults (critical enabled)
│
├── Check prefs.ShouldSendCritical()
│     └── False → return nil (user disabled critical alerts)
│
├── Build content:
│     title = "Risk Alert: {appName}"
│     body  = "{storeDomain} changed from {oldState} to {newState}"
│
├── If slackNotifier != nil AND prefs.SlackWebhookURL != "":
│     └── slackNotifier.SendSlack(webhookURL, title, body, color=RED)
│
└── For each device token (FindByUserID):
      └── pushProvider.SendPush(token, platform, title, body)
```

### Daily Summary Delivery
```
SendDailySummary(ctx, userID, appName, snapshot)
│
├── Load preferences → check ShouldSendDailySummary()
│
├── Build content:
│     title = "Daily Summary: {appName}"
│     body  = "MRR: ${mrr} | At Risk: ${atRisk} | Renewal Rate: {rate}%"
│
├── Slack (if configured) → color=BLUE (info)
│
└── Push to all registered devices
```

## Configuration
| Setting | Value | Notes |
|---------|-------|-------|
| Default critical alerts | Enabled | New users get critical alerts on by default |
| Default daily summary | Enabled | New users get daily summary on by default |
| Default summary time | 8:00 AM UTC | Stored in preferences, not yet used for scheduling |
| Slack HTTP timeout | 10 seconds | Hardcoded in SlackNotificationProvider |
| iOS sound | "default" | APNS payload includes sound and badge=1 |
| Android priority | "high" | Android config includes FLUTTER_NOTIFICATION_CLICK action |
| Firebase credentials | File path | Passed to `NewFirebaseMessagingService` constructor |

### Slack Color Constants
| Color | Hex | Usage |
|-------|-----|-------|
| Danger | `#dc3545` | Critical alerts (risk state changes) |
| Warning | `#ffc107` | Warnings |
| Success | `#28a745` | Success confirmations |
| Info | `#17a2b8` | Daily summaries |

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/notifications/devices` | Firebase | Register a device token for push notifications |
| DELETE | `/api/v1/notifications/devices` | Firebase | Unregister a device token |
| GET | `/api/v1/notifications/preferences` | Firebase | Get notification preferences |
| PUT | `/api/v1/notifications/preferences` | Firebase | Update notification preferences |

Internal service methods (called by other services, not directly by HTTP):
- `SendCriticalAlert()` — called by `WebhookService` on risk state changes
- `SendDailySummary()` — called by the sync/snapshot pipeline after daily metrics are computed

## Extension Points
- **New notification channels** — implement the `PushNotificationProvider` or `SlackNotifier` interface for email, SMS, Microsoft Teams, Discord, etc. The service dispatches to all configured providers.
- **Notification types** — add methods like `SendWeeklySummary()`, `SendMilestoneAlert()`, `SendAnomalyDetection()` following the same pattern: check preferences, build content, dispatch to channels.
- **Scheduling** — `DailySummaryTime` is stored in preferences but not yet wired to a scheduler. A cron job or background worker could use this to send summaries at the user's preferred time.
- **Notification history** — add a `notification_log` table to track what was sent, when, and to which channel. Currently notifications are fire-and-forget.
- **Rate limiting** — add per-user or per-channel rate limits to prevent notification spam during rapid risk state changes.
- **Multicast** — `FirebaseMessagingService` has a `SendMulticast()` method that sends to multiple tokens in a single API call. Currently unused; the service loops and sends individually.

## Gotchas
- **Last error wins.** Both `SendCriticalAlert` and `SendDailySummary` iterate over all device tokens and Slack, but only return the last error encountered. If push to device A fails and device B succeeds, the error from device A is lost. Partial delivery is not reported.
- **Slack color constants are duplicated.** They are defined in both `notification_service.go` and `slack_provider.go`. Changes to one file's constants do not affect the other.
- **Device token ownership transfer.** If a device token already exists for a different user (e.g., user logged out and new user logged in on same device), the old registration is deleted and a new one is created. The old user loses push notifications on that device silently.
- **Duplicate key race condition.** When creating default `NotificationPreferences`, a concurrent registration for the same user can trigger a PostgreSQL unique violation (error 23505). The service detects this by string-matching the error message, not by using a typed error. This works for PostgreSQL but may not work for other databases.
- **UnregisterDevice checks ownership.** A user can only delete their own device tokens. If the token belongs to another user, `ErrDeviceTokenNotFound` is returned rather than a permission error, which obscures the real issue.
- **DailySummaryTime is stored but unused.** The `DailySummaryTime` field in preferences exists in the entity and database, but no scheduler checks it. Daily summaries are triggered by the sync pipeline, not by a time-based scheduler.
- **FCM credential file must exist at startup.** `NewFirebaseMessagingService` requires a credentials file path and fails immediately if the file is missing or invalid. There is no lazy initialization or fallback.

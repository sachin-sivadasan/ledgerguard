# 35. Notification Scheduler

## What It Does
Background goroutine that sends daily summary notifications at each user's preferred hour. Ticks every 15 minutes but deduplicates so each UTC hour is only processed once.

## How It Works

```
Server starts
│
└── NotificationScheduler.Start(ctx)
      │
      ├── Immediate checkAndSend()
      │
      └── Every 15 minutes:
            │
            ├── currentHour = now.UTC().Hour()
            │
            ├── currentHour == lastCheckedHour?
            │     └── Yes → skip
            │
            └── No → process this hour:
                  │
                  ├── Query: FindUsersWithDailySummaryAtHour(hour)
                  │     SQL: SELECT user_id FROM notification_preferences
                  │          WHERE daily_summary_enabled = true
                  │          AND EXTRACT(HOUR FROM daily_summary_time) = $1
                  │
                  ├── No users → return
                  │
                  └── For each user:
                        │
                        ├── FindByUserID → partnerAccount
                        │     └── Not found → skip
                        │
                        ├── FindByPartnerAccountID → []apps
                        │
                        └── For each app:
                              │
                              ├── FindLatestByAppID → snapshot
                              │     └── Not found → skip
                              │
                              └── SendDailySummary(userID, appName, snapshot)
                                    │
                                    ├── Check prefs.ShouldSendDailySummary()
                                    │     └── false → skip
                                    │
                                    ├── Build message:
                                    │     title: "📊 Daily Summary: {appName}"
                                    │     body:  "MRR: $X | At Risk: $Y | Renewal Rate: Z%"
                                    │
                                    ├── Slack webhook (if configured)
                                    │     POST → color: info (#17a2b8)
                                    │
                                    └── Push to all device tokens (FCM)
                                          iOS: APNS sound + badge
                                          Android: high priority
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/application/scheduler/notification_scheduler.go` | Scheduler: 15-min tick, hour dedup, user iteration |
| `internal/application/service/notification_service.go` | SendDailySummary: prefs check, message build, dispatch |
| `internal/domain/entity/notification_preferences.go` | Entity: DailySummaryEnabled, DailySummaryTime |
| `internal/infrastructure/persistence/notification_preferences_repository.go` | FindUsersWithDailySummaryAtHour SQL query |
| `internal/infrastructure/external/firebase_messaging.go` | FCM push with platform-specific config |
| `internal/infrastructure/external/slack_provider.go` | Slack webhook delivery |

## Configuration

| Setting | Value | Notes |
|---------|-------|-------|
| Check interval | 15 minutes | `checkInterval` field, configurable via `SetCheckInterval()` |
| Hour dedup | `lastCheckedHour` | Ensures each UTC hour is processed at most once |
| Default summary time | 8:00 AM UTC | Set in `NewNotificationPreferences()` |
| Default enabled | true | Both critical and daily summary enabled by default |

## Lifecycle

| Event | Behavior |
|-------|----------|
| Server start | `scheduler.Start(ctx)` — launches goroutine, runs first check immediately |
| Every 15 min | Ticker fires → `checkAndSend()` |
| New hour | Queries DB for eligible users, sends summaries |
| Same hour | Skipped (already processed) |
| Server shutdown | `scheduler.Stop()` — closes `stopCh`, goroutine exits, closes `doneCh` |
| Context cancelled | Goroutine exits via `ctx.Done()` |

## Database Query

```sql
-- Find users whose preferred daily summary hour matches current UTC hour
SELECT user_id
FROM notification_preferences
WHERE daily_summary_enabled = true
  AND EXTRACT(HOUR FROM daily_summary_time) = $1
```

The `daily_summary_time` column stores a `TIME` value (e.g., `08:00:00`). Only the hour component is used for matching.

## Sequence Diagram
See `docs/diagrams/puml/35-notification-scheduler-sequence.puml`

## Gotchas
- **Hour-level precision only.** The minute component of `daily_summary_time` is ignored — summaries fire at the top of the matching hour (whenever the 15-min tick hits).
- **UTC only.** No timezone conversion. Users set their preferred hour in UTC.
- **No delivery tracking.** Summaries are fire-and-forget. No record of what was sent or when. If FCM fails, the error is logged but the summary is not retried.
- **Preferences queried twice.** The scheduler calls `FindUsersWithDailySummaryAtHour()`, then `SendDailySummary` calls `FindByUserID()` again to re-check `ShouldSendDailySummary()`. Minor redundancy.
- **Single-instance only.** No distributed lock — if multiple server instances run, each will send duplicate summaries. Needs Redis-based leader election for multi-node deployment.
- **Uses FindByUserID (not org-scoped).** The scheduler is a background goroutine with no HTTP context, so it uses `partnerRepo.FindByUserID()` directly instead of `resolvePartnerAccount`. This is intentional — org-scoped resolution requires `OrgContextMiddleware` which only runs in HTTP request pipelines.

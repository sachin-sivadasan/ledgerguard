# Tech Design: Notifications & Daily Insights (as-built)

**PRD:** docs/prds/09-notifications.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public prefs/device API shapes bind the Flutter client; changing them breaks settings + push registration. The delivery side is outbound (FCM/Slack) — reversible. No irreversible data model beyond additive tables. Main risk is **behavioral**: duplicate or missed sends erode trust in the channel.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — scheduler** (`notification_scheduler.go`): a ticker runs every 15 min; `checkAndSend` (`:81-111`) returns early if `currentHour == lastCheckedHour` (**in-memory** guard, `:86-90`), else queries `FindUsersWithDailySummaryAtHour(currentHourUTC)` and, per user → per app, loads the latest snapshot and calls `SendDailySummary` (`:113-143`). An admin `RunForHour` bypasses the guard.

**FACT — service** (`notification_service.go`): `SendCriticalAlert` (`:141-187`) and `SendDailySummary` (`:189-237`) load the user's devices + prefs, push to **all** devices via FCM (`firebase_messaging.go`, platform-aware APNS/Android), and post to Slack if `SlackWebhookURL` is set (`slack_provider.go`, color-coded, 10s timeout). Slack failure does **not** block push (tested). Device registration (`:69-111`) validates platform, transfers a token that moved users, and auto-creates default prefs.

**FACT — email is unwired.** `EmailEnabled` exists in prefs + schema, but a repo-wide search finds **no SMTP/SendGrid/Mailgun/mail-send code** anywhere in `backend/` — enabling it sends nothing.

**FACT — AI daily insight (separate concept)** (`ai_insight_service.go`): Pro-tier-gated; builds a prompt from the daily snapshot; upserts an 80–120-word brief into `daily_insight` (`ON CONFLICT (app_id, date)`), shown in-app. It is **on-demand only** — no background job generates it, and the daily-summary notification sends **snapshot metrics, not this insight**.

**FACT — security.** Firebase credentials are git-ignored (`**/firebase-credentials.json`, `.gitignore:35`) and no credential JSON is tracked — **not committed** (verified).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **hourly UTC scheduler → per-user/per-app fan-out → multi-channel send (push + Slack)**, plus an event-driven critical-alert path and a decoupled on-demand AI-insight generator. See `docs/designs/09-notifications-sequence.puml` (validated `plantuml -checkonly`) — 3+ components (scheduler, service, FCM, Slack, AI) interact.

### Data model

**FACT — additive tables:** `notification_preferences` (per-user unique; toggles critical/daily/churn/revenue/review, `daily_summary_time` TIME, `slack_webhook_url`, `email_enabled`, `slack_enabled`, `risk_threshold_days`; migrations 000009 + 000039); `device_tokens` (user_id, `device_token` unique, platform check ios|android|web; migration 000008 — **no `expires_at`/TTL/soft-delete**); `daily_insight` (unique `(app_id, date)`; migration 000007). No new migration for this backfill.

### API & events

**FACT.** `GET/PUT /api/v1/users/notification-preferences` (full prefs object, `daily_summary_time` as "HH:MM"); `POST /api/v1/devices` (`{device_token, platform}` → 201, auto-creates prefs); `DELETE /api/v1/devices` (→ 204, owner-checked). Admin `POST /admin/notifications/daily-summary` triggers `RunForHour`. Outbound: FCM push + Slack webhook. No inbound webhooks.
**Breaking-change check:** prefs field names bind the settings screen; the `HH:MM` time format is a contract.

## Alternatives considered

- **In-memory hour guard vs persisted send-log.** As-built uses `lastCheckedHour` (`:86`) — simple, but a restart mid-hour re-fires. A `notification_sends` dedupe table was **not recorded as considered**. **Choose a persisted send-log if** at-most-once delivery matters (metric #1).
- **Snapshot-templated summary vs AI-insight summary.** The notification sends templated snapshot metrics; the richer AI insight exists but isn't sent (decoupled). **Choose AI-in-notification if** the insight proves more engaging — requires scheduling insight generation first.
- **FCM + Slack now, email later vs all three at launch.** Email was scaffolded (flag/schema) but deferred (no provider). **Choose to implement email if** partners request it — or remove the flag to avoid a false promise.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| User has no devices | empty device list | **FACT** graceful no-op (tested) | register device |
| Slack webhook fails | HTTP non-200 | **FACT** logged; push still sent (`notification_service_test.go:596-757`) | fix webhook |
| No snapshot for app | repo err | **FACT** skip that app, continue (`scheduler:132-135`) | wait for sync |
| Duplicate token across users | transfer logic | **FACT** token reassigned to new user (`service:69-111`) | — |
| Pref disabled | pref check | **FACT** send skipped per type | — |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. In-memory idempotency → duplicate sends on restart.** A process restart within the hour resets `lastCheckedHour`; users at that hour get a second summary. → proposed test + persisted send-log (metric #1).
- **D2. Daily summary ≠ AI insight; insight has no scheduler.** Two "insight" concepts diverge; the AI brief is generated only on demand, so many app-days have none. → open question (schedule generation + include in notification?).
- **D3. No dead-token cleanup.** Invalid/expired FCM tokens are never pruned (no `expires_at`, no prune-on-failure) → DB bloat + wasted sends. → proposed prune-on-send-failure (metric #2).
- **D4. Email is a stub.** `EmailEnabled` is settable in the UI but sends nothing — a silent false promise. → implement or remove.
- **D5. UTC-only send time.** `DailySummaryTime` + scheduler use UTC; "08:00" is not the user's local morning. → add per-user timezone.
- **D6. No threshold/anomaly alerting.** `RiskThresholdDays` is stored but there's no "MRR dropped X%" engine. → open question (Product).
- **D7. Scheduler untested.** No `notification_scheduler_test.go` (service + Slack are tested; the hour-guard/fan-out is not). → add tests.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Exactly-once daily send across a simulated restart (metric #1) — **TODO**, needs the persisted send-log to pass.
2. Dead-token prune keeps push-failure rate < 5% (metric #2) — **TODO**, needs D3.

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- No-devices no-op, Slack-fails-push-continues, pref-disabled-skip, token-transfer — **exist** (`notification_service_test.go`). Slack payload/error/timeout — `slack_provider_test.go`.

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 restart duplicate-send · D3 invalid-token prune · D5 timezone send-time · D7 scheduler hour-guard/fan-out. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/application/service/ -run Notification -v && go test ./internal/infrastructure/external/ -run Slack -v`
2. Set `daily_summary_time` to the next UTC hour in the settings screen; register a device; wait for the tick (or hit `POST /admin/notifications/daily-summary`) → confirm one push + Slack message.
3. Set a Slack webhook only (no device) → confirm Slack-only delivery.
4. Toggle `email_enabled=true` → confirm nothing is emailed (documents D4).

**External-platform reality:** FCM and Slack are the external dependencies; unit tests fake the providers below the client interface. Real push delivery + APNS behavior + Slack rendering MUST be smoke-tested against a real device + a real webhook — never fully reproducible in unit tests.

## Pricing & policy touchpoints

**FACT:** AI daily-insight generation is **Pro-tier-gated** (`ai_insight_service.go`). **UNKNOWN** whether notification channels are tiered. Compliance: opt-in prefs are the only consent mechanism (no separate CMP) — note for GDPR review. → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Additive; default prefs auto-created. Deployed on Hetzner ([[hetzner-migration]]); the scheduler runs in-process with the API. **Before scaling to multiple instances:** the in-memory hour-guard (D1) means each instance would fire — a persisted send-log becomes required, same class of issue as the Revenue API rate limiter. Rollback = revert; no data stranded.

## Observability

**FACT:** rich `log.Printf` through the scheduler (users found, per-app send result with MRR, `:91-141`) and send failures. **DIVERGENCE:** no metric/alert on send success rate, duplicate sends, or dead-token rate. On-call greps Hetzner logs. → proposed observability TODO tied to metrics #1/#2.

## Open questions

- **Persisted send-log for at-most-once (D1)** — required before multi-instance. Owner: Eng.
- **Schedule AI-insight generation + include in notifications (D2)?** Owner: Product/Eng.
- **Dead-token prune (D3).** Owner: Eng.
- **Email: implement or remove (D4).** Owner: Eng/Product.
- **Per-user timezone (D5).** Owner: Eng/Product.
- **Threshold/anomaly alerts (D6)?** Owner: Product.
- Notification-channel tiering + GDPR consent review. Owner: Product.

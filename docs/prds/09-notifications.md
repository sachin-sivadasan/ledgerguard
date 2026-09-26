# PRD: Notifications & Daily Insights (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/application/service/notification_service.go`, `scheduler/notification_scheduler.go`,
> `service/ai_insight_service.go`, `external/{firebase_messaging,slack_provider}.go`,
> `handler/{notification_preferences,device_handler}.go`, `frontend-flutter/lib/` (notification settings + insights).

## Problem & evidence

A Shopify app partner won't log into a dashboard daily — they need LedgerGuard to **push the important stuff to them**: a critical risk change ("a store just entered a missed cycle") and a daily revenue pulse, wherever they already work (phone, Slack).

- **FACT:** two notification types are fully built and tested (critical risk alert, daily summary) across push + Slack.
- **UNKNOWN — original demand evidence not recorded.** Open question: validate which alerts partners actually want — owner: Product.

## Target users

**INFERRED:** the **Shopify app partner** (founder/growth) who wants proactive, low-effort awareness. Per-user preferences within an org.
- **Not the target:** merchants; end-customers; high-frequency/real-time trading-style alerting (this is daily + event-driven risk changes).

## Proposed solution (as-built behavior)

**FACT — two notifications** (`notification_service.go`):
- **Critical Risk Alert** ("🚨 Risk Alert: {App}") on risk-state changes, gated by `CriticalAlertsEnabled`; sent to **all the user's devices (push)** + **Slack** if a webhook is set (`:141-187`).
- **Daily Summary** ("📊 Daily Summary: {App}") with MRR / at-risk / renewal rate at the user's chosen time, push + Slack (`:189-237`).

**FACT — channels:** **Push (FCM)** and **Slack** are fully implemented (`firebase_messaging.go`, `slack_provider.go` — color-coded attachments, 10s timeout) and unit-tested. **Email is a stub** — `EmailEnabled` exists in prefs + schema but **no mail provider is wired**, so enabling it silently sends nothing.

**FACT — configuration:** a notification-settings screen calls `GET/PUT /api/v1/users/notification-preferences`; prefs include per-type toggles (critical/daily/churn/revenue/review), `DailySummaryTime`, channel flags (email/slack), `SlackWebhookURL`, and `RiskThresholdDays` (`notification_preferences.go`). Devices register via `POST /api/v1/devices` (token + platform ios/android/web), auto-creating default prefs.

**FACT — daily insights (separate concept):** an **AI-generated 80–120-word brief per app per day** (`ai_insight_service.go`, Pro-tier-gated, prompt built from the daily snapshot, upserted one-per-app-per-day) is stored in `daily_insight` and shown in-app. **Note:** the *daily summary notification* sends **snapshot metrics, not this AI insight** — the two are decoupled (see Open questions).

### Key screens

**No new wireframe** (as-built). Real surfaces: the Flutter notification-settings screen (prefs + Slack webhook) and the in-app daily-insights display.

## Platform & policy constraints

**FACT.** Push depends on Firebase Cloud Messaging (credentials/config); Slack depends on a per-user incoming webhook. Delivery timing is **UTC-based** — `DailySummaryTime` and the scheduler both use UTC hours (no per-user timezone). Data comes from the daily snapshot (bounded by sync).

## Pricing-tier impact

**FACT (partial):** AI **daily insight generation is Pro-tier-gated** (`ai_insight_service.go`). **UNKNOWN** whether notification *channels* (Slack/push/future email) are tiered. Open question — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Additive; default prefs auto-created on first device registration. No migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Delivery reliability:** ≥ 99% of due daily summaries are sent exactly once per day per opted-in user (guards the in-memory-idempotency restart risk).
2. **No dead sends:** push send-failure rate (invalid/expired tokens) < 5% — implies dead-token cleanup exists (it doesn't yet).
3. **Actionability:** ≥ 20% of critical risk alerts are followed by the user opening the app within 24h.

## Non-goals (deliberately absent in code)

1. **No email delivery** — flag exists, no provider; nothing is sent. **FACT.**
2. **No timezone-aware scheduling** — UTC-only send times. **FACT.**
3. **No AI insight in the daily notification** — the notification uses templated snapshot metrics; the AI insight is in-app only and generated on demand. **FACT.**
4. **No threshold/anomaly alerts** — `RiskThresholdDays` exists as a pref but there's no "MRR dropped X%" alerting engine. **INFERRED.**

## Open questions

- **Email is a stub** — implement a provider (SendGrid/SMTP) or remove `EmailEnabled` from the UI so it doesn't mislead. Owner: Eng/Product.
- **Scheduler idempotency is in-memory** (`lastCheckedHour`) — a restart within the hour risks duplicate sends; add a persisted send-log? Owner: Eng.
- **Daily insight vs daily summary are decoupled** — should the notification include the AI insight, and should insight generation be a scheduled background job (it's on-demand today)? Owner: Product/Eng.
- **No dead-token cleanup** — invalid FCM tokens accumulate; prune on send-failure? Owner: Eng.
- **UTC-only send time** — add per-user timezone so "08:00" means their morning. Owner: Eng/Product.
- **Verify Firebase credentials are git-ignored** (not committed). Owner: Eng/Sec.
- Notification-channel tiering by plan? Owner: Product.

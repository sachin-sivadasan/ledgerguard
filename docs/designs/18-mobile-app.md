# Tech Design: Mobile Push Client & Store Readiness

**PRD:** docs/prds/18-mobile-app.md · **Author:** sachin.s · **Date:** 2026-09-27
**Status:** Draft
**Reversibility:** Two hard-to-undo bits: (1) the **store bundle IDs** — once an app is
published to App Store / Play, its ID is permanent; picking the wrong one is unfixable
without a new listing. (2) the **`PushNotificationProvider.SendPush` interface** gains a
`data` param — an internal Go interface change touching all call sites. The Flutter push
wiring itself is reversible.

## Current state

- **Client:** `frontend-flutter` (Provider, live) is a `go_router` `StatefulShellRoute` app
  (`lib/app.dart:76-114,415-419` — `MaterialApp.router`) with an adaptive shell
  (`lib/shell/app_shell.dart`). It depends on `firebase_core` + `firebase_auth` only
  (`pubspec.yaml`) — **no `firebase_messaging`**, and nothing calls the device endpoint, so
  the app cannot receive push. Bundle IDs are Flutter defaults
  (`android/app/build.gradle.kts:24` `com.example.ledgerguard_flutter`;
  `ios/.../project.pbxproj:371` `com.example.ledgerguardFlutter`).
- **Backend (exists, don't rebuild):** `POST /api/v1/devices` (`handler/device_handler.go:31`,
  body `{device_token, platform}`, `platform ∈ ios|android|web` per `entity/device_token.go:13`)
  → `NotificationService.RegisterDevice`; plus `UnregisterDevice` (`:77`). FCM send is
  `FirebaseMessagingService.SendPush(ctx, token, platform, title, body)`
  (`external/firebase_messaging.go:35`), called from `notification_service.go:181,231`
  (critical-risk + daily-summary). **The message carries only `Notification{Title,Body}` —
  no `data` map** (`firebase_messaging.go:36-79`), so a tap has nothing to route on.

## Proposed design

Two workstreams; happy path in `18-mobile-app-sequence.puml`.

**A. Push client (`frontend-flutter`).** Add `firebase_messaging`. A new
`PushNotificationService` (a Provider, mirroring existing service style): after sign-in,
request OS permission, `getToken()`, and `POST /api/v1/devices {device_token, platform}`
reusing the existing `ApiClient`; subscribe to `onTokenRefresh` → re-register; on
`signOut()` (`providers/auth_provider.dart:123`) call unregister **then** Firebase sign-out.
Register foreground/background/tap handlers; on tap, read the message **`data`** and
`context.go(...)` the matching route (e.g. `data.type=risk` → the subscription/app screen),
falling back to Dashboard when data is missing/invalid or the `app_id` isn't in the user's org.

**B. Backend deep-link payload.** Extend `SendPush` (and the `PushNotificationProvider`
interface) with `data map[string]string`; populate `{type, app_id, subscription_id|store_id}`
at the two call sites so the client can route. Add **dead-token pruning**: on an FCM
`Unregistered`/`InvalidArgument` send error, delete that `device_token` (closes the PRD #9
"no dead-token cleanup" gap).

**C. Store readiness (config, no runtime logic).** Real bundle IDs (e.g.
`com.ledgerspear.app`) in Gradle + Xcode; app signing (keystore / provisioning);
iOS **APNs key uploaded to Firebase** (required for FCM on iOS) + `aps-environment` entitlement;
launcher icons + splash; `NSUserTrackingUsageDescription`-style privacy strings and the iOS
**PrivacyInfo** manifest. No store *publishing pipeline* in this design (out of scope).

### Data model

**No new tables.** `device_tokens` already exists (`entity/device_token.go`) with
`{user_id, device_token, platform, created_at}`. Change is behavioral: register is an
**upsert** on `device_token` (idempotent — a re-sent token must not duplicate rows;
confirm the repo does ON CONFLICT), and dead tokens are **deleted** on send failure. No
migration/backfill; existing rows are unaffected.

### API & events

- **`POST /api/v1/devices`** — unchanged contract `{device_token, platform}` → 200. (Exists.)
- **Unregister on logout** — reuse the existing unregister handler (`device_handler.go:77`).
- **Internal (breaking within the module):** `PushNotificationProvider.SendPush(...)` gains
  `data map[string]string`. Not a public/external API — no third-party integration observes it.
  FCM message shape gains a `data` block **alongside** the existing `notification` block, so
  current title/body behavior is unchanged for clients that ignore data.

## Alternatives considered

1. **Data-only FCM messages (no `notification` block), render locally with
   `flutter_local_notifications`.** Rejected: more client code and background-handler
   complexity; the backend already sends `notification` blocks that work today. *Choose this
   instead if* we need full control of notification styling or actions per platform.
2. **Deep-link via a URL string field instead of typed `data` keys.** Rejected: couples the
   backend to client route paths. Typed keys (`type`, `app_id`) let the client own routing.
   *Choose this instead if* web and mobile must share identical link URLs.
3. **Skip deep-linking (tap just opens the app to Dashboard).** Rejected for v1 of the
   feature since the PRD metric is "delivers to the right context," but it's the trivial
   fallback already built into option A. *Choose this instead if* we must ship push before
   the backend `data` change lands.

## Failure modes

| Failure | Detection | Behavior | Recovery |
|---|---|---|---|
| OS notification permission denied | `requestPermission()` result | App works fully; no token registered | Re-prompt from a settings toggle |
| `getToken()` / register call fails (offline, backend down) | non-200 / exception | App continues; push silently off | Retry on next launch + on `onTokenRefresh` |
| FCM token rotates | `onTokenRefresh` event | Old token goes stale | Re-register new token; old pruned on next send failure |
| Send to dead/expired token | FCM `Unregistered`/`InvalidArgument` on `SendPush` | That push lost | **Delete** the token (pruning) so it isn't retried |
| Duplicate registration (same token) | repo upsert | No duplicate row | Idempotent upsert |
| Tap with missing/invalid `data` or cross-org `app_id` | route lookup / org check | No crash | Fall back to Dashboard |
| Logout on a shared device | `signOut()` path | Prior user must stop getting alerts | Unregister token **before** Firebase sign-out |

## Test plan

1. **Acceptance (from PRD metrics):**
   - Push round-trip: a simulated risk-change send reaches a registered token with the
     correct `data` and taps into the target route (integration test w/ fake FCM).
   - Nav parity unaffected; permission-denied path still renders the app.
2. **Adversarial (one per failure row):** permission denied → no register; register failure →
   retry-on-relaunch; token refresh → re-register; dead-token send → row deleted; duplicate
   register → single row; malformed/cross-org `data` → Dashboard fallback; logout → token
   unregistered before sign-out.
3. **Manual script:** (1) sign in on a device/emulator → accept permission; (2) confirm a
   `device_tokens` row exists for the user; (3) trigger a risk change (or admin re-send) →
   notification arrives; (4) tap → lands on the right subscription/app screen; (5) sign out →
   confirm the row is gone; (6) send again → no delivery.

**FCM reality:** unit/integration tests fake the FCM SDK below the `PushNotificationProvider`
(backend) and below the `PushNotificationService` (client). **Must** smoke-test on a **real
device** against a Firebase project for: iOS APNs delivery, background/killed-state taps, and
permission prompts — these cannot be emulated faithfully.

## Pricing & policy touchpoints

**FACT (PRD #9 open item):** whether push is tier-gated is UNKNOWN — owner: Product. **Store
policy:** App Store + Play require a privacy manifest / data-safety disclosure covering the
device token + analytics (Mixpanel); iOS requires APNs. No WhatsApp/Shopify policy here.

## Rollout

Order: (1) backend `data` payload + dead-token pruning (backward-compatible — old clients
ignore `data`); (2) ship client `firebase_messaging` behind app-store release with real
bundle IDs; (3) enable. **No server flag needed** — an unregistered client simply gets no
push. **Rollback:** reverting the client stops registration; backend `data` is additive and
safe to leave. Bundle-ID choice is **not** rollback-able post-publish — decide before first submit.

## Observability

Tie to PRD metrics via the logs-over-MCP stack ([[logs-mcp-setup]] / `docs/LOGS_MCP_SETUP.md`,
once live): structured logs on the backend send path — `push_sent` (with `app_id`, `type`,
result), `push_token_pruned`, and `device_registered`. On-call for "I didn't get a
notification": query the log index by `user_id`/`app_id` for the send result + prune events;
check `device_tokens` for a live row. Client: log permission-grant + register outcome.

## Open questions

- **Bundle ID + Apple/Google developer accounts + signing ownership** — who owns store setup? Owner: Eng/Product.
- **iOS APNs key** provisioning into Firebase — owner: Eng.
- **Tier-gating** of push — owner: Product.
- **Deep-link route map** — exact `data.type` → route table (risk, daily-summary, churn…) — owner: Eng/Design.
- **Per-user timezone** for daily summary (PRD #9 notes UTC-only) — fix alongside? Owner: Eng.
- **Notification settings screen** already calls `notification-preferences`; does it also need a device/permission status row? Owner: Design.

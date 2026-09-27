# PRD: Mobile App (iOS / Android) (as-built)

**Date:** 2026-09-27 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `frontend-flutter/lib/shell/app_shell.dart`, `lib/theme/app_breakpoints.dart`,
> `lib/core/config/app_config.dart`, `frontend-flutter/pubspec.yaml`,
> `android/app/build.gradle.kts`, `ios/Runner.xcodeproj/project.pbxproj`.
> (Push/notifications: PRD #9. Auth: PRD #15. The dead Bloc app `frontend/app/` is out of scope.)

## Problem & evidence

A Shopify app partner is often away from a desk — they want their revenue/risk cockpit
**on their phone**, and to be pinged when something breaks, not only when they open a
laptop. The product is architected for that: **one Flutter codebase → Web + iOS + Android**.

- **FACT:** the live app (`frontend-flutter/`, Provider) has full `android/` + `ios/`
  projects and CI builds a release **APK** (`.github/workflows/ci.yml`, `flutter build apk`).
- **FACT:** the UI is genuinely adaptive, not just web-shrunk (see solution).
- **UNKNOWN — demand evidence not recorded.** Open question: is native mobile a shipping
  target or a by-product of Flutter? — owner: Product. (PRD.md lists "Native mobile app"
  under **roadmap**, and `frontend/REQUIREMENTS.md` scopes the app "Web-first".)

## Target users

**INFERRED:** the **Shopify app partner** (founder / growth / success) checking revenue,
risk, and store health on the go, per-app and per-org — the same user as the web app,
on a phone. **Not the target:** merchants; tablet-first or offline-first use.

## Proposed solution (as-built behavior)

**FACT — adaptive shell** (`app_shell.dart:77-200`, `app_breakpoints.dart`): the same
screens re-flow by width — **mobile (<600px) → BottomNavigationBar**, **desktop → NavigationRail**;
metric grids drop to **2 columns on mobile** (3 tablet, 4 desktop, `app_breakpoints.dart:25-30`).
Mobile bottom-nav exposes 10 primary sections (Dashboard, Subscriptions, Stores,
Transactions, Events, Webhooks, Risk, Analytics, Earnings, Apps; `app_shell.dart:14-23`).
- **FACT — routing:** `go_router` (`pubspec.yaml`), so URLs/back-stack work on all platforms.
- **FACT — config:** API host is compile-time via `--dart-define=API_BASE_URL`
  (`app_config.dart:7`), so the mobile build points at `api.ledgerspear.com`.
- **FACT — auth:** Firebase Auth works on device (`firebase_auth` in `pubspec.yaml`;
  PRD #15) — sign-in, session, logout are shared with web.

**FACT — the push gap (most important):** the backend fully sends **FCM push**
(critical-risk + daily-summary) and exposes device registration (PRD #9), **but the live
mobile app cannot receive it** — `firebase_messaging` is **not a dependency**
(`pubspec.yaml` has only `firebase_core` + `firebase_auth`), and nothing calls the
device-registration endpoint (`grep` of `lib/` finds no token registration). So today the
"push the important stuff to their phone" promise is unfulfilled on the client side.

### Key screens

**Adapted — no new wireframe** (as-built, per the `docs/prds/` convention). The real UI is
the responsive shell (`app_shell.dart`) rendering the existing 60 screens; mobile chrome =
bottom nav + 2-col metric grids. Visual language is Shopify-Polaris purple (`theme/app_colors.dart`).

## Platform & policy constraints

**FACT — not release-ready for the stores.** Bundle identifiers are still the Flutter
defaults: `com.example.ledgerguard_flutter` (`android/app/build.gradle.kts:24`) and
`com.example.ledgerguardFlutter` (`ios/.../project.pbxproj:371`). App Store / Play Store
submission requires real IDs, signing, icons/splash, privacy disclosures, and (for push)
APNs + `firebase_messaging`. **INFERRED:** distribution today is the **web PWA**
(`app.ledgerspear.com`) and side-loaded APKs, not the stores.

## Pricing-tier impact

**UNKNOWN — no mobile-specific plan gating found.** Feature access mirrors the web app
(same providers/services). Open question — owner: Product.

## Migration for existing users

**INFERRED — none.** Same account, same Firebase Auth, same backend; a user signs into the
mobile build with existing credentials. No data migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Push round-trip:** once `firebase_messaging` + device registration ship, a critical-risk
   event delivers to a registered device in ≤ 60s, p95 (closes the PRD #9 client gap).
2. **Mobile parity:** every web nav destination is reachable and renders without horizontal
   scroll at 390px width (observable check in a widget/golden test).
3. **Store-readiness:** a signed build with production bundle IDs passes store validation
   (no `com.example.*`, required privacy manifest present).

## Non-goals (deliberately absent in code)

1. **No push on device** — receiving/notification-tap handling is not built (`firebase_messaging`
   absent). **FACT.**
2. **No offline mode** — all screens are live API reads; no local cache/persistence. **INFERRED.**
3. **No store distribution pipeline** — no fastlane/signing/CI publish; APK build is
   `continue-on-error` in CI. **FACT.**
4. **No mobile-only features** — mobile is a responsive projection of the web app, not a
   distinct product surface. **INFERRED.**
5. **Chat & billing not present** — those two features live only in the dead Bloc app and
   were never ported (see [[frontend-feature-gap-and-ci]]); absent on mobile too. **FACT.**

## Open questions

- **Is native mobile a shipping target** (stores) or just a Flutter by-product? Owner: Product.
- **Push client**: add `firebase_messaging` + device-token registration (`POST /api/v1/devices`)
  and notification-tap deep links (go_router)? This is the highest-leverage gap. Owner: Eng/Product.
- **Production bundle IDs + signing + store assets** — who owns store setup? Owner: Eng.
- **Tablet layout**: breakpoints define a `tablet` tier, but is the tablet experience
  designed or just "not-mobile"? Owner: Design.
- **Offline / poor-connectivity** behavior on mobile — acceptable to require live network? Owner: Product.
- **Pricing/tier gating** on mobile. Owner: Product.

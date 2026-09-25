# PRD: Dashboard & KPIs (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/interfaces/http/handler/metrics.go`, `application/service/metrics_aggregation_service.go`,
> `domain/service/metrics_engine.go`, `frontend-flutter/lib/screens/dashboard/`, `providers/dashboard_provider.dart`.

## Problem & evidence

A partner opening LedgerGuard needs a **single at-a-glance answer** to "how is my revenue doing right now, and which way is it moving?" across their app — without running individual reports.

- **FACT:** the dashboard is the app's home surface and is live-wired to real metrics (deltas, trend, forecast all wired in commits #83/#84).
- **UNKNOWN — original demand evidence not recorded.** No tickets/usage metrics stored. Open question: validate the specific KPI set — owner: Product.

## Target users

**INFERRED:** the primary **Shopify app partner** persona (founder/growth), per-app within an org, who wants a fast pulse check before deciding whether to dig into reports.
- **Not the target:** merchants; users needing deep drill-down (that's the Reports platform, PRD #1).

## Proposed solution (as-built behavior)

**FACT.** The Dashboard (`dashboard_screen.dart`) shows:
- **Four primary KPI cards** (defaults, `dashboard_registry.dart:82-87`): **Active MRR**, **Renewal Rate**, **Revenue at Risk**, **Usage Revenue** — each with a **period-over-period delta badge** (real %, from the backend `delta` block).
- **Optional secondary widgets** (`dashboard_registry.dart`): MRR Trend (12-mo line), Risk Distribution (donut), Forecast (next month), Revenue Mix (stacked %), Weekly Activity, Earnings Overview.
- A **date-range selector** — This Week / This Month / Last Month / 3 Months (`dashboard_screen.dart:140-160`) — that re-fetches `/metrics` for the chosen window (trend & forecast are period-independent and not re-fetched).
- A **customizable layout** — add/remove/reorder KPI cards & widgets — **persisted per user** via backend UserPreferences (`/dashboard` GET/PUT, `router.go:138-139`).
- **Demo mode** — `DemoModeCoordinator` toggles all providers to mock data for a populated demo (`demo_mode_coordinator.dart`).

**FACT (KPI semantics).** Point-in-time KPIs (MRR, Revenue at Risk, Renewal Rate) come from the **end-of-period `daily_metrics_snapshot`**; period revenue (usage/total) is **summed from transactions**; deltas = `(current − previous) / previous × 100` over the prior equal-length window, **inverted for risk/churn** (a decrease is shown as good) (`metrics.go:535-543`).

### Key screens

**No new wireframe** (as-built). Real surface: `frontend-flutter/lib/screens/dashboard/dashboard_screen.dart` (KPI grid + widgets + date-range chips + customization). Wired via `dashboard_provider.dart` → `metrics_service.dart` → `GET /api/v1/apps/{appID}/metrics` (+ `/metrics/trend`).

## Platform & policy constraints

**FACT/INFERRED.** KPIs are derived from `daily_metrics_snapshot` (produced by the daily MetricsEngine from subscriptions + transactions) plus live transaction sums — no external calls at read time. The dashboard therefore reflects the **last successful sync + snapshot**; freshness is bounded by the sync pipeline. Forecast tile is gated on ≥90 snapshots (shows "not enough history" otherwise, `dashboard_screen.dart:404-407`).

## Pricing-tier impact

**UNKNOWN — no plan-gating found.** All KPIs/widgets available to any authenticated org member; layout customization is per-user. Open question: any tier limits on widgets? — owner: Product.

## Migration for existing users

**INFERRED.** Actively evolving (commits #83/#84 moved deltas/trend/forecast from mock to real). One residual: `GetLatestMetrics` is still mocked (`metrics.go:82`) — any surface relying on it shows placeholder data until wired. No data migration; layout prefs are additive.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Correctness:** dashboard Active MRR for a period equals the MRR report's headline for the same period (cross-surface consistency) in ≥99% of checks.
2. **Freshness:** ≥95% of dashboard loads reflect a snapshot ≤24h old (ties dashboard trust to the sync SLA).
3. **Engagement:** ≥50% of active orgs customize their layout (add/remove/reorder ≥1 card) within 30 days — validates that customization is worth its complexity.

## Non-goals (deliberately absent in code)

1. **No drill-down analytics on the dashboard** — it summarizes; depth lives in Reports (PRD #1). **INFERRED.**
2. **No real-time/streaming metrics** — snapshot + last-sync bound, not live. **FACT.**
3. **No cross-app blended dashboard as the default** — dashboard is per selected app (an `/metrics/aggregate` endpoint exists but the main dashboard is app-scoped). **FACT/INFERRED.**

## Open questions

- `GetLatestMetrics` still mocked (`metrics.go:82`) — what consumes it, and when is it wired? Owner: Eng.
- Validate the default KPI set + which widgets earn their place (metric #3). Owner: Product.
- Delta semantics: prior-equal-length window is the baseline — is that the right comparison for "This Week" vs "3 Months"? Owner: Product.
- Freshness signal: should the dashboard show "as of <snapshot date>" so stale data is visible? Owner: Eng/Product.
- Pricing/tier gating. Owner: Product.

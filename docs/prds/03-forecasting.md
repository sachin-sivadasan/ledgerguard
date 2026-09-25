# PRD: Revenue Forecasting (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/domain/service/forecasting_engine.go`, `handler/forecast_handler.go`,
> `frontend-flutter/lib/screens/analytics/forecasting_tab.dart`, `providers/analytics_provider.dart`.

## Problem & evidence

A partner watching current MRR wants to know **where it's heading** — "what will my recurring revenue be next month / in 12 months?" — to plan hiring, spend, and fundraising.

- **FACT:** the capability is fully built AND wired end-to-end (unlike AI Chat), so it answers a real, prioritized need.
- **UNKNOWN — original demand evidence not recorded.** No tickets/usage metrics stored. Open question: validate demand — owner: Product.

## Target users

**INFERRED:** the same **Shopify app partner** persona (founder/finance/growth) who needs a forward-looking number for planning. Per-app, per-org.
- **Not the target:** merchants; anyone needing investor-grade financial projections (the model is intentionally simple — see Non-goals).

## Proposed solution (as-built behavior)

**FACT.** Under **Analytics → Forecasting** (`forecasting_tab.dart`), a partner sees:
- An **empty state** ("Forecasting not yet available") until **≥90 days** of daily snapshots exist (`forecasting_tab.dart:28-34`; backend returns 422 below the threshold, `forecast_handler.go`).
- Two summary cards — **Next-Month Expected MRR** and the **12-month projection** — each showing a **Pessimistic – Expected – Optimistic** band.
- A **12-month line chart**: expected line plus shaded ±band overlays.
- A **model selector** — **Linear** vs **Exponential** — and a **data-points-used** badge.
- **MRR only.** No churn forecast, no usage/one-time projection.

**FACT (method — be honest).** Two deterministic algorithms in `forecasting_engine.go`:
- **Linear:** ordinary least-squares regression through daily `ActiveMRRCents` (`y = a + b·x`), projected forward.
- **Exponential:** Holt's double-exponential smoothing (level + trend); `alpha` user-tunable (default 0.3, clamped 0.1–0.9), `beta = alpha × 0.5` (hardcoded).
- Both project in **30-day "months"** and wrap the point estimate in a **hardcoded ±15% band** (optimistic ×1.15, pessimistic ×0.85, floored at 0). The band is a fixed heuristic, **not** a statistical confidence interval.

### Key screens

**No new wireframe** (as-built). Real surface: `frontend-flutter/lib/screens/analytics/forecasting_tab.dart` (summary cards + 12-month chart + model toggle). Wired via `analytics_provider.dart` → `metrics_service.dart:fetchForecast()` → `GET /api/v1/apps/{appID}/forecast`.

## Platform & policy constraints

**FACT/INFERRED.** Forecast is computed purely from stored `daily_metrics_snapshot` history — no external calls at forecast time. The **90-day minimum** is a hard gate (new/low-history apps get an empty state). Quality is bounded by snapshot coverage (gaps are treated as real data — see Open questions). No Shopify-policy interaction.

## Pricing-tier impact

**UNKNOWN — no plan-gating found.** Forecast is available to any authenticated org member. Open question: is forecasting a premium/planning-tier feature? — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Additive Analytics subtab reading existing snapshots. No migration; apps simply cross the 90-day threshold over time.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Accuracy:** median absolute % error of the **Expected** next-month MRR vs actual, measured 30 days later, ≤ 10% across apps with ≥180 days history.
2. **Calibration:** actual next-month MRR falls within the Pessimistic–Optimistic band ≥ 80% of the time (if it's far off, the ±15% heuristic needs replacing with real intervals).
3. **Engagement:** ≥ 25% of eligible (≥90-day) apps open the Forecasting tab monthly.

## Non-goals (deliberately absent in code)

1. **No churn forecast** — MRR only; at-risk-dollar projection is not modeled. **FACT.**
2. **No usage/one-time revenue projection** — only `ActiveMRRCents` (recurring). **FACT.**
3. **No statistical confidence intervals** — bands are a fixed ±15% heuristic. **FACT.**
4. **No seasonality / calendar-aware months** — assumes 30-day months. **FACT.**

## Open questions

- Validate demand (no evidence recorded). Owner: Product.
- **Are the ±15% bands acceptable, or should they become real prediction intervals?** (drives metric #2). Owner: Eng/Product.
- Snapshot **gaps vs zeros** — a missing day is treated as data; should the model interpolate/skip? Owner: Eng.
- Extend to **churn / usage / revenue-mix** forecasts? (currently MRR-only). Owner: Product.
- Pricing/tier gating. Owner: Product.
- `beta = alpha×0.5` and default `alpha=0.3` are unvalidated magic numbers — tune or document rationale. Owner: Eng.

# Tech Design: Dashboard & KPIs (as-built)

**PRD:** docs/prds/04-dashboard-kpis.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API contracts (`/metrics`, `/metrics/trend`, `/metrics/latest`, `/dashboard` prefs) are the hard-to-undo surfaces the Flutter client binds to. No new data model (reads existing `daily_metrics_snapshot` + transactions; layout prefs already persisted). KPI math is swappable behind the contract.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT.** Dashboard reads three backend surfaces (handler `metrics.go`, routes `router.go:193,212-214`; prefs `router.go:138-139`):
- **`GET /{appID}/metrics?start&end`** → `MetricsAggregationService.GetPeriodMetrics` → `{period, current, previous, delta}` (the real KPI + delta source).
- **`GET /{appID}/metrics/trend?months=1..12&granularity`** → downsampled snapshot series (503 if `trendProvider` nil, 500 on error; `metrics.go:118-179`).
- **`GET /{appID}/metrics/latest`** → **still hardcoded sample data** (`metrics.go:82-95`, `TODO: Calculate real metrics`) while also firing a `dashboard_viewed` analytics event (`metrics.go:76-80`).

**FACT (aggregation).** `GetPeriodMetrics` (`metrics_aggregation_service.go:36-81`) fetches current + `dateRange.PreviousPeriod()` snapshots AND current + previous transactions (`FindByAppID(..., End.Add(24h))` to include the end day), then `aggregateSnapshots` (`:86-129`): **point-in-time KPIs** (Active MRR, Revenue at Risk, Renewal Rate, risk counts) taken from the **latest snapshot in range**; **period revenue** (usage/total) computed from transactions via `MetricsEngine.CalculateUsageRevenue/CalculateTotalRevenue`. `NewPeriodMetrics` bundles current+previous; **a period with no snapshots yields a nil summary** (`:70-78`).

**FACT (delta).** The handler builds `delta` as `(current − previous)/previous × 100`, **inverted for risk/churn** so a decrease shows as positive (`metrics.go` delta builder, ~`:535-543`).

**FACT (frontend, live-wired).** `dashboard_screen.dart` renders 4 customizable primary KPI cards + optional widgets (`dashboard_registry.dart:82-95`); `dashboard_provider.dart` computes date-range windows with `periodRangeFor(chip, now)` (DST-safe, tested), `setTimeRange` re-fetches `/metrics` only (trend/forecast are period-independent, fetched in parallel on load, each degrading independently). Layout persisted per-user via UserPreferences (`/dashboard` GET/PUT). `DemoModeCoordinator` swaps all providers to mocks.

## Proposed design

**No redesign — documents the shipped design.** Pattern: **snapshot-backed aggregation.** The daily sync writes `daily_metrics_snapshot`; the dashboard reads end-of-period snapshots for point-in-time KPIs and sums transactions for period revenue, comparing against the prior equal-length window for deltas. Happy path: see `docs/designs/04-dashboard-kpis-sequence.puml` (validated `plantuml -checkonly`). Component interaction spans handler → aggregation service → engine → repos (diagrammed).

### Data model

**FACT — no new tables.** Reads `daily_metrics_snapshot` (produced by the daily `MetricsEngine` from subscriptions+transactions) and `transactions`. Dashboard **layout preferences** persist in the existing UserPreferences store (per-user, via `/dashboard`). No migration.

### API & events

**FACT.** `/metrics` current/previous KPI fields (`metrics.go:198-208`): `active_mrr_cents, revenue_at_risk_cents, usage_revenue_cents, total_revenue_cents, renewal_success_rate, safe_count, one_cycle_missed_count, two_cycles_missed_count, churned_count`. `delta` fields (`:210-217`): `*_percent` per KPI + `churn_count_percent`. Trend: `{granularity, data_points, snapshots[]}`. Analytics event `dashboard_viewed` emitted on `/metrics/latest`.
**Breaking-change check:** `*_cents`/`*_percent` field names are the contract; the Flutter models bind directly.

## Alternatives considered

- **Snapshot-backed KPIs vs live recompute per request.** As-built reads pre-computed daily snapshots (cheap, deterministic, historically consistent) and only sums transactions live for period revenue. Full live recompute was **not recorded as considered**. **Choose live recompute if** snapshot lag makes "today's" number feel stale (see D3 freshness).
- **Delta baseline = prior equal-length window** (`dateRange.PreviousPeriod()`). A fixed calendar baseline (e.g. same period last month/year) was not recorded as considered. **Choose calendar baseline if** users read "This Week Δ" as week-over-week rather than vs the immediately preceding 7 days.
- **Two endpoints (`/metrics` real + `/metrics/latest` mock)** — not a deliberate alternative; the mock is unfinished work (D1). No rationale recorded.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Unauthenticated | `UserFromContext==nil` | **FACT** 401 (`metrics.go:63-67,119-123`) | re-auth |
| App not found / bad appID | `resolveAppFromRequest` | **FACT** 404/400 | pick valid app |
| Trend provider unconfigured | `trendProvider==nil` | **FACT** 503 (`:131-134`) | ops config |
| Trend/snapshot repo error | repo err | **FACT** 500 (`:151-154`) | retry |
| No snapshots in selected period | `len==0` | **FACT** nil summary; delta can't compute → **degraded** card (`agg:70-78`) | widen range / wait for sync |
| Trend/forecast fetch fails | provider catch | **FACT** only that widget degrades, dashboard survives (`dashboard_provider.dart:141-155`) | reload |

### DIVERGENCE — failure paths / correctness gaps NOT handled (LISTED, not fixed)

- **D1. `/metrics/latest` returns hardcoded sample data** (`metrics.go:82-95`) yet is a live route and fires a real `dashboard_viewed` event — any consumer of it shows fake KPIs. → open question: what still calls it? proposed test/removal.
- **D2. Systemic tenant isolation.** Same `resolveAppFromRequest` (no org-ownership check) as reports/forecast/chat. → cross-ref `02-ai-chat.md` D1 / `03-forecasting.md` D3.
- **D3. No freshness signal.** KPIs reflect the last snapshot but the UI shows no "as of <date>"; a stalled sync silently serves stale numbers as current. → proposed test + PRD metric #2.
- **D4. Snapshot gaps skew point-in-time KPIs.** "Latest in range" is whatever the newest present snapshot is; if the true end-of-period day is missing, a stale mid-period value is presented as end-of-period. → proposed test.
- **D5. Division-by-zero / new-app deltas.** When previous == 0 or previous summary is nil (new app), delta % is undefined; behavior (0? hidden? ∞?) needs a defined contract. → proposed test.
- **D6. Cross-surface drift.** Dashboard Active MRR (snapshot) vs MRR report headline (also snapshot but computed in a different handler) can diverge if logic drifts. → proposed consistency test (PRD metric #1).

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Cross-surface consistency: `/metrics` Active MRR == MRR report headline for the same period (metric #1). **TODO.**
2. Freshness: response should carry / UI should show snapshot date ≤24h (metric #2). **TODO — no freshness field today.**

**Adversarial (one per Failure-modes row — partly EXISTING):**
- Period aggregation with mock snapshot/tx repos — **exists** (`metrics_aggregation_service_test.go`). DST-safe range math — **exists** (`dashboard_period_range_test.dart`).
- Missing today: 401/404/503/500 HTTP mapping per endpoint; no-snapshots nil-summary path.

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 `/metrics/latest` mock (assert-or-remove) · D2 cross-org appID · D4 gapped end-of-period · D5 previous==0 / new-app delta · D6 dashboard-vs-report MRR equality. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/application/service/ -run Metrics -v`
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/metrics?start=2026-08-01&end=2026-08-31" | jq '.current,.delta'`
3. Compare `.current.active_mrr_cents` to `/reports/mrr` headline for the same window (D6).
4. Hit `/metrics/latest` and confirm whether it still returns the 125000/15000 sample constants (D1).
5. `cd frontend-flutter && flutter test test/**/dashboard_period_range_test.dart`

**External-platform reality:** no external call at read time — KPIs are local over snapshots/transactions; fully deterministic/unit-testable. Freshness depends on the upstream sync pipeline (faked below the repo in tests).

## Pricing & policy touchpoints

**UNKNOWN — no plan-gating in code.** No Shopify-policy interaction. → Open question (Product).

## Rollout

**N/A (as-built, actively wired).** Recent commits (#83 real deltas/trend/forecast, #84 date-range re-fetch) already migrated mock→real; residual = D1 `/metrics/latest`. Layout prefs additive. Rollback = revert handler/provider; per-user prefs unaffected (no schema change).

## Observability

**FACT (partial):** `dashboard_viewed` analytics event on `/metrics/latest`; errors via `log`. **DIVERGENCE:** no metric on `/metrics` latency, no snapshot-age/freshness gauge, no alert when a period returns nil summary. On-call greps Hetzner container logs. → proposed observability TODO tied to PRD freshness metric.

## Open questions

- **Remove or wire `/metrics/latest`** (D1) — what consumes it? Owner: Eng.
- Delta baseline semantics per chip (D5 + prior-window vs calendar). Owner: Product.
- Surface snapshot freshness "as of <date>" (D3). Owner: Eng/Product.
- **Systemic org-ownership check (D2).** Owner: Eng/Sec.
- Dashboard↔report MRR consistency guarantee (D6). Owner: Eng.
- Pricing/tier gating. Owner: Product.

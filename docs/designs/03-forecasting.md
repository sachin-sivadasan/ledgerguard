# Tech Design: Revenue Forecasting (as-built)

**PRD:** docs/prds/03-forecasting.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API contract (`GET /api/v1/apps/{appID}/forecast` response shape) is the hard-to-undo surface — the Flutter client parses it. No new data model. The forecasting **math** is swappable behind that contract, so algorithm changes are low-cost; contract changes are not.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT.** One endpoint: `GET /{appID}/forecast` (`router.go:263`), handled by `ForecastHandler.GetForecast` (`forecast_handler.go:37`). It: checks auth (401, `:38-42`) → `resolveAppFromRequest` (404/400/503, `:44-48`) → parses `months` (1–36, default 12; invalid silently → 12, `:52-55`), `model` (`linear`|`exponential`, else 400, `:58-65`), `alpha` (default 0.3, `:67-72`) → fetches **only the last 12 months** of snapshots (`from = now.AddDate(-1,0,0)`, `:75-77`; repo error → **500**, `:79`) → dispatches to the engine → serializes `{model, generated_at, data_points_used, forecast[]}` (`:137-145`).

**FACT.** `ForecastingEngine` (`forecasting_engine.go`) is a stateless struct with two deterministic methods, both gated at **`MinDataPointsForForecast = 90`** (`:14,38,106`) returning `ErrInsufficientData` → 422 below threshold:
- **Linear** (`:33-96`): OLS fit `y=a+b·x` over daily `ActiveMRRCents` with day-**index** as x (`:45-61`); projects `futureDays = lastIndex + m*30` (`:70`), `expected = a + b·futureDays`.
- **Exponential** (`:100-166`): Holt's double-exponential smoothing; `alpha` clamped 0.1–0.9 (`:111-116`), `beta = alpha*0.5` (`:119`), trend seeded from day-0→day-30 slope (`:124-126`), projects `level + trend*(m*30)` (`:141`).
- Both wrap the point estimate in **hardcoded ±15% bands** (`optimistic = expected*1.15`, `pessimistic = expected*0.85`, floored at 0; `:76-80, 146-150`) and set `futureDate = lastDate.AddDate(0,m,0)` (calendar month, `:69,139`).

**FACT (frontend — live-wired).** `analytics/forecasting_tab.dart` → `AnalyticsProvider` → `MetricsService.fetchForecast()` → the endpoint; parsed via `ForecastResult.fromJson`. Empty state below 90 points; model toggle re-fetches; mock used **only** in demo mode (`analytics_provider.dart`, `mock_analytics.dart`).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **thin handler (param parse + app lookup + snapshot fetch) → pure stateless engine (two algorithms) → JSON.** Deterministic, no persistence, recomputed per request. Happy path: see `docs/designs/03-forecasting-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — no new tables.** Reads `daily_metrics_snapshot.ActiveMRRCents` only. Forecast output is **not persisted** (recomputed each call). No migration.

### API & events

**FACT.** `GET /api/v1/apps/{appID}/forecast?months=1..36&model=linear|exponential&alpha=0.3`. 200 body:
```json
{ "model":"linear", "generated_at":"<RFC3339>", "data_points_used":90,
  "forecast":[ {"date":"2026-10-25","expected_cents":580000,
                "optimistic_cents":667000,"pessimistic_cents":493000} ] }
```
422 body: `{"error":"insufficient data for forecasting","data_points":<n>,"required":90}`. No events.
**Breaking-change check:** field names (`*_cents`, `data_points_used`) are the contract; renames break the Flutter model.

## Alternatives considered

- **Linear (OLS) vs Exponential (Holt's).** Both are shipped and user-selectable — the code treats them as peers, not one-rejected. **Choose exponential when** recent trend should dominate (it weights latest data); **linear when** the whole 12-month history should count equally. No rationale recorded for offering both vs picking one.
- **±15% heuristic bands vs statistical prediction intervals.** As-built uses a fixed ±15% (`:76-77`). A residual-based interval (std error of regression) was **not recorded as considered**. **Choose real intervals if** PRD calibration metric (#2) fails. Honest entry: no alternative rationale in code/history.
- **Index-based x vs date-based x.** Regression uses array index as x, assuming contiguous daily snapshots. A date-based x was not recorded as considered. **Choose date-based if** snapshot gaps prove common (see D2).

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Unauthenticated | `UserFromContext==nil` | **FACT** 401 (`:38-42`) | re-auth |
| Invalid `model` | not in {linear,exponential} | **FACT** 400 (`:62-65`) | client fixes |
| App not found / bad appID | `resolveAppFromRequest` | **FACT** 404/400 (`:44-48`) | pick valid app |
| Snapshot repo error | repo returns err | **FACT** 500 (`:79`) — note reports use 503 (inconsistency) | retry |
| < 90 snapshots | engine `ErrInsufficientData` | **FACT** 422 + counts (`:83-91`) | wait for history |
| Projected value negative | `expected<0` | **FACT** floored to 0 (`:72-74,142-144`) | — |
| `denominator==0` (all-same x) | guarded | **FACT** set to 1 (`:56-59`) — "shouldn't happen with daily data" | — |

### DIVERGENCE — failure paths / correctness gaps NOT handled (LISTED, not fixed)

- **D1. Bands are not statistical.** ±15% is fixed regardless of variance/fit quality — a noisy series and a clean one get identical band width. → proposed calibration test (PRD metric #2).
- **D2. Snapshot gaps distort the model.** x = array **index**, not calendar day (`:46`), while output dates are calendar months (`:69`). Missing days compress the x-axis → slope/step mismatch, and the displayed `date` won't align with the x used to compute it. → proposed test: forecast with gapped snapshots.
- **D3. Systemic tenant isolation.** `resolveAppFromRequest` (`app_lookup.go:54`) resolves the app by `appRepo.FindByID(appID)` with **no org-ownership check** — identical to the AI-chat D1 concern but shared by ALL app-scoped endpoints (reports + forecast). Whether repos scope by org is UNVERIFIED. → cross-ref `docs/designs/02-ai-chat.md` D1; systemic open question.
- **D4. Only 12 months of history used** (`:75-76`) even if more exists; and **no caching** — every request recomputes OLS/Holt's over all snapshots. → proposed load/behavior tests.
- **D5. Magic numbers unvalidated:** `alpha=0.3` default, `beta=alpha*0.5`, 30-day months, day-0→day-30 trend seed. No sensitivity analysis. → open question.
- **D6. Silent param coercion:** `months` outside 1–36 and unparseable `alpha` are silently defaulted, not rejected — a client bug looks like success. → proposed test.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Accuracy/calibration harness: backtest Expected vs actual next-month MRR on ≥180-day fixtures; assert median abs % error ≤10% and band-coverage ≥80%. **TODO — not present.**

**Adversarial (one per Failure-modes row — partly EXISTING):**
- 422 insufficient data, linear happy path, invalid model → 400 — **exist** (`forecast_handler_test.go`). Engine trend cases (growing/flat/declining) + alpha clamp — **exist** (`forecasting_engine_test.go`).
- Missing today: exponential-via-HTTP, 500 repo-error path, negative-MRR flooring via HTTP.

**Adversarial — DIVERGENCE TODO (none exist today):**
- D2 gapped snapshots · D3 cross-org appID · D4 caching/12-mo window · D6 out-of-range `months`/bad `alpha`. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/domain/service/ -run Forecast -v && go test ./internal/interfaces/http/handler/ -run Forecast -v`
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/forecast?model=exponential&alpha=0.5&months=6" | jq .`
3. Repeat with an app < 90 days → expect 422 with `data_points`/`required`.
4. `cd frontend-flutter && flutter test` for analytics/forecast widget tests.

**External-platform reality:** none — forecast is pure local computation over snapshots; no Shopify/LLM dependency. Fully deterministic and unit-testable without any sandbox.

## Pricing & policy touchpoints

**UNKNOWN — no plan-gating in code.** No Shopify-policy interaction. → Open question (Product).

## Rollout

**N/A (as-built / shipped, live-wired).** Additive Analytics subtab; apps become eligible as they cross 90 days. Algorithm changes ship behind the stable JSON contract; rollback = revert handler/engine, no data stranded (nothing persisted).

## Observability

**DIVERGENCE:** no metrics on forecast requests, model mix, or eligibility (how many apps are <90 days). Errors only via `log`. On-call greps Hetzner container logs (no forecast index). → proposed observability TODO tied to PRD engagement metric #3.

## Open questions

- Replace ±15% bands with residual-based intervals? (D1, PRD metric #2). Owner: Eng/Product.
- Fix index-vs-date x-axis on gaps (D2). Owner: Eng.
- **Systemic org-ownership check on `resolveAppFromRequest` (D3)** — verify/enforce. Owner: Eng/Sec.
- Validate/justify magic numbers (D5). Owner: Eng.
- Extend to churn/usage forecasts (PRD non-goal today). Owner: Product.
- Pricing/tier gating. Owner: Product.

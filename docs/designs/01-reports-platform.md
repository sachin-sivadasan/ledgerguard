# Tech Design: Reports Platform (as-built)

**PRD:** docs/prds/01-reports-platform.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API surface (`/api/v1/apps/{appID}/reports/*`) — external clients + the Flutter app depend on the JSON contracts; changing them is a breaking change. No new data model is introduced by this capability (it reads existing tables), so the irreversible risk is **API contract**, not schema.

> **As-built.** "Current state IS the design." Every claim cites a real path. Markers: **FACT** (code proves it), **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded). **Nothing here is refactored or fixed** — gaps are LISTED as DIVERGENCE, not repaired.

## Current state

**FACT.** Reports is a set of ~22 independent read-only HTTP handlers, one per report, wired in `internal/interfaces/http/router/router.go:268-405` under `GET /{appID}/reports/{report}`. Each lives in `internal/interfaces/http/handler/<name>_report_handler.go` with a sibling `_test.go`. A representative handler (`mrr_report_handler.go:73-116`) shows the shared shape: (1) `middleware.UserFromContext` auth check → (2) `resolveAppFromRequest` (`app_lookup.go:54`) to resolve the app from the URL `appID` + org/user partner context → (3) `parseDateRange` (`revenue_at_risk_handler.go:169`) → (4) fetch **full** data via repos (`subRepo.FindByAppID`, `snapshotRepo.FindByAppIDRange`) → (5) aggregate via pure helper funcs and/or a domain service → (6) branch to CSV (`?format=csv`) or JSON.

**FACT.** Aggregation reuses domain services in `internal/domain/service/`: `metrics_engine`, `earnings_calculator`, `fee_verification_service` (`VerifyTransaction`, `CalculateFeeSummary`, `CalculateTierSavings`), `risk_engine`, `ledger_service`, `forecasting_engine`. Trend series are computed over `daily_metrics_snapshot` rows, then downsampled by `resolveTrendInterval`/`downsampleSnapshots` (`trend_interval.go:22,52`) to day/week/month.

**FACT.** Paging is shared (`report_paging.go`): `parsePaging` reads `?limit`/`?offset`, clamps `limit` to **200**, treats `limit<=0` as "all rows" (CSV/legacy); `pageSlice` returns a bounds-safe window and never `null`. **KPIs and trends are always computed over the full dataset — paging only slices the list table** (`report_paging.go:18-20`).

**FACT.** Frontend (`frontend-flutter/lib/screens/reports/`, Provider) layers **service → provider → screen**: e.g. `services/mrr_report_service.dart` (Dio, path `/api/v1/apps/$appId/reports/mrr`) → `providers/mrr_report_provider.dart` (`ChangeNotifier`, exposes `error` + `isServiceUnavailable`, maps `DioException` → "Service temporarily unavailable.") → `screens/reports/mrr_report_screen.dart`. Report + detail screens are paired (e.g. `revenue_at_risk_screen.dart` → `revenue_at_risk_stores_screen.dart`).

## Proposed design

**No redesign — this documents the shipped design.** The pattern is a **per-report handler + shared cross-cutting helpers** (`app_lookup`, `parseDateRange`, `parsePaging`, `trend_interval`), each handler delegating math to a domain service, over the pre-computed `daily_metrics_snapshot` plus live `subscriptions`/`transactions`. Happy path:

![sequence](01-reports-platform-sequence.puml) — see `docs/designs/01-reports-platform-sequence.puml` (validated with `plantuml -checkonly`).

### Data model

**FACT — no new tables for this capability.** Reports is read-only over existing stores (per PRD "packaging play"): `daily_metrics_snapshot` (pre-aggregated MRR, risk counts, renewal rate — the trend/KPI source), `subscriptions` (risk_state, base price, `StartDate()` business date, plan), `transactions` (charge_type, gross/fee/net — for revenue-mix/usage/fee-audit/reconciliation). **INFERRED:** business dates come from `subscription.StartDate()` / `churnedDateOf`, deliberately **not** `CreatedAt` which resets on every ledger rebuild (`mrr_report_handler.go:164-166`, aligns with memory [[createdat-is-record-date]]). No migration — nothing to backfill.

### API & events

**FACT.** Uniform contract: `GET /api/v1/apps/{appID}/reports/{report}` with `?from=YYYY-MM-DD&to=YYYY-MM-DD` (defaulted by `parseDateRange`), optional `?segment=`, `?format=csv`, and `?limit`/`?offset` on list-bearing reports. Auth = Firebase ID token + `X-Org-Id` org context (router middleware). ~22 report slugs registered at `router.go:268-405` (mrr, revenue-at-risk, churn, retention, active-customers, earnings, revenue-mix, usage, usage-trends, subscriptions, payout-schedule, payout-history, uninstall-context, installs, activation, fee-audit, customer-insights, ledger-reconciliation, net-new-subscriptions, cohorts, reviews).
**Breaking-change check:** JSON field names are hand-authored per report (e.g. `mrrReport` struct, `mrr_report_handler.go:59-69`); renaming/removing a field breaks the Flutter model + any API-key client. No events emitted (reports are pull-only). **UNKNOWN:** whether a versioning policy exists for these contracts → Open question.

## Alternatives considered

- **One generic "report engine" endpoint (`/reports?type=`) with a report registry**, vs the as-built one-handler-per-report. *Not recorded as considered.* Rejected-in-practice by the code, which duplicates the fetch/paging/CSV skeleton per handler. **Choose the generic engine instead if** the count keeps growing and skeleton drift/duplication becomes the maintenance cost. (No alternative rationale recorded in code/history.)
- **Compute on the fly vs read from `daily_metrics_snapshot`.** The code does BOTH: trends/headline KPIs read snapshots (cheap, deterministic), while period movements (new/churned MRR) are computed live from subs. **FACT** this split is intentional (`mrr_report_handler.go:101-105` vs `161-181`). **Choose full live-compute instead if** snapshot freshness/lag becomes a correctness problem.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Unauthenticated request | `UserFromContext == nil` | **FACT** 401 "authentication required" (`mrr_report_handler.go:75`) | client re-auths |
| Malformed / missing `appID` | `uuid.Parse` fails / empty | **FACT** 400 (`app_lookup.go:56-63`) | client fixes path |
| App / partner not found | `isNotFoundError` sentinel | **FACT** 404 (`app_lookup.go:67-74`) | client picks valid app |
| Any repo/DB error mid-read | repo returns non-nil err | **FACT** 503 "service temporarily unavailable", logged (ADR-042 — no not-found sentinel, every err = infra) (`mrr_report_handler.go:120-123`) | frontend shows `isServiceUnavailable` banner (`mrr_report_provider.dart:79-80`) |
| Oversized page request (whale store) | `limit > 200` | **FACT** clamped to 200 (`report_paging.go:27-29`) | bounded payload |
| < 2 snapshots in range | `len<2` guard | **FACT** MoM delta returns 0, frontend hides it (`mrr_report_handler.go:132-135`) | — |

### DIVERGENCE — failure paths the code does NOT handle (LISTED, not fixed)

- **D1. Upstream Shopify Partner API rate-limiting / errors are invisible to reports** — reports read local DB only; stale sync = silently stale numbers, no freshness signal in the response. → proposed test / open question.
- **D2. CSV export ignores `limit` and streams ALL rows** (`report_paging.go:14-15`) — a whale store's CSV is unbounded (the 200 clamp is JSON-only). → proposed load test.
- **D3. No request timeout / cancellation surfaced** — a very wide date range loads all subs + snapshots into memory and aggregates synchronously; no pagination on the *fetch*, only on the response slice. → proposed test for large-tenant latency.
- **D4. Partial/empty snapshot coverage** (sparse days) is treated as real zeros in trends — no "no data" vs "zero" distinction. → open question (Product): is a gap a 0 or a null?
- **D5. Contract drift** — no automated check that Go JSON structs match the Flutter models; a field rename breaks the app silently at runtime. → proposed contract test.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Determinism: same underlying data → identical KPIs/trend on re-run (guards the idempotency FACT). Covered today by per-handler `_test.go` golden-style assertions.
2. Reconciliation surfaces discrepancies: Fee Audit / Ledger Reconciliation flag a known-bad fixture (`fee_audit_report_handler_test.go`, `ledger_reconciliation_report_handler_test.go`).

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- 401 / 400 / 404 / 503 mapping — present in handler `_test.go` suites. Confirm coverage per new report.
- `limit>200` clamp + `offset` past end → `[]` — `report_paging_test.go`.
- `<2 snapshots` → MoM 0 — `mrr_report_handler_test.go`.

**Adversarial — DIVERGENCE TODO (not yet covered):**
- D2 unbounded CSV · D3 wide-range latency/memory · D5 Go↔Flutter contract test · D4 gap-vs-zero semantics. **TODO — none exist today.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/interfaces/http/handler/... -run Report -v`
2. Run server; `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/reports/mrr?from=2026-08-01&to=2026-08-31" | jq .`
3. Append `&format=csv` → expect `text/csv`; append `&limit=8` → expect ≤8 plan rows but unchanged headline KPIs.
4. `cd frontend-flutter && flutter test` for report widget/provider tests.

**External-platform reality:** reports never call Shopify at read time — the Partner API is faked below the sync layer, not the report layer, so report tests need no Shopify sandbox. Freshness (D1) can only be validated end-to-end after a real sync run.

## Pricing & policy touchpoints

**UNKNOWN — no plan-gating in the report handlers** (mirrors PRD open question). Shopify policy is upstream at sync time, not at report read. → Open question (Product): gate Guard/advanced reports?

## Rollout

**N/A (as-built / already shipped).** No migration, no flag observed. **INFERRED** future rollout for *new* reports = additive: register route + handler + Flutter service/provider/screen; no schema change, so rollback = revert the handler, no data stranded.

## Observability

**FACT (partial):** handlers `log.Printf` repo errors with an operation tag (e.g. `"mrr: repo error in %s"`, `mrr_report_handler.go:121`); frontend distinguishes 503 via `isServiceUnavailable`. **DIVERGENCE:** no per-report latency/volume metric, no dashboard tying report usage to the PRD adoption metric. On-call today greps container logs on the Hetzner box (no report-specific index). → proposed observability TODO.

## Open questions

- API contract versioning for `/reports/*` — none found. Owner: Eng.
- D4 gap-vs-zero semantics in trends. Owner: Product/Eng.
- Plan/pricing gating (esp. Guard). Owner: Product.
- Generic report-engine refactor if handler count keeps growing (currently ~22 near-duplicate skeletons). Owner: Eng.
- Cross-app aggregate variant coverage across all reports (PRD open question). Owner: Eng.

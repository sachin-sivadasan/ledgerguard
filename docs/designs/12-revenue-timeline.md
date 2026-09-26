# Tech Design: Earnings & Revenue Timeline (as-built)

**PRD:** docs/prds/12-revenue-timeline.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API shapes (earnings/periods, revenue/concentration, earnings/status, daily earnings) bind the Flutter client. No new data model. Low irreversibility; the main risk is **metric coherence** — multiple net-revenue paths (this + the earnings report + payouts, #7) must not drift.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE. Deep payout/reconciliation design is in `docs/designs/07-fees-and-payouts.md` (referenced, not repeated).

## Current state

**FACT — endpoints** (`revenue_handler.go`, `revenue_metrics_service.go`, routes `router.go:226-231`):
- `GET /apps/{appID}/earnings?start&end&mode=combined|split` (`:66`) — daily totals, optional subscription/usage split; `start`/`end` required (400 on bad/missing). **Live endpoint, no frontend route.**
- `GET /apps/{appID}/earnings/periods?start&end` (`:130-177`) — monthly cards; `GetMonthlyEarnings` (400 on invalid range) returns gross/net + per-status counts; the handler derives `shopify_cut = gross − net` and a month status. **Live-wired** (`earnings_screen.dart`).
- `GET /apps/{appID}/revenue/concentration?top=N&start&end` (`:179-229`) — top-N stores by net over 90d default. **Live-wired** (`analytics/revenue_tab.dart`).
- `GET /apps/{appID}/earnings/status` (`:34-62`) — pending/available/paid totals + pending-by-available-date + 30-day upcoming. **Live endpoint, no screen.**

**FACT — concentration logic** (`revenue_metrics_service.go:284-324`): iterate transactions, **skip `NetAmountCents <= 0`** (`:288`), bucket by `MyshopifyDomain` (empty → `"unknown"`, `:292-293`), sum net, `pct = revenue/total*100` (`:310`), then an **O(n²) bubble sort** descending (`:322-324`).

**FACT — earnings math** shared with #7: net = gross − shopify_share − processing − tax; availability = charge+7d; status by available-date (`earnings_calculator.go`). Net-by-status + per-charge table are the earnings report (`earnings_report_handler.go`, PRD #7).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **thin handlers over aggregate reads (monthly/daily) + a raw-transaction aggregation (concentration)**, all deterministic, read-only. Read paths: see `docs/designs/12-revenue-timeline-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — no new tables.** Monthly/daily views read revenue aggregates (`RevenueRepository`); concentration + status read raw `transactions` (`TransactionRepository`). No migration.

### API & events

**FACT.** Response shapes: `MonthlyEarningsResponse{Earnings:[{Month,GrossCents,ShopifyCutCents,NetEarningsCents,Status}]}`; `RevenueConcentrationResponse{TotalRevenueCents,Stores:[{Domain,ShopName,RevenueCents,TransactionCount,PctOfTotal}]}`; `EarningsStatusResponse{TotalPending/Available/PaidOutCents,PendingByDate[],UpcomingAvailability[]}`; `EarningsTimelineResponse{Earnings:[{Date,TotalAmountCents,(Subscription/UsageAmountCents if split)}]}`. No events.
**Breaking-change check:** `*_cents` fields bind the earnings + analytics providers.

## Alternatives considered

- **Raw-transaction concentration vs a pre-aggregated read model.** As-built aggregates raw transactions per request. **Choose a read model if** whale apps make per-request aggregation slow (compounded by the O(n²) sort). No alternative recorded.
- **Aggregate-backed monthly cards vs recompute from transactions.** As-built reads `RevenueRepository` aggregates (cheap). **Choose recompute if** aggregate freshness lags. Consistent with the dashboard's snapshot approach (#4).
- **Skip-negative-net in concentration vs include refunds.** Code skips net≤0 (`:288`) — a deliberate "top revenue stores" framing, but it hides refund-heavy stores. **Choose include if** the view should show *net* contribution including refunders.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Missing/bad date on `/earnings` | parse | **FACT** 400 (`revenue_handler.go:155-162`) | fix params |
| Invalid range on `/earnings/periods` | `ErrInvalidDateRange` | **FACT** 400 (`:166-168`) | fix range |
| Repo error | err | **FACT** 500 (`:170-172`) | retry |
| Empty domain on a tx | `== ""` | **FACT** bucketed as `"unknown"` (`:292-293`) | fix sync data |
| `total == 0` | guard | **FACT** pct = 0, no div-by-zero (`:309`) | — |
| All-refund store | net≤0 skip | **FACT** store omitted from concentration (`:288`) | — (see D1) |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. Concentration hides refund-heavy stores** — `net<=0` transactions are skipped, so a store that's net-negative never appears; "top stores" silently excludes churny/refunding customers. → open question (include net contribution?).
- **D2. O(n²) bubble sort** for store ranking (`:322-324`) — fine for small store counts, degrades for whale apps with many domains. → proposed: sort.Slice + read model if needed.
- **D3. Two live endpoints unwired** — daily `/earnings` and `/earnings/status` return data no screen consumes. Dead surface or pending UI. → wire or remove.
- **D4. Single-currency + no per-app split** — all views assume one currency; endpoints are single-app (no portfolio roll-up). → open question.
- **D5. Untested** — `GetRevenueConcentration` and `GetEarningsStatus` have no `_test.go` (GetEarnings/GetEarningPeriods are tested). → add tests.
- **D6. Metric-coherence risk** — net appears via monthly cards, daily timeline, and the earnings report (#7) through different code paths; no test asserts they agree. → cross-surface consistency test (metric #1).
- **D7. Systemic tenant isolation** — `resolveAppFromRequest`, no org-ownership check. → cross-ref `02-ai-chat.md` D1.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Cross-surface net consistency: monthly-card net == daily-timeline net == earnings-report net for a fixed period (metric #1). **TODO.**
2. Concentration self-consistency: Σ store pct (+ rounding) = 100% and store revenues sum to `TotalRevenueCents` (metric #2). **TODO.**

**Adversarial (one per Failure-modes row — partly EXISTING):**
- `/earnings` combined/split, auth, missing params, app-not-found, invalid appID — **exist** (`revenue_handler_test.go`). `/earnings/periods` status derivation, bad date, start>end, Shopify-cut calc — **exist**.

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 all-refund store omission · D2 large-store-count sort · D3 unwired-endpoint contracts · D5 concentration multi-store grouping + "unknown" bucket + negative-net skip · D6 net cross-surface equality. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/interfaces/http/handler/ -run Revenue -v`
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/earnings/periods" | jq '.Earnings[0]'` — confirm `ShopifyCutCents == GrossCents - NetEarningsCents`.
3. `curl ... /revenue/concentration?top=5 | jq '[.Stores[].PctOfTotal] | add'` — expect ≈ (100 − rest%).
4. Cross-check a month's net vs `/reports/earnings` for the same window (D6).

**External-platform reality:** all views are local over synced data — no Shopify/LLM call at read time; deterministic/unit-testable. Only upstream freshness depends on sync.

## Pricing & policy touchpoints

**UNKNOWN — no plan-gating.** No Shopify-policy interaction. → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Read-only; no migration. Rollback = revert. If concentration performance becomes an issue (D2), a read model can be added behind the stable contract.

## Observability

**FACT (partial):** handler errors logged. **DIVERGENCE:** no metric on view usage (would settle D3 wire-or-remove), concentration latency, or net cross-surface drift. On-call greps Hetzner logs. → proposed observability TODO tied to metrics #1/#3.

## Open questions

- **Wire or remove `/earnings` + `/earnings/status` (D3).** Owner: Product/Eng.
- **Concentration: include net-negative stores? (D1)** Owner: Eng/Product.
- **O(n²) sort → sort.Slice / read model for whale apps (D2).** Owner: Eng.
- **Add tests for concentration + status (D5).** Owner: Eng.
- **Net cross-surface consistency test (D6).** Owner: Eng.
- Multi-currency + pricing gating + systemic org-check (D4/D7). Owner: Eng/Product/Sec.

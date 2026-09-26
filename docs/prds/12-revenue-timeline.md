# PRD: Earnings & Revenue Timeline (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** / **INFERRED** / **UNKNOWN**.
> Success metrics are **PROPOSED** (future), not historical. **Full standalone** capability doc —
> self-contained. It covers the whole earnings/revenue *monitoring* surface; the deep
> **payout-schedule / payout-history / fee-audit / ledger-reconciliation** detail is specified in
> [PRD #7](07-fees-and-payouts.md) and referenced (not re-derived) where it connects here.
> Sources: `handler/{revenue_handler,earnings_report_handler}.go`,
> `application/service/revenue_metrics_service.go`, `domain/service/earnings_calculator.go`,
> `frontend-flutter/lib/screens/{earnings/earnings_screen,earnings/earnings_report_screen,analytics/revenue_tab}.dart`.

## Problem & evidence

A Shopify app partner needs an ongoing, at-a-glance answer to **"how much am I actually making, where is it coming from, and where is it heading?"** — net take-home over time, the Shopify cut made visible (gross vs net), which stores drive the revenue (concentration risk), and what's pending vs already paid out.

- **FACT:** a complete earnings/revenue monitoring surface exists across `revenue_handler.go` (timeline, monthly cards, concentration, status) and `earnings_report_handler.go` (net-by-status + per-charge table), live-wired to the earnings + analytics screens.
- **UNKNOWN — original demand evidence not recorded.** Open question: validate which of these views partners actually rely on — owner: Product.

## Target users

**INFERRED:** the **Shopify app partner** (founder/finance/growth) monitoring revenue health per app within an org.
- **Not the target:** merchants; users needing the reconciliation/"did Shopify pay me correctly?" audit (that's the Guard wedge, #7) or investor-grade projections (forecasting, #3).

## Proposed solution (as-built behavior)

**FACT — the earnings/revenue monitoring surface has five views:**

1. **Net earnings by status** (`earnings_report_handler.go`, screen `earnings_report_screen.dart`): net take-home split into **pending / available / paid-out**, plus a per-charge table (date, domain, gross, net, status, available-date). *(Full detail: PRD #7.)*
2. **Monthly earnings cards** (`GET /apps/{appID}/earnings/periods`, `revenue_handler.go:130-177`; screen `earnings_screen.dart`): per calendar month **gross / Shopify cut (= gross − net) / net** + a derived month status (pending → available → paid-out precedence). **Live-wired.**
3. **Daily earnings timeline** (`GET /apps/{appID}/earnings?start&end&mode=combined|split`, `:66`): per-day totals, optionally split into subscription vs usage. Endpoint **live**; no frontend route detected.
4. **Customer concentration** (`GET /apps/{appID}/revenue/concentration?top=N`, `:179-229`; `analytics/revenue_tab.dart`): top-N stores by **net revenue** over a 90-day default window with `pct_of_total` — surfaces revenue-concentration/churn risk. **Live-wired.**
5. **Earnings-status state machine** (`GET /apps/{appID}/earnings/status`, `:34-62`): total pending/available/paid + pending-bucketed-by-available-date + a 30-day upcoming-availability view. Endpoint **live but NOT wired to any screen.**

**FACT — computation.** Net = `gross − shopify_share − processing − tax_on_fees` (`transaction.go`); availability = charge date **+ 7 days** (refunds immediate); status PENDING/AVAILABLE/PAID_OUT by comparing now to available-date (`earnings_calculator.go`). Monthly cards/timeline read revenue aggregates; concentration reads raw transactions grouped by `MyshopifyDomain`.

### Key screens

**No new wireframe** (as-built). Real surfaces: `earnings_report_screen.dart` (net-by-status + charges), `earnings_screen.dart` (monthly cards), `analytics/revenue_tab.dart` ("Top Stores by Revenue"). The daily-timeline and earnings-status endpoints have no screen.

## Platform & policy constraints

**FACT.** All views derive from synced Shopify transaction fee fields; freshness is sync-bound. The **7-day payout delay** and a **30-day upcoming-availability lookahead** are hardcoded (the delay is shared with #7; real Shopify payouts are 7–37 days). Single currency assumed per app.

## Pricing-tier impact

**UNKNOWN — no plan-gating found** for any earnings/revenue view. Open question — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Read-only over synced transactions/aggregates. No migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Cross-surface consistency:** monthly-card net, daily-timeline net, and the earnings-report net all agree for the same period ≥ 99% (they read overlapping data via different paths).
2. **Concentration self-consistency:** top-N store `pct_of_total` + "rest" = 100% (± rounding) for any app.
3. **View utility:** each live view is opened by ≥ X% of active partners monthly; any view (e.g. earnings-status) below the bar is either wired to a screen or removed.

## Non-goals (deliberately absent in code)

1. **No reconciliation / "paid correctly?" audit here** — that's the Guard wedge (#7). **INFERRED.**
2. **No refund-heavy store visibility in concentration** — negative-net transactions are skipped, so an all-refund store never appears. **FACT.**
3. **No multi-currency detection** — all views assume one currency per app. **FACT.**
4. **No cross-app portfolio revenue roll-up** — endpoints are single-app. **INFERRED.**

## Open questions

- **`/earnings/status` and the daily `/earnings` timeline are live endpoints with no UI** — wire them (upcoming availability is genuinely useful) or remove? Owner: Product/Eng.
- **Concentration skips negative-net txs** (`revenue_metrics_service.go:288`) — a refund-heavy store is invisible; intended? Owner: Eng.
- **Empty domain → "unknown" bucket** (`:292-293`) — if sync ever nulls a valid domain, revenue mis-buckets. Owner: Eng.
- **30-day upcoming lookahead is hardcoded** — intentional (UI paging) or placeholder? Owner: Eng.
- **`GetRevenueConcentration` & `GetEarningsStatus` are untested.** Owner: Eng.
- Multi-currency handling + pricing gating. Owner: Eng/Product.

# PRD: Subscriptions Browsing & ARPU/LTV (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/interfaces/http/handler/subscription.go`, `handler/subscriptions_report_handler.go`,
> `entity/subscription.go`, `repository/subscription_repository.go`,
> `frontend-flutter/lib/screens/subscriptions/`. (Risk classification: PRD #5. Report framing: PRD #1.)

## Problem & evidence

A partner needs to move from aggregate KPIs to **the individual customers behind them** — find a specific store, filter to who's at risk, and inspect one subscription's plan, tenure, payment history, and risk trajectory. Plus a per-plan **ARPU/LTV** read to understand unit economics.

- **FACT:** a full list/detail/search browsing surface + ARPU/LTV report exist and are tested.
- **UNKNOWN — original demand evidence not recorded.** Open question: validate the filter/search set partners actually use — owner: Product.

## Target users

**INFERRED:** the **Shopify app partner** (founder/growth/success) drilling from dashboard into specific customers. Per-app, per-org.
- **Not the target:** merchants; bulk data-export/BI users (there's CSV on the report, but this isn't a data-warehouse surface).

## Proposed solution (as-built behavior)

**FACT — list** (`subscription.go:52-197`, `subscription_list_screen.dart`): paginated (25/page, max 100), with **DB-side** filters (risk state, subscription status incl. UNINSTALLED, plan, price range, billing interval), **server-side case-insensitive search** on shop name/domain, and sort (risk_state / price / shop_name). KPI cards show active / at-risk / churned / avg price (`/subscriptions/summary`).

**FACT — detail** (`subscription.go:354-434`, `subscription_detail_screen.dart`): billing info (plan, base price, MRR, interval, created date), status + risk badges, next-charge / days-since-last-payment, **payment history** (last ~12 months via `/history`), and a **risk timeline** of state-change events via `/risk-timeline` (this surfaces the `subscription_event` transitions).

**FACT — ARPU/LTV report** (`subscriptions_report_handler.go:74-127`): active subs, active MRR, **ARPU** = Σ MRR of SAFE ÷ count of SAFE (floored), **LTV** = ARPU ÷ monthly churn rate (0 when churn = 0), churn rate = churned ÷ latest-snapshot total, plus per-plan breakdown and CSV export.

### Key screens

**No new wireframe** (as-built). Real surfaces: `frontend-flutter/lib/screens/subscriptions/` (list + detail) and the subscriptions ARPU/LTV report screen.

## Platform & policy constraints

**FACT.** All data is synced-and-rebuilt subscription state (risk/plan/price from the ledger rebuild, PRD #5/#6); freshness is sync-bound. ARPU/LTV depend on snapshot-derived churn, so their accuracy inherits snapshot freshness. Currency is per-subscription; the report collapses to a single currency (see Non-goals).

## Pricing-tier impact

**UNKNOWN — no plan-gating found** for browsing or the ARPU/LTV report. Open question — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Read-only over rebuilt subscription state; no migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Findability:** a partner can locate a specific store (search) or a filtered cohort (e.g. all 2-cycle-missed) in ≤ 2 interactions, p95.
2. **Cross-surface consistency:** the list's active/at-risk/churned counts equal the Dashboard + Risk Summary for the same sync ≥ 99% (shares the risk-convergence guarantee, PRD #5).
3. **LTV credibility:** per-plan LTV is within ±X% of a per-plan-churn recomputation (validates the app-level-churn approximation before partners lean on it).

## Non-goals (deliberately absent in code)

1. **No per-plan churn in LTV** — LTV uses app-level churn for every plan (approximation). **FACT.**
2. **No multi-currency reporting** — the report picks the first non-empty currency; mixed-currency apps render as one. **FACT.**
3. **No forward cohort-decay LTV** — LTV is `ARPU ÷ current churn`, a point-in-time estimate, not a cohort model. **FACT.**
4. **No inline charge history in the list** — detail/history is a separate endpoint (deliberate API separation). **INFERRED.**

## Open questions

- **LTV = 0 when churn = 0** — the frontend must render "—" not "$0"; is that handled everywhere? Owner: Eng/Product.
- **Per-plan LTV uses app-level churn** — acceptable approximation, or compute per-plan churn (metric #3)? Owner: Eng/Product.
- **Annual-MRR round vs floor mismatch:** frontend rounds `price/12`, backend floors — a 1¢ cosmetic discrepancy. Align? Owner: Eng.
- **Churn denominator uses the latest snapshot in a 90-day window** while the Churn report uses 30 days — should ARPU/LTV and Churn share one window? Owner: Eng.
- **Multi-currency apps** render a single currency — detect/segregate? Owner: Eng/Product.
- Pricing/tier gating. Owner: Product.

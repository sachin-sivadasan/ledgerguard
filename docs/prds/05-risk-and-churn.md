# PRD: Risk & Churn (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/domain/service/risk_engine.go`, `valueobject/risk_state.go`,
> `handler/{risk_handler,churn_handler,retention_handler,revenue_at_risk_handler}.go`,
> `frontend-flutter/lib/screens/risk/`. (The Revenue-at-Risk *report* framing lives in PRD #1.)

## Problem & evidence

A Shopify app partner's recurring revenue quietly erodes when subscriptions miss renewal charges. Partners need an **unambiguous, deterministic signal** of *which* subscriptions are slipping, *how far*, and *how much MRR* is at stake — plus a clear churn/retention read.

- **FACT:** the capability is fully built, persisted, and the most heavily-tested in the codebase (31 risk-engine tests incl. all day-boundaries), signaling it's treated as core.
- **INFERRED (from `docs/REPORTS.md`):** deterministic risk is positioned as a wedge vs black-box churn scores. **UNKNOWN — original demand evidence not recorded** → Open question (owner: Product).

## Target users

**INFERRED:** the **Shopify app partner** (founder/growth/success) who wants to act on at-risk revenue before it's lost. Per-app, per-org.
- **Not the target:** merchants; users wanting predictive ML churn scoring (the engine is rule-based by design — see Non-goals).

## Proposed solution (as-built behavior)

**FACT — the risk state machine** (`risk_engine.go:24-88`), deterministic & pure, classifies each subscription into one of four states:
- **CANCELLED/EXPIRED → CHURNED** (terminal, `:28-30`); **FROZEN → ONE_CYCLE_MISSED** (payment-failure short-circuit, `:33-35`); **PENDING → SAFE** (`:38-40`); **ACTIVE with future/nil next-charge → SAFE** (`:44-52`).
- Otherwise by **days past due** (`:75-88`): ≤30 → SAFE (grace), 31–60 → ONE_CYCLE_MISSED, 61–90 → TWO_CYCLES_MISSED, >90 → CHURNED.

**FACT — surfaces:**
- **Risk Summary** (`/risk/summary`, `risk_handler.go:39`): 4-state distribution + at-risk stores ranked (health score, install date). Frontend `screens/risk/`.
- **Churn report** (`/reports/churn`): churned count, MRR lost, churn rate (clamped 0–1), churned stores by MRR lost, tenure, churn trend.
- **Retention report** (`/reports/retention`): renewal rate, retained MRR (Σ SAFE MRR), **reactivations** (distinct shops with `REACTIVAT*` events in range), per-plan renewal, renewal trend.
- **Revenue at Risk** (`/reports/revenue-at-risk`): at-risk MRR + **recoverable revenue** (60% of 1-cycle + 25% of 2-cycle, `revenue_at_risk_handler.go:23-26`). *(Report framing: PRD #1.)*

**FACT — lifecycle:** risk state is computed **once per sync** during ledger rebuild (`RiskEngine.ClassifyAll`), **persisted** on the subscription (`risk_state` column), and **never recomputed on read** (RISK-1b, `risk_handler.go:55`) so Dashboard/Risk/Subscriptions stay convergent. It then feeds `daily_metrics_snapshot` counts (`SafeCount`, `OneCycleMissedCount`, …, `RevenueAtRiskCents`).

### Key screens

**No new wireframe** (as-built). Real surface: `frontend-flutter/lib/screens/risk/` (distribution + at-risk list) plus the churn/retention/revenue-at-risk report screens.

## Platform & policy constraints

**FACT.** Classification depends entirely on synced Shopify subscription `status` + `expectedNextChargeDate`; accuracy is bounded by sync freshness. Time math assumes **UTC** (no DST/timezone correction, `risk_engine.go:62-72`). The 30/60/90-day thresholds are product policy encoded in code (CLAUDE.md §13).

## Pricing-tier impact

**UNKNOWN — no plan-gating found.** Open question: is Revenue-at-Risk / recoverable-revenue a premium feature? — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Risk state is (re)derived every sync (deterministic/idempotent), so no migration; a threshold change would reclassify on the next sync.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Determinism/consistency:** risk distribution shown on Dashboard, Risk Summary, and Subscriptions match for the same sync (guards RISK-1b) in ≥99% of checks.
2. **Actionability:** ≥ 40% of partners with at-risk MRR view the ranked at-risk store list within a week of it appearing.
3. **Recovery-model honesty:** measured actual recovery rate of 1-cycle vs 2-cycle subs is within ±10pp of the hardcoded 60%/25% — else the constants must be recalibrated (they're currently guesses).

## Non-goals (deliberately absent in code)

1. **No ML/predictive churn scoring** — rule-based state machine only. **FACT.**
2. **No per-subscription risk history/audit UI** — transitions are recorded (`subscription_event` from/to risk state) but not surfaced as a timeline. **INFERRED.**
3. **No timezone-aware days-late** — UTC-only. **FACT.**
4. **No handling of PAUSED status** — falls through to ACTIVE logic (see Open questions). **FACT.**

## Open questions

- **PAUSED (and any non-enumerated status) falls through to ACTIVE logic** (`risk_engine.go` default path) — is that intended, or should PAUSED be its own state? Owner: Product/Eng.
- Recoverable-revenue constants (60%/25%) are unvalidated guesses — calibrate from reactivation data (metric #3). Owner: Eng/Product.
- Churn count is **live** while retention/renewal is **snapshot-sourced** — acceptable drift, or should both use one source? Owner: Eng.
- **CLAUDE.md §13 pseudo-code is simplified** vs the real (more defensive) engine — update the doc to match. Owner: Eng.
- Timezone/DST: are all upstream dates truly UTC? A merchant near a day-boundary could shift a state by one. Owner: Eng.

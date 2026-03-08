# PLAN-04: Revenue Intelligence Core (Ledger, Risk, Metrics)

**Date:** 2026-02-26
**Status:** Completed

## Scope
- Deterministic Ledger Rebuild (12-month rolling window)
- RiskEngine — classify subscriptions into risk states (SAFE, ONE_CYCLE_MISSED, TWO_CYCLE_MISSED, CHURNED)
- MetricsEngine — calculate MRR, usage revenue, renewal rate, revenue at risk, churn rate
- AIInsightService — generate daily AI insights via Claude API
- Daily metrics snapshots (one per app per day, never deleted)

## Key Decisions
- ADR-002: Full Ledger Rebuild over Incremental Updates

## Architecture
- Ledger rebuild is idempotent and deterministic (same input = same output)
- Risk classification based on days past expected charge date
- Revenue strictly separated by ChargeType (RECURRING, USAGE, ONE_TIME, REFUND)
- MRR = RECURRING only; Usage Revenue = USAGE only

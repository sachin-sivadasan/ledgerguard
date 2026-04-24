# 07. Risk Engine

## What It Does
Classifies each subscription into one of four risk states based on subscription status and payment timeliness. This is the authoritative risk classification for all of LedgerGuard. The engine determines which subscriptions are healthy, which are at risk of churning, and which have already churned. It also computes aggregate risk metrics: a risk distribution summary and total revenue at risk.

The four risk states form a linear progression:
- **SAFE** — Active, current, or within 30-day grace period
- **ONE_CYCLE_MISSED** — 31-60 days past due, or payment frozen
- **TWO_CYCLES_MISSED** — 61-90 days past due
- **CHURNED** — Over 90 days past due, or terminal status (CANCELLED/EXPIRED)

## Architecture
Domain service layer (`internal/domain/service/`). Pure Go with zero external dependencies — no database calls, no HTTP clients, no configuration. The RiskEngine is a stateless struct that takes subscription data and a timestamp as input and returns risk classifications. This makes it trivially testable and safe to call from any context.

The same risk classification logic also exists as a method on the `Subscription` entity itself (`ClassifyRisk(now)`), providing both a standalone service interface and an entity-level convenience method. The two implementations follow identical logic.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/domain/service/risk_engine.go` | ~139 | RiskEngine service: ClassifyRisk, DaysPastDue, summary, revenue at risk |
| `backend/internal/domain/service/risk_engine_test.go` | ~530 | Comprehensive tests for all states, edge cases, boundary conditions |
| `backend/internal/domain/entity/subscription.go` | ~137 | Entity-level ClassifyRisk(now) method with identical logic |
| `backend/internal/domain/valueobject/risk_state.go` | ~49 | RiskState type: SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED |

## Data Flow
```
ClassifyRisk(subscription, now)
│
├── Is status terminal (CANCELLED, EXPIRED)?
│     └── Yes → CHURNED
│
├── Is status FROZEN?
│     └── Yes → ONE_CYCLE_MISSED
│
├── Is status PENDING?
│     └── Yes → SAFE
│
├── Is status ACTIVE with future/current ExpectedNextChargeDate?
│     └── Yes → SAFE
│
├── Is ExpectedNextChargeDate nil?
│     └── Yes → SAFE (cannot classify without date)
│
└── Calculate days past due:
      daysPastDue = (now - ExpectedNextChargeDate) in days
      │
      ├── ≤ 0 days  → SAFE
      ├── 1-30 days → SAFE (grace period)
      ├── 31-60 days → ONE_CYCLE_MISSED
      ├── 61-90 days → TWO_CYCLES_MISSED
      └── > 90 days  → CHURNED
```

### Aggregate Methods
```
CalculateRiskSummary(subscriptions)        CalculateRevenueAtRisk(subscriptions)
│                                          │
├── Count by RiskState:                    ├── For each subscription:
│   SafeCount                              │   if RiskState.IsAtRisk():
│   OneCycleMissedCount                    │     atRisk += sub.MRRCents()
│   TwoCyclesMissedCount                   │
│   ChurnedCount                           └── Returns total cents at risk
│
└── Returns RiskSummary struct
```

## Configuration
No configuration needed. All thresholds are hardcoded constants in the classification logic:

| Threshold | Value | Risk State |
|-----------|-------|------------|
| Grace period | 0-30 days past due | SAFE |
| First missed cycle | 31-60 days past due | ONE_CYCLE_MISSED |
| Second missed cycle | 61-90 days past due | TWO_CYCLES_MISSED |
| Churned | > 90 days past due | CHURNED |

Status-based overrides (applied before days-past-due calculation):

| Status | Risk State | Rationale |
|--------|------------|-----------|
| CANCELLED | CHURNED | Terminal — subscription ended |
| EXPIRED | CHURNED | Terminal — subscription ended |
| FROZEN | ONE_CYCLE_MISSED | Payment failure — Shopify has frozen billing |
| PENDING | SAFE | Not yet active — no risk signal |

## API Surface
The RiskEngine is not exposed directly via HTTP. Risk states are computed during the ledger rebuild and stored on each subscription record. They surface through:

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/subscriptions/{appId}` | Firebase | Subscription list includes `risk_state` field |
| GET | `/api/v1/metrics/{appId}` | Firebase | Snapshot includes risk counts and revenue at risk |
| POST | `/api/v1/sync` | Firebase | Sync response includes `RiskSummary` |
| POST | `/graphql` | Firebase | Subscriptions and risk queries via internal GraphQL |

The AI chat modules (`risk` module) also expose risk data via tool calls: `risk__get_risk_summary`, `risk__get_at_risk_subscriptions`, `risk__get_risk_details`.

## Extension Points
- **Custom thresholds** — the 30/60/90 day boundaries could be made configurable per app or per plan tier. Currently hardcoded for simplicity.
- **Additional risk signals** — the engine currently uses only status and days-past-due. Future signals could include: payment retry count, historical churn patterns, store health score, review sentiment.
- **RiskState value object** — adding new states (e.g., `REACTIVATED`, `GRACE_PERIOD`) requires updating the `valueobject.RiskState` constants, the `ClassifyRisk` switch, and the `RiskSummary` struct.
- **IsAtRisk() predicate** — defined on the `RiskState` value object. Returns true for `ONE_CYCLE_MISSED` and `TWO_CYCLES_MISSED`. Used by `CalculateRevenueAtRisk()` and the metrics engine.
- **ClassifyAll()** — batch classification method that mutates risk state on a slice of subscriptions in place. Useful for bulk operations.

## Gotchas
- **Two implementations exist.** Both `RiskEngine.ClassifyRisk(sub, now)` (domain service) and `Subscription.ClassifyRisk(now)` (entity method) contain the same logic. The entity method does not check for FROZEN/PENDING/terminal statuses as comprehensively as the service — it only checks ACTIVE status and days-past-due. The service version is the authoritative one; the entity version is a convenience for the ledger rebuild.
- **Grace period is generous.** The first 30 days past due are still classified as SAFE. This aligns with Shopify's billing retry window but means truly delinquent subscriptions are not flagged immediately.
- **FROZEN maps to ONE_CYCLE_MISSED, not TWO.** Even if a subscription has been frozen for months, the status-based override always returns ONE_CYCLE_MISSED. The days-past-due path is only reached for ACTIVE subscriptions.
- **Nil ExpectedNextChargeDate defaults to SAFE.** If no expected charge date is set (e.g., free plans, newly created subscriptions), the engine cannot calculate days past due and returns SAFE. This is a safe default but may mask genuinely missing data.
- **DaysPastDue truncates to int.** The calculation is `int(hours / 24)`, which truncates partial days. A subscription 30.9 days past due is classified as 30 days (SAFE), not 31 days (ONE_CYCLE_MISSED).
- **Revenue at risk includes only ONE_CYCLE_MISSED and TWO_CYCLES_MISSED.** CHURNED subscriptions are excluded because that revenue is considered already lost, not "at risk."
- **MRRCents() for annual subscriptions divides by 12.** Revenue at risk for a $1200/year subscription is reported as $100/month MRR equivalent, not the full annual amount.

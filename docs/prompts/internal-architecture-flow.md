# Internal Architecture Flow - Implementation Prompt

## Overview

Create an animated, interactive visualization showing how LedgerGuard integrates with Shopify's Partner API and processes data internally. This is for internal/technical audiences (team, investors, technical partners).

---

## Purpose

- **Internal documentation** - Team onboarding, architecture understanding
- **Investor presentations** - Show technical sophistication
- **Technical sales** - Demonstrate data flow to technical buyers

---

## Data Flow Stages

### Stage 1: Data Ingestion (Shopify → LedgerGuard)

```
[Shopify Partner Account]
        │
        ▼ OAuth (read-only)
[Partner API GraphQL]
        │
        ├─── App Subscriptions (appSubscription)
        │    └─ id, name, status, currentPeriodEnd, price
        │
        ├─── Transactions (transactions)
        │    └─ id, type, amount, shop, createdAt
        │    └─ Types: AppSubscriptionSale, AppUsageSale, AppOneTimeSale
        │
        └─── Shop Details (shop)
             └─ domain, name, plan
        │
        ▼
[LedgerGuard Sync Engine]
```

### Stage 2: Data Processing (Internal Pipeline)

```
[Raw Transactions]
        │
        ▼ Every 12 hours
[Transaction Repository]
        │ Batch Upsert (idempotent)
        ▼
[Ledger Rebuild Service]
        │
        ├─── Classify Charge Types
        │    └─ RECURRING (MRR)
        │    └─ USAGE (metered)
        │    └─ ONE_TIME (setup fees)
        │    └─ REFUND (negative)
        │
        ├─── Build Subscriptions
        │    └─ Group by shop
        │    └─ Calculate billing interval
        │    └─ Set expected next charge
        │
        └─── Compute Risk State
             └─ Compare expected vs actual
             └─ Days overdue → risk tier
```

### Stage 3: Risk Classification

```
[Expected Next Charge Date]
        │
        ▼ Compare to Today
[Days Overdue Calculation]
        │
        ├─── 0-30 days  → SAFE (green)
        ├─── 31-60 days → ONE_CYCLE_MISSED (yellow)
        ├─── 61-90 days → TWO_CYCLES_MISSED (orange)
        └─── 90+ days   → CHURNED (red)
```

### Stage 4: Metrics Computation

```
[Subscription States]
        │
        ▼
[Metrics Engine]
        │
        ├─── Renewal Success Rate
        │    └─ SAFE / Total Active
        │
        ├─── Active MRR
        │    └─ Sum of SAFE subscription MRR
        │
        ├─── Revenue at Risk
        │    └─ Sum of AT_RISK subscription MRR
        │
        ├─── Usage Revenue
        │    └─ Sum of USAGE transactions in period
        │
        └─── Total Revenue
             └─ RECURRING + USAGE + ONE_TIME - REFUNDS
        │
        ▼
[Daily Metrics Snapshot]
        │ Stored per app per day
        ▼
[Trend Analysis]
```

### Stage 5: Output Layer

```
[Metrics + Risk Data]
        │
        ├─────────────────┬──────────────────┬──────────────────┐
        ▼                 ▼                  ▼                  ▼
[Dashboard UI]    [Push Alerts]      [AI Daily Brief]    [Revenue API]
        │              │                   │                   │
        │         Slack/Email          Claude AI           REST/GraphQL
        │              │                   │                   │
        ▼              ▼                   ▼                   ▼
   [User Views]   [Notifications]    [Executive          [External
    KPIs, Risk,   Critical alerts,    Summary]            Integrations]
    Cohorts]      Daily digests
```

---

## Visual Design

### Color Coding by Data Type

| Data Type | Color | Hex |
|-----------|-------|-----|
| Shopify/External | Green | #22c55e |
| Raw Data | Blue | #3b82f6 |
| Processing | Purple | #8b5cf6 |
| Risk/Alerts | Amber/Red | #f59e0b / #ef4444 |
| Output | Cyan | #06b6d4 |

### Animation Sequences

1. **Data Ingestion Flow**
   - API request pulse from LedgerGuard to Shopify
   - Data packets flowing back
   - Transaction types splitting

2. **Ledger Rebuild**
   - Transactions grouping by type
   - Subscriptions being constructed
   - Risk badges appearing

3. **Risk Classification**
   - Timeline visualization
   - Days counting up
   - Risk state changing colors

4. **Metrics Computation**
   - Numbers calculating/counting
   - Gauges filling
   - Delta indicators appearing

5. **Output Distribution**
   - Data splitting to multiple outputs
   - Dashboard updating
   - Alerts firing
   - API responses

---

## Interactive Controls

### View Modes
1. **Full Pipeline** - See entire flow end-to-end
2. **Ingestion Focus** - Zoom into Shopify API details
3. **Processing Focus** - Zoom into ledger rebuild
4. **Risk Focus** - Zoom into classification logic
5. **Output Focus** - Zoom into delivery channels

### Animation Controls
- Play/Pause
- Speed (0.5x, 1x, 2x)
- Step-by-step mode
- Reset

### Data Toggles
- Show/hide raw data examples
- Show/hide timing info (12h sync, 7-37 day payout)
- Show/hide code snippets

---

## Code Snippets to Show

### GraphQL Query (Ingestion)
```graphql
query FetchTransactions($appId: ID!, $after: String) {
  app(id: $appId) {
    transactions(first: 100, after: $after) {
      edges {
        node {
          id
          type: __typename
          grossAmount { amount }
          netAmount { amount }
          shop { name domain }
          createdAt
        }
      }
    }
  }
}
```

### Risk Classification (Go)
```go
func ClassifyRisk(expectedNext time.Time, now time.Time) RiskState {
    daysLate := int(now.Sub(expectedNext).Hours() / 24)
    switch {
    case daysLate <= 30:
        return RiskSafe
    case daysLate <= 60:
        return RiskOneCycleMissed
    case daysLate <= 90:
        return RiskTwoCyclesMissed
    default:
        return RiskChurned
    }
}
```

### Metrics Calculation (Go)
```go
func (e *MetricsEngine) ComputeMetrics(subs []Subscription) Metrics {
    var activeMRR, atRiskMRR int64
    var safeCount, atRiskCount int

    for _, sub := range subs {
        switch sub.RiskState {
        case RiskSafe:
            activeMRR += sub.MRRCents
            safeCount++
        case RiskOneCycleMissed, RiskTwoCyclesMissed:
            atRiskMRR += sub.MRRCents
            atRiskCount++
        }
    }

    renewalRate := float64(safeCount) / float64(safeCount + atRiskCount)
    return Metrics{ActiveMRR: activeMRR, AtRiskMRR: atRiskMRR, RenewalRate: renewalRate}
}
```

---

## Section Layout

### Section 1: Header
- Title: "LedgerGuard Architecture"
- Subtitle: "How we turn Shopify data into revenue intelligence"
- View mode selector

### Section 2: Main Visualization
- Animated flow diagram
- Entity boxes with icons
- Animated connection lines
- Data packets moving

### Section 3: Detail Panels
- Expandable code snippets
- Timing information
- Data examples

### Section 4: Metrics Output Preview
- Live-updating sample metrics
- Risk distribution
- Revenue breakdown

---

## Technical Implementation

### Route
`/architecture` or `/internal/flow`

### Component
`marketing/site/components/ArchitectureFlow.tsx`

### Animation Library
- CSS animations + transitions
- Optional: Framer Motion for complex sequences

### State Management
- Current view mode
- Animation phase
- Active step (for step-by-step)
- Speed multiplier

---

## Sample Data for Animation

### Transaction Examples
```json
{
  "transactions": [
    {"type": "AppSubscriptionSale", "amount": 29.00, "shop": "acme-store.myshopify.com"},
    {"type": "AppUsageSale", "amount": 12.50, "shop": "techgadgets.myshopify.com"},
    {"type": "AppSubscriptionSale", "amount": 99.00, "shop": "fashionhub.myshopify.com"}
  ]
}
```

### Subscription Examples
```json
{
  "subscriptions": [
    {"shop": "acme-store", "mrr": 2900, "riskState": "SAFE", "daysSinceCharge": 5},
    {"shop": "techgadgets", "mrr": 4900, "riskState": "ONE_CYCLE_MISSED", "daysSinceCharge": 42},
    {"shop": "fashionhub", "mrr": 9900, "riskState": "SAFE", "daysSinceCharge": 12}
  ]
}
```

### Metrics Output
```json
{
  "activeMRR": 47392,
  "revenueAtRisk": 8240,
  "renewalRate": 94.2,
  "usageRevenue": 3420,
  "totalRevenue": 51812
}
```

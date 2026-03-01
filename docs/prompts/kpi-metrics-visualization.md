# KPI Metrics Visualization - Interactive Dashboard Guide

## Context
You are a senior frontend + visualization engineer building an interactive animated guide showing how LedgerGuard calculates and presents KPIs for Shopify app developers.

Build an educational visualization that helps developers understand:
1. What each KPI measures and WHY it matters
2. How each metric is calculated (formulas, data sources)
3. How risk classification works (the payment lifecycle)
4. How period-over-period comparisons reveal trends

---

## Design Philosophy

### Target Audience
Shopify app developers who:
- Understand basic SaaS metrics (MRR, churn)
- May NOT know how Shopify Partner API works
- Want to understand their revenue health at a glance
- Need actionable insights, not just numbers

### Key Principles
1. **Show, don't tell** - Animated data flows beat static text
2. **Progressive disclosure** - Start simple, allow deep dives
3. **Contextual meaning** - Every number needs "is this good?"
4. **Real math** - Show actual formulas with example calculations

---

## KPI Categories

### Category 1: Revenue KPIs (Money Metrics)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           REVENUE KPIs                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  💰 ACTIVE MRR              📉 REVENUE AT RISK         📊 USAGE REVENUE     │
│  ─────────────              ─────────────────          ───────────────      │
│  MRR from healthy           MRR that may be            Revenue from         │
│  subscriptions only         lost (at-risk stores)      metered billing      │
│                                                                              │
│  💵 TOTAL REVENUE           🔴 CHURNED REVENUE                              │
│  ─────────────              ────────────────                                │
│  All revenue combined       MRR already lost                                │
│  (recurring + usage)        (stores that left)                              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Category 2: Health KPIs (Status Metrics)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           HEALTH KPIs                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ✅ RENEWAL SUCCESS RATE                                                    │
│  ────────────────────────                                                   │
│  % of subscriptions renewing on time                                        │
│  Formula: (Safe Count / Total Subscriptions) × 100                          │
│                                                                              │
│  📊 RISK DISTRIBUTION                                                       │
│  ─────────────────────                                                      │
│  How your subscriptions are distributed across risk states:                 │
│                                                                              │
│  [████████████████████░░░░░░░░]  Safe (72%)                                │
│  [████░░░░░░░░░░░░░░░░░░░░░░░░]  At Risk (15%)                             │
│  [██░░░░░░░░░░░░░░░░░░░░░░░░░░]  Critical (8%)                             │
│  [█░░░░░░░░░░░░░░░░░░░░░░░░░░░]  Churned (5%)                              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## KPI Definitions & Formulas

### 1. Active MRR (Monthly Recurring Revenue)

**What it measures:** Predictable monthly revenue from healthy, renewing subscriptions.

**Why it matters:** This is your "safe" revenue - money you can count on next month.

**Formula:**
```
Active MRR = SUM(MRR) for all subscriptions WHERE RiskState = SAFE

For each subscription:
  - If MONTHLY: MRR = BasePriceCents
  - If ANNUAL:  MRR = BasePriceCents / 12
```

**Example Calculation:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ Store              │ Plan      │ Price    │ Interval │ Risk   │ MRR │
├────────────────────┼───────────┼──────────┼──────────┼────────┼─────┤
│ cool-store.myshop  │ Pro       │ $49/mo   │ Monthly  │ SAFE   │ $49 │
│ mega-shop.myshop   │ Business  │ $588/yr  │ Annual   │ SAFE   │ $49 │
│ tiny-biz.myshop    │ Starter   │ $19/mo   │ Monthly  │ AT_RISK│ $0  │ ← Excluded!
│ big-corp.myshop    │ Enterprise│ $99/mo   │ Monthly  │ SAFE   │ $99 │
├────────────────────┴───────────┴──────────┴──────────┴────────┼─────┤
│                                              Active MRR Total: │$197 │
└────────────────────────────────────────────────────────────────┴─────┘
```

**Semantic:** Higher is good ↑ (green indicator)

**Data Flow Animation:**
```
[Subscriptions DB] → Filter: RiskState=SAFE → Calculate MRR → Sum → [Active MRR]
       ↓                      ↓                    ↓
   847 stores          →   612 SAFE         →   Sum each    →   $12,450
                            stores              MRR value
```

---

### 2. Revenue at Risk

**What it measures:** MRR from stores that missed payment(s) but haven't churned yet.

**Why it matters:** This is revenue you might LOSE if you don't intervene. It's an early warning system.

**Formula:**
```
Revenue at Risk = SUM(MRR) for all subscriptions WHERE RiskState IN (ONE_CYCLE_MISSED, TWO_CYCLES_MISSED)
```

**Risk State Breakdown:**
```
ONE_CYCLE_MISSED   = 31-60 days past due  (⚠️ Warning)
TWO_CYCLES_MISSED  = 61-90 days past due  (🔴 Critical)
```

**Example:**
```
┌─────────────────────────────────────────────────────────────────────┐
│ Store              │ Days Past Due │ Risk State          │ MRR     │
├────────────────────┼───────────────┼─────────────────────┼─────────┤
│ slow-payer.myshop  │ 45 days       │ ONE_CYCLE_MISSED    │ $29     │
│ trouble-co.myshop  │ 72 days       │ TWO_CYCLES_MISSED   │ $49     │
│ late-again.myshop  │ 38 days       │ ONE_CYCLE_MISSED    │ $19     │
├────────────────────┴───────────────┴─────────────────────┼─────────┤
│                                      Revenue at Risk:     │ $97     │
└──────────────────────────────────────────────────────────┴─────────┘
```

**Semantic:** Lower is good ↓ (red when high, green when low)

**Action Prompt:** "You have $97 at risk. Consider reaching out to these 3 stores."

---

### 3. Renewal Success Rate

**What it measures:** The percentage of your subscriptions that are healthy and renewing.

**Why it matters:** A high renewal rate means your app is sticky and valuable.

**Formula:**
```
Renewal Success Rate = (Safe Count / Total Subscriptions) × 100

Where:
  Safe Count = COUNT(*) WHERE RiskState = SAFE
  Total = COUNT(*) for all subscriptions
```

**Example:**
```
Total Subscriptions: 100
├── Safe:               72  ← In numerator
├── One Cycle Missed:    8
├── Two Cycles Missed:   5
└── Churned:            15

Renewal Success Rate = (72 / 100) × 100 = 72%
```

**Visual Representation:**
```
Renewal Success: 72%
[████████████████████████████░░░░░░░░░░]
 ↑ Safe (renewing)            ↑ Not safe
```

**Semantic:** Higher is good ↑

**Benchmarks:**
```
< 70%  = Poor (🔴)    - Significant retention issues
70-85% = Okay (🟡)    - Room for improvement
85-95% = Good (🟢)    - Healthy retention
> 95%  = Excellent (💎) - Best in class
```

---

### 4. Usage Revenue

**What it measures:** Revenue from metered/usage-based billing (not subscriptions).

**Why it matters:** Shows additional revenue beyond fixed subscriptions. Scales with merchant success.

**Formula:**
```
Usage Revenue = SUM(NetAmountCents) for all transactions WHERE ChargeType = USAGE
```

**Usage Charge Examples:**
```
┌────────────────────────────────────────────────────────────────┐
│ Use Case            │ Example           │ How It's Charged     │
├─────────────────────┼───────────────────┼──────────────────────┤
│ Per order           │ $0.05/order       │ APP_USAGE_SALE       │
│ Per SMS sent        │ $0.02/message     │ APP_USAGE_SALE       │
│ Per API call        │ $0.001/call       │ APP_USAGE_SALE       │
│ Per AI generation   │ $0.10/use         │ APP_USAGE_SALE       │
│ Storage overage     │ $0.10/GB          │ APP_USAGE_SALE       │
└─────────────────────┴───────────────────┴──────────────────────┘
```

**Revenue Separation (Critical!):**
```
                    ┌─────────────────┐
                    │  Total Revenue  │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         ↓                   ↓                   ↓
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│    RECURRING    │ │     USAGE       │ │    ONE-TIME     │
│   (from subs)   │ │  (metered)      │ │   (add-ons)     │
└─────────────────┘ └─────────────────┘ └─────────────────┘
  APP_SUBSCRIPTION    APP_USAGE_SALE     APP_ONE_TIME_SALE
        _SALE

⚠️ NEVER mix these in calculations!
   MRR = RECURRING only
   Usage Revenue = USAGE only
```

**Semantic:** Higher is good ↑

---

### 5. Total Revenue

**What it measures:** All revenue combined for a period.

**Formula:**
```
Total Revenue = RECURRING + USAGE + ONE_TIME - REFUNDS

From transactions:
  + SUM(Amount) WHERE ChargeType = RECURRING
  + SUM(Amount) WHERE ChargeType = USAGE
  + SUM(Amount) WHERE ChargeType = ONE_TIME
  - SUM(Amount) WHERE ChargeType = REFUND   ← Subtracted!
```

**Visual Breakdown:**
```
Total Revenue: $15,450

Composition:
[████████████████████████] Recurring    $12,000 (78%)
[██████░░░░░░░░░░░░░░░░░░] Usage        $3,000  (19%)
[█░░░░░░░░░░░░░░░░░░░░░░░] One-time     $500    (3%)
[░░░░░░░░░░░░░░░░░░░░░░░░] Refunds      -$50    (-0.3%)
```

**Semantic:** Higher is good ↑

---

### 6. Churned Revenue

**What it measures:** MRR from subscriptions that have fully churned (90+ days past due).

**Why it matters:** This is revenue you've LOST. Understanding churn helps prevent future losses.

**Formula:**
```
Churned Revenue = SUM(MRR) for all subscriptions WHERE RiskState = CHURNED
```

**Semantic:** Lower is good ↓ (this is a loss metric)

---

## Risk Classification Engine

### The Payment Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     SUBSCRIPTION PAYMENT LIFECYCLE                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Day 0              Day 30            Day 60            Day 90+             │
│    │                  │                 │                  │                │
│    ▼                  ▼                 ▼                  ▼                │
│  [CHARGE]          [EXPECTED]        [STILL]            [LOST]             │
│  Payment           Next charge       waiting...         Customer           │
│  successful        due date                             churned            │
│                                                                              │
│                                                                              │
│  ════════════════════════════════════════════════════════════════════════  │
│                                                                              │
│  DAYS PAST DUE:                                                             │
│                                                                              │
│  [0]─────[30]─────────[60]─────────[90]───────────────[∞]                  │
│   │        │            │            │                  │                   │
│   │   ✅ SAFE      ⚠️ ONE_CYCLE   🔴 TWO_CYCLES    💀 CHURNED            │
│   │   Grace period   MISSED         MISSED           Lost forever         │
│   │                                                                         │
│   └── Includes 30-day grace period for payment processing                   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Risk State Definitions

```
┌───────────────────────────────────────────────────────────────────────────┐
│ Risk State          │ Days Past Due │ Color │ Action Required            │
├─────────────────────┼───────────────┼───────┼────────────────────────────┤
│ ✅ SAFE             │ 0 - 30 days   │ Green │ None - healthy             │
│ ⚠️ ONE_CYCLE_MISSED │ 31 - 60 days  │ Yellow│ Reach out, offer help      │
│ 🔴 TWO_CYCLES_MISSED│ 61 - 90 days  │ Orange│ Urgent - last chance       │
│ 💀 CHURNED          │ 90+ days      │ Red   │ Lost - analyze why         │
└─────────────────────┴───────────────┴───────┴────────────────────────────┘
```

### Animated Risk Classification Flow

```
INPUT: Subscription Data
         │
         ▼
┌─────────────────────────────┐
│ Last Charge: Jan 15, 2026   │
│ Billing: Monthly            │
│ Expected Next: Feb 15, 2026 │
└─────────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ TODAY: March 20, 2026       │
│ Days Since Expected:        │
│   March 20 - Feb 15 = 33    │
└─────────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ CLASSIFICATION:             │
│                             │
│ 33 days > 30 (grace)        │
│ 33 days ≤ 60 (one cycle)    │
│                             │
│ → ONE_CYCLE_MISSED ⚠️       │
└─────────────────────────────┘
```

### Risk Classification Code (Actual Logic)

```go
func ClassifyRisk(subscription Subscription, now time.Time) RiskState {
    // If actively paid and not past due
    if subscription.Status == "ACTIVE" && now.Before(subscription.ExpectedNextChargeDate) {
        return SAFE
    }

    // Calculate days past expected charge date
    daysPastDue := int(now.Sub(subscription.ExpectedNextChargeDate).Hours() / 24)

    switch {
    case daysPastDue <= 30:
        return SAFE                 // Grace period
    case daysPastDue <= 60:
        return ONE_CYCLE_MISSED     // 31-60 days
    case daysPastDue <= 90:
        return TWO_CYCLES_MISSED    // 61-90 days
    default:
        return CHURNED              // 90+ days
    }
}
```

---

## Period-over-Period Comparison

### Delta Calculation

**Formula:**
```
Delta % = ((Current - Previous) / Previous) × 100

Special cases:
  - If Previous = 0 and Current ≠ 0: Show "New" (no comparison)
  - If Previous = 0 and Current = 0: Show 0% (no change)
```

### Semantic Interpretation (Is the change GOOD?)

```
┌────────────────────────┬─────────────┬─────────────┬──────────────────┐
│ Metric                 │ Direction   │ Positive Δ  │ Negative Δ       │
├────────────────────────┼─────────────┼─────────────┼──────────────────┤
│ Active MRR             │ Higher ↑    │ 🟢 Good     │ 🔴 Bad           │
│ Revenue at Risk        │ Lower ↓     │ 🔴 Bad      │ 🟢 Good          │
│ Usage Revenue          │ Higher ↑    │ 🟢 Good     │ 🔴 Bad           │
│ Total Revenue          │ Higher ↑    │ 🟢 Good     │ 🔴 Bad           │
│ Renewal Success Rate   │ Higher ↑    │ 🟢 Good     │ 🔴 Bad           │
│ Churn Count            │ Lower ↓     │ 🔴 Bad      │ 🟢 Good          │
│ Churned Revenue        │ Lower ↓     │ 🔴 Bad      │ 🟢 Good          │
└────────────────────────┴─────────────┴─────────────┴──────────────────┘
```

### Visual Delta Display

```
Active MRR                     Revenue at Risk
$12,450                        $1,850
  ↑ +5.2%  🟢                    ↓ -12.3%  🟢
  vs. last month                 vs. last month
  (was $11,835)                  (was $2,110)
```

### Animated Comparison Flow

```
           FEBRUARY 2026                    MARCH 2026
┌─────────────────────────┐         ┌─────────────────────────┐
│ Active MRR:    $11,835  │  ────▶  │ Active MRR:    $12,450  │
│ At Risk:       $2,110   │         │ At Risk:       $1,850   │
│ Renewal Rate:  89.2%    │         │ Renewal Rate:  91.5%    │
└─────────────────────────┘         └─────────────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────────────┐
                                    │      DELTA ANALYSIS     │
                                    │                         │
                                    │ MRR:     +$615  (+5.2%) │
                                    │ At Risk: -$260 (-12.3%) │
                                    │ Renewal: +2.3 pts       │
                                    │                         │
                                    │ 📈 Improving trend!     │
                                    └─────────────────────────┘
```

---

## Data Flow Architecture

### From Partner API to Dashboard

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        DATA FLOW: API → KPIs                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐                                                            │
│  │  Shopify    │                                                            │
│  │ Partner API │                                                            │
│  └──────┬──────┘                                                            │
│         │                                                                    │
│         │  GraphQL: transactions(last: 12 months)                           │
│         ▼                                                                    │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│  │ Transaction │     │ Transaction │     │ Transaction │  ...               │
│  │ RECURRING   │     │ USAGE       │     │ REFUND      │                   │
│  └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                   │
│         │                   │                   │                           │
│         └───────────────────┼───────────────────┘                           │
│                             │                                                │
│                             ▼                                                │
│                    ┌─────────────────┐                                      │
│                    │ LEDGER REBUILD  │                                      │
│                    │                 │                                      │
│                    │ 1. Group by     │                                      │
│                    │    domain       │                                      │
│                    │ 2. Build subs   │                                      │
│                    │ 3. Classify     │                                      │
│                    │    risk         │                                      │
│                    └────────┬────────┘                                      │
│                             │                                                │
│              ┌──────────────┼──────────────┐                                │
│              ▼              ▼              ▼                                │
│      ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                       │
│      │Subscriptions│ │ Transactions│ │Risk Summary │                       │
│      └──────┬──────┘ └──────┬──────┘ └──────┬──────┘                       │
│             │               │               │                               │
│             └───────────────┼───────────────┘                               │
│                             │                                                │
│                             ▼                                                │
│                    ┌─────────────────┐                                      │
│                    │ METRICS ENGINE  │                                      │
│                    │                 │                                      │
│                    │ • Active MRR    │                                      │
│                    │ • Revenue@Risk  │                                      │
│                    │ • Usage Revenue │                                      │
│                    │ • Renewal Rate  │                                      │
│                    │ • Risk Counts   │                                      │
│                    └────────┬────────┘                                      │
│                             │                                                │
│                             ▼                                                │
│                    ┌─────────────────┐                                      │
│                    │  DAILY SNAPSHOT │◄── Stored permanently                │
│                    │  (immutable)    │    One per app per day               │
│                    └────────┬────────┘                                      │
│                             │                                                │
│                             ▼                                                │
│                    ┌─────────────────┐                                      │
│                    │   DASHBOARD     │                                      │
│                    │   KPI Cards     │                                      │
│                    └─────────────────┘                                      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Interactive Elements

### Toggle Controls
- **Time Range:** This Month | Last 30 Days | Last 90 Days | Custom
- **View Mode:** Summary | Detailed | Comparison
- **KPI Focus:** All | Revenue | Health | Risk

### Animated Scenarios
1. **New Subscription Flow:** Watch MRR increase
2. **Missed Payment Flow:** Watch risk state change
3. **Churn Flow:** Watch revenue move to churned
4. **Recovery Flow:** Watch at-risk return to safe

### Hover Interactions
- Hover on any KPI card → Show formula + example calculation
- Hover on risk state → Show count + total MRR in that state
- Hover on delta → Show previous vs current values

---

## Visual Requirements

### Color Scheme
```
✅ Safe / Good:        #22c55e (Green)
⚠️ Warning / At Risk:  #f59e0b (Amber)
🔴 Critical:           #ef4444 (Red)
💀 Churned:            #6b7280 (Gray)
📈 Positive Delta:     #22c55e (Green)
📉 Negative Delta:     #ef4444 (Red)
💰 Revenue:            #3b82f6 (Blue)
📊 Neutral:            #8b5cf6 (Purple)
```

### KPI Card Layout
```
┌─────────────────────────────────┐
│ 💰 Active MRR                   │
│                                 │
│     $12,450                     │  ← Large, prominent
│     ↑ +5.2% vs last month       │  ← Delta with direction
│                                 │
│ ─────────────────────────────── │
│ From 612 healthy subscriptions  │  ← Context
└─────────────────────────────────┘
```

### Risk Distribution Visualization
```
     Risk Distribution (847 subscriptions)

     ┌────────────────────────────────────┐
     │████████████████████████████░░░░░░░░│ Safe: 612 (72%)
     │████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░│ At Risk: 127 (15%)
     │████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│ Critical: 68 (8%)
     │██░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│ Churned: 40 (5%)
     └────────────────────────────────────┘
```

---

## Animation Sequences

### Sequence 1: Understanding Active MRR
```
Step 1: Show all subscriptions as boxes
Step 2: Highlight SAFE subscriptions (green glow)
Step 3: Show MRR values floating up from each
Step 4: Values converge into total
Step 5: Display final Active MRR card
```

### Sequence 2: Risk Classification
```
Step 1: Show a subscription timeline
Step 2: Mark expected charge date
Step 3: Animate time passing (days counter)
Step 4: Show risk state changing at thresholds
Step 5: Display final risk state badge
```

### Sequence 3: Period Comparison
```
Step 1: Show February snapshot (left)
Step 2: Show March snapshot (right)
Step 3: Draw connecting lines between metrics
Step 4: Calculate and animate delta values
Step 5: Color code good/bad changes
```

### Sequence 4: Revenue Breakdown
```
Step 1: Show total revenue as full bar
Step 2: Split into RECURRING segment
Step 3: Split into USAGE segment
Step 4: Split into ONE_TIME segment
Step 5: Show REFUND as subtraction
Step 6: Display final composition percentages
```

---

## Key Messages to Convey

1. **"Active MRR is YOUR safe money"**
   - Only counts healthy subscriptions
   - Excludes at-risk and churned

2. **"Revenue at Risk is your early warning"**
   - These stores might still save
   - Take action before they churn

3. **"Risk states are deterministic"**
   - Days past due → Risk state
   - No guesswork, clear thresholds

4. **"Deltas tell the story"**
   - Red/green isn't always obvious
   - Lower revenue-at-risk is GOOD (green)

5. **"Usage revenue scales with success"**
   - Not limited to orders
   - SMS, API, AI, storage, etc.

---

## File Locations
- **Prompt:** `docs/prompts/kpi-metrics-visualization.md`
- **Component:** `marketing/site/components/KPIMetricsGuide.tsx`
- **Page:** `marketing/site/app/kpi-guide/page.tsx`
- **View:** http://localhost:3000/kpi-guide

---

## Example API Response (Reference)

```json
{
  "period": {
    "start": "2026-02-01",
    "end": "2026-02-28"
  },
  "current": {
    "active_mrr_cents": 1245000,
    "revenue_at_risk_cents": 185000,
    "usage_revenue_cents": 350000,
    "total_revenue_cents": 1750000,
    "renewal_success_rate": 0.915,
    "safe_count": 612,
    "one_cycle_missed_count": 85,
    "two_cycles_missed_count": 42,
    "churned_count": 108
  },
  "previous": {
    "active_mrr_cents": 1183500,
    "revenue_at_risk_cents": 211000,
    "usage_revenue_cents": 318000,
    "total_revenue_cents": 1508000,
    "renewal_success_rate": 0.892,
    "safe_count": 578,
    "one_cycle_missed_count": 92,
    "two_cycles_missed_count": 48,
    "churned_count": 95
  },
  "delta": {
    "active_mrr_percent": 5.2,
    "revenue_at_risk_percent": -12.3,
    "usage_revenue_percent": 10.1,
    "total_revenue_percent": 16.0,
    "renewal_success_rate_percent": 2.58,
    "churn_count_percent": 13.7
  }
}
```

---

## Implementation Checklist

- [ ] Create page at `/kpi-guide`
- [ ] Build KPIMetricsGuide component
- [ ] Implement KPI card animations
- [ ] Implement risk classification visualization
- [ ] Implement period comparison animations
- [ ] Implement revenue breakdown visualization
- [ ] Add interactive toggles (time range, view mode)
- [ ] Add hover tooltips with formulas
- [ ] Add responsive design for mobile
- [ ] Test all animation sequences

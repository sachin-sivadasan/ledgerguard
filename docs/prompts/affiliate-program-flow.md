# Affiliate Program Flow - Interactive Visualization

## Context
You are a senior frontend + visualization engineer building an interactive animated guide showing common affiliate and referral program patterns used by SaaS companies and Shopify apps.

Build an educational visualization that helps developers understand:
1. How referral attribution works with cookie windows
2. Different commission models (one-time, recurring, hybrid)
3. Multi-tier affiliate structures
4. Affiliate lifecycle from signup to tier progression

---

## Design Philosophy

### Target Audience
SaaS founders and Shopify app developers who:
- Want to implement an affiliate/referral program
- Need to understand different commission structures
- Want to see real-world examples from successful companies
- Are evaluating build vs buy decisions

### Key Principles
1. **Show real patterns** - Based on actual programs (Shopify, ConvertKit, etc.)
2. **Compare models** - One-time vs recurring vs hybrid
3. **Interactive parameters** - Adjustable cookie windows, commission rates
4. **Practical guidance** - Implementation considerations

---

## Flow Types

### Flow 1: Attribution Flow (Cookie-based)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         REFERRAL ATTRIBUTION FLOW                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  🔗 Affiliate    →    👤 Visitor    →    🍪 Browser    →    💳 Customer     │
│  Shares Link          Clicks Link        Cookie Set        Signs Up         │
│                                                                              │
│  Cookie Window: [30 days ▾]                                                 │
│                                                                              │
│  Unique Link: yourapp.com/?ref=john123                                      │
│  Cookie: ref_id=john123, expires=30d                                        │
│  Attribution: Cookie read at checkout, affiliate credited                   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

Configurable windows: 30 / 60 / 90 / Lifetime
```

### Flow 2: Commission Calculation

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       COMMISSION CALCULATION FLOW                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ONE-TIME MODEL (25%)           RECURRING MODEL (20%)                       │
│  ────────────────────           ─────────────────────                       │
│  Customer pays: $49/mo          Customer pays: $49/mo                       │
│  Commission: $12.25 (once)      Commission: $9.80/mo (ongoing)              │
│  Total (12mo): $12.25           Total (12mo): $117.60                       │
│                                                                              │
│  HYBRID MODEL (15% + 10%)                                                   │
│  ────────────────────────                                                   │
│  First payment: 15% = $7.35                                                 │
│  Ongoing: 10% = $4.90/mo                                                    │
│  Total (12mo): $7.35 + $53.90 = $61.25                                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 3: Multi-Tier Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MULTI-TIER AFFILIATE STRUCTURE                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Tier 1: Direct Referral                                                    │
│  ────────────────────────                                                   │
│  Affiliate A → Customer X                                                   │
│  Commission: 20%                                                            │
│                                                                              │
│  Tier 2: Sub-Affiliate Override                                             │
│  ──────────────────────────────                                             │
│  Affiliate A recruits Affiliate B                                           │
│  Affiliate B → Customer Y                                                   │
│  Affiliate B gets: 20%                                                      │
│  Affiliate A gets: 5% override                                              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 4: Affiliate Lifecycle

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          AFFILIATE LIFECYCLE                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  📝 Apply    →    ✅ Approved    →    🔗 Share    →    💰 Earn    →    ⭐ Tier Up │
│                                                                              │
│  Tier Progression:                                                          │
│  Starter (10%) → Bronze (15%) → Silver (20%) → Gold (25%) → Platinum (30%)  │
│                                                                              │
│  Thresholds: $0 → $1,000 → $5,000 → $10,000 → $25,000 total revenue        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Real-World Examples

| Company | Commission | Type | Description |
|---------|-----------|------|-------------|
| Shopify | Up to $150/referral | One-time | Per merchant signup, bonus for Plus |
| ConvertKit | 30% recurring | Lifetime | Popular with creators |
| HubSpot | 20% recurring (1yr) | Partner | Tiered partner program |
| Webflow | 50% first payment | One-time | High initial commission |
| Stripe | $5/user for 1 year | Referral | Limited credits, not full affiliate |
| Teachable | 30% recurring | Lifetime | Course creator platform |

---

## Program Types

### Referral Links
- Unique tracking URLs: `?ref=john`, `/r/affiliate123`
- Simple implementation
- Works across all channels

### Coupon Codes
- Discount codes tied to affiliates: `JOHN20`, `SAVE15`
- Great for influencers
- Tracks offline conversions

### Partner Programs
- Formal agreements with agencies/consultants
- Higher commission tiers
- Often includes certifications

---

## Implementation Considerations

### Attribution Window
- 30-day: Industry standard
- 60-90 days: Higher affiliate attraction
- Last-click vs first-click attribution
- Cross-device tracking challenges

### Fraud Prevention
- Self-referral detection
- Minimum payout thresholds
- Chargeback/refund clawbacks
- IP/fingerprint analysis

### Payout Logistics
- PayPal, Stripe, wire transfer
- Monthly vs bi-weekly payouts
- Tax forms (W-9, W-8BEN)
- International payment handling

### Platform Options
- Build custom (full control)
- Rewardful, PartnerStack, FirstPromoter
- Shopify Collabs (for merchants)
- Impact, ShareASale (enterprise)

---

## Technical Requirements

### Framework
- Next.js 14+ with App Router
- TailwindCSS for styling
- React hooks for state and animation

### Visual Style
- Dark theme (slate-950 background)
- Green to blue gradient accents
- Card-based layouts
- Animated flow diagrams

### Interactions
- Flow type selector
- Configurable parameters (cookie window, commission rate)
- Play/pause animation
- Real-world example cards

---

## Component Structure

```
marketing/site/
├── app/affiliate-program/page.tsx    # Page wrapper with documentation
└── components/
    └── AffiliateFlowVisualization.tsx  # Main visualization
        ├── FlowSelector               # Tab buttons
        ├── AttributionFlow            # Cookie tracking flow
        ├── CommissionFlow             # Calculation comparison
        ├── MultiTierFlow              # Override structure
        ├── LifecycleFlow              # Tier progression
        └── RealWorldExamples          # Company cards
```

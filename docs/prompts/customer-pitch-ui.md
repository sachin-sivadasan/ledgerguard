# Customer Pitch UI - Implementation Prompt

## Overview

Create an interactive, scroll-based customer pitch page for LedgerGuard targeting Shopify app developers. The page tells a problem→solution story with animated visualizations and ends with a clear CTA.

---

## Target Audience

**Primary:** Shopify App Developers
- Revenue range: $10K–$100K MRR
- Teams who've passed initial traction
- Need serious visibility into renewals, churn, and expansion
- Currently using manual spreadsheets or generic SaaS tools

**Not targeting (v1):**
- Theme developers (one-off purchases)
- Service partners/agencies (project-based)

---

## Pain Points to Highlight

### 1. Shopify's Native Dashboard Gap
> "Shopify's Partner Dashboard shows historical earnings and payouts, but it doesn't give renewal-rate analytics, cohort retention, or churn-risk scores for app subscriptions."

**What Shopify provides:**
- Revenue over time
- App earnings and payouts
- Basic time-based performance metrics

**What's missing (our opportunity):**
- Cohort-based renewal/retention rates
- Predictive churn risk scoring
- LTV, segmentation, health scoring
- Proactive alerts and workflows

### 2. Current Workarounds
> "Today, Shopify app developers typically export CSV reports and maintain manual spreadsheets, or force-fit generic SaaS analytics tools that don't understand Shopify's app billing nuances."

- Export CSVs → Google Sheets / Excel
- Generic tools (Baremetrics, ChartMogul) with integration mismatches
- Usage charges, prorations, app credits not handled properly

### 3. Delayed Churn Detection
> "For most Shopify app teams under ~$100K MRR, churn is detected reactively—days to weeks after the fact, during monthly revenue reviews."

- No real-time alerting
- Discover churn when checking weekly/monthly revenue
- Don't know which accounts were responsible

---

## Competitive Positioning

### Our Differentiators

1. **Verticalized:** Built specifically for Shopify app billing, not generic Stripe/Braintree data
2. **Native Integration:** Plugs directly into Shopify Partner API
3. **Shopify-Aware:** Understands recurring charges, usage charges, app credits, prorations
4. **Opinionated:** Ships with Shopify app retention benchmarks ("what good looks like")

### vs. Generic SaaS Tools
| Feature | Generic Tools | LedgerGuard |
|---------|--------------|-------------|
| Shopify billing objects | Partial/manual | Native |
| Usage charges | Often missed | First-class |
| App credits/refunds | Manual reconciliation | Automatic |
| Shopify benchmarks | None | Built-in |

---

## Social Proof Strategy

### Metrics to Highlight (Beta Goals)
- "Reduced churn detection lag from 14 days to <24 hours"
- "Identified $X in at-risk MRR in the first week"
- "Helped apps increase 3-month retention by 8%"
- "Surfaced upsell opportunities for Y% expansion MRR growth"

### Case Study Structure (Future)
1. App profile: category, MRR range, team size
2. Before: manual exports, spreadsheets
3. Pain: missed churn, reactive support
4. After: metrics + founder quote + screenshots

### Placeholder for Now
Use realistic mock data with disclaimer: "Based on typical Shopify app metrics"

---

## Call-to-Action Flow

### Primary CTA
"Connect your Shopify Partner account" or "Start free trial"

### Secondary CTA
"Book a demo" (for cautious users)

### Early Access Mode
"Join the beta" / "Request early access"

### Trust Elements Before Connection
- Brief explanation of data accessed
- Link to Shopify Partner API docs
- Security/privacy statement

---

## Page Structure (Scroll-Based Story)

### Section 1: Hero
**Headline:** "Know Which Merchants Are About to Churn—Before They Do"

**Subheadline:** "Revenue intelligence built for Shopify app developers. Stop guessing, start retaining."

**Visual:** Animated dashboard preview showing:
- MRR ticker counting up
- Risk alert notification sliding in
- Cohort chart with retention curve

**CTA:** "Connect Partner Account" + "Watch Demo"

---

### Section 2: The Problem
**Headline:** "Your Partner Dashboard Shows What You Earned. Not Why It's Changing."

**Three-column pain points:**

1. **Blind to Churn**
   - Icon: Eye with slash
   - "Most app developers discover churn days or weeks late—during monthly revenue reviews"
   - Animated: Calendar flipping pages, then alert appearing

2. **Spreadsheet Hell**
   - Icon: Spreadsheet chaos
   - "Exporting CSVs, manual formulas, inconsistent tracking"
   - Animated: Spreadsheet cells multiplying chaotically

3. **Wrong Tools**
   - Icon: Puzzle piece that doesn't fit
   - "Generic SaaS analytics don't understand Shopify billing"
   - Animated: Square peg trying to fit round hole

---

### Section 3: The Solution
**Headline:** "Revenue Intelligence That Speaks Shopify"

**Interactive Dashboard Preview:**
Show a mini-dashboard with tabs:

1. **MRR Health**
   - Active MRR card (animated counter)
   - Revenue at Risk card (pulsing warning)
   - Churned this month

2. **Risk Radar**
   - List of merchants with risk scores
   - Color-coded: Safe (green), At Risk (yellow), Critical (red)
   - Hover to see "Days since last payment" and "Usage trend"

3. **Cohort Retention**
   - Animated retention curve by install month
   - Benchmark line showing "Top 10% Shopify apps"

---

### Section 4: How It Works
**Headline:** "From Partner API to Actionable Insights in Minutes"

**Animated Flow Diagram:**
```
[Shopify Partner Account]
        ↓ Connect (OAuth)
[LedgerGuard Sync Engine]
        ↓ Every 12 hours
[Transaction Ledger]
        ↓ Rebuild
[Risk Classification]
        ↓ Alerts
[Your Dashboard]
```

**Three Steps:**
1. **Connect** - "Link your Shopify Partner account (read-only access)"
2. **Sync** - "We pull 12 months of transaction history automatically"
3. **Act** - "See who's at risk and take action before they churn"

---

### Section 5: Features Grid
**Headline:** "Everything You Need to Protect Your MRR"

**Feature Cards (2x3 grid):**

| Feature | Description |
|---------|-------------|
| **Renewal Success Rate** | Track what % of merchants successfully renew each billing cycle |
| **Revenue at Risk** | See exactly how much MRR is tied to at-risk merchants |
| **Churn Prediction** | Know who's likely to cancel before they do (30/60/90 day risk) |
| **Usage Revenue** | Track usage-based billing separately from subscriptions |
| **Daily AI Brief** | Get a 100-word executive summary of your revenue health (Pro) |
| **Revenue API** | Query merchant payment status from your own app |

---

### Section 6: Comparison
**Headline:** "Why App Developers Switch to LedgerGuard"

**Before/After Split:**

| Before LedgerGuard | After LedgerGuard |
|-------------------|-------------------|
| Export CSVs weekly | Real-time dashboard |
| Notice churn 2 weeks late | Same-day alerts |
| "We think MRR is around..." | "MRR is $47,392 (+3.2%)" |
| Manual spreadsheet cohorts | Auto-generated retention curves |
| Generic SaaS tools | Shopify-native intelligence |

---

### Section 7: Social Proof
**Headline:** "Trusted by Growing Shopify Apps"

**If beta metrics available:**
- Three stat cards: "X% faster churn detection", "$Y at-risk MRR identified", "Z apps connected"

**If pre-beta:**
- "Join 50+ app developers on the waitlist"
- Logos placeholder or "Featured in Shopify Partner community"

**Testimonial Card (future):**
> "We used to find out about churned merchants a month later. Now we get alerts the same day and have saved $8K in MRR through proactive outreach."
> — *Founder, [App Name]*

---

### Section 8: Pricing Preview
**Headline:** "Start Free. Scale When Ready."

**Two Tiers:**

| Free | Pro ($49/mo) |
|------|-------------|
| 1 app | Unlimited apps |
| Core KPIs | All KPIs + trends |
| 7-day history | 12-month history |
| - | AI Daily Brief |
| - | Revenue API access |
| - | Slack alerts |

**CTA:** "Start Free" / "Go Pro"

---

### Section 9: Final CTA
**Headline:** "Stop Losing Merchants You Could Have Saved"

**Subheadline:** "Connect your Partner account in 60 seconds. No credit card required."

**Large CTA Button:** "Connect Shopify Partner Account"

**Secondary:** "Book a Demo" | "View Documentation"

**Trust badges:**
- "Read-only access"
- "SOC 2 compliant" (if applicable)
- "Your data stays yours"

---

## Visual Design Specifications

### Match Marketing Site
- Same typography, colors, icon style
- Dark/light mode consistent with main site
- Similar chart styling

### Animations
1. **Scroll-triggered:** Elements fade/slide in as user scrolls
2. **Dashboard preview:** Live-updating numbers, pulsing alerts
3. **Flow diagram:** Sequential highlighting as user scrolls
4. **Retention curve:** Draws itself when in viewport

### Charts to Include
1. **MRR Timeline:** Area chart with gradient fill
2. **Risk Distribution:** Horizontal stacked bar (Safe/At Risk/Critical/Churned)
3. **Cohort Retention:** Heatmap or line chart by month
4. **Revenue Mix:** Donut chart (Recurring/Usage/One-time)

### Interactive Elements
- Dashboard tabs (click to switch views)
- Merchant risk list (hover for details)
- Feature cards (expand on click for more info)

---

## Technical Implementation

### Route
`/pitch` or `/why-ledgerguard`

### Component Structure
```
marketing/site/
├── app/pitch/page.tsx
└── components/
    ├── PitchHero.tsx
    ├── ProblemSection.tsx
    ├── SolutionPreview.tsx
    ├── HowItWorks.tsx
    ├── FeaturesGrid.tsx
    ├── ComparisonTable.tsx
    ├── SocialProof.tsx
    ├── PricingPreview.tsx
    └── FinalCTA.tsx
```

### Animation Library
- Framer Motion for scroll animations
- Or CSS animations with Intersection Observer

### Responsive
- Mobile-first
- Dashboard preview simplified on mobile
- Feature grid collapses to single column

---

## Content Tone

### Voice
- Confident but not arrogant
- Technical but accessible
- Empathetic to the pain ("We've been there")

### Avoid
- Jargon overload
- Vague promises ("revolutionary", "game-changing")
- Overly salesy language

### Emphasize
- Specificity ("14 days → 24 hours")
- Shopify-native understanding
- Time savings and peace of mind

---

## Success Metrics

Track after launch:
1. Scroll depth (how far do visitors get?)
2. CTA click rate (Connect vs Demo vs Waitlist)
3. Time on page
4. Conversion to signup/connection

---

## Implementation Phases

### Phase 1: Core Page
- Hero + Problem + Solution sections
- Basic scroll animations
- Primary CTA

### Phase 2: Interactive Elements
- Dashboard preview with tabs
- Animated flow diagram
- Merchant risk list hover states

### Phase 3: Polish
- Testimonials (when available)
- Video embed option
- A/B test different headlines

---

## Notes

- This page should work standalone (direct link from ads, social, etc.)
- Also linked from main marketing site navigation
- Consider a shorter "elevator pitch" version for homepage hero

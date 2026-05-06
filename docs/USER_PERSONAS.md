# LedgerGuard — User Personas & Feature Map

> Revenue Intelligence Platform for Shopify App Developers

## Why This Matters

Every Shopify app developer has different needs. Some want a full dashboard, others just need an API call, and some only want a Slack ping when something goes wrong. This document maps **15 distinct user personas** to LedgerGuard features, ensuring the platform delivers value across every segment — and justifies charging for it.

---

## Personas Overview

| # | Persona | Description | Likely Tier | Primary Value |
|---|---------|-------------|-------------|---------------|
| 1 | **Embedded-Only Dev** | No website or external dashboard. Built a Shopify embedded app and needs LedgerGuard as their sole analytics dashboard. | Starter | Full dashboard replacement — MRR, installs, subscriptions, risk, earnings |
| 2 | **Non-Embedded Dev** | Has their own app with a dashboard, but needs subscription intelligence. Wants to know which shops are paying, who's churning, and risk state changes. | Starter | Subscription intelligence + risk alerts on top of their existing stack |
| 3 | **API-First Dev** | Uses the Revenue API (API key) to gate features in their app based on subscription status. Detects cycle-missed stores and enforces access control programmatically. | Starter/Pro | Real-time subscription gating — one API call decides feature access |
| 4 | **Notification-Only** | Doesn't need a dashboard at all. Just wants Slack or email alerts when something important happens — churn, billing failure, risk escalation. | Starter | Zero-effort monitoring — set up once, get alerts forever |
| 5 | **Mobile-First Dev** | Primarily uses the LedgerGuard mobile app. Wants push notifications for events and a quick glance at MRR and installs while on the go. | Starter | On-the-go monitoring without opening a laptop |
| 6 | **AI Power User** | Wants AI-generated insights: daily revenue briefs, natural language queries ("what's my churn rate this month?"), forecasting, cohort analysis. | Pro | AI does the analysis — no manual number crunching |
| 7 | **Agency / Multi-App** | Manages 5+ Shopify apps (own or clients'). Needs portfolio-wide metrics, aggregate MRR, per-app breakdowns, and bulk notifications. | Pro/Enterprise | Single pane of glass for an entire app portfolio |
| 8 | **Finance / Compliance** | Cares about audit trails, revenue reconciliation, Shopify fee verification, and historical snapshots. Needs to match Shopify payouts against their books. | Pro | Immutable snapshots + fee breakdown = reconciliation confidence |
| 9 | **Marketplace Veteran** | Has been on the Shopify App Store for years. Obsesses over reviews, ratings, install trends, and reputation management. | Starter | Review scraping + install velocity = marketplace pulse |
| 10 | **Freemium App Dev** | Offers a free app with optional paid upgrades. Needs to track free-to-paid conversion rates, trial expiry, and which stores upgrade. | Starter | Conversion funnel visibility — know exactly where users drop off |
| 11 | **Investor / Acquirer** | Evaluating a Shopify app for acquisition or investment. Needs due diligence data: MRR trends, churn rates, revenue concentration, growth trajectory. | Pro | Export-ready reports for investment decisions |
| 12 | **Customer Success Mgr** | Works at a larger app company. Manages merchant relationships and needs per-store risk alerts to proactively reach out before churn happens. | Starter/Pro | Proactive outreach triggers — save accounts before they leave |
| 13 | **Support Team Lead** | Needs to quickly check if a support ticket sender is a paying subscriber before prioritizing. Uses domain-based lookup via API or dashboard. | Starter | Instant subscriber verification — prioritize paying customers |
| 14 | **Side-Project Dev** | Runs a Shopify app as a side hustle. Checks it once a week. Needs a weekly email digest with key changes, zero daily engagement required. | Free/Starter | Weekly digest — stay informed with zero effort |
| 15 | **Growth / Marketing** | Focused on growing installs and improving app store presence. Wants install velocity trends, review sentiment over time, and campaign correlation. | Starter/Pro | Install analytics + review sentiment = data-driven growth |

---

## Feature Coverage Matrix

Shows which features each persona uses. Heavy use = primary driver, Light use = secondary benefit.

| Feature | P1 | P2 | P3 | P4 | P5 | P6 | P7 | P8 | P9 | P10 | P11 | P12 | P13 | P14 | P15 |
|---------|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Dashboard & KPIs** | | | | | | | | | | | | | | | |
| MRR / ARR / churn rate | XX | | | | | XX | XX | | XX | XX | XX | | | XX | |
| Install count & trends | XX | | | | | | | | XX | XX | | | | XX | XX |
| Subscription list | XX | XX | | | | | | | | XX | | XX | XX | | |
| Multi-app aggregate | | | | | | | XX | | | | | | | | |
| Dashboard widgets | | | | | | | XX | | | | | | | | |
| Store health per shop | XX | | | | | | | | | | | XX | XX | | |
| **Revenue & Earnings** | | | | | | | | | | | | | | | |
| Earnings timeline | XX | | | | | | XX | XX | | | | | | | |
| Fee breakdown | | | | | | | XX | XX | | | | | | | |
| Daily snapshots | | XX | | | | | | XX | | | | | | | |
| Revenue by charge type | | | | | | XX | | XX | | | | | | | |
| **Risk Engine** | | | | | | | | | | | | | | | |
| Risk classification | XX | XX | | | | | | | | | | | | | |
| At-risk alerts | | XX | | XX | | | | | | | | XX | | | |
| Churn detection | | XX | | XX | | | | | | XX | | XX | | | |
| **Revenue API** | | | | | | | | | | | | | | | |
| Subscription by GID/domain | | | XX | | | | | | | | | XX | XX | | |
| Batch lookup | | | XX | | | | | | | | | | | | |
| Usage charge status | | | XX | | | | | | | | | | | | |
| GraphQL endpoint | | | XX | | | | | | | | | | | | |
| API key management | | | XX | | | | | | | | | | | | |
| **Notifications** | | | | | | | | | | | | | | | |
| Slack alerts | | XX | | XX | | | XX | | | | | XX | | | |
| Email daily summary | | | | XX | | | XX | | | | | | | | |
| Email weekly digest | | | | | | | | | | | | | | XX | |
| Push notifications (FCM) | | | | | XX | | | | XX | | | | | XX | |
| Alert preferences | | | | XX | | | | | | | | | | | |
| **Mobile App** | | | | | | | | | | | | | | | |
| Mobile dashboard | | | | | XX | | | | | | | | | | |
| Real-time push | | | | | XX | | | | | | | | | | |
| Quick sub lookup | | | | | XX | | | | | | | XX | | | |
| **AI & Analytics** | | | | | | | | | | | | | | | |
| AI Daily Brief | | | | | | XX | | | | | | | | | |
| NL chat | | | | | | XX | | | | | | | | | |
| Churn prediction | | | | | | XX | | | | | XX | | | | |
| Cohort analysis | | | | | | XX | | | | | XX | | | | |
| **Conversion & Growth** | | | | | | | | | | | | | | | |
| Free-to-paid conversion | | | | | | | | | | XX | | | | | XX |
| Install velocity | | | | | | | | | | | | | | | XX |
| Trial expiry tracking | | | | | | | | | | XX | | | | | |
| **Reporting & Export** | | | | | | | | | | | | | | | |
| PDF / CSV export | | | | | | | XX | XX | | | XX | | | | XX |
| Revenue concentration | | | | | | | | | | | XX | | | | |
| 12-month history | | | | | | | | XX | | | XX | | | | |
| **Sync Engine** | | | | | | | | | | | | | | | |
| Full sync | XX | | | | | | | | | | | | | | |
| Daily catchup | | | | | | | | XX | | | | | | | |
| Webhook updates | XX | XX | XX | | | | | | | | | | | | |
| Queue management | | | | | | | XX | | | | | | | | |
| **Reviews** | | | | | | | | | | | | | | | |
| Review scraping | | | | | | | | | XX | | | | | | XX |
| Sentiment tracking | | | | | | | | | XX | | | | | | XX |

`XX` = persona actively uses this feature

---

## Pricing Tier Mapping

### Free (USD $0/month)
- 1 app, 30-day history
- Basic MRR/install metrics
- Email weekly digest
- **Target personas:** Side-Project Dev (P14)

### Starter (USD $249/month)
- 1 app, 90-day history
- Full dashboard, subscriptions, risk engine
- Email + push notifications
- Revenue API access (60 req/min)
- Review scraping
- **Target personas:** Embedded-Only (P1), Non-Embedded (P2), Notification-Only (P4), Mobile-First (P5), Marketplace Veteran (P9), Freemium Dev (P10), Support Lead (P13)

### Pro (USD $499/month)
- Unlimited apps, 12-month history
- AI Daily Brief + NL chat
- Churn prediction + cohort analysis
- Slack integration
- PDF/CSV export
- Priority support
- **Target personas:** API-First (P3), AI Power User (P6), Agency (P7), Finance (P8), Investor (P11), CS Manager (P12), Growth/Marketing (P15)

### Enterprise (Custom pricing)
- Everything in Pro
- Dedicated account manager
- Custom integrations
- SLA guarantees
- SSO / role-based access
- **Target personas:** Large agencies (P7 at scale)

---

## Feature Gap Analysis

Features **not yet built** but needed to serve all 15 personas:

| Feature | Needed By | Status | Priority |
|---------|-----------|--------|----------|
| Free-to-paid conversion funnel | P10, P15 | Not built | High — unlocks Freemium persona |
| Weekly email digest | P14 | Not built | Medium — serves side-project devs |
| PDF/CSV export | P7, P8, P11, P15 | Not built | High — unlocks Investor + Finance |
| Revenue concentration analysis | P11 | Not built | Medium — investment due diligence |
| Install velocity / campaign correlation | P15 | Not built | Medium — Growth persona |
| Trial expiry tracking | P10 | Not built | Medium — Freemium persona |
| Cohort analysis & retention curves | P6, P11 | Not built | Low — AI roadmap |
| Churn prediction model | P6, P11 | Not built | Low — AI roadmap |

---

## Diagrams

| Diagram | File | Description |
|---------|------|-------------|
| Overview Mindmap | [`USER_PERSONAS.puml`](diagrams/puml/USER_PERSONAS.puml) | All 15 personas grouped by pricing tier |
| Group 1: Dashboard | [`USER_PERSONAS_1_dashboard.puml`](diagrams/puml/USER_PERSONAS_1_dashboard.puml) | Embedded, Non-Embedded, Mobile, Side-Project, Marketplace |
| Group 2: API & Ops | [`USER_PERSONAS_2_api_ops.puml`](diagrams/puml/USER_PERSONAS_2_api_ops.puml) | API-First, Notification-Only, CS Manager, Support, Freemium |
| Group 3: Power Users | [`USER_PERSONAS_3_power.puml`](diagrams/puml/USER_PERSONAS_3_power.puml) | AI, Agency, Finance, Investor, Growth/Marketing |

Render all: `plantuml docs/diagrams/puml/USER_PERSONAS*.puml`

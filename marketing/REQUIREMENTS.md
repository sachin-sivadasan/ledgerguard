# Marketing Site Requirements – LedgerGuard

## Overview
Public-facing marketing site for LedgerGuard, a Revenue Intelligence Platform for Shopify App Developers.

## Tech Stack
- **Framework:** Next.js 14+ (App Router)
- **Styling:** TailwindCSS
- **Design:** Minimal, professional, responsive
- **Authentication:** None (public landing page)

---

## Page Structure

### 1. Hero Section
**Purpose:** Capture attention, communicate value proposition

**Content:**
- Headline: "Stop Guessing. Start Knowing."
- Subheadline: "Revenue intelligence for Shopify app developers. Track renewals, predict churn, and protect your MRR."
- Primary CTA: "Connect Shopify Partner" (button)
- Secondary CTA: "See How It Works" (link)

---

### 2. Problem Statement
**Purpose:** Resonate with pain points

**Content:**
- Headline: "Flying Blind on Revenue?"
- Pain points (3 cards):
  1. **Spreadsheet Hell** – "Manually tracking subscriptions across multiple apps is error-prone and time-consuming."
  2. **Surprise Churn** – "You only discover a customer churned when the payment fails—too late to save them."
  3. **No Single Source of Truth** – "Partner dashboard data doesn't match your records. Which one is right?"

---

### 3. Renewal Success Rate
**Purpose:** Explain key metric

**Content:**
- Headline: "Know Your Renewal Success Rate"
- Description: "See the percentage of subscriptions that successfully renewed vs. those at risk or churned. One number that tells you if your business is healthy."
- Visual: Simple metric card mockup showing "94.2% Renewal Rate"

---

### 4. Revenue at Risk
**Purpose:** Explain key metric

**Content:**
- Headline: "See Revenue at Risk Before It's Gone"
- Description: "Identify subscriptions that missed payment cycles. Reach out before they churn. Every dollar you save goes straight to your bottom line."
- Visual: Simple metric card showing "$2,450 at risk" with breakdown

---

### 5. AI Daily Revenue Brief
**Purpose:** Highlight Pro feature

**Content:**
- Headline: "Your Daily Revenue Brief, Powered by AI"
- Description: "Every morning, get an 80-word executive summary of your revenue health. What changed, what needs attention, and what's trending—delivered to your inbox or Slack."
- Badge: "Pro Feature"
- Example brief in a card/quote style

---

### 6. Pricing Tiers
**Purpose:** Convert visitors

**Tiers:**

#### Starter Tier
- Price: $149/month
- Features:
  - 1 Shopify app
  - Renewal Success Rate
  - Revenue at Risk alerts
  - 90-day transaction history
  - Email notifications
- CTA: "Get Started"

#### Pro Tier
- Price: $299/month
- Features:
  - Unlimited apps
  - AI Daily Revenue Brief
  - 12-month transaction history
  - Slack integration
  - Priority support
- CTA: "Start Free Trial"

---

### 7. Final CTA Section
**Purpose:** Drive conversion

**Content:**
- Headline: "Ready to Protect Your Revenue?"
- Subheadline: "Connect your Shopify Partner account in 60 seconds. No credit card required."
- CTA Button: "Connect Shopify Partner"

---

## Design Guidelines

### Colors
- **Primary:** Blue (#2563eb) – trust, professionalism
- **Secondary:** Slate (#475569) – text, subtle elements
- **Accent:** Green (#10b981) – success, positive metrics
- **Warning:** Amber (#f59e0b) – at-risk indicators
- **Danger:** Red (#ef4444) – churned, critical alerts
- **Background:** White (#ffffff) / Slate-50 (#f8fafc)

### Typography
- **Headings:** Inter or system font, bold
- **Body:** Inter or system font, regular
- **Sizes:** Mobile-first, responsive scaling

### Spacing
- Consistent section padding (py-16 to py-24)
- Container max-width (max-w-6xl or max-w-7xl)
- Generous whitespace

### Components
- Cards with subtle shadows and rounded corners
- Buttons with hover states
- Simple iconography (Heroicons or similar)

---

## Responsive Breakpoints
- Mobile: < 640px (default)
- Tablet: 640px - 1024px (sm:, md:)
- Desktop: > 1024px (lg:, xl:)

---

## SEO Considerations
- Semantic HTML (h1, h2, sections)
- Meta title: "LedgerGuard – Revenue Intelligence for Shopify App Developers"
- Meta description: "Track renewals, predict churn, and protect your MRR. Connect your Shopify Partner account and get insights in minutes."

---

## No Authentication
This is a public marketing site. Authentication happens in the main app (separate frontend).

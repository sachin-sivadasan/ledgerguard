# Billing System Flow - Interactive Visualization

## Context
You are a senior frontend + visualization engineer building an interactive animated guide showing how LedgerGuard's billing and subscription system works. The system uses Stripe for payment processing with a Trial-Freemium model.

Build an educational visualization that helps admins and developers understand:
1. How the trial-to-freemium-to-paid lifecycle works
2. How Stripe webhooks keep billing state in sync
3. How feature gating enforces plan limits
4. How plan upgrades/downgrades are handled

---

## Design Philosophy

### Target Audience
LedgerGuard admins and developers who:
- Need to understand the billing lifecycle
- Want to configure plans and features
- Need to debug billing state issues
- Want to see how Stripe integration works end-to-end

### Key Principles
1. **Show the lifecycle** - Trial → Free → Pro → Enterprise transitions
2. **Stripe-first** - All payment logic delegated to Stripe
3. **Feature gating** - Clear visualization of what's locked/unlocked per plan
4. **Webhook-driven** - Backend reacts to Stripe events, never polls

---

## Why Stripe (Not Shopify Billing)

```
+-----------------------------------------------------------------------+
|                    PAYMENT GATEWAY DECISION                           |
+===============+======================+================================+
|               | Stripe [chosen]      | Shopify Billing                |
+===============+======================+================================+
| Target        | Any customer         | Merchants with shops only      |
| Customer      | (partners, devs,     | (requires app installed        |
|               | businesses)          | in a Shopify store)            |
+---------------+----------------------+--------------------------------+
| Billing       | Stripe invoices      | Shopify merchant invoice       |
| Entity        | (your control)       | (Shopify's control)            |
+---------------+----------------------+--------------------------------+
| Partner       | Yes                  | No - Partner API has no        |
| Billing       |                      | billing endpoints              |
+---------------+----------------------+--------------------------------+
| Manual Token  | Yes (all users       | No - requires OAuth +          |
| Users         | can pay)             | app installation               |
+---------------+----------------------+--------------------------------+
| Marketplace   | N/A (direct)         | No Partner Tools               |
|               |                      | marketplace exists             |
+---------------+----------------------+--------------------------------+
| Subscription  | Full control         | Shopify controls lifecycle     |
| Management    | (trials, proration,  | (limited customization)        |
|               | coupons, invoices)   |                                |
+---------------+----------------------+--------------------------------+
| Revenue Share | 0% (Stripe fees      | 0-20% Shopify rev share        |
|               | only: ~2.9% + 30c)   | + Stripe fees                  |
+---------------+----------------------+--------------------------------+
| Industry      | B2B SaaS standard    | Merchant app standard          |
| Standard      | (Baremetrics, Ship,  | (NOT for partner tools)        |
|               | all partner tools)   |                                |
+---------------+----------------------+--------------------------------+

Decision: Stripe
- LedgerGuard is a partner-facing SaaS, NOT a merchant-installed Shopify app
- Shopify Billing API requires app installed in a shop (merchant context)
- Partner API has NO billing endpoints
- No Partner Tools marketplace exists
- Manual token users have no Shopify OAuth — can't use Shopify Billing
- All comparable tools (Baremetrics, Ship) use external billing
```

---

## Plan Tiers

```
+-----------------------------------------------------------------------+
|                         PLAN TIERS                                    |
+==========+================+================+==========================+
|          | FREE           | PRO            | ENTERPRISE               |
+==========+================+================+==========================+
| Price    | $0/mo          | $XX/mo         | Custom                   |
|          | (after trial)  | (monthly/annual)| (contact sales)         |
+----------+----------------+----------------+--------------------------+
| Trial    | 14-day full    | --             | --                       |
|          | access on      |                |                          |
|          | signup         |                |                          |
+----------+----------------+----------------+--------------------------+
| Apps     | 1              | Unlimited      | Unlimited                |
+----------+----------------+----------------+--------------------------+
| Dashboard| Yes            | Yes            | Yes                      |
+----------+----------------+----------------+--------------------------+
| Risk     | Basic          | Full           | Full + Custom Rules      |
| Analytics| (overview)     | (breakdown,    |                          |
|          |                | timeline)      |                          |
+----------+----------------+----------------+--------------------------+
| AI Chat  | No             | Yes            | Yes + Priority           |
+----------+----------------+----------------+--------------------------+
| API Keys | No             | Yes            | Yes + Higher Limits      |
+----------+----------------+----------------+--------------------------+
| Slack    | No             | Yes            | Yes                      |
+----------+----------------+----------------+--------------------------+
| Export   | No             | CSV/PDF        | CSV/PDF + API            |
+----------+----------------+----------------+--------------------------+
| Notifs   | Push only      | Push + Slack   | Push + Slack + Email     |
+----------+----------------+----------------+--------------------------+
| Support  | Community      | Email          | Dedicated + SLA          |
+----------+----------------+----------------+--------------------------+
```

---

## Flow Types

### Flow 1: Signup -> Trial -> Conversion

```
+-----------------------------------------------------------------------+
|                    TRIAL-TO-PAID LIFECYCLE                             |
+-----------------------------------------------------------------------+
|                                                                       |
|  +----------+     +-------------+     +----------+                    |
|  | Firebase  |---->|   Backend   |---->|  Stripe  |                   |
|  | Signup    |     | Create User |     | Customer |                   |
|  +----------+     +------+------+     +----+-----+                   |
|                          |                  |                         |
|                          v                  v                         |
|                   +--------------+   +--------------+                 |
|                   | Set plan_tier|   | Create Stripe|                 |
|                   | = TRIAL      |   | Customer obj |                 |
|                   | trial_ends_at|   | (no payment  |                 |
|                   | = NOW + 14d  |   |  method yet) |                 |
|                   +--------------+   +--------------+                 |
|                                                                       |
|  Day 1-14: TRIAL (all features unlocked)                              |
|  +-----------------------------------------------------------------+  |
|  | User sees: "14 days left in trial" banner                       |  |
|  | Day 7: Email reminder "7 days left, add payment method"         |  |
|  | Day 12: Email reminder "2 days left"                            |  |
|  | Day 13: In-app warning "Trial ends tomorrow"                    |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|  Day 14: Trial Expires                                                |
|  +-----------------------------------------------------------------+  |
|  | Option A: User added payment -> Stripe subscription created     |  |
|  |           plan_tier = PRO, billing starts                       |  |
|  |                                                                 |  |
|  | Option B: User didn't pay -> plan_tier = FREE                   |  |
|  |           Gated features locked, upgrade prompts shown          |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
+-----------------------------------------------------------------------+
```

### Flow 2: Stripe Integration (Webhook-Driven)

```
+-----------------------------------------------------------------------+
|                    STRIPE WEBHOOK FLOW                                 |
+-----------------------------------------------------------------------+
|                                                                       |
|  +----------+     +----------------+     +----------+                 |
|  |  Stripe  |---->|  POST          |---->|  Backend |                 |
|  |  Event   |     | /webhooks/     |     |  Handler |                 |
|  |          |     |  stripe        |     |          |                 |
|  +----------+     +----------------+     +----+-----+                 |
|                                               |                       |
|                                               v                       |
|                                    +---------------------+            |
|                                    | Update billing_     |            |
|                                    | subscriptions table |            |
|                                    | + user.plan_tier    |            |
|                                    +---------------------+            |
|                                                                       |
|  Events Handled:                                                      |
|  +---------------------------------------------------------------+   |
|  | Event                          | Action                        |   |
|  |--------------------------------+-------------------------------|   |
|  | checkout.session.completed     | Create billing_subscription,  |   |
|  |                                | set plan_tier = PRO           |   |
|  | invoice.paid                   | Renew subscription period     |   |
|  | invoice.payment_failed         | Mark as past_due, send alert  |   |
|  | customer.subscription.updated  | Handle plan change            |   |
|  | customer.subscription.deleted  | Downgrade to FREE             |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Security:                                                            |
|  - Verify Stripe webhook signature (STRIPE_WEBHOOK_SECRET)            |
|  - Idempotent processing (event ID deduplication)                     |
|  - Return 200 immediately, process async                              |
|                                                                       |
+-----------------------------------------------------------------------+
```

### Flow 3: Feature Gating

```
+-----------------------------------------------------------------------+
|                    FEATURE GATING FLOW                                 |
+-----------------------------------------------------------------------+
|                                                                       |
|  Request arrives at protected endpoint (e.g. /api/v1/chat)            |
|         |                                                             |
|         v                                                             |
|  +------------------+                                                 |
|  | Auth Middleware   |  (existing: verify Firebase token)              |
|  +--------+---------+                                                 |
|           |                                                           |
|           v                                                           |
|  +------------------+                                                 |
|  | Plan Middleware   |  (NEW: check feature access)                   |
|  +--------+---------+                                                 |
|           |                                                           |
|           v                                                           |
|  +-----------------------------+                                      |
|  | 1. Get user.plan_tier       |                                      |
|  | 2. Check trial_ends_at      |                                      |
|  |    (if TRIAL, check expiry) |                                      |
|  | 3. Lookup plan_features     |                                      |
|  |    for required feature     |                                      |
|  | 4. If allowed -> continue   |                                      |
|  |    If denied -> 403 +       |                                      |
|  |    upgrade_required response |                                      |
|  +-----------------------------+                                      |
|                                                                       |
|  Frontend Gating:                                                     |
|  +---------------------------------------------------------------+   |
|  | Feature     | FREE          | TRIAL/PRO      | ENTERPRISE     |   |
|  |-------------+---------------+----------------+----------------|   |
|  | AI Chat     | Lock icon +   | Full access    | Full access    |   |
|  |             | "Upgrade to   |                | + priority     |   |
|  |             | PRO" button   |                |                |   |
|  | API Keys    | Hidden        | Full access    | Higher limits  |   |
|  | Slack       | Hidden        | Full access    | Full access    |   |
|  | Export      | Hidden        | CSV/PDF        | CSV/PDF + API  |   |
|  | Multi-app   | "1 app" badge | Unlimited      | Unlimited      |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
+-----------------------------------------------------------------------+
```

### Flow 4: Plan Upgrade / Downgrade

```
+-----------------------------------------------------------------------+
|                    PLAN CHANGE FLOW                                    |
+-----------------------------------------------------------------------+
|                                                                       |
|  UPGRADE (FREE -> PRO):                                               |
|  1. User clicks "Upgrade" in app                                      |
|  2. Frontend creates Stripe Checkout Session                          |
|     POST /api/v1/billing/checkout                                     |
|     { plan_id: "pro_monthly", success_url, cancel_url }               |
|  3. Backend creates Stripe Checkout Session                           |
|  4. User redirected to Stripe Checkout (hosted page)                  |
|  5. User enters payment method + confirms                             |
|  6. Stripe fires checkout.session.completed webhook                   |
|  7. Backend updates billing_subscription + plan_tier = PRO            |
|  8. User redirected to success_url, sees PRO features unlocked        |
|                                                                       |
|  DOWNGRADE (PRO -> FREE):                                             |
|  1. User clicks "Cancel Subscription" in settings                     |
|  2. Frontend calls POST /api/v1/billing/cancel                        |
|  3. Backend calls Stripe: cancel at period end                        |
|  4. User keeps PRO until current period ends                          |
|  5. At period end: Stripe fires customer.subscription.deleted          |
|  6. Backend sets plan_tier = FREE                                     |
|  7. Gated features locked on next request                             |
|                                                                       |
|  PLAN CHANGE (PRO Monthly -> PRO Annual):                             |
|  1. User selects annual plan in billing settings                      |
|  2. Backend calls Stripe: update subscription with proration          |
|  3. Stripe calculates credit/charge and invoices immediately          |
|  4. Webhook confirms update                                           |
|                                                                       |
+-----------------------------------------------------------------------+
```

---

## Data Model

### plans table
```sql
CREATE TABLE plans (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,  -- 'free', 'pro_monthly', 'pro_annual', 'enterprise'
    display_name VARCHAR(100) NOT NULL,        -- 'Free', 'Pro (Monthly)', 'Enterprise'
    tier        VARCHAR(20) NOT NULL,          -- 'FREE', 'PRO', 'ENTERPRISE'
    price_cents INTEGER NOT NULL DEFAULT 0,    -- 2900 = $29.00
    interval    VARCHAR(20) NOT NULL,          -- 'month', 'year', 'custom'
    stripe_price_id VARCHAR(100),              -- Stripe Price ID (e.g. price_xxx)
    trial_days  INTEGER NOT NULL DEFAULT 0,    -- 14 for initial trial
    max_apps    INTEGER NOT NULL DEFAULT 1,    -- 1 for free, -1 for unlimited
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed data
INSERT INTO plans (name, display_name, tier, price_cents, interval, trial_days, max_apps) VALUES
('free',         'Free',            'FREE',       0,    'month', 0,  1),
('pro_monthly',  'Pro (Monthly)',   'PRO',        2900, 'month', 14, -1),
('pro_annual',   'Pro (Annual)',    'PRO',        24900,'year',  14, -1),
('enterprise',   'Enterprise',     'ENTERPRISE', 0,    'custom',0,  -1);
```

### plan_features table
```sql
CREATE TABLE plan_features (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id     UUID NOT NULL REFERENCES plans(id),
    feature_key VARCHAR(50) NOT NULL,   -- 'ai_chat', 'api_keys', 'slack', 'export', 'multi_app'
    enabled     BOOLEAN NOT NULL DEFAULT false,
    limit_value INTEGER,                -- NULL = unlimited, e.g. api_keys limit = 5
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(plan_id, feature_key)
);

-- Feature keys:
-- ai_chat, api_keys, slack_notifications, data_export, multi_app,
-- advanced_risk, custom_risk_rules, priority_support, sla_support
```

### billing_subscriptions table
```sql
CREATE TABLE billing_subscriptions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID NOT NULL REFERENCES users(id),
    plan_id               UUID NOT NULL REFERENCES plans(id),
    stripe_customer_id    VARCHAR(100) NOT NULL,
    stripe_subscription_id VARCHAR(100),         -- NULL during trial/free
    status                VARCHAR(20) NOT NULL,  -- 'trialing', 'active', 'past_due', 'canceled', 'free'
    current_period_start  TIMESTAMPTZ,
    current_period_end    TIMESTAMPTZ,
    trial_ends_at         TIMESTAMPTZ,
    canceled_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### billing_events table (audit log)
```sql
CREATE TABLE billing_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id),
    stripe_event_id   VARCHAR(100) NOT NULL UNIQUE,  -- idempotency
    event_type        VARCHAR(50) NOT NULL,           -- 'checkout.session.completed', etc.
    payload           JSONB NOT NULL,
    processed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/billing/plans` | Public | List available plans + features |
| GET | `/api/v1/billing/subscription` | Firebase | Get user's current subscription |
| POST | `/api/v1/billing/checkout` | Firebase | Create Stripe Checkout Session |
| POST | `/api/v1/billing/portal` | Firebase | Create Stripe Customer Portal link |
| POST | `/api/v1/billing/cancel` | Firebase | Cancel subscription at period end |
| POST | `/webhooks/stripe` | Stripe Sig | Handle Stripe webhook events |

---

## Stripe Webhook Events

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Create billing_subscription, upgrade plan_tier |
| `invoice.paid` | Renew subscription, update period dates |
| `invoice.payment_failed` | Set status=past_due, send alert, retry |
| `customer.subscription.updated` | Handle plan changes, proration |
| `customer.subscription.deleted` | Downgrade to FREE, lock features |
| `customer.subscription.trial_will_end` | Send "trial ending" reminder (3 days before) |

---

## Email Triggers (via n8n + Postmark)

| Trigger | Email | Timing |
|---------|-------|--------|
| Trial started | "Welcome! Your 14-day trial has started" | Immediate |
| Trial day 7 | "7 days left — add payment to keep PRO features" | Day 7 |
| Trial day 12 | "2 days left — don't lose access to AI Chat" | Day 12 |
| Trial expired (no payment) | "Your trial ended — upgrade anytime" | Day 14 |
| Trial converted to PRO | "Welcome to PRO!" | On payment |
| Payment failed | "Payment failed — update your card" | On failure |
| Subscription canceled | "We're sorry to see you go" | On cancel |

---

## Configuration

### Environment Variables
```env
# Stripe
STRIPE_SECRET_KEY=sk_live_xxx
STRIPE_PUBLISHABLE_KEY=pk_live_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# Stripe Price IDs (created in Stripe Dashboard)
STRIPE_PRICE_PRO_MONTHLY=price_xxx
STRIPE_PRICE_PRO_ANNUAL=price_xxx
```

---

## Technical Requirements

### Framework (Marketing Visualization)
- Next.js 14+ with App Router
- TailwindCSS for styling
- React hooks for state and animation

### Animation Approach
- SVG-based flow diagrams with animated paths
- Step-by-step lifecycle progression
- Play/pause controls
- Tab selector for flow types

### Visual Style
- Dark theme (slate-950 background)
- Gradient accents (green for active, amber for trial, grey for free)
- Color-coded plan tiers
- Stripe brand purple for payment flows

### Interactions
- Flow type selector (5 tabs: Lifecycle, Stripe, Feature Gate, Upgrade, Data Model)
- Plan comparison table with hover highlights
- Animated webhook event processing

---

## Component Structure

```
marketing/site/
+-- app/billing-flow/page.tsx            # Page wrapper
+-- components/
    +-- BillingFlowVisualization.tsx      # Main visualization
        +-- FlowSelector                  # Tab buttons
        +-- LifecycleFlow                 # Trial -> Free -> Pro
        +-- StripeWebhookFlow            # Webhook event processing
        +-- FeatureGateFlow              # Plan-based access control
        +-- UpgradeFlow                  # Checkout + cancel flows
        +-- DataModelDiagram             # ER diagram of billing tables
        +-- PlanComparison               # Feature comparison table
```

---

## Implementation Notes

1. **Stripe Checkout (hosted):** Use Stripe's hosted checkout page — no PCI compliance needed
2. **Customer Portal:** Use Stripe's hosted portal for plan management, payment method updates
3. **Webhook idempotency:** Store `stripe_event_id` in `billing_events`, skip duplicates
4. **Trial clock:** Backend cron checks `trial_ends_at` daily, downgrades expired trials
5. **Feature check middleware:** `PlanMiddleware("ai_chat")` wraps protected endpoints
6. **Frontend plan context:** `BillingBloc` fetches subscription state, provides to all widgets
7. **Proration:** Stripe handles proration math — backend just updates subscription
8. **Naming:** `billing_subscriptions` (not `subscriptions`) to avoid collision with Shopify subscription data

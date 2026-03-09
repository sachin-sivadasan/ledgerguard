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

**No free plan.** All paid. Starter plan includes a 14-day trial. After trial expires without payment, user enters **read-only mode** (can view dashboard, cannot sync/chat/export/take actions).

```
+-----------------------------------------------------------------------+
|                         PLAN TIERS                                    |
+==========+=============+===============+===============+==============+
|          | TRIAL       | STARTER       | PRO           | ENTERPRISE   |
|          | (14 days)   | (base paid)   | (advanced)    | (custom)     |
+==========+=============+===============+===============+==============+
| Price    | $0          | $X/mo or      | $XX/mo or     | Custom       |
|          | (14 days)   | $X/yr         | $XX/yr        | (sales)      |
+----------+-------------+---------------+---------------+--------------+
| Trial    | This IS the | Paid after    | No trial      | No trial     |
|          | Starter     | trial ends    | (pay upfront) | (negotiated) |
|          | trial       |               |               |              |
+----------+-------------+---------------+---------------+--------------+
| Apps     | 1           | 1             | Unlimited     | Unlimited    |
+----------+-------------+---------------+---------------+--------------+
| Dashboard| Full        | Full          | Full          | Full         |
+----------+-------------+---------------+---------------+--------------+
| Risk     | Full        | Full          | Full          | Full +       |
|          |             |               |               | Custom Rules |
+----------+-------------+---------------+---------------+--------------+
| Sync     | Yes         | Yes           | Yes           | Yes          |
+----------+-------------+---------------+---------------+--------------+
| AI Chat  | Yes (trial) | No            | Yes           | Yes +        |
|          |             |               |               | Priority     |
+----------+-------------+---------------+---------------+--------------+
| API Keys | Yes (trial) | No            | Yes           | Yes +        |
|          |             |               |               | Higher Limits|
+----------+-------------+---------------+---------------+--------------+
| Slack    | Yes (trial) | No            | Yes           | Yes          |
+----------+-------------+---------------+---------------+--------------+
| Export   | Yes (trial) | No            | CSV/PDF       | CSV/PDF +    |
|          |             |               |               | API          |
+----------+-------------+---------------+---------------+--------------+
| Notifs   | Push +Slack | Push only     | Push + Slack  | Push + Slack |
|          | (trial)     |               |               | + Email      |
+----------+-------------+---------------+---------------+--------------+
| Support  | Community   | Community     | Email         | Dedicated +  |
|          |             |               |               | SLA          |
+----------+-------------+---------------+---------------+--------------+

EXPIRED TRIAL (read-only mode):
- Can view dashboard (last synced data, no new syncs)
- Cannot sync, use AI chat, export, manage API keys
- Upgrade prompts shown throughout UI
- Data preserved (not deleted)
- User must subscribe to Starter or higher to regain access
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
|  |           plan_tier = STARTER, billing starts                   |  |
|  |           (or PRO if user chose PRO during checkout)            |  |
|  |                                                                 |  |
|  | Option B: User didn't pay -> plan_tier = EXPIRED                |  |
|  |           Read-only mode: view dashboard, no sync/chat/export   |  |
|  |           Upgrade prompts shown, data preserved                 |  |
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
|  +------------------------------------------------------------------+ |
|  | Feature  | EXPIRED     | STARTER     | PRO         | ENTERPRISE  | |
|  |----------+-------------+-------------+-------------+-------------| |
|  | Dashboard| Read-only   | Full        | Full        | Full        | |
|  | Sync     | Blocked     | Yes         | Yes         | Yes         | |
|  | AI Chat  | Blocked     | Blocked     | Full        | Full+       | |
|  | API Keys | Blocked     | Blocked     | Full        | Higher lim  | |
|  | Slack    | Blocked     | Blocked     | Full        | Full        | |
|  | Export   | Blocked     | Blocked     | CSV/PDF     | CSV/PDF+API | |
|  | Apps     | Blocked     | 1 app       | Unlimited   | Unlimited   | |
|  | Actions  | Subscribe   | Upgrade     | Full        | Full        | |
|  |          | prompts     | prompts     |             |             | |
|  +------------------------------------------------------------------+ |
|                                                                       |
+-----------------------------------------------------------------------+
```

### Flow 4: Plan Upgrade / Downgrade

```
+-----------------------------------------------------------------------+
|                    PLAN CHANGE FLOW                                    |
+-----------------------------------------------------------------------+
|                                                                       |
|  SUBSCRIBE (EXPIRED/TRIAL -> STARTER):                                |
|  1. User clicks "Subscribe" in app                                    |
|  2. Frontend calls POST /api/v1/billing/checkout                      |
|     { plan_id: "starter_monthly", success_url, cancel_url }           |
|  3. Backend creates Stripe Checkout Session                           |
|  4. User redirected to Stripe Checkout (hosted page)                  |
|  5. User enters payment method + confirms                             |
|  6. Stripe fires checkout.session.completed webhook                   |
|  7. Backend updates billing_subscription + plan_tier = STARTER        |
|  8. User redirected to success_url, sees features unlocked            |
|                                                                       |
|  UPGRADE (STARTER -> PRO) — IMMEDIATE with proration:                 |
|  1. User clicks "Upgrade to Pro" in settings                         |
|  2. Frontend calls POST /api/v1/billing/upgrade                       |
|     { plan_id: "pro_monthly" }                                        |
|  3. Backend calls Stripe: update subscription                         |
|     - proration_behavior: "create_prorations"                         |
|     - Stripe credits unused Starter time                              |
|     - Stripe charges Pro price minus credit                           |
|     - Prorated invoice generated immediately                          |
|  4. Stripe fires customer.subscription.updated webhook                |
|  5. Backend sets plan_tier = PRO immediately                          |
|  6. AI Chat, API Keys, Slack, export, multi-app unlocked right away   |
|                                                                       |
|  DOWNGRADE (PRO -> STARTER) — SCHEDULED at period end:                |
|  1. User clicks "Downgrade to Starter" in settings                   |
|  2. Frontend calls POST /api/v1/billing/downgrade                     |
|     { plan_id: "starter_monthly" }                                    |
|  3. Backend records scheduled downgrade:                              |
|     - billing_subscription.scheduled_plan_id = starter plan ID        |
|     - billing_subscription.scheduled_change_at = current_period_end   |
|     - Does NOT call Stripe yet                                        |
|  4. User keeps PRO features until current period ends                 |
|  5. Daily cron job checks scheduled_change_at <= NOW():               |
|     - Calls Stripe: update subscription to Starter price              |
|     - Sets plan_tier = STARTER                                        |
|     - Clears scheduled_plan_id and scheduled_change_at                |
|     - Locks PRO-only features (AI Chat, API Keys, etc.)               |
|  6. User sees "Downgrade scheduled for [date]" in billing settings    |
|  7. User can cancel the scheduled downgrade before it takes effect    |
|                                                                       |
|  CANCEL (any paid -> EXPIRED) — at period end:                        |
|  1. User clicks "Cancel Subscription" in settings                     |
|  2. Frontend calls POST /api/v1/billing/cancel                        |
|  3. Backend calls Stripe: cancel at period end                        |
|     - cancel_at_period_end: true                                      |
|  4. User keeps current plan until period ends                         |
|  5. At period end: Stripe fires customer.subscription.deleted          |
|  6. Backend sets plan_tier = EXPIRED, read-only mode                  |
|                                                                       |
|  PLAN CHANGE (PRO Monthly -> PRO Annual) — IMMEDIATE with proration:  |
|  1. User selects annual plan in billing settings                      |
|  2. Backend calls Stripe: update subscription                         |
|     - proration_behavior: "create_prorations"                         |
|  3. Stripe credits remaining monthly, charges annual (prorated)       |
|  4. Webhook confirms update, same tier, new interval                  |
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

-- Seed data (no free plan — all paid, Starter has trial)
INSERT INTO plans (name, display_name, tier, price_cents, interval, trial_days, max_apps) VALUES
('starter_monthly', 'Starter (Monthly)', 'STARTER',    0,     'month', 14, 1),
('starter_annual',  'Starter (Annual)',  'STARTER',    0,     'year',  14, 1),
('pro_monthly',     'Pro (Monthly)',     'PRO',        0,     'month', 0,  -1),
('pro_annual',      'Pro (Annual)',      'PRO',        0,     'year',  0,  -1),
('enterprise',      'Enterprise',       'ENTERPRISE', 0,     'custom',0,  -1);
-- NOTE: price_cents set to 0 as placeholder — actual prices set via Stripe Price IDs
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
    status                VARCHAR(20) NOT NULL,  -- 'trialing', 'active', 'past_due', 'canceled', 'expired'
    current_period_start  TIMESTAMPTZ,
    current_period_end    TIMESTAMPTZ,
    trial_ends_at         TIMESTAMPTZ,
    canceled_at           TIMESTAMPTZ,
    scheduled_plan_id     UUID REFERENCES plans(id),       -- pending downgrade target plan
    scheduled_change_at   TIMESTAMPTZ,                     -- when downgrade takes effect (period end)
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
| POST | `/api/v1/billing/upgrade` | Firebase | Upgrade plan (immediate, prorated) |
| POST | `/api/v1/billing/downgrade` | Firebase | Schedule downgrade (takes effect at period end) |
| DELETE | `/api/v1/billing/downgrade` | Firebase | Cancel a scheduled downgrade |
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
| `customer.subscription.deleted` | Set plan_tier = EXPIRED, read-only mode |
| `customer.subscription.trial_will_end` | Send "trial ending" reminder (3 days before) |

---

## Daily Subscription Check Job (Cron)

```
+-----------------------------------------------------------------------+
|                    DAILY SUBSCRIPTION CHECK                            |
+-----------------------------------------------------------------------+
|                                                                       |
|  Runs: Every day at 00:00 UTC (backend cron / Cloud Scheduler)        |
|                                                                       |
|  Step 1: Check expired trials                                         |
|  +---------------------------------------------------------------+   |
|  | SELECT * FROM billing_subscriptions                            |   |
|  | WHERE status = 'trialing'                                      |   |
|  |   AND trial_ends_at < NOW()                                    |   |
|  |                                                                |   |
|  | For each expired trial:                                        |   |
|  |   1. Set billing_subscription.status = 'expired'               |   |
|  |   2. Set user.plan_tier = 'EXPIRED'                            |   |
|  |   3. Fire webhook: { event: "trial.expired" }                  |   |
|  |   4. Send "Trial ended" email via n8n/Postmark                 |   |
|  |   5. Log to billing_events                                     |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Step 2: Check past-due subscriptions                                 |
|  +---------------------------------------------------------------+   |
|  | SELECT * FROM billing_subscriptions                            |   |
|  | WHERE status = 'past_due'                                      |   |
|  |   AND updated_at < NOW() - INTERVAL '7 days'                   |   |
|  |                                                                |   |
|  | For each past-due > 7 days:                                    |   |
|  |   1. Call Stripe API to check latest invoice status             |   |
|  |   2. If still unpaid after Stripe retries:                      |   |
|  |      - Cancel Stripe subscription                              |   |
|  |      - Set plan_tier = 'EXPIRED'                               |   |
|  |      - Send "Subscription canceled due to payment failure"      |   |
|  |   3. If paid (Stripe retry succeeded):                          |   |
|  |      - Set status = 'active' (webhook may have already done)    |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Step 3: Execute scheduled downgrades                                 |
|  +---------------------------------------------------------------+   |
|  | SELECT * FROM billing_subscriptions                            |   |
|  | WHERE scheduled_plan_id IS NOT NULL                             |   |
|  |   AND scheduled_change_at <= NOW()                              |   |
|  |                                                                |   |
|  | For each scheduled downgrade:                                  |   |
|  |   1. Call Stripe: update subscription to scheduled plan price   |   |
|  |   2. Set plan_id = scheduled_plan_id                           |   |
|  |   3. Set user.plan_tier = new plan's tier (e.g. STARTER)       |   |
|  |   4. Clear scheduled_plan_id and scheduled_change_at            |   |
|  |   5. Lock features no longer available on lower plan            |   |
|  |   6. Send "Plan downgraded" email                              |   |
|  |   7. Log to billing_events                                     |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Step 4: Send trial reminders                                         |
|  +---------------------------------------------------------------+   |
|  | SELECT * FROM billing_subscriptions                            |   |
|  | WHERE status = 'trialing'                                      |   |
|  |                                                                |   |
|  | For each active trial:                                         |   |
|  |   days_left = trial_ends_at - NOW()                            |   |
|  |   IF days_left = 7: Send "7 days left" email                   |   |
|  |   IF days_left = 2: Send "2 days left" email                   |   |
|  |   IF days_left = 1: Send "Last day!" email + in-app warning    |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Summary of upgrade vs downgrade behavior:                            |
|  +---------------------------------------------------------------+   |
|  | Action         | When it takes effect | Proration              |   |
|  |----------------+----------------------+------------------------|   |
|  | Upgrade        | Immediately          | Yes (credit + charge)  |   |
|  | Downgrade      | Next period start    | No (use paid period)   |   |
|  | Cancel         | Period end           | No (use paid period)   |   |
|  | Monthly->Annual| Immediately          | Yes (credit + charge)  |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Implementation:                                                      |
|  - Go: goroutine with time.Ticker (simple, runs in-process)           |
|  - OR: GCP Cloud Scheduler -> POST /internal/cron/billing-check       |
|  - Endpoint protected by internal-only auth (not public)              |
|  - Idempotent: safe to run multiple times per day                     |
|  - Logs all actions to billing_events for audit trail                 |
|                                                                       |
+-----------------------------------------------------------------------+
```

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
- Flow type selector (6 tabs: Lifecycle, Payment, Checkout, Webhooks, Gating, Upgrade)
- Plan comparison table with hover highlights
- Animated webhook event processing
- Payment money flow with fee breakdown

---

## Component Structure

```
marketing/site/
+-- app/billing-flow/page.tsx            # Page wrapper (metadata, layout, explanation sections)
+-- components/
    +-- BillingFlowVisualization.tsx      # Main interactive visualization (6 tabs)
        +-- Tab 1: Lifecycle              # Trial → Starter → Pro → Enterprise state machine
        +-- Tab 2: Payment Flow           # Customer USD → Stripe fees → Developer bank (INR)
        +-- Tab 3: Checkout               # Frontend → Backend → Stripe → Webhook → DB
        +-- Tab 4: Webhooks               # Event processing with dedup/idempotency
        +-- Tab 5: Feature Gating         # Auth → Plan Check → Allow/403
        +-- Tab 6: Upgrade/Downgrade      # Proration vs scheduled changes
        +-- PaymentReference              # Fee breakdown card
        +-- LifecycleReference            # Plan states card
        +-- WebhookReference              # Event list card
        +-- GatingReference               # Feature access table
        +-- UpgradeReference              # Plan change behavior card
```

---

### Flow 5: Payment Money Flow

```
+-----------------------------------------------------------------------+
|                    PAYMENT MONEY FLOW                                   |
+-----------------------------------------------------------------------+
|                                                                       |
|  Customer → pays $29/mo (USD via card)                                |
|    → Stripe processes payment                                         |
|    → Stripe deducts: platform fee (~2.9% + $0.30)                     |
|      Processing: 2.9% × $29.00 = $0.84                               |
|      Fixed fee: $0.30                                                 |
|      Total fee: $1.14                                                 |
|    → Net: $27.86                                                      |
|    → Stripe India: converts USD → INR at market rate                  |
|      $27.86 × ₹83.30 = ~₹2,321 (rate varies daily)                   |
|    → Payout to developer's Indian bank account                        |
|    → Settlement: T+2 (standard) to T+7 (first payout)                |
|                                                                       |
|  Monthly Revenue Example (50 Pro subscribers):                        |
|  +---------------------------------------------------------------+   |
|  | 50 × $29.00 = $1,450.00 gross                                 |   |
|  | Stripe fees: ~$57.00                                           |   |
|  | Net payout: ~$1,393.00                                         |   |
|  | INR equivalent: ~₹116,037                                      |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
+-----------------------------------------------------------------------+
```

### Flow 6: Stripe Customer Creation Strategies

```
+-----------------------------------------------------------------------+
|                    STRIPE CUSTOMER CREATION                            |
+-----------------------------------------------------------------------+
|                                                                       |
|  Option A: At Signup [CHOSEN]                                         |
|  +---------------------------------------------------------------+   |
|  | Create Stripe Customer immediately on user registration        |   |
|  | No payment method yet — just email + metadata                  |   |
|  |                                                                |   |
|  | Why chosen:                                                    |   |
|  |   1. Simplifies trial tracking (Stripe knows from day 1)      |   |
|  |   2. Stripe sends trial_will_end reminders automatically       |   |
|  |   3. No lazy-init complexity in checkout flow                  |   |
|  |   4. Enables Customer Portal access during trial               |   |
|  |   5. Industry standard for B2B SaaS                           |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Option B: At First Checkout                                          |
|  +---------------------------------------------------------------+   |
|  | + Fewer Stripe Customers (only paying users)                   |   |
|  | + Lower Stripe API calls at signup                             |   |
|  | - Cannot use Stripe trial features                             |   |
|  | - Must build own trial tracking                                |   |
|  | - More complex checkout (create customer + subscription)       |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
|  Option C: At Trial End                                               |
|  +---------------------------------------------------------------+   |
|  | + Fewest Stripe Customers                                      |   |
|  | - Cannot use Stripe trial reminders                            |   |
|  | - Delayed Stripe integration                                   |   |
|  | - Risk: user leaves before Stripe Customer created             |   |
|  +---------------------------------------------------------------+   |
|                                                                       |
+-----------------------------------------------------------------------+
```

---

## Auto-Deduct Behaviors

### 1. Auto-Renewal Charges
Stripe automatically charges the saved payment method each billing cycle (monthly/annual) without user action. This is standard SaaS auto-debit behavior.
- Stripe creates an invoice at each period end
- Charges the default payment method on file
- Fires `invoice.paid` on success, `invoice.payment_failed` on failure
- User sees charge on card statement as "LEDGERGUARD" (configurable in Stripe Dashboard)
- No user action required — fully automatic recurring billing

### 2. Auto-Downgrade on Payment Failure
After Stripe exhausts smart retry attempts (typically 3-4 retries over ~2 weeks), the system automatically downgrades the user.
- `invoice.payment_failed` → set `status = past_due`, send alert
- Daily cron checks `past_due > 7 days` → verify with Stripe API
- If Stripe retries exhausted → cancel subscription → `plan_tier = EXPIRED`
- User enters read-only mode automatically (no manual intervention)
- Send "Subscription canceled due to payment failure" email

### 3. Auto Trial-to-Paid Conversion
If user adds a payment method during trial, Stripe automatically creates the subscription when trial ends.
- During trial: user visits Stripe Customer Portal or checkout to add card
- At trial expiry: Stripe creates subscription + first invoice
- Fires `checkout.session.completed` → backend sets `plan_tier = STARTER`
- Seamless transition — user keeps access without interruption
- If no payment method added: `plan_tier = EXPIRED` (read-only mode)

### 4. RBI Auto-Debit Mandate (India)
For Indian cards, RBI (Reserve Bank of India) requires additional authorization for recurring payments exceeding ₹15,000/transaction.
- Stripe handles e-mandate registration during first checkout
- Payments ≤ ₹15,000: auto-debit without additional auth
- Payments > ₹15,000: Stripe sends pre-debit notification 24h before charge
- Customer must approve via bank/UPI app for amounts > ₹15,000
- If customer doesn't approve: payment fails → follow auto-downgrade flow
- LedgerGuard plans likely under ₹15,000 threshold ($29 Pro ≈ ₹2,400), so most charges auto-debit seamlessly
- Enterprise custom pricing may exceed threshold — handle with Stripe's mandate flow

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
9. **Auto-renewal:** Stripe handles recurring charges automatically — no backend cron needed for billing
10. **Auto-downgrade:** Daily cron + Stripe webhooks handle failed payment → EXPIRED transition
11. **RBI compliance:** Stripe India handles e-mandate registration; most LedgerGuard charges below ₹15,000 threshold

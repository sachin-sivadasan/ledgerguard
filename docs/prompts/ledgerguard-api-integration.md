# LedgerGuard API Integration - Interactive Visualization

## Context

You are a senior frontend + visualization engineer building an interactive animated diagram showing how **LedgerGuard's Revenue Status API** integrates into Shopify app developer workflows.

This visualization helps app developers understand:
1. **Where LedgerGuard fits** in their infrastructure
2. **What data flows** through the system
3. **How to integrate** the API into their apps
4. **Real-time status checks** for subscription health

---

## The Problem LedgerGuard Solves

### Without LedgerGuard
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    CURRENT STATE: FLYING BLIND                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Your App                           Shopify Partner API                    │
│   ┌──────────┐                       ┌──────────────────┐                   │
│   │          │ ─── Query? ──────────▶│   transactions   │                   │
│   │  "Is     │                       │   (raw data)     │                   │
│   │  store-x │ ◀─── Giant JSON ─────│                  │                   │
│   │  paying?"│                       │   No risk info   │                   │
│   │          │                       │   No MRR calc    │                   │
│   └──────────┘                       │   No alerts      │                   │
│        │                             └──────────────────┘                   │
│        │                                                                     │
│        ▼                                                                     │
│   ┌──────────────────────────────────────────┐                              │
│   │  YOU HAVE TO:                            │                              │
│   │  • Parse complex transaction objects      │                              │
│   │  • Calculate days past due               │                              │
│   │  • Determine risk state                  │                              │
│   │  • Handle edge cases (prorations, etc)   │                              │
│   │  • Build your own alerting               │                              │
│   │  • Store historical data                 │                              │
│   └──────────────────────────────────────────┘                              │
│                                                                              │
│   ⚠️  Time-consuming, error-prone, every app rebuilds the same logic       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### With LedgerGuard
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    WITH LEDGERGUARD: INSTANT ANSWERS                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Your App               LedgerGuard API              Shopify Partner API   │
│   ┌──────────┐           ┌──────────────┐            ┌──────────────────┐  │
│   │          │           │              │            │                  │  │
│   │  "Is     │ ─ GET ──▶ │  /v1/sub/    │ ◀── Sync ─│   transactions   │  │
│   │  store-x │           │  status      │            │                  │  │
│   │  paying?"│ ◀─ JSON ─ │              │            │                  │  │
│   │          │           │  ✅ SAFE     │            └──────────────────┘  │
│   └──────────┘           │  or          │                                   │
│                          │  ⚠️ AT_RISK  │                                   │
│   Response in <50ms      │  or          │                                   │
│                          │  💀 CHURNED  │                                   │
│                          └──────────────┘                                   │
│                                                                              │
│   ✅ Pre-calculated risk state                                              │
│   ✅ MRR normalized (monthly/annual)                                        │
│   ✅ Days past due computed                                                 │
│   ✅ Historical data stored                                                 │
│   ✅ Webhook alerts available                                               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Architecture

### Complete System Flow
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        LEDGERGUARD DATA FLOW                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌───────────────┐                                                          │
│  │   Shopify     │                                                          │
│  │ Partner API   │                                                          │
│  └───────┬───────┘                                                          │
│          │                                                                   │
│          │ GraphQL: transactions(last: 12 months)                           │
│          │ Every 12 hours (00:00, 12:00 UTC)                                │
│          ▼                                                                   │
│  ┌───────────────────────────────────────────────────────────────────┐      │
│  │                      LEDGERGUARD BACKEND                          │      │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │      │
│  │  │ Sync Engine │─▶│   Ledger    │─▶│    Risk     │               │      │
│  │  │             │  │   Engine    │  │   Engine    │               │      │
│  │  │ Fetch &     │  │ Classify    │  │ 0-30: SAFE  │               │      │
│  │  │ Store       │  │ transactions│  │ 31-60: WARN │               │      │
│  │  └─────────────┘  └─────────────┘  │ 61-90: CRIT │               │      │
│  │                                     │ 90+: CHURN │               │      │
│  │                                     └──────┬──────┘               │      │
│  │                                            │                      │      │
│  │  ┌─────────────────────────────────────────┼───────────────────┐ │      │
│  │  │              SUBSCRIPTION STATUS DB     │                   │ │      │
│  │  │  ┌──────────┬───────────┬─────────┬────┴─────┬─────────┐   │ │      │
│  │  │  │ shop_gid │ mrr_cents │ risk    │ days_due │ plan    │   │ │      │
│  │  │  ├──────────┼───────────┼─────────┼──────────┼─────────┤   │ │      │
│  │  │  │ gid://123│ 4900      │ SAFE    │ 5        │ Pro     │   │ │      │
│  │  │  │ gid://456│ 2900      │ AT_RISK │ 45       │ Starter │   │ │      │
│  │  │  │ gid://789│ 9900      │ CHURNED │ 95       │ Business│   │ │      │
│  │  │  └──────────┴───────────┴─────────┴──────────┴─────────┘   │ │      │
│  │  └─────────────────────────────────────────────────────────────┘ │      │
│  └───────────────────────────────────────────────────────────────────┘      │
│          │                                     │                             │
│          │ REST API                            │ Webhooks (coming soon)      │
│          ▼                                     ▼                             │
│  ┌───────────────────────────────────────────────────────────────────┐      │
│  │                     YOUR APPLICATION                               │      │
│  │                                                                    │      │
│  │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐            │      │
│  │   │   Checkout  │   │  Admin      │   │  Alerting   │            │      │
│  │   │   Flow      │   │  Dashboard  │   │  System     │            │      │
│  │   │             │   │             │   │             │            │      │
│  │   │ "Is store   │   │ "Show all   │   │ "Notify     │            │      │
│  │   │  active?"   │   │  at-risk"   │   │  when risk  │            │      │
│  │   │             │   │             │   │  changes"   │            │      │
│  │   └─────────────┘   └─────────────┘   └─────────────┘            │      │
│  │                                                                    │      │
│  └───────────────────────────────────────────────────────────────────┘      │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

### Authentication

All API requests require an API key in the header:

```
Authorization: Bearer lgk_live_xxxxxxxxxxxxxxxxxxxx
```

API keys are generated in the LedgerGuard dashboard under Settings > API Keys.

### Endpoint Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/subscription/{shopify_gid}/status` | GET | Get status by Shopify GID |
| `/v1/subscription/status?domain={domain}` | GET | Get status by myshopify domain |
| `/v1/subscriptions/status/batch` | POST | Get multiple statuses (max 100) |
| `/v1/usage/{shopify_gid}/status` | GET | Get usage billing status |

---

## Response Objects

### Subscription Status Response

```json
{
  "shopify_gid": "gid://shopify/AppSubscription/12345678",
  "myshopify_domain": "cool-store.myshopify.com",
  "status": "ACTIVE",
  "risk_state": "SAFE",
  "plan_name": "Pro Plan",
  "mrr_cents": 4900,
  "billing_interval": "MONTHLY",
  "days_past_due": 5,
  "current_period_end": "2026-03-15T00:00:00Z",
  "last_charge_date": "2026-02-15T00:00:00Z",
  "last_charge_amount_cents": 4900,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2026-03-01T12:00:00Z"
}
```

### Risk State Values

| Risk State | Days Past Due | Color | Meaning |
|------------|---------------|-------|---------|
| `SAFE` | 0-30 | Green | Payment on track or within grace |
| `ONE_CYCLE_MISSED` | 31-60 | Amber | Missed one billing cycle |
| `TWO_CYCLES_MISSED` | 61-90 | Red | Critical - two cycles missed |
| `CHURNED` | 90+ | Gray | Customer lost |

### Batch Response

```json
{
  "results": [
    {
      "shopify_gid": "gid://shopify/AppSubscription/123",
      "risk_state": "SAFE",
      "mrr_cents": 4900
    },
    {
      "shopify_gid": "gid://shopify/AppSubscription/456",
      "risk_state": "ONE_CYCLE_MISSED",
      "mrr_cents": 2900
    }
  ],
  "not_found": [
    "gid://shopify/AppSubscription/999"
  ]
}
```

---

## Integration Patterns

### Pattern 1: Checkout/Install Flow
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    INTEGRATION: CHECKOUT FLOW                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Merchant installs your app                                                │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────┐                                                       │
│   │  Your App       │                                                       │
│   │  Checkout Page  │                                                       │
│   └────────┬────────┘                                                       │
│            │                                                                 │
│            │ GET /v1/subscription/status?domain=store.myshopify.com         │
│            ▼                                                                 │
│   ┌─────────────────┐         ┌─────────────────────────────────┐          │
│   │  LedgerGuard    │────────▶│  Response:                      │          │
│   │  API            │         │  { "risk_state": "SAFE" }       │          │
│   └─────────────────┘         │  OR                             │          │
│                               │  { "error": "not_found" }       │          │
│                               └─────────────────────────────────┘          │
│            │                                                                 │
│            ▼                                                                 │
│   ┌─────────────────────────────────────────────────────────────┐          │
│   │  DECISION LOGIC:                                            │          │
│   │                                                              │          │
│   │  if (status.risk_state === "SAFE") {                        │          │
│   │    // Existing customer, good standing                      │          │
│   │    showWelcomeBack();                                       │          │
│   │  } else if (status.risk_state === "CHURNED") {              │          │
│   │    // Previous customer who churned                         │          │
│   │    showReactivationOffer();                                 │          │
│   │  } else if (status.error === "not_found") {                 │          │
│   │    // Brand new customer                                    │          │
│   │    showNewCustomerFlow();                                   │          │
│   │  }                                                           │          │
│   │                                                              │          │
│   └─────────────────────────────────────────────────────────────┘          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Pattern 2: Admin Dashboard
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    INTEGRATION: ADMIN DASHBOARD                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Your App's Admin Panel (for app developers)                               │
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                      SUBSCRIPTION HEALTH                             │  │
│   │  ────────────────────────────────────────────────────────────────   │  │
│   │                                                                      │  │
│   │   ✅ SAFE                 ⚠️ AT RISK               💀 CHURNED       │  │
│   │   ═══════                 ═════════                ════════         │  │
│   │   612 stores              127 stores               108 stores       │  │
│   │   $29,988 MRR             $6,223 MRR               $5,292 MRR       │  │
│   │                                                                      │  │
│   │   [███████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]  │  │
│   │                                                                      │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│   Built using:                                                              │
│                                                                              │
│   // Fetch all subscriptions from your DB                                   │
│   const subscriptions = await db.subscriptions.findAll();                   │
│                                                                              │
│   // Batch lookup statuses from LedgerGuard                                 │
│   const statuses = await ledgerguard.batch({                                │
│     ids: subscriptions.map(s => s.shopify_gid)                              │
│   });                                                                        │
│                                                                              │
│   // Group by risk state                                                    │
│   const grouped = groupBy(statuses.results, 'risk_state');                  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Pattern 3: Proactive Alerting
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    INTEGRATION: PROACTIVE ALERTING                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Scheduled Job (cron: every 6 hours)                                       │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  1. Fetch all active subscriptions from your DB                     │  │
│   │  2. Batch lookup from LedgerGuard API                               │  │
│   │  3. Compare with previous state                                     │  │
│   │  4. Alert on state changes                                          │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                      │  │
│   │   if (previous === "SAFE" && current === "ONE_CYCLE_MISSED") {      │  │
│   │     sendSlackAlert(`⚠️ ${store} moved to AT RISK`);                 │  │
│   │     sendEmailToCustomerSuccess(store);                              │  │
│   │   }                                                                  │  │
│   │                                                                      │  │
│   │   if (previous === "ONE_CYCLE_MISSED" && current === "SAFE") {      │  │
│   │     sendSlackAlert(`✅ ${store} recovered!`);                       │  │
│   │   }                                                                  │  │
│   │                                                                      │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│         │                                                                    │
│         ▼                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  SLACK NOTIFICATION                                                  │  │
│   │  ─────────────────                                                   │  │
│   │  🔔 LedgerGuard Alert                                               │  │
│   │                                                                      │  │
│   │  ⚠️ Subscription moved to AT RISK                                   │  │
│   │                                                                      │  │
│   │  Store: cool-store.myshopify.com                                    │  │
│   │  Plan: Pro ($49/mo)                                                 │  │
│   │  Days Past Due: 35                                                  │  │
│   │                                                                      │  │
│   │  [View in Dashboard]  [Contact Store]                               │  │
│   │                                                                      │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Pattern 4: Feature Gating
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    INTEGRATION: FEATURE GATING                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Your App (Server-side middleware)                                         │
│                                                                              │
│   async function checkSubscriptionMiddleware(req, res, next) {              │
│     const domain = req.headers['x-shopify-shop-domain'];                    │
│                                                                              │
│     const status = await ledgerguard.getStatus({ domain });                 │
│                                                                              │
│     if (status.risk_state === 'CHURNED') {                                  │
│       // Block access, subscription expired                                 │
│       return res.status(402).json({                                         │
│         error: 'subscription_expired',                                      │
│         message: 'Please renew your subscription',                          │
│         reactivate_url: '/billing/reactivate'                               │
│       });                                                                    │
│     }                                                                        │
│                                                                              │
│     if (status.risk_state === 'TWO_CYCLES_MISSED') {                        │
│       // Soft block - show warning, limit features                          │
│       req.subscriptionWarning = true;                                       │
│     }                                                                        │
│                                                                              │
│     req.subscription = status;                                              │
│     next();                                                                  │
│   }                                                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Code Examples

### JavaScript/TypeScript SDK
```typescript
import { LedgerGuard } from '@ledgerguard/sdk';

const lg = new LedgerGuard({
  apiKey: process.env.LEDGERGUARD_API_KEY,
});

// Single lookup by domain
const status = await lg.subscriptions.getByDomain('cool-store.myshopify.com');
console.log(status.risk_state); // 'SAFE' | 'ONE_CYCLE_MISSED' | ...

// Single lookup by Shopify GID
const status2 = await lg.subscriptions.getByGID('gid://shopify/AppSubscription/123');

// Batch lookup (up to 100)
const batch = await lg.subscriptions.getBatch([
  'gid://shopify/AppSubscription/123',
  'gid://shopify/AppSubscription/456',
]);
```

### cURL Examples
```bash
# Get by domain
curl -X GET "https://api.ledgerguard.io/v1/subscription/status?domain=cool-store.myshopify.com" \
  -H "Authorization: Bearer lgk_live_xxxxxxxxxxxx"

# Get by Shopify GID
curl -X GET "https://api.ledgerguard.io/v1/subscription/gid://shopify/AppSubscription/123/status" \
  -H "Authorization: Bearer lgk_live_xxxxxxxxxxxx"

# Batch lookup
curl -X POST "https://api.ledgerguard.io/v1/subscriptions/status/batch" \
  -H "Authorization: Bearer lgk_live_xxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["gid://shopify/AppSubscription/123", "gid://shopify/AppSubscription/456"]}'
```

---

## Animation Sequences

### Sequence 1: Single Status Lookup
```
Step 1: Show your app making a request
Step 2: Animate request traveling to LedgerGuard
Step 3: Show LedgerGuard checking its database
Step 4: Animate response traveling back
Step 5: Show your app displaying the status
```

### Sequence 2: Batch Lookup
```
Step 1: Show array of subscription IDs
Step 2: Single request with all IDs
Step 3: LedgerGuard processes in parallel
Step 4: Single response with all statuses
Step 5: Your app updates UI for all
```

### Sequence 3: Risk State Transition
```
Step 1: Show subscription in SAFE state
Step 2: Days counter advances (30 → 31)
Step 3: Status changes to ONE_CYCLE_MISSED
Step 4: Alert fires (webhook or poll detection)
Step 5: Your app takes action
```

### Sequence 4: Full Data Flow
```
Step 1: Shopify Partner API (raw transactions)
Step 2: LedgerGuard syncs (every 12 hours)
Step 3: Ledger rebuild + risk classification
Step 4: Status stored in LedgerGuard DB
Step 5: Your app queries via API
Step 6: Instant response (<50ms)
```

---

## Visual Requirements

### Layout
```
╔══════════════════════════════════════════════════════════════════════════╗
║                           LEDGERGUARD API                                 ║
║                    Revenue Status for Shopify Apps                        ║
╠══════════════════════════════════════════════════════════════════════════╣
║                                                                           ║
║   ┌──────────────┐         ┌──────────────┐         ┌──────────────┐    ║
║   │   Shopify    │ ──────▶ │  LedgerGuard │ ──────▶ │   Your App   │    ║
║   │ Partner API  │  sync   │    API       │  query  │              │    ║
║   └──────────────┘         └──────────────┘         └──────────────┘    ║
║                                    │                                      ║
║                            ┌───────┴───────┐                             ║
║                            │ risk_state:   │                             ║
║                            │ ✅ SAFE       │                             ║
║                            │ mrr: $49/mo   │                             ║
║                            └───────────────┘                             ║
║                                                                           ║
╚══════════════════════════════════════════════════════════════════════════╝
```

### Interactive Elements
- **Toggle**: Single Lookup vs Batch Lookup
- **Input**: Enter a test domain or GID
- **Animated request/response**: Shows data flowing
- **Risk state selector**: Show different response scenarios
- **Code snippet tabs**: JavaScript, cURL, Python

### Color Scheme
```
Shopify Partner API:    #96bf48 (Shopify green)
LedgerGuard:            #6366f1 (Indigo)
Your App:               #3b82f6 (Blue)
SAFE:                   #22c55e (Green)
ONE_CYCLE_MISSED:       #f59e0b (Amber)
TWO_CYCLES_MISSED:      #ef4444 (Red)
CHURNED:                #6b7280 (Gray)
Request arrows:         #818cf8 (Light indigo)
Response arrows:        #22c55e (Green)
```

---

## Key Messages to Convey

1. **"One API call = Instant subscription health"**
   - No complex transaction parsing
   - Pre-calculated risk state
   - <50ms response time

2. **"Same thresholds as the dashboard"**
   - SAFE: 0-30 days
   - AT_RISK: 31-90 days
   - CHURNED: 90+ days

3. **"Batch operations for efficiency"**
   - Up to 100 subscriptions per request
   - Single HTTP call
   - Parallel processing

4. **"Always fresh data"**
   - Synced every 12 hours
   - Data never more than 12 hours stale
   - On-demand sync available (Pro)

5. **"Build features, not infrastructure"**
   - Checkout flow integration
   - Admin dashboard widgets
   - Alerting systems
   - Feature gating

---

## Rate Limits & Best Practices

### Rate Limits
| Tier | Requests/minute | Batch size |
|------|-----------------|------------|
| Free | 60 | 10 |
| Pro | 300 | 100 |
| Enterprise | Custom | Custom |

### Best Practices
- **Cache responses** for 5-15 minutes
- **Use batch endpoints** when checking multiple subscriptions
- **Implement exponential backoff** for 429 errors
- **Subscribe to webhooks** (coming soon) for real-time updates

---

## Error Responses

```json
// 401 Unauthorized
{
  "error": "unauthorized",
  "message": "Invalid or missing API key"
}

// 403 Forbidden
{
  "error": "access_denied",
  "message": "API key does not have access to this app"
}

// 404 Not Found
{
  "error": "not_found",
  "message": "Subscription not found"
}

// 429 Too Many Requests
{
  "error": "rate_limited",
  "message": "Rate limit exceeded",
  "retry_after": 60
}
```

---

## File Locations

- **Prompt:** `docs/prompts/ledgerguard-api-integration.md`
- **Component:** `marketing/site/components/APIIntegrationGuide.tsx`
- **Page:** `marketing/site/app/api-guide/page.tsx`
- **View:** http://localhost:3000/api-guide

---

## Implementation Checklist

- [ ] Create page at `/api-guide`
- [ ] Build APIIntegrationGuide component
- [ ] Implement animated data flow (Shopify → LedgerGuard → Your App)
- [ ] Add request/response animation
- [ ] Add code snippet tabs (JS, cURL, Python)
- [ ] Add interactive "try it" panel with mock responses
- [ ] Add integration pattern selector (Checkout, Dashboard, Alerting, Gating)
- [ ] Add risk state visualization matching KPI guide
- [ ] Add responsive design for mobile
- [ ] Test all animation sequences

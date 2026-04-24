# 19. Razorpay Billing

## What It Does
Handles LedgerGuard's own B2B SaaS billing using Razorpay Subscriptions. Users can subscribe to Starter ($249/mo) or Pro ($499/mo) plans via a Razorpay-hosted checkout page. Webhook events drive subscription lifecycle transitions (activated, charged, halted, cancelled). This is entirely separate from the Shopify subscription monitoring -- `BillingSubscription` tracks LedgerGuard's paying customers, while `Subscription` tracks their Shopify app subscribers.

## Architecture
```
┌────────────┐    ┌──────────────┐    ┌─────────────────┐    ┌──────────────────┐
│ Flutter    │───▶│BillingHandler│───▶│ BillingService   │───▶│ RazorpayClient   │
│ App        │    │              │    │                  │    │ (Basic Auth)     │
└────────────┘    │ POST checkout│    │ CreateCheckout() │    │                  │
                  │ GET  status  │    │ GetBillingStatus│    │ CreateCustomer() │
                  └──────────────┘    │ HandleWebhook() │    │ CreateSubscription│
                                      └────────┬────────┘    │ FetchSubscription│
                                               │             │ CancelSubscription│
┌────────────┐                                 │             │ VerifyWebhookSig │
│ Razorpay   │    POST /webhooks/razorpay      │             └──────────────────┘
│ Webhook    │─────────────────────────────────▶│
└────────────┘                                  │
                                               ▼
                              ┌──────────────────────────┐
                              │ BillingSubscriptionRepo  │
                              │ (PostgreSQL)              │
                              │                          │
                              │ + UserRepository         │
                              │   (plan tier updates)    │
                              └──────────────────────────┘
```

### Domain Separation
| Concept | Entity | Table | Purpose |
|---------|--------|-------|---------|
| Shopify subscriber store | `Subscription` | `subscriptions` | Monitored app subscription from Shopify Partner API |
| LedgerGuard paying user | `BillingSubscription` | `billing_subscriptions` | LedgerGuard's own SaaS billing |

These two never interact. They share no code paths.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/interfaces/http/handler/billing.go` | ~97 | HTTP handlers: checkout, status, webhook |
| `backend/internal/application/service/billing_service.go` | ~289 | Business logic: checkout flow, webhook routing, plan tier updates |
| `backend/internal/infrastructure/external/razorpay_client.go` | ~195 | Razorpay API client (Basic Auth, HMAC verification) |
| `backend/internal/domain/entity/billing_subscription.go` | ~99 | BillingSubscription entity with state transitions |
| `backend/internal/domain/valueobject/billing_plan.go` | ~47 | BillingPlan value object (STARTER, PRO) with prices |
| `backend/internal/domain/valueobject/billing_subscription_status.go` | ~62 | BillingSubscriptionStatus value object (CREATED through COMPLETED) |
| `backend/internal/domain/repository/billing_subscription_repository.go` | ~18 | Repository interface |

## Data Flow

### Checkout Flow
```
Flutter App                    Backend                          Razorpay
    │                             │                               │
    │ POST /api/v1/billing/checkout                               │
    │ { "plan": "STARTER" }       │                               │
    │────────────────────────────▶│                               │
    │                             │                               │
    │                             │ 1. Validate plan              │
    │                             │ 2. Look up user email         │
    │                             │ 3. Reuse or create customer   │
    │                             │    POST /v1/customers ────────▶│
    │                             │◀────── { id: "cust_xxx" } ───│
    │                             │                               │
    │                             │ 4. Create subscription         │
    │                             │    POST /v1/subscriptions ───▶│
    │                             │    { plan_id, customer_id,    │
    │                             │      total_count: 120,        │
    │                             │      customer_notify: 1 }     │
    │                             │◀── { id, short_url } ────────│
    │                             │                               │
    │                             │ 5. Persist BillingSubscription │
    │                             │    (status: CREATED)           │
    │                             │                               │
    │◀── { subscription_id,      │                               │
    │      short_url }           │                               │
    │                             │                               │
    │ Open short_url in browser   │                               │
    │─────────────────────────────────────────────────────────────▶│
    │                             │                               │
    │                             │ Razorpay sends webhooks:      │
    │                             │◀─ subscription.activated ─────│
    │                             │◀─ subscription.charged ───────│
    │                             │                               │
```

### Webhook Processing
```
Razorpay                          Backend
    │                               │
    │ POST /webhooks/razorpay       │
    │ X-Razorpay-Signature: hmac    │
    │ { event, payload }            │
    │──────────────────────────────▶│
    │                               │
    │                  1. HMAC-SHA256 verification
    │                  2. Parse event type
    │                  3. Find BillingSubscription by razorpay_id
    │                  4. Route to handler:
    │                     │
    │                     ├─ subscription.activated → Activate() + update user plan tier
    │                     ├─ subscription.charged   → UpdatePeriod()
    │                     ├─ subscription.pending   → MarkPending()
    │                     ├─ subscription.halted    → Halt()
    │                     └─ subscription.cancelled → Cancel() + downgrade to FREE
    │                               │
    │◀──── 200 OK (always) ────────│
    │                               │
```

### Status Lifecycle
```
CREATED ──▶ ACTIVE ──▶ HALTED ──▶ CANCELLED
   │           │           │
   │           │           └──▶ ACTIVE (retry succeeded)
   │           │
   │           └──▶ CANCELLED
   │
   └──▶ PENDING ──▶ ACTIVE
```

## Configuration
| Setting | Source | Description |
|---------|--------|-------------|
| `RAZORPAY_KEY_ID` | Environment / config | Razorpay API key ID (Basic Auth username) |
| `RAZORPAY_KEY_SECRET` | Environment / config | Razorpay API key secret (Basic Auth password) |
| `RAZORPAY_WEBHOOK_SECRET` | Environment / config | HMAC-SHA256 webhook verification secret |
| `RAZORPAY_STARTER_PLAN_ID` | Environment / config | Razorpay plan ID for Starter ($249/mo) |
| `RAZORPAY_PRO_PLAN_ID` | Environment / config | Razorpay plan ID for Pro ($499/mo) |

**Important**: Plan IDs must be created in the Razorpay Dashboard first. They are not created programmatically. Billing routes are optional -- the server starts without error if Razorpay config is missing.

### Pricing
| Plan | Monthly Price | Razorpay Plan ID Config Key |
|------|---------------|----------------------------|
| STARTER | $249 (24900 cents) | `starterPlanID` |
| PRO | $499 (49900 cents) | `proPlanID` |

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/v1/billing/checkout | Firebase | Create Razorpay subscription, return hosted checkout URL |
| GET | /api/v1/billing/status | Firebase | Get current billing status for authenticated user |
| POST | /webhooks/razorpay | None (HMAC) | Razorpay webhook receiver |

### Checkout Request / Response
```json
// Request
{ "plan": "STARTER" }

// Response
{
  "subscription_id": "sub_xxx",
  "short_url": "https://rzp.io/i/xxx"
}
```

### Status Response
```json
{
  "plan": "STARTER",
  "status": "ACTIVE",
  "amount_cents": 24900,
  "currency": "USD",
  "current_period_start": "2025-06-01T00:00:00Z",
  "current_period_end": "2025-07-01T00:00:00Z",
  "short_url": "https://rzp.io/i/xxx"
}
```

### No Active Subscription
```json
{
  "plan": "FREE",
  "status": "NONE"
}
```

## Extension Points
- Add plan upgrade/downgrade flow (cancel current, create new with prorated logic)
- Implement `CancelSubscription` endpoint for self-service cancellation (client method already exists)
- Add billing history endpoint using Razorpay's invoices API
- Integrate with email service to send payment receipts
- Add annual billing plans alongside monthly
- Implement trial periods using Razorpay's `start_at` parameter

## Gotchas
- **Webhook always returns 200**: The `HandleWebhookEvent` method returns `nil` for processing errors (logs them instead). Only signature verification failure returns a non-nil error (HTTP 400). This prevents Razorpay from retrying on transient failures and creating duplicate events.
- **customer_notify must be 1**: Required for Razorpay's hosted checkout page to work. Without it, the `short_url` is not generated.
- **total_count is 120**: Represents 10 years of monthly billing cycles. Razorpay requires a finite total count for subscriptions.
- **Customer reuse**: If a user already has a billing subscription, the existing `razorpay_customer_id` is reused instead of creating a duplicate customer.
- **Plan tier sync**: On `subscription.activated`, the user's `PlanTier` is updated (Starter or Pro). On `subscription.cancelled`, it is downgraded to `FREE`. This drives feature gating across the platform.
- **Test mode only**: Currently configured with Razorpay test keys. Switch to live keys for production.
- **RazorpayClient is injectable**: `NewRazorpayClientWithHTTPClient()` accepts a custom `http.Client` and `baseURL` for testing with mock servers.
- **Period timestamps are Unix**: Razorpay sends `current_start` and `current_end` as Unix timestamps (int64 pointers). The `extractPeriod()` helper converts them to `time.Time`.
- **Billing routes are optional**: If Razorpay config is missing, the billing endpoints are simply not registered. The server starts and runs without billing functionality.

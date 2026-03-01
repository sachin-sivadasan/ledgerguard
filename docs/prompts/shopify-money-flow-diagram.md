# Shopify Revenue Flows - Complete Visualization

## Context
You are a senior frontend + visualization engineer.

Build an interactive animated diagram showing the TWO SEPARATE money flows in the Shopify ecosystem, helping app developers understand the difference between **Transaction Revenue** (merchant's money) and **App Revenue** (developer's money).

---

## The Two Flows (CRITICAL DISTINCTION)

### Flow 1: Transaction Revenue (Merchant's Money)
```
┌──────────┐         ┌──────────┐         ┌──────────┐
│ Customer │ ─$100── │ Shopify  │ ─$97─── │ Merchant │
│ (buyer)  │ payment │ Payments │ payout  │ (seller) │
└──────────┘         └──────────┘         └──────────┘
                           │
                    keeps ~2.9% + $0.30
                    (transaction fee)
```

**This is NOT your money as an app developer.**
- Customer buys product from merchant's store
- Shopify processes payment (2.9% + $0.30 fee)
- Merchant receives payout (net amount)
- Payouts are typically every 1-3 business days

### Flow 2: App Revenue (Developer's Money)

**What Gets Deducted From Your App Revenue:**
```
┌─────────────────────────────────────────────────────────────┐
│                    DEDUCTIONS FROM APP SALE                  │
├─────────────────────────────────────────────────────────────┤
│  Merchant Pays (grossAmount)                    $100.00     │
│  ─────────────────────────────────────────────────────────  │
│  - Revenue Share (shopifyFee)                   -$20.00     │
│    (0%, 15%, or 20% depending on tier)                      │
│  - Processing Fee (processingFee) 2.9%          -$2.90      │
│  - Sales Tax (on fees)                          -$X.XX      │
│  - Regulatory Operating Fee (some regions)      -$X.XX      │
│  ─────────────────────────────────────────────────────────  │
│  You Receive (netAmount)                        $77.10*     │
│  *varies by tier and region                                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Revenue Share Tiers (NOT Always 20%!)

### Default: 20% Revenue Share
If you haven't registered for the reduced plan.

### Reduced Revenue Share Plan ($19 one-time registration)

**Tier 1: Small Developers**
- Prior-year app-store gross earnings < $20M AND company revenue < $100M
- **0% revenue share** on first $1,000,000 USD lifetime gross app revenue (earned on/after Jan 1, 2025)
- **15% revenue share** on lifetime earnings over $1,000,000 USD

**Tier 2: Large Developers**
- Prior-year app-store gross earnings ≥ $20M OR company revenue ≥ $100M
- **15% revenue share** on ALL app revenue (no 0% band)

### Visual: Revenue Share Tiers
```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SHOPIFY REVENUE SHARE TIERS                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  DEFAULT (no registration)                                               │
│  ════════════════════════════════════════════════════════════           │
│  ████████████████████ 20% on ALL revenue                                │
│                                                                          │
│  REDUCED PLAN - SMALL DEVELOPER (<$20M app / <$100M company)            │
│  ════════════════════════════════════════════════════════════           │
│  [         0%          ][███████ 15% ████████████████████████]          │
│  └── First $1M lifetime ─┘└── Everything over $1M ──────────┘           │
│                                                                          │
│  REDUCED PLAN - LARGE DEVELOPER (≥$20M app OR ≥$100M company)           │
│  ════════════════════════════════════════════════════════════           │
│  ███████████████ 15% on ALL revenue (no 0% band)                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## All Fees Breakdown

| Fee Type | Amount | Shown In API As | Notes |
|----------|--------|-----------------|-------|
| Revenue Share | 0% / 15% / 20% | `shopifyFee` | Depends on tier |
| Processing Fee | 2.9% | `processingFee` | Always applies |
| Sales Tax | Varies | `TaxTransaction` | On Shopify's fees, not your revenue |
| Regulatory Operating Fee | Varies | `regulatoryOperatingFee` | Some regions only |

### Example: $49/month Subscription (Small Developer, Under $1M)
```
grossAmount:              $49.00
- Revenue Share (0%):     -$0.00
- Processing Fee (2.9%):  -$1.42
- Sales Tax (est 8%):     -$0.11  (on fees only)
─────────────────────────────────
netAmount:                $47.47  (you receive)
```

### Example: $49/month Subscription (Default 20% Tier)
```
grossAmount:              $49.00
- Revenue Share (20%):    -$9.80
- Processing Fee (2.9%):  -$1.42
- Sales Tax (est 8%):     -$0.90  (on fees only)
─────────────────────────────────
netAmount:                $36.88  (you receive)
```

---

## Billing Models

### Understanding Usage Charges

Usage charges are NOT limited to order processing. Apps can charge for ANY measurable action:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        USAGE CHARGE CATEGORIES                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  📦 TRANSACTION-BASED                    💬 COMMUNICATION                    │
│  ────────────────────                    ───────────────                     │
│  • Orders processed                      • SMS messages sent                 │
│  • Checkouts completed                   • Email campaigns                   │
│  • Fulfillments created                  • Push notifications               │
│  • Returns handled                       • WhatsApp messages                │
│                                                                              │
│  🔗 API & INTEGRATIONS                   🤖 AI & AUTOMATION                 │
│  ────────────────────                    ─────────────────                   │
│  • External API calls                    • AI text generation               │
│  • Webhook deliveries                    • Image processing                 │
│  • Third-party syncs                     • Product recommendations          │
│  • Data imports/exports                  • Chatbot interactions             │
│                                                                              │
│  💾 DATA & STORAGE                       📊 REPORTING                        │
│  ─────────────────                       ──────────                          │
│  • GB of storage used                    • Reports generated                │
│  • Files uploaded                        • Analytics queries                │
│  • Backup operations                     • CSV exports                      │
│  • CDN bandwidth                         • PDF invoices created             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Subscription-Only Model
```
┌──────────┐         ┌──────────┐         ┌──────────┐
│ Merchant │ ─$49─── │ Shopify  │ ─$39*── │   You    │
│          │ monthly │ App Store│  ~80%   │(Developer)│
└──────────┘         └──────────┘         └──────────┘
                           │
              keeps revenue share + 2.9% + tax
              *exact amount depends on tier
```

### Usage-Based Model
```
┌──────────┐               ┌──────────┐         ┌──────────┐
│ Merchant │ ─$500──────── │ Shopify  │ ─$400*─ │   You    │
│ (usage   │ usage fees    │ App Store│  ~80%   │(Developer)│
│  events) │               │          │         │          │
└──────────┘               └──────────┘         └──────────┘
                                 │
                    keeps revenue share + 2.9% + tax
```

**Usage Charge Examples (NOT limited to orders):**
| Use Case | Pricing Example | Trigger |
|----------|-----------------|---------|
| Order fulfillment | $0.05/order | Each order processed |
| SMS/Messaging | $0.02/message | Each SMS/notification sent |
| API calls | $0.001/call | External API usage |
| AI features | $0.10/generation | AI text/image generation |
| Data exports | $0.50/export | CSV/report downloads |
| Storage | $0.10/GB/month | Data storage overage |
| Third-party integrations | $0.05/sync | External service calls |

### Hybrid Model (Subscription + Usage)
```
┌──────────┐               ┌──────────────┐         ┌──────────┐
│ Merchant │               │   Shopify    │         │   You    │
│          │ ─$29 sub───── │   App Store  │         │(Developer)│
│ (high    │               │              │ ─$423*─ │          │
│  usage)  │ ─$500 usage── │ keeps fees   │  ~80%   │          │
│          │               │              │         │          │
└──────────┘               └──────────────┘         └──────────┘
                Total: $529
                *exact amount depends on tier + fees
```

**Hybrid Model Examples:**
| Base Plan | Usage Component | Total Example |
|-----------|-----------------|---------------|
| $29/mo Pro Plan | + $0.05/order over 1,000 | $29 + $500 usage = $529 |
| $49/mo Business | + $0.02/SMS sent | $49 + $200 usage = $249 |
| $99/mo Enterprise | + $0.10/AI generation | $99 + $1,000 usage = $1,099 |

---

## Payout Timing (IMPORTANT)

### When Do Earnings Appear?

| Charge Type | Time to Appear in Partner Dashboard |
|-------------|-------------------------------------|
| RecurringApplicationCharge | Up to **37 days** from merchant acceptance |
| ApplicationCharge (one-time) | Within **7 days** |
| UsageCharge | After merchant pays Shopify invoice |

**Key Insight:** Earnings appear AFTER the merchant pays their Shopify invoice, not when they accept the charge.

### Payout Schedule
- Exact calendar schedule (bi-weekly dates) in Partner Help Center
- Minimum payout threshold: Check Partner Help Center
- Payout currency depends on your account settings

```
┌─────────────────────────────────────────────────────────────────────┐
│                      EARNINGS TIMELINE                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Day 0        Day 1-30           Day 30-37        Day 37+           │
│    │              │                  │               │               │
│    ▼              ▼                  ▼               ▼               │
│  Merchant     Merchant's          Merchant        Earning           │
│  accepts      billing cycle       pays their      appears in        │
│  charge       (30 days)           invoice         Partner Dashboard │
│                                                                      │
│  [────────── Up to 37 days for recurring charges ──────────]        │
│                                                                      │
│  [─── 7 days for one-time charges ───]                              │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Partner API & Payout System

### Payout Lifecycle
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          PARTNER PAYOUT LIFECYCLE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. MERCHANT ACTION                                                          │
│     └── Installs app, subscribes, uses features                             │
│                                                                              │
│  2. SHOPIFY BILLS MERCHANT                                                   │
│     └── Added to merchant's Shopify invoice                                 │
│                                                                              │
│  3. MERCHANT PAYS INVOICE                                                    │
│     └── Up to 37 days for recurring, 7 days for one-time                   │
│                                                                              │
│  4. TRANSACTION CREATED                                                      │
│     ├── grossAmount: What merchant paid                                     │
│     ├── shopifyFee: Revenue share (0%/15%/20%)                              │
│     ├── processingFee: 2.9%                                                 │
│     ├── regulatoryOperatingFee: (if applicable)                             │
│     └── netAmount: What YOU receive                                         │
│                                                                              │
│  5. PARTNER BALANCE ACCRUES                                                  │
│     └── netAmount adds to your Partner balance                              │
│                                                                              │
│  6. PAYOUT SCHEDULED (bi-weekly)                                            │
│     ├── Shopify aggregates net amounts                                      │
│     └── TaxTransaction rolled up once per payout                            │
│                                                                              │
│  7. PAYOUT EXECUTED                                                          │
│     ├── Status: scheduled → in_transit → paid (or failed)                   │
│     └── Transferred to your bank/PayPal                                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Partner API Transaction Types

| TransactionType | Description | Key Fields | Effect on Payout |
|-----------------|-------------|------------|------------------|
| `APP_SUBSCRIPTION_SALE` | Recurring subscription | grossAmount, shopifyFee, processingFee, netAmount | +netAmount |
| `APP_USAGE_SALE` | Usage-based charge | grossAmount, shopifyFee, processingFee, netAmount | +netAmount |
| `APP_ONE_TIME_SALE` | One-time purchase | grossAmount, shopifyFee, processingFee, netAmount | +netAmount |
| `APP_SALE_ADJUSTMENT` | Refunds, downgrades, chargebacks | grossAmount, shopifyFee, netAmount | -netAmount |
| `APP_SALE_CREDIT` | Credits applied | amount | -amount |
| `TAX_TRANSACTION` | Tax on Shopify's fees (1 per payout) | amount, type | ±amount |
| `THEME_SALE` | Theme purchase | grossAmount, shopifyFee, netAmount | +netAmount |
| `REFERRAL_TRANSACTION` | Store referral commission | amount, category | +amount |

### Transaction Object Structure
```graphql
# Example: APP_SUBSCRIPTION_SALE
{
  id: "gid://partners/AppSubscriptionSale/123456"
  createdAt: "2026-03-01T10:30:00Z"
  __typename: "AppSubscriptionSale"

  app { id, name }
  shop { id, myshopifyDomain }

  grossAmount {              # What merchant paid
    amount: "49.00"
    currencyCode: "USD"
  }
  shopifyFee {               # Revenue share (0%/15%/20%)
    amount: "9.80"
    currencyCode: "USD"
  }
  processingFee {            # 2.9% processing fee
    amount: "1.42"
    currencyCode: "USD"
  }
  regulatoryOperatingFee {   # Regional fees (if applicable)
    amount: "0.00"
    currencyCode: "USD"
  }
  netAmount {                # What YOU receive
    amount: "37.78"
    currencyCode: "USD"
  }
}
```

---

## TAX_TRANSACTION Explained

**Important:** TaxTransaction is tax on SHOPIFY'S FEES, not on your gross revenue.

```
┌─────────────────────────────────────────────────────────────────────┐
│                      TAX TRANSACTION TYPES                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  FOR SALES (App Subscriptions, Theme Sales):                        │
│  ─────────────────────────────────────────────                      │
│  Tax on Shopify's brokerage fee                                     │
│  Amount: NEGATIVE (reduces your payout)                             │
│  Example: -$0.90 (8% tax on $11.22 fees)                            │
│                                                                      │
│  FOR REFERRALS (Store Referral Commissions):                        │
│  ─────────────────────────────────────────────                      │
│  Tax on commission Shopify pays you                                 │
│  Amount: POSITIVE (adds to your payout)                             │
│  Example: +$2.40 (tax included in referral)                         │
│                                                                      │
│  TIMING: Rolled up ONCE per payout (not per transaction)           │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Edge Cases

### Failed Payments
- If merchant never pays Shopify invoice → **No transaction created**
- You don't see failed payment attempts in Partner API
- Check merchant's subscription status via Admin API (`frozen`, `pending`, `declined`)

### Prorated Charges
| Event | What Happens |
|-------|--------------|
| **Upgrade mid-cycle** | `AppSubscriptionSale` with prorated new charge amount |
| **Downgrade mid-cycle** | `AppSaleAdjustment` or `AppSaleCredit` with negative amount (credit) |

### Free Trials
- **During trial:** No transaction (merchant not billed)
- **After trial converts:** `AppSubscriptionSale` when first paid billing cycle starts
- Note: First charge may include prorated credit for trial days

### Chargebacks
- No separate `CHARGEBACK` type in Partner API
- Appears as `AppSaleAdjustment` with negative netAmount
- Associated `TaxTransaction` adjustments also reversed

---

## Usage-Based Billing Details

### What Are Usage Charges?

Usage charges allow apps to bill merchants based on **actual consumption** rather than a flat fee. This aligns developer revenue with merchant value.

**Common Usage Charge Types:**

| Category | Examples | Typical Pricing |
|----------|----------|-----------------|
| **Transaction-based** | Orders processed, checkouts, fulfillments | $0.01 - $0.10 per event |
| **Communication** | SMS sent, emails, push notifications | $0.01 - $0.05 per message |
| **API/Integration** | External API calls, webhooks, syncs | $0.001 - $0.01 per call |
| **AI/ML Features** | Text generation, image processing, recommendations | $0.05 - $0.50 per use |
| **Data Operations** | Reports generated, exports, imports | $0.10 - $1.00 per operation |
| **Storage** | GB of data stored, files uploaded | $0.05 - $0.20 per GB/month |
| **Third-party costs** | External service pass-through (shipping rates, translations) | Variable |

### How Usage Charges Work
```
┌─────────────────────────────────────────────────────────────────────┐
│                     USAGE CHARGE LIFECYCLE                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  1. USAGE EVENT OCCURS                                              │
│     └── Order processed, SMS sent, AI call made, etc.              │
│                                                                      │
│  2. YOUR APP creates usage charge                                   │
│     POST /recurring_application_charges/{id}/usage_charges.json    │
│     OR: appUsageRecordCreate (GraphQL)                              │
│     Include: description, price, quantity (optional)               │
│                                                                      │
│  3. SHOPIFY adds to merchant's next invoice                         │
│     (Not billed immediately - batched with invoice)                 │
│                                                                      │
│  4. MERCHANT pays Shopify invoice                                   │
│                                                                      │
│  5. APP_USAGE_SALE transaction appears in Partner API               │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Usage Charge Best Practices

| Practice | Description |
|----------|-------------|
| **Clear descriptions** | "50 SMS messages sent" not just "$2.50 charge" |
| **Batch small charges** | Combine multiple $0.01 calls into periodic summaries |
| **Transparent pricing** | Show usage dashboard in app UI |
| **Usage alerts** | Notify merchants approaching their cap |

### Capped Amount (IMPORTANT)
- **Required:** Must define `capped_amount` on parent subscription
- **Per 30-day billing cycle:** Resets each cycle
- **Enforced by Shopify:** Attempts over cap return 422 error

```json
HTTP/1.1 422 Unprocessable Entity
{
  "errors": {
    "base": ["Total price exceeds balance remaining"]
  }
}
```

- **Merchant can modify cap** from their admin
- Listen for changes via `APP_SUBSCRIPTIONS_UPDATE` webhook

### No Per-Charge Maximum
- Only constraint is the capped_amount per billing cycle
- No documented per-charge limit (but be reasonable)

---

## Visual Requirements

### Layout
```
╔══════════════════════════════════════════════════════════════╗
║                    TRANSACTION FLOW                          ║
║              (Merchant's Product Revenue)                    ║
║                                                              ║
║   [Customer] ──$100──▶ [Shopify] ──$97──▶ [Merchant]        ║
║                         keeps 3%                             ║
║                     NOT YOUR MONEY                           ║
╠══════════════════════════════════════════════════════════════╣
║                      APP REVENUE FLOW                        ║
║               (Your Developer Revenue)                       ║
║                                                              ║
║   [Merchant] ──$49──▶ [Shopify] ──$37*──▶ [You]             ║
║                   keeps fees + tax                           ║
║                     YOUR MONEY                               ║
║              *varies by tier (0%/15%/20%)                    ║
╚══════════════════════════════════════════════════════════════╝
```

### Interactive Elements
- Toggle between billing models (Subscription/Usage/Hybrid)
- Toggle between revenue share tiers (Default 20% / Reduced 0%+15% / Large 15%)
- **Usage type selector** (when Usage or Hybrid selected):
  - Per Order ($0.05/order)
  - Per SMS ($0.02/message)
  - Per API Call ($0.001/call)
  - Per AI Generation ($0.10/use)
  - Custom (user input)
- Hover tooltips on each entity and fee
- Play/Pause controls
- "Show Both" / "Transaction Only" / "App Revenue Only" tabs

### Color Scheme
- Transaction Flow: Green/teal (merchant money)
- App Revenue Flow: Blue/purple (developer money)
- Fees: Red (deductions)
- Net Amount: Bright green (what you keep)

---

## Key Messages to Convey

1. **"Transactions ≠ Your Revenue"**
   - Customer payments go to merchants
   - You don't get a cut of product sales

2. **"Revenue Share is NOT Always 20%"**
   - 0% on first $1M (if registered + eligible)
   - 15% over $1M (or for large developers)
   - 20% only if not registered

3. **"There Are Multiple Fees"**
   - Revenue share: 0% / 15% / 20%
   - Processing fee: 2.9% (always)
   - Sales tax on fees
   - Regulatory fees (some regions)

4. **"Earnings Take Time"**
   - Up to 37 days for recurring charges
   - 7 days for one-time charges
   - Only after merchant pays their invoice

5. **"Usage Charges Are Flexible"**
   - Not just for orders - ANY measurable action
   - SMS, API calls, AI features, storage, exports
   - capped_amount per 30-day cycle (merchant-adjustable)
   - Align your revenue with merchant value

---

## File Locations
- **Component:** `marketing/site/components/ShopifyMoneyFlow.tsx`
- **Page:** `marketing/site/app/money-flow/page.tsx`
- **View:** http://localhost:3000/money-flow

---

## Example GraphQL Query

```graphql
query PartnerTransactions($first: Int!, $createdAtMin: DateTime, $types: [TransactionType!]) {
  transactions(first: $first, createdAtMin: $createdAtMin, types: $types) {
    edges {
      node {
        id
        createdAt
        __typename

        ... on AppSubscriptionSale {
          app { name }
          shop { myshopifyDomain }
          grossAmount { amount currencyCode }
          shopifyFee { amount currencyCode }
          processingFee { amount currencyCode }
          netAmount { amount currencyCode }
        }

        ... on AppUsageSale {
          app { name }
          shop { myshopifyDomain }
          grossAmount { amount currencyCode }
          shopifyFee { amount currencyCode }
          processingFee { amount currencyCode }
          netAmount { amount currencyCode }
        }

        ... on AppSaleAdjustment {
          app { name }
          shop { myshopifyDomain }
          grossAmount { amount currencyCode }
          netAmount { amount currencyCode }
        }

        ... on TaxTransaction {
          amount { amount currencyCode }
          type
        }
      }
    }
    pageInfo { hasNextPage }
  }
}
```

---

## Payout Reconciliation

```
Partner Dashboard Payout              Partner API Transactions
─────────────────────────             ────────────────────────
Payout Date: 2026-02-15               SUM(netAmount) for period
Net Amount: $4,234.50                 ────────────────────────
                                      APP_SUBSCRIPTION_SALE: $3,500.00
                        ═══════       APP_USAGE_SALE: $1,000.00
                        SHOULD        APP_SALE_ADJUSTMENT: -$200.00
                        MATCH         TAX_TRANSACTION: -$65.50
                        ═══════       ─────────────────────────
                                      TOTAL: $4,234.50 ✓
```

### Rate Limits
- 4 requests/second per Partner API client
- Returns `{"errors": [{"message": "Too many requests"}]}` when throttled

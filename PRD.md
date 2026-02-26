# PRD – LedgerGuard

## 1. Product Vision

LedgerGuard is a Revenue Intelligence Platform for Shopify App Developers.

It helps developers:
- Monitor Renewal Success Rate
- Detect missed billing cycles before churn
- Track Revenue at Risk in real-time
- Separate Recurring and Usage revenue streams
- Receive AI-powered Daily Revenue Briefs (Pro tier)

---

## 2. Target Users

**Primary:**
- Shopify app developers using Shopify Partner API
- Solo developers and small teams managing subscription apps

**Secondary:**
- Agencies managing multiple Shopify apps
- Finance teams needing revenue reconciliation

---

## 3. Core Value Proposition

> "Know immediately if your subscription revenue is stable."

**Key Differentiator:** Proactive churn detection based on billing cycle analysis, not reactive reporting after revenue is lost.

---

## 4. Core KPIs

| KPI | Definition | Priority |
|-----|------------|----------|
| Renewal Success Rate | (Renewed subscriptions / Expected renewals) × 100 | P0 |
| Active MRR | Sum of all ACTIVE subscription amounts | P0 |
| Revenue at Risk | MRR from ONE_CYCLE_MISSED + TWO_CYCLE_MISSED states | P0 |
| Usage Revenue | Non-recurring charges (usage billing) | P1 |
| Total Revenue | Recurring + Usage revenue | P1 |

---

## 5. Functional Requirements

### 5.1 Authentication
- Firebase Authentication
- Google OAuth login
- Email/password login
- Role-based access control:
  - **OWNER**: Full access, billing, delete workspace
  - **ADMIN**: Manage integrations, view all data
  - **VIEWER**: Read-only dashboard access (future)

### 5.2 Partner Integration

**Shopify Partner OAuth (Primary):**
- OAuth 2.0 flow with Shopify Partners
- Scopes: `read_financials`, `read_apps`
- Automatic token refresh

**Manual Partner Token (ADMIN only):**
- For development/testing
- Encrypted storage (AES-256)
- Token validation on save

**Post-Connection:**
- Fetch list of apps from Partner API
- User selects which app(s) to track
- Store app_id mapping per workspace

### 5.3 Sync Engine

**Schedule:**
- 12-hour batch sync (00:00, 12:00 UTC)
- On-demand sync via API/UI

**Behavior:**
- Recalculate full 12-month rolling window every sync
- Idempotent: same input → same output
- Background historical backfill on first connection

**Data Fetched:**
- `transactions` (Partner API) - all app earnings
- `appSubscription` events
- Pagination handled automatically

### 5.4 Ledger Engine

**Principles:**
- Deterministic rebuild from raw transactions
- Immutable source records
- Computed fields clearly separated

**Revenue Classification:**
| Type | Source | Examples |
|------|--------|----------|
| RECURRING | AppSubscriptionSale | Monthly/annual plans |
| USAGE | AppUsageSale | Usage-based charges |
| ONE_TIME | AppOneTimeSale | Setup fees, add-ons |
| REFUND | AppRefund | Negative adjustment |

**Expected Renewal Calculation:**
```
For each ACTIVE subscription:
  expected_renewal_date = current_period_end
  expected_amount = plan_price × (1 - shopify_cut)
```

### 5.5 Risk Engine

**States (Authoritative):**

| State | Condition | Action |
|-------|-----------|--------|
| SAFE | Status = ACTIVE | No action |
| SAFE | Days past period_end ≤ 30 | Grace period |
| ONE_CYCLE_MISSED | 31–60 days past period_end | Yellow alert |
| TWO_CYCLE_MISSED | 61–90 days past period_end | Red alert |
| CHURNED | >90 days past period_end | Mark as lost |

**Implementation:**
```go
func ClassifyRisk(sub *Subscription, now time.Time) RiskState {
    if sub.Status == ACTIVE {
        return SAFE
    }
    daysLate := daysSince(sub.CurrentPeriodEnd, now)
    switch {
    case daysLate <= 30:
        return SAFE
    case daysLate <= 60:
        return ONE_CYCLE_MISSED
    case daysLate <= 90:
        return TWO_CYCLE_MISSED
    default:
        return CHURNED
    }
}
```

### 5.6 KPI Engine

**Compute Daily:**

```sql
-- Renewal Success Rate
SELECT
  COUNT(CASE WHEN renewed = true THEN 1 END)::float /
  COUNT(*)::float * 100 AS renewal_rate
FROM expected_renewals
WHERE expected_date BETWEEN :start AND :end;

-- Active MRR
SELECT SUM(price_cents) FROM subscriptions WHERE status = 'ACTIVE';

-- Revenue at Risk
SELECT SUM(price_cents) FROM subscriptions
WHERE risk_state IN ('ONE_CYCLE_MISSED', 'TWO_CYCLE_MISSED');
```

### 5.7 Snapshot Engine

**Purpose:** Immutable daily metrics for historical analysis

**Schema:**
```sql
CREATE TABLE daily_metrics_snapshot (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    renewal_success_rate DECIMAL(5,2),
    active_mrr_cents BIGINT,
    at_risk_mrr_cents BIGINT,
    usage_revenue_cents BIGINT,
    total_revenue_cents BIGINT,
    safe_count INT,
    one_cycle_missed_count INT,
    two_cycle_missed_count INT,
    churned_count INT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(workspace_id, snapshot_date)
);
```

**Retention:** Permanent (never delete)

### 5.8 AI Insight Engine (Pro Tier)

**Input:** Structured JSON snapshot
```json
{
  "today": { "mrr": 5000, "at_risk": 200, "renewal_rate": 94.5 },
  "7_days_ago": { "mrr": 4800, "at_risk": 150, "renewal_rate": 96.0 },
  "changes": {
    "new_at_risk": ["store-a.myshopify.com"],
    "recovered": ["store-b.myshopify.com"]
  }
}
```

**Output:** 80–120 word summary
- Highlight significant changes
- Call out stores needing attention
- Positive reinforcement when stable

**Delivery:** Daily at 08:00 user timezone

### 5.9 Notifications

**Critical Alerts (Immediate):**
- State change: SAFE → ONE_CYCLE_MISSED
- State change: ONE_CYCLE_MISSED → TWO_CYCLE_MISSED
- Large revenue at risk (>10% of MRR)

**Daily Summary:**
- Morning brief with KPIs
- Optional: AI insight (Pro)

**Channels:**
- Email (default)
- Slack webhook (Pro)
- In-app notifications

**User Configurable:**
- Enable/disable per alert type
- Quiet hours
- Threshold customization

---

## 6. Non-Functional Requirements

| Requirement | Specification |
|-------------|---------------|
| **Determinism** | Same input always produces same output |
| **Security** | Tokens encrypted at rest (AES-256-GCM) |
| **Access Control** | Role-based, workspace-scoped |
| **Auditability** | All state changes logged with timestamp |
| **Architecture** | Modular monolith (Go backend) |
| **Database** | PostgreSQL 15+ |
| **Availability** | 99.9% uptime target |
| **Latency** | Dashboard load <2s, API p95 <500ms |

---

## 7. Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.22+, Clean Architecture |
| Database | PostgreSQL, pgcrypto |
| Cache | Redis (session, rate limiting) |
| Auth | Firebase Authentication |
| Frontend | Flutter Web |
| State | Riverpod |
| Hosting | Cloud Run / Fly.io |
| CI/CD | GitHub Actions |

---

## 8. API Endpoints (v1)

### Auth
- `POST /api/v1/auth/login` - Firebase token exchange
- `GET /api/v1/auth/me` - Current user info

### Workspace
- `POST /api/v1/workspaces` - Create workspace
- `GET /api/v1/workspaces/:id` - Get workspace

### Integration
- `GET /api/v1/integrations/shopify/oauth` - Start OAuth
- `GET /api/v1/integrations/shopify/callback` - OAuth callback
- `POST /api/v1/integrations/shopify/token` - Manual token (ADMIN)
- `GET /api/v1/integrations/shopify/apps` - List connected apps

### Sync
- `POST /api/v1/sync` - Trigger manual sync
- `GET /api/v1/sync/status` - Sync status

### Dashboard
- `GET /api/v1/dashboard/kpis` - Current KPIs
- `GET /api/v1/dashboard/risk-summary` - Risk breakdown
- `GET /api/v1/dashboard/trends` - Historical trends

### Subscriptions
- `GET /api/v1/subscriptions` - List with filters
- `GET /api/v1/subscriptions/:id` - Single subscription

### Payouts
- `GET /api/v1/payouts` - List transactions
- `GET /api/v1/payouts/reconcile` - Reconciliation report

---

## 9. Data Model

```
workspace
├── id, name, owner_id, created_at
│
├── workspace_members (user_id, role)
│
├── integrations
│   └── shopify_partner (org_id, encrypted_token, app_ids[])
│
├── subscriptions
│   └── id, shopify_gid, shop_domain, plan, price, status, risk_state, period_end
│
├── transactions (immutable)
│   └── id, shopify_gid, type, amount, shop_domain, app_name, transaction_date
│
└── daily_metrics_snapshot
    └── date, mrr, at_risk, renewal_rate, counts...
```

---

## 10. Milestones

| Phase | Scope | Target |
|-------|-------|--------|
| MVP | Auth, Shopify sync, Risk engine, Dashboard | Week 1-2 |
| Beta | KPI engine, Snapshots, Email alerts | Week 3-4 |
| Launch | AI insights, Slack, Polish | Week 5-6 |
| Growth | Multi-app, Forecasting | Post-launch |

---

## 11. Future Features

Tracked in `future.md`:
- Multi-app support per workspace
- Revenue forecast modeling
- Anomaly detection (ML)
- Affiliate/referral program
- Native mobile app (iOS/Android)
- Stripe integration for non-Shopify revenue
- Team collaboration features
- Custom report builder

---

## 12. Success Metrics

| Metric | Target |
|--------|--------|
| Time to first sync | <5 minutes |
| Daily active users | 60% of registered |
| Churn prediction accuracy | >85% |
| Alert response rate | >70% |
| NPS | >50 |

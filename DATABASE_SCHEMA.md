# Database Schema – LedgerGuard

## Entity Relationship Diagram

```
users
  │
  ├──< partner_accounts
  │         │
  │         └──< apps
  │               │
  │               ├──< transactions
  │               │
  │               ├──< subscriptions
  │               │
  │               ├──< daily_metrics_snapshot
  │               │
  │               └──< daily_insight
  │
  ├──< device_tokens
  │
  └──< notification_preferences
```

---

## Tables

### users
Primary user account (linked to Firebase).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, DEFAULT gen_random_uuid() | Internal user ID |
| firebase_uid | VARCHAR(128) | UNIQUE, NOT NULL | Firebase Auth UID |
| email | VARCHAR(255) | NOT NULL | User email |
| role | VARCHAR(20) | NOT NULL, CHECK (OWNER, ADMIN) | User role |
| plan_tier | VARCHAR(20) | DEFAULT 'FREE' | FREE / PRO |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Account creation |

### partner_accounts
Shopify Partner API connections.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Partner account ID |
| user_id | UUID | FK → users.id, NOT NULL | Owner |
| integration_type | VARCHAR(20) | NOT NULL, CHECK (OAUTH, MANUAL) | How connected |
| partner_id | VARCHAR(100) | NOT NULL | Shopify Partner org ID |
| encrypted_access_token | BYTEA | NOT NULL | AES-256-GCM encrypted |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Connection date |

### apps
Shopify apps being tracked.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | App ID |
| partner_account_id | UUID | FK → partner_accounts.id, NOT NULL | Parent account |
| partner_app_id | VARCHAR(100) | NOT NULL | Shopify app GID |
| name | VARCHAR(255) | NOT NULL | App name |
| tracking_enabled | BOOLEAN | DEFAULT TRUE | Active tracking |

### transactions
Immutable ledger of all Partner API transactions.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Transaction ID |
| app_id | UUID | FK → apps.id, NOT NULL | Parent app |
| shopify_gid | VARCHAR(255) | UNIQUE, NOT NULL | Shopify transaction GID |
| myshopify_domain | VARCHAR(255) | NOT NULL | Store domain |
| charge_type | VARCHAR(20) | NOT NULL, CHECK (RECURRING, USAGE, ONE_TIME, REFUND) | Revenue type |
| amount_cents | BIGINT | NOT NULL | Amount in cents |
| currency | VARCHAR(3) | DEFAULT 'USD' | Currency code |
| transaction_date | TIMESTAMPTZ | NOT NULL | When charged |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Record creation |

### subscriptions
Current state of each subscription.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Subscription ID |
| app_id | UUID | FK → apps.id, NOT NULL | Parent app |
| shopify_gid | VARCHAR(255) | UNIQUE, NOT NULL | Shopify subscription GID |
| myshopify_domain | VARCHAR(255) | NOT NULL | Store domain |
| plan_name | VARCHAR(255) | | Plan name |
| base_price_cents | BIGINT | NOT NULL | Price in cents |
| currency | VARCHAR(3) | DEFAULT 'USD' | Currency code |
| billing_interval | VARCHAR(20) | DEFAULT 'MONTHLY' | MONTHLY / ANNUAL |
| status | VARCHAR(20) | NOT NULL | ACTIVE, CANCELLED, FROZEN, PENDING |
| last_recurring_charge_date | TIMESTAMPTZ | | Last successful charge |
| expected_next_charge_date | TIMESTAMPTZ | | Next expected charge |
| risk_state | VARCHAR(30) | NOT NULL | SAFE, ONE_CYCLE_MISSED, TWO_CYCLE_MISSED, CHURNED |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | First seen |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last updated |

### daily_metrics_snapshot
Immutable daily KPI snapshots.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Snapshot ID |
| app_id | UUID | FK → apps.id, NOT NULL | Parent app |
| date | DATE | NOT NULL | Snapshot date |
| active_mrr_cents | BIGINT | | Monthly recurring revenue |
| renewal_success_rate | DECIMAL(5,2) | | Percentage (0-100) |
| revenue_at_risk_cents | BIGINT | | At-risk MRR |
| usage_revenue_cents | BIGINT | | Usage-based revenue |
| total_revenue_cents | BIGINT | | Total revenue |
| safe_count | INT | | Subscriptions in SAFE |
| one_cycle_missed_count | INT | | ONE_CYCLE_MISSED count |
| two_cycle_missed_count | INT | | TWO_CYCLE_MISSED count |
| churned_count | INT | | CHURNED count |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Record creation |
| UNIQUE | | (app_id, date) | One snapshot per app per day |

### daily_insight
AI-generated daily summaries (Pro tier).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Insight ID |
| app_id | UUID | FK → apps.id, NOT NULL | Parent app |
| date | DATE | NOT NULL | Insight date |
| insight_text | TEXT | NOT NULL | AI-generated summary (80-120 words) |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Generation time |
| UNIQUE | | (app_id, date) | One insight per app per day |

### device_tokens
Push notification tokens.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Token ID |
| user_id | UUID | FK → users.id, NOT NULL | Owner |
| device_token | VARCHAR(500) | NOT NULL | FCM/APNs token |
| platform | VARCHAR(20) | NOT NULL | ios, android, web |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Registration time |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last refresh |

### notification_preferences
User notification settings.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Preference ID |
| user_id | UUID | FK → users.id, UNIQUE, NOT NULL | Owner (one per user) |
| critical_enabled | BOOLEAN | DEFAULT TRUE | Risk state change alerts |
| daily_summary_enabled | BOOLEAN | DEFAULT TRUE | Daily summary email |
| daily_summary_time | TIME | DEFAULT '08:00' | Local time for summary |
| slack_webhook_url | VARCHAR(500) | | Slack integration (Pro) |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | Creation time |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | Last modified |

---

## SQL DDL

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid VARCHAR(128) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('OWNER', 'ADMIN')),
    plan_tier VARCHAR(20) DEFAULT 'FREE' CHECK (plan_tier IN ('FREE', 'PRO')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- partner_accounts
CREATE TABLE partner_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    integration_type VARCHAR(20) NOT NULL CHECK (integration_type IN ('OAUTH', 'MANUAL')),
    partner_id VARCHAR(100) NOT NULL,
    encrypted_access_token BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- apps
CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_account_id UUID NOT NULL REFERENCES partner_accounts(id) ON DELETE CASCADE,
    partner_app_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    tracking_enabled BOOLEAN DEFAULT TRUE,
    UNIQUE (partner_account_id, partner_app_id)
);

-- transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    shopify_gid VARCHAR(255) UNIQUE NOT NULL,
    myshopify_domain VARCHAR(255) NOT NULL,
    charge_type VARCHAR(20) NOT NULL CHECK (charge_type IN ('RECURRING', 'USAGE', 'ONE_TIME', 'REFUND')),
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    shopify_gid VARCHAR(255) UNIQUE NOT NULL,
    myshopify_domain VARCHAR(255) NOT NULL,
    plan_name VARCHAR(255),
    base_price_cents BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_interval VARCHAR(20) DEFAULT 'MONTHLY' CHECK (billing_interval IN ('MONTHLY', 'ANNUAL')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'CANCELLED', 'FROZEN', 'PENDING')),
    last_recurring_charge_date TIMESTAMPTZ,
    expected_next_charge_date TIMESTAMPTZ,
    risk_state VARCHAR(30) NOT NULL CHECK (risk_state IN ('SAFE', 'ONE_CYCLE_MISSED', 'TWO_CYCLE_MISSED', 'CHURNED')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- daily_metrics_snapshot
CREATE TABLE daily_metrics_snapshot (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    active_mrr_cents BIGINT,
    renewal_success_rate DECIMAL(5,2),
    revenue_at_risk_cents BIGINT,
    usage_revenue_cents BIGINT,
    total_revenue_cents BIGINT,
    safe_count INT,
    one_cycle_missed_count INT,
    two_cycle_missed_count INT,
    churned_count INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (app_id, date)
);

-- daily_insight
CREATE TABLE daily_insight (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    insight_text TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (app_id, date)
);

-- device_tokens
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(500) NOT NULL,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- notification_preferences
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    critical_enabled BOOLEAN DEFAULT TRUE,
    daily_summary_enabled BOOLEAN DEFAULT TRUE,
    daily_summary_time TIME DEFAULT '08:00',
    slack_webhook_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

---
-- INDEXES
---

-- transactions: high-volume queries
CREATE INDEX idx_transactions_app_date ON transactions(app_id, transaction_date DESC);
CREATE INDEX idx_transactions_domain ON transactions(app_id, myshopify_domain);
CREATE INDEX idx_transactions_type ON transactions(app_id, charge_type);

-- subscriptions: risk queries
CREATE INDEX idx_subscriptions_app_status ON subscriptions(app_id, status);
CREATE INDEX idx_subscriptions_app_risk ON subscriptions(app_id, risk_state);
CREATE INDEX idx_subscriptions_domain ON subscriptions(myshopify_domain);

-- snapshots: time-series queries
CREATE INDEX idx_snapshots_app_date ON daily_metrics_snapshot(app_id, date DESC);

-- device_tokens: push notifications
CREATE INDEX idx_device_tokens_user ON device_tokens(user_id);

---
-- TRIGGERS
---

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER device_tokens_updated_at
    BEFORE UPDATE ON device_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
```

---

## Notes

1. **Immutability:** `transactions` and `daily_metrics_snapshot` are append-only
2. **Encryption:** `encrypted_access_token` uses AES-256-GCM with app-level master key
3. **Soft Delete:** Not implemented; use `tracking_enabled` for apps
4. **Retention:** Transactions kept for 12 months; snapshots kept permanently
5. **Timezone:** All timestamps in UTC (TIMESTAMPTZ)

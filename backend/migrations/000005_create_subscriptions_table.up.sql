-- Create subscriptions table
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
    risk_state VARCHAR(30) NOT NULL CHECK (risk_state IN ('SAFE', 'ONE_CYCLE_MISSED', 'TWO_CYCLES_MISSED', 'CHURNED')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for app and status queries
CREATE INDEX idx_subscriptions_app_status ON subscriptions(app_id, status);

-- Index for risk state queries
CREATE INDEX idx_subscriptions_app_risk ON subscriptions(app_id, risk_state);

-- Index for domain lookups
CREATE INDEX idx_subscriptions_domain ON subscriptions(myshopify_domain);

-- Index for expected charge date (for identifying at-risk subscriptions)
CREATE INDEX idx_subscriptions_expected_charge ON subscriptions(app_id, expected_next_charge_date);

-- Auto-update updated_at trigger
CREATE TRIGGER subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

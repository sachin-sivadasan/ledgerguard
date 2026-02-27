-- Create api_subscription_status table (CQRS read model)
CREATE TABLE api_subscription_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shopify_gid VARCHAR(255) NOT NULL,
    app_id UUID NOT NULL,
    myshopify_domain VARCHAR(255) NOT NULL,
    shop_name VARCHAR(255),
    plan_name VARCHAR(255),
    risk_state VARCHAR(30) NOT NULL CHECK (risk_state IN ('SAFE', 'ONE_CYCLE_MISSED', 'TWO_CYCLES_MISSED', 'CHURNED')),
    is_paid_current_cycle BOOLEAN NOT NULL,
    months_overdue INT NOT NULL DEFAULT 0 CHECK (months_overdue >= 0),
    last_successful_charge_date TIMESTAMPTZ,
    expected_next_charge_date TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'CANCELLED', 'FROZEN', 'PENDING')),
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique index on Shopify GID (primary lookup)
CREATE UNIQUE INDEX idx_api_sub_status_gid ON api_subscription_status(shopify_gid);

-- Index for app-level queries
CREATE INDEX idx_api_sub_status_app ON api_subscription_status(app_id);

-- Index for risk state filtering
CREATE INDEX idx_api_sub_status_risk ON api_subscription_status(app_id, risk_state);

-- Index for domain lookups
CREATE INDEX idx_api_sub_status_domain ON api_subscription_status(myshopify_domain);

-- Index for sync timestamp
CREATE INDEX idx_api_sub_status_sync ON api_subscription_status(last_synced_at);

-- Comment on table
COMMENT ON TABLE api_subscription_status IS 'CQRS read model for subscription payment status. Populated after ledger sync.';
COMMENT ON COLUMN api_subscription_status.shopify_gid IS 'Shopify GraphQL ID (e.g., gid://shopify/AppSubscription/123)';
COMMENT ON COLUMN api_subscription_status.is_paid_current_cycle IS 'TRUE if status is ACTIVE and not overdue';
COMMENT ON COLUMN api_subscription_status.months_overdue IS 'Number of billing cycles without successful payment';

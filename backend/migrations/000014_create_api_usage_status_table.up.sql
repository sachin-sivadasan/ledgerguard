-- Create api_usage_status table (CQRS read model)
CREATE TABLE api_usage_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shopify_gid VARCHAR(255) NOT NULL,
    subscription_shopify_gid VARCHAR(255) NOT NULL,
    subscription_id UUID NOT NULL,
    billed BOOLEAN NOT NULL DEFAULT FALSE,
    billing_date TIMESTAMPTZ,
    amount_cents INT NOT NULL CHECK (amount_cents >= 0),
    description TEXT,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique index on Shopify GID (primary lookup)
CREATE UNIQUE INDEX idx_api_usage_status_gid ON api_usage_status(shopify_gid);

-- Index for subscription lookups
CREATE INDEX idx_api_usage_status_subscription ON api_usage_status(subscription_id);

-- Index for subscription GID lookups (for nested queries)
CREATE INDEX idx_api_usage_status_sub_gid ON api_usage_status(subscription_shopify_gid);

-- Index for billed status filtering
CREATE INDEX idx_api_usage_status_billed ON api_usage_status(subscription_id, billed);

-- Index for sync timestamp
CREATE INDEX idx_api_usage_status_sync ON api_usage_status(last_synced_at);

-- Comment on table
COMMENT ON TABLE api_usage_status IS 'CQRS read model for usage billing status. Populated after ledger sync.';
COMMENT ON COLUMN api_usage_status.shopify_gid IS 'Shopify GraphQL ID (e.g., gid://shopify/AppUsageRecord/456)';
COMMENT ON COLUMN api_usage_status.subscription_shopify_gid IS 'Parent subscription Shopify GID for nested lookups';
COMMENT ON COLUMN api_usage_status.billed IS 'TRUE if Shopify has billed this usage charge';

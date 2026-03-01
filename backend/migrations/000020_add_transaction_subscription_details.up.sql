-- Add shop and subscription details to transactions table
-- These fields come from the expanded GraphQL query per Shopify Partner API docs

-- Shop details
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS shopify_shop_gid VARCHAR(255);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS shop_plan VARCHAR(100);

-- Subscription reference (for AppSubscriptionSale transactions)
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS subscription_gid VARCHAR(255);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(50);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS subscription_period_end TIMESTAMP WITH TIME ZONE;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS billing_interval VARCHAR(20);

-- Index for faster subscription lookups
CREATE INDEX IF NOT EXISTS idx_transactions_subscription_gid ON transactions(subscription_gid) WHERE subscription_gid IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_subscription_status ON transactions(subscription_status) WHERE subscription_status IS NOT NULL;

-- Comment on columns
COMMENT ON COLUMN transactions.shopify_shop_gid IS 'Shopify shop GID (gid://shopify/Shop/xxx)';
COMMENT ON COLUMN transactions.shop_plan IS 'Shop Shopify plan (Basic, Shopify, Advanced, Plus)';
COMMENT ON COLUMN transactions.subscription_gid IS 'Shopify subscription GID for AppSubscriptionSale';
COMMENT ON COLUMN transactions.subscription_status IS 'Subscription status (ACTIVE, CANCELLED, FROZEN, etc.)';
COMMENT ON COLUMN transactions.subscription_period_end IS 'Current billing period end date';
COMMENT ON COLUMN transactions.billing_interval IS 'Billing interval (MONTHLY, ANNUAL)';

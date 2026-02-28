-- Add indexes for subscription filtering and search
-- This improves performance for the enhanced subscription list endpoint

-- Composite index for common filter combinations
CREATE INDEX IF NOT EXISTS idx_subscriptions_filters
ON subscriptions(app_id, risk_state, base_price_cents, billing_interval);

-- Index for search by shop_name and domain (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_subscriptions_search_shop_name
ON subscriptions(app_id, lower(shop_name));

CREATE INDEX IF NOT EXISTS idx_subscriptions_search_domain
ON subscriptions(app_id, lower(myshopify_domain));

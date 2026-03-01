-- Add shop GID column to subscriptions for events lookup
ALTER TABLE subscriptions ADD COLUMN shopify_shop_gid VARCHAR(255);

-- Index for shop-based queries
CREATE INDEX idx_subscriptions_shop_gid ON subscriptions(shopify_shop_gid) WHERE shopify_shop_gid IS NOT NULL;

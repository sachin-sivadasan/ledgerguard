-- Add shop_name column to subscriptions table
ALTER TABLE subscriptions ADD COLUMN shop_name VARCHAR(255);

-- Index for shop name searches
CREATE INDEX idx_subscriptions_shop_name ON subscriptions(shop_name);

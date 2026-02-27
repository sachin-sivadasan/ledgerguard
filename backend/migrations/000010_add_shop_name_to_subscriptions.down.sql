-- Remove shop_name column from subscriptions table
DROP INDEX IF EXISTS idx_subscriptions_shop_name;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS shop_name;

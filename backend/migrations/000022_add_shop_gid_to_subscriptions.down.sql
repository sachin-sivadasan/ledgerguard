-- Remove shop GID index and column
DROP INDEX IF EXISTS idx_subscriptions_shop_gid;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS shopify_shop_gid;

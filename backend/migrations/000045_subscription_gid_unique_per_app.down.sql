-- Revert to global uniqueness. NOTE: fails if the same shopify_gid now exists under more
-- than one app_id (which this migration was created to allow) — dedupe first if rolling back.
DROP INDEX IF EXISTS idx_api_sub_status_app_gid;
CREATE UNIQUE INDEX idx_api_sub_status_gid ON api_subscription_status (shopify_gid);

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_app_id_shopify_gid_key;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_shopify_gid_key UNIQUE (shopify_gid);

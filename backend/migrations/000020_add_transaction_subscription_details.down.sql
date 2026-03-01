-- Remove shop and subscription details from transactions table

DROP INDEX IF EXISTS idx_transactions_subscription_status;
DROP INDEX IF EXISTS idx_transactions_subscription_gid;

ALTER TABLE transactions DROP COLUMN IF EXISTS billing_interval;
ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_period_end;
ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_status;
ALTER TABLE transactions DROP COLUMN IF EXISTS subscription_gid;
ALTER TABLE transactions DROP COLUMN IF EXISTS shop_plan;
ALTER TABLE transactions DROP COLUMN IF EXISTS shopify_shop_gid;

-- Revert to the original allowed set. NOTE: this fails if any row still has
-- status='UNINSTALLED' — remap those first (e.g. UPDATE ... SET status='CANCELLED') before
-- rolling back.
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check
    CHECK (status IN ('ACTIVE', 'CANCELLED', 'FROZEN', 'PENDING'));

ALTER TABLE api_subscription_status DROP CONSTRAINT IF EXISTS api_subscription_status_status_check;
ALTER TABLE api_subscription_status ADD CONSTRAINT api_subscription_status_status_check
    CHECK (status IN ('ACTIVE', 'CANCELLED', 'FROZEN', 'PENDING'));

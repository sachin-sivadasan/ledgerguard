DROP INDEX IF EXISTS idx_subscriptions_stable_domain_key;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS stable_domain_key;

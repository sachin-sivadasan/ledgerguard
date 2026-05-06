-- Add stable_domain_key to subscriptions table
-- Deterministic SHA1 of myshopify_domain (lg_sub_...), stable across reinstalls.
-- Used for future churn-return analysis and store lifetime tracking.
ALTER TABLE subscriptions ADD COLUMN stable_domain_key VARCHAR(255);

-- Index for stable domain key lookups
CREATE INDEX idx_subscriptions_stable_domain_key ON subscriptions(stable_domain_key);

-- shopify_gid was GLOBALLY unique on both the subscriptions write model and the
-- api_subscription_status read model. But the same AppSubscription GID can appear under
-- more than one app_id (e.g. two LedgerGuard app records pointing at the same Partner app,
-- or overlapping org apps). The global constraint let one app's ledger rebuild silently
-- clobber another app's row: Upsert ... ON CONFLICT (shopify_gid) DO UPDATE (no app_id in
-- the target/SET) updated the OTHER app's row instead of inserting, with NO error — so a
-- "completed" sync could drop hundreds of real subscriptions. Scope uniqueness per app so
-- each app owns its rows independently.

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_shopify_gid_key;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_app_id_shopify_gid_key UNIQUE (app_id, shopify_gid);

DROP INDEX IF EXISTS idx_api_sub_status_gid;
CREATE UNIQUE INDEX idx_api_sub_status_app_gid ON api_subscription_status (app_id, shopify_gid);

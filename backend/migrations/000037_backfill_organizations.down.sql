-- Reverse backfill: make org_id nullable again, drop indexes, remove backfilled data.

BEGIN;

-- 1. Drop org_id indexes.
DROP INDEX IF EXISTS idx_api_keys_org_id;
DROP INDEX IF EXISTS idx_billing_subscriptions_org_id;
DROP INDEX IF EXISTS idx_partner_accounts_org_id;

-- 2. Make org_id nullable again.
ALTER TABLE api_keys ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE billing_subscriptions ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE partner_accounts ALTER COLUMN org_id DROP NOT NULL;

-- 3. Clear org_id values.
UPDATE api_keys SET org_id = NULL;
UPDATE billing_subscriptions SET org_id = NULL;
UPDATE partner_accounts SET org_id = NULL;

-- 4. Remove backfilled org_members (only auto-created OWNER rows).
DELETE FROM org_members
WHERE role = 'OWNER'
  AND org_id IN (
    SELECT id FROM organizations WHERE slug LIKE 'personal-%'
  );

-- 5. Remove auto-created personal organizations.
DELETE FROM organizations WHERE slug LIKE 'personal-%';

COMMIT;

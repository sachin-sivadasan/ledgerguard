-- Backfill migration: create personal org for each existing user,
-- set org_id on partner_accounts / billing_subscriptions / api_keys,
-- then make org_id NOT NULL.
--
-- This is idempotent: re-running skips users who already have an org.

BEGIN;

-- 1. Create a personal organization for every user who doesn't have one yet.
--    Name = user email, slug = 'personal-' || left 8 chars of user.id,
--    plan_tier copied from user's current plan_tier.
INSERT INTO organizations (id, name, slug, plan_tier, created_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.email,
    'personal-' || LEFT(u.id::text, 8),
    COALESCE(u.plan_tier, 'FREE'),
    u.id,
    COALESCE(u.created_at, NOW()),
    NOW()
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM org_members om WHERE om.user_id = u.id
);

-- 2. Create OWNER membership for each user → their personal org.
INSERT INTO org_members (id, org_id, user_id, role, status, joined_at)
SELECT
    gen_random_uuid(),
    o.id,
    o.created_by,
    'OWNER',
    'ACTIVE',
    NOW()
FROM organizations o
WHERE NOT EXISTS (
    SELECT 1 FROM org_members om
    WHERE om.org_id = o.id AND om.user_id = o.created_by
);

-- 3. Backfill partner_accounts.org_id from user → user's org.
UPDATE partner_accounts pa
SET org_id = (
    SELECT om.org_id
    FROM org_members om
    WHERE om.user_id = pa.user_id
      AND om.role = 'OWNER'
    LIMIT 1
)
WHERE pa.org_id IS NULL;

-- 4. Backfill billing_subscriptions.org_id from user → user's org.
UPDATE billing_subscriptions bs
SET org_id = (
    SELECT om.org_id
    FROM org_members om
    WHERE om.user_id = bs.user_id
      AND om.role = 'OWNER'
    LIMIT 1
)
WHERE bs.org_id IS NULL;

-- 5. Backfill api_keys.org_id from user → user's org.
UPDATE api_keys ak
SET org_id = (
    SELECT om.org_id
    FROM org_members om
    WHERE om.user_id = ak.user_id
      AND om.role = 'OWNER'
    LIMIT 1
)
WHERE ak.org_id IS NULL;

-- 6. Make org_id NOT NULL now that all rows are backfilled.
ALTER TABLE partner_accounts ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE billing_subscriptions ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE api_keys ALTER COLUMN org_id SET NOT NULL;

-- 7. Add indexes for org_id lookups.
CREATE INDEX IF NOT EXISTS idx_partner_accounts_org_id ON partner_accounts(org_id);
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_org_id ON billing_subscriptions(org_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_org_id ON api_keys(org_id);

COMMIT;

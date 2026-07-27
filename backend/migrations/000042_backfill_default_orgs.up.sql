-- Backfill: give every user with NO org membership a default FREE organization and an
-- OWNER membership. Fixes accounts created before signup-time org provisioning existed
-- (a user row could exist with zero orgs, breaking org-scoped routes like /token with
-- "organization selection required"). Only affects users who currently have none, so
-- re-running is safe (no new orgs once every user has a membership).
WITH orphaned AS (
    SELECT u.id AS user_id, u.email
    FROM users u
    WHERE NOT EXISTS (
        SELECT 1 FROM org_members m WHERE m.user_id = u.id
    )
),
new_orgs AS (
    INSERT INTO organizations (name, plan_tier, created_by)
    SELECT
        CASE
            WHEN o.email IS NOT NULL AND position('@' IN o.email) > 1
                THEN split_part(o.email, '@', 1) || '''s Organization'
            ELSE 'My Organization'
        END,
        'FREE',
        o.user_id
    FROM orphaned o
    RETURNING id AS org_id, created_by AS user_id
)
INSERT INTO org_members (org_id, user_id, role, status)
SELECT n.org_id, n.user_id, 'OWNER', 'ACTIVE'
FROM new_orgs n;

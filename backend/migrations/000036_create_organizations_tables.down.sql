-- Reverse migration: drop org_id columns and organization tables

ALTER TABLE api_keys DROP COLUMN IF EXISTS org_id;
ALTER TABLE billing_subscriptions DROP COLUMN IF EXISTS org_id;
ALTER TABLE partner_accounts DROP COLUMN IF EXISTS org_id;

DROP TABLE IF EXISTS org_audit_log;
DROP TABLE IF EXISTS org_invitations;
DROP TABLE IF EXISTS org_members;
DROP TABLE IF EXISTS organizations;

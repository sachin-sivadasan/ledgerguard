-- Intentionally not reversible: this is a data backfill that created default
-- organizations and OWNER memberships for users who had none. Those orgs may now
-- own real data (partner accounts, subscriptions) and hold additional members, so
-- deleting them on a rollback would be destructive and unsafe to guess at.
SELECT 1;

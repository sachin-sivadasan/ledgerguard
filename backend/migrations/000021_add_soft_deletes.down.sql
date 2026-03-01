-- Remove soft delete indexes
DROP INDEX IF EXISTS idx_subscriptions_app_not_deleted;
DROP INDEX IF EXISTS idx_subscriptions_deleted_at;

-- Remove soft delete column from subscriptions
ALTER TABLE subscriptions DROP COLUMN IF EXISTS deleted_at;

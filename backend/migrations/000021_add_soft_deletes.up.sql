-- Add soft delete column to subscriptions
ALTER TABLE subscriptions ADD COLUMN deleted_at TIMESTAMPTZ;

-- Index for soft delete queries (find non-deleted records efficiently)
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions(deleted_at) WHERE deleted_at IS NULL;

-- Composite index for app queries excluding deleted records
CREATE INDEX idx_subscriptions_app_not_deleted ON subscriptions(app_id) WHERE deleted_at IS NULL;

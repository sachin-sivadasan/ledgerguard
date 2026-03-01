-- General audit log table for tracking user actions
-- Separate from api_audit_log which is specific to Revenue API requests
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for user-based queries (most common pattern)
CREATE INDEX idx_audit_log_user_created ON audit_log(user_id, created_at DESC);

-- Index for time-based queries
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);

-- Index for action-based queries
CREATE INDEX idx_audit_log_action ON audit_log(action, created_at DESC);

-- Index for resource-based queries
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id) WHERE resource_id IS NOT NULL;

-- Comments
COMMENT ON TABLE audit_log IS 'General audit trail for user actions. Used for compliance, debugging, and activity monitoring.';
COMMENT ON COLUMN audit_log.user_id IS 'User who performed the action. NULL for system-initiated actions.';
COMMENT ON COLUMN audit_log.action IS 'Action type (LOGIN, APP_SELECT, SYNC_START, etc.)';
COMMENT ON COLUMN audit_log.resource_type IS 'Type of resource affected (USER, APP, SUBSCRIPTION, etc.)';
COMMENT ON COLUMN audit_log.resource_id IS 'ID of the affected resource';
COMMENT ON COLUMN audit_log.details IS 'Additional context as JSON (sanitized, no sensitive data)';

-- Create api_audit_log table for Revenue API request logging
CREATE TABLE api_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    request_params JSONB,
    response_status INT NOT NULL,
    response_time_ms INT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for key + time queries (most common query pattern)
CREATE INDEX idx_api_audit_key_created ON api_audit_log(api_key_id, created_at DESC);

-- Index for time-based queries
CREATE INDEX idx_api_audit_created ON api_audit_log(created_at DESC);

-- Partial index for error investigation (status >= 400)
CREATE INDEX idx_api_audit_errors ON api_audit_log(response_status, created_at DESC) WHERE response_status >= 400;

-- Comment on table
COMMENT ON TABLE api_audit_log IS 'Audit log for all Revenue API requests. Used for debugging, analytics, and compliance.';
COMMENT ON COLUMN api_audit_log.request_params IS 'Sanitized query params or body (no sensitive data)';
COMMENT ON COLUMN api_audit_log.ip_address IS 'Client IP (IPv4 or IPv6, max 45 chars)';

-- Create api_keys table for Revenue API authentication
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(64) NOT NULL,
    name VARCHAR(100),
    rate_limit_per_minute INT NOT NULL DEFAULT 60 CHECK (rate_limit_per_minute BETWEEN 1 AND 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Unique index on key hash (for fast lookups during authentication)
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);

-- Index for user's keys
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

-- Partial index for active keys only
CREATE INDEX idx_api_keys_active ON api_keys(user_id) WHERE revoked_at IS NULL;

-- Comment on table
COMMENT ON TABLE api_keys IS 'API keys for Revenue API external access. key_hash is SHA-256 of raw key.';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the raw API key. Raw key is only shown once at creation.';

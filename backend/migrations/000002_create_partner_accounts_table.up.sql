-- Create partner_accounts table
CREATE TABLE partner_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    integration_type VARCHAR(20) NOT NULL CHECK (integration_type IN ('OAUTH', 'MANUAL')),
    partner_id VARCHAR(100) NOT NULL,
    encrypted_access_token BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for user lookups
CREATE INDEX idx_partner_accounts_user_id ON partner_accounts(user_id);

-- Index for partner ID lookups
CREATE INDEX idx_partner_accounts_partner_id ON partner_accounts(partner_id);

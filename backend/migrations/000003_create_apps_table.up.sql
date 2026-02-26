-- Create apps table
CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_account_id UUID NOT NULL REFERENCES partner_accounts(id) ON DELETE CASCADE,
    partner_app_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    tracking_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (partner_account_id, partner_app_id)
);

-- Index for partner account lookups
CREATE INDEX idx_apps_partner_account_id ON apps(partner_account_id);

-- Index for tracking enabled apps
CREATE INDEX idx_apps_tracking_enabled ON apps(partner_account_id, tracking_enabled);

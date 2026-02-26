-- Create transactions table (immutable ledger)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    shopify_gid VARCHAR(255) UNIQUE NOT NULL,
    myshopify_domain VARCHAR(255) NOT NULL,
    charge_type VARCHAR(20) NOT NULL CHECK (charge_type IN ('RECURRING', 'USAGE', 'ONE_TIME', 'REFUND')),
    amount_cents BIGINT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    transaction_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for app and date queries (most common)
CREATE INDEX idx_transactions_app_date ON transactions(app_id, transaction_date DESC);

-- Index for domain queries
CREATE INDEX idx_transactions_domain ON transactions(app_id, myshopify_domain);

-- Index for charge type queries
CREATE INDEX idx_transactions_type ON transactions(app_id, charge_type);

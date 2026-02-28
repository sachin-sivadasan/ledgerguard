-- Add revenue share tier and updated_at to apps table
ALTER TABLE apps
ADD COLUMN revenue_share_tier VARCHAR(20) DEFAULT 'DEFAULT_20'
    CHECK (revenue_share_tier IN ('DEFAULT_20', 'SMALL_DEV_0', 'SMALL_DEV_15', 'LARGE_DEV_15'));

ALTER TABLE apps
ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Update existing apps to have updated_at equal to created_at
UPDATE apps SET updated_at = created_at WHERE updated_at IS NULL;

-- Add fee breakdown columns to transactions table
-- These store the actual fees from Shopify Partner API
ALTER TABLE transactions
ADD COLUMN gross_amount_cents BIGINT;

ALTER TABLE transactions
ADD COLUMN shopify_fee_cents BIGINT;

ALTER TABLE transactions
ADD COLUMN processing_fee_cents BIGINT;

ALTER TABLE transactions
ADD COLUMN tax_on_fees_cents BIGINT;

ALTER TABLE transactions
ADD COLUMN net_amount_cents BIGINT;

-- Rename existing amount_cents to be clearer (it was the net amount)
-- First, populate net_amount_cents with existing data
UPDATE transactions SET net_amount_cents = amount_cents WHERE net_amount_cents IS NULL;

-- Comment to clarify the columns
COMMENT ON COLUMN transactions.gross_amount_cents IS 'What the merchant paid (from Shopify Partner API)';
COMMENT ON COLUMN transactions.shopify_fee_cents IS 'Revenue share deducted (0%, 15%, or 20%)';
COMMENT ON COLUMN transactions.processing_fee_cents IS 'Processing fee (2.9%)';
COMMENT ON COLUMN transactions.tax_on_fees_cents IS 'Tax on Shopify fees';
COMMENT ON COLUMN transactions.net_amount_cents IS 'What the developer receives';
COMMENT ON COLUMN transactions.amount_cents IS 'Legacy: same as net_amount_cents';
COMMENT ON COLUMN apps.revenue_share_tier IS 'Shopify revenue share tier: DEFAULT_20, SMALL_DEV_0, SMALL_DEV_15, LARGE_DEV_15';

-- Remove fee breakdown columns from transactions table
ALTER TABLE transactions DROP COLUMN IF EXISTS gross_amount_cents;
ALTER TABLE transactions DROP COLUMN IF EXISTS shopify_fee_cents;
ALTER TABLE transactions DROP COLUMN IF EXISTS processing_fee_cents;
ALTER TABLE transactions DROP COLUMN IF EXISTS tax_on_fees_cents;
ALTER TABLE transactions DROP COLUMN IF EXISTS net_amount_cents;

-- Remove revenue share tier and updated_at from apps table
ALTER TABLE apps DROP COLUMN IF EXISTS revenue_share_tier;
ALTER TABLE apps DROP COLUMN IF EXISTS updated_at;

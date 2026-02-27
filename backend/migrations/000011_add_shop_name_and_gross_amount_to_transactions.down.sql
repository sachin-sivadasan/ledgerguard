-- Remove shop_name and gross_amount_cents columns from transactions table
DROP INDEX IF EXISTS idx_transactions_shop_name;
ALTER TABLE transactions DROP COLUMN IF EXISTS gross_amount_cents;
ALTER TABLE transactions DROP COLUMN IF EXISTS shop_name;

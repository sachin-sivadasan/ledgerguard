-- Add shop_name and gross_amount_cents columns to transactions table
-- Also rename amount_cents to net_amount_cents for clarity

-- Add new columns
ALTER TABLE transactions ADD COLUMN shop_name VARCHAR(255);
ALTER TABLE transactions ADD COLUMN gross_amount_cents BIGINT;

-- Copy existing amount to net_amount (amount_cents is already net amount)
-- We'll keep amount_cents as an alias via application layer

-- Create index for shop name
CREATE INDEX idx_transactions_shop_name ON transactions(shop_name);

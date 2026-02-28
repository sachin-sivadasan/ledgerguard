-- Add earnings tracking fields to transactions table
-- Tracks when charges were created and when earnings become available

-- Add new columns
ALTER TABLE transactions ADD COLUMN created_date TIMESTAMPTZ;
ALTER TABLE transactions ADD COLUMN available_date TIMESTAMPTZ;
ALTER TABLE transactions ADD COLUMN earnings_status VARCHAR(20)
    CHECK (earnings_status IN ('PENDING', 'AVAILABLE', 'PAID_OUT'));

-- Backfill existing transactions: use transaction_date as created_date
-- and calculate available_date as created_date + 7 days (conservative estimate)
UPDATE transactions
SET
    created_date = COALESCE(transaction_date, created_at),
    available_date = COALESCE(transaction_date, created_at) + INTERVAL '7 days',
    earnings_status = CASE
        WHEN COALESCE(transaction_date, created_at) + INTERVAL '7 days' <= NOW() THEN 'AVAILABLE'
        ELSE 'PENDING'
    END
WHERE created_date IS NULL;

-- Make columns NOT NULL after backfill
ALTER TABLE transactions ALTER COLUMN created_date SET NOT NULL;
ALTER TABLE transactions ALTER COLUMN available_date SET NOT NULL;
ALTER TABLE transactions ALTER COLUMN earnings_status SET NOT NULL;

-- Set default for new transactions
ALTER TABLE transactions ALTER COLUMN earnings_status SET DEFAULT 'PENDING';

-- Add index for querying by earnings status
CREATE INDEX idx_transactions_earnings_status ON transactions(app_id, earnings_status);
CREATE INDEX idx_transactions_available_date ON transactions(app_id, available_date);

-- Add comment explaining the fields
COMMENT ON COLUMN transactions.created_date IS 'When the charge was created in Shopify';
COMMENT ON COLUMN transactions.available_date IS 'When earnings become available for payout';
COMMENT ON COLUMN transactions.earnings_status IS 'PENDING (not yet available), AVAILABLE (ready for payout), PAID_OUT (disbursed)';

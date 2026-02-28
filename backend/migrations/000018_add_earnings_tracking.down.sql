-- Rollback earnings tracking fields

DROP INDEX IF EXISTS idx_transactions_available_date;
DROP INDEX IF EXISTS idx_transactions_earnings_status;

ALTER TABLE transactions DROP COLUMN IF EXISTS earnings_status;
ALTER TABLE transactions DROP COLUMN IF EXISTS available_date;
ALTER TABLE transactions DROP COLUMN IF EXISTS created_date;

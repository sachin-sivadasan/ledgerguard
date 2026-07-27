-- Add a real business subscription-start date, distinct from created_at.
-- created_at is the record-created / ingestion timestamp — it is reset on every ledger
-- rebuild (which hard-deletes + re-inserts subscriptions), so it cannot be used as the
-- subscription's actual start date. activated_at is the earliest charge date for the
-- subscription and is populated by the rebuild going forward.
ALTER TABLE subscriptions ADD COLUMN activated_at TIMESTAMPTZ;

-- Backfill existing rows from the earliest RECURRING charge date per (app, store),
-- matching how the rebuild sets it going forward: MIN(created_date, falling back to
-- transaction_date) over recurring transactions only (a subscription starts at its first
-- recurring charge, not a preceding one-time/usage charge).
UPDATE subscriptions s
SET activated_at = t.first_charge
FROM (
    SELECT app_id,
           myshopify_domain,
           MIN(COALESCE(created_date, transaction_date)) AS first_charge
    FROM transactions
    WHERE charge_type = 'RECURRING'
    GROUP BY app_id, myshopify_domain
) t
WHERE s.app_id = t.app_id
  AND s.myshopify_domain = t.myshopify_domain
  AND s.activated_at IS NULL;

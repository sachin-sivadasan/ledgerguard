ALTER TABLE user_preferences
  DROP COLUMN IF EXISTS auto_sync,
  DROP COLUMN IF EXISTS sync_frequency,
  DROP COLUMN IF EXISTS workspace_name,
  DROP COLUMN IF EXISTS currency,
  DROP COLUMN IF EXISTS timezone;

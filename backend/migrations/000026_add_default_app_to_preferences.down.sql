-- Remove default_app_id from user_preferences
DROP INDEX IF EXISTS idx_user_preferences_default_app;
ALTER TABLE user_preferences DROP COLUMN IF EXISTS default_app_id;

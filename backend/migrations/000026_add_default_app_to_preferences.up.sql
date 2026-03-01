-- Add default_app_id to user_preferences for multi-app support
ALTER TABLE user_preferences ADD COLUMN IF NOT EXISTS default_app_id UUID REFERENCES apps(id) ON DELETE SET NULL;

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS idx_user_preferences_default_app ON user_preferences(default_app_id);

COMMENT ON COLUMN user_preferences.default_app_id IS 'Default app to show in dashboard for multi-app users';

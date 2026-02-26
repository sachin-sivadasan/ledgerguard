-- Notification preferences per user
CREATE TABLE IF NOT EXISTS notification_preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    critical_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    daily_summary_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    daily_summary_time TIME NOT NULL DEFAULT '08:00:00',
    slack_webhook_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for finding preferences by user
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);

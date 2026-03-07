ALTER TABLE notification_preferences ADD COLUMN IF NOT EXISTS ai_provider VARCHAR(20) DEFAULT 'openai';

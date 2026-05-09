ALTER TABLE notification_preferences
  DROP COLUMN IF EXISTS email_enabled,
  DROP COLUMN IF EXISTS slack_enabled,
  DROP COLUMN IF EXISTS churn_alerts_enabled,
  DROP COLUMN IF EXISTS revenue_alerts_enabled,
  DROP COLUMN IF EXISTS review_alerts_enabled,
  DROP COLUMN IF EXISTS risk_threshold_days;

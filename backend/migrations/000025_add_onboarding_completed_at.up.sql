-- Add onboarding completion tracking to users
ALTER TABLE users ADD COLUMN onboarding_completed_at TIMESTAMPTZ;

-- Comment
COMMENT ON COLUMN users.onboarding_completed_at IS 'Timestamp when user completed onboarding flow. NULL means onboarding not yet completed.';

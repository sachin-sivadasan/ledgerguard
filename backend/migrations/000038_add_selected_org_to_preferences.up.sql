ALTER TABLE user_preferences ADD COLUMN IF NOT EXISTS selected_org_id UUID REFERENCES organizations(id) ON DELETE SET NULL;

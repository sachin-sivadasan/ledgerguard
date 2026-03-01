-- Add install_count column to apps table
ALTER TABLE apps ADD COLUMN IF NOT EXISTS install_count INTEGER DEFAULT 0;

-- Add index for sorting by install count
CREATE INDEX IF NOT EXISTS idx_apps_install_count ON apps(install_count);

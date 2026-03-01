-- Remove install_count column from apps table
DROP INDEX IF EXISTS idx_apps_install_count;
ALTER TABLE apps DROP COLUMN IF EXISTS install_count;

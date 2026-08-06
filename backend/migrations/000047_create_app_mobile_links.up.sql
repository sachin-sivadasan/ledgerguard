-- Per-app mobile store identifiers so LedgerGuard can pull the developer's iOS/Android
-- app ratings + reviews from the PUBLIC store endpoints (no credentials). Downloads /
-- revenue / crashes are NOT public and are intentionally out of scope here.
CREATE TABLE IF NOT EXISTS app_mobile_links (
    app_id UUID PRIMARY KEY REFERENCES apps(id) ON DELETE CASCADE,
    ios_app_id VARCHAR(32) NOT NULL DEFAULT '',      -- Apple numeric app id (e.g. 310633997)
    play_package VARCHAR(255) NOT NULL DEFAULT '',    -- Android package (e.g. com.whatsapp)
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

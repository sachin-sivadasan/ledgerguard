CREATE TABLE IF NOT EXISTS app_events (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL,
    shopify_shop_gid VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    raw_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_events_app_id ON app_events(app_id);
CREATE INDEX idx_app_events_shop_gid ON app_events(shopify_shop_gid);
CREATE INDEX idx_app_events_app_shop ON app_events(app_id, shopify_shop_gid);
CREATE INDEX idx_app_events_occurred_at ON app_events(occurred_at);
CREATE UNIQUE INDEX idx_app_events_unique ON app_events(app_id, shopify_shop_gid, event_type, occurred_at);

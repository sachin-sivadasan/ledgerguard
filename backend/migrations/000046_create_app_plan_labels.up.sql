-- Developer-assigned friendly names for price tiers, so plan-based reports can show
-- "Starter" / "Starter (old)" instead of derived "$29.00/mo" labels. Keyed by the price
-- tier (billing_interval + price_cents); a mid-life price change yields a new tier row,
-- letting the same plan be named at each price it has had.
CREATE TABLE IF NOT EXISTS app_plan_labels (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    billing_interval VARCHAR(20) NOT NULL,
    price_cents BIGINT NOT NULL,
    label VARCHAR(120) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, billing_interval, price_cents)
);

CREATE INDEX idx_app_plan_labels_app_id ON app_plan_labels(app_id);

-- Daily metrics snapshots - immutable audit trail
-- One snapshot per app per day, never deleted
CREATE TABLE IF NOT EXISTS daily_metrics_snapshot (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    date DATE NOT NULL,

    -- Revenue metrics (in cents)
    active_mrr_cents BIGINT NOT NULL DEFAULT 0,
    revenue_at_risk_cents BIGINT NOT NULL DEFAULT 0,
    usage_revenue_cents BIGINT NOT NULL DEFAULT 0,
    total_revenue_cents BIGINT NOT NULL DEFAULT 0,

    -- Renewal success rate (0.0 to 1.0)
    renewal_success_rate DECIMAL(5, 4) NOT NULL DEFAULT 0,

    -- Subscription counts by risk state
    safe_count INTEGER NOT NULL DEFAULT 0,
    one_cycle_missed_count INTEGER NOT NULL DEFAULT 0,
    two_cycles_missed_count INTEGER NOT NULL DEFAULT 0,
    churned_count INTEGER NOT NULL DEFAULT 0,
    total_subscriptions INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure one snapshot per app per day
    CONSTRAINT unique_app_date UNIQUE (app_id, date)
);

-- Index for querying snapshots by app and date range
CREATE INDEX idx_daily_metrics_app_date ON daily_metrics_snapshot(app_id, date DESC);

-- Index for finding latest snapshot per app
CREATE INDEX idx_daily_metrics_app_latest ON daily_metrics_snapshot(app_id, date DESC);

COMMENT ON TABLE daily_metrics_snapshot IS 'Daily KPI snapshots - immutable audit trail for trends and reconciliation';
COMMENT ON COLUMN daily_metrics_snapshot.active_mrr_cents IS 'MRR from SAFE subscriptions only';
COMMENT ON COLUMN daily_metrics_snapshot.revenue_at_risk_cents IS 'MRR from ONE_CYCLE_MISSED + TWO_CYCLES_MISSED';
COMMENT ON COLUMN daily_metrics_snapshot.renewal_success_rate IS 'SAFE / Total subscriptions as decimal';

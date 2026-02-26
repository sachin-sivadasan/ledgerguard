-- Daily AI-generated insights (Pro tier only)
-- One insight per app per day
CREATE TABLE IF NOT EXISTS daily_insight (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    insight_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure one insight per app per day
    CONSTRAINT unique_app_insight_date UNIQUE (app_id, date)
);

-- Index for querying insights by app and date range
CREATE INDEX idx_daily_insight_app_date ON daily_insight(app_id, date DESC);

COMMENT ON TABLE daily_insight IS 'AI-generated daily executive briefs (80-120 words) - Pro tier only';
COMMENT ON COLUMN daily_insight.insight_text IS 'AI-generated summary from MetricsSnapshot data';

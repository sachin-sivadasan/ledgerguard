-- User dashboard preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    primary_kpis TEXT[] DEFAULT ARRAY['renewal_success_rate', 'active_mrr', 'revenue_at_risk', 'churned'],
    secondary_widgets TEXT[] DEFAULT ARRAY['usage_revenue', 'total_revenue', 'revenue_mix_chart', 'risk_distribution_chart', 'earnings_timeline'],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

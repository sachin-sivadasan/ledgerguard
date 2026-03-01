-- Subscription lifecycle events table
-- Tracks all state transitions for churn analysis and auditing
CREATE TABLE IF NOT EXISTS subscription_events (
    id UUID PRIMARY KEY,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    from_risk_state VARCHAR(50) NOT NULL,
    to_risk_state VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- webhook, sync, manual, billing_failure, app_uninstalled
    reason TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for finding events by subscription
CREATE INDEX idx_subscription_events_subscription_id ON subscription_events(subscription_id);

-- Index for finding events by type and date (for churn analysis)
CREATE INDEX idx_subscription_events_type_occurred ON subscription_events(event_type, occurred_at);

-- Index for finding churn events (to_risk_state = 'CHURNED')
CREATE INDEX idx_subscription_events_churn ON subscription_events(to_risk_state, occurred_at) WHERE to_risk_state = 'CHURNED';

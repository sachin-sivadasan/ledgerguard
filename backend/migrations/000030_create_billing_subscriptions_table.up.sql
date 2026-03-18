CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    razorpay_subscription_id VARCHAR(255) NOT NULL UNIQUE,
    razorpay_plan_id VARCHAR(255) NOT NULL,
    razorpay_customer_id VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    short_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_billing_subscriptions_user_id ON billing_subscriptions(user_id);
CREATE INDEX idx_billing_subscriptions_razorpay_sub_id ON billing_subscriptions(razorpay_subscription_id);
CREATE INDEX idx_billing_subscriptions_status ON billing_subscriptions(status);

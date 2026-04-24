CREATE TABLE app_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    source_review_id TEXT NOT NULL,
    author TEXT NOT NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body TEXT NOT NULL DEFAULT '',
    review_date TIMESTAMP NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    time_using TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'shopify_app_store',
    scraped_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(app_id, source_review_id)
);

CREATE INDEX idx_app_reviews_app_id ON app_reviews(app_id);
CREATE INDEX idx_app_reviews_review_date ON app_reviews(app_id, review_date DESC);

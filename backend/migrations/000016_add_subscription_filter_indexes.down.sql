-- Remove subscription filter indexes

DROP INDEX IF EXISTS idx_subscriptions_search_domain;
DROP INDEX IF EXISTS idx_subscriptions_search_shop_name;
DROP INDEX IF EXISTS idx_subscriptions_filters;

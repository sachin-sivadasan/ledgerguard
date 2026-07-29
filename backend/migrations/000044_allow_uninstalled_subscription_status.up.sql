-- 'UNINSTALLED' is a legitimate domain subscription status (entity.Subscription.Status:
-- "ACTIVE, CANCELLED, FROZEN, PENDING, UNINSTALLED"; produced by GetLatestSubscriptionStatus
-- from RELATIONSHIP_UNINSTALLED events and set by the webhook/status paths), but the CHECK
-- constraints from migrations 000005 / 000013 never included it. It stayed latent until the
-- StatusProcessor actually persisted an UNINSTALLED status at scale (full-history sync +
-- working event fetch), which failed with subscriptions_status_check (SQLSTATE 23514). Add
-- UNINSTALLED to both the write model and the CQRS read model (which copies sub.Status verbatim).

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check
    CHECK (status IN ('ACTIVE', 'CANCELLED', 'FROZEN', 'PENDING', 'UNINSTALLED'));

ALTER TABLE api_subscription_status DROP CONSTRAINT IF EXISTS api_subscription_status_status_check;
ALTER TABLE api_subscription_status ADD CONSTRAINT api_subscription_status_status_check
    CHECK (status IN ('ACTIVE', 'CANCELLED', 'FROZEN', 'PENDING', 'UNINSTALLED'));

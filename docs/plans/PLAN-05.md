# PLAN-05: Notifications & Webhooks

**Date:** 2026-02-27
**Status:** Completed

## Scope
- NotificationService for scheduling and sending push notifications
- Firebase Cloud Messaging (FCM) integration for iOS/Android/Web
- SlackNotificationProvider for channel alerts
- Shopify webhook handlers (subscription updates, billing failures, app uninstalls)
- Device token management (register/unregister)
- Notification preferences (critical alerts, daily summary, summary hour)

## Notification Types
| Type | Trigger | Channel |
|------|---------|---------|
| Critical Alert | Risk state change | Push + Slack |
| Daily Summary | Scheduled (user's hour) | Push + Slack |
| Billing Failure | Payment failed webhook | Push + Slack |
| App Uninstalled | Shop removed app | Push + Slack |

## API Endpoints
- `POST /api/v1/devices` — Register device token
- `DELETE /api/v1/devices` — Unregister device
- `GET/PUT /api/v1/users/notification-preferences` — Manage preferences
- `POST /webhooks/shopify/subscriptions` — Subscription updates
- `POST /webhooks/shopify/uninstalled` — App uninstalls
- `POST /webhooks/shopify/billing-failure` — Billing failures

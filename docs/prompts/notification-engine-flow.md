# Notification Engine Flow - Interactive Visualization

## Context
You are a senior frontend + visualization engineer building an interactive animated guide showing how LedgerGuard's notification engine delivers real-time alerts and scheduled summaries to Shopify app developers.

Build an educational visualization that helps developers understand:
1. How webhooks trigger real-time notifications on risk state changes
2. How the daily summary scheduler works (hour-based delivery)
3. How device registration enables push notifications
4. How multi-channel delivery routes to FCM, APNs, and Slack

---

## Design Philosophy

### Target Audience
Shopify app developers who:
- Want instant alerts when subscriptions show risk
- Need daily summaries at their preferred time
- Use mobile apps and/or Slack for notifications
- Want to understand the notification pipeline

### Key Principles
1. **Show the flow** - Animated data paths from trigger to delivery
2. **Multiple scenarios** - Critical alerts vs daily summaries vs registration
3. **Multi-channel** - Show parallel delivery to Push + Slack
4. **Risk context** - Tie notifications to risk state changes

---

## Flow Types

### Flow 1: Critical Alert (Real-time)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CRITICAL ALERT FLOW                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  🛍️ Shopify    →    🔔 Webhook    →    ⚠️ Risk     →    📬 Notification    │
│  Webhook           Service            Detection         Service             │
│                                                                              │
│                                                              ↓               │
│                                                    ┌─────────┴─────────┐    │
│                                                    ↓                   ↓    │
│                                                📱 Push            💬 Slack  │
│                                                FCM/APNs           Webhook   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

Trigger: Shopify sends webhook (subscription update, billing failure, uninstall)
Process: WebhookService detects risk state change (SAFE → ONE_CYCLE_MISSED)
Action: NotificationService sends to all user devices + Slack (if configured)
```

### Flow 2: Daily Summary (Scheduled)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         DAILY SUMMARY FLOW                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ⏰ Scheduler    →    ⚙️ Preferences    →    📊 Metrics    →    👤 User     │
│  (15-min tick)        Repository            Snapshot          Devices       │
│                                                                              │
│  Check: Is current hour = user's summary_hour?                              │
│  Query: SELECT user_id WHERE daily_summary = true AND summary_hour = ?      │
│  Fetch: Latest daily_metrics_snapshot for each app                          │
│  Send: "MRR: $5,000 | At Risk: $200 | Rate: 94%"                           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 3: Device Registration

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       DEVICE REGISTRATION FLOW                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  📱 Mobile App    →    🔥 Firebase    →    🖥️ API Server    →    💾 Database │
│  (Flutter)             Get Token          POST /devices        device_tokens │
│                                                                              │
│  1. App calls FirebaseMessaging.getToken()                                  │
│  2. App sends POST /api/v1/devices { token, platform: "android" }           │
│  3. Server stores token in device_tokens table                              │
│  4. Server creates default notification_preferences if not exist            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Flow 4: Multi-Channel Delivery

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       MULTI-CHANNEL DELIVERY                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  📬 NotificationService                                                      │
│         │                                                                    │
│         ├── Check user preferences (critical_alerts, daily_summary)         │
│         │                                                                    │
│         ├──→ 🔥 Firebase FCM ──→ Android devices                            │
│         │                                                                    │
│         ├──→ 🍎 Apple APNs ──→ iOS devices                                  │
│         │                                                                    │
│         └──→ 💬 Slack Webhook ──→ Slack channel (if URL configured)         │
│                                                                              │
│  All channels are sent in PARALLEL for low latency                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Risk States Reference

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    RISK STATES THAT TRIGGER ALERTS                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ✅ SAFE              ⚠️ ONE_CYCLE_MISSED    🔴 TWO_CYCLES_MISSED    💀 CHURNED │
│  0-30 days            31-60 days             61-90 days              >90 days  │
│  past due             past due               past due                past due  │
│                                                                              │
│  Notifications sent when state CHANGES (e.g., SAFE → ONE_CYCLE_MISSED)     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Notification Types

| Type | Trigger | Channel | Color |
|------|---------|---------|-------|
| Critical Alert | Risk state change | Push + Slack | 🔴 Red |
| Daily Summary | Scheduled (user's hour) | Push + Slack | 🔵 Blue |
| Billing Failure | Payment failed webhook | Push + Slack | 🟠 Orange |
| App Uninstalled | Shop removed app | Push + Slack | ⚫ Gray |

---

## Webhook Events

### app_subscriptions/update
```json
{
  "admin_graphql_api_id": "gid://partners/AppSubscription/123",
  "status": "CANCELLED",
  "name": "Pro Plan"
}
```
**Risk Impact:** Status-based risk state update

### app/uninstalled
```json
{
  "id": 12345,
  "myshopify_domain": "store.myshopify.com"
}
```
**Risk Impact:** Immediately marks as CHURNED

### subscription_billing_attempts/failure
```json
{
  "subscription_contract_id": "gid://...",
  "error_code": "card_declined",
  "error_message": "Your card was declined"
}
```
**Risk Impact:** Escalates risk (SAFE → ONE_CYCLE → TWO_CYCLES → CHURNED)

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/devices` | Register device token |
| DELETE | `/api/v1/devices` | Unregister device token |
| GET | `/api/v1/users/notification-preferences` | Get preferences |
| PUT | `/api/v1/users/notification-preferences` | Update preferences |

---

## Technical Requirements

### Framework
- Next.js 14+ with App Router
- TailwindCSS for styling
- React hooks for state and animation

### Animation Approach
- SVG-based flow diagrams
- requestAnimationFrame for smooth animation
- Step-by-step progression with timing
- Play/pause controls

### Visual Style
- Dark theme (slate-950 background)
- Gradient accents (red to purple for notifications)
- Glowing effects for active states
- Color-coded risk states

### Interactions
- Flow type selector (tabs)
- Play/pause animation
- Show/hide step details
- Progress indicator

---

## Component Structure

```
marketing/site/
├── app/notifications/page.tsx          # Page wrapper
└── components/
    └── NotificationFlowVisualization.tsx  # Main visualization
        ├── FlowSelector                   # Tab buttons
        ├── FlowDiagram                    # Animated SVG
        ├── StepDetails                    # Step descriptions
        ├── RiskStatesRef                  # Risk state cards
        └── NotificationTypes              # Type documentation
```

---

## Implementation Notes

1. **Animation Loop:** Use requestAnimationFrame with 1.5s step duration
2. **SVG Entities:** Circles with icons, connected by animated lines
3. **Step Highlighting:** Glow effect and color change for active steps
4. **Multi-path:** Critical alert flow splits to Push AND Slack (parallel)
5. **Reset on Flow Change:** Reset animation step when switching flows

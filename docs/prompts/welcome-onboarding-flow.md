# Welcome & Onboarding Flow - Interactive Visualization

## Context
You are a senior frontend + visualization engineer building an interactive animated guide showing how LedgerGuard's Hybrid Welcome & Onboarding system works. The system combines webhook-triggered email drip campaigns (via n8n + Postmark) with in-app onboarding checklist tracking, plus support for custom external API endpoints.

Build an educational visualization that helps developers/admins understand:
1. How new signups trigger a welcome email drip campaign
2. How the in-app onboarding checklist tracks progress
3. How webhook events sync drip state with checklist completion
4. How the custom webhook option enables third-party flow builders
5. Why n8n and Postmark were chosen (decision comparison)

---

## Design Philosophy

### Target Audience
LedgerGuard admins and developers who:
- Want to understand the post-signup user journey
- Need to configure welcome flows (n8n or custom webhook)
- Want to see how email drip + in-app checklist work together
- May want to plug in their own flow builder (Customer.io, Brevo, ActiveCampaign, etc.)

### Key Principles
1. **Show the hybrid flow** - Animated paths from signup through both email and in-app channels
2. **Decision transparency** - Visual comparison of automation + email options with rationale
3. **Extensibility** - Custom webhook option clearly shown as alternative to n8n
4. **Milestone-driven** - Each onboarding step maps to both a checklist item and a drip email

---

## Flow Types

### Flow 1: New Signup - Welcome Email Drip

```
+---------------------------------------------------------------------------+
|                    NEW SIGNUP -> WELCOME EMAIL DRIP                        |
+---------------------------------------------------------------------------+
|                                                                           |
|  +------------+     +-------------+     +----------+     +-----------+   |
|  |  Firebase   |---->|   Backend   |---->|  Webhook |---->|    n8n    |   |
|  | Auth Signup |     | Auth Middle |     |  Dispatch |    | Workflow  |   |
|  +------------+     +-------------+     +----------+     +-----------+   |
|                      |                                    |               |
|                      | Creates user                       | Drip:        |
|                      | record with                        |              |
|                      | onboarding_                        | Day 0:       |
|                      | completed_at                       |  Welcome     |
|                      | = NULL                             |              |
|                      |                                    | Day 1:       |
|                      |                                    |  Connect     |
|                      |                                    |  Shopify     |
|                      |                                    |              |
|                      |                                    | Day 3:       |
|                      |                                    |  Explore     |
|                      |                                    |  Dashboard   |
|                      |                                    |              |
|                      |                                    | Day 7:       |
|                      |                                    |  Try AI Chat |
|                                                                           |
|  Webhook Payload:                                                         |
|  POST <configured_webhook_url>                                            |
|  {                                                                        |
|    "event": "user.created",                                               |
|    "user_id": "uuid",                                                     |
|    "email": "dev@example.com",                                            |
|    "display_name": "Jane Dev",                                            |
|    "plan_tier": "FREE",                                                   |
|    "created_at": "2026-03-09T12:00:00Z"                                   |
|  }                                                                        |
+---------------------------------------------------------------------------+

Trigger: First valid Firebase token -> user doesn't exist -> create
Process: Auth middleware fires "user.created" event after DB insert
Action: EventDispatcher POSTs webhook payload to configured URL
```

### Flow 2: In-App Onboarding Checklist

```
+---------------------------------------------------------------------------+
|                    IN-APP ONBOARDING CHECKLIST                             |
+---------------------------------------------------------------------------+
|                                                                           |
|  User lands on Dashboard (first login)                                    |
|         |                                                                 |
|         v                                                                 |
|  GET /api/v1/users/onboarding-status                                     |
|         |                                                                 |
|         v                                                                 |
|  { is_complete: false, next_step: "connect_partner" }                     |
|         |                                                                 |
|         v                                                                 |
|  +--------------------------------------------------+                    |
|  |         ONBOARDING CHECKLIST BANNER               |                    |
|  |                                                    |                   |
|  |  Welcome to LedgerGuard! Complete these steps:     |                   |
|  |                                                    |                   |
|  |  [x] Sign up for LedgerGuard           (auto)     |                   |
|  |  [ ] Connect Shopify Partner Account    (step 1)   |                   |
|  |  [ ] Select an App to Track             (step 2)   |                   |
|  |  [ ] View Your Dashboard Metrics        (step 3)   |                   |
|  |  [ ] Try the AI Revenue Assistant       (step 4)   |                   |
|  |                                                    |                   |
|  |  Progress: [=====>                    ] 1/5        |                   |
|  +--------------------------------------------------+                    |
|                                                                           |
|  On each step completion:                                                 |
|    1. Update local checklist state                                        |
|    2. POST /api/v1/users/onboarding-step                                  |
|       { "step": "connect_partner", "completed_at": "..." }               |
|    3. Backend fires webhook: { "event": "onboarding.step_completed" }     |
|    4. n8n cancels pending drip for that step                              |
|                                                                           |
|  On ALL steps complete:                                                   |
|    1. POST /api/v1/users/onboarding-complete                              |
|    2. Backend fires webhook: { "event": "onboarding.completed" }          |
|    3. n8n sends "You're all set!" email                                   |
|    4. Hide checklist, show success message                                |
+---------------------------------------------------------------------------+
```

### Flow 3: Custom External API (Third-Party Flow Builder)

```
+---------------------------------------------------------------------------+
|                   CUSTOM WEBHOOK OPTION                                    |
+---------------------------------------------------------------------------+
|                                                                           |
|  Instead of n8n, admins can configure ANY external webhook URL.            |
|  The backend sends the same event payloads to the custom endpoint.         |
|                                                                           |
|  Config (env or admin settings):                                          |
|    WELCOME_WEBHOOK_URL=https://hooks.customer.io/v1/events                |
|    WELCOME_WEBHOOK_SECRET=whsec_xxx  (HMAC signing)                       |
|                                                                           |
|  +------------+     +-------------+     +-------------------+            |
|  |   Backend   |---->|   Event     |---->|  External API     |            |
|  |   Event     |     |  Dispatcher |     |  (any provider)   |            |
|  +------------+     +-------------+     +-------------------+            |
|                                          |                               |
|                                          |  Examples:                    |
|                                          |                               |
|                           +--------------+--------------+                |
|                           |              |              |                |
|                           v              v              v                |
|                     +----------+   +----------+   +-----------+          |
|                     |Customer  |   |  Brevo   |   |ActiveCamp |          |
|                     |   .io    |   |(Sendinbl)|   |  aign     |          |
|                     +----------+   +----------+   +-----------+          |
|                     Own mailer     Own mailer      Own mailer             |
|                     Own flows      Own flows       Own flows              |
|                                                                           |
|  Webhook Events Sent:                                                     |
|  +---------------------------------------------------------------+       |
|  | Event                       | When                             |       |
|  |-----------------------------+----------------------------------|       |
|  | user.created                | New user first authenticates     |       |
|  | onboarding.step_completed   | User completes checklist step    |       |
|  | onboarding.completed        | All steps done                  |       |
|  +---------------------------------------------------------------+       |
|                                                                           |
|  Security:                                                                |
|  - HMAC-SHA256 signature in X-LedgerGuard-Signature header               |
|  - Timestamp in X-LedgerGuard-Timestamp for replay protection             |
|  - Retry with exponential backoff (3 attempts, 1s/5s/30s)                |
+---------------------------------------------------------------------------+
```

### Flow 4: Complete Hybrid Architecture

```
+---------------------------------------------------------------------------+
|                    HYBRID WELCOME ARCHITECTURE                             |
+---------------------------------------------------------------------------+
|                                                                           |
|                        +------------------+                               |
|                        |  Firebase Auth   |                               |
|                        |  (Signup/Login)  |                               |
|                        +--------+---------+                               |
|                                 |                                         |
|                                 v                                         |
|                        +------------------+                               |
|                        |  Auth Middleware  |                               |
|                        | (Create User if  |                               |
|                        |  first login)    |                               |
|                        +--------+---------+                               |
|                                 |                                         |
|                    +------------+------------+                            |
|                    |                         |                            |
|                    v                         v                            |
|          +------------------+      +------------------+                   |
|          |  Event Dispatcher |      |  Flutter App     |                  |
|          |  (Webhook POST)  |      |  (In-App State)  |                   |
|          +--------+---------+      +--------+---------+                   |
|                   |                         |                             |
|          +--------+---------+      +--------+---------+                   |
|          |                  |      |                  |                    |
|          v                  v      v                  v                    |
|   +-------------+  +-----------+  +-------------+  +----------+          |
|   |    n8n      |  |  Custom   |  | Onboarding  |  | Checklist|          |
|   |  (default)  |  |  Webhook  |  |   Bloc      |  |  Banner  |          |
|   +------+------+  +-----------+  +------+------+  +----------+          |
|          |                               |                                |
|          v                               v                                |
|   +-------------+              +------------------+                       |
|   |  Postmark   |              | Step Completion  |                       |
|   | (Email Drip)|              | Events -> Webhook|                       |
|   +-------------+              +------------------+                       |
|                                                                           |
|  SYNC MECHANISM:                                                          |
|  When user completes a step in-app:                                       |
|    App -> Backend -> Webhook -> n8n -> Cancel that drip email             |
|  This prevents sending "Connect Shopify" email to someone who already did |
+---------------------------------------------------------------------------+
```

---

## Decision Visualizations

### Automation Platform Comparison

```
+---------------------------------------------------------------------------+
|                  AUTOMATION PLATFORM COMPARISON                            |
+==============+===========+===========+===========+========================+
|              | n8n       | Make      | Zapier    | Cloud Function         |
|              | [chosen]  |           |           |                        |
+==============+===========+===========+===========+========================+
| Cost         | Free      | Free      | $20+/mo  | Pay-per-use            |
|              | (self-    | (1K ops/  |          | (~$0.40/M invocations) |
|              | hosted)   | month)    |          |                        |
+--------------+-----------+-----------+-----------+------------------------+
| Hosting      | Own infra | Cloud     | Cloud    | GCP                    |
|              | (Hetzner) | (EU/US)   | (US)     |                        |
+--------------+-----------+-----------+-----------+------------------------+
| Visual       | Yes       | Yes       | Yes      | No (code only)         |
| Builder      | (Node-    | (Drag &   | (Simple  |                        |
|              | based)    | drop)     | linear)  |                        |
+--------------+-----------+-----------+-----------+------------------------+
| Self-host    | Yes       | No        | No       | N/A (managed)          |
+--------------+-----------+-----------+-----------+------------------------+
| Complexity   | Medium    | Low       | Low      | High                   |
+--------------+-----------+-----------+-----------+------------------------+
| Data Privacy | Full      | Shared    | Shared   | Full control           |
|              | control   | infra     | infra    | (GCP)                  |
+--------------+-----------+-----------+-----------+------------------------+
| Scalability  | Manual    | Auto      | Auto     | Auto                   |
+--------------+-----------+-----------+-----------+------------------------+
| Integrations | 400+      | 1500+    | 6000+    | Custom code            |
+--------------+-----------+-----------+-----------+------------------------+

Decision: n8n
- Free (self-hosted on existing Hetzner infrastructure)
- Visual workflow builder for non-technical iteration
- Full data control (GDPR-friendly, no data leaves your servers)
- Docker deployment: docker run -d n8n
- Webhook receiver built-in for bidirectional communication
```

### Email Provider Comparison

```
+---------------------------------------------------------------------------+
|                    EMAIL PROVIDER COMPARISON                               |
+==============+===========+===========+===========+========================+
|              | SendGrid  | Postmark  | Resend    | Firebase Extensions    |
|              |           | [chosen]  |           |                        |
+==============+===========+===========+===========+========================+
| Cost         | Free      | $15/mo    | Free      | Free (Firestore-       |
|              | 100/day   | 10K msgs  | 100/day   | triggered)             |
+--------------+-----------+-----------+-----------+------------------------+
| Delivery     | Good      | Best      | Good      | Variable               |
| Speed        | (seconds) | (<1 sec)  | (seconds) | (depends on ext)       |
+--------------+-----------+-----------+-----------+------------------------+
| Templates    | Visual    | HTML +    | React     | Limited                |
|              | editor    | Mustache  | Email     | (basic HTML)           |
+--------------+-----------+-----------+-----------+------------------------+
| Transactional| Mixed     | Pure      | Mixed     | N/A                    |
| Focus        | (+ mktg)  | (txn only)| (+ mktg) |                        |
+--------------+-----------+-----------+-----------+------------------------+
| Analytics    | Full      | Good      | Basic     | None                   |
|              | (opens,   | (opens,   | (opens)   |                        |
|              | clicks,   | bounces,  |           |                        |
|              | bounces)  | spam)     |           |                        |
+--------------+-----------+-----------+-----------+------------------------+
| API Quality  | Good      | Excellent | Excellent | N/A                    |
+--------------+-----------+-----------+-----------+------------------------+
| Reputation   | Mixed     | Premium   | New       | Google                 |
|              | (shared   | (dedicated| (growing) | (shared)               |
|              | IPs)      | IPs)      |           |                        |
+--------------+-----------+-----------+-----------+------------------------+

Decision: Postmark
- Best-in-class delivery speed (<1 second to inbox)
- Pure transactional focus (no marketing spam reputation issues)
- B2B SaaS industry standard for welcome/notification emails
- Clean analytics (open rate, bounce rate, spam complaints)
- Dedicated IP reputation (not shared with bulk mailers)
- $15/mo covers 10K emails (sufficient for early-stage B2B SaaS)
```

---

## Webhook Event Payloads

### user.created
```json
{
  "event": "user.created",
  "timestamp": "2026-03-09T12:00:00Z",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "dev@example.com",
    "display_name": "Jane Dev",
    "role": "OWNER",
    "plan_tier": "FREE",
    "created_at": "2026-03-09T12:00:00Z"
  }
}
```

### onboarding.step_completed
```json
{
  "event": "onboarding.step_completed",
  "timestamp": "2026-03-09T13:30:00Z",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "dev@example.com",
    "step": "connect_partner",
    "step_number": 1,
    "total_steps": 5,
    "completed_at": "2026-03-09T13:30:00Z"
  }
}
```

### onboarding.completed
```json
{
  "event": "onboarding.completed",
  "timestamp": "2026-03-10T09:15:00Z",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "dev@example.com",
    "completed_at": "2026-03-10T09:15:00Z",
    "duration_hours": 21.25
  }
}
```

---

## Email Drip Campaign

| Email | Delay | Subject | Purpose | Cancel If |
|-------|-------|---------|---------|-----------|
| Welcome | Immediate | Welcome to LedgerGuard! | Intro + quick start link | Never |
| Connect | Day 1 | Connect your Shopify Partner account | Guide to OAuth setup | `connect_partner` done |
| Dashboard | Day 3 | Your revenue dashboard is waiting | Feature highlights | `view_dashboard` done |
| AI Chat | Day 7 | Meet your AI Revenue Assistant | AI chat demo | `try_ai_chat` done |
| Complete | On completion | You're all set! | Celebrate + pro tips | N/A (triggered by event) |

---

## Onboarding Steps Definition

| Step | ID | Trigger | Detection |
|------|----|---------|-----------|
| Sign Up | `signup` | Auto-complete on user creation | `user.created` event |
| Connect Shopify | `connect_partner` | OAuth callback success or manual token saved | `has_partner_account = true` |
| Select App | `select_app` | App selected and saved | `has_apps = true` |
| View Dashboard | `view_dashboard` | User visits `/dashboard` with loaded metrics | Frontend event |
| Try AI Chat | `try_ai_chat` | User sends first chat message | Frontend event |

---

## Configuration

### Environment Variables (Backend)
```env
# Webhook dispatch (choose one)
WELCOME_WEBHOOK_URL=http://n8n.internal:5678/webhook/user-events
WELCOME_WEBHOOK_SECRET=whsec_your_hmac_secret_here

# Or custom third-party
# WELCOME_WEBHOOK_URL=https://hooks.customer.io/v1/events
# WELCOME_WEBHOOK_SECRET=whsec_customerio_secret
```

### n8n Workflow (Self-hosted)
```
Docker: docker run -d --name n8n -p 5678:5678 n8nio/n8n
Webhook URL: http://n8n.internal:5678/webhook/user-events
Postmark API Key: configured in n8n credentials
```

---

## Technical Requirements

### Framework
- Next.js 14+ with App Router (marketing site visualization)
- TailwindCSS for styling
- React hooks for state and animation

### Animation Approach
- SVG-based flow diagrams with animated paths
- requestAnimationFrame for smooth transitions
- Step-by-step progression with 1.5s step duration
- Play/pause controls

### Visual Style
- Dark theme (slate-950 background)
- Gradient accents (blue to green for onboarding progress)
- Glowing effects for active webhook dispatch
- Color-coded steps (grey=pending, blue=active, green=complete)

### Interactions
- Flow type selector (5 tabs: Signup, Checklist, Custom, Architecture, Decisions)
- Play/pause animation
- Hover for step details
- Toggle between n8n and custom webhook views

---

## Component Structure

```
marketing/site/
+-- app/welcome-flow/page.tsx              # Page wrapper
+-- components/
    +-- WelcomeFlowVisualization.tsx        # Main visualization
        +-- FlowSelector                    # Tab buttons
        +-- SignupFlow                      # Flow 1: Signup -> Email
        +-- ChecklistFlow                   # Flow 2: In-app checklist
        +-- CustomWebhookFlow              # Flow 3: Custom API
        +-- HybridArchitecture             # Flow 4: Full architecture
        +-- DecisionComparison             # Flow 5: Platform/email tables
        +-- EventPayloads                  # JSON payload reference
        +-- DripTimeline                   # Email drip schedule
```

---

## Implementation Notes

1. **Event Dispatcher:** Create `internal/application/service/event_dispatcher.go` with webhook POST + HMAC signing
2. **Idempotency:** Include `X-LedgerGuard-Idempotency-Key` header (event_type + user_id + timestamp)
3. **Retry Logic:** Exponential backoff (1s, 5s, 30s) with dead letter logging
4. **n8n Integration:** n8n receives webhook, routes to Postmark Send Email node
5. **Drip Cancellation:** n8n workflow checks step completion before each drip email
6. **Custom Webhook:** Same payload format regardless of destination — n8n or third-party
7. **HMAC Signing:** `HMAC-SHA256(webhook_secret, timestamp + "." + payload_json)`

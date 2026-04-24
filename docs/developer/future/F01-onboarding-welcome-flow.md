# F01. Onboarding & Welcome Flow

## What It Will Do
A hybrid onboarding experience combining an in-app setup checklist with an automated email drip campaign. New users are guided through: connecting Shopify Partner API, selecting apps to track, completing their first sync, and exploring key features. Emails are triggered by backend events and delivered via Postmark through n8n workflows.

## Why It Matters
First-time users who don't connect their Partner account within 48 hours have a high churn risk. A structured onboarding flow that combines in-app guidance with email nudges significantly improves activation rate. The welcome drip campaign keeps users engaged even when they're not in the app.

## Dependencies
- Backend: Onboarding status tracking (implemented — `GET /api/v1/users/onboarding-status`)
- n8n: Self-hosted automation platform (to be deployed on Hetzner)
- Postmark: Transactional email provider (account setup needed)
- See [ADR-015](../../../DECISIONS.md) (n8n) and [ADR-016](../../../DECISIONS.md) (Postmark)

## Integration Points
- Backend emits webhook events: `user.created`, `onboarding.step_completed`, `onboarding.completed`
- n8n receives events and triggers email sequences
- Custom webhook escape hatch: admins can configure any external URL instead of n8n
- Flutter: `LgOnboardingChecklist` widget already exists in prototype

## Estimated Scope
- Backend: 2-3 days (webhook event emission, drip campaign state machine)
- n8n setup: 1-2 days (Docker deploy, workflow creation)
- Postmark: 1 day (account setup, domain verification, template creation)
- Flutter: 1 day (wire checklist to real API)
- Total: ~5-7 days

## Open Questions
- How many emails in the drip sequence? (Suggested: 4 — welcome, connect reminder, first sync celebration, weekly digest intro)
- Should we include a "setup wizard" route (`/setup`) as fallback? (See future.md)
- What's the trial conversion nudge timing? (Day 7? Day 12?)

## Suggested Approach
1. Deploy n8n on Hetzner (Docker Compose alongside existing services)
2. Create Postmark account, verify sending domain, design email templates
3. Add webhook event emission to backend (user.created, onboarding events)
4. Build n8n workflow: receive webhook → check step → send appropriate email
5. Wire Flutter checklist to real onboarding API (replace mock data)
6. Test full flow: signup → events → n8n → emails → checklist completion

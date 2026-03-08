# PLAN-19: Welcome & Onboarding Flow (Hybrid)

**Date:** 2026-03-09
**Status:** Designed (awaiting implementation)

## Overview
Hybrid Welcome & Onboarding system combining webhook-triggered email drip campaigns (via n8n + Postmark) with in-app onboarding checklist tracking, plus custom webhook support for third-party flow builders.

## Architecture

### Email Drip (via n8n + Postmark)
```
Firebase Auth Signup
    -> Backend creates user (auth middleware)
    -> Fire "user.created" event
    -> POST webhook to n8n (or custom URL)
    -> n8n workflow:
        -> Day 0: Welcome email (Postmark)
        -> Day 1: "Connect Shopify" email (skip if done)
        -> Day 3: "Explore Dashboard" email (skip if done)
        -> Day 7: "Try AI Chat" email (skip if done)
```

### In-App Checklist
```
User lands on Dashboard (first time)
    -> GET /api/v1/users/onboarding-status
    -> Show checklist banner:
        [x] Sign up
        [ ] Connect Shopify Partner
        [ ] Select an App
        [ ] View Dashboard Metrics
        [ ] Try AI Chat
    -> Each step -> POST event -> webhook cancels drip
    -> All complete -> celebration + hide checklist
```

### Custom Webhook (Third-Party)
Same HMAC-signed event payloads sent to any configured URL:
- Customer.io, Brevo, ActiveCampaign, or custom endpoint
- Configured via `WELCOME_WEBHOOK_URL` env var
- HMAC-SHA256 signature for security

## Key Decisions
- ADR-015: n8n for Automation Platform (free, self-hosted, visual builder)
- ADR-016: Postmark for Transactional Email (best delivery, pure transactional)

## Webhook Events
| Event | When |
|-------|------|
| `user.created` | New user first authenticates |
| `onboarding.step_completed` | User completes checklist step |
| `onboarding.completed` | All steps done |

## Implementation Required (Backend)
- `internal/application/service/event_dispatcher.go` — Webhook POST + HMAC signing
- Modify auth middleware to fire `user.created` event
- New endpoint: `POST /api/v1/users/onboarding-step`
- Retry logic: 3 attempts (1s, 5s, 30s backoff)

## Implementation Required (Frontend)
- `OnboardingBloc` — Checklist state management
- `OnboardingChecklistBanner` widget — Progress display
- Conditional routing based on onboarding status
- Step completion event firing

## Visualization
- Prompt: `docs/prompts/welcome-onboarding-flow.md`
- Diagram: `docs/welcome-onboarding-flow.puml`

## Reference
- See `docs/prompts/welcome-onboarding-flow.md` for complete flow diagrams, decision comparison tables, webhook payloads, and email drip schedule

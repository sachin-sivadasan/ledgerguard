# PLAN-20: Billing System (Stripe + Trial-Freemium)

**Date:** 2026-03-09
**Status:** Designed (awaiting implementation)

## Overview
Stripe-based billing system with Trial-Freemium model. 14-day trial with full features, then FREE (limited) or PRO (paid). Database-driven feature gating via `plans` + `plan_features` tables.

## Why Stripe (Not Shopify Billing)
- LedgerGuard is a partner-facing SaaS, NOT a Shopify-installed merchant app
- Shopify Billing API requires app installed in a shop (merchant context only)
- Partner API has NO billing endpoints
- No Partner Tools marketplace exists
- Manual token users can't use Shopify Billing (no OAuth/app installation)
- All comparable tools (Baremetrics, Ship) use external billing

## Plan Tiers
| Tier | Price | Apps | AI Chat | API Keys | Slack | Export |
|------|-------|------|---------|----------|-------|--------|
| TRIAL | $0 (14 days) | Unlimited | Yes | Yes | Yes | Yes |
| FREE | $0 | 1 | No | No | No | No |
| PRO | $XX/mo | Unlimited | Yes | Yes | Yes | CSV/PDF |
| ENTERPRISE | Custom | Unlimited | Yes+ | Yes+ | Yes | Full |

## Data Model
- `plans` — tier definitions with Stripe Price IDs
- `plan_features` — feature flags per plan (database-driven, admin-editable)
- `billing_subscriptions` — user subscription state (Stripe-synced)
- `billing_events` — webhook audit log (idempotency via stripe_event_id)

## Key Flows
1. **Signup → Trial:** Create user + Stripe Customer, 14-day trial, all features
2. **Trial → PRO:** User adds payment, Stripe Checkout, webhook updates plan_tier
3. **Trial → FREE:** No payment, daily cron downgrades, features locked
4. **Upgrade:** Stripe Checkout Session → hosted page → webhook
5. **Cancel:** Cancel at period end, keep PRO until period expires
6. **Feature Gate:** PlanMiddleware checks plan_features before allowing access

## API Endpoints
- `GET /api/v1/billing/plans` — List plans + features
- `GET /api/v1/billing/subscription` — Current subscription
- `POST /api/v1/billing/checkout` — Create Stripe Checkout
- `POST /api/v1/billing/portal` — Stripe Customer Portal
- `POST /api/v1/billing/cancel` — Cancel at period end
- `POST /webhooks/stripe` — Stripe webhook handler

## Key Decisions
- ADR-017: Stripe for Billing (Not Shopify Billing)
- ADR-018: Trial-Freemium Billing Model

## Visualization
- Prompt: `docs/prompts/billing-system-flow.md`
- Diagram: `docs/billing-system-flow.puml`

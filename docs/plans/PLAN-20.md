# PLAN-20: Billing System (Stripe + Trial-Freemium)

**Date:** 2026-03-09
**Status:** Designed (awaiting implementation)

## Overview
Stripe-based billing system. All plans are paid — no free tier. Starter plan includes a 14-day trial. After trial expires without payment, user enters read-only mode. Database-driven feature gating via `plans` + `plan_features` tables.

## Why Stripe (Not Shopify Billing)
- LedgerGuard is a partner-facing SaaS, NOT a Shopify-installed merchant app
- Shopify Billing API requires app installed in a shop (merchant context only)
- Partner API has NO billing endpoints
- No Partner Tools marketplace exists
- Manual token users can't use Shopify Billing (no OAuth/app installation)
- All comparable tools (Baremetrics, Ship) use external billing

## Plan Tiers (No Free Plan)
| Tier | Price | Trial | Apps | AI Chat | API Keys | Slack | Export |
|------|-------|-------|------|---------|----------|-------|--------|
| TRIAL | $0 (14 days) | This IS the trial | 1 | Yes | Yes | Yes | Yes |
| STARTER | $X/mo | 14-day trial | 1 | No | No | No | No |
| PRO | $XX/mo | No trial | Unlimited | Yes | Yes | Yes | CSV/PDF |
| ENTERPRISE | Custom | No trial | Unlimited | Yes+ | Yes+ | Yes | Full |
| EXPIRED | — | — | Read-only | No | No | No | No |

## Data Model
- `plans` — tier definitions with Stripe Price IDs
- `plan_features` — feature flags per plan (database-driven, admin-editable)
- `billing_subscriptions` — user subscription state (Stripe-synced)
- `billing_events` — webhook audit log (idempotency via stripe_event_id)

## Key Flows
1. **Signup → Trial:** Create user + Stripe Customer, 14-day Starter trial
2. **Trial → STARTER:** User subscribes, Stripe Checkout, plan_tier = STARTER
3. **Trial → EXPIRED:** No payment, daily cron sets read-only mode
4. **STARTER → PRO:** Upgrade via Stripe, proration handled
5. **Upgrade (immediate):** Stripe proration — credit unused time on old plan, charge new plan
6. **Downgrade (scheduled):** NOT immediate — scheduled for period end, daily cron executes it
7. **Daily Cron Job:** 4 checks (trials, past-due, scheduled downgrades, reminders)

## Upgrade vs Downgrade Behavior
| Action | When | Proration |
|--------|------|-----------|
| Upgrade (STARTER → PRO) | Immediate | Yes (credit remaining Starter, charge Pro) |
| Downgrade (PRO → STARTER) | Period end (scheduled) | No (user keeps PRO until paid period ends) |
| Cancel | Period end | No (user keeps plan until paid period ends) |
| Monthly → Annual | Immediate | Yes (credit remaining monthly) |

## Daily Subscription Check Job
- Runs daily at 00:00 UTC (in-process goroutine or GCP Cloud Scheduler)
- **Step 1:** Expire trials where `trial_ends_at < NOW()` → set plan_tier = EXPIRED
- **Step 2:** Check past-due subscriptions > 7 days → verify with Stripe, cancel if still unpaid
- **Step 3:** Execute scheduled downgrades where `scheduled_change_at <= NOW()` → update Stripe subscription, set new plan_tier, clear schedule
- **Step 4:** Send trial reminders (7 days, 2 days, 1 day left)
- Idempotent, logs all actions to `billing_events`
- Internal-only endpoint: `POST /internal/cron/billing-check`

## API Endpoints
- `GET /api/v1/billing/plans` — List plans + features
- `GET /api/v1/billing/subscription` — Current subscription (includes scheduled downgrade info)
- `POST /api/v1/billing/checkout` — Create Stripe Checkout (new subscription)
- `POST /api/v1/billing/upgrade` — Upgrade plan (immediate, prorated)
- `POST /api/v1/billing/downgrade` — Schedule downgrade (period end)
- `DELETE /api/v1/billing/downgrade` — Cancel scheduled downgrade
- `POST /api/v1/billing/portal` — Stripe Customer Portal
- `POST /api/v1/billing/cancel` — Cancel subscription at period end
- `POST /webhooks/stripe` — Stripe webhook handler
4. **Upgrade:** Stripe Checkout Session → hosted page → webhook
5. **Cancel:** Cancel at period end, keep PRO until period expires
6. **Feature Gate:** PlanMiddleware checks plan_features before allowing access

## Key Decisions
- ADR-017: Stripe for Billing (Not Shopify Billing)
- ADR-018: All-Paid Billing Model (Trial on Starter, no free tier)

## Visualization
- Prompt: `docs/prompts/billing-system-flow.md`
- Diagram: `docs/billing-system-flow.puml`

# PRD: Billing (Razorpay) (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** / **INFERRED** / **UNKNOWN**.
> Success metrics are **PROPOSED** (future), not historical. Sources:
> `application/service/billing_service.go`, `external/razorpay_client.go`, `handler/billing.go`,
> `valueobject/{plan_tier,billing_plan,billing_subscription_status}.go`, `entity/{billing_subscription,organization}.go`,
> migration 000030, `frontend-flutter/lib/screens/settings/connect_shopify_screen.dart`.

## ⚠️ Headline finding (read first)

**FACT — the billing backend is production-grade and fully tested, but there is no frontend to reach it.** Checkout + webhook lifecycle + plan enforcement all work server-side, yet the only entry point in the app is a **TODO comment** (`connect_shopify_screen.dart:523` — "navigate to upgrade/billing screen"); no billing/upgrade screen exists. So paid plans are currently **unreachable by users** (the same built-but-not-wired pattern as AI Chat, #2).

## Problem & evidence

LedgerGuard monetizes via paid plans (STARTER/PRO) that lift limits and unlock features. It needs a way for users to subscribe, and for plan state to stay in sync with payments.

- **FACT:** a complete Razorpay subscription system exists (checkout, webhooks, state machine, plan-tier sync) with comprehensive tests — but no UI path to it.
- **UNKNOWN — original demand evidence not recorded.** Open question — owner: Product.

## Target users

**INFERRED:** the workspace **OWNER** upgrading their org to a paid tier.
- **Not the target:** merchants; non-owner members.

## Proposed solution (as-built behavior)

**FACT — plans** (`billing_plan.go`, `plan_tier.go`): FREE / STARTER ($249/mo) / PRO ($499/mo). Limits (`organization.go:40-60`): **members** FREE 1 / STARTER 3 / PRO 10; **SSO** PRO-only; **webhooks** STARTER+PRO; **AI insights** PRO-only (`ai_insight_service.go:57` → `ErrProTierRequired`).

**FACT — checkout** (`billing.go:25`, `billing_service.go`): `POST /api/v1/billing/checkout {plan}` finds/creates a Razorpay customer, creates a subscription (TotalCount=120 ≈ 10y), persists a `billing_subscription`, and returns `{subscription_id, short_url}` for Razorpay **hosted checkout**. `GET /api/v1/billing/status` returns current billing state.

**FACT — webhooks** (`POST /webhooks/razorpay`, `billing.go:78`): verifies the `X-Razorpay-Signature` via **HMAC-SHA256** (`razorpay_client.go:157`), then drives a state machine on `subscription.activated/charged/pending/halted/cancelled` (`billing_service.go:205-298`): activated → ACTIVE + set `user.PlanTier`; cancelled → CANCELLED + **immediate** downgrade to FREE. Always returns 200 (prevents Razorpay retry spam).

**FACT — the gap:** none of this is reachable in the Flutter app — no upgrade/plan-selection/billing screen; the upgrade action is an unimplemented TODO.

### Key screens

**No new wireframe** (as-built). The **only** relevant frontend touchpoint is the TODO at `connect_shopify_screen.dart:523`. There is no billing screen to reference — its absence is the finding.

## Platform & policy constraints

**FACT.** Payments via **Razorpay** (Basic-auth API, hosted checkout, HMAC webhooks); keys/plan-ids/webhook-secret from config/env. Billing initializes **only if `cfg.Razorpay.KeyID != ""`** (`main.go:654`) — gracefully absent otherwise. **Currency note:** prices are expressed in **USD** ($249/$499) though Razorpay is India-centric (typically INR) — a currency/gateway question.

## Pricing-tier impact

**FACT:** this IS the pricing system — STARTER/PRO lift member limits and unlock SSO/webhooks/AI-insights; FREE is the default. Downgrade on cancel is **immediate** (no grace period).

## Migration for existing users

**INFERRED — N/A.** New capability; existing users are FREE until they subscribe. `billing_subscriptions` is additive (migration 000030).

## Success metrics (PROPOSED — future-facing)

None measured today (no users can subscribe). Proposed once the UI is wired:
1. **Reachability:** an OWNER can complete checkout from the app and see `plan_tier` update within one webhook cycle (today: impossible — no UI).
2. **State fidelity:** 100% of Razorpay webhook events (activated/charged/halted/cancelled) result in the correct `billing_subscription` status + `plan_tier` (backend already tested).
3. **Involuntary-churn handling:** halted/failed payments are surfaced to the OWNER (today `halted` sets state but there's no user-facing notice).

## Non-goals (deliberately absent in code)

1. **No billing/upgrade UI** — backend-only; checkout unreachable in-app. **FACT.**
2. **No plan-change/downgrade endpoint** — only subscribe + cancel (cancel = immediate FREE). **FACT.**
3. **No grace period / dunning** — cancel/halt take effect immediately; no retry-window UX. **FACT.**
4. **No invoices/receipts** — no Invoice entity. **FACT.**

## Open questions

- **Build the billing/upgrade UI** — the backend is ready; the TODO (`connect_shopify_screen.dart:523`) is the whole gap. Owner: Product/Eng.
- **Currency:** USD prices on Razorpay (India/INR-centric) — is the gateway/currency correct for the market? Owner: Product/Eng.
- **Immediate downgrade on cancel** — add a grace period until period end? Owner: Product.
- **Halted-payment UX** — notify the OWNER on failed charges (dunning)? Owner: Product/Eng.
- **Plan change** (upgrade/downgrade between STARTER↔PRO) — no endpoint; add? Owner: Eng.
- Test vs live Razorpay key handling per environment. Owner: Eng.

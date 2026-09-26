# Tech Design: Billing (Razorpay) (as-built)

**PRD:** docs/prds/16-billing.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Handles **money + external payment state**. The Razorpay subscription IDs, plan IDs, and the `billing_subscriptions` ↔ `plan_tier` mapping are hard to change once real subscriptions exist. Webhook signature handling is security-critical (a forged webhook could grant a free upgrade).

> **As-built.** "Current state IS the design." Every claim cites a real path:line; security-critical ones verified. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — backend is complete & tested; frontend is absent.** Checkout: `POST /api/v1/billing/checkout {plan}` (`billing.go:25`) → `BillingService.CreateCheckout` finds/creates a Razorpay customer, `CreateSubscription` (TotalCount=120), persists a `billing_subscription` (status CREATED), returns `{subscription_id, short_url}` for hosted checkout. Status: `GET /api/v1/billing/status`. **The only frontend touchpoint is a TODO** (`connect_shopify_screen.dart:523`) — no billing screen exists.

**FACT — Razorpay client** (`razorpay_client.go`): `CreateCustomer/CreateSubscription/FetchSubscription/CancelSubscription`, Basic-auth (`SetBasicAuth`, `:175`), base `https://api.razorpay.com/v1`. `VerifyWebhookSignature` (`:157-162`) = **HMAC-SHA256 with `hmac.Equal` (constant-time)** — verified.

**FACT — webhook state machine** (`billing.go:78`, `billing_service.go:205-298`): verifies signature, dispatches `subscription.activated/charged/pending/halted/cancelled`; **activated** → status ACTIVE + set period + `user.PlanTier` = STARTER/PRO; **cancelled** → CANCELLED + **immediate** downgrade to FREE; unhandled events logged; handler always returns 200.

**FACT — plan enforcement** (`organization.go:40-60`, `ai_insight_service.go:57`): members 1/3/10 (FREE/STARTER/PRO), SSO PRO-only, webhooks STARTER+PRO, AI insights PRO-only. Billing initializes only if `cfg.Razorpay.KeyID != ""` (`main.go:654`).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **hosted-checkout + signed-webhook state machine → plan-tier sync**, backend-only (no UI). Flow: see `docs/designs/16-billing-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — `billing_subscriptions`** (migration 000030): id, user_id, razorpay_subscription_id, razorpay_plan_id, razorpay_customer_id, plan, status, amount_cents, currency, current_period_start/end, short_url, timestamps; indexed by user_id, razorpay_subscription_id, status. Status enum: CREATED/ACTIVE/PENDING/HALTED/CANCELLED/COMPLETED. `user.plan_tier` is the derived authorization field. No new migration.

### API & events

**FACT.** `POST /api/v1/billing/checkout` (auth), `GET /api/v1/billing/status` (auth), `POST /webhooks/razorpay` (signature-verified, no auth middleware — signature IS the auth). Inbound events from Razorpay only; no outbound events.
**Breaking-change check:** the checkout/status response shapes would bind a future billing UI; the webhook event names are Razorpay's contract.

## Alternatives considered

- **Razorpay hosted checkout (short_url) vs in-app SDK checkout.** As-built returns a hosted-page URL — simplest, PCI-offloaded. **Choose the in-app SDK if** a seamless in-app payment sheet is wanted (more integration + compliance work).
- **Webhook-driven plan sync vs client-confirmed upgrade.** As-built trusts signed webhooks as the source of truth for `plan_tier` (robust to client drop-off). Correct choice; no alternative recorded.
- **Immediate downgrade on cancel vs grace-until-period-end.** As-built downgrades immediately (`handleCancelled`). **Choose grace** if cancelling mid-period should retain access — a contestable UX decision.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Forged/invalid webhook | `VerifyWebhookSignature` | **FACT** rejected (HMAC-SHA256, constant-time, tested `billing_service_test.go`) | — |
| Payment succeeds, client drops off | webhook still fires | **FACT** plan_tier synced via webhook regardless | — |
| Payment fails / subscription halted | `subscription.halted` | **FACT** status HALTED (tested) | re-pay (no dunning UX, D4) |
| Cancellation | `subscription.cancelled` | **FACT** immediate FREE (tested) | re-subscribe |
| Unknown event type | default case | **FACT** logged, 200 returned | — |
| Razorpay unconfigured | `KeyID == ""` | **FACT** billing not initialized (endpoints absent) | set keys |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1 (blocking). No frontend** — checkout/status endpoints are unreachable; the upgrade action is a TODO (`connect_shopify_screen.dart:523`). Paid plans cannot be purchased in-app. → build the billing UI.
- **D2. Currency mismatch risk** — prices are USD ($249/$499) on Razorpay (India/INR-centric). → confirm gateway/currency for the target market.
- **D3. Immediate downgrade on cancel** — no grace period until period end; a mid-period cancel loses access instantly. → optional grace window.
- **D4. No dunning/failed-payment UX** — `halted` sets state but the OWNER isn't notified; involuntary churn is silent. → wire a notification (engine exists, #9).
- **D5. No plan-change endpoint** — only subscribe + cancel; STARTER↔PRO change requires cancel+resubscribe. → add plan-change.
- **D6. No invoices/receipts** — no Invoice entity. → generate/expose receipts.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. State fidelity: each webhook event → correct `billing_subscription` status + `user.plan_tier` — **exist** (`billing_service_test.go:204-289`).
2. Reachability: OWNER completes checkout from the app → plan updates (metric #1) — **TODO** (blocked by D1, no UI).

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- Invalid signature rejected, activated/cancelled transitions, plan validation, checkout creation, status — **exist** (`billing_service_test.go`, `razorpay_client_test.go`, `billing_subscription_test.go`).

**Adversarial — DIVERGENCE TODO (none exist today):**
- D3 grace-period behavior · D4 halted→notification · D5 plan-change. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/application/service/ -run Billing -v && go test ./internal/infrastructure/external/ -run Razorpay -v`
2. `curl -X POST -H "Authorization: Bearer <idToken>" -H "Content-Type: application/json" -d '{"plan":"STARTER"}' http://localhost:8080/api/v1/billing/checkout | jq .short_url` — open the URL, pay in Razorpay **test mode**.
3. Confirm the webhook fires → `GET /api/v1/billing/status` shows ACTIVE and `/api/v1/me` shows the upgraded `plan_tier`.
4. Send a forged `POST /webhooks/razorpay` with a bad signature → expect rejection.

**External-platform reality:** Razorpay is the payment gateway; unit tests fake the client + craft signed webhook bodies. Real checkout + webhook delivery + signature must be smoke-tested against a Razorpay **test** account; live-key behavior only in production.

## Pricing & policy touchpoints

**FACT:** this is the pricing system (STARTER/PRO gate members/SSO/webhooks/AI-insights). Payment-card compliance is offloaded to Razorpay hosted checkout. → Currency/market question (D2). GDPR/financial-record retention for `billing_subscriptions` — note for review.

## Rollout

**Backend deployed; the release blocker is the UI (D1).** Razorpay keys (test/live) + plan IDs + webhook secret must be set per environment on Hetzner ([[hetzner-migration]]); the webhook URL must be registered in the Razorpay dashboard. Rollout order: build UI → test-mode E2E → switch to live keys. Rollback = revert UI; backend + subscriptions persist.

## Observability

**FACT:** billing events tracked (Mixpanel `billing_activated`, etc.) + webhook logs. **DIVERGENCE:** no dashboard/alert on failed payments (halted rate), signature-rejection spikes, or MRR from billing. On-call reads logs + Razorpay dashboard. → proposed observability TODO.

## Open questions

- **Build the billing/upgrade UI (D1)** — the sole blocker to monetization. Owner: Product/Eng. **(highest priority)**
- **Currency/gateway fit (D2).** Owner: Product/Eng.
- **Grace period on cancel (D3) + dunning on halt (D4).** Owner: Product/Eng.
- **Plan-change endpoint (D5) + invoices (D6).** Owner: Eng.
- Test/live key + webhook-URL management per env. Owner: Eng.

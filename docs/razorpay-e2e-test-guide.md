# Razorpay Billing — E2E Test Guide

End-to-end testing checklist for the Razorpay subscription billing flow (test mode).

---

## Pre-requisites (Razorpay Dashboard)

1. **Sign in** to [Razorpay Test Mode Dashboard](https://dashboard.razorpay.com)
2. **Create two Plans:**
   - Starter: `period=monthly`, `interval=1`, `amount=24900` (in paise for INR, or cents for USD), `currency=USD`
   - Pro: same but `amount=49900`
3. **Copy the Plan IDs** (e.g., `plan_XXXXX`)
4. **Get test API keys** from Settings → API Keys (Key ID + Key Secret)
5. **Get webhook secret** from Settings → Webhooks → create one pointing to your tunnel URL

---

## Configure Backend

Add to `config.local.yaml` or set environment variables:

```yaml
razorpay:
  key_id: "rzp_test_XXXXX"
  key_secret: "XXXXX"
  webhook_secret: "XXXXX"
  starter_plan_id: "plan_XXXXX"
  pro_plan_id: "plan_XXXXX"
```

**Environment variable equivalents:**
```bash
export RAZORPAY_KEY_ID=rzp_test_XXXXX
export RAZORPAY_KEY_SECRET=XXXXX
export RAZORPAY_WEBHOOK_SECRET=XXXXX
export RAZORPAY_STARTER_PLAN_ID=plan_XXXXX
export RAZORPAY_PRO_PLAN_ID=plan_XXXXX
```

---

## Test Steps

| # | Test | How | Expected |
|---|------|-----|----------|
| 1 | Server starts with config | `go run ./cmd/server -config config.local.yaml` | Log: "Razorpay billing handler initialized" |
| 2 | Server starts without config | Remove razorpay block from config | No error — billing routes absent, everything else works |
| 3 | POST /billing/checkout | `curl -X POST localhost:8080/api/v1/billing/checkout -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"plan":"STARTER"}'` | Returns `{"subscription_id":"sub_XXX","short_url":"https://rzp.io/i/XXX"}` |
| 4 | Open short_url | Paste short_url in browser | Razorpay hosted checkout page loads |
| 5 | GET /billing/status | `curl localhost:8080/api/v1/billing/status -H "Authorization: Bearer <token>"` | Returns `{"plan":"FREE","status":"NONE"}` (before payment) or `{"plan":"STARTER","status":"ACTIVE",...}` (after) |
| 6 | Webhook (valid sig) | Send test payload with valid HMAC-SHA256 signature to `POST localhost:8080/webhooks/razorpay` | Returns 200 |
| 7 | Webhook (invalid sig) | Send payload with bad `X-Razorpay-Signature` header | Returns 400 |
| 8 | Flutter UI | Open app → Settings → Plan & Billing | Shows current plan card + Starter/Pro plan cards with Subscribe buttons |
| 9 | Flutter checkout | Tap "Subscribe" on a plan card | Opens Razorpay hosted checkout URL in browser via url_launcher |

---

## Webhook Testing with curl

Generate a valid HMAC-SHA256 signature for testing:

```bash
# Set your webhook secret
WEBHOOK_SECRET="your_webhook_secret_here"

# Create test payload
PAYLOAD='{"event":"subscription.activated","payload":{"subscription":{"entity":{"id":"sub_test123","status":"active","current_start":1700000000,"current_end":1702592000}}}}'

# Generate HMAC-SHA256 signature
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$WEBHOOK_SECRET" | awk '{print $2}')

# Send webhook
curl -X POST localhost:8080/webhooks/razorpay \
  -H "Content-Type: application/json" \
  -H "X-Razorpay-Signature: $SIGNATURE" \
  -d "$PAYLOAD"
```

---

## Razorpay Test Cards

| Card Number | Description |
|-------------|-------------|
| 4111 1111 1111 1111 | Successful payment |
| 5267 3181 8797 5449 | Successful payment (Mastercard) |
| 4000 0000 0000 0002 | Card declined |

- **CVV:** Any 3 digits
- **Expiry:** Any future date
- **OTP (if prompted):** Use Razorpay test OTP flow (auto-completes in test mode)

---

## API Endpoints Summary

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/api/v1/billing/checkout` | Firebase Bearer token | Create checkout → returns Razorpay hosted URL |
| GET | `/api/v1/billing/status` | Firebase Bearer token | Current billing plan and status |
| POST | `/webhooks/razorpay` | HMAC-SHA256 signature | Razorpay webhook handler |

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| "Razorpay billing handler initialized" not in logs | Check `razorpay.key_id` is set in config or env |
| 401 on /billing/checkout | Ensure Firebase auth token is valid and not expired |
| Webhook returns 400 | Signature mismatch — verify webhook secret matches Razorpay dashboard |
| short_url doesn't load | Verify Razorpay plan exists and subscription was created (check Razorpay dashboard → Subscriptions) |
| Flutter "Could not open checkout page" | Ensure `url_launcher` is configured for the platform (web/iOS/Android) |

---

*Last updated: 2026-03-18*

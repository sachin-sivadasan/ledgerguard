# Future Features – LedgerGuard

Postponed ideas and features for later implementation.

---

## Backlog

| Feature | Priority | Notes |
|---------|----------|-------|
| Multi-app support | P1 | Track multiple apps per workspace |
| Revenue forecasting | P2 | ML-based prediction |
| Anomaly detection | P2 | Alert on unusual patterns |
| Stripe integration | P3 | Non-Shopify revenue |
| Native mobile app | P3 | iOS/Android standalone |
| Custom report builder | P3 | User-defined reports |
| Subscription detail view | P2 | View individual subscription details, history, risk timeline |
| Affiliate program | P4 | Referral system |

---

## Ideas (Unvalidated)

-

---

## Feature Details

### Subscription Detail View (P2)
**Added:** 2026-02-27

**Description:**
View individual subscription details with full history and risk analysis.

**Proposed Features:**
- Subscription overview (shop name, plan, MRR, status)
- Risk state with timeline visualization
- Payment history with charge types (RECURRING, USAGE, ONE_TIME, REFUND)
- Expected next charge date
- Days since last payment
- Risk state change history
- Actions: Mark as churned, Add note, Export history

**Navigation:**
- From Risk Breakdown page → tap subscription row
- From Dashboard → search or list view

**Endpoints needed:**
- `GET /api/v1/subscriptions/{id}` - Subscription details
- `GET /api/v1/subscriptions/{id}/history` - Payment history
- `GET /api/v1/subscriptions/{id}/risk-timeline` - Risk state changes

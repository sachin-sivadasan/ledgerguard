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
| Subscription list page | P2 | List all subscriptions with filters (risk state, plan, search) |
| Onboarding flow | P1 | Guide new users through setup (connect partner, select app, first sync) |
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

### Subscription List Page (P2)
**Added:** 2026-02-27

**Description:**
List all subscriptions with filtering, sorting, and search capabilities.

**Proposed Features:**
- Paginated list of all subscriptions
- Filter by risk state (SAFE, ONE_CYCLE_MISSED, TWO_CYCLES_MISSED, CHURNED)
- Filter by plan tier
- Search by shop name or email
- Sort by MRR, risk state, last payment date
- Quick stats summary (total count, at-risk count, churned count)
- Tap row to navigate to Subscription Detail View

**UI Components:**
- Filter chips or dropdown for risk state
- Search input field
- Sortable column headers
- Subscription row with shop name, MRR, risk badge, last payment
- Pagination controls or infinite scroll

**Endpoints needed:**
- `GET /api/v1/subscriptions?risk_state=&plan=&search=&sort=&page=&limit=`

### Onboarding Flow (P1)
**Added:** 2026-02-27

**Description:**
Guide new users through the initial setup process after signup.

**Proposed Steps:**
1. **Welcome Screen** - Brief intro to LedgerGuard value proposition
2. **Connect Partner Account** - OAuth or manual token entry for Shopify Partner API
3. **Select App** - Choose which app to track from available apps
4. **Initial Sync** - Trigger first data sync with progress indicator
5. **Setup Complete** - Success screen with link to dashboard

**UI Components:**
- Step indicator (1/5, 2/5, etc.)
- Progress bar across steps
- Skip option (where appropriate)
- Back navigation
- Loading states during API calls

**State Management:**
- OnboardingBloc with steps: welcome, connectPartner, selectApp, syncing, complete
- Persist onboarding progress (resume if interrupted)
- Track completion status in user profile

**Navigation:**
- Auto-redirect new users to onboarding after signup
- Redirect to dashboard after completion
- Allow re-entry from settings if setup incomplete

**Endpoints needed:**
- `GET /api/v1/users/onboarding-status` - Check if onboarding complete
- `POST /api/v1/users/onboarding-complete` - Mark onboarding as done

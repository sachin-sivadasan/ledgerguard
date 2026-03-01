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
| Dark mode support | P3 | System/manual theme toggle with dark color palette |
| Affiliate program | P4 | Referral system |

---

## Completed

| Feature | Completed | Notes |
|---------|-----------|-------|
| Subscription detail view | 2026-03-01 | GET /api/v1/subscriptions/{id}, /history, /risk-timeline |
| Subscription list page | 2026-02-28 | GET /api/v1/apps/{appID}/subscriptions with filters, pagination, sorting |
| Onboarding flow (backend) | 2026-03-01 | GET /api/v1/users/onboarding-status, POST /api/v1/users/onboarding-complete |
| Config validation | 2026-03-01 | Added Validate() and HasCriticalWarnings() to config.go |
| RegisterDevice error handling | 2026-03-01 | Fixed to only ignore duplicate key errors |
| Webhook integration | 2026-03-01 | Real-time subscription updates, billing failures, app uninstalls |
| GitHub Actions CI | 2026-03-01 | Backend tests, lint, frontend tests, marketing site build |
| io.ReadAll error handling | 2026-03-01 | Verified all usages handle errors correctly |
| Repository contract clarity | 2026-03-01 | Added documentation to AppRepository interface |

---

## Technical Debt / Code Quality

All items resolved. See "Completed" section above.

---

## Ideas (Unvalidated)

-

---

## Feature Details

### Multi-App Support (P1)
**Added:** 2026-03-01

**Description:**
Allow users to track multiple Shopify apps within a single workspace/account.

**Proposed Features:**
- App selector/switcher in dashboard header
- Aggregate metrics view across all apps (optional)
- Per-app metrics and subscription views
- App management page (add/remove tracked apps)
- Default app preference setting

**Backend Requirements:**
- Already supports multiple apps per partner account
- Add aggregate metrics endpoint: `GET /api/v1/metrics/aggregate`
- Add user preference for default app

**Frontend Requirements:**
- App selector dropdown in header/sidebar
- "All Apps" aggregate view option
- Persist selected app in local storage
- Update all data fetching to use selected app ID

**Database:**
- No changes needed (apps table already supports multiple per account)

---

### Dark Mode Support (P3)
**Added:** 2026-02-27

**Description:**
Add dark theme support with system preference detection and manual toggle.

**Proposed Features:**
- Dark color palette matching brand identity
- System theme detection (follow device settings)
- Manual toggle in settings (Light/Dark/System)
- Persist preference locally
- Smooth transition animation between themes

**Implementation:**
- Create `AppTheme.darkTheme` in `core/theme/app_theme.dart`
- Add `ThemeBloc` or use `ValueNotifier` for theme state
- Update `MaterialApp` to use `themeMode` property
- Store preference in SharedPreferences
- Add theme toggle in Profile/Settings page

**Color Considerations:**
- Dark backgrounds: grey[900], grey[850]
- Card surfaces: grey[800]
- Primary colors remain consistent
- Ensure WCAG contrast compliance
- Charts and badges need dark-mode variants

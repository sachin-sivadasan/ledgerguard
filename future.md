# Future Features – LedgerGuard

Postponed ideas and features for later implementation.

---

## Backlog

| Feature | Priority | Notes |
|---------|----------|-------|
| Revenue forecasting | P2 | ML-based prediction |
| Anomaly detection | P2 | Alert on unusual patterns |
| Stripe integration | P3 | Non-Shopify revenue |
| Native mobile app | P3 | iOS/Android standalone |
| Custom report builder | P3 | User-defined reports |
| Dark mode support | P3 | System/manual theme toggle with dark color palette |
| Home screen widgets | P3 | iOS/Android widgets for MRR, at-risk count |
| Smart search | P3 | Fuzzy matching for store names |
| Voice AI assistant | P4 | Voice commands for navigation and queries |
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
| Multi-app support | 2026-03-01 | Aggregate metrics, default app preference, app selector support |

---

## Technical Debt / Code Quality

All items resolved. See "Completed" section above.

---

## Ideas (Unvalidated)

-

---

## Feature Details

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

---

### Voice AI Assistant (P4)
**Added:** 2026-03-02

**Description:**
Voice-enabled assistant for hands-free navigation and queries. Users can speak commands like "Show store Acme health" or "List subscriptions at risk".

**Why P4 (Low Priority):**
- Target users (developers) prefer tapping/typing over voice
- Privacy concerns with speaking financial data aloud
- High implementation complexity vs value delivered
- Better alternatives exist (widgets, smart search, better notifications)

**Specification:**
- Full visualization and spec at `/voice-assistant` marketing page
- Prompt file: `docs/prompts/voice-assistant-flow.md`

**Proposed Features:**
- Voice capture using `speech_to_text` Flutter package
- Intent classification via Claude API or local model
- Entity extraction (store names, filters, metrics)
- Navigation via GoRouter deep links
- Fallback: Show suggestions if intent unclear

**Supported Commands:**
- "Show details of store [name]" → Subscription details
- "Store [name] health" → Health score page
- "List subscriptions at risk" → Filtered list
- "What's my MRR?" → Dashboard metrics
- "Any billing failures?" → Alerts page

**Higher Priority Alternatives:**
1. Home screen widgets (P3) - Instant access without opening app
2. Smart search (P3) - Type "acme" to find store instantly
3. Better push notifications - Proactive alerts eliminate need to ask

**Implementation Effort:** High (speech recognition, AI integration, entity matching)

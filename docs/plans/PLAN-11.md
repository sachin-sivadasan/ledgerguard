# PLAN-11: Flutter AI Insights, Settings, Profile

**Date:** 2026-02-27
**Status:** Completed

## Scope
- AI Insight card on dashboard (collapsible, PRO tier only, shimmer loading)
- Notification Settings screen (critical alerts toggle, daily summary toggle, time picker)
- Profile page (email, role, plan tier, upgrade button, logout with confirmation)
- Global error handling (SnackbarService, LoadingOverlay, API interceptor)

## Blocs
- `InsightBloc` — AI daily insight data
- `NotificationPreferencesBloc` — Notification settings CRUD

## Routes
- `/settings/notifications`
- `/profile`

## Tests
- InsightBloc tests
- NotificationPreferencesBloc tests
- Profile page widget tests

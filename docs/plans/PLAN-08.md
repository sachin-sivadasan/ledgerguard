# PLAN-08: Flutter Partner Integration & App Selection

**Date:** 2024-01-XX
**Status:** Completed

## Scope
- Partner Integration screen (OAuth connect + manual token form for admin)
- App Selection screen (fetch app list, radio selection, save preference)
- PartnerIntegrationBloc and AppSelectionBloc
- Onboarding page (4-step carousel: Welcome, Partner, App, Insights)

## Screens
- `/onboarding` — 4-step intro carousel
- `/partner-integration` — OAuth button + manual token form
- `/app-selection` — App list with radio selection + confirm

## Blocs
- `PartnerIntegrationBloc` — Manages OAuth flow state
- `AppSelectionBloc` — Fetches apps, manages selection, saves to SharedPreferences

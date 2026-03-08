# PLAN-14: Revenue Share Tier & Preferences

**Date:** 2026-03-01
**Status:** Completed

## Scope
- Revenue Share Tier tracking (SMALL_DEV_0, DEFAULT_20, ENTERPRISE_15, CUSTOM)
- TierSelector widget for choosing tier in App Settings
- FeeInsightsCard showing fee calculations based on selected tier
- EarningsStatusCard and EarningsTimelineChart
- Dashboard preferences backend endpoint (GET/PUT `/api/v1/user/preferences/dashboard`)
- PreferencesPage for dashboard KPI and widget configuration
- Default tier changed from 20% to 0% (most indie devs qualify)

## Key Decisions
- ADR-008: Default Revenue Share Tier Changed to 0%

## Database
- `user_preferences` table (migration 000026)
- Default app preference support

## Routes
- `/settings/app` — App Settings (tier selector)
- `/settings/preferences` — Dashboard preferences

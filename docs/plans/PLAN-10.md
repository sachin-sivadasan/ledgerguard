# PLAN-10: Flutter Risk & Subscription Analytics

**Date:** 2026-02-27
**Status:** Completed

## Scope
- Risk Breakdown screen (pie chart, risk distribution counts)
- Subscription List page with risk filtering, search, pagination
- Subscription Detail page (status, plan, pricing, risk timeline)
- RiskBadge widget with color-coded risk states
- SubscriptionTile widget with initials/avatar
- Store Health page (domain-level metrics)

## Blocs
- `RiskBloc` — Risk distribution data
- `SubscriptionListBloc` — Filtered/paginated subscription list
- `SubscriptionDetailBloc` — Single subscription details
- `StoreHealthBloc` — Domain-level health metrics

## Routes
- `/risk-breakdown`
- `/apps/:appId/subscriptions`
- `/apps/:appId/subscriptions/:subscriptionId`
- `/apps/:appId/stores/:domain/health`

## Tests
- 54 tests: SubscriptionListBloc, RiskBadge, SubscriptionTile

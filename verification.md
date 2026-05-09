# Verification Checklist

Tracks verification points from each implementation plan. Run these after deployment or during QA.

---

## Plan: Migrate from Numeric Shopify App IDs to Internal UUIDs
**Date:** 2026-05-08

- [x] `go test ./...` — all pass
- [x] `flutter analyze` — no issues
- [ ] Login → Dashboard loads (API calls use UUID in path)
- [ ] Navigate all screens → data loads correctly
- [ ] App dropdown works → switching apps reloads with UUID
- [ ] Apps page → "Sync Now" works (UUID used directly, no 400)
- [ ] Demo mode → toggle off → live data loads
- [ ] Check network tab: paths show `/apps/{uuid}/subscriptions`
- [ ] Curl with UUID → 200
- [ ] Curl with numeric → 400 "invalid app ID format"

---

## Plan: Remove "All Apps" Filter & Enforce App Limits Per Plan
**Date:** 2026-05-08

- [ ] **STARTER + 1 app**: Connect → only 1 app selectable → all screens show that app, no "All Apps" dropdown
- [ ] **STARTER + try adding 2nd app**: Backend returns 403, frontend shows disabled checkbox + upgrade prompt
- [ ] **PRO + 2 apps**: Entity screens show app-switcher (no "All Apps"). Dashboard/Analytics show "All Apps" option
- [ ] **PRO + "All Apps" on Dashboard**: Shows portfolio-level MRR, risk distribution across all apps
- [ ] **App switch**: Select different app → data reloads on current screen
- [ ] **`go test ./internal/domain/entity/... ./internal/interfaces/http/handler/...`**: All pass
- [ ] **`flutter analyze`**: No issues

---

## Plan: Unify Flutter Provider Data-Loading Patterns
**Date:** 2026-05-08

- [ ] **Cold start**: Login → Dashboard shows spinner → loads data with correct app in filter chip
- [ ] **Navigation**: Click each sidebar item → spinner → data loads → app filter shows correct app
- [ ] **App switch**: Select different app from dropdown → data reloads for that app
- [ ] **Demo toggle**: Demo ON → Demo OFF → all screens reload with live data via coordinator
- [ ] **Retry**: Stop backend → navigate → error state → start backend → Retry → loads correctly
- [ ] **Clear filters**: Click "Clear" on subscriptions → app stays selected, only facet filters reset
- [ ] **`flutter analyze`**: No issues

---

## Plan: DataLoadingMixin Refactor
**Date:** 2026-05-08

- [ ] **Cold start**: Launch app → login → verify Dashboard loads data automatically (no blank screen)
- [ ] **Navigation**: Click each sidebar menu → verify data loads on first visit
- [ ] **Demo toggle**: Settings → toggle demo OFF → verify all screens show live data
- [ ] **Error recovery**: Stop backend → navigate to screen → see error state → start backend → tap "Retry" → data loads
- [ ] **App switch**: If multi-app, switch apps → verify old request cancelled, new data loads
- [ ] **Logout/login**: Logout → login → verify data reloads fresh (no stale state)

---

## Plan: Mock Shopify Server for Usage Charges
**Date:** 2026-05-05

- [ ] Start mock server: `cd mock-shopify-api && ruby app.rb`
- [ ] Visit `http://localhost:4000/admin` — whale persona (1005) visible in list
- [ ] Click into whale persona — stats card shows ~50K shops, ~50K subs, usage charges count
- [ ] Usage charges table displayed with app/shop/frequency/amount columns
- [ ] Add usage charge via form → appears in table after redirect
- [ ] Hit GraphQL endpoint: `curl -X POST http://localhost:4000/1005/api/2025-07/graphql.json` — returns paginated results
- [ ] Second call to same endpoint uses cache (fast response)
- [ ] Full pipeline test: connect backend to mock server org 1005 → select app → verify queue processes pages of ~1M transactions

---

## Plan: Load Data with Org Context
**Date:** 2026-05-04

- [ ] `flutter analyze` — no errors
- [ ] Start backend + Flutter → login → check network tab: API calls include `X-Org-Id` header
- [ ] First launch: demo mode OFF → API calls hit backend (may show empty data if no Shopify connected)
- [ ] Settings → toggle demo ON → mock data shown
- [ ] Restart app → demo mode stays as last set (SharedPreferences)
- [ ] Logout → login → orgs reload, header set again

---

## Plan: Fix Stale Data — Navigation Reload + Sync Awareness + Manual Refresh
**Date:** 2026-05-08

### Commit 1 — Navigation staleness
- [ ] Navigate to Stores → wait 2+ min → click Dashboard → click Stores → loading + fresh API call
- [ ] Navigate to Stores → immediately re-click Stores (<2 min) → no reload, cached data instant

### Commit 2 — Refresh button
- [ ] Every data screen (Dashboard, Subscriptions, Stores, Transactions, Events, Risk, Analytics, Earnings, Insights) shows a refresh icon in the header
- [ ] Click refresh → loading → fresh data from API
- [ ] Error state "Retry" still works independently

### Commit 3 — Apps page sync
- [ ] Apps page shows "Synced X ago" per app
- [ ] Click "Sync Now" → progress bar appears, polls updates
- [ ] Click "Cancel" → sync cancelled, idle state restored
- [ ] Sync completes → "Synced just now" shown

### Commit 4 — Auto-refresh
- [ ] While sync running, data screens show subtle "Syncing..." next to title
- [ ] Sync completes → "Syncing..." disappears → page auto-reloads with fresh data
- [ ] No sync running → no indicator, no extra polling

### Cross-cutting
- [ ] `flutter analyze` — no issues
- [ ] Demo mode unaffected — no API calls, no sync polling

---

## [2026-05-09] Events & Webhooks Page Improvements

### Events Page
- [ ] `flutter analyze` — no issues
- [ ] Events screen "This Month" → KPI cards show non-zero installs
- [ ] Set Type filter → list filtered, KPI cards still show unfiltered totals
- [ ] Network tab: type filter sends `eventType=RELATIONSHIP_INSTALLED` query param
- [ ] Events with SUBSCRIPTION_FROZEN display "FROZEN" badge (not "INSTALL")
- [ ] Store detail timeline shows correct badges for new types

### Webhooks Page
- [ ] `flutter analyze` — no issues
- [ ] Webhooks KPI "Today" counts refresh correctly
- [ ] App filter chip filters webhook list
- [ ] Time range toggle works (Today/Week/Month)
- [ ] Demo mode off → webhooks list empty
- [ ] Demo mode on → mock data returns

---

## Plan: Analytics & Earnings — Complete Wiring & Missing Features
**Date:** 2026-05-09

### Phase 1 — Mock data fix
- [ ] `flutter analyze` — no issues
- [ ] Live mode + Forecasting tab → shows empty state (not mock numbers)
- [ ] Demo mode + Forecasting tab → shows mock data as before
- [ ] Same for Cohorts, Profit, Multi-App tabs

### Phase 2 — Earnings API wiring + MRR movements
- [ ] Fees & Tiers tab shows tiers from `GET /api/v1/tiers` (not hardcoded)
- [ ] Fee calculator calls `/fees/breakdown` API
- [ ] Earnings status shows pending/available/paidOut from API
- [ ] Fee summary shows savings vs default tier
- [ ] Revenue tab MRR movement chart shows live data

### Phase 3 — Backend endpoints
- [ ] `go test ./...` passes
- [ ] `GET /api/v1/apps/{appID}/forecast?model=linear` returns 12 forecast points
- [ ] `GET /api/v1/apps/{appID}/forecast?model=exponential` returns 12 forecast points
- [ ] `GET /api/v1/apps/{appID}/cohorts` returns cohort retention data
- [ ] `GET /api/v1/apps/{appID}/fees/monthly` returns monthly P&L breakdown
- [ ] `GET /api/v1/apps/{appID}/revenue/concentration` returns top stores
- [ ] Insufficient data (< 90 snapshots) returns 422 for forecast

### Phase 4 — Frontend wiring
- [ ] Forecasting tab shows live forecast with model toggle (Linear/Exponential)
- [ ] Switching models re-fetches data from API
- [ ] Cohorts tab shows live retention heatmap
- [ ] Profit tab shows live monthly P&L from fees API
- [ ] Multi-App tab shows live per-app comparison from aggregate API

### Phase 5 — Enhancements
- [ ] Revenue tab shows top stores by revenue (live mode only)
- [ ] Date picker on Analytics page loads historical snapshot
- [ ] Comparison card shows current vs historical MRR with % change
- [ ] Earnings page shows upcoming 30-day earnings timeline
- [ ] Earnings page shows fee savings card

---

## Plan: Snapshot Strategy — Daily Backfill, Query-Time Aggregation & Trend API
**Date:** 2026-05-09

### DEV LIMIT Fix
- [ ] `snapshot_processor.go` uses `AddDate(-1, 0, 0)` (12 months)
- [ ] `go test ./internal/infrastructure/queue/processors/...` passes

### Backfill Optimization
- [ ] Transactions sorted once before day loop
- [ ] Each transaction visited exactly once (pointer advancement)
- [ ] `UpsertBatch` added to repository interface + persistence
- [ ] All mock repos updated with `UpsertBatch`
- [ ] Tests: optimized backfill matches original output
- [ ] `go test ./internal/domain/service/...` passes

### Granularity API
- [ ] `DownsampleSnapshots` picks last snapshot per period
- [ ] REST: `GET /apps/{appID}/metrics/trend?granularity=weekly` returns ~26 points for 6m
- [ ] GraphQL: `metricsTrend(granularity: WEEKLY)` works
- [ ] `go test ./...` passes

### Frontend
- [ ] Revenue tab chart uses weekly granularity (cleaner trend line)
- [ ] Forecasting tab still uses raw daily data (unchanged)

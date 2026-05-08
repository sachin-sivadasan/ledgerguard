# Verification Checklist

Tracks verification points from each implementation plan. Run these after deployment or during QA.

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

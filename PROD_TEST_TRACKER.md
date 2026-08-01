# Prod Testing Tracker

Page-by-page verification of the live app (app.ledgerspear.com → api.ledgerspear.com,
app `a4d7dfd1` = Zoko — WhatsApp Marketing API). Log every issue here; fix in batches.

**Legend:** 🔴 Critical (wrong data)  🟠 Important  🟡 Minor/polish  ✅ Verified OK  ⏳ To test

Last updated: 2026-07-30

---

## Status by page

| Page | State | Notes |
|------|-------|-------|
| Transactions | ✅ Fixed (PR #45) | Server-side Gross/Net/Cut totals + full-history window; live count 49,607 verified |
| Subscriptions | 🔴 Open findings | See SUB-1..SUB-3 below — partial fix shipped (PR #46), reconciliation fix pending |
| Dashboard | 🔴 Open findings | route is `/` (not `/dashboard`); see DASH-1..DASH-3 |
| Stores | 🟠 Open findings | 2926 stores; see STORE-1..STORE-2 |
| Store Detail | 🔴 Open findings | route `/#/stores/{domain}`; see SD-1..SD-4 |
| Events | 🟠 Open findings | 139 events; see EVT-1..EVT-3 |
| Risk | 🔴 Open findings | see RISK-1..RISK-3 |
| Earnings | 🔴 Open findings | all $0.00 in UI, data exists in API; see EARN-1..EARN-2 |
| Analytics | 🟢 Mostly OK | Revenue/MRR-movement render; see ANALYTICS-1 (minor) — only Revenue tab checked |
| Reports | ⏳ Catalog OK | ~16 report cards render; per-report deep-test pending (many inherit risk/earnings root causes) |
| Webhooks | 🟢 Mostly OK | empty (no webhooks); WEBHOOK-1 minor |
| AI Insights | 🟢 Mostly OK | route `/insights`; empty daily brief; AII-1 |
| Apps | 🔴 Open finding | "0 installs" despite 2926 stores; see APPS-1 |
| API Keys | 🟢 OK | 3 active + 1 revoked; see APIKEYS-1 (verify seed vs real) |
| Webhooks | ⏳ | |
| Risk | ⏳ | |
| Analytics | ⏳ | |
| Earnings | ⏳ | |
| AI Insights | ⏳ | |
| Reports | ⏳ | |
| Apps | ⏳ | |
| API Keys | ⏳ | |

---

## Findings

### Subscriptions

**✅ SUB-1 / SUB-2 FIXED + VERIFIED (PR #49, deployed+resynced).** Active 718 → **1,057**; `lokjoylokjoy` (billed Jul 22, was CANCELLED/CHURNED) → **ACTIVE/SAFE**. Subscriptions (1,057) and Risk (1,068) now converge (were 718/729).

**✅ RISK-1 Dashboard convergence FIXED + VERIFIED (PR #50, deployed+resynced).** After the Wave-3 snapshot + RefreshTodaySnapshot: Subscriptions `/summary` and Dashboard `/metrics` now report **identical** safe **1,061** / at-risk 21 / churned **1,848** (Dashboard was 1,176). Remaining: the **Risk page `/risk/summary` (stores.risk_state)** still differs — safe 1,076, at-risk 130, churned 1,724 — because store health/risk is a **separate cycle-based grading**, not the subscription risk. Unifying store risk with subscription risk is a follow-up (RISK-1b).

**RISK-1 🟠 residual (historical writeup) — Dashboard snapshot lags.** After the fix: Subscriptions `/summary` safe **1,057**, Risk `/summary` safe **1,068** (converged), but Dashboard `/metrics` safe **1,176**. Cause: `/metrics` reads the **daily snapshot written during ledger rebuild — BEFORE `StatusProcessor` reconciliation** (pure charge-recency), while the other two read the post-reconciliation live state. The reconciled ~1,060 is the more-correct number (excludes ~120 subs that genuinely uninstalled/cancelled with no billing after). **Fix:** recompute/refresh the daily snapshot AFTER StatusProcessor runs (or have the metrics endpoint read live subs). Then all three agree.

---

**✅ SUB-1 / SUB-2 (original writeup) — cancel-trap reconciliation.** `Subscription.ApplyEventStatus` now reconciles the event-derived status against billing: a terminal CANCELLED/UNINSTALLED churns only when NO recurring charge post-dates the event (else it's a stale/plan-change cancel → kept ACTIVE), and risk is re-derived from charge recency after every status refresh (fixes the ACTIVE/CHURNED contradiction). Applied in both `status_processor.go` and `sync_service.go` via `GetLatestSubscriptionStatusWithTime`. Needs deploy + full resync. **RISK-1** should then converge (all three pages ultimately derive from `sub.RiskState`; the 718/729/1175 split was live-subs over-churned vs stores vs snapshot) — verify after resync.

**SUB-1 🔴 Cancel trap — active subs mis-churned by stale/plan-change CANCELED events** (original finding)
- Live data: real active ≈ **1,167** (subs billed within 60d, Shopify's grace), but page shows **718**.
- Cross-tab: `CANCELLED/CHURNED` = 1,934; of 2,189 churned, **382 were billed within the last 35 days** (e.g. `lokjoylokjoy` charged Jul 22, yet CANCELLED/CHURNED with a future next-charge Aug 22).
- Root cause: `StatusProcessor` force-churns on any `SUBSCRIPTION_CHARGE_CANCELED`/UNINSTALLED event, ignoring that the sub has a recent recurring charge. Shopify emits CANCELED on **plan changes** too (old plan cancelled + new activated).
- **Proposed fix:** reconcile event-status against charge data — a CANCELLED/UNINSTALLED event only churns if there's been **no recurring charge since that event** (terminal event newer than last charge). Billed-after-cancel ⇒ still active.

**SUB-2 🔴 Risk never re-classified after status update** (pending fix)
- `ACTIVE/CHURNED` = 22 subs (status says active, risk still stale-churned).
- Root cause: `StatusProcessor` updates `status` but only re-runs risk for the churn branch; non-terminal status updates leave `risk_state` stale.
- **Proposed fix:** re-run `ClassifyRisk` after any status update so risk always reflects charge recency.

**SUB-3 ✅ Missing app-event types (SHIPPED, PR #46)**
- `GetLatestSubscriptionStatus` ignored `SUBSCRIPTION_CHARGE_ACTIVATED` (renewal event), `RELATIONSHIP_REACTIVATED`, `DEACTIVATED`, `EXPIRED`, `DECLINED` → active shops fell through to stale UNINSTALLED/CANCELED. Fixed + resynced: active 417 → 718. (Remaining gap is SUB-1/SUB-2.)

### Dashboard (route = `/`)

Live KPIs: MRR $62,917 (+4.2%), Renewal 40.2%, Revenue at Risk $6,241, Usage Rev $27,417.
`/metrics` response: `safe 1175 / 1-cycle 71 / 2-cycle 65 / churned 1615` (= 2926 ✓).

**DASH-1 🔴 Dashboard risk/active disagrees with Subscriptions page** (root cause = SUB-1/SUB-2)
- Dashboard `/metrics` → `safe_count = 1175`; Subscriptions `/summary` → `activeCount(SAFE) = 718`. Same metric, **two different numbers** across pages.
- Cause: `/metrics` (MetricsEngine) recomputes risk from **charge data** (≈ correct, 1175 ≈ true active), while `/summary` reads the **stored `risk_state`** polluted by the StatusProcessor over-churn (SUB-1). They will converge once SUB-1/SUB-2 are fixed and a resync rewrites `risk_state`.
- ⭐ This is strong corroboration that 1175 (not 718) is the real active count.

**DASH-2 🟠 "MRR Trend (12 months)" chart renders blank**
- Only `/metrics` is fetched on load (returns current/previous/delta — **no 12-month series**). No trend/snapshot endpoint is called, so the chart has no data. Confirm whether a trend endpoint exists / should be wired, or the chart should read daily snapshots.

**DASH-3 🟡 "This Week Activity" panel appears empty** — verify whether legit (no events this week under the "This Week" filter) or unwired.

### Stores (route `/#/stores`)

2,926 stores (1:1 with subs). Cards show Health %, LTV, Apps, risk badge.

**STORE-1 🔴 Store risk badges over-churned** — inherits the polluted `risk_state` (same root cause as SUB-1). Many "Churned/10% health" stores are actually active. Fixes with SUB-1.

**STORE-2 🟠 `first_install_date` / `last_interaction` = record-created time, not real dates**
- `/stores` returns `first_install_date: 2026-07-30T07:18:45Z` for every store — that's the resync record-creation timestamp, not the shop's real first-install date. `last_interaction` likewise = today. Violates CreatedAt-is-record-date. Source install date from the earliest `RELATIONSHIP_INSTALLED` event or earliest transaction (cf. `activated_at` backfill already done for subs). Affects tenure/LTV-age if any consumer uses it.

**Not a bug:** hashed store domains (`00430d-6a.myshopify.com`) are Shopify's anonymized handles for churned/uninstalled shops; alphabetically sorted so they lead the list.

### Store Detail (route `/#/stores/{shop_domain}`)

The detail page makes **no store-by-ID request** — it resolves the store (and its subs) **client-side from the already-loaded first list page** (`/stores?page=1&pageSize=20` + `/subscriptions?page=1&pageSize=100`).

**✅ SD-1 / SD-2 / SD-3 FIXED (PR pending) — store detail loads by domain.** `StoreProvider.loadStoreDetail(appId, domain)` fetches the store + its subscriptions server-side via the existing `search` param on both endpoints (exact-domain match), so deep-links and any store past list page 1 resolve; the screen is now stateful (DataLoadingMixin) with loading/not-found/error states. SD-3: Installed Apps shows the real connected-app name (from AppsProvider) instead of the app UUID. Frontend-only. (STORE-2 record-date on install/interaction is still open — separate.)

**SD-1 🔴 Deep-link / any store beyond list page 1 → "Store not found"** (original finding)
- `…/stores/lokjoylokjoy.myshopify.com` (real active store, sorts under 'l') → **"Store not found"**; `…/stores/00430d-6a.myshopify.com` (first on page 1) → renders. Store detail must fetch the store by domain from the server, not rely on the in-memory paginated list. Breaks deep-links, bookmarks, and clicking any store on list pages ≥2.

**SD-2 🟠 "Subscriptions (0) — No subscriptions" on a store that has a subscription**
- Every store == a subscription (2926/2926), yet the detail's Subscriptions card shows 0. Same client-side-lookup limitation (the store's sub isn't in the loaded page-1 slice). Link store→subscriptions server-side.

**SD-3 🟡 Installed Apps shows the app UUID** (`a4d7dfd1-d27f-4cd1-ab08-6e96cbc8ff3c`) instead of the app name ("Zoko — WhatsApp Marketing API"). Resolve app name.

**SD-4 🟡 First Install / Last Interaction = Jul 30 2026** (record-created date, same as STORE-2); **Timeline card empty** (no events rendered — verify wiring vs genuinely no events).

### Events (route `/#/events`)

139 events. This-week KPIs: Installs 3, Uninstalls 0, Churns 4, Billing Failures 0.

**EVT-1 🔴 Plan changes counted as churns (cancel-trap, → SUB-1)**
- `85c635` shows `SUBSCRIPTION_CHARGE_ACTIVATED` **and** "Subscription Cancelled" at the same 7:14 AM; `onewoofclub` same at 6:04 AM. These are plan changes (old charge cancelled + new activated), but each cancel counts toward "Churns This Week: 4". The events feed and the churn KPI both over-count plan-change cancels as churn — same root cause as SUB-1.

**EVT-2 🟠 Wrong event-type badge/category** — `SUBSCRIPTION_CHARGE_ACTIVATED` rows are badged green **"INSTALL"** (it's a charge activation, not an install). Fix the event-type → category/badge mapping.

**EVT-3 🟡 Inconsistent event titles** — some rows humanized ("Subscription Cancelled"), others show the raw enum ("SUBSCRIPTION_CHARGE_ACTIVATED"). Humanize all event-type titles.

**To verify:** all event timestamps are Jul 30 (today) — confirm these are real Shopify `occurredAt` vs sync-detection time.

### Risk (route `/#/risk`)

Funnel: safe 729 / 1-cycle 126 / 2-cycle 5 / churned 2066 (= 2926 ✓). At-Risk Stores (131).

**RISK-1 🔴 THREE inconsistent risk distributions across pages** (same metric, 3 answers):
| Source (endpoint) | Safe | 1-cycle | 2-cycle | Churned |
|---|---|---|---|---|
| Subscriptions `/subscriptions/summary` (subscriptions.risk_state) | 718 | — (at-risk 19) | | 2189 |
| Dashboard `/metrics` (MetricsEngine recompute, charge-based) | 1175 | 71 | 65 | 1615 |
| Risk `/risk/summary` (stores.risk_state) | 729 | 126 | 5 | 2066 |
- Three code paths compute risk differently and read different stores (subscriptions.risk_state vs stores.risk_state vs on-the-fly MetricsEngine). Need a **single source of truth**. The `/metrics` (charge-based, 1175) is closest to correct; the two stored-state paths are polluted by SUB-1 AND disagree with each other (718 vs 729). Fixing SUB-1/SUB-2 + unifying the risk source should collapse all three to one number.

**RISK-2 🟡 `/risk/summary` returns `installed_app_ids: []`** (empty) while `/stores` populates it — inconsistent store serialization.

**RISK-3 🟡 Negative LTV shown** — `423ca4-5` has `lifetime_value_cents: -2800` (−$28.00; refunds > revenue). Decide display (show negative w/ context, or floor at 0) and confirm it's expected.

(STORE-2 recurs here: every at-risk store's `first_install_date` = 2026-07-30 record time.)

### Earnings (route `/#/earnings`)

**EARN-1 🔴 Summary cards show $0.00 despite real data in API** (frontend field-mapping bug)
- UI: Total Earned / Pending / Available all **$0.00**. But `/earnings/status` returns `total_available_cents: 116,257,226 ($1,162,572)`, `total_pending_cents: 1,615,352 ($16,153)`, `total_paid_out_cents: 0`. Data is there — the frontend isn't mapping the `/earnings/status` fields onto the cards.

**EARN-2 🔴 Earnings list rows show Gross/Shopify/Net = $0.00** — the `/earnings` response rows only carry `{date, total_amount_cents}` (e.g. $1,384.22), but the row widget reads `gross`/`shopify`/`net` fields that don't exist in the payload → renders 0 + "PENDING" for every row. Frontend↔backend contract mismatch.

**✅ EARN-1/EARN-2 FIXED (PR pending) — frontend field-mapping.** `EarningsStatus.fromJson` now reads `total_pending/available/paid_out_cents` + `upcoming_availability`; `totalEarned` sums the status totals in live mode; `EarningPeriod.fromJson` maps `total_amount_cents`→net and `date`→dates; per-date rows show **Net** (Gross/Shopify breakdown only when the source provides it, via `hasFeeBreakdown`).
**✅ EARN-3 FIXED (PR pending) — monthly earnings periods.** New `GET /earnings/periods` endpoint aggregates transactions by month (gross, net, Shopify cut = gross−net, derived status), and the Earnings tab now renders the wireframe's monthly cards (Month + PENDING/AVAILABLE/PAID_OUT badge + Gross / Shopify Fee (rate%) / Net). Replaces the hundreds of daily rows. Needs a backend deploy.

(Known: `total_paid_out_cents: 0` — PAID_OUT never populated, already in future.md.)

**EARN-4 🟡** "Upcoming 30 Days" card lists *every* upcoming-availability entry (hundreds of rows) → very long scroll. Cap/paginate (e.g. top 10 + "view all"). Minor UX.

### Analytics (route `/#/analytics`) — Revenue tab only

Revenue Breakdown: Recurring $62,917 (69.6%) / Usage $27,417 (30.4%) / One-Time $0. MRR Movement bar chart renders (Feb–Jul).

**ANALYTICS-1 🟡** MRR Movement is almost entirely "New" (green) bars with a single Contraction and no Churned/Expansion, despite heavy churn in the data — verify the New/Expansion/Contraction/Churn attribution. Forecasting / Profit & Expense / Cohorts / Multi-App tabs not yet tested.

---

### Reports (route `/#/reports`)

Catalog renders: Retention & Risk (Revenue at Risk, Churn, Retention, Retention Cohorts, Reviews, Uninstall Context), Revenue & Billing (Earnings, MRR, Revenue Mix, Usage & One-Time, Usage Trends, Subscriptions, Payout Schedule, Payout History), Growth (below fold). **Per-report values not yet verified** — several depend on `risk_state` (Revenue at Risk, Churn, Retention, Subscriptions) and earnings, so expect them to move once the root-cause fixes land. Re-test each after fixes.

### Webhooks (route `/#/webhooks`)
Empty: 0 webhook events, "No webhooks match your filters". Likely legitimate (app syncs via Partner API polling, not webhooks).
**WEBHOOK-1 🟡** Success Rate shows **0%** with 0 webhooks — should render "—"/N/A to avoid implying total failure.

### AI Insights (route `/#/insights`)
Daily Briefs: "No insights available yet. Insights are generated daily after your first sync." (Jul 30 2:19 PM). Revenue chat panel renders with input.
**AII-1 🟡** Daily brief empty — verify the daily-brief generation job runs post-sync. Chat not interactively tested.

### Apps (route `/#/apps`)
1 connected app (Zoko — WhatsApp Marketing API), "Synced". Tabs: Connected Apps, Reviews.
**✅ APPS-1 FIXED (PR pending) — install count populated during sync.** `StoreProcessor` (store_sync, runs every full sync) now persists `app.InstallCount = ` distinct installed shops (domains) — the same count the Stores page shows (≈2,926) — instead of leaving it 0 until the manual `RefreshInstallCount` (Partner-API `FetchInstallCount`) is called. Backend; needs deploy + resync. (Count is total distinct shops with activity, matching Stores; not net-active — deliberate, matches user expectation. Sync now **owns** `InstallCount`; the manual Partner-API `RefreshInstallCount`/`FetchInstallCount` net-active path is **superseded** — could be retired later.)

**APPS-1 🔴 "0 installs" despite 2,926 stores/subscriptions** (original finding) — the app-card install count is 0. `FetchInstallCount` (counts `RELATIONSHIP_INSTALLED` events) isn't populating the real number (and likely also misses `RELATIONSHIP_REACTIVATED` reinstalls, cf. SUB-3). Should reflect the true install base.
**APPS-2 🟡** Rating shows ★ 0 — verify against real App Store rating / Reviews tab (not deep-tested).

**✅ APPS-3 FIXED (PR pending) — sync progress now derives from completed/total.** `SyncJob.progress` computes `completed_items/total_items` (clamped, falls back to `progress_pct`); `SyncStatusProvider._poll` picks the furthest-along child job (not the 0/0 parent) and cancels the parent full_sync. Frontend-only.

**APPS-3 🟠 Sync progress stuck at 0% on the Apps screen** (user-reported; ROOT CAUSE CONFIRMED live)
- Symptom: at `/#/apps` during a sync, the "Syncing…" state, progress bar, and Cancel button **do** render — but the bar/percentage is **stuck at 0%** even though the job is well underway (observed `status_sync` at `completed_items 2154 / total_items 2930` ≈ 73%).
- **Root cause (field mismatch, like EARN-1):** `GET /api/v1/sync/jobs?status=processing` returns each job with **`completed_items` / `total_items`** but **no `progress_pct`**; the frontend model `SyncJobStatus.fromJson` reads `json['progress_pct'] ?? 0` → always 0. Compounding: the parent `full_sync` job reports `total_items: 0` (`0/0`), while the child `status_sync`/`event_sync` jobs carry the real counts — so the provider must pick the child (or aggregate), not the parent.
- **Fix (frontend, small):** compute progress = `completed_items / total_items` (guard `total>0`) in `SyncJobStatus`; in `SyncStatusProvider._poll`, prefer the child job with `total_items>0` (or aggregate across children) over the `0/0` parent. Files: `frontend-flutter/lib/services/sync_status_service.dart`, `lib/providers/sync_status_provider.dart`. (Alternative: backend emits `progress_pct` on `/sync/jobs`.)

### API Keys (route `/#/api-keys`)
"3 active keys" (CI/CD Pipeline, Staging, Production) + Old Integration (REVOKED). Scopes/dates/revoke/create all render.
**APIKEYS-1 🟡** Key names/dates look like they could be seed/demo data (CI/CD Pipeline, Staging Key, lg_live_/lg_test_ prefixes) — confirm these are real user-created keys, not seeded fixtures showing in prod.

### Routing note 🟡
Some nav destinations don't match their obvious hash path: `/#/dashboard` → "no routes" (real route `/`), `/#/ai-insights` → "no routes" (real route `/#/insights`). Deep-linking those literal paths 404s. Confirm no internal link/bookmark uses them.

---

## Proposed fix batch (to do together)

**Group A — Risk/churn correctness (biggest impact):**
- **SUB-1** cancel-trap reconciliation: a CANCELLED/UNINSTALLED event churns only if no recurring charge since that event (billed-after-cancel ⇒ active). *(backend, StatusProcessor + needs event timestamp)*
- **SUB-2** re-run `ClassifyRisk` after any status update. *(backend)*
- **RISK-1** unify risk to a single source of truth so Subscriptions / Dashboard / Risk agree. *(backend)*
- Fixes cascade to STORE-1, SD (badges), EVT-1 (plan-change churns), and risk-dependent Reports.

**Group B — Earnings frontend contract:**
- **EARN-1/EARN-2** map `/earnings/status` (available/pending/paid_out) to the summary cards and `total_amount_cents` to the row Net; stop reading non-existent gross/shopify/net fields. *(frontend)*

**Group C — Store detail:**
- **SD-1/SD-2** fetch store (and its subscriptions) by domain server-side instead of client-side page-1 lookup. *(frontend + maybe a store-by-domain endpoint)*
- **SD-3** resolve app name; **STORE-2/SD-4** real install/interaction dates (not record time).

**Group D — Counts & polish:**
- **APPS-1** install count = 0 → populate real install base from `RELATIONSHIP_INSTALLED` (+ REACTIVATED). *(backend)*
- **DASH-2** MRR Trend chart data source; **EVT-2/EVT-3** event badge/title mapping; **RISK-2** store serialization; **RISK-3** negative LTV display; **ANALYTICS-1** MRR-movement attribution; **DASH-3** This Week Activity; **WEBHOOK-1** success-rate N/A; **AII-1** daily-brief job; **APIKEYS-1** verify seed vs real; routing note (`/dashboard`,`/ai-insights` literal paths 404).

## Still to deep-test (top-level nav done)
- The ~16 individual **Reports** (values, CSV export) — several will shift after Group A/B fixes; re-test post-fix.
- **Analytics** sub-tabs: Forecasting, Profit & Expense, Cohorts, Multi-App.
- **Apps → Reviews** tab; AI Insights **chat** (interactive).

---

**Deferred (already in future.md):**
- Transactions `loadMore` swallows page errors; store filter is client-only/page-scoped (PR #45 follow-ups).

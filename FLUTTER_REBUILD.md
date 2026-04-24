# LedgerGuard Flutter Rebuild — Revenue Intelligence Dashboard

## How to Use This File

```bash
cd /Users/sachins/development/ai_projects/ledgerguard
claude "Read FLUTTER_REBUILD.md and follow the steps. Start with Step 1."
```

---

## Steps

### Step 1: Plan
Read all reference files listed below, then create a detailed implementation plan
with every file listed, grouped by phase with dependencies. Get my approval before
building anything.

### Step 2: Build
After plan approval, implement phase by phase. Run `flutter analyze` after each
phase to catch issues early. Run `flutter build web` at the end.

### Step 3: Verify
- `flutter analyze` — zero issues
- `flutter build web` — succeeds
- `flutter run -d chrome` — app launches, all nav works, all screens render

---

## Project Spec

### What to Build
A new Flutter web admin dashboard at `ledgerguard-flutter/` (sibling to existing
`ledgerguard/`). Mock data only, no backend. This is a UI prototype of the full
LedgerGuard platform with enhanced features.

### Design System Reference
Port the Polaris-inspired design system from the BookFlow Flutter build:
`/Users/sachins/development/ai_projects/idea-2/bookflow-flutter/`
- `lib/theme/app_theme.dart` — Material 3 theme with Polaris-scale typography
- `lib/theme/app_colors.dart` — Semantic color palette
- `lib/theme/app_spacing.dart` — 4px-base spacing scale (s100=4 through s1600=64)
- `lib/widgets/` — 9 reusable widgets (page wrapper, card, badge, metric card,
  data table, search field, status badge, empty state, confirmation dialog)

Adapt with LG prefix (LgPage, LgCard, LgBadge, LgMetricCard, etc). Keep the same
Polaris typography scale (headlineSmall=20px, titleMedium=16px, bodyMedium=13px,
bodySmall=12px) and 4px spacing grid.

### Feature Reference Files
Read these to understand the domain and existing features:
- PRD: `/Users/sachins/development/ai_projects/ledgerguard/PRD.md`
- Existing pages: `/Users/sachins/development/ai_projects/ledgerguard/frontend/app/lib/presentation/pages/`
- Existing models: `/Users/sachins/development/ai_projects/ledgerguard/frontend/app/lib/data/models/`
- Existing widgets: `/Users/sachins/development/ai_projects/ledgerguard/frontend/app/lib/presentation/widgets/`
- Shopify app ideas: `/Users/sachins/development/ai_projects/idea-2/shopify-app-ideas.md`

---

## Screens

### Core Screens (port from existing app)

1. **Dashboard** — MRR, Renewal Rate, Revenue at Risk, Usage Revenue metrics +
   MRR trend chart + risk distribution donut + recent alerts + forecast card

2. **Subscriptions** — List with filters (status, risk state, plan), search,
   detail view with billing timeline

3. **Store Health** — Per-store health cards, risk state badges, revenue contribution

4. **Risk Breakdown** — Risk funnel (SAFE → ONE_CYCLE_MISSED → TWO_CYCLE_MISSED → CHURNED),
   at-risk store list with days overdue, recovery actions

5. **Analytics** — Revenue trends (recurring vs usage vs one-time), renewal rate over time,
   churn analysis, MRR movement (new/expansion/contraction/churned)

6. **AI Insights** — Daily revenue brief cards, chat interface for asking about revenue

7. **Settings** — Notification preferences (alert types, channels, thresholds),
   sync schedule, workspace config

8. **Apps** — Connected Shopify apps list, sync status, last sync time

9. **Transactions** — Full ledger view of all revenue events. Columns: date, store,
   type (RECURRING/USAGE/ONE_TIME/REFUND), app, amount, net amount (after Shopify's
   20% cut). Filters: type, app, store, date range. Summary row showing totals.
   Export button (stub). Drill-down to subscription detail on click.

### Enhanced Features (from idea-2 analysis)

10. **Profit & Expense Overlay** (idea #10 — Expense & Profit Analytics)
    Layer Shopify's 20% revenue share, payment processing fees, and infrastructure
    costs onto revenue data. Show net profit per app. Add "Profit Margin" metric
    to dashboard. Monthly P&L summary card. Accessible as a tab within Analytics.

11. **Revenue Forecasting** (idea #2 — Smart Inventory Forecasting pattern)
    Predict next month's MRR based on renewal pipeline, at-risk subs, and historical
    trends. Confidence range (optimistic/expected/pessimistic). "Expected MRR next
    month: $5,200–$5,800" card on dashboard. 12-month forecast chart.
    Accessible as a tab within Analytics.

12. **Churn Recovery Playbooks** (idea #13 — Abandoned Cart Recovery pattern)
    When a store enters ONE_CYCLE_MISSED, show suggested recovery actions:
    email templates, discount offers, feature highlights. Track which actions
    were taken and whether the store recovered. Recovery success rate metric.
    Accessible within Risk Breakdown and Store detail views.

13. **Store CRM Timeline** (CRM gap — only 11 apps in dataset)
    Per-store relationship page: install date, every transaction, risk state
    changes over time, notes. Like a CRM contact page but for stores that
    use your app. "Last interaction" and "lifetime value" per store.
    Accessed by tapping a store from Store Health or Subscriptions.

14. **App Review Monitor** (idea #9 — Social Proof & Review Aggregator pattern)
    Track Shopify app store ratings and reviews. Correlate negative reviews
    with churn spikes. Alert on new 1-star reviews. Review sentiment trend chart.
    Accessible as a tab within the Apps screen.

15. **Multi-App Comparison** (idea #12 — Multi-Channel Sync pattern)
    Side-by-side comparison across your apps: which app has better renewal rate,
    faster growth, more at-risk revenue. Helps developers decide where to invest.
    Accessible as a tab within Analytics.

16. **Cohort Retention** (idea #20 — Benchmarking pattern)
    Group stores by install month. Track retention curves per cohort.
    "Stores installed in Jan have 85% 6-month retention vs 72% for March."
    Identify which cohorts are healthiest and investigate why.
    Accessible as a tab within Analytics.

---

## Navigation Structure

Sidebar with 9 destinations:
1. Dashboard `/` — overview with KPI cards, trend chart, alerts, forecast
2. Subscriptions `/subscriptions` — list + detail `/subscriptions/:id`
3. Stores `/stores` — health list + detail/CRM `/stores/:id`
4. Transactions `/transactions` — full ledger
5. Risk `/risk` — breakdown funnel + recovery playbooks
6. Analytics `/analytics` — tabs: Revenue, Forecasting, Profit, Cohorts, Multi-App
7. Apps `/apps` — connected apps + reviews tab
8. AI Insights `/insights` — briefs + chat
9. Settings `/settings` — notifications, sync, workspace

---

## Architecture

- **State**: Provider + ChangeNotifier (NOT Bloc — this is a UI prototype)
- **Navigation**: GoRouter with ShellRoute (NavigationDrawer / NavigationRail)
- **Charts**: fl_chart (bar charts, line charts, area charts, pie/donut)
- **Data**: All mock — no Firebase, no Dio, no backend
- **Structure**: Flat layout:
  ```
  ledgerguard-flutter/lib/
  ├── theme/        → app_theme.dart, app_colors.dart, app_spacing.dart
  ├── widgets/      → lg_page, lg_card, lg_badge, lg_metric_card, etc.
  ├── models/       → subscription, store, transaction, analytics, etc.
  ├── mock_data/    → all mock data files
  ├── providers/    → ChangeNotifier providers
  ├── screens/      → one folder per screen group
  ├── shell/        → app_shell.dart (nav drawer/rail)
  ├── app.dart      → MaterialApp.router + GoRouter
  └── main.dart     → MultiProvider entry point
  ```

---

## Mock Data Requirements

Generate realistic mock data matching the PRD's domain:
- 3 Shopify apps being tracked (e.g. "InventorySync Pro", "ReviewBoost", "ShipTracker")
- ~40 subscriptions across risk states (25 SAFE, 8 ONE_CYCLE_MISSED,
  4 TWO_CYCLE_MISSED, 3 CHURNED)
- ~120 transactions across 3 months (mix of all 4 types, weighted toward RECURRING)
- 12 months of daily MRR snapshots (trending upward with realistic dips)
- 15 stores with varying health scores and CRM timeline events
- 5 AI insight briefs (recent days)
- Revenue mix: ~80% recurring, ~15% usage, ~5% one-time
- 12-month forecast data (optimistic/expected/pessimistic)
- 6 monthly cohorts with retention percentages declining over time
- 3 churn recovery playbook templates with mock outcomes
- 15 app reviews across 3 apps (mix of 1-5 stars, recent dates)
- Expense/cost data: Shopify 20% cut, $49/mo infrastructure, 2.9% payment processing

---

## Lessons from BookFlow Flutter Build

Avoid these issues encountered in the BookFlow build (`bookflow-flutter/STATUS.md`):
- **Do NOT** wrap `DataTable` in `SizedBox(width: double.infinity)` — crashes inside horizontal ScrollView
- **Do NOT** use deprecated `DropdownButtonFormField.value` — use `initialValue` (Flutter 3.33+)
- **Do NOT** use deprecated `RadioListTile.groupValue/onChanged` — use `RadioGroup` wrapper (Flutter 3.32+)
- Run `flutter analyze` after EACH phase, not just at the end
- Test every screen in browser after building — compile success ≠ runtime success

## Excluded

- Authentication / login screens (show the app directly)
- Shopify OAuth flow
- Real API integration
- Mobile-specific layouts (web dashboard only)

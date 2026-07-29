# LedgerGuard — Reports

**Status:** Spec (Revenue at Risk = first build)
**Created:** 2026-07-26
**Owner:** Product
**Related:** [[future.md]] · `docs/wireframes/19-reports.svg`, `19a-report-revenue-at-risk.svg` · `docs/diagrams/puml/45-report-revenue-at-risk-flow.puml`

---

## 1. Why Reports

Today the app mostly shows **raw lists** (subscriptions, transactions, events). For a Shopify app **partner**, lists are thin — they don't want to scroll, they want **answers**: *Am I churning? Where's my money? Who's my best customer? Did Shopify pay me correctly?*

We already have the **data** (subscriptions with risk state + MRR, transactions with RECURRING/USAGE/ONE_TIME/REFUND + full fee breakdown, app events, daily snapshots) **and the compute engines** (`risk_engine`, `metrics_engine`, `forecasting_engine`, `earnings_calculator`, `fee_verification_service`). Reports is largely a **packaging + light-aggregation + export** play, not a data-collection project.

### Positioning vs Mantle
Mantle is the reference (category rail + card grid). We can match ~13 of its ~19 reports with existing data. The gaps (feature engagement, uninstall *reasons*, discounts, affiliates) are structural — Mantle **embeds an SDK / uninstall survey**; LedgerGuard is **read-only partner analytics**. We don't fake those. Instead we lean into two things Mantle can't do as well:

- **Deterministic risk** — our `risk_engine` (missed-cycle detection) → *Revenue at Risk* as a hard $ number, not a black-box score.
- **🛡️ Guard / reconciliation** — verify Shopify paid you correctly (`fee_verification_service`). This is the product's namesake and a genuine wedge.

---

## 2. Information architecture

A top-level **Reports** nav item → a Reports landing page modeled on Mantle: a **category rail** + **card grid** + filter + "Custom report" CTA. Each card opens a **report page** with a date-range picker, segment filters, states (loading/empty/error), and **CSV export**.

### Categories & catalog (✅ now · 🟡 light aggregation · 🔴 needs data we don't have)

**Revenue & Billing**
| Report | Answers | Feasibility | Reuses |
|---|---|---|---|
| Earnings | net take-home; pending/available/paid | ✅ | `earnings_calculator` |
| Monthly Recurring Revenue | MRR growth/shrink over time | ✅ | daily snapshot |
| Revenue Mix | RECURRING vs USAGE vs ONE_TIME split | ✅ | transactions |
| Usage & One-Time Charges | usage/one-time billed | ✅ | transactions |
| Usage Trends | usage over time, top usage stores | ✅ | transactions |
| Subscriptions (ARPU / LTV) | active subs, ARPU, LTV | ✅ **Shipped** (PR) | subs + snapshot |
| Payout Schedule | upcoming Shopify payouts + when | ✅ **Shipped** (PR) | `AvailableDate`, `EarningsStatus` |
| Payout History | historical payouts by app/period | ✅ **Shipped** (PR) | paid transactions |
| Discounts | discount impact on revenue | 🔴 | not synced |

**Retention & Risk**
| Report | Answers | Feasibility | Reuses |
|---|---|---|---|
| **Revenue at Risk** ⭐ | MRR about to churn, ranked | ✅ | `risk_engine`, snapshot |
| Churn | churned count + MRR lost, churn % | ✅ | risk state, events, snapshot |
| Retention / Renewal | renewal-success trend | ✅ | snapshot `RenewalSuccessRate` |
| Retention Cohorts | retention by signup month | ✅ | `CohortHandler` |
| Reviews | ratings + sentiment | ✅ | `app_review` |
| Uninstall Context | state *before* churn (not self-reported reason) | 🟡 | events + subs join |
| Uninstall Reasons | merchant-stated reason | 🔴 | needs a survey (read-only) |

**Growth**
| Report | Answers | Feasibility |
|---|---|---|
| Installs | installs vs uninstalls trend | ✅ **Shipped** (PR) — RELATIONSHIP_INSTALLED/UNINSTALLED events |
| Activation | install → paid conversion | 🟡 (events↔subs join) |
| Net-New Subscriptions | new subs per period | ✅ **Shipped** (PR) — subs created/churned in range |

**Customers** *(the feasible slice of Mantle's Engagement/Customers — the rest needs a Pixel/SDK we don't have)*
| Report | Answers | Feasibility | Reuses |
|---|---|---|---|
| Active Customers | active paying customers over time | ✅ | snapshot `TotalSubscriptions`/`SafeCount` |
| Customer Insights (Segments) | slice customers by plan / Shopify plan / risk / revenue | 🟡 | subs + `transactions.shop_plan` |
| Trials | trial → paid conversion | 🔴 | no trial data synced (only inferrable) |
| Engagement / Web / Marketing / Email | feature usage, pixel, attribution | 🔴 | needs Mantle-style SDK / Pixel / Email — out of scope for read-only |

**🛡️ Guard (Reconciliation)** — *our differentiator*
| Report | Answers | Feasibility | Reuses |
|---|---|---|---|
| Fee Audit | were you charged the correct revenue-share tier? | ✅ | `fee_verification_service` |
| Payout Accuracy | did every charge pay out / any missing? | ✅ | txns + earnings status |
| Ledger Reconciliation | recomputed ledger vs Shopify figures | ✅ | `ledger_service` |

**Later (needs new syncs):** Affiliates/Referrals (Partner API has referrals — new sync), Engagement (needs in-app SDK — out of scope for read-only).

### Product inventions layered on top
- **"Recoverable revenue"** — not all at-risk MRR churns. Estimate recoverable $ by weighting risk states with historical recovery rate (1-cycle recovers far more than 2-cycle). Turns a scary number into an *actionable* one.
- **Export-first** — CSV now, PDF later. The real partner value is sending a report to a board/accountant.
- **Scheduled/emailed reports** — weekly "Revenue at Risk" digest (reuses the notification engine).
- **Threshold alerts** — notify when Revenue at Risk crosses $X or jumps N% WoW.
- **Custom reports** — saved filter/column sets (Mantle parity), P2.

---

## 3. First report — **Revenue at Risk** ⭐

### 3.1 Purpose & audience
**Purpose:** Show the partner exactly how much recurring revenue is *about to churn* and which stores to act on **now**. It is the bridge between the passive Risk screen and money.
**Audience:** Founder / growth / success at a Shopify app partner.
**Job-to-be-done:** *"Tell me the $ I'm about to lose, who's causing it, how much I can still save, and what to do."*

### 3.2 Data sources (all existing)
| Field | Source |
|---|---|
| Risk classification (Safe / 1-cycle / 2-cycle / Churned) | `risk_engine.ClassifyRisk(status, expectedNextCharge, now)` |
| MRR per subscription | `subscription.BasePriceCents` |
| Days late / expected next charge | `subscription.ExpectedNextChargeDate`, `LastRecurringChargeDate` |
| Store, plan, tenure | `subscription.MyshopifyDomain`, `ShopName`, `PlanName`, `CreatedAt` |
| At-risk MRR time series | `daily_metrics_snapshot.RevenueAtRiskCents` (already stored daily) |
| Risk-state counts over time | snapshot `OneCycleMissedCount`, `TwoCyclesMissedCount` |

### 3.3 Computation
```
atRisk        = Σ BasePriceCents for subs where riskState ∈ {ONE_CYCLE_MISSED, TWO_CYCLES_MISSED}
byState       = { oneCycle: Σ MRR (1-cycle), twoCycle: Σ MRR (2-cycle) }
recoverable   = oneCycleMRR * r1 + twoCycleMRR * r2      // r1,r2 = historical recovery rates (default r1=0.6, r2=0.25; later learn from reactivation events)
trend[]       = daily_metrics_snapshot.RevenueAtRiskCents over [from,to]
stores[]      = subs in at-risk states, sorted by MRR desc, each with {domain, mrr, riskState, daysLate, expectedChargeDate, recoverableCents}
```
- **Deterministic + idempotent** — matches the ledger-rebuild philosophy (same input → same output).
- Recovery rates start as constants, then get **learned** from `RELATIONSHIP_REACTIVATED` / re-charge events (invention: closes the loop).

### 3.4 API contract
```
GET /api/v1/apps/{appID}/reports/revenue-at-risk?from=YYYY-MM-DD&to=YYYY-MM-DD&segment=all|plan:<name>
Auth: Firebase + org context.  (Aggregate variant: /api/v1/reports/revenue-at-risk across all apps.)

200 {
  "currency": "USD",
  "totalAtRiskCents": 184200,
  "recoverableCents": 98400,
  "byState": { "oneCycleCents": 120000, "twoCycleCents": 64200 },
  "counts": { "oneCycle": 8, "twoCycle": 4 },
  "trend": [ { "date": "2026-07-01", "atRiskCents": 150000 }, ... ],
  "stores": [
    { "domain": "acme-widgets.myshopify.com", "shopName": "Acme",
      "mrrCents": 4900, "riskState": "TWO_CYCLES_MISSED", "daysLate": 47,
      "expectedChargeDate": "2026-06-09", "planName": "Pro",
      "recoverableCents": 1225 }
  ]
}
```
CSV export: `?format=csv` → the `stores[]` rows (one line per store).

### 3.5 UI (see `19a-report-revenue-at-risk.svg`)
Follows app standards (LgPage: title + subtitle + refresh, app filter chip, cards `rx=8` stroke `#E5E7EB`, risk color tokens):
1. **Header** — "Revenue at Risk" + subtitle; **date-range** picker; **Export CSV**; app filter chip.
2. **Hero row (3 KPI cards)** — *Total at Risk* ($), *Recoverable* ($, green), *At-Risk Stores* (count, split 1-/2-cycle).
3. **Trend card** — at-risk MRR over the range (from snapshots), with a subtle "recoverable" band.
4. **Ranked table** — stores sorted by MRR at risk: Store · MRR · Risk badge (amber/orange) · Days late · Expected charge · Recoverable · `›` (→ `/stores/:id`).
5. **States** — loading skeletons; **empty** = green check + "No revenue at risk 🎉"; error = LgServiceUnavailable pattern.

### 3.6 Non-goals / honesty
- No *reason* for lateness (Shopify doesn't expose it) — we show **state + days-late + tenure** as context.
- Recovery estimate is a **model**, labeled as such, not a promise.

### 3.7 Acceptance
- [ ] Sum of `stores[].mrrCents` (at-risk states) == `totalAtRiskCents`
- [ ] `totalAtRiskCents` matches latest snapshot `RevenueAtRiskCents` (± same-day drift)
- [ ] Trend renders from snapshots; empty range → empty state
- [ ] CSV export downloads the ranked stores
- [ ] Deterministic: same data → same numbers

---

## 4. Build order
- **P0:** Reports shell (nav + landing + category rail + card grid) · **Revenue at Risk** · **Fee Audit (Guard)**
- **P1:** Churn · MRR · Earnings · Usage · Payout Schedule · Cohorts (surface existing)
- **P2:** CSV→PDF · scheduled/emailed reports · threshold alerts · Custom reports · segment drill-downs

---

## 5. Wireframe & archetype library

Every report is an **instance of an archetype**. To introduce a new report: pick the archetype → clone its wireframe → reuse its PUML data-flow → define the data contract. This is the reusable design system.

**Archetype flows (PUML):**
`46` Trend+Table (A) · `47` Composition (B) · `48` Cohort (C) · `49` Schedule (D) · `50` Funnel (E) · `51` Reconciliation (F) · `52` List+Sentiment (G) — all under `docs/diagrams/puml/`.

| Report | Category | Archetype | Wireframe | Flow PUML |
|---|---|---|---|---|
| **Revenue at Risk** ⭐ | Retention & Risk | A | `19a` | `45` (+`46`) |
| Churn | Retention & Risk | A | `19j` | `46` |
| Retention | Retention & Risk | A/C | `19k` | `46`/`48` |
| Retention Cohorts | Retention & Risk | C | `19l` | `48` |
| Reviews | Retention & Risk | G | `19m` | `52` |
| Uninstall Context | Retention & Risk | G | `19n` | `52` |
| Earnings | Revenue & Billing | D | `19b` | `49` |
| Monthly Recurring Revenue | Revenue & Billing | A | `19c` | `46` |
| Revenue Mix | Revenue & Billing | B | `19d` | `47` |
| Usage & One-Time Charges | Revenue & Billing | A/B | `19e` | `46`/`47` |
| Usage Trends | Revenue & Billing | A | `19f` | `46` |
| Subscriptions (ARPU/LTV) | Revenue & Billing | B | `19g` | `47` |
| Payout Schedule | Revenue & Billing | D | `19h` | `49` |
| Payout History | Revenue & Billing | D | `19i` | `49` |
| Installs | Growth | A | `19o` | `46` |
| Activation | Growth | E | `19p` | `50` |
| Net-New Subscriptions | Growth | A | `19q` | `46` |
| Active Customers | Customers | A | `19u` | `46` |
| Customer Insights (Segments) | Customers | B | `19v` | `47` |
| 🛡️ Fee Audit | Guard | F | `19r` | `51` |
| 🛡️ Payout Accuracy | Guard | F | `19s` | `51` |
| 🛡️ Ledger Reconciliation | Guard | F | `19t` | `51` |

Index page: `docs/wireframes/19-reports.svg`.

---

## 6. Build status (source of truth — keep updated as reports ship)

| Report | Status |
|---|---|
| **Revenue at Risk** | ✅ **Shipped** — `backend/.../revenue_at_risk_handler.go` + `frontend-flutter/.../revenue_at_risk_screen.dart`; live on app.ledgerspear.com |
| **Churn** | ✅ **Shipped** — `backend/.../churn_handler.go` (12 tests) + `frontend-flutter/.../churn_screen.dart`; churn %, MRR lost, ranked churned stores, trend, CSV. Merged to main (PR #4). |
| **Retention / Renewal** | ✅ **Shipped** — `backend/.../retention_handler.go` (15 tests) + `frontend-flutter/.../retention_screen.dart`; renewal rate + trend, retained MRR, reactivations, Retention-by-Plan table, CSV. Merged to main (PR #5). |
| **Retention Cohorts** | ✅ **Shipped** — hardened `backend/.../cohort_handler.go` (13 tests, was 0; M0 baseline fix, deterministic churn date, 500→503, CSV, /reports/cohorts alias) + `frontend-flutter/.../cohorts_screen.dart` + reusable `CohortHeatmap` widget. Merged to main (PR #6). |
| **Reviews** | ✅ **Shipped** — `backend/.../review_report.go` (16 tests, was 0; new `FindAllByAppID` repo method, avg/distribution/sentiment/recent, out-of-range guard, 503, CSV) + `frontend-flutter/.../reviews_screen.dart`, reusing the `AppReview` model. Merged to main (PR #7). |
| **Uninstall Context** | ✅ **Shipped** — `backend/.../uninstall_context_handler.go` (16 tests; UNINSTALL* events ↔ subs via ShopifyShopGID, inferred pre-uninstall state, %at-risk-first, median tenure, 503, CSV) + `frontend-flutter/.../uninstall_context_screen.dart`. Merged to main (PR #8). **Retention & Risk category complete.** |
| **Earnings** | ✅ **Shipped** — `backend/.../earnings_report_handler.go` (12 tests + calculator unit test; reuses EarningsCalculator; Net/Pending/Available/Paid Out reconcile, per-charge timeline, 503, CSV) + `frontend-flutter/.../earnings_report_screen.dart`. Merged to main (PR #9). |
| **Monthly Recurring Revenue** | ✅ **Shipped** — `backend/.../mrr_report_handler.go` (14 tests; ActiveMRR + signed unclamped MoM, new/churned movement, MRR-by-plan; 503, CSV) + `frontend-flutter/.../mrr_report_screen.dart`. Merged to main (PR #10). |
| **Revenue Mix** | ✅ **Shipped** — `backend/.../revenue_mix_report_handler.go` (13 tests; RECURRING/USAGE/ONE_TIME split, refund adjustment, gross/net, composition segments; unknown-type logged; 503, CSV) + `frontend-flutter/.../revenue_mix_screen.dart` (stacked bar + breakdown). Merged to main (PR #14). |
| **Usage & One-Time Charges** | ✅ **Shipped** — `backend/.../usage_report_handler.go` (10 tests; USAGE+ONE_TIME only, per-store table, snapshot usage trend; 503, CSV) + `frontend-flutter/.../usage_screen.dart`. Merged to main (PR #15). |
| **Usage Trends** | ✅ **Shipped** — `backend/.../usage_trends_report_handler.go` (13 tests; USAGE-only ISO-week (Monday-anchored) buckets, signed unclamped WoW app- + per-store level, distinct active-store count; 503, CSV) + `frontend-flutter/.../usage_trends_screen.dart` (Usage MRR-equiv / WoW / active-stores KPIs, weekly trend, top-usage-customers table). Merged to main (PR #16). |
| **Subscriptions (ARPU / LTV)** | ✅ **Shipped** — `backend/.../subscriptions_report_handler.go` (13 tests; Archetype B composition. Active = SAFE subs; ARPU = ActiveMRR ÷ active subs (floored); LTV = ARPU ÷ monthly churn rate (shared `churnRate` helper → consistent with Churn report, rounded, 0/"—" when churn 0); per-plan ARPU/LTV + share-of-subs; 503, CSV) + `frontend-flutter/.../subscriptions_screen.dart` (Active-subs / ARPU / LTV KPIs, subscriptions-by-plan horizontal bars, plan-detail table). Merged to main (PR #17). |
| **Payout Schedule** | ✅ **Shipped** — `backend/.../payout_schedule_report_handler.go` (11 tests; Archetype D schedule/timeline. Upcoming (PENDING+AVAILABLE) net earnings grouped by (AvailableDate, status); PAID_OUT excluded [→ Payout History], unknown status logged+dropped, KPIs reconcile with rows; Upcoming/Next-date/Pending KPIs; rows sorted scheduled-asc, Available-before-Pending, unscheduled last; 503, CSV) + `frontend-flutter/.../payout_schedule_screen.dart` (3 KPIs + upcoming-payouts timeline with status pills). Merged to main (PR #18). |
| **Payout History** | ✅ **Shipped** — `backend/.../payout_history_report_handler.go` (14 tests; Archetype D completed-payouts. PAID_OUT-only [PENDING/AVAILABLE → Payout Schedule], grouped by charge-month period (keyed off the Shopify charge date `CreatedDate`→`TransactionDate`, NOT ingestion `CreatedAt`); Total-Paid / Payout-count / Avg-payout [floored] KPIs; per-period amount + charge count + availableDate=latest availability date in period (est., not Shopify's disbursement date); rows sorted period-desc; reuses shared `earningsCurrency`; 503, CSV) + `frontend-flutter/.../payout_history_screen.dart` (3 KPIs + payout-log timeline). Merged to main (PR #19). **NOTE: PAID_OUT is not populated by the pipeline yet (P2 in future.md), so this report is empty with live data until that path exists.** **Revenue & Billing category COMPLETE (9 reports).** |
| **Installs** | ✅ **Shipped** — `backend/.../installs_report_handler.go` (11 tests; Growth Archetype A. Install/uninstall counts + net from RELATIONSHIP_INSTALLED/UNINSTALLED app-events [UNINSTALL matched FIRST — "UNINSTALLED" contains "INSTALL"]; daily two-series trend; recent-events table [newest-first, capped 50, domain resolved via sub ShopifyShopGID join]; date-window boundary; 503×2, 401, CSV) + `frontend-flutter/.../installs_screen.dart` (Installs/Uninstalls/Net KPIs, two-line trend, events table with pills). Merged to main (PR #20). |
| **Net-New Subscriptions** | ✅ **Shipped** — `backend/.../net_new_subs_report_handler.go` (Growth Archetype A. New (subs **StartDate()=activated_at** in range — see ADR-044; NOT record-created CreatedAt) vs churned (churnedDateOf in range, UpdatedAt fallback for nil-charge-date churns) + net; start-and-churn-same-period counts both; daily new/churned/net trend; recent-new-subs table [newest-first, capped 50, UI "N of M", plan/MRR/started]; currency first-non-empty break-loop [no sentinel collision]; date boundary; 503, 401, CSV) + `frontend-flutter/.../net_new_subs_screen.dart` (New/Churned/Net KPIs, net-new bar chart, new-subs table). Merged to main (PR #21). |
| **Active Customers** | ✅ **Shipped** — `backend/.../active_customers_report_handler.go` (6 tests; Customers · Archetype A. Headline = current active (non-churned = SAFE + at-risk) from latest in-range snapshot, else current non-churned sub count; New (StartDate/activated_at in range) / Churned (churn date in range) / Net-Change movement; active-customers trend from daily snapshots (Total − CHURNED, adaptive granularity); "active by plan" table (count + MRR + share of active count, sorted MRR desc, churned excluded); 503, CSV) + `frontend-flutter/.../active_customers_screen.dart` (3 KPIs + green trend + active-by-plan table). |
| Activation | ⬜ Pending (Growth — Archetype E funnel, events↔subs join) |
| Customer Insights | ⬜ Pending (Customers) |
| 🛡️ Fee Audit · Payout Accuracy · Ledger Reconciliation | ⬜ Pending (Guard — differentiators) |

**Reports shell** (nav + index card grid) ✅ shipped with Revenue at Risk. Each pending report = clone its archetype (see §5), add a backend `/reports/<name>` endpoint + a Provider screen, wire into `DemoModeCoordinator`.

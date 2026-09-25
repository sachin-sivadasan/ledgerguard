# PRD: Reports Platform (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD.** Reverse-derived from shipped code, not written before it. Markers:
> **FACT** = the code proves it · **INFERRED** = behavior implies it · **UNKNOWN** = rationale
> not recorded (→ Open question). Success metrics are **PROPOSED** (future), not historical.
> Primary sources: `internal/interfaces/http/handler/*_report_handler.go`,
> `domain/service/{metrics_engine,ledger_service,earnings_calculator,fee_verification_service}`,
> `frontend-flutter/lib/screens/reports/`, and the pre-build spec `docs/REPORTS.md`.

## Problem & evidence

A Shopify app **partner** using LedgerGuard needs *answers* about their revenue — "Am I churning? Where's my money? Is Shopify paying me correctly?" — not raw lists.

- **INFERRED (from `docs/REPORTS.md §1`):** the app previously exposed mostly raw lists (subscriptions, transactions, events); Reports was built to package the existing data + compute engines into decision-ready views. This is design reasoning captured in-repo.
- **UNKNOWN — original demand evidence not recorded.** No support tickets, churn reasons, or usage metrics are stored that prove partners asked for this. `docs/REPORTS.md` cites competitive positioning vs **Mantle** as the rationale, which is a design argument, not demand data.
- Open question: validate demand with real partner signal — owner: Product.

## Target users

**FACT/INFERRED (`docs/REPORTS.md §3.1`):** founder / growth / customer-success at a **Shopify app partner** — a developer earning recurring + usage revenue through the Shopify Partner program, viewing **per-app** analytics within an org workspace.
- **Not the target:** Shopify *merchants* (store owners); partners wanting in-app product-engagement analytics (pixel/SDK feature usage) — deliberately out of scope (see Non-goals).

## Proposed solution (as-built behavior)

**FACT.** A top-level **Reports** section (`frontend-flutter/lib/screens/reports/reports_screen.dart`) presents a catalog of reports grouped into categories. Each report opens a page with a **date-range**, an optional **segment filter**, loading/empty/error states, a **preview table**, and a linked **detail page** for the full paged list. Every report offers **CSV export**.

Report catalog implemented today (**FACT** — one `*_report_handler.go` + screen each, all with `_test.go`):
- **Revenue & Billing:** MRR, Revenue Mix, Usage & One-Time, Usage Trends, Subscriptions (ARPU/LTV), Earnings, Payout Schedule, Payout History.
- **Retention & Risk:** Revenue at Risk ⭐, Churn, Retention/Renewal, Retention Cohorts, Reviews, Uninstall Context.
- **Growth:** Installs, Activation, Net-New Subscriptions.
- **Customers:** Active Customers, Customer Insights (segments).
- **🛡️ Guard (reconciliation — the product's namesake wedge):** Fee Audit, Payout Accuracy, Ledger Reconciliation.

Cross-cutting behaviors (**FACT**):
- **Two-tier paging** (`report_paging.go`): report page requests a small preview (e.g. `limit=8`); the detail page requests a full window (`limit`/`offset`, clamped to **200/page**); `limit<=0` returns all rows (CSV/legacy). **KPIs and trends are always computed over the full dataset, independent of paging.**
- **Product inventions layered on raw data:** "Recoverable revenue" (weights at-risk MRR by historical recovery rate) in Revenue at Risk; deterministic/idempotent computation matching the ledger-rebuild philosophy.

### Key screens

**No new wireframes** — this is as-built; we reference the real screens (per the CHECKMATE adaptation, wireframes only if we intend to change the UI). Information architecture (FACT, from `reports_screen.dart` + screen file pairs):

```
Reports hub (reports_screen.dart)
  └─ Category groups → report cards
       └─ Report page  (e.g. mrr_report_screen.dart)      → KPIs + trend + preview table + CSV
            └─ Detail page (e.g. revenue_at_risk_stores_screen.dart) → full paged list
```
Representative real routes: `mrr_report_screen.dart`, `revenue_at_risk_screen.dart` → `revenue_at_risk_stores_screen.dart`, `fee_audit_screen.dart`, `ledger_recon_screen.dart`, `cohorts_screen.dart`.

## Platform & policy constraints

**FACT/INFERRED.** Reports is **read-only partner analytics** over the Shopify **Partner API** + App Store data. Hard data-availability ceilings (see `docs/REPORTS.md §2`): no discount data, no trial data, no in-app engagement/pixel data, no merchant-stated uninstall *reasons* are synced — so those reports cannot exist without new data sources (a survey/SDK the product deliberately avoids). Partner API version support-window + rate limits apply upstream ([[shopify-partner-api-gotchas]]). Auth: **Firebase + org context** on every report endpoint (**FACT**, e.g. `mrr_report_handler.go`).

## Pricing-tier impact

**UNKNOWN — no plan-gating logic for reports found in the handlers.** All reports appear available to any authenticated org member. Open question: should Guard/reconciliation or advanced reports be a paid tier? — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Reports is additive: a new nav section layered on existing subscription/transaction/event data. No data migration or behavior change to existing screens observed in code.

## Success metrics (PROPOSED — future-facing, not historical)

These are how we'd know the platform *still works / improves*; none are measured today.
1. **Adoption:** ≥ 60% of active orgs open ≥ 1 report per week within 30 days of a partner connecting an app.
2. **Reconciliation value:** ≥ 1 in 5 Fee Audit / Ledger Reconciliation runs surfaces a discrepancy the partner acknowledges (proves the Guard wedge).
3. **Determinism guard:** a report run on unchanged underlying data returns byte-identical KPIs/trends on re-run (regression check — enforces the idempotency FACT).

## Non-goals (observed as deliberately absent in code)

1. **No engagement/pixel/SDK analytics** — feature usage, web/marketing attribution (needs an SDK LedgerGuard doesn't ship). **FACT** (absent).
2. **No merchant-stated uninstall *reasons*, discounts, or trial reports** — underlying data not synced. **FACT.** (Uninstall *Context* — pre-churn state — exists; reasons do not.)
3. **No PDF export** — CSV only today (`format=csv`). **FACT.**
4. **No scheduled/emailed reports and no custom/saved reports yet** — noted as later/P2 in `docs/REPORTS.md`. **INFERRED.**

## Open questions

- Validate original demand (tickets/usage) — no evidence recorded. Owner: Product.
- Plan/pricing gating for reports (esp. Guard) — none in code. Owner: Product.
- Recoverable-revenue recovery rates are constants (`r1≈0.6, r2≈0.25`); intended to be *learned* from reactivation events — is that still the plan? Owner: Eng/Product.
- Cross-app aggregate variant (`/api/v1/reports/{report}`) — confirmed for Revenue at Risk in `docs/REPORTS.md`; **INFERRED** whether all reports expose it. Owner: Eng.
- Which reports are considered "shipped/verified" vs partial — source of truth is `docs/REPORTS.md §6 build-status`; keep this PRD pointing there rather than duplicating. Owner: Product.

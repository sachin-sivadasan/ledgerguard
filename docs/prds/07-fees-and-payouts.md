# PRD: Fees, Payouts & Reconciliation (as-built)

**Date:** 2026-09-25 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** (code proves it),
> **INFERRED** (behavior implies it), **UNKNOWN** (rationale not recorded → Open question).
> Success metrics are **PROPOSED** (future), not historical. Primary sources:
> `backend/internal/domain/service/{fee_verification_service,earnings_calculator}.go`,
> `valueobject/revenue_share_tier.go`, `handler/{fee_audit,payout_schedule,payout_history,earnings,ledger_reconciliation}_report_handler.go`,
> `handler/fee_shared.go`, `frontend-flutter/lib/screens/reports/`. (Report framing: PRD #1.)

## Problem & evidence

A Shopify app partner is paid by Shopify after Shopify deducts a **revenue-share cut** and a **processing fee** — and partners have little independent way to check they were paid the *right* amount, or to see *when* money will land. This is LedgerGuard's namesake wedge (**🛡️ Guard**): *did Shopify pay you correctly, and does the money add up?*

- **INFERRED (`docs/REPORTS.md §2`):** reconciliation is explicitly positioned as the differentiator Mantle can't match.
- **UNKNOWN — original demand evidence not recorded.** Open question: validate demand (owner: Product).

## Target users

**INFERRED:** the **Shopify app partner** (founder/finance) who wants confidence in Shopify's math and visibility into payout timing. Per-app, per-org.
- **Not the target:** merchants; partners needing a full accounting system (this reconciles Shopify's numbers, it isn't a GL — see Non-goals).

## Proposed solution (as-built behavior)

**FACT — five surfaces** (all `/api/v1/apps/{appID}/reports/*`, real screens in `frontend-flutter/lib/screens/reports/`):
- **Fee Audit** (`fee_audit_screen.dart`): per-month configured tier % vs the **detected** rate derived from observed `shopifyFee/gross`, **snapped per-month to the nearest known tier** (0/15/20%) so a mid-year $1M crossing doesn't create a false blended-rate mismatch (`fee_shared.go:84-91`). Flags a month when `|actual − expected| > 1% of gross`; shows **savings vs the 20% default**.
- **Payout Schedule** (`payout_schedule_screen.dart`): upcoming **PENDING + AVAILABLE** net earnings grouped by available-date, with `nextPayoutDate`.
- **Payout History** (`payout_history_screen.dart`): **PAID_OUT** net earnings by calendar month (by Shopify charge date).
- **Earnings** (`earnings_report_screen.dart`): net take-home split into **pending / available / paid-out**, plus a per-charge table.
- **Ledger Reconciliation** (`ledger_recon_screen.dart`): verifies the identity **`gross = net + revenue_share + processing`** per month, absorbing 1% rounding, and flags `processing_suspect` when the derived processing % exceeds 6% (a sign Shopify's share didn't sync).

**FACT — the math.** Fee verification (`fee_verification_service.go:46-90`) compares actual fees to `tier × gross` (tax excluded as variable); "verified" iff revenue-share **and** processing variance are within tolerance. Net = `gross − shopify_share − processing − tax_on_fees` (`transaction.go:117-120`). Availability = **charge date + 7 days** (refunds immediate), status PENDING/AVAILABLE by comparing to now (`earnings_calculator.go:44-67`). All deterministic.

### Key screens

**No new wireframe** (as-built). Real surfaces: the five report screens above.

## Platform & policy constraints

**FACT.** Everything is derived from synced Shopify transaction fee fields; accuracy is bounded by what Shopify's Partner API returns (if the revenue-share cut doesn't sync, reconciliation uses the `processing_suspect` heuristic to avoid a false pass). **Key assumptions are hardcoded** (see below): tier rates 0/15/20/20-default, processing **2.9%**, payout delay **7 days**. Real Shopify payout delays are 7–37 days — the 7-day figure is an MVP simplification.

## Pricing-tier impact

**UNKNOWN — no plan-gating found**, though `docs/REPORTS.md` positions Guard/reconciliation as the premium wedge. Open question: gate reconciliation behind a paid tier? — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Reconciliation is read-only over synced transactions; recomputed each request. A rate/assumption change reflects immediately. No migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Reconciliation trust:** ledger-reconciliation `reconciled=true` for ≥ 95% of app-months with complete synced fee data (a lower rate means either real Shopify discrepancies *or* our heuristics misfiring — both worth knowing).
2. **Discrepancy value:** among flagged months, ≥ 1 in 5 reflects a *real* fee/payout discrepancy on manual audit (validates the wedge; guards against false positives).
3. **Payout-timing accuracy:** predicted `nextPayoutDate` lands within ±X days of actual disbursement ≥ 80% of the time (drives replacing the hardcoded 7-day delay).

## Non-goals (deliberately absent in code)

1. **Not a general ledger / accounting system** — it reconciles Shopify's numbers, doesn't do bookkeeping. **INFERRED.**
2. **No tax reconciliation** — fee verification passes `tax=0` (tax treated as variable, `fee_verification_service.go:53`). **FACT.**
3. **No admin UI to edit tier/processing rates** — hardcoded, code-redeploy to change. **FACT.**
4. **No actual disbursement dates** — payout dates are *estimates* (charge + 7d / max available-date), not Shopify's real payout timestamps. **FACT.**

## Open questions

- **Hardcoded assumptions** (tier 0/15/20%, processing 2.9%, 7-day delay) — Shopify can change these; no config path. Move to config/data? Owner: Eng/Product.
- **`processing_suspect` 6% ceiling** is a heuristic — a genuinely high-fee merchant could be flagged as suspect (false positive). Calibrate/validate? Owner: Eng.
- **Refund handling:** refunds assumed negative net + immediate availability — if sync ever sends them differently, earnings buckets break. Owner: Eng.
- **Multi-currency:** currency defaults to USD with no mixed-currency detection — could a mixed-currency app misreport? Owner: Eng.
- **Real payout delay is 7–37 days** — should availability model Shopify's actual schedule? (metric #3). Owner: Eng/Product.
- Pricing/tier gating for Guard. Owner: Product.

# Tech Design: Fees, Payouts & Reconciliation (as-built)

**PRD:** docs/prds/07-fees-and-payouts.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API contracts (5 report endpoints) + the hardcoded fee assumptions are the hard-to-undo surfaces — a wrong tier/processing/delay constant silently changes every reconciliation verdict. No new data model (reads synced transaction fee fields). The *reconciliation verdict* is the highest-stakes output: false positives erode trust in the "Guard" wedge; false negatives defeat its purpose.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — fee verification** (`fee_verification_service.go:46-90`): `VerifyTransaction` computes expected fees as `tier.CalculateFeeBreakdown(gross, taxRate=0)` (tax excluded as variable, `:53`), compares to actual synced `ShopifyFeeCents`/`ProcessingFeeCents`, and sets `IsVerified` iff **both** revenue-share and processing variances are `≤ gross × tolerancePercent` (`:83-87`). `CalculateTierSavings` baselines against the 20% default (`:144-162`). Tiers + rates are an enum (`revenue_share_tier.go`): DEFAULT 20% / SMALL_DEV_0 0% / SMALL_DEV_15 & LARGE_DEV 15%; **processing always 2.9%** (`:29`); fee math truncates via `int64(float64…)` (`:97,102`). `ParseRevenueShareTier` defaults **unknown → SMALL_DEV_0 (0%)** (`:144`).

**FACT — fee audit report** (`fee_audit_report_handler.go` + `fee_shared.go`): instead of trusting a configured tier, `buildFeeAudit` **detects** the rate from observed `shopifyFee/gross` and **snaps per-month** to the nearest known tier (0/15/20), robust to a mid-year $1M crossing (`fee_shared.go:84-91`); flags a month when `|actual − expected| > 1% of gross` (`:96`); shows savings vs 20% default.

**FACT — earnings/payouts** (`earnings_calculator.go`): net = `gross − shopify_share − processing − tax_on_fees` (`transaction.go:117-120`); availability = charge date **+ 7 days** (refunds immediate, `:44-57`); status PENDING/AVAILABLE by comparing now to available-date; `SummarizeEarnings` sums `NetAmountCents` per status bucket, ignoring unknown statuses. Payout Schedule = PENDING+AVAILABLE grouped by available-date; Payout History = PAID_OUT grouped by charge-month; per-period date is `MAX(availableDate)` — an **estimate**, not actual disbursement.

**FACT — ledger reconciliation** (`ledger_reconciliation_report_handler.go`): per-month checks `gross = net + revenue_share + processing`, tolerating `|residual| ≤ gross/100`; flags `processing_suspect` when derived processing % > 6.0% (`:45-51`); `reconciled = bucketsClose AND !suspect` (or gross ≤ 0).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **deterministic, read-only reconciliation over synced transaction fee fields**, with data-derived tier detection (report path) and identity-checking (`gross = net + share + processing`). Read paths: see `docs/designs/07-fees-and-payouts-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — no new tables.** Reads `transactions` fee fields (`GrossAmountCents`, `ShopifyFeeCents`, `ProcessingFeeCents`, `TaxOnFeesCents`, `NetAmountCents`, charge type/date) populated during sync. Tier rates + processing % + delay are **compile-time constants**, not config/data. No migration.

### API & events

**FACT.** Five endpoints (`router.go:316-365`), all Firebase+org auth: `/reports/fee-audit?months=1..24`, `/reports/payout-schedule?from&to&limit&offset`, `/reports/payout-history?from&to&limit&offset`, `/reports/earnings?from&to&limit&offset`, `/reports/ledger-reconciliation?months=1..24`. Response shapes per PRD/inventory. No events.
**Breaking-change check:** `*_cents` fields + the `reconciled`/`processing_suspect`/`fee_guard_ok`/`tier_matches` flags are the contract the Flutter screens bind to.

## Alternatives considered

- **Data-derived tier detection (report) vs configured-tier verification (`VerifyTransaction`).** Both exist. The report path detects the tier from observed fees (robust, no config needed); `VerifyTransaction` trusts a passed tier and **defaults unknown to 0%** — divergent behavior (D5). No recorded rationale for keeping both. **Prefer detection** unless a per-transaction verifier against a *known-correct* tier is specifically needed.
- **Hardcoded constants vs config/data-driven rates.** As-built hardcodes tier/processing/delay. **Choose config if** Shopify changes any rate (currently a code redeploy). No alternative recorded.
- **`processing_suspect > 6%` heuristic vs matching refund fee-reversals explicitly.** The heuristic catches unsynced revenue-share cheaply but can mis-flag genuinely high-fee months. **Choose explicit refund-GID matching if** false positives prove common.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Fee variance within tolerance | `≤ gross×tol` | **FACT** IsVerified true / FeeGuardOk true (`fee_verification_service.go:86`, `fee_shared.go:96`) | — |
| Tier crossing ($1M mid-year) | per-month snap | **FACT** snapped per-month, no false blended mismatch (`fee_shared.go:84-91`) | — |
| Unsynced revenue-share cut | processing% > 6% | **FACT** `processing_suspect`, month NOT reconciled despite residual≈0 (`:45-51,128-132`) | re-sync |
| gross ≤ 0 (refund-only month) | guard | **FACT** treated reconciled / FeeGuardOk (avoids div-by-zero) | — |
| Refund in earnings | negative net | **FACT** subtracts naturally, available immediately (`earnings_calculator_test.go:239-253`) | — |
| Rounding drift | 1% tolerance | **FACT** absorbed (`:136`, `fee_shared.go:96`) | — |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. Hardcoded assumptions** — tier 0/15/20 + default 20, processing **2.9%**, payout delay **7 days** (real Shopify 7–37d). Shopify rate changes require a code redeploy; wrong constants silently skew every verdict. → open question (move to config).
- **D2. `processing_suspect` 6% ceiling is a heuristic** — a genuinely high-fee month is flagged suspect (false positive). → proposed calibration test (PRD metric #2).
- **D3. Payout dates are estimates** (charge+7d / max-available), not Shopify's real disbursement timestamps — schedule/history can mislead. → proposed accuracy test (metric #3).
- **D4. Systemic tenant isolation** — same `resolveAppFromRequest` (no org-ownership check). → cross-ref `02-ai-chat.md` D1.
- **D5. `VerifyTransaction` unknown-tier → 0%** (`revenue_share_tier.go:144`) would expect zero revenue-share and flag any real cut as a discrepancy; only the report path (data-detected) avoids this. → open question (align the two paths).
- **D6. No tax reconciliation** (`tax=0`) — tax-on-fees variance is hidden by tolerance, never reconciled. → open question.
- **D7. No multi-currency detection** — currency defaults to USD; a mixed-currency app could sum disparate currencies as one. → proposed test.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Reconciliation trust: on fully-itemized fixtures, `reconciled=true` for balanced months; residual/suspect flagged when the share is unsynced (metric #1) — **exists** (`ledger_reconciliation_report_handler_test.go`).
2. Discrepancy validity + payout-timing accuracy (metrics #2/#3) — **TODO**, require real-account data.

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- VerifyTransaction at 20/0/15%, tier savings — `fee_verification_service_test.go`. Availability/refund-immediate, status transitions, unknown-status ignored, negative-refund subtract — `earnings_calculator_test.go`. Tier mismatch + off-tier flagged + $1M snap — `fee_audit_report_handler_test.go`. Balanced vs residual vs processing_suspect — `ledger_reconciliation_report_handler_test.go`.

**Adversarial — DIVERGENCE TODO (none exist today):**
- D2 high-fee false-positive · D3 payout-date accuracy · D4 cross-org appID · D5 unknown-tier verifier default · D7 mixed-currency. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/domain/service/ -run 'Fee|Earnings' -v && go test ./internal/interfaces/http/handler/ -run 'FeeAudit|Ledger|Payout|Earnings' -v`
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/reports/ledger-reconciliation?months=6" | jq '.reconciled,.months[].processing_suspect'`
3. Fee audit: `curl ... /reports/fee-audit?months=12 | jq '.detected_fee_pct,.tier_matches,.flagged_months'` and sanity-check vs the app's real Shopify tier.

**External-platform reality:** all math is local over synced fee fields — no Shopify call at report time; fully deterministic/unit-testable. But the *correctness* of the hardcoded constants (tier %, 2.9%, 7-day delay) can only be validated against a real Partner account's actual payouts — never in unit tests.

## Pricing & policy touchpoints

**Platform:** encodes Shopify's Reduced Revenue Share Plan rules as constants (`revenue_share_tier.go:4-25`) — these are Shopify policy that can change. **UNKNOWN** whether Guard is tier-gated (no code); `docs/REPORTS.md` positions it as the premium wedge. → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Read-only + deterministic; a constant/logic change reflects on the next report load. Rollback = revert. No migration. Because constants are compile-time, a Shopify rate change is a redeploy, not a config toggle (D1).

## Observability

**FACT (partial):** unset-charge-date and unknown-status rows are logged (`payout_history_report_handler.go:180-181`, earnings filtering). **DIVERGENCE:** no metric/alert on flagged-month rate, `processing_suspect` frequency, or reconciliation pass-rate trends — the very signals that would validate the wedge. On-call greps Hetzner logs. → proposed observability TODO tied to metrics #1/#2.

## Open questions

- **Move tier/processing/delay to config or derive from data** (D1). Owner: Eng/Product.
- Calibrate/replace the 6% `processing_suspect` heuristic (D2). Owner: Eng.
- Model real Shopify payout schedule vs the 7-day estimate (D3). Owner: Eng/Product.
- Align `VerifyTransaction` unknown-tier default with the report's data-detection (D5). Owner: Eng.
- Tax reconciliation (D6) + multi-currency detection (D7). Owner: Eng.
- Systemic org-ownership check (D4). Owner: Eng/Sec.
- Guard tier-gating. Owner: Product.

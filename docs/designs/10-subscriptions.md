# Tech Design: Subscriptions Browsing & ARPU/LTV (as-built)

**PRD:** docs/prds/10-subscriptions.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API shapes (list/detail/history/risk-timeline/report) bind the Flutter client. No new data model (reads rebuilt subscription state + transactions + snapshots). Low irreversibility overall; the ARPU/LTV *formulas* are the main correctness surface (partners may make decisions on LTV).

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — list** (`subscription.go:52-197`): `FindWithFilters` (`subscription_repository.go:268-411`) builds a dynamic `WHERE` (risk state, subscription status, plan, price range, billing interval) with **DB-side** `LOWER(...) LIKE` search on shop name/domain and `LIMIT/OFFSET` pagination (25/page, max 100); all queries exclude `deleted_at IS NULL`. Shop enrichment (logos/names) is a **batch** `FindByDomains` (N+1 mitigated). `GetSummary`/`GetPriceStats` are single aggregate queries. **No in-memory filtering or slicing.**

**FACT — detail** (`subscription.go:354-434`): `GetDetail` + `GetPaymentHistory` (last ~12 months) + `GetRiskTimeline` (state-change events from `subscription_event`). *(This is where the risk-transition audit trail surfaces — partially resolving PRD #5's D6.)*

**FACT — derived fields** (`subscription.go`): `MRRCents()` returns `BasePriceCents/12` for annual (integer **floor**, `:206-211`), else base price. `StartDate()` = `ActivatedAt` if set, else `CreatedAt` (business start, not record-created — `:38-43`, per the createdat-is-record-date convention).

**FACT — ARPU/LTV** (`subscriptions_report_handler.go`): `arpuCents = Σ MRR(SAFE) / count(SAFE)` floored (`:267-272`); `ltvCents = round(ARPU / monthlyChurn)`, **0 when churn ≤ 0** (undefined → frontend shows "—", documented `:274-284`); churn = churned ÷ latest-snapshot total. Per-plan breakdown reuses the **app-level** churn.

## Proposed design

**No redesign — documents the shipped design.** Pattern: **thin handlers over DB-side filtered/paginated repository queries**, with a derived-field entity and a snapshot-backed ARPU/LTV computation. Read paths: see `docs/designs/10-subscriptions-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — no new tables.** Reads `subscriptions` (rebuilt each sync, PRD #6), `transactions` (payment history), `subscription_event` (risk timeline), `daily_metrics_snapshot` (churn denominator). No migration.

### API & events

**FACT.** `GET /apps/{appID}/subscriptions` (page/pageSize/limit/offset, status=comma-sep risk, subscription_status, plan, priceMin/Max, billingInterval, search, sortBy, sortOrder) → `{subscriptions,total,page,pageSize,totalPages}`; `/subscriptions/{subscriptionID}`; `/subscriptions/summary`; `/subscriptions/price-stats`; plus user-scoped `/subscriptions/{id}`, `/{id}/history`, `/{id}/risk-timeline`; and `/apps/{appID}/reports/subscriptions[?format=csv]`. No events.
**Breaking-change check:** `*_cents` fields + pagination envelope bind the Flutter models.

## Alternatives considered

- **DB-side filtering/pagination vs in-memory.** As-built pushes everything to Postgres (`FindWithFilters`) — the right call for large tenants, and a deliberate contrast with the report platform's in-memory `pageSlice` (PRD #1). No alternative recorded; this is simply the better-built path.
- **App-level churn vs per-plan churn for LTV.** As-built uses app-level churn for every plan (`:144`) because per-plan churn isn't computed. **Choose per-plan churn if** plans have materially different retention (metric #3).
- **`ActivatedAt`-else-`CreatedAt` for StartDate.** A recorded, deliberate choice (comment `:36-37`) to avoid `CreatedAt` (which resets on rebuild). No alternative contested.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| No active subs (ARPU) | `activeSubs<=0` | **FACT** ARPU 0 (`:268-270`) | — |
| Churn = 0 (LTV) | `churnRate<=0` | **FACT** LTV 0 → UI "—" (`:280-282`) | — |
| Stale snapshot vs live churned | count > snapshot total | **FACT** warned/logged (`:191-193`) | next sync |
| Empty/over-large page | limit clamp (max 100) | **FACT** clamped, DB paginates | — |
| Deleted subscriptions | `deleted_at IS NULL` | **FACT** excluded everywhere (`:279`) | — |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. Per-plan LTV uses app-level churn** — inflates high-churn plans' LTV. → proposed per-plan-churn recompute (metric #3).
- **D2. Annual-MRR round vs floor mismatch** — backend floors `price/12`, frontend rounds → 1¢ cosmetic discrepancy. → align rounding.
- **D3. Churn-window mismatch** — ARPU/LTV churn uses the latest snapshot in a 90-day window; the Churn report (PRD #5) uses 30 days → the same app can show two churn rates. → unify the window.
- **D4. Currency collapse** — the report picks the first non-empty currency; a mixed-currency app is misrepresented as single-currency. → detect/segregate.
- **D5. Systemic tenant isolation** — the `/apps/{appID}/...` routes use `resolveAppFromRequest` (no org-ownership check). *Note:* the user-scoped `/subscriptions/{id}` routes should be checked separately for owner scoping. → cross-ref `02-ai-chat.md` D1.
- **D6. Search special chars unescaped** — `LIKE` pattern chars (`%`, `_`) in a shop name aren't escaped → rare false-positive matches. → escape input.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Cross-surface consistency: list summary counts == Dashboard + Risk for the same sync (metric #2). **TODO.**
2. LTV credibility: per-plan LTV vs a per-plan-churn recompute within ±X% (metric #3). **TODO.**

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- ARPU floor (not round), LTV round (not truncate), churn stale-snapshot detection, per-plan sort — **exist** (`subscriptions_report_handler_test.go`). List filters/pagination, GetByID, Summary, PriceStats — **exist** (`subscription_test.go`). `StartDate` fallback — **exists** (`subscription_start_test.go`).

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 per-plan churn · D2 round/floor parity (backend↔frontend) · D3 window unification · D4 mixed-currency · D5 cross-org appID · D6 LIKE-escaping. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/interfaces/http/handler/ -run 'Subscription' -v && go test ./internal/domain/entity/ -run 'Start|MRR' -v`
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/subscriptions?subscription_status=ACTIVE&search=acme&sortBy=base_price_cents&sortOrder=desc&pageSize=25" | jq '.total,.subscriptions[0]'`
3. `curl ... /reports/subscriptions | jq '.arpuCents,.ltvCents,.churnRate'`; confirm LTV=0 renders as "—" in the app when churn is 0.
4. Compare summary counts to `/metrics` + `/risk/summary` (metric #2).

**External-platform reality:** all queries are local over rebuilt state — no Shopify/LLM call; fully deterministic/unit-testable. Only upstream freshness depends on sync (faked below the repo in tests).

## Pricing & policy touchpoints

**UNKNOWN — no plan-gating** for browsing or the ARPU/LTV report. No Shopify-policy interaction. → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Read-only over rebuilt state; no migration. Rollback = revert. Formula changes (e.g. per-plan churn) ship behind the stable JSON contract.

## Observability

**FACT (partial):** stale-snapshot warnings logged (`:191-193`). **DIVERGENCE:** no metric on query latency, filter usage, or ARPU/LTV distribution. On-call greps Hetzner logs. → proposed observability TODO.

## Open questions

- Per-plan churn for LTV (D1). Owner: Eng/Product.
- Annual-MRR round/floor parity (D2). Owner: Eng.
- Unify churn window with the Churn report (D3). Owner: Eng.
- Mixed-currency handling (D4). Owner: Eng/Product.
- Escape `LIKE` special chars in search (D6). Owner: Eng.
- Systemic org-ownership check (D5). Owner: Eng/Sec.
- Pricing/tier gating. Owner: Product.

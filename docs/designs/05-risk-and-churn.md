# Tech Design: Risk & Churn (as-built)

**PRD:** docs/prds/05-risk-and-churn.md · **Author:** sachin.s · **Date:** 2026-09-25
**Status:** Draft (backfilled from implementation)
**Reversibility:** The 30/60/90-day thresholds + status→state mapping are **product policy encoded in a pure function** — cheap to change, but the persisted `risk_state` column and the snapshot counts derived from it are what Dashboard/Reports read, so a policy change silently reclassifies everything on the next sync. Public API shapes (`/risk/summary`, `/reports/{churn,retention,revenue-at-risk}`) are the hard-to-undo surface.

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — the engine** (`risk_engine.go:24-88`) is a stateless, pure, deterministic classifier. `ClassifyRisk(sub, now)` order: terminal (CANCELLED/EXPIRED) → CHURNED (`:28-30`); FROZEN → ONE_CYCLE_MISSED (`:33-35`); PENDING → SAFE (`:38-40`); ACTIVE **and** non-nil future/today charge → SAFE (`:44-48`); nil charge → SAFE (`:51-52`); else `DaysPastDue` (`:62-72`, `int(hours/24)`, 0 if future/nil) → `RiskStateFromDaysPastDue` (`:75-88`, ≤30 SAFE / ≤60 1-cycle / ≤90 2-cycle / >90 CHURNED). `ClassifyAll` (`:91-95`) mutates in place; `CalculateRiskSummary`/`CalculateRevenueAtRisk` (`:97-128`) aggregate.

**FACT — lifecycle.** The engine runs **once per sync** inside ledger rebuild; results are **persisted** to `subscriptions.risk_state` (UpsertBatch). Read handlers **never re-run the engine** — `risk_handler.go:55` (RISK-1b) explicitly reads persisted state so Dashboard/Risk/Subscriptions converge. `SnapshotProcessor` then derives `daily_metrics_snapshot` counts + `RevenueAtRiskCents` + `RenewalSuccessRate` from the persisted states, refreshing today's snapshot post-sync (RISK-1).

**FACT — read handlers.** Risk Summary (`risk_handler.go:39`) = distribution + ranked at-risk stores. Churn (`churn_handler.go:75`) = **live** count of CHURNED subs + Σ MRR lost + rate (÷ snapshot TotalSubscriptions, clamped 0–1) + snapshot trend. Retention (`retention_handler.go:73`) = renewal rate from **snapshot** headline, retained MRR (Σ SAFE MRR), reactivations (distinct shops with `REACTIVAT*` events in range), per-plan renewal. Revenue at Risk (`revenue_at_risk_handler.go`) = at-risk MRR + recoverable (`1c*0.60 + 2c*0.25`, `:23-26`).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **classify-once-at-sync, persist, read-only aggregation** (write/read split prevents read-time drift). Lifecycle + read paths: see `docs/designs/05-risk-and-churn-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — no new tables.** Reads/writes existing `subscriptions.risk_state` (enum column), `subscription_event` (records `from_risk_state`/`to_risk_state` transitions — an audit trail, not surfaced), and `daily_metrics_snapshot` (derived counts). No migration; a threshold change is code-only and takes effect on next sync.

### API & events

**FACT.** `/risk/summary` → `{distribution:{safe,one_cycle,two_cycle,churned}, at_risk_stores:[...]}`. `/reports/churn` → `{currency, churnRate, churnedMrrLostCents, churnedCount, trend[], stores[], storesTotal}`. `/reports/retention` → `{currency, renewalRate, retainedMrrCents, reactivations, trend[], plans[]}`. `/reports/revenue-at-risk` → at-risk + recoverable (PRD #1). All Firebase+org auth, `?from&to[&segment][&format=csv]`.
**Breaking-change check:** the `risk_state` enum values (`SAFE/ONE_CYCLE_MISSED/TWO_CYCLES_MISSED/CHURNED`) are a contract shared by DB, API, and Flutter — renaming breaks all three.

## Alternatives considered

- **Classify-at-sync-and-persist vs classify-on-read.** As-built persists (RISK-1b, `risk_handler.go:55`) explicitly to keep surfaces convergent and to survive the "cancel-trap" reconciliation. On-read classification was **tried/considered and rejected** (the RISK-1b comment is the recorded rationale). **Choose on-read only if** near-real-time reclassification between syncs becomes required (accepting cross-surface drift).
- **Rule-based thresholds vs ML churn score.** As-built is rule-based (deterministic, explainable — a positioning choice per `docs/REPORTS.md`). **Choose ML if** explainability stops mattering and labeled churn data exists. No ML alternative recorded beyond the positioning note.
- **Recoverable-revenue constants (60%/25%) vs learned rates.** `docs/REPORTS.md` notes intent to learn from reactivation events; shipped code hardcodes. **Choose learned rates if** metric #3 shows the constants are off.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| Terminal status | `IsTerminal()` | **FACT** → CHURNED regardless of timing (`:28-30`) | — |
| FROZEN (payment fail) | `IsFrozen()` | **FACT** → ONE_CYCLE_MISSED short-circuit (`:33-35`) | — |
| nil `ExpectedNextChargeDate` | nil guard | **FACT** → SAFE (defensive, `:51-52`) | — |
| Future charge date | `hours<0` | **FACT** daysPastDue=0 → SAFE (`:68-70`) | — |
| Exact day boundaries (30/60/90) | `<=` bounds | **FACT** tested: 30→SAFE, 31→1c, 60→1c, 61→2c, 90→2c, 91→CHURNED (`risk_engine_test.go:195-307`) | — |
| Churn count vs snapshot drift | count > total | **FACT** logged, still returned (`churn_handler.go:229-231`) | reconciles next sync |

### DIVERGENCE — paths NOT handled (LISTED, not fixed)

- **D1. PAUSED (and any non-enumerated status) falls through to ACTIVE days-logic.** Non-terminal/frozen/pending/active subs skip the `:44` future-charge short-circuit and get classified purely by days-past-due — a PAUSED sub with an old charge date is reported CHURNED. Intended? → open question + proposed test.
- **D2. UTC-only days math** (`:67`, `int(hours/24)`) — no DST/timezone; a sub within ~1 day of a boundary can flip state depending on the assumed zone. → proposed test.
- **D3. Recovery constants unvalidated** (0.60/0.25) — no learning loop despite `docs/REPORTS.md` intent. → proposed calibration (PRD metric #3).
- **D4. Churn (live) vs retention (snapshot) source split** — can momentarily disagree; only logged, not reconciled at read. → open question.
- **D5. Systemic tenant isolation** — same `resolveAppFromRequest` (no org-ownership check) as reports/chat/forecast/dashboard. → cross-ref `02-ai-chat.md` D1.
- **D6. Risk-transition audit not surfaced** — `subscription_event` stores from/to risk state but there's no timeline UI/endpoint; data accrues unused. → open question (Product).

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Cross-surface consistency: risk distribution equal across Dashboard `/metrics`, `/risk/summary`, and Subscriptions for the same sync (metric #1). **TODO.**
2. Recovery calibration harness: compare actual reactivation rates to 0.60/0.25 (metric #3). **TODO.**

**Adversarial (one per Failure-modes row — mostly EXISTING):**
- Terminal/FROZEN/PENDING/nil/future + all 30/60/90 boundaries — **exist** (`risk_engine_test.go`, 31 tests). Churn count+MRR incl. annual÷12 — **exists** (`churn_handler_test.go`). Retention per-plan + reactivations — **exists** (`retention_handler_test.go`).

**Adversarial — DIVERGENCE TODO (none exist today):**
- D1 PAUSED classification · D2 timezone/DST near boundary · D4 churn-vs-retention source agreement · D5 cross-org appID. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/domain/service/ -run Risk -v` (expect all boundary tests green)
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/risk/summary" | jq '.distribution'`
3. Compare distribution to `/metrics` risk counts and to `/reports/churn` churnedCount for the same app (D4/metric #1).
4. Construct a PAUSED sub with an old charge date in a test DB → observe reported state (D1).

**External-platform reality:** classification is pure over synced fields — no Shopify/LLM call at classify or read time; fully deterministic and unit-testable. Only upstream `status`/`expectedNextChargeDate` freshness depends on sync (faked below the repo in tests).

## Pricing & policy touchpoints

**Policy encoded, not platform.** The 30/60/90 grace/cycle thresholds are LedgerGuard product policy in `risk_engine.go` (CLAUDE.md §13). **UNKNOWN** whether Revenue-at-Risk/recoverable is tier-gated (no code). → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Deterministic + idempotent: any threshold/status-mapping change ships as code and reclassifies on the next sync — no migration, no backfill script (snapshots rebuild). Rollback = revert the engine; next sync restores prior states.

## Observability

**FACT (partial):** churn-vs-total drift is logged (`churn_handler.go:229-231`); risk transitions recorded to `subscription_event`. **DIVERGENCE:** no metric/alert on distribution shifts, no dashboard for reclassification volume per sync, no surfaced transition timeline. On-call greps Hetzner logs. → proposed observability TODO tied to metric #1.

## Open questions

- **PAUSED / non-enumerated status handling (D1).** Owner: Product/Eng.
- Calibrate recovery constants from reactivations (D3). Owner: Eng/Product.
- Unify churn(live) vs retention(snapshot) source (D4). Owner: Eng.
- **Update CLAUDE.md §13 pseudo-code** to match the real (more defensive) engine. Owner: Eng.
- Systemic org-ownership check (D5). Owner: Eng/Sec.
- Surface risk-transition history (D6)? Owner: Product.

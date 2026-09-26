# Tech Design: Stores (as-built)

**PRD:** docs/prds/13-stores.md · **Author:** sachin.s · **Date:** 2026-09-26
**Status:** Draft (backfilled from implementation)
**Reversibility:** Public API shapes (store list/detail) bind the Flutter client. No new data model beyond the existing `shops` enrichment cache. Low irreversibility; the correctness risks (persisted-state reuse, event-sourced dates) are already regression-guarded (STORE-1/STORE-2).

> **As-built.** "Current state IS the design." Every claim cites a real path:line. Markers: **FACT / INFERRED / UNKNOWN**. **Nothing is refactored or fixed** — gaps are LISTED as DIVERGENCE.

## Current state

**FACT — store list** (`store_handler.go:42`, `:119-195`): `GET /apps/{appID}/stores?page&pageSize&search` loads subscriptions (using **persisted** `risk_state` — STORE-1, `:57-61`) + app events, **deduplicates by domain** into store cards with a priority risk state, `health_score` (`healthScoreFromRisk`, `:213-226`: SAFE 90 / 1c 50 / 2c 25 / churned 10, default 50), 3-year `lifetime_value_cents` from transactions, and **event-sourced** `first_install_date`/`last_interaction` (`store_dates.go`: event → subscription start → `UpdatedAt` only as last resort so a rebuild timestamp never wins — STORE-2).

**FACT — store detail** (`store_health.go:91`): `GET /apps/{appID}/stores/{domain}/health` returns the subscription(s) + 3-month transactions + earnings, optionally merged with shop brand (logo/name) when a `ShopRepository` is attached (`SetShopRepo`, `:38-41`; nil-tolerant). The frontend keys detail state by **domain** with a request guard to avoid A→B navigation races (SD-1/SD-2, `store_provider.dart`).

**FACT — enrichment** (`shopify_storefront_client.FetchBrand`, `shop_repository.go`): logo/name/branding fetched from the Storefront API and upserted into `shops`, **opportunistically during sync** (no scheduled refresh).

## Proposed design

**No redesign — documents the shipped design.** Pattern: **domain-level aggregation over persisted subscription/event/transaction data + opportunistic shop-brand enrichment**, read-only. Read paths: see `docs/designs/13-stores-sequence.puml` (validated `plantuml -checkonly`).

### Data model

**FACT — reads existing tables** (`subscriptions`, `app_events`, `transactions`) + the `shops` enrichment cache (`entity/shop.go`: domain, GID, name, LogoURL/SquareLogoURL/CoverImageURL, PrimaryDomain, CountryCode, CurrencyCode). No new migration.

### API & events

**FACT.** `GET /apps/{appID}/stores` → `{stores:[{id,shop_domain,installed_app_ids,health_score,lifetime_value_cents,first_install_date,last_interaction,risk_state}],total,page,pageSize,totalPages}`. `GET /apps/{appID}/stores/{domain}/health` → `{subscription,transactions,earnings, shop?}`. No dedicated store-by-domain entity endpoint — detail is subscription-centric. No events.
**Breaking-change check:** the store card fields bind `StoreModel`/`StoreProvider`.

## Alternatives considered

- **Domain-aggregated store view vs reusing the subscriptions list.** As-built builds a real per-domain roll-up (priority risk, LTV, event dates) — genuinely distinct from #10. **Reuse subscriptions only if** one-store-one-subscription always held (it doesn't — multiple subs per store). Deliberate separate surface.
- **Persisted risk vs re-classify on read.** As-built reads persisted state (STORE-1) so cancel-trap stores don't re-churn — a recorded decision tied to RISK-1b (#5). **Re-classify only if** near-real-time between-sync reclassification is needed.
- **Risk-derived health score vs composite model.** As-built maps risk→{90,50,25,10}. **Choose a composite** (tenure, payment history, usage) if a single risk-derived number proves too coarse. No composite alternative recorded.

## Failure modes

| Failure | Detection | Behavior (as-built) | Recovery |
|---------|-----------|---------------------|----------|
| No shop-brand enrichment | ShopRepo nil / miss | **FACT** renders with raw domain, no logo (nil-tolerant, `store_health.go:38-41`) | enrichment on next sync |
| Cancel-trap store (cancelled-then-active) | persisted state | **FACT** stays SAFE, not re-churned (STORE-1, tested `store_handler_test.go:90`) | — |
| Only rebuild timestamps available | date resolution | **FACT** `UpdatedAt` used only as last resort (STORE-2, tested `:18`) | — |
| Refund-negative LTV | negative handling | **FACT** frontend formats negatives (RISK-3, `store_model.dart`) | — |
| Store A→B nav race | domain-keyed guard | **FACT** stale response discarded (SD-1/SD-2, `store_provider.dart`) | — |

### DIVERGENCE — paths NOT handled well (LISTED, not fixed)

- **D1. Health score is a 4-value lookup**, not a composite — two very different SAFE stores get identical 90. → open question (richer model?).
- **D2. Logo/brand enrichment is opportunistic with no refresh cadence/TTL** — a rebranded store shows a stale logo indefinitely. → proposed refresh job.
- **D3. LTV is a fixed 3-year window** — arbitrary horizon; interaction with refund-negative stores under-explored. → open question.
- **D4. Systemic tenant isolation** — `resolveAppFromRequest`, no org-ownership check. → cross-ref `02-ai-chat.md` D1.
- **D5. Detail is subscription-centric** — there's no store entity endpoint independent of subscriptions; a store with only events (no subscription) may render thin. → verify empty-subscription store detail.

## Test plan

**Acceptance (from PRD PROPOSED metrics, as observable checks):**
1. Cross-surface consistency: store risk badge == that store's subscription risk == Dashboard/Risk counts for the same sync (metric #1). **TODO** (extend store tests).
2. Enrichment coverage: ≥90% of active stores resolve a real name/logo (metric #2). **TODO** (needs enrichment metric).

**Adversarial (one per Failure-modes row — key ones EXISTING):**
- Event-sourced dates (not CreatedAt/UpdatedAt) — **exists** (`store_handler_test.go:18`, STORE-2). Persisted risk (cancel-trap SAFE) — **exists** (`:90`, STORE-1). Date-resolution helpers — `store_dates_test.go`.

**Adversarial — DIVERGENCE TODO (none exist today):**
- D2 stale-logo refresh · D3 3-year LTV + refund-negative store · D4 cross-org appID · D5 subscription-less store detail. **TODO.**

**Manual verification (copy-paste):**
1. `cd backend && go test ./internal/interfaces/http/handler/ -run Store -v`
2. `curl -H "Authorization: Bearer <idToken>" -H "X-Org-Id: <org>" "http://localhost:8080/api/v1/apps/<appID>/stores?pageSize=25&search=acme" | jq '.stores[0]'` — confirm health_score matches the risk_state map.
3. `curl ... /stores/<domain>/health | jq '.subscription.risk_state, (.shop.LogoURL // "no-logo")'`
4. Compare a store's risk badge to `/subscriptions` risk + `/risk/summary` (metric #1).

**External-platform reality:** store data is local over synced state; the Storefront API (brand enrichment) is the only external dependency, faked below `shopify_storefront_client` in tests. Real logo/name resolution must smoke-test against real stores.

## Pricing & policy touchpoints

**UNKNOWN — no plan-gating.** Storefront API brand fetch subject to Shopify API limits. → Open question (Product).

## Rollout

**N/A (as-built, shipped).** Read-only aggregation; no migration. Rollback = revert. A logo-refresh job (D2) would be additive.

## Observability

**FACT (partial):** enrichment/handler errors logged. **DIVERGENCE:** no metric on enrichment coverage, store-list latency, or health-score distribution. On-call greps Hetzner logs. → proposed observability TODO tied to metric #2.

## Open questions

- **Composite health score (D1)?** Owner: Product.
- **Logo/brand refresh cadence (D2).** Owner: Eng.
- **LTV horizon + refund-negative stores (D3).** Owner: Eng/Product.
- **Subscription-less store detail (D5).** Owner: Eng.
- **Systemic org-ownership check (D4).** Owner: Eng/Sec.
- Pricing/tier gating. Owner: Product.

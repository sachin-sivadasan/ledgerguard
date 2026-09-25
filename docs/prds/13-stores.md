# PRD: Stores (as-built)

**Date:** 2026-09-26 · **Author:** sachin.s · **Status:** Draft (backfilled from implementation)

> **As-built PRD**, reverse-derived from code. Markers: **FACT** / **INFERRED** / **UNKNOWN**.
> Success metrics are **PROPOSED** (future), not historical. Sources:
> `handler/{store_handler,store_health,store_dates}.go`, `entity/shop.go`,
> `external/shopify_storefront_client.go`, `frontend-flutter/lib/screens/stores/`,
> `providers/store_provider.dart`. (Risk engine: #5. Subscriptions: #10. store_health chat tool: #2.)

## Problem & evidence

A partner's app is installed across many merchant **stores**, and a store — not a subscription — is the unit they think about ("which merchants are my best/riskiest customers?"). They need a store-centric view that rolls up all of a store's subscriptions into one health/LTV read.

- **FACT:** Stores is a first-class nav tab with real domain-level aggregation (not a subscriptions re-skin, not a report drill-down), backed by regression-guarded logic (STORE-1/STORE-2).
- **UNKNOWN — original demand evidence not recorded.** Open question — owner: Product.

## Target users

**INFERRED:** the **Shopify app partner** (success/growth) doing account-level analysis of their merchant base. Per-app, per-org.
- **Not the target:** merchants; users who want the raw subscription list (#10) or report-filtered store subsets (churn/at-risk/usage drill-downs).

## Proposed solution (as-built behavior)

**FACT — store list** (`GET /apps/{appID}/stores`, `store_handler.go:42`; `store_list_screen.dart`): a paginated, searchable (domain substring) card grid. Each store card shows: `shop_domain`, **health_score** (0–100), **lifetime_value_cents** (3-year LTV from transactions), **first_install_date** + **last_interaction** (**event-sourced** from `app_events`, not record timestamps — STORE-2, `store_dates.go`), **risk_state** badge, and **installed_app_ids**. Stores are **deduplicated by domain** — multiple subscriptions per store roll up into one card with a priority risk state.

**FACT — health score** (`store_handler.go:213-226`): derived on-the-fly from the persisted risk state — `SAFE→90, ONE_CYCLE_MISSED→50, TWO_CYCLES_MISSED→25, CHURNED→10`. Not a separate composite metric.

**FACT — store detail** (`GET /apps/{appID}/stores/{domain}/health`, `store_health.go:91`; `store_detail_screen.dart`): an overview card (health, LTV, install/last-interaction, risk), installed apps, **linked subscriptions** (clickable → subscription detail, #10), a 3-month transaction view, earnings, and an app-event timeline — enriched with shop **logo/name** when available.

**FACT — regression-guarded correctness:** uses the **persisted** risk state (STORE-1, `store_handler.go:57-61`) rather than re-running the engine, so cancel-trap stores stay SAFE and converge with other surfaces (RISK-1b, #5).

### Key screens

**No new wireframe** (as-built). Real surfaces: `frontend-flutter/lib/screens/stores/store_list_screen.dart` (card grid) + `store_detail_screen.dart` (overview / apps / subscriptions / timeline).

## Platform & policy constraints

**FACT.** Store data is derived from synced subscriptions/transactions/events; shop **logo/name/branding** come from the Shopify **Storefront API** (`ShopifyStorefrontClient.FetchBrand`), cached in `shops` and enriched **opportunistically** during sync (no scheduled refresh). Enrichment is nil-tolerant — a store renders without a logo. Health/dates depend on sync freshness + event coverage.

## Pricing-tier impact

**UNKNOWN — no plan-gating found.** Open question — owner: Product.

## Migration for existing users

**INFERRED — N/A.** Read-only aggregation over rebuilt data; no migration.

## Success metrics (PROPOSED — future-facing)

None measured today. Proposed:
1. **Cross-surface consistency:** a store's risk badge matches that store's subscription risk (and Dashboard/Risk counts) for the same sync ≥ 99% (shares the persisted-state guarantee).
2. **Enrichment coverage:** ≥ 90% of active stores show a real name/logo (not the raw domain) — measures Storefront enrichment health.
3. **Date correctness:** first-install/last-interaction reflect event-sourced dates, never rebuild timestamps, in 100% of audits (guards STORE-2).

## Non-goals (deliberately absent in code)

1. **No composite health model** — health is a fixed risk→number map, not a multi-factor score. **FACT.**
2. **No scheduled logo/branding refresh** — enrichment is opportunistic; stale branding persists. **FACT.**
3. **No cross-app store view** — stores are scoped per app (`installed_app_ids` is informational, not a portfolio roll-up). **INFERRED.**
4. **No store-level actions** — it's a read/browse view; no messaging/outreach from here. **INFERRED.**

## Open questions

- **Health score is a 4-value lookup** — is a richer composite (payment history, tenure, usage) wanted, or is the risk-derived score sufficient? Owner: Product.
- **Opportunistic-only enrichment** — add a refresh cadence so logos/names don't go stale? Owner: Eng.
- **LTV is a 3-year window** — is that the right horizon, and how does it handle refund-negative stores? Owner: Eng/Product.
- **Store detail depends on shop-repo enrichment being present** — graceful without it, but confirm the degraded state is acceptable UX. Owner: Eng.
- Pricing/tier gating. Owner: Product.

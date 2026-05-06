# 12. Subscription Management

## What It Does
Provides CRUD and query operations for Shopify app subscriptions. Supports filtered listing with pagination, sorting, and search. Includes subscription detail views with payment history and risk timeline. Enriches responses with shop brand data (logos, names) when available.

## Architecture
Interface layer handler (`internal/interfaces/http/handler/`) backed by domain repository (`internal/domain/repository/`). Subscriptions are created by the Ledger Rebuild Engine during sync — not directly by users. The handler provides read-only access with rich filtering. Optional dependencies (shop repo, detail service) added via setter injection.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/interfaces/http/handler/subscription.go` | ~519 | HTTP handlers for subscription endpoints |
| `backend/internal/domain/entity/subscription.go` | ~137 | Subscription entity with risk classification |
| `backend/internal/domain/repository/subscription_repository.go` | ~30 | Repository interface with FindWithFilters |
| `backend/internal/infrastructure/persistence/subscription_repository.go` | ~250 | PostgreSQL implementation with filtering |
| `backend/internal/application/service/subscription_detail_service.go` | ~150 | Rich detail with payment history |

## Data Flow
```
┌──────────┐    ┌───────────────────┐    ┌───────────────┐
│ Flutter  │───▶│ SubscriptionHandler│───▶│ SubRepository │
│ App      │    │                   │    │ .FindWithFilters│
└──────────┘    │ 1. Auth check     │    └───────┬───────┘
                │ 2. Parse filters  │            │
                │ 3. Resolve appID  │            ▼
                │ 4. Fetch subs     │    ┌───────────────┐
                │ 5. Enrich logos   │───▶│ ShopRepository│
                │ 6. JSON response  │    │ .FindByDomains│
                └───────────────────┘    └───────────────┘
```

1. Request arrives with appID (numeric, e.g., `4599915`)
2. Handler constructs full Shopify GID (`gid://partners/App/4599915`)
3. Resolves to internal UUID via `appRepo.FindByPartnerAppID()`
4. Applies filters (risk state, price range, billing interval, search)
5. Batch-fetches shop logos for all domains in result set
6. Returns paginated JSON with subscription data + logos

## Configuration
No feature-specific configuration. Uses standard Firebase Auth middleware.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/v1/apps/{appID}/subscriptions | Firebase | Filtered, paginated subscription list |
| GET | /api/v1/apps/{appID}/subscriptions/summary | Firebase | Aggregate stats (active, at-risk, churned counts) |
| GET | /api/v1/apps/{appID}/subscriptions/price-stats | Firebase | Price range and distinct price points |
| GET | /api/v1/apps/{appID}/subscriptions/{subscriptionID} | Firebase | Single subscription with shop logo |
| GET | /api/v1/subscriptions/{id} | Firebase | Subscription detail with risk assessment |
| GET | /api/v1/subscriptions/{id}/history | Firebase | Payment history (limit param) |
| GET | /api/v1/subscriptions/{id}/risk-timeline | Firebase | Risk state change events |

### Query Parameters (List endpoint)

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Comma-separated risk states (SAFE, ONE_CYCLE_MISSED, etc.) |
| `priceMin` | int64 | Minimum price in cents |
| `priceMax` | int64 | Maximum price in cents |
| `billingInterval` | string | MONTHLY or ANNUAL |
| `search` | string | Search by shop name or domain |
| `sortBy` | string | Sort field |
| `sortOrder` | string | asc or desc |
| `page` | int | Page number (default 1) |
| `pageSize` | int | Items per page (default 25, max 100) |

## Dual Identity: ShopifyGID + StableDomainKey

Subscriptions have two identity fields:

- **`shopify_gid`** — The real Shopify subscription GID, mapped from the Partner API `chargeId` field on `AppSubscriptionSale` transactions. This changes when a store uninstalls and reinstalls the app (new subscription = new GID).
- **`stable_domain_key`** — Deterministic key computed as `lg_sub_` + SHA1(`myshopify_domain`). This survives reinstalls because it's derived from the store's permanent domain.

This dual identity enables churn-return analysis: when the same `stable_domain_key` appears with a different `shopify_gid`, it indicates a store reinstalled the app. See ADR-031.

## Extension Points
- Add new filter parameters by extending `SubscriptionFilters` struct
- Implement new detail service methods (e.g., GetChurnPrediction)
- Add CSV/JSON export for subscription lists
- Extend `subscriptionToJSON()` for new response fields

## Gotchas
- **appID in URL is numeric** (e.g., `4599915`), not a UUID. Handler constructs the full Shopify GID prefix `gid://partners/App/`
- **Subscriptions are read-only** via API — they are created/updated only by the sync pipeline
- **Legacy support**: `risk_state` single param and `limit`/`offset` params are still accepted alongside newer `status` and `page`/`pageSize`
- **Shop logo enrichment** is optional — if shopRepo is nil, responses omit logo fields
- **Soft deletes**: FindWithFilters excludes subscriptions where `deleted_at IS NOT NULL`
- **Ownership verification**: All endpoints verify the user owns the partner account that owns the app

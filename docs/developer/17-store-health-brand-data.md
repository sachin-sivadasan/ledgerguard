# 17. Store Health & Brand Data

## What It Does
Provides a detailed health view for a specific Shopify store: its subscription status, recent transaction history, earnings summary, and brand data (logo, shop name, country). Brand data is fetched from the Shopify Storefront API during sync and cached in the `shops` table so the store health endpoint can enrich responses without hitting Shopify on every request.

## Architecture
The feature spans three layers. The **domain layer** defines the `Shop` entity and `ShopRepository` interface. The **infrastructure layer** contains both the `ShopifyStorefrontClient` (external API) and `PostgresShopRepository` (persistence). The **interface layer** provides the `StoreHealthHandler` which aggregates data from four repositories (subscriptions, transactions, partner accounts, apps) plus an optional shop repository for logo enrichment.

The shop repository is injected via setter (`SetShopRepo`) rather than the constructor, making it an optional dependency. If the shop repo is nil, the handler still works but returns responses without logo URLs.

```
Domain Layer                Infrastructure Layer               Interface Layer
┌─────────────┐    ┌──────────────────────────┐    ┌───────────────────────┐
│ Shop entity │    │ ShopifyStorefrontClient  │    │ StoreHealthHandler    │
│             │    │   (Storefront GraphQL)   │    │  - subscriptionRepo   │
│ ShopRepo    │    │                          │    │  - transactionRepo    │
│ (interface) │◀───│ PostgresShopRepository   │    │  - partnerRepo        │
└─────────────┘    │   (upsert, find)         │    │  - appRepo            │
                   └──────────────────────────┘    │  - shopRepo (optional)│
                                                   └───────────────────────┘
```

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/interfaces/http/handler/store_health.go` | ~224 | HTTP handler for store health endpoint |
| `backend/internal/infrastructure/external/shopify_storefront_client.go` | ~152 | Fetches brand data from Shopify Storefront API |
| `backend/internal/domain/entity/shop.go` | ~23 | Shop entity with brand fields |
| `backend/internal/domain/repository/shop_repository.go` | ~14 | ShopRepository interface (Upsert, FindByDomain, FindByDomains) |
| `backend/internal/infrastructure/persistence/shop_repository.go` | ~219 | PostgreSQL implementation with upsert on domain conflict |
| `backend/migrations/000029_create_shops_table.up.sql` | ~17 | Creates `shops` table with unique domain index |

## Data Flow
```
┌──────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐
│ Flutter  │───▶│StoreHealthHandler│───▶│SubscriptionRepo │    │ ShopRepository   │
│ App      │    │                  │    │.FindByAppIDAndDomain│  │ .FindByDomain    │
└──────────┘    │ 1. Auth check    │    └────────┬────────┘    └────────┬─────────┘
                │ 2. Resolve app   │             │                      │
                │ 3. Fetch sub     │             ▼                      ▼
                │ 4. Fetch txns    │    ┌─────────────────┐    ┌──────────────────┐
                │ 5. Fetch earnings│───▶│TransactionRepo  │    │ Shop data        │
                │ 6. Enrich brand  │    │.FindByDomain    │    │ (logo, name,     │
                │ 7. JSON response │    │.GetEarningsSummary│   │  country)        │
                └──────────────────┘    └─────────────────┘    └──────────────────┘
```

### Brand Data Fetch (during sync)
```
SyncService
  │
  ├─ Discovers new myshopify_domain from subscription
  │
  ├─ ShopifyStorefrontClient.FetchBrand(domain)
  │   │
  │   ├─ POST https://{domain}/api/2026-01/graphql.json
  │   │   query: { shop { name, brand { logo, squareLogo, coverImage }, paymentSettings { countryCode } } }
  │   │
  │   └─ Returns Shop entity with brand fields populated
  │
  └─ ShopRepository.Upsert(shop)
      └─ ON CONFLICT (myshopify_domain) DO UPDATE
```

1. During sync, the sync service encounters `myshopify_domain` values from subscriptions
2. For new domains not yet in the `shops` table, `ShopifyStorefrontClient.FetchBrand()` is called
3. The Storefront API is public (no auth token needed) and returns shop name, logo URLs, country code, and currency code
4. The result is upserted into `shops` keyed on `myshopify_domain`
5. On subsequent store health requests, the handler looks up cached brand data from the shop repository

## Configuration
| Setting | Source | Description |
|---------|--------|-------------|
| Storefront API Version | Hardcoded constant `2026-01` | Shopify Storefront API version |
| HTTP timeout | `5 * time.Second` | Timeout for Storefront API calls |
| Transaction window | Last 3 months | `store_health.go` line 136 — `now.AddDate(0, -3, 0)` |

No feature-flag or environment variable required. The Storefront API is unauthenticated (public brand data only).

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/v1/apps/{appID}/stores/{domain}/health | Firebase | Full store health detail |

### Response Shape
```json
{
  "subscription": {
    "id": "uuid",
    "shopify_gid": "gid://partners/AppSubscription/...",
    "myshopify_domain": "example.myshopify.com",
    "shop_name": "Example Store",
    "plan_name": "Premium",
    "base_price_cents": 2999,
    "billing_interval": "MONTHLY",
    "risk_state": "SAFE",
    "status": "ACTIVE",
    "created_at": "2025-01-15T00:00:00Z",
    "last_charge_date": "2025-06-01T00:00:00Z",
    "expected_next_charge": "2025-07-01T00:00:00Z",
    "shop_logo_url": "https://cdn.shopify.com/...",
    "shop_square_logo_url": "https://cdn.shopify.com/..."
  },
  "transactions": [
    {
      "id": "uuid",
      "charge_type": "RECURRING",
      "gross_amount_cents": 2999,
      "net_amount_cents": 2399,
      "currency": "USD",
      "transaction_date": "2025-06-01T00:00:00Z",
      "earnings_status": "AVAILABLE"
    }
  ],
  "earnings": {
    "pending_cents": 0,
    "available_cents": 2399,
    "paid_out_cents": 14394,
    "total_cents": 16793
  }
}
```

## Extension Points
- Add a `FindByDomains()` batch method to enrich subscription lists with logos (already exists in the interface)
- Extend the Storefront query to fetch more brand fields (e.g., `shortDescription`, `colors`)
- Add store health history by combining daily snapshots with per-store transaction trends
- Implement webhook-driven brand refresh when Shopify sends shop/update events

## Gotchas
- **Storefront API is public** -- no access token required, but some shops may have it disabled or return empty brand data. `FetchBrand()` returns an empty `Shop` on failure rather than erroring.
- **appID in URL is numeric** (e.g., `4599915`). The handler constructs the full GID prefix `gid://partners/App/` internally.
- **Shop repo is optional** -- injected via `SetShopRepo()` setter, not the constructor. If nil, responses omit `shop_logo_url` and `shop_square_logo_url` fields.
- **Brand data is cached** -- fetched once during sync and stored in `shops`. There is no TTL or automatic refresh; the upsert overwrites on re-sync.
- **Transactions window is 3 months** -- hardcoded in the handler. The earnings summary covers all time for that store.
- **Storefront API version** is a constant (`2026-01`). Update this when Shopify deprecates the version.
- **Nullable columns** in `shops` -- all brand fields except `myshopify_domain` are nullable. The persistence layer uses pointer scanning to handle NULLs.
- **Org-scoped resolution.** The handler resolves the partner account via `resolvePartnerAccount(r, partnerRepo)` (org context → `FindByOrgID`, fallback → `FindByUserID`). See ADR-035 for the org-scoped data access pattern.

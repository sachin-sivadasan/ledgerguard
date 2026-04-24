# 04. App Management & Selection

## What It Does
Enables users to discover, select, and manage which Shopify apps LedgerGuard tracks. After connecting a Shopify Partner account (via OAuth or manual token), users fetch their available apps from the Partner API and select which ones to monitor. Each app tracks its revenue share tier (determining Shopify's commission), install count, and app store slug. The app entity is the central reference point that links a Partner account to its transactions, subscriptions, metrics, and store health data.

## Architecture
Spans all four DDD layers. The app management flow bridges the Shopify Partner integration (doc 03) with the transaction sync engine (doc 05):

- **Domain layer** — `App` entity with fields for partner account linkage, tracking state, revenue share tier, and install count. `RevenueShareTier` value object encapsulates Shopify's tiered commission structure with fee calculation logic. `AppRepository` interface defines the port.
- **Infrastructure layer** — `PostgresAppRepository` implements CRUD operations with COALESCE defaults for backward-compatible columns. `ShopifyPartnerClient.FetchApps()` discovers apps by querying transaction history from the Partner API. `FetchInstallCount()` calculates current installs from relationship events.
- **Interfaces layer** — `AppHandler` exposes endpoints for listing available apps, selecting apps to track, listing tracked apps, updating revenue share tiers, updating store slugs, and refreshing install counts.
- **Application layer** — No dedicated app service; the handler directly coordinates between the Partner client, repositories, and encryption. The `SyncService` later uses the app's `TrackingEnabled` flag to determine which apps to sync.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/interfaces/http/handler/app.go` | ~498 | AppHandler: GetAvailableApps, SelectApp, ListApps, UpdateAppTier, RefreshInstallCount, GetInstallCount, UpdateStoreSlug |
| `backend/internal/domain/entity/app.go` | ~49 | App entity: ID, PartnerAccountID, PartnerAppID (Shopify GID), Name, TrackingEnabled, RevenueShareTier, InstallCount, AppStoreSlug |
| `backend/internal/domain/repository/app_repository.go` | ~45 | AppRepository interface: Create, FindByID, FindByPartnerAccountID, FindByPartnerAppID, FindAllByPartnerAppID, Update, Delete |
| `backend/internal/infrastructure/persistence/app_repository.go` | ~247 | PostgreSQL implementation with COALESCE defaults for nullable columns |
| `backend/internal/domain/valueobject/revenue_share_tier.go` | — | RevenueShareTier: SMALL_DEV_0, DEFAULT_20, PREMIUM_15. Includes fee calculation, display names, and validation |
| `backend/internal/infrastructure/external/shopify_partner_client.go` | ~1045 | FetchApps (from transaction history), FetchInstallCount (from relationship events with pagination) |

## Data Flow

### App Discovery and Selection
```
┌──────────┐                    ┌───────────────┐                  ┌──────────────────┐
│  Flutter  │                    │  LedgerGuard  │                  │ Shopify Partners │
│   App     │                    │   Backend     │                  │      API         │
└─────┬─────┘                    └───────┬───────┘                  └────────┬─────────┘
      │                                  │                                   │
      │  1. GET /api/v1/apps/available   │                                   │
      │ ────────────────────────────────▶│                                   │
      │                                  │                                   │
      │                         2. partnerRepo.FindByUserID()                │
      │                         3. decryptor.Decrypt(token)                  │
      │                                  │                                   │
      │                         4. FetchApps(orgID, token)                   │
      │                                  │──────────────────────────────────▶│
      │                                  │  5. GraphQL: transactions query   │
      │                                  │  (extracts unique apps)           │
      │                                  │◀──────────────────────────────────│
      │                                  │                                   │
      │  6. Return {apps: [{id, name}]}  │                                   │
      │ ◀────────────────────────────────│                                   │
      │                                  │                                   │
      │  7. POST /api/v1/apps/select     │                                   │
      │  {partner_app_id, name}          │                                   │
      │ ────────────────────────────────▶│                                   │
      │                                  │                                   │
      │                         8. Check for duplicates                      │
      │                         9. entity.NewApp()                           │
      │                            → TrackingEnabled: true                   │
      │                            → RevenueShareTier: SMALL_DEV_0          │
      │                        10. appRepo.Create()                          │
      │                                  │                                   │
      │  11. Return {id, name, tier}     │                                   │
      │ ◀────────────────────────────────│                                   │
```

### Tracked App Lifecycle
```
                    ┌───────────────────────────────┐
                    │         App Entity             │
                    │                                │
    SelectApp ──────▶  Created (TrackingEnabled=true)│
                    │  RevenueShareTier=SMALL_DEV_0  │
                    │                                │
                    │         │                      │
                    │         ▼                      │
                    │  SyncService checks            │
    SyncAllApps ───▶  TrackingEnabled flag           │
                    │  → true: fetch transactions    │
                    │  → false: skip                 │
                    │                                │
                    │         │                      │
                    │         ▼                      │
    UpdateAppTier──▶  Tier updated                   │
                    │  (affects fee calculations)     │
                    │                                │
                    │         │                      │
                    │         ▼                      │
    RefreshInstall─▶  InstallCount updated           │
    Count           │  (from Partner API events)     │
                    │                                │
                    │         │                      │
                    │         ▼                      │
    UpdateStoreSlug▶  AppStoreSlug set               │
                    │  (enables review scraping)     │
                    └───────────────────────────────┘
```

### App ID Resolution
```
Handler receives appID from URL path (e.g., /apps/{appID}/...)

  appID string
     │
     ├── Try uuid.Parse(appID)
     │   ├── Success → appRepo.FindByID(uuid)
     │   │
     │   └── Failure → treat as Shopify GID
     │       ├── starts with "gid://" → use as-is
     │       └── numeric only → prepend "gid://partners/App/"
     │           └── appRepo.FindByPartnerAppID(partnerAccountID, gid)
     │
     ▼
  *entity.App
```

## Configuration
No feature-specific configuration variables. App management uses the Shopify Partner client (configured in doc 03) and the database connection (configured in doc 01).

The default revenue share tier (`SMALL_DEV_0` = 0% commission) is hardcoded in `entity.NewApp()` per ADR-008, reflecting Shopify's policy of 0% revenue share on the first $1M in annual app revenue for developers earning under that threshold.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/apps/available` | Firebase | Fetch available apps from Shopify Partner API (from transaction history). Returns `{apps: [{id, name}]}`. |
| `POST` | `/api/v1/apps/select` | Firebase | Select an app to track. Body: `{partner_app_id, name}`. Returns 409 if already tracked. |
| `GET` | `/api/v1/apps` | Firebase | List all tracked apps for the current user. Returns `{apps: [{id, name, tracking_enabled, revenue_share_tier, install_count, app_store_slug, created_at, updated_at}]}`. |
| `PATCH` | `/api/v1/apps/{appID}/tier` | Firebase | Update revenue share tier. Body: `{revenue_share_tier}`. AppID accepts UUID or numeric Shopify ID. |
| `GET` | `/api/v1/apps/{appID}/install-count` | Firebase | Get current install count for an app. |
| `POST` | `/api/v1/apps/{appID}/refresh-install-count` | Firebase | Refresh install count from Partner API (fetches install/uninstall events, calculates net installs). |
| `PATCH` | `/api/v1/apps/{appID}/store-slug` | Firebase | Set the Shopify App Store slug (e.g., `whatsapp-button-chat`). Enables review scraping. |

### App Entity Fields
| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Internal primary key |
| `partner_account_id` | UUID | FK to partner_accounts table |
| `partner_app_id` | string | Shopify GID format: `gid://partners/App/12345` |
| `name` | string | App display name from Shopify |
| `tracking_enabled` | bool | Whether sync should process this app (default: true) |
| `revenue_share_tier` | RevenueShareTier | Shopify commission tier (default: SMALL_DEV_0) |
| `install_count` | int | Current number of shops with app installed |
| `app_store_slug` | string | Shopify App Store URL slug for review scraping |
| `created_at` | time.Time | Record creation timestamp |
| `updated_at` | time.Time | Last modification timestamp |

### Revenue Share Tiers
| Tier | Commission | Description |
|------|-----------|-------------|
| `SMALL_DEV_0` | 0% | First $1M annual revenue (default for new apps) |
| `DEFAULT_20` | 20% | Standard Shopify rate above $1M |
| `PREMIUM_15` | 15% | Reduced rate for certain premium partners |

## Extension Points
- **New app discovery method** — The current `FetchApps` extracts apps from transaction history. An alternative could query the Partner API's app listing directly. Implement in `ShopifyPartnerClient` and wire to the handler.
- **Tracking toggle** — The `TrackingEnabled` field exists on the entity but there is no dedicated toggle endpoint. Add a `POST /api/v1/apps/{id}/tracking` handler that calls `appRepo.Update()` to flip the flag.
- **Custom revenue share tiers** — Add new tiers to `valueobject.RevenueShareTier`. Each tier needs `RevenueSharePercent()`, `DisplayName()`, `Description()`, and `CalculateFeeBreakdown()` implementations.
- **Bulk app import** — Currently apps are selected one at a time. A bulk endpoint could accept an array of `{partner_app_id, name}` pairs and create all at once.

## Gotchas
- **`partner_app_id` is a Shopify GID.** Format: `gid://partners/App/12345`. The handler extracts the numeric part for display but stores the full GID. When accepting IDs in URL paths, the handler tries UUID first, then prepends the GID prefix if the input is numeric-only.
- **Default revenue tier is 0% (SMALL_DEV_0).** Per ADR-008, new apps default to Shopify's 0% tier for developers earning under $1M annually. This is the correct default for most indie developers. Fee calculations using a wrong tier will produce incorrect results.
- **Install count from events, not API count.** The install count is computed by fetching all `RELATIONSHIP_INSTALLED` and `RELATIONSHIP_UNINSTALLED` events and calculating the difference. This is paginated and can be slow for apps with many installs. The count is cached in the database and updated only when explicitly refreshed.
- **App discovery depends on transaction history.** `FetchApps` finds apps by scanning transaction records. An app with no transactions (e.g., free app, new app with no paying users) will not appear in the available apps list.
- **COALESCE in SQL queries.** The PostgreSQL repository uses `COALESCE` for `revenue_share_tier`, `install_count`, and `app_store_slug` to handle rows created before these columns existed. Defaults are `DEFAULT_20`, `0`, and `''` respectively.
- **App store slug enables review scraping.** Without a slug set, the review handler cannot scrape reviews from the Shopify App Store. The slug must be set manually via the PATCH endpoint or updated during sync.

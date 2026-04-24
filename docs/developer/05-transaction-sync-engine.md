# 05. Transaction Sync Engine

## What It Does
Orchestrates the synchronization of financial transactions from Shopify Partner API into LedgerGuard. Fetches a rolling 12-month window of transactions, processes earnings tracking, stores them idempotently, and triggers the ledger rebuild pipeline. After the core sync, it runs three optional enrichment phases: subscription status enrichment from app lifecycle events, shop brand data fetching from Shopify Storefront API, and app review scraping from the Shopify App Store.

## Architecture
Application layer service (`internal/application/service/`). Uses the builder pattern for optional dependencies — the constructor requires only the core dependencies (TransactionFetcher, txRepo, appRepo, partnerRepo, Decryptor, LedgerRebuilder), while enrichment dependencies are attached via `WithEventFetcher()`, `WithSubscriptionRepo()`, `WithShopBrandFetcher()`, and `WithReviewScraper()` methods. This keeps the sync functional even when optional subsystems are not configured.

The service coordinates between external APIs (Shopify Partner, Storefront, App Store) and domain services (LedgerService, EarningsCalculator). It follows the full-rebuild strategy defined in ADR-002: every sync fetches the entire 12-month history, stores raw transactions immutably, and rebuilds the entire ledger from scratch.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/application/service/sync_service.go` | ~401 | Sync orchestration, builder pattern, enrichment phases |
| `backend/internal/application/service/sync_service_test.go` | ~356 | Unit tests with mocked dependencies |
| `backend/internal/infrastructure/external/shopify_partner_client.go` | ~1044 | Partner API GraphQL client with rate limiting |
| `backend/internal/interfaces/http/handler/sync.go` | ~127 | POST /api/v1/sync and /api/v1/sync/{appID} endpoints |
| `backend/internal/domain/service/earnings_calculator.go` | ~118 | Earnings availability date calculation and status tracking |

## Data Flow
```
┌────────────────────┐
│   HTTP Handler     │
│ POST /api/v1/sync  │
└─────────┬──────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        SyncService.SyncApp()                        │
│                                                                     │
│  1. appRepo.FindByID()           ← Get app metadata                │
│  2. partnerRepo.FindByID()       ← Get partner account             │
│  3. decryptor.Decrypt()          ← Decrypt Shopify access token     │
│  4. fetcher.FetchTransactions()  ← Partner API (12-month window)   │
│  5. earningsCalc.Process()       ← Set available dates & statuses  │
│  6. txRepo.UpsertBatch()         ← Idempotent transaction storage  │
│  7. ledger.RebuildFromTx()       ← Full ledger rebuild + snapshot  │
│  8. ledger.BackfillSnapshots()   ← Historical monthly snapshots    │
│                                                                     │
│  ── Optional Enrichment (errors ignored) ──                         │
│  9. enrichSubscriptionStatus()   ← App lifecycle events → status   │
│ 10. fetchShopBrands()            ← Storefront API → logos          │
│ 11. scrapeAndStoreReviews()      ← App Store → reviews             │
└─────────────────────────────────────────────────────────────────────┘
          │
          ▼
┌────────────────────┐
│    SyncResult      │
│ - TransactionCount │
│ - RiskSummary      │
│ - TotalMRRCents    │
│ - SyncedAt         │
└────────────────────┘
```

### SyncAllApps Flow
```
SyncAllApps(partnerAccountID)
  │
  ├── appRepo.FindByPartnerAccountID()
  │
  └── for each app where TrackingEnabled == true:
        └── SyncApp(appID)
              │
              ├── success → append SyncResult
              └── error   → append SyncResult with Error field set, continue
```

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `SHOPIFY_PARTNER_API_KEY` | — | Yes | Partner API access key |
| `ENCRYPTION_KEY` | — | Yes | AES-256-GCM key for decrypting stored access tokens |
| `SYNC_INTERVAL` | 12h | No | Scheduled sync interval for background job |

The Partner API client also respects Shopify's rate limit of 4 requests/second using an internal token bucket.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/sync` | Firebase | Sync all apps for the current user's partner account |
| POST | `/api/v1/sync/{appID}` | Firebase | Sync a specific app (with tenant isolation check) |

Both endpoints return JSON with `app_id`, `app_name`, `transaction_count`, and `synced_at`. The `/sync/{appID}` endpoint verifies that the app belongs to the authenticated user's partner account before proceeding.

## Extension Points
- **TransactionFetcher interface** — implement for non-Shopify data sources (e.g., Stripe, Paddle). The interface requires only `FetchTransactions(ctx, accessToken, appID, from, to)`.
- **Builder pattern** — add new enrichment steps via `WithXxx()` methods without modifying the constructor or breaking existing callers.
- **EventFetcher interface** — swap the Shopify app event source for a webhook-based approach.
- **ShopBrandFetcher interface** — replace Storefront API with a different brand data provider.
- **ReviewScraper interface** — swap the HTML scraping strategy for an official API when available.
- **LedgerRebuilder interface** — the sync service is decoupled from ledger internals; any implementation that satisfies `RebuildFromTransactions` and `BackfillHistoricalSnapshots` works.

## Gotchas
- **Enrichment errors are silently ignored.** Subscription status enrichment, shop brand fetching, and review scraping all swallow errors to avoid blocking the core transaction sync. Check application logs for enrichment failures.
- **Rate limiting:** Shopify Partner API allows 4 requests/second; the client uses a token bucket internally. First sync backfills 12 months of history, which may trigger rate limiting and take 30+ seconds.
- **SyncAllApps skips disabled apps.** Only apps with `TrackingEnabled == true` are synced. Newly added apps must be explicitly enabled.
- **Access token is encrypted at rest.** The Decryptor interface is called on every sync to decrypt the AES-256-GCM encrypted Shopify access token. If the encryption key rotates, tokens must be re-encrypted.
- **Review scraping is capped at 2 pages (~20 reviews)** during sync to keep the operation fast. A separate full scrape can be triggered independently.
- **BackfillHistoricalSnapshots errors are ignored** (double underscore: both the backfill call itself and FindByAppID failures are swallowed). Historical snapshots are best-effort.
- **Tenant isolation** is enforced at the handler level — the `SyncApp` endpoint verifies that `app.PartnerAccountID == user's partnerAccount.ID` before proceeding.

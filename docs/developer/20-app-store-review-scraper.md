# 20. App Store Review Scraper

## What It Does
Scrapes user reviews from the Shopify App Store HTML pages and stores them for analysis. Reviews are fetched by crawling `https://apps.shopify.com/{slug}/reviews?page=N`, parsing the HTML with goquery, and upserting into the `app_reviews` table with deduplication on a deterministic hash of author + date + body. Supports both automated scraping during sync (2 pages) and manual triggering via API (configurable page count).

## Architecture
```
┌────────────┐    ┌──────────────────┐    ┌───────────────────────┐
│ Flutter    │───▶│  ReviewHandler   │───▶│ ShopifyAppStoreClient │
│ App        │    │                  │    │                       │
└────────────┘    │ GET  list        │    │ ScrapeReviews(slug,   │
                  │ POST scrape      │    │   maxPages)           │
                  └──────┬───────────┘    │                       │
                         │                │ scrapePage()           │
                         │                │ parseReview()          │
                         ▼                └───────────────────────┘
                  ┌──────────────────┐              │
                  │AppReviewRepository│              ▼
                  │                  │    ┌───────────────────────┐
                  │ UpsertBatch()   │    │ Shopify App Store     │
                  │ FindByAppID()   │    │ HTML pages            │
                  │ CountByAppID()  │    │ (public, no auth)     │
                  └──────────────────┘    └───────────────────────┘
```

The scraper client lives in the infrastructure layer (`external/`) since it accesses an external service. The handler in the interface layer coordinates: resolving the app, triggering the scrape, converting scraped reviews to domain entities, and persisting via the repository.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/infrastructure/external/shopify_appstore_client.go` | ~215 | HTML scraper using goquery; parses rating, author, body, date, location, time using |
| `backend/internal/domain/entity/app_review.go` | ~23 | AppReview domain entity |
| `backend/internal/domain/repository/app_review_repository.go` | ~14 | Repository interface (UpsertBatch, FindByAppID, CountByAppID) |
| `backend/internal/infrastructure/persistence/app_review_repository.go` | ~121 | PostgreSQL implementation with upsert on (app_id, source_review_id) |
| `backend/internal/interfaces/http/handler/review.go` | ~222 | HTTP handlers for listing reviews and triggering scrapes |
| `backend/migrations/000032_create_app_reviews_table.up.sql` | ~20 | Creates `app_reviews` table with dedup constraint and indexes |

## Data Flow

### Scrape Flow
```
ReviewHandler.Scrape()
  │
  ├─ 1. Auth check (Firebase middleware)
  ├─ 2. Resolve app (UUID or partner app ID)
  ├─ 3. Verify app_store_slug is set
  │
  ├─ 4. ShopifyAppStoreClient.ScrapeReviews(slug, maxPages)
  │      │
  │      ├─ For each page (1..maxPages):
  │      │    │
  │      │    ├─ GET https://apps.shopify.com/{slug}/reviews?page={n}
  │      │    │   User-Agent: Chrome 120 (avoids bot detection)
  │      │    │
  │      │    ├─ Parse HTML with goquery
  │      │    │   └─ div[data-merchant-review]  →  one review per div
  │      │    │       ├─ div[aria-label$="stars"]     →  rating (1-5)
  │      │    │       ├─ .tw-text-body-xs              →  date
  │      │    │       ├─ div[data-truncate-content-copy] p  →  body
  │      │    │       ├─ .tw-text-heading-xs span[title]    →  author
  │      │    │       └─ .tw-order-1 children               →  location, time using
  │      │    │
  │      │    ├─ Check for a[rel="next"] to determine if more pages exist
  │      │    └─ Rate limit: 1 second delay between pages
  │      │
  │      └─ Return []ScrapedReview
  │
  ├─ 5. Convert to []*entity.AppReview
  │      ├─ Generate SourceReviewID: SHA-256(author + date + body[:100])[:32hex]
  │      ├─ Skip reviews with rating=0 or zero date (malformed)
  │      └─ Set source = "shopify_app_store"
  │
  └─ 6. AppReviewRepository.UpsertBatch(reviews)
         └─ INSERT ... ON CONFLICT (app_id, source_review_id) DO UPDATE
            (updates rating, body, location, time_using, scraped_at, updated_at)
```

### Deduplication Strategy
```
SourceReviewID = SHA-256( author + date("2006-01-02") + body[:100] )[:32 hex chars]

Unique constraint: (app_id, source_review_id)

On conflict: UPDATE mutable fields (body may be edited, location/time_using may change)
```

This deterministic hash means re-scraping the same reviews produces the same `source_review_id`, triggering an upsert instead of a duplicate insert.

### Sentiment Derivation
The handler derives sentiment from the numeric rating at read time (not stored):
| Rating | Sentiment |
|--------|-----------|
| 1-2 | negative |
| 3 | neutral |
| 4-5 | positive |

## Configuration
| Setting | Source | Description |
|---------|--------|-------------|
| HTTP timeout | `15 * time.Second` | Timeout per page fetch |
| Rate limit | `1 * time.Second` | Delay between page fetches |
| Default max pages | 5 | Used when `maxPages <= 0` in ScrapeReviews |
| Default max pages (sync) | 2 | During automated sync, only 2 pages (~20 reviews) for speed |
| User-Agent | Chrome 120 on macOS | Hardcoded to avoid bot detection |

### Prerequisite
The app entity must have `app_store_slug` set (e.g., `"my-cool-app"`). Without it, the scrape endpoint returns HTTP 400. The slug corresponds to the URL path on `apps.shopify.com`.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/v1/apps/{appID}/reviews | Firebase | Paginated review list with sentiment |
| POST | /api/v1/apps/{appID}/reviews/scrape | Firebase | Trigger manual review scrape |

### List Reviews Request
```
GET /api/v1/apps/{appID}/reviews?page=1&per_page=20
```

### List Reviews Response
```json
{
  "reviews": [
    {
      "id": "uuid",
      "author": "Store Owner",
      "rating": 5,
      "body": "Great app, really helped with...",
      "review_date": "2025-06-15",
      "location": "United Kingdom",
      "time_using": "4 months using the app",
      "sentiment": "positive",
      "source": "shopify_app_store",
      "scraped_at": "2025-07-01T12:00:00Z"
    }
  ],
  "total": 47,
  "page": 1,
  "per_page": 20
}
```

### Scrape Request / Response
```json
// POST /api/v1/apps/{appID}/reviews/scrape
// Request (optional body)
{ "max_pages": 10 }

// Response
{
  "message": "Scrape completed",
  "new_reviews": 23,
  "total_reviews": 47
}
```

### App ID Resolution
The `{appID}` parameter accepts two formats:
1. **UUID** -- looked up directly via `appRepo.FindByID()`
2. **Numeric partner app ID** (e.g., `4599915`) -- converted to GID `gid://partners/App/4599915` and resolved via `appRepo.FindByPartnerAppID()`

## Extension Points
- Add sentiment analysis using an LLM or NLP library instead of rating-based derivation
- Implement review change detection (compare body hashes to flag edited reviews)
- Add review aggregation stats (average rating, rating distribution, trend over time)
- Support scraping from other sources (G2, Capterra) by adding new scraped review sources
- Integrate with the AI chat module to answer questions like "What are customers saying?"
- Add webhook/notification when new negative reviews are detected
- Implement review response tracking (developer replies)

## Gotchas
- **HTML selectors are fragile**: The scraper depends on Shopify App Store CSS class names (e.g., `.tw-text-heading-xs`, `.tw-order-1`, `div[data-merchant-review]`). Shopify can change these at any time, breaking the scraper silently (it returns empty results rather than errors).
- **Rate limiting**: There is a 1-second delay between page fetches. Aggressive scraping risks IP bans from Shopify. During sync, only 2 pages are scraped to stay fast.
- **No auth required**: The Shopify App Store is public. No API key or OAuth token is needed.
- **Malformed reviews are skipped**: Reviews with `rating == 0` or `date.IsZero()` are silently dropped during conversion to domain entities.
- **Date parsing**: Supports three layouts (`"January 2, 2006"`, `"Jan 2, 2006"`, `"2006-01-02"`). If Shopify changes the date format, reviews will have zero-value dates and be skipped.
- **app_store_slug must be set manually**: There is no automatic discovery of the slug from the Shopify Partner API. It must be set on the app entity (via the migration `000031_add_app_store_slug`).
- **Context cancellation**: The scraper checks `ctx.Done()` between pages, so cancelling the request stops further page fetches but returns whatever was already scraped.
- **Stops on first error**: If any page fetch fails (network error, non-200 status), the scraper stops pagination and returns the reviews collected so far. It does not retry.
- **new_reviews count may be misleading**: The response field `new_reviews` counts all reviews passed to `UpsertBatch`, including those that triggered updates (not just inserts). The actual number of new-vs-updated reviews is not tracked.

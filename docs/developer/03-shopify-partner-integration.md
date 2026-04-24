# 03. Shopify Partner Integration

## What It Does
Connects LedgerGuard to a user's Shopify Partners account so the platform can fetch financial transaction data. Supports two integration paths: an OAuth 2.0 flow where Shopify redirects back with an authorization code, and a manual token upload where an admin pastes a Partner API token directly. In both cases, the access token is encrypted with AES-256-GCM before storage. Once connected, the ShopifyPartnerClient provides a rate-limited, retry-aware GraphQL client for querying the Partner API for transactions, app events, and install counts.

## Architecture
Spans all four DDD layers (ADR-006 for CSRF state parameter protection):

- **Domain layer** — `PartnerAccount` entity stores the integration record. `IntegrationType` value object distinguishes OAUTH vs MANUAL.
- **Infrastructure layer** — `ShopifyOAuthService` handles the OAuth flow (auth URL generation, code exchange, organization ID fetch). `ShopifyPartnerClient` handles all Partner API GraphQL queries with token bucket rate limiting, exponential backoff, and per-partner rate limiter isolation.
- **Interfaces layer** — `OAuthHandler` exposes `/oauth` and `/callback` endpoints. `ManualTokenHandler` exposes token CRUD for admin users. `IntegrationStatusHandler` reports connection status.
- **Application layer** — The `OAuthStateStore` (in-memory cache with 10-min TTL) validates CSRF state parameters with one-time use semantics.

## Key Files
| File | Lines | Purpose |
|------|-------|---------|
| `backend/internal/infrastructure/external/shopify_oauth.go` | ~178 | OAuth flow: GenerateAuthURL(), ExchangeCodeForToken(), FetchOrganizationID() via Partner API GraphQL |
| `backend/internal/interfaces/http/handler/oauth.go` | ~169 | StartOAuth (GET, generates URL + state), Callback (GET, exchanges code, encrypts token, creates PartnerAccount) |
| `backend/internal/interfaces/http/handler/manual_token.go` | ~163 | AddToken (POST), GetToken (GET, masked), RevokeToken (DELETE) for manual Partner API tokens |
| `backend/internal/domain/entity/partner_account.go` | ~29 | PartnerAccount entity: id, user_id, partner_id, integration_type, encrypted_access_token |
| `backend/internal/infrastructure/external/shopify_partner_client.go` | ~1045 | Partner API GraphQL client: FetchApps, FetchTransactions (paginated), FetchAppEvents, FetchInstallCount; token bucket rate limiting with per-partner isolation, exponential backoff with jitter |
| `backend/internal/infrastructure/cache/oauth_state_store.go` | — | In-memory state store with TTL and one-time-use validation |
| `backend/pkg/crypto/aes_encryptor.go` | — | AES-256-GCM encryption/decryption for token storage |
| `backend/internal/domain/valueobject/integration_type.go` | — | IntegrationType: OAUTH, MANUAL |
| `backend/internal/domain/repository/partner_account_repository.go` | — | PartnerAccountRepository interface (port) |

## Data Flow

### OAuth Flow
```
┌──────────┐                    ┌───────────────┐                  ┌──────────────────┐
│  Flutter  │                    │  LedgerGuard  │                  │ Shopify Partners │
│   App     │                    │   Backend     │                  │                  │
└─────┬─────┘                    └───────┬───────┘                  └────────┬─────────┘
      │                                  │                                   │
      │  1. GET /integrations/           │                                   │
      │     shopify/oauth                │                                   │
      │ ────────────────────────────────▶│                                   │
      │                                  │                                   │
      │                         2. Generate random state                     │
      │                            Store state + userID                      │
      │                            (10-min TTL, one-time use)                │
      │                                  │                                   │
      │  3. Return {url, state}          │                                   │
      │ ◀────────────────────────────────│                                   │
      │                                  │                                   │
      │  4. Open URL in browser          │                                   │
      │ ─────────────────────────────────────────────────────────────────────▶│
      │                                  │                                   │
      │                                  │  5. Redirect to callback          │
      │                                  │     ?code=xxx&state=yyy           │
      │                                  │ ◀─────────────────────────────────│
      │                                  │                                   │
      │                         6. Validate state (one-time use)             │
      │                         7. ExchangeCodeForToken(code)                │
      │                                  │ ─────────────────────────────────▶│
      │                                  │  8. Return access_token           │
      │                                  │ ◀─────────────────────────────────│
      │                                  │                                   │
      │                         9. FetchOrganizationID(token)                │
      │                                  │ ─────────────────────────────────▶│
      │                                  │  10. Return org GID               │
      │                                  │ ◀─────────────────────────────────│
      │                                  │                                   │
      │                        11. Encrypt token (AES-256-GCM)               │
      │                        12. Create PartnerAccount record              │
      │                                  │                                   │
      │  13. Return {message, id}        │                                   │
      │ ◀────────────────────────────────│                                   │
```

### Manual Token Flow
```
┌──────────┐                    ┌───────────────┐
│  Admin    │                    │  LedgerGuard  │
│  User     │                    │   Backend     │
└─────┬─────┘                    └───────┬───────┘
      │                                  │
      │  POST /integrations/             │
      │  shopify/token                   │
      │  {token, partner_id}             │
      │ ────────────────────────────────▶│
      │                                  │
      │                         1. Verify ADMIN/OWNER role
      │                         2. Encrypt token (AES-256-GCM)
      │                         3. Upsert PartnerAccount
      │                                  │
      │  Return {id, masked_token}       │
      │ ◀────────────────────────────────│
```

### Partner API Rate Limiting
```
┌────────────────────────────────────────────────────────┐
│               ShopifyPartnerClient                      │
│                                                         │
│  Per-Partner Rate Limiters (keyed by org ID):           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐               │
│  │ Org 123  │ │ Org 456  │ │ Org 789  │  ...          │
│  │ 4 RPS    │ │ 4 RPS    │ │ 4 RPS    │               │
│  │ Bucket   │ │ Bucket   │ │ Bucket   │               │
│  └──────────┘ └──────────┘ └──────────┘               │
│                                                         │
│  Retry Strategy:                                        │
│  - Max 3 retries                                        │
│  - Exponential backoff: 1s, 2s, 4s (capped at 30s)    │
│  - Jitter: +/-25%                                       │
│  - Respects Retry-After header on 429                  │
│  - Retries on 429 and 5xx only                         │
└────────────────────────────────────────────────────────┘
```

## Configuration
| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `SHOPIFY_CLIENT_ID` | — | Yes (for OAuth) | Shopify Partners app client ID |
| `SHOPIFY_CLIENT_SECRET` | — | Yes (for OAuth) | Shopify Partners app client secret |
| `SHOPIFY_REDIRECT_URI` | — | Yes (for OAuth) | Callback URL registered with Shopify (e.g., `https://api.example.com/api/v1/integrations/shopify/callback`) |
| `SHOPIFY_SCOPES` | — | Yes (for OAuth) | OAuth scopes, typically `read_financials,read_apps` |
| `SHOPIFY_RATE_LIMIT_RPS` | `4` | No | Requests per second for Partner API calls (per partner) |
| `ENCRYPTION_KEY` | — | Yes | 32-byte AES-256 master key for encrypting access tokens at rest |

If `SHOPIFY_CLIENT_ID` is empty, the OAuth service is not initialized and OAuth routes are not registered. Manual token upload still works if encryption is configured.

## API Surface
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/integrations/shopify/oauth` | Firebase | Generates OAuth URL and CSRF state parameter. Returns `{url, state}`. |
| `GET` | `/api/v1/integrations/shopify/callback` | None (state-validated) | OAuth callback from Shopify. Exchanges code for token, encrypts, stores. |
| `POST` | `/api/v1/integrations/shopify/token` | Firebase + ADMIN | Upload a manual Partner API token. Body: `{token, partner_id}`. |
| `GET` | `/api/v1/integrations/shopify/token` | Firebase + ADMIN | Retrieve masked token info for the current partner account. |
| `DELETE` | `/api/v1/integrations/shopify/token` | Firebase + ADMIN | Revoke (delete) the partner account and stored token. |
| `GET` | `/api/v1/integrations/shopify/status` | Firebase | Check if user has a connected partner account. |

### Partner API Endpoint
All Partner API queries go to: `https://partners.shopify.com/{org_id}/api/2025-07/graphql.json`

The client supports these queries:
- `FetchApps` — List apps from transaction history
- `FetchTransactions` — Paginated transaction fetch with date range
- `FetchAppEvents` — App lifecycle events (install, uninstall, subscription changes)
- `FetchInstallCount` — Current install count from relationship events

## Extension Points
- **New OAuth provider** — Create a new service implementing the `OAuthService` interface defined in `handler/oauth.go` (GenerateAuthURL, ExchangeCodeForToken, FetchOrganizationID). Wire it in `main.go`.
- **Persistent state store** — Replace the in-memory `OAuthStateStore` with a Redis or database-backed implementation. The interface (`Store`, `Validate`) is defined in `handler/oauth.go`.
- **Additional Partner API queries** — Add methods to `ShopifyPartnerClient`. The `executeWithRetry` method handles rate limiting and retries automatically for any new GraphQL query.
- **Token rotation** — The `PartnerAccount` entity supports updating encrypted tokens via the repository's `Update` method. A background job could periodically re-encrypt with rotated keys.

## Gotchas
- **State parameter is one-time use.** The CSRF state token is consumed on validation and cannot be reused. If the callback fails after state validation but before account creation, the user must restart the OAuth flow.
- **Token encrypted at rest.** Access tokens are never stored in plaintext. All reads require decryption. If the encryption key changes, existing tokens become unreadable.
- **Rate limit is 4 req/s per partner.** Shopify enforces this on the Partner API. The client creates a separate token bucket per organization ID. Exceeding the limit results in 429 responses that trigger exponential backoff.
- **OAuth callback is unauthenticated.** The `/callback` endpoint cannot use Firebase auth because it receives a redirect from Shopify (no Bearer token). Security comes from the CSRF state parameter which maps back to a specific user ID.
- **Organization ID extraction.** The `FetchOrganizationID` method queries the Partner API to get the org GID (e.g., `gid://partners/Organization/12345`) and extracts the numeric ID. This ID is used as the path segment in all subsequent Partner API calls.
- **Manual token path requires ADMIN role.** Only users with ADMIN or OWNER role can upload tokens directly. This is enforced by the `RequireRoles` middleware on the manual token routes.

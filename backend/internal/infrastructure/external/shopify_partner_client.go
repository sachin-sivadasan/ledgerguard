package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Rate limiting errors
var (
	ErrRateLimited      = errors.New("rate limited by Shopify Partner API")
	ErrMaxRetriesExceed = errors.New("max retries exceeded for Shopify Partner API request")
)

// PartnerApp represents a Shopify app from the Partner API
type PartnerApp struct {
	ID   string `json:"id"`   // Shopify GID (e.g., "gid://partners/App/12345")
	Name string `json:"name"` // App name
}

// RateLimiterConfig configures rate limiting behavior
type RateLimiterConfig struct {
	RequestsPerSecond float64       // Target requests per second (default: 4)
	BurstSize         int           // Burst capacity (default: 4)
	MaxRetries        int           // Max retry attempts (default: 3)
	BaseBackoff       time.Duration // Base backoff duration (default: 1s)
	MaxBackoff        time.Duration // Max backoff duration (default: 30s)
}

// DefaultRateLimiterConfig returns sensible defaults for Shopify Partner API
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 4,                // Shopify's documented rate
		BurstSize:         4,                // Allow small bursts
		MaxRetries:        3,                // Retry up to 3 times
		BaseBackoff:       time.Second,      // Start with 1s backoff
		MaxBackoff:        30 * time.Second, // Max 30s backoff
	}
}

// tokenBucket implements a simple token bucket rate limiter
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newTokenBucket(tokensPerSecond float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: tokensPerSecond,
		lastRefill: time.Now(),
	}
}

// wait blocks until a token is available or context is cancelled
func (tb *tokenBucket) wait(ctx context.Context) error {
	for {
		tb.mu.Lock()
		tb.refill()

		if tb.tokens >= 1 {
			tb.tokens--
			tb.mu.Unlock()
			return nil
		}

		// Calculate wait time until next token
		waitTime := time.Duration(float64(time.Second) / tb.refillRate)
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to try again
		}
	}
}

func (tb *tokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = math.Min(tb.maxTokens, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now
}

// partnerAPIVersion is the Shopify Partner API version used for all GraphQL calls
// (single source of truth). 2025-07 fell out of support (Shopify returned
// {"errors":[{"message":"Invalid API version"}]} with a 404), so this must track a
// currently-supported version — 2026-04 is the current stable per shopify.dev/docs/api/partner.
// Bump DELIBERATELY and re-validate the queries against the new schema. See future.md.
const partnerAPIVersion = "2026-04"

// ShopifyPartnerClient handles communication with Shopify Partner API
type ShopifyPartnerClient struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *tokenBucket            // Deprecated: single limiter for backward compat
	limiters    map[string]*tokenBucket // Per-partner rate limiters keyed by org ID
	limiterMu   sync.RWMutex            // Protects limiters map
	config      RateLimiterConfig
}

// ShopifyPartnerClientOption is a functional option for configuring the client
type ShopifyPartnerClientOption func(*ShopifyPartnerClient)

// WithRateLimiterConfig sets custom rate limiter configuration
func WithRateLimiterConfig(config RateLimiterConfig) ShopifyPartnerClientOption {
	return func(c *ShopifyPartnerClient) {
		c.config = config
		c.rateLimiter = newTokenBucket(config.RequestsPerSecond, config.BurstSize)
	}
}

// WithBaseURL overrides the Partner API base URL (for mock server).
func WithBaseURL(url string) ShopifyPartnerClientOption {
	return func(c *ShopifyPartnerClient) {
		c.baseURL = url
		log.Printf("Shopify Partner API base URL overridden: %s", url)
	}
}

// WithRequestsPerSecond sets the rate limit for Partner API calls.
// This is used for per-partner rate limiting (each partner gets their own limiter).
func WithRequestsPerSecond(rps float64) ShopifyPartnerClientOption {
	return func(c *ShopifyPartnerClient) {
		if rps > 0 {
			c.config.RequestsPerSecond = rps
			c.config.BurstSize = int(rps) // Burst size equals RPS
			if c.config.BurstSize < 1 {
				c.config.BurstSize = 1
			}
			// Update global limiter for backward compatibility
			c.rateLimiter = newTokenBucket(rps, c.config.BurstSize)
			log.Printf("Shopify Partner API rate limit configured: %.1f RPS", rps)
		}
	}
}

// NewShopifyPartnerClient creates a new client with rate limiting
func NewShopifyPartnerClient(opts ...ShopifyPartnerClientOption) *ShopifyPartnerClient {
	config := DefaultRateLimiterConfig()
	c := &ShopifyPartnerClient{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://partners.shopify.com",
		rateLimiter: newTokenBucket(config.RequestsPerSecond, config.BurstSize),
		limiters:    make(map[string]*tokenBucket),
		config:      config,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// getLimiterForPartner returns the rate limiter for a specific partner (org ID).
// Creates a new limiter if one doesn't exist for this partner.
func (c *ShopifyPartnerClient) getLimiterForPartner(orgID string) *tokenBucket {
	// Fast path: read lock
	c.limiterMu.RLock()
	if limiter, ok := c.limiters[orgID]; ok {
		c.limiterMu.RUnlock()
		return limiter
	}
	c.limiterMu.RUnlock()

	// Slow path: write lock to create new limiter
	c.limiterMu.Lock()
	defer c.limiterMu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok := c.limiters[orgID]; ok {
		return limiter
	}

	// Create new limiter for this partner
	limiter := newTokenBucket(c.config.RequestsPerSecond, c.config.BurstSize)
	c.limiters[orgID] = limiter
	log.Printf("Created rate limiter for partner %s: %.1f RPS, burst %d", orgID, c.config.RequestsPerSecond, c.config.BurstSize)
	return limiter
}

// executeWithRetry executes an HTTP request with rate limiting and exponential backoff
func (c *ShopifyPartnerClient) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	var lastErr error

	// Use default config if not initialized (for backward compatibility in tests)
	maxRetries := c.config.MaxRetries
	if c.rateLimiter == nil && maxRetries == 0 {
		maxRetries = 0 // No retries when rate limiter not initialized
	}

	// Get per-partner rate limiter if org ID is in context
	var limiter *tokenBucket
	if orgID := c.getOrganizationID(ctx); orgID != "" && c.limiters != nil {
		limiter = c.getLimiterForPartner(orgID)
	} else {
		// Fallback to global limiter for backward compatibility
		limiter = c.rateLimiter
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait for rate limiter (skip if not initialized)
		if limiter != nil {
			if err := limiter.wait(ctx); err != nil {
				return nil, nil, err
			}
		}

		// Execute request. Clone per attempt AND rewind the body: the body is a
		// bytes.Reader that's consumed on the first Do, so a retry (after a 429/5xx backoff)
		// would otherwise POST an EMPTY body — which Cloudflare rejects with 400 Bad Request.
		// GetBody (set by http.NewRequest for a bytes.Reader body) returns a fresh reader.
		attemptReq := req.Clone(ctx)
		if req.GetBody != nil {
			freshBody, gerr := req.GetBody()
			if gerr != nil {
				return nil, nil, fmt.Errorf("failed to rewind request body: %w", gerr)
			}
			attemptReq.Body = freshBody
		}
		resp, err := c.httpClient.Do(attemptReq)
		if err != nil {
			lastErr = err
			log.Printf("Shopify API request failed (attempt %d/%d): %v", attempt+1, c.config.MaxRetries+1, err)
			if attempt < c.config.MaxRetries {
				c.backoff(ctx, attempt)
				continue
			}
			break
		}

		// Read body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Check for rate limiting (429) or server errors (5xx)
		if resp.StatusCode == http.StatusTooManyRequests {
			log.Printf("Shopify API rate limited (attempt %d/%d), backing off", attempt+1, c.config.MaxRetries+1)
			lastErr = ErrRateLimited
			if attempt < c.config.MaxRetries {
				// Check for Retry-After header
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
						time.Sleep(seconds)
						continue
					}
				}
				c.backoff(ctx, attempt)
				continue
			}
			break
		}

		if resp.StatusCode >= 500 {
			log.Printf("Shopify API server error %d (attempt %d/%d)", resp.StatusCode, attempt+1, c.config.MaxRetries+1)
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			if attempt < c.config.MaxRetries {
				c.backoff(ctx, attempt)
				continue
			}
			break
		}

		// Success or client error (don't retry 4xx except 429)
		return resp, body, nil
	}

	if lastErr != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrMaxRetriesExceed, lastErr)
	}
	return nil, nil, ErrMaxRetriesExceed
}

// backoff performs exponential backoff with jitter
func (c *ShopifyPartnerClient) backoff(ctx context.Context, attempt int) {
	// Exponential backoff: base * 2^attempt
	backoff := c.config.BaseBackoff * time.Duration(1<<uint(attempt))
	if backoff > c.config.MaxBackoff {
		backoff = c.config.MaxBackoff
	}

	// Add jitter (±25%)
	jitter := time.Duration(float64(backoff) * 0.25 * (0.5 - float64(time.Now().UnixNano()%100)/100))
	backoff += jitter

	log.Printf("Backing off for %v before retry", backoff)

	select {
	case <-ctx.Done():
	case <-time.After(backoff):
	}
}

// isRateLimitError checks if an error indicates rate limiting
func isRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimited) ||
		strings.Contains(err.Error(), "429") ||
		strings.Contains(err.Error(), "rate limit")
}

// FetchApps retrieves all apps for the given partner organization
func (c *ShopifyPartnerClient) FetchApps(ctx context.Context, organizationID, accessToken string) ([]PartnerApp, error) {
	// Fetch transactions and extract apps from AppSubscriptionSale
	query := `
		query {
			transactions(first: 100) {
				edges {
					node {
						id
						... on AppSubscriptionSale {
							app {
								id
								name
							}
						}
						... on AppUsageSale {
							app {
								id
								name
							}
						}
						... on AppOneTimeSale {
							app {
								id
								name
							}
						}
					}
				}
			}
		}
	`
	_ = organizationID

	url := fmt.Sprintf("%s/%s/api/%s/graphql.json", c.baseURL, organizationID, partnerAPIVersion)
	log.Printf("Fetching apps from: %s (org: %s, token length: %d)", url, organizationID, len(accessToken))

	reqBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", accessToken)

	// Use rate-limited execution with retry
	resp, body, err := c.executeWithRetry(ctx, req)
	if err != nil {
		log.Printf("Partner API request failed: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	log.Printf("Partner API response - Status: %d, Body length: %d", resp.StatusCode, len(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Transactions struct {
				Edges []struct {
					Node struct {
						ID  string `json:"id"`
						App *struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"app,omitempty"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"transactions"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	// Extract unique apps from transactions
	appMap := make(map[string]PartnerApp)
	for _, edge := range result.Data.Transactions.Edges {
		if edge.Node.App != nil && edge.Node.App.ID != "" {
			appMap[edge.Node.App.ID] = PartnerApp{
				ID:   edge.Node.App.ID,
				Name: edge.Node.App.Name,
			}
		}
	}

	apps := make([]PartnerApp, 0, len(appMap))
	for _, app := range appMap {
		apps = append(apps, app)
	}

	log.Printf("Found %d unique apps from transactions", len(apps))
	return apps, nil
}

// FetchTransactions retrieves transactions from the Shopify Partner API for a given app
// within the specified date range. Handles pagination automatically.
func (c *ShopifyPartnerClient) FetchTransactions(
	ctx context.Context,
	accessToken string,
	appID uuid.UUID,
	from, to time.Time,
) ([]*entity.Transaction, error) {
	// Get organization ID from the client context or use a default
	// In production, this would come from the partner account
	organizationID := c.getOrganizationID(ctx)
	if organizationID == "" {
		return nil, fmt.Errorf("organization ID not set")
	}

	var allTransactions []*entity.Transaction
	var cursor string
	hasNextPage := true

	for hasNextPage {
		transactions, nextCursor, more, err := c.fetchTransactionPage(
			ctx, organizationID, accessToken, appID, from, to, cursor,
		)
		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)
		cursor = nextCursor
		hasNextPage = more

		log.Printf("Fetched %d transactions (total: %d, hasMore: %v)",
			len(transactions), len(allTransactions), hasNextPage)
	}

	log.Printf("Total transactions fetched: %d for app %s", len(allTransactions), appID)
	return allTransactions, nil
}

// fetchTransactionPage fetches a single page of transactions
func (c *ShopifyPartnerClient) fetchTransactionPage(
	ctx context.Context,
	organizationID, accessToken string,
	appID uuid.UUID,
	from, to time.Time,
	cursor string,
) ([]*entity.Transaction, string, bool, error) {
	// Build the GraphQL query with pagination and date filtering
	// Includes all required fields per Shopify Partner API documentation
	query := `
		query($first: Int!, $after: String, $createdAtMin: DateTime!, $createdAtMax: DateTime!) {
			transactions(first: $first, after: $after, createdAtMin: $createdAtMin, createdAtMax: $createdAtMax) {
				edges {
					cursor
					node {
						__typename
						id
						createdAt
						... on AppSubscriptionSale {
							chargeId
							app { id name }
							shop {
								id
								myshopifyDomain
								name
							}
							grossAmount { amount currencyCode }
							netAmount { amount currencyCode }
						}
						... on AppUsageSale {
							chargeId
							app { id name }
							shop {
								id
								myshopifyDomain
								name
							}
							grossAmount { amount currencyCode }
							netAmount { amount currencyCode }
						}
						... on AppOneTimeSale {
							chargeId
							app { id name }
							shop {
								id
								myshopifyDomain
								name
							}
							grossAmount { amount currencyCode }
							netAmount { amount currencyCode }
						}
						... on AppSaleAdjustment {
							chargeId
							app { id name }
							shop {
								id
								myshopifyDomain
								name
							}
							grossAmount { amount currencyCode }
							netAmount { amount currencyCode }
						}
					}
				}
				pageInfo {
					hasNextPage
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first":        100,
		"createdAtMin": from.Format(time.RFC3339),
		"createdAtMax": to.Format(time.RFC3339),
	}
	if cursor != "" {
		variables["after"] = cursor
	}
	// Pass app GID so the API returns only this app's transactions
	if appGID := c.getPartnerAppGID(ctx); appGID != "" {
		variables["appId"] = appGID
	}

	url := fmt.Sprintf("%s/%s/api/%s/graphql.json", c.baseURL, organizationID, partnerAPIVersion)

	reqBody, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", accessToken)

	// Use rate-limited execution with retry
	resp, body, err := c.executeWithRetry(ctx, req)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result transactionsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "", false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, "", false, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	// Convert to domain entities
	var transactions []*entity.Transaction
	var lastCursor string

	for _, edge := range result.Data.Transactions.Edges {
		lastCursor = edge.Cursor
		tx := c.parseTransaction(edge.Node, appID)
		if tx != nil {
			transactions = append(transactions, tx)
		}
	}

	return transactions, lastCursor, result.Data.Transactions.PageInfo.HasNextPage, nil
}

// transactionsResponse represents the GraphQL response structure
type transactionsResponse struct {
	Data struct {
		Transactions struct {
			Edges []struct {
				Cursor string          `json:"cursor"`
				Node   transactionNode `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"transactions"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// transactionNode represents a transaction from the Partner API
type transactionNode struct {
	Typename  string `json:"__typename"`
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	ChargeID  string `json:"chargeId,omitempty"`
	App       *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"app,omitempty"`
	Shop *struct {
		ID              string `json:"id"`
		MyshopifyDomain string `json:"myshopifyDomain"`
		Name            string `json:"name"`
	} `json:"shop,omitempty"`
	GrossAmount *struct {
		Amount       string `json:"amount"`
		CurrencyCode string `json:"currencyCode"`
	} `json:"grossAmount,omitempty"`
	NetAmount *struct {
		Amount       string `json:"amount"`
		CurrencyCode string `json:"currencyCode"`
	} `json:"netAmount,omitempty"`
}

// parseTransaction converts a Partner API transaction to a domain entity
func (c *ShopifyPartnerClient) parseTransaction(node transactionNode, appID uuid.UUID) *entity.Transaction {
	if node.App == nil {
		return nil
	}
	// Shop can be nil for ReferralTransaction
	shopDomain := ""
	shopName := ""
	shopGID := ""
	if node.Shop != nil {
		shopDomain = node.Shop.MyshopifyDomain
		shopName = node.Shop.Name
		shopGID = node.Shop.ID
	}

	// Determine charge type based on transaction type (inferred from fields present)
	chargeType := c.inferChargeType(node)

	// Get both amounts - gross (subscription price) and net (revenue)
	grossCents, netCents, currency := c.parseAmounts(node)

	// Parse transaction date
	transactionDate, err := time.Parse(time.RFC3339, node.CreatedAt)
	if err != nil {
		log.Printf("Failed to parse transaction date %s: %v", node.CreatedAt, err)
		transactionDate = time.Now()
	}

	tx := entity.NewTransaction(
		appID,
		node.ID,
		shopDomain,
		shopName,
		chargeType,
		grossCents,
		netCents,
		currency,
		transactionDate,
	)

	// Add shop details and source app GID for per-app filtering
	tx.ShopifyShopGID = shopGID
	tx.PartnerAppGID = node.App.ID

	// chargeId for AppSubscriptionSale is the subscription GID (gid://shopify/AppSubscription/...)
	if node.ChargeID != "" {
		tx.SubscriptionGID = node.ChargeID
	}

	return tx
}

// inferChargeType determines the charge type based on GraphQL __typename
func (c *ShopifyPartnerClient) inferChargeType(node transactionNode) valueobject.ChargeType {
	switch node.Typename {
	case "AppSubscriptionSale":
		return valueobject.ChargeTypeRecurring
	case "AppUsageSale":
		return valueobject.ChargeTypeUsage
	case "AppOneTimeSale":
		return valueobject.ChargeTypeOneTime
	case "AppSaleAdjustment":
		// Refund/downgrade/chargeback of an app charge. Its netAmount is NEGATIVE (deducted
		// from payout) and is stored as-is — the truthful signed effect on the payout, so
		// any SUM(net) nets it out. Consumers that display refunds as a positive magnitude
		// negate locally (see revenue_mix / metrics_engine / earnings resolver). NOTE: only
		// AppSaleAdjustment is fetched (fragment above); AppCredit would need its own
		// verified fragment before it can be ingested, so it's intentionally not mapped here.
		return valueobject.ChargeTypeRefund
	default:
		return valueobject.ChargeTypeRecurring
	}
}

// parseAmounts extracts both gross and net amounts in cents and currency from the transaction
// - grossAmount: Subscription price (what customer pays)
// - netAmount: Revenue (what you receive after Shopify's cut)
func (c *ShopifyPartnerClient) parseAmounts(node transactionNode) (grossCents, netCents int64, currency string) {
	currency = "USD"

	if node.GrossAmount != nil {
		var dollars float64
		fmt.Sscanf(node.GrossAmount.Amount, "%f", &dollars)
		grossCents = int64(dollars * 100)
		currency = node.GrossAmount.CurrencyCode
	}

	if node.NetAmount != nil {
		var dollars float64
		fmt.Sscanf(node.NetAmount.Amount, "%f", &dollars)
		netCents = int64(dollars * 100)
		if currency == "USD" && node.NetAmount.CurrencyCode != "" {
			currency = node.NetAmount.CurrencyCode
		}
	}

	return grossCents, netCents, currency
}

// getOrganizationID retrieves the organization ID from context
// This should be set when creating requests
func (c *ShopifyPartnerClient) getOrganizationID(ctx context.Context) string {
	if orgID, ok := ctx.Value(organizationIDKey).(string); ok {
		return orgID
	}
	return ""
}

// organizationIDKey is the context key for organization ID
type contextKey string

const organizationIDKey contextKey = "organizationID"
const partnerAppGIDKey contextKey = "partnerAppGID"

// WithOrganizationID returns a new context with the organization ID set
func WithOrganizationID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, organizationIDKey, orgID)
}

// WithPartnerAppGID returns a new context with the partner app GID set
// (e.g. "gid://partners/App/2001") so the API query can filter by app
func WithPartnerAppGID(ctx context.Context, gid string) context.Context {
	return context.WithValue(ctx, partnerAppGIDKey, gid)
}

// getPartnerAppGID retrieves the partner app GID from context
func (c *ShopifyPartnerClient) getPartnerAppGID(ctx context.Context) string {
	if gid, ok := ctx.Value(partnerAppGIDKey).(string); ok {
		return gid
	}
	return ""
}

// AppEvent represents an app lifecycle event from the Partner API
type AppEvent struct {
	Type       string // one of Partner API AppEventTypes, e.g. RELATIONSHIP_INSTALLED/REACTIVATED/UNINSTALLED/DEACTIVATED, SUBSCRIPTION_CHARGE_ACTIVATED/ACCEPTED/CANCELED/EXPIRED/FROZEN/UNFROZEN/DECLINED (see GetLatestSubscriptionStatus)
	ShopID     string
	ShopName   string
	OccurredAt time.Time
}

// FetchAppEvents retrieves lifecycle events for an app, optionally filtered by shop.
// Events span the full Partner API AppEventTypes enum (relationship + subscription-charge
// lifecycle); GetLatestSubscriptionStatus interprets the status-relevant ones.
func (c *ShopifyPartnerClient) FetchAppEvents(
	ctx context.Context,
	organizationID, accessToken string,
	appGID string,
	shopGID string, // Optional: filter by shop
) ([]AppEvent, error) {
	// App.events has no sortKey/reverse (verified against Partner API 2026-04), and the
	// Partner API rejects `last` without a `before` cursor ("Using last without before is
	// not supported"). So forward-paginate with first/after and collect EVERY event for the
	// shop — completeness matters (installs, subscription-charge accepts, and the most
	// recent status-deciding events for busy shops). GetLatestSubscriptionStatus sorts by
	// OccurredAt desc, so page order doesn't matter.
	// The Partner API's PageInfo has NO endCursor (unlike Admin/Storefront) — the next
	// cursor is the LAST edge's `cursor`, and pageInfo only exposes hasNextPage. (Same
	// pattern as FetchTransactions.)
	const query = `
		query($appId: ID!, $shopId: ID, $cursor: String) {
			app(id: $appId) {
				events(shopId: $shopId, first: 100, after: $cursor) {
					edges {
						cursor
						node {
							type
							occurredAt
							shop {
								id
								name
							}
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}
		}
	`

	url := fmt.Sprintf("%s/%s/api/%s/graphql.json", c.baseURL, organizationID, partnerAPIVersion)

	var events []AppEvent
	cursor := ""
	// Safety bound: 1000 pages × 100 = 100k events; per-shop app-events are far fewer.
	for page := 0; page < 1000; page++ {
		variables := map[string]interface{}{
			"appId": appGID,
		}
		if shopGID != "" {
			variables["shopId"] = shopGID
		}
		if cursor != "" {
			variables["cursor"] = cursor
		}

		reqBody, err := json.Marshal(map[string]interface{}{
			"query":     query,
			"variables": variables,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Shopify-Access-Token", accessToken)

		resp, body, err := c.executeWithRetry(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Data struct {
				App struct {
					Events struct {
						Edges []struct {
							Cursor string `json:"cursor"`
							Node   struct {
								Type       string `json:"type"`
								OccurredAt string `json:"occurredAt"`
								Shop       *struct {
									ID   string `json:"id"`
									Name string `json:"name"`
								} `json:"shop"`
							} `json:"node"`
						} `json:"edges"`
						PageInfo struct {
							HasNextPage bool `json:"hasNextPage"`
						} `json:"pageInfo"`
					} `json:"events"`
				} `json:"app"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
		}

		edges := result.Data.App.Events.Edges
		lastCursor := ""
		for _, edge := range edges {
			lastCursor = edge.Cursor
			event := AppEvent{Type: edge.Node.Type}
			if edge.Node.OccurredAt != "" {
				if t, err := time.Parse(time.RFC3339, edge.Node.OccurredAt); err == nil {
					event.OccurredAt = t
				}
			}
			if edge.Node.Shop != nil {
				event.ShopID = edge.Node.Shop.ID
				event.ShopName = edge.Node.Shop.Name
			}
			events = append(events, event)
		}

		// Next cursor is the last edge's cursor (Partner API has no pageInfo.endCursor).
		if !result.Data.App.Events.PageInfo.HasNextPage || lastCursor == "" {
			break
		}
		cursor = lastCursor
	}

	return events, nil
}

// CountInstalls derives install metrics from the app-wide RELATIONSHIP event
// stream by running a per-shop state machine over the relationship events:
//   - active  = shops whose LATEST relationship event is INSTALLED or REACTIVATED
//     (currently installed; matches Shopify's Partner "active installs" / Mantle
//     "active users").
//   - total   = distinct shops that have ever installed (any INSTALLED event) —
//     the lifetime install base (Mantle "Installed").
//
// Shops are keyed by ShopID (GID), falling back to ShopName. Non-relationship
// events (subscription charges etc.) are ignored.
func CountInstalls(events []AppEvent) (active, total int) {
	type shopState struct {
		latest   time.Time
		seen     bool
		active   bool
		everInst bool
	}
	byShop := make(map[string]*shopState)
	for _, ev := range events {
		var isRel, isActive, isInstall bool
		switch ev.Type {
		case "RELATIONSHIP_INSTALLED":
			isRel, isActive, isInstall = true, true, true
		case "RELATIONSHIP_REACTIVATED":
			isRel, isActive = true, true
		case "RELATIONSHIP_UNINSTALLED", "RELATIONSHIP_DEACTIVATED":
			isRel = true // inactive
		}
		if !isRel {
			continue
		}
		key := ev.ShopID
		if key == "" {
			key = ev.ShopName
		}
		if key == "" {
			continue
		}
		s := byShop[key]
		if s == nil {
			s = &shopState{}
			byShop[key] = s
		}
		if isInstall {
			s.everInst = true
		}
		if !s.seen || ev.OccurredAt.After(s.latest) {
			s.latest = ev.OccurredAt
			s.active = isActive
			s.seen = true
		}
	}
	for _, s := range byShop {
		if s.everInst {
			total++
		}
		if s.active {
			active++
		}
	}
	return active, total
}

// GetLatestSubscriptionStatus determines subscription status from the latest
// (by OccurredAt) status-relevant app event.
// Returns: "ACTIVE", "FROZEN", "CANCELLED", "UNINSTALLED", "PENDING", or "" if unknown.
func GetLatestSubscriptionStatus(events []AppEvent) string {
	status, _ := GetLatestSubscriptionStatusWithTime(events)
	return status
}

// GetLatestSubscriptionStatusWithTime is GetLatestSubscriptionStatus plus the
// OccurredAt of the deciding event, so callers can reconcile a terminal status
// against billing reality (a recurring charge dated AFTER a CANCELLED/UNINSTALLED
// event means that event was a stale/plan-change cancel, not real churn).
func GetLatestSubscriptionStatusWithTime(events []AppEvent) (string, time.Time) {
	if len(events) == 0 {
		return "", time.Time{}
	}

	// Decide status from the LATEST event by OccurredAt — never trust the
	// slice/API order (it isn't guaranteed sorted). Tie-break: an accepted
	// charge outranks a cancel at the same timestamp so an upgrade/downgrade
	// (Shopify cancels the old subscription and accepts a new one, often at the
	// same instant) is not misread as churn — the "cancel trap".
	sorted := make([]AppEvent, len(events))
	copy(sorted, events)
	sort.SliceStable(sorted, func(i, j int) bool {
		if !sorted[i].OccurredAt.Equal(sorted[j].OccurredAt) {
			return sorted[i].OccurredAt.After(sorted[j].OccurredAt)
		}
		return statusEventPriority(sorted[i].Type) > statusEventPriority(sorted[j].Type)
	})

	for _, event := range sorted {
		switch event.Type {
		case "RELATIONSHIP_UNINSTALLED", "RELATIONSHIP_DEACTIVATED":
			// Uninstalled, or the relationship was deactivated (store closed/frozen
			// by Shopify) — terminal churn.
			return "UNINSTALLED", event.OccurredAt
		case "SUBSCRIPTION_CHARGE_CANCELED", "SUBSCRIPTION_CHARGE_EXPIRED":
			// Cancelled, or the charge expired (approval window lapsed / not renewed).
			return "CANCELLED", event.OccurredAt
		case "SUBSCRIPTION_CHARGE_FROZEN", "SUBSCRIPTION_CHARGE_DECLINED":
			// Frozen, or a payment was declined — a recoverable payment issue, not
			// terminal churn.
			return "FROZEN", event.OccurredAt
		case "SUBSCRIPTION_CHARGE_ACTIVATED", "SUBSCRIPTION_CHARGE_UNFROZEN", "SUBSCRIPTION_CHARGE_ACCEPTED":
			// Activated = a recurring charge went live (the event monthly renewals
			// emit — NOT ACCEPTED, which only fires at initial approval); unfrozen
			// restores billing; accepted (re)starts it — all ACTIVE.
			return "ACTIVE", event.OccurredAt
		case "RELATIONSHIP_INSTALLED", "RELATIONSHIP_REACTIVATED":
			// Installed or reinstalled (reactivated) but no subscription-charge event
			// yet — trial or pending. A later charge event, if any, outranks this.
			return "PENDING", event.OccurredAt
		}
	}

	return "", time.Time{}
}

// statusEventPriority ranks status-relevant events for same-timestamp tie-breaks.
// Higher wins. Active signals (accepted / unfrozen) outrank a cancel so a
// same-instant upgrade/downgrade (Shopify cancels the old sub and accepts a new
// one) reads as ACTIVE, not churn.
func statusEventPriority(eventType string) int {
	switch eventType {
	case "SUBSCRIPTION_CHARGE_ACTIVATED":
		return 7 // recurring charge live — strongest active signal
	case "SUBSCRIPTION_CHARGE_ACCEPTED":
		return 6
	case "SUBSCRIPTION_CHARGE_UNFROZEN":
		return 5
	case "RELATIONSHIP_REACTIVATED":
		return 4
	case "RELATIONSHIP_INSTALLED":
		return 3
	case "SUBSCRIPTION_CHARGE_FROZEN", "SUBSCRIPTION_CHARGE_DECLINED":
		return 2
	case "SUBSCRIPTION_CHARGE_CANCELED", "SUBSCRIPTION_CHARGE_EXPIRED":
		return 1
	case "RELATIONSHIP_UNINSTALLED", "RELATIONSHIP_DEACTIVATED":
		return 0
	default:
		return -1
	}
}

// FetchInstallCount retrieves the number of shops that have installed the app
// by counting RELATIONSHIP_INSTALLED events with pagination
func (c *ShopifyPartnerClient) FetchInstallCount(
	ctx context.Context,
	organizationID, accessToken, partnerAppID string,
) (int, error) {
	// Use a set to track unique shop IDs
	installedShops := make(map[string]bool)
	uninstalledShops := make(map[string]bool)
	var cursor string
	hasNextPage := true

	for hasNextPage {
		count, installs, uninstalls, nextCursor, more, err := c.fetchInstallEventPage(
			ctx, organizationID, accessToken, partnerAppID, cursor,
		)
		if err != nil {
			return 0, err
		}

		// Track installed shops
		for _, shopID := range installs {
			installedShops[shopID] = true
		}

		// Track uninstalled shops
		for _, shopID := range uninstalls {
			uninstalledShops[shopID] = true
		}

		cursor = nextCursor
		hasNextPage = more

		log.Printf("Fetched %d install events (total installs: %d, uninstalls: %d, hasMore: %v)",
			count, len(installedShops), len(uninstalledShops), hasNextPage)
	}

	// Calculate current installs = installed - uninstalled
	currentInstalls := 0
	for shopID := range installedShops {
		if !uninstalledShops[shopID] {
			currentInstalls++
		}
	}

	log.Printf("Total current installs: %d (installed: %d, uninstalled: %d)",
		currentInstalls, len(installedShops), len(uninstalledShops))

	return currentInstalls, nil
}

// fetchInstallEventPage fetches a single page of install/uninstall events
func (c *ShopifyPartnerClient) fetchInstallEventPage(
	ctx context.Context,
	organizationID, accessToken, partnerAppID, cursor string,
) (int, []string, []string, string, bool, error) {
	// Query for both RELATIONSHIP_INSTALLED and RELATIONSHIP_UNINSTALLED events
	query := `
		query($appId: ID!, $first: Int!, $after: String) {
			app(id: $appId) {
				events(
					first: $first
					after: $after
					types: [RELATIONSHIP_INSTALLED, RELATIONSHIP_UNINSTALLED]
				) {
					edges {
						cursor
						node {
							type
							shop {
								id
							}
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"appId": partnerAppID,
		"first": 100,
	}
	if cursor != "" {
		variables["after"] = cursor
	}

	url := fmt.Sprintf("%s/%s/api/%s/graphql.json", c.baseURL, organizationID, partnerAPIVersion)

	reqBody, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return 0, nil, nil, "", false, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return 0, nil, nil, "", false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", accessToken)

	resp, body, err := c.executeWithRetry(ctx, req)
	if err != nil {
		return 0, nil, nil, "", false, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, nil, nil, "", false, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			App struct {
				Events struct {
					Edges []struct {
						Cursor string `json:"cursor"`
						Node   struct {
							Type string `json:"type"`
							Shop *struct {
								ID string `json:"id"`
							} `json:"shop"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage bool `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"events"`
			} `json:"app"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, nil, nil, "", false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return 0, nil, nil, "", false, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	var installs, uninstalls []string
	var lastCursor string

	for _, edge := range result.Data.App.Events.Edges {
		lastCursor = edge.Cursor
		if edge.Node.Shop == nil {
			continue
		}

		switch edge.Node.Type {
		case "RELATIONSHIP_INSTALLED":
			installs = append(installs, edge.Node.Shop.ID)
		case "RELATIONSHIP_UNINSTALLED":
			uninstalls = append(uninstalls, edge.Node.Shop.ID)
		}
	}

	return len(result.Data.App.Events.Edges), installs, uninstalls, lastCursor, result.Data.App.Events.PageInfo.HasNextPage, nil
}

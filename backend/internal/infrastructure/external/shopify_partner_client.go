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
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Rate limiting errors
var (
	ErrRateLimited       = errors.New("rate limited by Shopify Partner API")
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
		RequestsPerSecond: 4,              // Shopify's documented rate
		BurstSize:         4,              // Allow small bursts
		MaxRetries:        3,              // Retry up to 3 times
		BaseBackoff:       time.Second,    // Start with 1s backoff
		MaxBackoff:        30 * time.Second, // Max 30s backoff
	}
}

// tokenBucket implements a simple token bucket rate limiter
type tokenBucket struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastRefill   time.Time
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

// ShopifyPartnerClient handles communication with Shopify Partner API
type ShopifyPartnerClient struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *tokenBucket
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

// NewShopifyPartnerClient creates a new client with rate limiting
func NewShopifyPartnerClient(opts ...ShopifyPartnerClientOption) *ShopifyPartnerClient {
	config := DefaultRateLimiterConfig()
	c := &ShopifyPartnerClient{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     "https://partners.shopify.com",
		rateLimiter: newTokenBucket(config.RequestsPerSecond, config.BurstSize),
		config:      config,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// executeWithRetry executes an HTTP request with rate limiting and exponential backoff
func (c *ShopifyPartnerClient) executeWithRetry(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	var lastErr error

	// Use default config if not initialized (for backward compatibility in tests)
	maxRetries := c.config.MaxRetries
	if c.rateLimiter == nil && maxRetries == 0 {
		maxRetries = 0 // No retries when rate limiter not initialized
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait for rate limiter (skip if not initialized)
		if c.rateLimiter != nil {
			if err := c.rateLimiter.wait(ctx); err != nil {
				return nil, nil, err
			}
		}

		// Execute request
		resp, err := c.httpClient.Do(req.Clone(ctx))
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

	url := fmt.Sprintf("%s/%s/api/2025-07/graphql.json", c.baseURL, organizationID)
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
								plan { displayName }
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
								plan { displayName }
							}
							grossAmount { amount currencyCode }
							netAmount { amount currencyCode }
							appUsageRecord {
								id
								description
							}
						}
						... on AppOneTimeSale {
							chargeId
							app { id name }
							shop {
								id
								myshopifyDomain
								name
								plan { displayName }
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

	url := fmt.Sprintf("%s/%s/api/2025-07/graphql.json", c.baseURL, organizationID)

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
		Plan            *struct {
			DisplayName string `json:"displayName"`
		} `json:"plan,omitempty"`
	} `json:"shop,omitempty"`
	GrossAmount *struct {
		Amount       string `json:"amount"`
		CurrencyCode string `json:"currencyCode"`
	} `json:"grossAmount,omitempty"`
	NetAmount *struct {
		Amount       string `json:"amount"`
		CurrencyCode string `json:"currencyCode"`
	} `json:"netAmount,omitempty"`
	// AppUsageSale specific fields
	AppUsageRecord *struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	} `json:"appUsageRecord,omitempty"`
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
	shopPlan := ""
	if node.Shop != nil {
		shopDomain = node.Shop.MyshopifyDomain
		shopName = node.Shop.Name
		shopGID = node.Shop.ID
		if node.Shop.Plan != nil {
			shopPlan = node.Shop.Plan.DisplayName
		}
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

	// Add shop details
	tx.ShopifyShopGID = shopGID
	tx.ShopPlan = shopPlan

	// Note: Subscription status/details are not available from transactions query.
	// Use FetchAppEvents to get subscription lifecycle events (SUBSCRIPTION_CHARGE_ACCEPTED,
	// SUBSCRIPTION_CHARGE_CANCELED, RELATIONSHIP_INSTALLED, RELATIONSHIP_UNINSTALLED)

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
	case "AppCredit":
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

// WithOrganizationID returns a new context with the organization ID set
func WithOrganizationID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, organizationIDKey, orgID)
}

// AppEvent represents an app lifecycle event from the Partner API
type AppEvent struct {
	Type      string // RELATIONSHIP_INSTALLED, SUBSCRIPTION_CHARGE_ACCEPTED, SUBSCRIPTION_CHARGE_CANCELED, RELATIONSHIP_UNINSTALLED
	ShopID    string
	ShopName  string
	OccurredAt time.Time
}

// FetchAppEvents retrieves lifecycle events for an app, optionally filtered by shop
// Events include: RELATIONSHIP_INSTALLED, SUBSCRIPTION_CHARGE_ACCEPTED,
// SUBSCRIPTION_CHARGE_CANCELED, RELATIONSHIP_UNINSTALLED
func (c *ShopifyPartnerClient) FetchAppEvents(
	ctx context.Context,
	organizationID, accessToken string,
	appGID string,
	shopGID string, // Optional: filter by shop
) ([]AppEvent, error) {
	// Build query with optional shop filter
	query := `
		query($appId: ID!, $shopId: ID) {
			app(id: $appId) {
				events(shopId: $shopId, first: 50) {
					edges {
						node {
							type
							occurredAt
							shop {
								id
								name
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"appId": appGID,
	}
	if shopGID != "" {
		variables["shopId"] = shopGID
	}

	url := fmt.Sprintf("%s/%s/api/2025-07/graphql.json", c.baseURL, organizationID)

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
						Node struct {
							Type       string `json:"type"`
							OccurredAt string `json:"occurredAt"`
							Shop       *struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"shop"`
						} `json:"node"`
					} `json:"edges"`
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

	var events []AppEvent
	for _, edge := range result.Data.App.Events.Edges {
		event := AppEvent{
			Type: edge.Node.Type,
		}

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

	return events, nil
}

// GetLatestSubscriptionStatus determines subscription status from events
// Returns: "ACTIVE", "CANCELLED", "UNINSTALLED", or empty if unknown
func GetLatestSubscriptionStatus(events []AppEvent) string {
	if len(events) == 0 {
		return ""
	}

	// Events are typically returned newest-first
	// Look at the most recent relevant event
	for _, event := range events {
		switch event.Type {
		case "RELATIONSHIP_UNINSTALLED":
			return "UNINSTALLED"
		case "SUBSCRIPTION_CHARGE_CANCELED":
			return "CANCELLED"
		case "SUBSCRIPTION_CHARGE_ACCEPTED":
			return "ACTIVE"
		case "RELATIONSHIP_INSTALLED":
			// Installed but no subscription event yet - could be trial or pending
			return "PENDING"
		}
	}

	return ""
}

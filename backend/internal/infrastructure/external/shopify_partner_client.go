package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// PartnerApp represents a Shopify app from the Partner API
type PartnerApp struct {
	ID   string `json:"id"`   // Shopify GID (e.g., "gid://partners/App/12345")
	Name string `json:"name"` // App name
}

// ShopifyPartnerClient handles communication with Shopify Partner API
type ShopifyPartnerClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewShopifyPartnerClient() *ShopifyPartnerClient {
	return &ShopifyPartnerClient{
		httpClient: &http.Client{},
		baseURL:    "https://partners.shopify.com",
	}
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Partner API request failed: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read body for logging
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Partner API response - Status: %d, Body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Re-create reader for JSON decoding
	resp.Body = io.NopCloser(bytes.NewReader(body))

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

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	query := `
		query($first: Int!, $after: String, $createdAtMin: DateTime!, $createdAtMax: DateTime!) {
			transactions(first: $first, after: $after, createdAtMin: $createdAtMin, createdAtMax: $createdAtMax) {
				edges {
					cursor
					node {
						id
						createdAt
						... on AppSubscriptionSale {
							chargeId
							app { id name }
							shop { myshopifyDomain }
							netAmount { amount currencyCode }
						}
						... on AppUsageSale {
							chargeId
							app { id name }
							shop { myshopifyDomain }
							netAmount { amount currencyCode }
						}
						... on AppOneTimeSale {
							chargeId
							app { id name }
							shop { myshopifyDomain }
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

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
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	ChargeID  string `json:"chargeId,omitempty"`
	App       *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"app,omitempty"`
	Shop *struct {
		MyshopifyDomain string `json:"myshopifyDomain"`
	} `json:"shop,omitempty"`
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
	if node.Shop != nil {
		shopDomain = node.Shop.MyshopifyDomain
	}

	// Determine charge type based on transaction type (inferred from fields present)
	chargeType := c.inferChargeType(node)

	// Get amount - either netAmount or amount (for credits)
	amountCents, currency := c.parseAmount(node)

	// Parse transaction date
	transactionDate, err := time.Parse(time.RFC3339, node.CreatedAt)
	if err != nil {
		log.Printf("Failed to parse transaction date %s: %v", node.CreatedAt, err)
		transactionDate = time.Now()
	}

	return entity.NewTransaction(
		appID,
		node.ID,
		shopDomain,
		chargeType,
		amountCents,
		currency,
		transactionDate,
	)
}

// inferChargeType determines the charge type based on transaction characteristics
func (c *ShopifyPartnerClient) inferChargeType(node transactionNode) valueobject.ChargeType {
	// Check if it's a one-time charge (no chargeId typically means usage or subscription)
	// The actual type would be determined by the GraphQL typename, but since we're
	// using inline fragments, we infer from context

	// For now, default to RECURRING as the most common case
	// In practice, you'd want to track the __typename or use separate queries
	if node.ChargeID != "" {
		// Has a chargeId - could be subscription, usage, or one-time
		// We'll default to recurring for subscription-like charges
		return valueobject.ChargeTypeRecurring
	}

	return valueobject.ChargeTypeRecurring
}

// parseAmount extracts amount in cents and currency from the transaction
func (c *ShopifyPartnerClient) parseAmount(node transactionNode) (int64, string) {
	if node.NetAmount == nil {
		return 0, "USD"
	}

	amountStr := node.NetAmount.Amount
	currency := node.NetAmount.CurrencyCode

	// Parse amount string to cents (Shopify returns decimal strings like "10.50")
	var dollars float64
	fmt.Sscanf(amountStr, "%f", &dollars)
	cents := int64(dollars * 100)

	return cents, currency
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

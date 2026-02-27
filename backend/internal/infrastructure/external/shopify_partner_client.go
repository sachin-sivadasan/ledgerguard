package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

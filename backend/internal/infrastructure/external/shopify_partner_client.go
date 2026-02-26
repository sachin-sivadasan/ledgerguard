package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	query := `
		query {
			apps(first: 100) {
				edges {
					node {
						id
						name
					}
				}
			}
		}
	`

	url := fmt.Sprintf("%s/%s/api/2024-01/graphql.json", c.baseURL, organizationID)

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
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Apps struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"apps"`
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

	apps := make([]PartnerApp, 0, len(result.Data.Apps.Edges))
	for _, edge := range result.Data.Apps.Edges {
		apps = append(apps, PartnerApp{
			ID:   edge.Node.ID,
			Name: edge.Node.Name,
		})
	}

	return apps, nil
}

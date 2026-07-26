package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ShopifyOAuthService provides Partner API calls that require an access token.
//
// Note: The Shopify Partner API has no OAuth authorization flow; access tokens
// are created manually in the Partner Dashboard. This service intentionally only
// retains token-based Partner API helpers (e.g. FetchOrganizationID).
type ShopifyOAuthService struct{}

func NewShopifyOAuthService() *ShopifyOAuthService {
	return &ShopifyOAuthService{}
}

// FetchOrganizationID retrieves the current organization ID using the access token
func (s *ShopifyOAuthService) FetchOrganizationID(ctx context.Context, accessToken string) (string, error) {
	// Query the Partner API to get the current organization
	query := `
		query {
			currentUser {
				organizations(first: 1) {
					edges {
						node {
							id
						}
					}
				}
			}
		}
	`

	reqBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal query: %w", err)
	}

	// Use the partners API endpoint
	apiURL := "https://partners.shopify.com/api/2025-07/graphql.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch organization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			CurrentUser struct {
				Organizations struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"` // e.g., "gid://partners/Organization/12345"
						} `json:"node"`
					} `json:"edges"`
				} `json:"organizations"`
			} `json:"currentUser"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	if len(result.Data.CurrentUser.Organizations.Edges) == 0 {
		return "", fmt.Errorf("no organizations found for user")
	}

	// Extract organization ID from GID (e.g., "gid://partners/Organization/12345" -> "12345")
	orgGID := result.Data.CurrentUser.Organizations.Edges[0].Node.ID
	parts := strings.Split(orgGID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return orgGID, nil
}

package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	shopifyAuthURL  = "https://partners.shopify.com/authorize"
	shopifyTokenURL = "https://partners.shopify.com/access_token"
)

type ShopifyOAuthService struct {
	clientID     string
	clientSecret string
	redirectURI  string
	scopes       string
	tokenURL     string // For testing
}

func NewShopifyOAuthService(clientID, clientSecret, redirectURI, scopes string) *ShopifyOAuthService {
	return &ShopifyOAuthService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		scopes:       scopes,
		tokenURL:     shopifyTokenURL,
	}
}

// GenerateAuthURL creates the OAuth authorization URL for Shopify Partners.
func (s *ShopifyOAuthService) GenerateAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("scope", s.scopes)
	params.Set("state", state)
	params.Set("response_type", "code")

	return fmt.Sprintf("%s?%s", shopifyAuthURL, params.Encode())
}

// ExchangeCodeForToken exchanges an authorization code for an access token.
func (s *ShopifyOAuthService) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return "", fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("token exchange failed: %s - %s", errResp.Error, errResp.Description)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return tokenResp.AccessToken, nil
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
	apiURL := "https://partners.shopify.com/api/2024-01/graphql.json"
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

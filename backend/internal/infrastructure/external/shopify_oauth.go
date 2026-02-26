package external

import (
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
		json.NewDecoder(resp.Body).Decode(&errResp)
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

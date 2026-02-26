package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateAuthURL(t *testing.T) {
	service := NewShopifyOAuthService(
		"test-client-id",
		"test-client-secret",
		"https://example.com/callback",
		"read_financials,read_apps",
	)

	state := "random-state-123"
	url := service.GenerateAuthURL(state)

	if !strings.Contains(url, "partners.shopify.com") {
		t.Errorf("URL should contain partners.shopify.com: %s", url)
	}

	if !strings.Contains(url, "client_id=test-client-id") {
		t.Errorf("URL should contain client_id: %s", url)
	}

	if !strings.Contains(url, "redirect_uri=https") {
		t.Errorf("URL should contain redirect_uri: %s", url)
	}

	if !strings.Contains(url, "state=random-state-123") {
		t.Errorf("URL should contain state: %s", url)
	}

	if !strings.Contains(url, "scope=read_financials") {
		t.Errorf("URL should contain scope: %s", url)
	}
}

func TestExchangeCodeForToken_Success(t *testing.T) {
	// Mock Shopify token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.FormValue("code") != "test-code" {
			t.Errorf("expected code 'test-code', got '%s'", r.FormValue("code"))
		}

		if r.FormValue("client_id") != "test-client-id" {
			t.Errorf("expected client_id 'test-client-id', got '%s'", r.FormValue("client_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "shppa_test_access_token",
			"scope":        "read_financials,read_apps",
		})
	}))
	defer server.Close()

	service := &ShopifyOAuthService{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURI:  "https://example.com/callback",
		scopes:       "read_financials,read_apps",
		tokenURL:     server.URL,
	}

	token, err := service.ExchangeCodeForToken(context.Background(), "test-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "shppa_test_access_token" {
		t.Errorf("expected token 'shppa_test_access_token', got '%s'", token)
	}
}

func TestExchangeCodeForToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "invalid_grant",
			"error_description": "The code has expired",
		})
	}))
	defer server.Close()

	service := &ShopifyOAuthService{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		redirectURI:  "https://example.com/callback",
		scopes:       "read_financials,read_apps",
		tokenURL:     server.URL,
	}

	_, err := service.ExchangeCodeForToken(context.Background(), "invalid-code")
	if err == nil {
		t.Error("expected error for invalid code")
	}
}

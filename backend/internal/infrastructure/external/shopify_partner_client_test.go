package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchApps_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("X-Shopify-Access-Token") != "test-token" {
			t.Errorf("expected token header")
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type application/json")
		}

		// Return mock response
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"apps": map[string]interface{}{
					"edges": []map[string]interface{}{
						{
							"node": map[string]interface{}{
								"id":   "gid://partners/App/12345",
								"name": "My App",
							},
						},
						{
							"node": map[string]interface{}{
								"id":   "gid://partners/App/67890",
								"name": "Another App",
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	apps, err := client.FetchApps(context.Background(), "org123", "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(apps))
	}

	if apps[0].ID != "gid://partners/App/12345" {
		t.Errorf("expected app ID 'gid://partners/App/12345', got %s", apps[0].ID)
	}

	if apps[0].Name != "My App" {
		t.Errorf("expected app name 'My App', got %s", apps[0].Name)
	}

	if apps[1].ID != "gid://partners/App/67890" {
		t.Errorf("expected app ID 'gid://partners/App/67890', got %s", apps[1].ID)
	}
}

func TestFetchApps_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Authentication required"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	_, err := client.FetchApps(context.Background(), "org123", "invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "graphql error: Authentication required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFetchApps_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	_, err := client.FetchApps(context.Background(), "org123", "invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchApps_EmptyApps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"apps": map[string]interface{}{
					"edges": []map[string]interface{}{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	apps, err := client.FetchApps(context.Background(), "org123", "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

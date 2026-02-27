package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
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

		// Return mock response - apps are extracted from transactions
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"transactions": map[string]interface{}{
					"edges": []map[string]interface{}{
						{
							"node": map[string]interface{}{
								"id": "gid://partners/AppSubscriptionSale/1",
								"app": map[string]interface{}{
									"id":   "gid://partners/App/12345",
									"name": "My App",
								},
							},
						},
						{
							"node": map[string]interface{}{
								"id": "gid://partners/AppSubscriptionSale/2",
								"app": map[string]interface{}{
									"id":   "gid://partners/App/67890",
									"name": "Another App",
								},
							},
						},
						{
							// Duplicate app should be deduplicated
							"node": map[string]interface{}{
								"id": "gid://partners/AppSubscriptionSale/3",
								"app": map[string]interface{}{
									"id":   "gid://partners/App/12345",
									"name": "My App",
								},
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

	// Should have 2 unique apps (one duplicate was deduplicated)
	if len(apps) != 2 {
		t.Fatalf("expected 2 unique apps, got %d", len(apps))
	}

	// Check that both apps are present (order may vary due to map iteration)
	appIDs := make(map[string]string)
	for _, app := range apps {
		appIDs[app.ID] = app.Name
	}

	if name, ok := appIDs["gid://partners/App/12345"]; !ok || name != "My App" {
		t.Errorf("expected app 'My App' with ID 'gid://partners/App/12345'")
	}

	if name, ok := appIDs["gid://partners/App/67890"]; !ok || name != "Another App" {
		t.Errorf("expected app 'Another App' with ID 'gid://partners/App/67890'")
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
				"transactions": map[string]interface{}{
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

func TestFetchTransactions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("X-Shopify-Access-Token") != "test-token" {
			t.Errorf("expected token header")
		}

		// Return mock response with transactions
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"transactions": map[string]interface{}{
					"edges": []map[string]interface{}{
						{
							"cursor": "cursor1",
							"node": map[string]interface{}{
								"id":        "gid://partners/AppSubscriptionSale/12345",
								"createdAt": "2024-02-15T10:00:00Z",
								"chargeId":  "charge123",
								"app": map[string]interface{}{
									"id":   "gid://partners/App/99",
									"name": "Test App",
								},
								"shop": map[string]interface{}{
									"myshopifyDomain": "test-shop.myshopify.com",
									"name":            "Test Shop",
								},
								"grossAmount": map[string]interface{}{
									"amount":       "35.99",
									"currencyCode": "USD",
								},
								"netAmount": map[string]interface{}{
									"amount":       "29.99",
									"currencyCode": "USD",
								},
							},
						},
						{
							"cursor": "cursor2",
							"node": map[string]interface{}{
								"id":        "gid://partners/AppUsageSale/67890",
								"createdAt": "2024-02-16T12:00:00Z",
								"chargeId":  "charge456",
								"app": map[string]interface{}{
									"id":   "gid://partners/App/99",
									"name": "Test App",
								},
								"shop": map[string]interface{}{
									"myshopifyDomain": "another-shop.myshopify.com",
									"name":            "Another Shop",
								},
								"grossAmount": map[string]interface{}{
									"amount":       "7.00",
									"currencyCode": "USD",
								},
								"netAmount": map[string]interface{}{
									"amount":       "5.50",
									"currencyCode": "USD",
								},
							},
						},
					},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
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

	appID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	ctx := WithOrganizationID(context.Background(), "org123")
	transactions, err := client.FetchTransactions(ctx, "test-token", appID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(transactions))
	}

	// Verify first transaction
	if transactions[0].ShopifyGID != "gid://partners/AppSubscriptionSale/12345" {
		t.Errorf("expected shopify GID 'gid://partners/AppSubscriptionSale/12345', got %s", transactions[0].ShopifyGID)
	}
	if transactions[0].MyshopifyDomain != "test-shop.myshopify.com" {
		t.Errorf("expected myshopify domain 'test-shop.myshopify.com', got %s", transactions[0].MyshopifyDomain)
	}
	if transactions[0].ShopName != "Test Shop" {
		t.Errorf("expected shop name 'Test Shop', got %s", transactions[0].ShopName)
	}
	if transactions[0].GrossAmountCents != 3599 {
		t.Errorf("expected gross amount 3599 cents, got %d", transactions[0].GrossAmountCents)
	}
	if transactions[0].AmountCents() != 2999 {
		t.Errorf("expected net amount 2999 cents, got %d", transactions[0].AmountCents())
	}
	if transactions[0].Currency != "USD" {
		t.Errorf("expected currency USD, got %s", transactions[0].Currency)
	}
	if transactions[0].ChargeType != valueobject.ChargeTypeRecurring {
		t.Errorf("expected charge type RECURRING, got %s", transactions[0].ChargeType)
	}

	// Verify second transaction
	if transactions[1].ShopName != "Another Shop" {
		t.Errorf("expected shop name 'Another Shop', got %s", transactions[1].ShopName)
	}
	if transactions[1].GrossAmountCents != 700 {
		t.Errorf("expected gross amount 700 cents, got %d", transactions[1].GrossAmountCents)
	}
	if transactions[1].AmountCents() != 550 {
		t.Errorf("expected net amount 550 cents, got %d", transactions[1].AmountCents())
	}
}

func TestFetchTransactions_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var response map[string]interface{}
		if callCount == 1 {
			// First page
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"transactions": map[string]interface{}{
						"edges": []map[string]interface{}{
							{
								"cursor": "cursor1",
								"node": map[string]interface{}{
									"id":        "gid://partners/AppSubscriptionSale/1",
									"createdAt": "2024-02-15T10:00:00Z",
									"chargeId":  "charge1",
									"app": map[string]interface{}{
										"id":   "gid://partners/App/99",
										"name": "Test App",
									},
									"shop": map[string]interface{}{
										"myshopifyDomain": "shop1.myshopify.com",
									},
									"netAmount": map[string]interface{}{
										"amount":       "10.00",
										"currencyCode": "USD",
									},
								},
							},
						},
						"pageInfo": map[string]interface{}{
							"hasNextPage": true,
						},
					},
				},
			}
		} else {
			// Second page (last)
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"transactions": map[string]interface{}{
						"edges": []map[string]interface{}{
							{
								"cursor": "cursor2",
								"node": map[string]interface{}{
									"id":        "gid://partners/AppSubscriptionSale/2",
									"createdAt": "2024-02-16T10:00:00Z",
									"chargeId":  "charge2",
									"app": map[string]interface{}{
										"id":   "gid://partners/App/99",
										"name": "Test App",
									},
									"shop": map[string]interface{}{
										"myshopifyDomain": "shop2.myshopify.com",
									},
									"netAmount": map[string]interface{}{
										"amount":       "20.00",
										"currencyCode": "USD",
									},
								},
							},
						},
						"pageInfo": map[string]interface{}{
							"hasNextPage": false,
						},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	appID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	ctx := WithOrganizationID(context.Background(), "org123")
	transactions, err := client.FetchTransactions(ctx, "test-token", appID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}

	if len(transactions) != 2 {
		t.Fatalf("expected 2 transactions total, got %d", len(transactions))
	}
}

func TestFetchTransactions_NoOrganizationID(t *testing.T) {
	client := &ShopifyPartnerClient{
		httpClient: &http.Client{},
		baseURL:    "https://partners.shopify.com",
	}

	appID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	// Context without organization ID
	_, err := client.FetchTransactions(context.Background(), "test-token", appID, from, to)
	if err == nil {
		t.Fatal("expected error for missing organization ID, got nil")
	}

	if err.Error() != "organization ID not set" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFetchTransactions_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Invalid access token"},
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

	appID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	ctx := WithOrganizationID(context.Background(), "org123")
	_, err := client.FetchTransactions(ctx, "invalid-token", appID, from, to)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "graphql error: Invalid access token" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFetchTransactions_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{
		httpClient: server.Client(),
		baseURL:    server.URL,
	}

	appID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	ctx := WithOrganizationID(context.Background(), "org123")
	_, err := client.FetchTransactions(ctx, "test-token", appID, from, to)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchTransactions_EmptyTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"transactions": map[string]interface{}{
					"edges": []map[string]interface{}{},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
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

	appID := uuid.New()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)

	ctx := WithOrganizationID(context.Background(), "org123")
	transactions, err := client.FetchTransactions(ctx, "test-token", appID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(transactions))
	}
}


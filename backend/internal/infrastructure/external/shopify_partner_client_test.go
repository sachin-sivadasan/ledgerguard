package external

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestFetchAppEvents_PaginatesForwardWithFirstAfter verifies the app-events fetch uses
// forward pagination (first/after, NOT `last` — which the Partner API rejects with
// "Using last without before is not supported") and collects events across all pages.
func TestFetchAppEvents_PaginatesForwardWithFirstAfter(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		body, _ := io.ReadAll(r.Body)
		q := string(body)
		if strings.Contains(q, "last:") {
			t.Errorf("query must not use `last:` (Partner API rejects it without before); got: %s", q)
		}
		if !strings.Contains(q, "first: 100") {
			t.Errorf("query must forward-paginate with `first: 100`; got: %s", q)
		}
		if strings.Contains(q, "endCursor") {
			t.Errorf("query must NOT request pageInfo.endCursor (not on Partner API PageInfo); got: %s", q)
		}

		var response map[string]interface{}
		if callCount == 1 {
			if strings.Contains(q, `"cursor":`) {
				t.Errorf("first page must not send a cursor variable; got: %s", q)
			}
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"app": map[string]interface{}{
						"events": map[string]interface{}{
							"edges": []map[string]interface{}{
								{"cursor": "c1", "node": map[string]interface{}{
									"type":       "RELATIONSHIP_INSTALLED",
									"occurredAt": "2024-02-15T10:00:00Z",
									"shop":       map[string]interface{}{"id": "gid://partners/Shop/1", "name": "Shop One"},
								}},
							},
							"pageInfo": map[string]interface{}{"hasNextPage": true},
						},
					},
				},
			}
		} else {
			if !strings.Contains(q, "c1") {
				t.Errorf("second page must send the last edge cursor `c1` as after; got: %s", q)
			}
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"app": map[string]interface{}{
						"events": map[string]interface{}{
							"edges": []map[string]interface{}{
								{"cursor": "c2", "node": map[string]interface{}{
									"type":       "SUBSCRIPTION_CHARGE_ACCEPTED",
									"occurredAt": "2024-02-16T10:00:00Z",
									"shop":       map[string]interface{}{"id": "gid://partners/Shop/1", "name": "Shop One"},
								}},
							},
							"pageInfo": map[string]interface{}{"hasNextPage": false},
						},
					},
				},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{httpClient: server.Client(), baseURL: server.URL}
	ctx := WithOrganizationID(context.Background(), "org123")

	events, err := client.FetchAppEvents(ctx, "org123", "test-token", "gid://partners/App/99", "gid://partners/Shop/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 paginated calls, got %d", callCount)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events across pages, got %d", len(events))
	}
	if events[0].Type != "RELATIONSHIP_INSTALLED" || events[1].Type != "SUBSCRIPTION_CHARGE_ACCEPTED" {
		t.Errorf("unexpected event types: %+v", events)
	}
}

// TestFetchAppEvents_SinglePageStops verifies a single page (hasNextPage=false) makes
// exactly one call.
func TestFetchAppEvents_SinglePageStops(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"app": map[string]interface{}{
					"events": map[string]interface{}{
						"edges":    []map[string]interface{}{},
						"pageInfo": map[string]interface{}{"hasNextPage": false},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &ShopifyPartnerClient{httpClient: server.Client(), baseURL: server.URL}
	ctx := WithOrganizationID(context.Background(), "org123")
	events, err := client.FetchAppEvents(ctx, "org123", "test-token", "gid://partners/App/99", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call for a single page, got %d", callCount)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// TestExecuteWithRetry_RewindsBodyOnRetry guards the empty-body-on-retry bug: after a
// 429, the retry must re-send the FULL request body (a consumed bytes.Reader would send an
// empty POST → Cloudflare 400). Verified via FetchAppEvents (429 → 200).
func TestExecuteWithRetry_RewindsBodyOnRetry(t *testing.T) {
	callCount := 0
	bodyLens := []int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		body, _ := io.ReadAll(r.Body)
		bodyLens = append(bodyLens, len(body))

		if callCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests) // force a retry
			return
		}
		// Retry must carry the query, not an empty body.
		if !strings.Contains(string(body), "events(") {
			t.Errorf("retry sent a body without the query (empty-body bug); got %q", string(body))
		}
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"app": map[string]interface{}{
					"events": map[string]interface{}{
						"edges": []map[string]interface{}{
							{"cursor": "c1", "node": map[string]interface{}{"type": "RELATIONSHIP_INSTALLED", "occurredAt": "2024-02-15T10:00:00Z"}},
						},
						"pageInfo": map[string]interface{}{"hasNextPage": false},
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
		config:     RateLimiterConfig{MaxRetries: 2, BaseBackoff: time.Millisecond, MaxBackoff: time.Millisecond},
	}
	ctx := WithOrganizationID(context.Background(), "org123")

	events, err := client.FetchAppEvents(ctx, "org123", "test-token", "gid://partners/App/99", "gid://partners/Shop/1")
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected a retry (2 calls), got %d", callCount)
	}
	if bodyLens[0] == 0 || bodyLens[1] == 0 {
		t.Errorf("both attempts must send a non-empty body; got lengths %v", bodyLens)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event after successful retry, got %d", len(events))
	}
}

// TestParseTransaction_AppSaleAdjustmentRefund: a refund/adjustment is classified REFUND
// with Shopify's natural NEGATIVE net preserved (so any SUM(net) nets it out).
func TestParseTransaction_AppSaleAdjustmentRefund(t *testing.T) {
	c := &ShopifyPartnerClient{}
	node := transactionNode{
		Typename:  "AppSaleAdjustment",
		ID:        "gid://partners/AppSaleAdjustment/1",
		CreatedAt: "2026-02-15T10:00:00Z",
		App: &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: "gid://partners/App/99", Name: "Test App"},
		Shop: &struct {
			ID              string `json:"id"`
			MyshopifyDomain string `json:"myshopifyDomain"`
			Name            string `json:"name"`
		}{ID: "gid://shopify/Shop/1", MyshopifyDomain: "s.myshopify.com", Name: "S"},
		GrossAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "-10.00", CurrencyCode: "USD"},
		NetAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "-7.00", CurrencyCode: "USD"},
	}
	tx := c.parseTransaction(node, uuid.New())
	if tx == nil {
		t.Fatal("expected a transaction, got nil")
	}
	if tx.ChargeType != valueobject.ChargeTypeRefund {
		t.Errorf("expected REFUND, got %s", tx.ChargeType)
	}
	if tx.NetAmountCents != -700 {
		t.Errorf("expected natural negative net -700, got %d", tx.NetAmountCents)
	}
}

// TestParseTransaction_ShopifyFee: the shopifyFee field (Shopify's retained cut) is
// parsed into ShopifyFeeCents — the signal that drives Profit & Expense + the Fee Guard.
func TestParseTransaction_ShopifyFee(t *testing.T) {
	c := &ShopifyPartnerClient{}
	node := transactionNode{
		Typename:  "AppSubscriptionSale",
		ID:        "gid://partners/AppSubscriptionSale/1",
		CreatedAt: "2026-02-15T10:00:00Z",
		App: &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: "gid://partners/App/99", Name: "Test App"},
		Shop: &struct {
			ID              string `json:"id"`
			MyshopifyDomain string `json:"myshopifyDomain"`
			Name            string `json:"name"`
		}{ID: "gid://shopify/Shop/1", MyshopifyDomain: "s.myshopify.com", Name: "S"},
		GrossAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "100.00", CurrencyCode: "USD"},
		NetAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "85.00", CurrencyCode: "USD"},
		ShopifyFee: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "15.00", CurrencyCode: "USD"},
	}
	tx := c.parseTransaction(node, uuid.New())
	if tx == nil {
		t.Fatal("expected a transaction, got nil")
	}
	if tx.ShopifyFeeCents != 1500 {
		t.Errorf("expected ShopifyFeeCents=1500 (from $15.00 shopifyFee), got %d", tx.ShopifyFeeCents)
	}
	// net (85) == gross (100) − shopifyFee (15): no residual, so no derived processing fee.
	if tx.ProcessingFeeCents != 0 {
		t.Errorf("expected ProcessingFeeCents=0 when net == gross − fee, got %d", tx.ProcessingFeeCents)
	}
}

// TestParseTransaction_DerivesProcessingFee: Shopify bakes a payment-processing deduction
// into netAmount that it does NOT expose as a field. When net < gross − shopifyFee, the
// remainder is recovered into ProcessingFeeCents so gross = net + shopifyFee + processing.
func TestParseTransaction_DerivesProcessingFee(t *testing.T) {
	c := &ShopifyPartnerClient{}
	node := transactionNode{
		Typename:  "AppSubscriptionSale",
		ID:        "gid://partners/AppSubscriptionSale/2",
		CreatedAt: "2026-02-15T10:00:00Z",
		App: &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: "gid://partners/App/99", Name: "Test App"},
		Shop: &struct {
			ID              string `json:"id"`
			MyshopifyDomain string `json:"myshopifyDomain"`
			Name            string `json:"name"`
		}{ID: "gid://shopify/Shop/1", MyshopifyDomain: "s.myshopify.com", Name: "S"},
		GrossAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "100.00", CurrencyCode: "USD"},
		NetAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "82.00", CurrencyCode: "USD"}, // Shopify deposited $82: $15 revenue share + $3 processing
		ShopifyFee: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "15.00", CurrencyCode: "USD"},
	}
	tx := c.parseTransaction(node, uuid.New())
	if tx == nil {
		t.Fatal("expected a transaction, got nil")
	}
	// 10000 − 1500 − 8200 = 300
	if tx.ProcessingFeeCents != 300 {
		t.Errorf("expected derived ProcessingFeeCents=300 (gross−fee−net), got %d", tx.ProcessingFeeCents)
	}
	// The identity must close: gross == net + shopifyFee + processing.
	if got := tx.NetAmountCents + tx.ShopifyFeeCents + tx.ProcessingFeeCents; got != tx.GrossAmountCents {
		t.Errorf("reconciliation identity broken: net+fee+processing=%d, gross=%d", got, tx.GrossAmountCents)
	}
}

// TestParseTransaction_AppSaleCredit: AppSaleCredit (a credit against an app charge) is now
// fetched and maps to REFUND — its negative net is the signed payout effect, and no
// processing fee is derived for it.
func TestParseTransaction_AppSaleCredit(t *testing.T) {
	c := &ShopifyPartnerClient{}
	node := transactionNode{
		Typename:  "AppSaleCredit",
		ID:        "gid://partners/AppSaleCredit/1",
		CreatedAt: "2026-02-15T10:00:00Z",
		App: &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: "gid://partners/App/99", Name: "Test App"},
		Shop: &struct {
			ID              string `json:"id"`
			MyshopifyDomain string `json:"myshopifyDomain"`
			Name            string `json:"name"`
		}{ID: "gid://shopify/Shop/1", MyshopifyDomain: "s.myshopify.com", Name: "S"},
		GrossAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "-10.00", CurrencyCode: "USD"},
		NetAmount: &struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currencyCode"`
		}{Amount: "-8.50", CurrencyCode: "USD"},
	}
	tx := c.parseTransaction(node, uuid.New())
	if tx == nil {
		t.Fatal("expected a transaction, got nil")
	}
	if tx.ChargeType != valueobject.ChargeTypeRefund {
		t.Errorf("expected REFUND for AppSaleCredit, got %s", tx.ChargeType)
	}
	if tx.NetAmountCents != -850 {
		t.Errorf("expected natural negative net -850, got %d", tx.NetAmountCents)
	}
	if tx.ProcessingFeeCents != 0 {
		t.Errorf("expected no derived processing fee on a credit, got %d", tx.ProcessingFeeCents)
	}
}

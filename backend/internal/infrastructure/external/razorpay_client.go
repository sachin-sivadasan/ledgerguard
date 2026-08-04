package external

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const razorpayBaseURL = "https://api.razorpay.com/v1"

// RazorpayClient communicates with the Razorpay API using Basic Auth.
type RazorpayClient struct {
	keyID      string
	keySecret  string
	httpClient *http.Client
	baseURL    string
}

// NewRazorpayClient creates a production Razorpay client.
func NewRazorpayClient(keyID, keySecret string) *RazorpayClient {
	return &RazorpayClient{
		keyID:     keyID,
		keySecret: keySecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: razorpayBaseURL,
	}
}

// NewRazorpayClientWithHTTPClient creates a test-injectable Razorpay client.
func NewRazorpayClientWithHTTPClient(keyID, keySecret string, httpClient *http.Client, baseURL string) *RazorpayClient {
	return &RazorpayClient{
		keyID:      keyID,
		keySecret:  keySecret,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// RazorpayCustomer represents a Razorpay customer.
type RazorpayCustomer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// RazorpaySubscription represents a Razorpay subscription response.
type RazorpaySubscription struct {
	ID             string `json:"id"`
	PlanID         string `json:"plan_id"`
	CustomerID     string `json:"customer_id"`
	Status         string `json:"status"`
	ShortURL       string `json:"short_url"`
	CurrentStart   *int64 `json:"current_start"`
	CurrentEnd     *int64 `json:"current_end"`
	TotalCount     int    `json:"total_count"`
	PaidCount      int    `json:"paid_count"`
	RemainingCount int    `json:"remaining_count"`
	ChargeAt       *int64 `json:"charge_at"`
}

// CreateCustomerRequest is the payload for creating a Razorpay customer.
type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateSubscriptionRequest is the payload for creating a Razorpay subscription.
type CreateSubscriptionRequest struct {
	PlanID         string `json:"plan_id"`
	CustomerID     string `json:"customer_id,omitempty"`
	TotalCount     int    `json:"total_count"`
	CustomerNotify int    `json:"customer_notify"`
}

// CreateCustomer creates a customer in Razorpay.
func (c *RazorpayClient) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*RazorpayCustomer, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal customer request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/customers", body)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}

	var customer RazorpayCustomer
	if err := json.Unmarshal(resp, &customer); err != nil {
		return nil, fmt.Errorf("unmarshal customer response: %w", err)
	}
	return &customer, nil
}

// CreateSubscription creates a subscription in Razorpay.
func (c *RazorpayClient) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*RazorpaySubscription, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal subscription request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/subscriptions", body)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	var sub RazorpaySubscription
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal subscription response: %w", err)
	}
	return &sub, nil
}

// FetchSubscription retrieves a subscription by ID.
func (c *RazorpayClient) FetchSubscription(ctx context.Context, subscriptionID string) (*RazorpaySubscription, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/subscriptions/"+subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription: %w", err)
	}

	var sub RazorpaySubscription
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal subscription response: %w", err)
	}
	return &sub, nil
}

// CancelSubscription cancels a subscription.
func (c *RazorpayClient) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtCycleEnd bool) (*RazorpaySubscription, error) {
	payload := map[string]bool{"cancel_at_cycle_end": cancelAtCycleEnd}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal cancel request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/subscriptions/"+subscriptionID+"/cancel", body)
	if err != nil {
		return nil, fmt.Errorf("cancel subscription: %w", err)
	}

	var sub RazorpaySubscription
	if err := json.Unmarshal(resp, &sub); err != nil {
		return nil, fmt.Errorf("unmarshal subscription response: %w", err)
	}
	return &sub, nil
}

// VerifyWebhookSignature verifies a Razorpay webhook signature using HMAC-SHA256.
func VerifyWebhookSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (c *RazorpayClient) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.SetBasicAuth(c.keyID, c.keySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("razorpay API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

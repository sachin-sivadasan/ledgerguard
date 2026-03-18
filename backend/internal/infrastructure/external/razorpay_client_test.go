package external

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRazorpayClient_CreateCustomer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/customers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		user, pass, ok := r.BasicAuth()
		if !ok || user != "rzp_test_key" || pass != "rzp_test_secret" {
			t.Errorf("unexpected auth: %s/%s", user, pass)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RazorpayCustomer{
			ID:    "cust_123",
			Name:  "Test User",
			Email: "test@example.com",
		})
	}))
	defer server.Close()

	client := NewRazorpayClientWithHTTPClient("rzp_test_key", "rzp_test_secret", server.Client(), server.URL)
	cust, err := client.CreateCustomer(context.Background(), CreateCustomerRequest{
		Name:  "Test User",
		Email: "test@example.com",
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error: %v", err)
	}
	if cust.ID != "cust_123" {
		t.Errorf("customer ID = %q, want cust_123", cust.ID)
	}
}

func TestRazorpayClient_CreateSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RazorpaySubscription{
			ID:       "sub_456",
			PlanID:   "plan_starter",
			ShortURL: "https://rzp.io/i/test",
			Status:   "created",
		})
	}))
	defer server.Close()

	client := NewRazorpayClientWithHTTPClient("key", "secret", server.Client(), server.URL)
	sub, err := client.CreateSubscription(context.Background(), CreateSubscriptionRequest{
		PlanID:     "plan_starter",
		CustomerID: "cust_123",
		TotalCount: 120,
	})
	if err != nil {
		t.Fatalf("CreateSubscription() error: %v", err)
	}
	if sub.ID != "sub_456" {
		t.Errorf("subscription ID = %q, want sub_456", sub.ID)
	}
	if sub.ShortURL != "https://rzp.io/i/test" {
		t.Errorf("ShortURL = %q, want https://rzp.io/i/test", sub.ShortURL)
	}
}

func TestRazorpayClient_FetchSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions/sub_789" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RazorpaySubscription{
			ID:     "sub_789",
			Status: "active",
		})
	}))
	defer server.Close()

	client := NewRazorpayClientWithHTTPClient("key", "secret", server.Client(), server.URL)
	sub, err := client.FetchSubscription(context.Background(), "sub_789")
	if err != nil {
		t.Fatalf("FetchSubscription() error: %v", err)
	}
	if sub.Status != "active" {
		t.Errorf("status = %q, want active", sub.Status)
	}
}

func TestRazorpayClient_CancelSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions/sub_789/cancel" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RazorpaySubscription{
			ID:     "sub_789",
			Status: "cancelled",
		})
	}))
	defer server.Close()

	client := NewRazorpayClientWithHTTPClient("key", "secret", server.Client(), server.URL)
	sub, err := client.CancelSubscription(context.Background(), "sub_789", true)
	if err != nil {
		t.Fatalf("CancelSubscription() error: %v", err)
	}
	if sub.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", sub.Status)
	}
}

func TestRazorpayClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"code":"BAD_REQUEST_ERROR","description":"Invalid plan_id"}}`))
	}))
	defer server.Close()

	client := NewRazorpayClientWithHTTPClient("key", "secret", server.Client(), server.URL)
	_, err := client.CreateSubscription(context.Background(), CreateSubscriptionRequest{})
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	body := []byte(`{"event":"subscription.activated"}`)
	secret := "webhook_secret_123"

	// Compute valid HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	validSig := hex.EncodeToString(mac.Sum(nil))

	if !VerifyWebhookSignature(body, validSig, secret) {
		t.Error("expected valid signature to pass")
	}

	if VerifyWebhookSignature(body, "invalid_sig", secret) {
		t.Error("expected invalid signature to fail")
	}

	if VerifyWebhookSignature(body, "", secret) {
		t.Error("expected empty signature to fail")
	}
}

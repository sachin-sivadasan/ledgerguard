package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	appservice "github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
)

// --- Mocks for BillingService dependencies ---

type mockBillingSubRepoForHandler struct {
	created *entity.BillingSubscription
}

func (m *mockBillingSubRepoForHandler) Create(ctx context.Context, bs *entity.BillingSubscription) error {
	m.created = bs
	return nil
}
func (m *mockBillingSubRepoForHandler) Update(ctx context.Context, bs *entity.BillingSubscription) error {
	return nil
}
func (m *mockBillingSubRepoForHandler) FindByID(ctx context.Context, id uuid.UUID) (*entity.BillingSubscription, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockBillingSubRepoForHandler) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.BillingSubscription, error) {
	return nil, nil
}
func (m *mockBillingSubRepoForHandler) FindByRazorpaySubscriptionID(ctx context.Context, rzpSubID string) (*entity.BillingSubscription, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockBillingSubRepoForHandler) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.BillingSubscription, error) {
	return nil, fmt.Errorf("not found")
}

type mockUserRepoForBillingHandler struct {
	user *entity.User
}

func (m *mockUserRepoForBillingHandler) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.user, nil
}
func (m *mockUserRepoForBillingHandler) FindByFirebaseUID(ctx context.Context, uid string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepoForBillingHandler) Create(ctx context.Context, user *entity.User) error {
	return nil
}
func (m *mockUserRepoForBillingHandler) Update(ctx context.Context, user *entity.User) error {
	return nil
}

func TestBillingHandler_CreateCheckout_Success(t *testing.T) {
	userID := uuid.New()
	user := &entity.User{ID: userID, Email: "test@example.com"}

	// Mock Razorpay API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/customers":
			json.NewEncoder(w).Encode(external.RazorpayCustomer{ID: "cust_1", Email: user.Email})
		case "/subscriptions":
			json.NewEncoder(w).Encode(external.RazorpaySubscription{
				ID: "sub_1", ShortURL: "https://rzp.io/i/test", Status: "created",
			})
		}
	}))
	defer server.Close()

	rzpClient := external.NewRazorpayClientWithHTTPClient("key", "secret", server.Client(), server.URL)
	billingRepo := &mockBillingSubRepoForHandler{}
	userRepo := &mockUserRepoForBillingHandler{user: user}
	svc := appservice.NewBillingService(rzpClient, billingRepo, userRepo, "secret", "plan_starter", "plan_pro")
	h := NewBillingHandler(svc)

	body := bytes.NewBufferString(`{"plan":"STARTER"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", body)
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	h.CreateCheckout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var resp appservice.CheckoutResult
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.SubscriptionID != "sub_1" {
		t.Errorf("SubscriptionID = %q, want sub_1", resp.SubscriptionID)
	}
	if resp.ShortURL != "https://rzp.io/i/test" {
		t.Errorf("ShortURL = %q, want https://rzp.io/i/test", resp.ShortURL)
	}
}

func TestBillingHandler_CreateCheckout_Unauthorized(t *testing.T) {
	h := NewBillingHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", nil)
	rec := httptest.NewRecorder()

	h.CreateCheckout(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestBillingHandler_CreateCheckout_InvalidPlan(t *testing.T) {
	user := &entity.User{ID: uuid.New(), Email: "test@example.com"}
	h := NewBillingHandler(nil)

	body := bytes.NewBufferString(`{"plan":"INVALID"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", body)
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	h.CreateCheckout(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestBillingHandler_GetStatus_Unauthorized(t *testing.T) {
	h := NewBillingHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/status", nil)
	rec := httptest.NewRecorder()

	h.GetStatus(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestBillingHandler_HandleWebhook_InvalidSignature(t *testing.T) {
	svc := appservice.NewBillingService(nil, nil, nil, "webhook_secret", "", "")
	h := NewBillingHandler(svc)

	body := []byte(`{"event":"subscription.activated"}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/razorpay", bytes.NewReader(body))
	req.Header.Set("X-Razorpay-Signature", "bad_signature")
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (always 200 to prevent Razorpay retries)", rec.Code)
	}
}

func TestBillingHandler_HandleWebhook_ValidSignature(t *testing.T) {
	secret := "webhook_secret"
	svc := appservice.NewBillingService(nil, &mockBillingSubRepoForHandler{}, nil, secret, "", "")
	h := NewBillingHandler(svc)

	payload := map[string]interface{}{
		"event": "subscription.activated",
		"payload": map[string]interface{}{
			"subscription": map[string]interface{}{
				"entity": map[string]interface{}{"id": "sub_unknown", "status": "active"},
			},
		},
	}
	body, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/webhooks/razorpay", bytes.NewReader(body))
	req.Header.Set("X-Razorpay-Signature", sig)
	rec := httptest.NewRecorder()

	h.HandleWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
)

// --- Mock BillingSubscriptionRepository ---

type mockBillingSubRepo struct {
	created    *entity.BillingSubscription
	updated    *entity.BillingSubscription
	byID       *entity.BillingSubscription
	byRzpID    *entity.BillingSubscription
	activeByUser *entity.BillingSubscription
	findErr    error
}

func (m *mockBillingSubRepo) Create(ctx context.Context, bs *entity.BillingSubscription) error {
	m.created = bs
	return nil
}

func (m *mockBillingSubRepo) Update(ctx context.Context, bs *entity.BillingSubscription) error {
	m.updated = bs
	return nil
}

func (m *mockBillingSubRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.BillingSubscription, error) {
	if m.byID != nil {
		return m.byID, nil
	}
	return nil, m.findErr
}

func (m *mockBillingSubRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.BillingSubscription, error) {
	return nil, nil
}

func (m *mockBillingSubRepo) FindByRazorpaySubscriptionID(ctx context.Context, razorpaySubID string) (*entity.BillingSubscription, error) {
	if m.byRzpID != nil {
		return m.byRzpID, nil
	}
	return nil, m.findErr
}

func (m *mockBillingSubRepo) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.BillingSubscription, error) {
	if m.activeByUser != nil {
		return m.activeByUser, nil
	}
	return nil, m.findErr
}

// --- Mock UserRepository ---

type mockUserRepoForBilling struct {
	user      *entity.User
	updated   *entity.User
	findErr   error
}

func (m *mockUserRepoForBilling) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if m.user != nil {
		return m.user, nil
	}
	return nil, m.findErr
}

func (m *mockUserRepoForBilling) FindByFirebaseUID(ctx context.Context, uid string) (*entity.User, error) {
	return nil, nil
}

func (m *mockUserRepoForBilling) Create(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUserRepoForBilling) Update(ctx context.Context, user *entity.User) error {
	m.updated = user
	return nil
}

// --- Tests ---

func TestBillingService_CreateCheckout(t *testing.T) {
	userID := uuid.New()
	user := &entity.User{ID: userID, Email: "test@example.com"}

	// Mock Razorpay server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/customers":
			json.NewEncoder(w).Encode(external.RazorpayCustomer{
				ID:    "cust_test",
				Email: "test@example.com",
			})
		case "/subscriptions":
			json.NewEncoder(w).Encode(external.RazorpaySubscription{
				ID:       "sub_test",
				PlanID:   "plan_starter_test",
				ShortURL: "https://rzp.io/i/checkout",
				Status:   "created",
			})
		}
	}))
	defer server.Close()

	rzpClient := external.NewRazorpayClientWithHTTPClient("key", "secret", server.Client(), server.URL)
	billingRepo := &mockBillingSubRepo{}
	userRepo := &mockUserRepoForBilling{user: user}

	svc := NewBillingService(rzpClient, billingRepo, userRepo, "webhook_secret", "plan_starter_test", "plan_pro_test")

	result, err := svc.CreateCheckout(context.Background(), userID, valueobject.BillingPlanStarter)
	if err != nil {
		t.Fatalf("CreateCheckout() error: %v", err)
	}

	if result.SubscriptionID != "sub_test" {
		t.Errorf("SubscriptionID = %q, want sub_test", result.SubscriptionID)
	}
	if result.ShortURL != "https://rzp.io/i/checkout" {
		t.Errorf("ShortURL = %q, want https://rzp.io/i/checkout", result.ShortURL)
	}
	if billingRepo.created == nil {
		t.Fatal("expected billing subscription to be created")
	}
	if billingRepo.created.Plan != valueobject.BillingPlanStarter {
		t.Errorf("created plan = %v, want STARTER", billingRepo.created.Plan)
	}
}

func TestBillingService_CreateCheckout_InvalidPlan(t *testing.T) {
	svc := NewBillingService(nil, nil, nil, "", "", "")
	_, err := svc.CreateCheckout(context.Background(), uuid.New(), valueobject.BillingPlan("INVALID"))
	if err == nil {
		t.Fatal("expected error for invalid plan")
	}
}

func TestBillingService_GetBillingStatus_Active(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	periodEnd := now.Add(30 * 24 * time.Hour)

	billingRepo := &mockBillingSubRepo{
		activeByUser: &entity.BillingSubscription{
			Plan:               valueobject.BillingPlanPro,
			Status:             valueobject.BillingSubscriptionStatusActive,
			AmountCents:        49900,
			Currency:           "USD",
			CurrentPeriodStart: &now,
			CurrentPeriodEnd:   &periodEnd,
		},
	}

	svc := NewBillingService(nil, billingRepo, nil, "", "", "")
	status, err := svc.GetBillingStatus(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetBillingStatus() error: %v", err)
	}
	if status.Plan != "PRO" {
		t.Errorf("Plan = %q, want PRO", status.Plan)
	}
	if status.Status != "ACTIVE" {
		t.Errorf("Status = %q, want ACTIVE", status.Status)
	}
}

func TestBillingService_GetBillingStatus_NoSubscription(t *testing.T) {
	billingRepo := &mockBillingSubRepo{findErr: fmt.Errorf("not found")}
	svc := NewBillingService(nil, billingRepo, nil, "", "", "")

	status, err := svc.GetBillingStatus(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetBillingStatus() error: %v", err)
	}
	if status.Plan != "FREE" {
		t.Errorf("Plan = %q, want FREE", status.Plan)
	}
}

func TestBillingService_HandleWebhookEvent_InvalidSignature(t *testing.T) {
	svc := NewBillingService(nil, nil, nil, "secret123", "", "")
	err := svc.HandleWebhookEvent(context.Background(), []byte(`{}`), "invalid_sig")
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestBillingService_HandleWebhookEvent_Activated(t *testing.T) {
	userID := uuid.New()
	user := &entity.User{ID: userID, Email: "test@example.com", PlanTier: valueobject.PlanTierFree}

	bs := entity.NewBillingSubscription(userID, "sub_rzp_1", "plan_1", "cust_1", valueobject.BillingPlanStarter, "")
	billingRepo := &mockBillingSubRepo{byRzpID: bs}
	userRepo := &mockUserRepoForBilling{user: user}

	secret := "webhook_secret"
	svc := NewBillingService(nil, billingRepo, userRepo, secret, "", "")

	startTS := int64(1700000000)
	endTS := int64(1702592000)
	payload := map[string]interface{}{
		"event": "subscription.activated",
		"payload": map[string]interface{}{
			"subscription": map[string]interface{}{
				"entity": map[string]interface{}{
					"id":            "sub_rzp_1",
					"status":        "active",
					"current_start": startTS,
					"current_end":   endTS,
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := computeHMACSignature(body, secret)

	err := svc.HandleWebhookEvent(context.Background(), body, sig)
	if err != nil {
		t.Fatalf("HandleWebhookEvent() error: %v", err)
	}

	if billingRepo.updated == nil {
		t.Fatal("expected subscription to be updated")
	}
	if billingRepo.updated.Status != valueobject.BillingSubscriptionStatusActive {
		t.Errorf("status = %v, want ACTIVE", billingRepo.updated.Status)
	}
	if userRepo.updated == nil {
		t.Fatal("expected user to be updated")
	}
	if userRepo.updated.PlanTier != valueobject.PlanTierStarter {
		t.Errorf("user plan tier = %v, want STARTER", userRepo.updated.PlanTier)
	}
}

func TestBillingService_HandleWebhookEvent_Cancelled(t *testing.T) {
	userID := uuid.New()
	user := &entity.User{ID: userID, PlanTier: valueobject.PlanTierStarter}

	bs := entity.NewBillingSubscription(userID, "sub_rzp_2", "plan_1", "cust_1", valueobject.BillingPlanStarter, "")
	bs.Status = valueobject.BillingSubscriptionStatusActive
	billingRepo := &mockBillingSubRepo{byRzpID: bs}
	userRepo := &mockUserRepoForBilling{user: user}

	secret := "webhook_secret"
	svc := NewBillingService(nil, billingRepo, userRepo, secret, "", "")

	payload := map[string]interface{}{
		"event": "subscription.cancelled",
		"payload": map[string]interface{}{
			"subscription": map[string]interface{}{
				"entity": map[string]interface{}{
					"id":     "sub_rzp_2",
					"status": "cancelled",
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	sig := computeHMACSignature(body, secret)

	err := svc.HandleWebhookEvent(context.Background(), body, sig)
	if err != nil {
		t.Fatalf("HandleWebhookEvent() error: %v", err)
	}

	if billingRepo.updated.Status != valueobject.BillingSubscriptionStatusCancelled {
		t.Errorf("status = %v, want CANCELLED", billingRepo.updated.Status)
	}
	if userRepo.updated.PlanTier != valueobject.PlanTierFree {
		t.Errorf("user plan tier = %v, want FREE", userRepo.updated.PlanTier)
	}
}

func computeHMACSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

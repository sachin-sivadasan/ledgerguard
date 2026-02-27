package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock subscription repository
type mockSubscriptionRepo struct {
	subscriptions []*entity.Subscription
	subscription  *entity.Subscription
	findErr       error
	findAllErr    error
}

func (m *mockSubscriptionRepo) Upsert(ctx context.Context, sub *entity.Subscription) error {
	return nil
}

func (m *mockSubscriptionRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.subscription, nil
}

func (m *mockSubscriptionRepo) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	return m.subscriptions, nil
}

func (m *mockSubscriptionRepo) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Subscription, error) {
	return m.subscription, m.findErr
}

func (m *mockSubscriptionRepo) FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, domain string) (*entity.Subscription, error) {
	return m.subscription, m.findErr
}

func (m *mockSubscriptionRepo) FindByRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.Subscription, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	// Filter subscriptions by risk state
	var filtered []*entity.Subscription
	for _, sub := range m.subscriptions {
		if sub.RiskState == riskState {
			filtered = append(filtered, sub)
		}
	}
	return filtered, nil
}

func (m *mockSubscriptionRepo) DeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	return nil
}

// Mock partner repo for subscription tests
type mockPartnerRepoForSub struct {
	account *entity.PartnerAccount
	findErr error
}

func (m *mockPartnerRepoForSub) Create(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForSub) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.findErr
}

func (m *mockPartnerRepoForSub) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.findErr
}

func (m *mockPartnerRepoForSub) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerRepoForSub) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForSub) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockPartnerRepoForSub) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	if m.account != nil {
		return []uuid.UUID{m.account.ID}, nil
	}
	return []uuid.UUID{}, nil
}

// Mock app repo for subscription tests
type mockAppRepoForSub struct {
	app                   *entity.App
	findErr               error
	expectedPartnerID     uuid.UUID // If set, only return app if partnerAccountID matches
	checkPartnerOwnership bool
}

func (m *mockAppRepoForSub) Create(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepoForSub) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return m.app, m.findErr
}

func (m *mockAppRepoForSub) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	return nil, nil
}

func (m *mockAppRepoForSub) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	// Simulate ownership check: only return app if it belongs to this partner
	if m.checkPartnerOwnership && m.app != nil && m.app.PartnerAccountID != partnerAccountID {
		return nil, errors.New("app not found")
	}
	return m.app, nil
}

func (m *mockAppRepoForSub) Update(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepoForSub) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// Helper to create a chi context with URL params
func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestSubscriptionHandler_List_Success(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	appID := uuid.New()
	numericAppID := "4599915" // Numeric ID used in URL
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/" + numericAppID,
	}

	now := time.Now()
	subscriptions := []*entity.Subscription{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/AppSubscription/123",
			MyshopifyDomain: "store1.myshopify.com",
			PlanName:        "Pro Plan",
			BasePriceCents:  2999,
			BillingInterval: valueobject.BillingIntervalMonthly,
			RiskState:       valueobject.RiskStateSafe,
			Status:          "ACTIVE",
			CreatedAt:       now,
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/AppSubscription/456",
			MyshopifyDomain: "store2.myshopify.com",
			PlanName:        "Basic Plan",
			BasePriceCents:  999,
			BillingInterval: valueobject.BillingIntervalMonthly,
			RiskState:       valueobject.RiskStateOneCycleMissed,
			Status:          "ACTIVE",
			CreatedAt:       now,
		},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	subs, ok := resp["subscriptions"].([]interface{})
	if !ok {
		t.Fatal("expected subscriptions array in response")
	}

	if len(subs) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subs))
	}

	// Check total count
	total, ok := resp["total"].(float64)
	if !ok || int(total) != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestSubscriptionHandler_List_FilterByRiskState(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	appID := uuid.New()
	numericAppID := "4599915"
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		PartnerAppID:     "gid://partners/App/" + numericAppID,
	}

	now := time.Now()
	subscriptions := []*entity.Subscription{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/AppSubscription/123",
			MyshopifyDomain: "store1.myshopify.com",
			PlanName:        "Pro Plan",
			BasePriceCents:  2999,
			BillingInterval: valueobject.BillingIntervalMonthly,
			RiskState:       valueobject.RiskStateSafe,
			Status:          "ACTIVE",
			CreatedAt:       now,
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/AppSubscription/456",
			MyshopifyDomain: "store2.myshopify.com",
			PlanName:        "Basic Plan",
			BasePriceCents:  999,
			BillingInterval: valueobject.BillingIntervalMonthly,
			RiskState:       valueobject.RiskStateOneCycleMissed,
			Status:          "ACTIVE",
			CreatedAt:       now,
		},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Filter for SAFE only
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?risk_state=SAFE", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	subs, ok := resp["subscriptions"].([]interface{})
	if !ok {
		t.Fatal("expected subscriptions array in response")
	}

	if len(subs) != 1 {
		t.Errorf("expected 1 subscription (filtered by SAFE), got %d", len(subs))
	}
}

func TestSubscriptionHandler_List_NoUser(t *testing.T) {
	handler := NewSubscriptionHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/4599915/subscriptions", nil)
	req = withURLParam(req, "appID", "4599915")

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSubscriptionHandler_List_NoPartnerAccount(t *testing.T) {
	partnerRepo := &mockPartnerRepoForSub{findErr: errors.New("not found")}
	handler := NewSubscriptionHandler(nil, partnerRepo, nil)

	numericAppID := "4599915"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_List_UnauthorizedApp(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	numericAppID := "4599915"

	// App belongs to different partner account
	app := &entity.App{
		ID:               uuid.New(),
		PartnerAccountID: uuid.New(), // Different partner
		PartnerAppID:     "gid://partners/App/" + numericAppID,
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	// Enable ownership check - FindByPartnerAppID will return error if app doesn't belong to partner
	appRepo := &mockAppRepoForSub{app: app, checkPartnerOwnership: true}
	handler := NewSubscriptionHandler(nil, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	// With FindByPartnerAppID, unauthorized apps return 404 (not found for this partner)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_List_AppNotFound(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{findErr: errors.New("not found")}
	handler := NewSubscriptionHandler(nil, partnerRepo, appRepo)

	numericAppID := "4599915"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_Success(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	appID := uuid.New()
	numericAppID := "4599915"
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		PartnerAppID:     "gid://partners/App/" + numericAppID,
	}

	subID := uuid.New()
	now := time.Now()
	nextCharge := now.Add(30 * 24 * time.Hour)
	subscription := &entity.Subscription{
		ID:                     subID,
		AppID:                  appID,
		ShopifyGID:             "gid://shopify/AppSubscription/123",
		MyshopifyDomain:        "store.myshopify.com",
		PlanName:               "Pro Plan",
		BasePriceCents:         2999,
		BillingInterval:        valueobject.BillingIntervalMonthly,
		RiskState:              valueobject.RiskStateSafe,
		Status:                 "ACTIVE",
		CreatedAt:              now,
		ExpectedNextChargeDate: &nextCharge,
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscription: subscription}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions/"+subID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", numericAppID)
	rctx.URLParams.Add("subscriptionID", subID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	sub, ok := resp["subscription"].(map[string]interface{})
	if !ok {
		t.Fatal("expected subscription object in response")
	}

	if sub["id"] != subID.String() {
		t.Errorf("expected subscription id %s, got %v", subID.String(), sub["id"])
	}

	if sub["plan_name"] != "Pro Plan" {
		t.Errorf("expected plan_name 'Pro Plan', got %v", sub["plan_name"])
	}
}

func TestSubscriptionHandler_GetByID_NotFound(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	appID := uuid.New()
	numericAppID := "4599915"
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		PartnerAppID:     "gid://partners/App/" + numericAppID,
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{findErr: errors.New("not found")}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	subID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions/"+subID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", numericAppID)
	rctx.URLParams.Add("subscriptionID", subID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_WrongApp(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}

	appID := uuid.New()
	numericAppID := "4599915"
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		PartnerAppID:     "gid://partners/App/" + numericAppID,
	}

	// Subscription belongs to a different app
	differentAppID := uuid.New()
	subscription := &entity.Subscription{
		ID:        uuid.New(),
		AppID:     differentAppID, // Different app
		RiskState: valueobject.RiskStateSafe,
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscription: subscription}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	subID := subscription.ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions/"+subID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", numericAppID)
	rctx.URLParams.Add("subscriptionID", subID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestSubscriptionHandler_GetByID_NoUser(t *testing.T) {
	handler := NewSubscriptionHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/4599915/subscriptions/"+uuid.New().String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", "4599915")
	rctx.URLParams.Add("subscriptionID", uuid.New().String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

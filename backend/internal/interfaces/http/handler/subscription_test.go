package handler

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock subscription repository
type mockSubscriptionRepo struct {
	subscriptions []*entity.Subscription
	subscription  *entity.Subscription
	summary       *repository.SubscriptionSummary
	priceStats    *repository.PriceStats
	findErr       error
	findAllErr    error
	summaryErr    error
	priceStatsErr error
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

func (m *mockSubscriptionRepo) FindWithFilters(ctx context.Context, appID uuid.UUID, filters repository.SubscriptionFilters) (*repository.SubscriptionPage, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}

	// Apply filters to subscriptions
	filtered := m.subscriptions

	// Filter by risk states
	if len(filters.RiskStates) > 0 {
		var result []*entity.Subscription
		for _, sub := range filtered {
			for _, rs := range filters.RiskStates {
				if sub.RiskState == rs {
					result = append(result, sub)
					break
				}
			}
		}
		filtered = result
	}

	// Filter by price range
	if filters.PriceMinCents != nil {
		var result []*entity.Subscription
		for _, sub := range filtered {
			if sub.BasePriceCents >= *filters.PriceMinCents {
				result = append(result, sub)
			}
		}
		filtered = result
	}
	if filters.PriceMaxCents != nil {
		var result []*entity.Subscription
		for _, sub := range filtered {
			if sub.BasePriceCents <= *filters.PriceMaxCents {
				result = append(result, sub)
			}
		}
		filtered = result
	}

	// Filter by billing interval
	if filters.BillingInterval != nil {
		var result []*entity.Subscription
		for _, sub := range filtered {
			if sub.BillingInterval == *filters.BillingInterval {
				result = append(result, sub)
			}
		}
		filtered = result
	}

	// Filter by search term
	if filters.SearchTerm != "" {
		searchLower := strings.ToLower(filters.SearchTerm)
		var result []*entity.Subscription
		for _, sub := range filtered {
			if strings.Contains(strings.ToLower(sub.ShopName), searchLower) ||
				strings.Contains(strings.ToLower(sub.MyshopifyDomain), searchLower) {
				result = append(result, sub)
			}
		}
		filtered = result
	}

	// Apply pagination
	total := len(filtered)
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return &repository.SubscriptionPage{
		Subscriptions: filtered[start:end],
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

func (m *mockSubscriptionRepo) GetSummary(ctx context.Context, appID uuid.UUID) (*repository.SubscriptionSummary, error) {
	if m.summaryErr != nil {
		return nil, m.summaryErr
	}
	if m.summary != nil {
		return m.summary, nil
	}

	// Calculate from subscriptions
	var activeCount, atRiskCount, churnedCount, totalCents int64
	for _, sub := range m.subscriptions {
		switch sub.RiskState {
		case valueobject.RiskStateSafe:
			activeCount++
		case valueobject.RiskStateOneCycleMissed, valueobject.RiskStateTwoCyclesMissed:
			atRiskCount++
		case valueobject.RiskStateChurned:
			churnedCount++
		}
		totalCents += sub.BasePriceCents
	}

	var avgPrice int64
	if len(m.subscriptions) > 0 {
		avgPrice = totalCents / int64(len(m.subscriptions))
	}

	return &repository.SubscriptionSummary{
		ActiveCount:   int(activeCount),
		AtRiskCount:   int(atRiskCount),
		ChurnedCount:  int(churnedCount),
		AvgPriceCents: avgPrice,
		TotalCount:    len(m.subscriptions),
	}, nil
}

func (m *mockSubscriptionRepo) GetPriceStats(ctx context.Context, appID uuid.UUID) (*repository.PriceStats, error) {
	if m.priceStatsErr != nil {
		return nil, m.priceStatsErr
	}
	if m.priceStats != nil {
		return m.priceStats, nil
	}

	// Return default stats for testing
	return &repository.PriceStats{
		MinCents: 499,
		MaxCents: 9999,
		AvgCents: 3999,
		Prices: []repository.PricePoint{
			{PriceCents: 499, Count: 10},
			{PriceCents: 2999, Count: 25},
			{PriceCents: 4999, Count: 15},
			{PriceCents: 9999, Count: 5},
		},
	}, nil
}

func (m *mockSubscriptionRepo) SoftDeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	return nil
}

func (m *mockSubscriptionRepo) FindDeletedByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	return nil, nil
}

func (m *mockSubscriptionRepo) RestoreByID(ctx context.Context, id uuid.UUID) error {
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

func (m *mockAppRepoForSub) FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error) {
	if m.app != nil {
		return []*entity.App{m.app}, nil
	}
	return nil, m.findErr
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

// Tests for Summary endpoint
func TestSubscriptionHandler_Summary_Success(t *testing.T) {
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

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, BasePriceCents: 2999, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 4999, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 1999, RiskState: valueobject.RiskStateOneCycleMissed},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 999, RiskState: valueobject.RiskStateChurned},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions/summary", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.Summary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	if int(resp["activeCount"].(float64)) != 2 {
		t.Errorf("expected activeCount 2, got %v", resp["activeCount"])
	}
	if int(resp["atRiskCount"].(float64)) != 1 {
		t.Errorf("expected atRiskCount 1, got %v", resp["atRiskCount"])
	}
	if int(resp["churnedCount"].(float64)) != 1 {
		t.Errorf("expected churnedCount 1, got %v", resp["churnedCount"])
	}
	if int(resp["totalCount"].(float64)) != 4 {
		t.Errorf("expected totalCount 4, got %v", resp["totalCount"])
	}
}

func TestSubscriptionHandler_Summary_NoUser(t *testing.T) {
	handler := NewSubscriptionHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/4599915/subscriptions/summary", nil)
	req = withURLParam(req, "appID", "4599915")

	rec := httptest.NewRecorder()
	handler.Summary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

// Tests for PriceStats endpoint
func TestSubscriptionHandler_PriceStats_Success(t *testing.T) {
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
	subRepo := &mockSubscriptionRepo{} // Uses default stats

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions/price-stats", nil)
	req = withURLParam(req, "appID", numericAppID)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.PriceStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	// Verify price stats fields
	if _, ok := resp["minCents"]; !ok {
		t.Error("expected minCents in response")
	}
	if _, ok := resp["maxCents"]; !ok {
		t.Error("expected maxCents in response")
	}
	if _, ok := resp["avgCents"]; !ok {
		t.Error("expected avgCents in response")
	}

	// Check values from mock
	if resp["minCents"].(float64) != 499 {
		t.Errorf("expected minCents 499, got %v", resp["minCents"])
	}

	// Verify prices array
	prices, ok := resp["prices"].([]interface{})
	if !ok {
		t.Fatal("expected prices array in response")
	}
	if len(prices) != 4 {
		t.Errorf("expected 4 price points, got %d", len(prices))
	}

	// Check first price
	firstPrice := prices[0].(map[string]interface{})
	if firstPrice["priceCents"].(float64) != 499 {
		t.Errorf("expected first price 499, got %v", firstPrice["priceCents"])
	}
	if firstPrice["count"].(float64) != 10 {
		t.Errorf("expected first count 10, got %v", firstPrice["count"])
	}
}

func TestSubscriptionHandler_PriceStats_NoUser(t *testing.T) {
	handler := NewSubscriptionHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/4599915/subscriptions/price-stats", nil)
	req = withURLParam(req, "appID", "4599915")

	rec := httptest.NewRecorder()
	handler.PriceStats(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

// Tests for enhanced List filtering
func TestSubscriptionHandler_List_MultipleStatusFilter(t *testing.T) {
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

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, RiskState: valueobject.RiskStateOneCycleMissed},
		{ID: uuid.New(), AppID: appID, RiskState: valueobject.RiskStateTwoCyclesMissed},
		{ID: uuid.New(), AppID: appID, RiskState: valueobject.RiskStateChurned},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Filter for SAFE and ONE_CYCLE_MISSED
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?status=SAFE,ONE_CYCLE_MISSED", nil)
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

	total := int(resp["total"].(float64))
	if total != 2 {
		t.Errorf("expected total 2 (filtered), got %d", total)
	}
}

func TestSubscriptionHandler_List_PriceRangeFilter(t *testing.T) {
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

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, BasePriceCents: 1000, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 2500, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 5000, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 10000, RiskState: valueobject.RiskStateSafe},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Filter for $20-$60 (2000-6000 cents)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?priceMin=2000&priceMax=6000", nil)
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

	total := int(resp["total"].(float64))
	if total != 2 {
		t.Errorf("expected total 2 (filtered by price), got %d", total)
	}
}

func TestSubscriptionHandler_List_BillingIntervalFilter(t *testing.T) {
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

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, BillingInterval: valueobject.BillingIntervalMonthly, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BillingInterval: valueobject.BillingIntervalMonthly, RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, BillingInterval: valueobject.BillingIntervalAnnual, RiskState: valueobject.RiskStateSafe},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Filter for ANNUAL only
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?billingInterval=ANNUAL", nil)
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

	total := int(resp["total"].(float64))
	if total != 1 {
		t.Errorf("expected total 1 (ANNUAL only), got %d", total)
	}
}

func TestSubscriptionHandler_List_SearchFilter(t *testing.T) {
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

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, ShopName: "Acme Store", MyshopifyDomain: "acme.myshopify.com", RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, ShopName: "Best Shop", MyshopifyDomain: "best.myshopify.com", RiskState: valueobject.RiskStateSafe},
		{ID: uuid.New(), AppID: appID, ShopName: "Cool Store", MyshopifyDomain: "cool.myshopify.com", RiskState: valueobject.RiskStateSafe},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Search for "acme"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?search=acme", nil)
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

	total := int(resp["total"].(float64))
	if total != 1 {
		t.Errorf("expected total 1 (search 'acme'), got %d", total)
	}
}

func TestSubscriptionHandler_List_Pagination(t *testing.T) {
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

	// Create 30 subscriptions
	subscriptions := make([]*entity.Subscription, 30)
	for i := 0; i < 30; i++ {
		subscriptions[i] = &entity.Subscription{
			ID:        uuid.New(),
			AppID:     appID,
			RiskState: valueobject.RiskStateSafe,
		}
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Request page 2 with pageSize 10
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?page=2&pageSize=10", nil)
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

	page := int(resp["page"].(float64))
	pageSize := int(resp["pageSize"].(float64))
	total := int(resp["total"].(float64))
	totalPages := int(resp["totalPages"].(float64))

	if page != 2 {
		t.Errorf("expected page 2, got %d", page)
	}
	if pageSize != 10 {
		t.Errorf("expected pageSize 10, got %d", pageSize)
	}
	if total != 30 {
		t.Errorf("expected total 30, got %d", total)
	}
	if totalPages != 3 {
		t.Errorf("expected totalPages 3, got %d", totalPages)
	}

	subs := resp["subscriptions"].([]interface{})
	if len(subs) != 10 {
		t.Errorf("expected 10 subscriptions on page 2, got %d", len(subs))
	}
}

func TestSubscriptionHandler_List_CombinedFilters(t *testing.T) {
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

	subscriptions := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, BasePriceCents: 5000, BillingInterval: valueobject.BillingIntervalMonthly, RiskState: valueobject.RiskStateSafe, ShopName: "Acme"},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 5000, BillingInterval: valueobject.BillingIntervalAnnual, RiskState: valueobject.RiskStateSafe, ShopName: "Beta"},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 10000, BillingInterval: valueobject.BillingIntervalMonthly, RiskState: valueobject.RiskStateSafe, ShopName: "Acme Plus"},
		{ID: uuid.New(), AppID: appID, BasePriceCents: 5000, BillingInterval: valueobject.BillingIntervalMonthly, RiskState: valueobject.RiskStateChurned, ShopName: "Acme Old"},
	}

	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	subRepo := &mockSubscriptionRepo{subscriptions: subscriptions}

	handler := NewSubscriptionHandler(subRepo, partnerRepo, appRepo)

	// Combine filters: SAFE, MONTHLY, price <= 6000, search "acme"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+numericAppID+"/subscriptions?status=SAFE&billingInterval=MONTHLY&priceMax=6000&search=acme", nil)
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

	total := int(resp["total"].(float64))
	if total != 1 {
		t.Errorf("expected total 1 (combined filters), got %d", total)
	}
}

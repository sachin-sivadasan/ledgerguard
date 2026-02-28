package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// mockRevenueRepository implements repository.RevenueRepository for testing
type mockRevenueRepository struct {
	aggregations []repository.RevenueAggregation
	err          error
}

func (m *mockRevenueRepository) GetRevenueByDateRange(ctx context.Context, appID uuid.UUID, startDate, endDate time.Time) ([]repository.RevenueAggregation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.aggregations, nil
}

// mockPartnerRepoForRevenue implements just what we need from PartnerAccountRepository
type mockPartnerRepoForRevenue struct {
	account *entity.PartnerAccount
	err     error
}

func (m *mockPartnerRepoForRevenue) Create(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForRevenue) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockPartnerRepoForRevenue) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockPartnerRepoForRevenue) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockPartnerRepoForRevenue) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForRevenue) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockPartnerRepoForRevenue) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

// mockAppRepoForRevenue implements just what we need from AppRepository
type mockAppRepoForRevenue struct {
	apps []*entity.App
	err  error
}

func (m *mockAppRepoForRevenue) Create(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepoForRevenue) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	if len(m.apps) > 0 {
		return m.apps[0], nil
	}
	return nil, m.err
}

func (m *mockAppRepoForRevenue) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	return m.apps, m.err
}

func (m *mockAppRepoForRevenue) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	for _, app := range m.apps {
		if app.PartnerAppID == partnerAppID {
			return app, nil
		}
	}
	return nil, m.err
}

func (m *mockAppRepoForRevenue) Update(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepoForRevenue) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestRevenueHandler_GetEarnings_Success(t *testing.T) {
	// Setup
	userID := uuid.New()
	partnerID := uuid.New()
	appID := uuid.New()

	partnerRepo := &mockPartnerRepoForRevenue{
		account: &entity.PartnerAccount{
			ID:     partnerID,
			UserID: userID,
		},
	}

	appRepo := &mockAppRepoForRevenue{
		apps: []*entity.App{
			{
				ID:               appID,
				PartnerAccountID: partnerID,
				PartnerAppID:     "gid://partners/App/12345",
				Name:             "Test App",
			},
		},
	}

	revenueRepo := &mockRevenueRepository{
		aggregations: []repository.RevenueAggregation{
			{
				Date:                    "2024-01-15",
				TotalAmountCents:        10000,
				SubscriptionAmountCents: 7000,
				UsageAmountCents:        3000,
			},
			{
				Date:                    "2024-01-16",
				TotalAmountCents:        12000,
				SubscriptionAmountCents: 8000,
				UsageAmountCents:        4000,
			},
		},
	}

	revenueSvc := service.NewRevenueMetricsService(revenueRepo)
	handler := NewRevenueHandler(revenueSvc, partnerRepo, appRepo)

	// Create request with chi router context
	req := httptest.NewRequest("GET", "/api/v1/apps/12345/earnings?start=2024-01-01&end=2024-01-31&mode=split", nil)

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", "12345")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add user to context
	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{
		ID:    userID,
		Email: "test@example.com",
	}))

	// Execute
	rr := httptest.NewRecorder()
	handler.GetEarnings(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response service.EarningsTimelineResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.StartDate != "2024-01-01" {
		t.Errorf("expected start_date 2024-01-01, got %s", response.StartDate)
	}

	if len(response.Earnings) != 2 {
		t.Errorf("expected 2 earnings entries, got %d", len(response.Earnings))
	}

	// Check split mode includes subscription and usage amounts
	if response.Earnings[0].SubscriptionAmountCents != 7000 {
		t.Errorf("expected subscription amount 7000, got %d", response.Earnings[0].SubscriptionAmountCents)
	}
	if response.Earnings[0].UsageAmountCents != 3000 {
		t.Errorf("expected usage amount 3000, got %d", response.Earnings[0].UsageAmountCents)
	}
}

func TestRevenueHandler_GetEarnings_CombinedMode(t *testing.T) {
	// Setup
	userID := uuid.New()
	partnerID := uuid.New()
	appID := uuid.New()

	partnerRepo := &mockPartnerRepoForRevenue{
		account: &entity.PartnerAccount{
			ID:     partnerID,
			UserID: userID,
		},
	}

	appRepo := &mockAppRepoForRevenue{
		apps: []*entity.App{
			{
				ID:               appID,
				PartnerAccountID: partnerID,
				PartnerAppID:     "gid://partners/App/12345",
				Name:             "Test App",
			},
		},
	}

	revenueRepo := &mockRevenueRepository{
		aggregations: []repository.RevenueAggregation{
			{
				Date:                    "2024-01-15",
				TotalAmountCents:        10000,
				SubscriptionAmountCents: 7000,
				UsageAmountCents:        3000,
			},
		},
	}

	revenueSvc := service.NewRevenueMetricsService(revenueRepo)
	handler := NewRevenueHandler(revenueSvc, partnerRepo, appRepo)

	// Create request with combined mode (default)
	req := httptest.NewRequest("GET", "/api/v1/apps/12345/earnings?start=2024-01-01&end=2024-01-31", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", "12345")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{
		ID:    userID,
		Email: "test@example.com",
	}))

	// Execute
	rr := httptest.NewRecorder()
	handler.GetEarnings(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response service.EarningsTimelineResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// In combined mode, subscription and usage should be omitted (zero values)
	if response.Earnings[0].SubscriptionAmountCents != 0 {
		t.Errorf("expected subscription amount 0 in combined mode, got %d", response.Earnings[0].SubscriptionAmountCents)
	}
	if response.Earnings[0].UsageAmountCents != 0 {
		t.Errorf("expected usage amount 0 in combined mode, got %d", response.Earnings[0].UsageAmountCents)
	}
	if response.Earnings[0].TotalAmountCents != 10000 {
		t.Errorf("expected total amount 10000, got %d", response.Earnings[0].TotalAmountCents)
	}
}

func TestRevenueHandler_GetEarnings_Unauthorized(t *testing.T) {
	handler := NewRevenueHandler(nil, nil, nil)

	req := httptest.NewRequest("GET", "/api/v1/apps/12345/earnings?start=2024-01-01&end=2024-01-31", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", "12345")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// No user in context

	rr := httptest.NewRecorder()
	handler.GetEarnings(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestRevenueHandler_GetEarnings_MissingParams(t *testing.T) {
	userID := uuid.New()
	partnerID := uuid.New()
	appID := uuid.New()

	partnerRepo := &mockPartnerRepoForRevenue{
		account: &entity.PartnerAccount{
			ID:     partnerID,
			UserID: userID,
		},
	}

	appRepo := &mockAppRepoForRevenue{
		apps: []*entity.App{
			{
				ID:               appID,
				PartnerAccountID: partnerID,
				PartnerAppID:     "gid://partners/App/12345",
				Name:             "Test App",
			},
		},
	}

	revenueSvc := service.NewRevenueMetricsService(&mockRevenueRepository{})
	handler := NewRevenueHandler(revenueSvc, partnerRepo, appRepo)

	// Missing start and end dates
	req := httptest.NewRequest("GET", "/api/v1/apps/12345/earnings", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", "12345")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{
		ID:    userID,
		Email: "test@example.com",
	}))

	rr := httptest.NewRecorder()
	handler.GetEarnings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestRevenueHandler_GetEarnings_AppNotFound(t *testing.T) {
	userID := uuid.New()
	partnerID := uuid.New()

	partnerRepo := &mockPartnerRepoForRevenue{
		account: &entity.PartnerAccount{
			ID:     partnerID,
			UserID: userID,
		},
	}

	// Empty app list - app not found
	appRepo := &mockAppRepoForRevenue{
		apps: []*entity.App{},
	}

	revenueSvc := service.NewRevenueMetricsService(&mockRevenueRepository{})
	handler := NewRevenueHandler(revenueSvc, partnerRepo, appRepo)

	req := httptest.NewRequest("GET", "/api/v1/apps/99999/earnings?start=2024-01-01&end=2024-01-31", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("appID", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	req = req.WithContext(middleware.SetUserContext(req.Context(), &entity.User{
		ID:    userID,
		Email: "test@example.com",
	}))

	rr := httptest.NewRecorder()
	handler.GetEarnings(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

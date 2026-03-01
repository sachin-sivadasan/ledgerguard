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
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock implementations for sync tests
type mockSyncTransactionFetcher struct {
	transactions []*entity.Transaction
	err          error
}

func (m *mockSyncTransactionFetcher) FetchTransactions(ctx context.Context, accessToken string, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.err
}

type mockSyncTransactionRepo struct {
	err error
}

func (m *mockSyncTransactionRepo) Upsert(ctx context.Context, tx *entity.Transaction) error {
	return m.err
}

func (m *mockSyncTransactionRepo) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	return m.err
}

func (m *mockSyncTransactionRepo) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return nil, nil
}

func (m *mockSyncTransactionRepo) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Transaction, error) {
	return nil, nil
}

func (m *mockSyncTransactionRepo) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *mockSyncTransactionRepo) GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}

func (m *mockSyncTransactionRepo) GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]repository.EarningsByDate, error) {
	return nil, nil
}

func (m *mockSyncTransactionRepo) GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]repository.EarningsByDate, error) {
	return nil, nil
}

func (m *mockSyncTransactionRepo) FindByDomain(ctx context.Context, appID uuid.UUID, domain string, from, to time.Time) ([]*entity.Transaction, error) {
	return nil, nil
}

func (m *mockSyncTransactionRepo) GetEarningsSummaryByDomain(ctx context.Context, appID uuid.UUID, domain string) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}

type mockSyncAppRepo struct {
	app  *entity.App
	apps []*entity.App
	err  error
}

func (m *mockSyncAppRepo) Create(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockSyncAppRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return m.app, m.err
}

func (m *mockSyncAppRepo) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	return m.apps, m.err
}

func (m *mockSyncAppRepo) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	return m.app, m.err
}

func (m *mockSyncAppRepo) Update(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockSyncAppRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockSyncAppRepo) FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error) {
	if m.app != nil {
		return []*entity.App{m.app}, nil
	}
	return m.apps, m.err
}

type mockSyncPartnerRepo struct {
	account *entity.PartnerAccount
	err     error
}

func (m *mockSyncPartnerRepo) Create(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockSyncPartnerRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockSyncPartnerRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockSyncPartnerRepo) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockSyncPartnerRepo) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockSyncPartnerRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockSyncPartnerRepo) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	if m.account != nil {
		return []uuid.UUID{m.account.ID}, nil
	}
	return []uuid.UUID{}, nil
}

type mockSyncDecryptor struct {
	decrypted []byte
	err       error
}

func (m *mockSyncDecryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	return m.decrypted, m.err
}

type mockSyncLedgerRebuilder struct {
	result *domainservice.LedgerRebuildResult
	err    error
}

func (m *mockSyncLedgerRebuilder) RebuildFromTransactions(ctx context.Context, appID uuid.UUID, now time.Time) (*domainservice.LedgerRebuildResult, error) {
	if m.result != nil {
		return m.result, m.err
	}
	return &domainservice.LedgerRebuildResult{
		AppID:                appID,
		SubscriptionsUpdated: 0,
		TotalMRRCents:        0,
		TotalUsageCents:      0,
		RiskSummary:          domainservice.RiskSummary{},
		RebuildAt:            now,
	}, m.err
}

func (m *mockSyncLedgerRebuilder) BackfillHistoricalSnapshots(ctx context.Context, appID uuid.UUID, transactions []*entity.Transaction) (int, error) {
	return 0, m.err
}

func TestSyncHandler_SyncAllApps_Success(t *testing.T) {
	partnerAccountID := uuid.New()
	appID := uuid.New()

	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		UserID:               uuid.New(),
		EncryptedAccessToken: []byte("encrypted"),
	}

	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
		Name:             "Test App",
		TrackingEnabled:  true,
	}

	transactions := []*entity.Transaction{
		{
			ID:         uuid.New(),
			AppID:      appID,
			ShopifyGID: "gid://shopify/Transaction/1",
			ChargeType: valueobject.ChargeTypeRecurring,
		},
	}

	fetcher := &mockSyncTransactionFetcher{transactions: transactions}
	txRepo := &mockSyncTransactionRepo{}
	appRepo := &mockSyncAppRepo{app: app, apps: []*entity.App{app}}
	partnerRepo := &mockSyncPartnerRepo{account: partnerAccount}
	decryptor := &mockSyncDecryptor{decrypted: []byte("token")}
	ledger := &mockSyncLedgerRebuilder{}

	syncService := service.NewSyncService(fetcher, txRepo, appRepo, partnerRepo, decryptor, ledger)
	handler := NewSyncHandler(syncService, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", nil)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.SyncAllApps(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	results, ok := resp["results"].([]interface{})
	if !ok {
		t.Fatal("expected results array in response")
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSyncHandler_SyncAllApps_NoUser(t *testing.T) {
	handler := NewSyncHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", nil)
	rec := httptest.NewRecorder()
	handler.SyncAllApps(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSyncHandler_SyncAllApps_NoPartnerAccount(t *testing.T) {
	partnerRepo := &mockSyncPartnerRepo{err: errors.New("not found")}
	handler := NewSyncHandler(nil, partnerRepo, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync", nil)
	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.SyncAllApps(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestSyncHandler_SyncApp_Success(t *testing.T) {
	partnerAccountID := uuid.New()
	appID := uuid.New()

	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		UserID:               uuid.New(),
		EncryptedAccessToken: []byte("encrypted"),
	}

	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
		Name:             "Test App",
		TrackingEnabled:  true,
	}

	transactions := []*entity.Transaction{
		{
			ID:         uuid.New(),
			AppID:      appID,
			ShopifyGID: "gid://shopify/Transaction/1",
			ChargeType: valueobject.ChargeTypeRecurring,
		},
	}

	fetcher := &mockSyncTransactionFetcher{transactions: transactions}
	txRepo := &mockSyncTransactionRepo{}
	appRepo := &mockSyncAppRepo{app: app}
	partnerRepo := &mockSyncPartnerRepo{account: partnerAccount}
	decryptor := &mockSyncDecryptor{decrypted: []byte("token")}
	ledger := &mockSyncLedgerRebuilder{}

	syncService := service.NewSyncService(fetcher, txRepo, appRepo, partnerRepo, decryptor, ledger)
	handler := NewSyncHandler(syncService, partnerRepo, appRepo)

	// Create router to handle URL params
	r := chi.NewRouter()
	r.Post("/api/v1/sync/{appID}", handler.SyncApp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/"+appID.String(), nil)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["app_name"] != "Test App" {
		t.Errorf("expected app_name 'Test App', got %v", resp["app_name"])
	}
}

func TestSyncHandler_SyncApp_NoUser(t *testing.T) {
	handler := NewSyncHandler(nil, nil, nil)

	r := chi.NewRouter()
	r.Post("/api/v1/sync/{appID}", handler.SyncApp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSyncHandler_SyncApp_InvalidAppID(t *testing.T) {
	partnerRepo := &mockSyncPartnerRepo{account: &entity.PartnerAccount{}}
	handler := NewSyncHandler(nil, partnerRepo, nil)

	r := chi.NewRouter()
	r.Post("/api/v1/sync/{appID}", handler.SyncApp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/invalid-uuid", nil)
	user := &entity.User{ID: uuid.New(), Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSyncHandler_SyncApp_TenantIsolation_Forbidden(t *testing.T) {
	// User A's partner account
	userAPartnerAccountID := uuid.New()
	userAPartnerAccount := &entity.PartnerAccount{
		ID:                   userAPartnerAccountID,
		UserID:               uuid.New(),
		EncryptedAccessToken: []byte("encrypted"),
	}

	// User B's app (belongs to different partner account)
	userBPartnerAccountID := uuid.New()
	appID := uuid.New()
	userBApp := &entity.App{
		ID:               appID,
		PartnerAccountID: userBPartnerAccountID, // Different partner account!
		Name:             "User B's App",
		TrackingEnabled:  true,
	}

	appRepo := &mockSyncAppRepo{app: userBApp}
	partnerRepo := &mockSyncPartnerRepo{account: userAPartnerAccount}

	handler := NewSyncHandler(nil, partnerRepo, appRepo)

	r := chi.NewRouter()
	r.Post("/api/v1/sync/{appID}", handler.SyncApp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/"+appID.String(), nil)
	userA := &entity.User{ID: userAPartnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), userA))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should be forbidden because User A is trying to sync User B's app
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d (Forbidden), got %d", http.StatusForbidden, rec.Code)
	}
}

func TestSyncHandler_SyncApp_AppNotFound(t *testing.T) {
	partnerAccountID := uuid.New()
	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		UserID:               uuid.New(),
		EncryptedAccessToken: []byte("encrypted"),
	}

	appRepo := &mockSyncAppRepo{err: errors.New("not found")}
	partnerRepo := &mockSyncPartnerRepo{account: partnerAccount}

	handler := NewSyncHandler(nil, partnerRepo, appRepo)

	r := chi.NewRouter()
	r.Post("/api/v1/sync/{appID}", handler.SyncApp)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/"+uuid.New().String(), nil)
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

package handler

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// mockTxRepo implements repository.TransactionRepository for handler tests.
type mockTxRepo struct {
	transactions []*entity.Transaction
	findErr      error
}

func (m *mockTxRepo) Upsert(ctx context.Context, tx *entity.Transaction) error { return nil }
func (m *mockTxRepo) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	return nil
}
func (m *mockTxRepo) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.findErr
}
func (m *mockTxRepo) FindByShopifyGID(ctx context.Context, gid string) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTxRepo) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	return int64(len(m.transactions)), nil
}
func (m *mockTxRepo) GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}
func (m *mockTxRepo) GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]repository.EarningsByDate, error) {
	return nil, nil
}
func (m *mockTxRepo) GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]repository.EarningsByDate, error) {
	return nil, nil
}
func (m *mockTxRepo) FindByDomain(ctx context.Context, appID uuid.UUID, domain string, from, to time.Time) ([]*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTxRepo) GetEarningsSummaryByDomain(ctx context.Context, appID uuid.UUID, domain string) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}
func (m *mockTxRepo) FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters repository.TransactionFilters) (*repository.TransactionPage, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	total := len(m.transactions)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return &repository.TransactionPage{
		Transactions: m.transactions[start:end],
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

func makeTxTestData(appID uuid.UUID, count int) []*entity.Transaction {
	txs := make([]*entity.Transaction, count)
	for i := 0; i < count; i++ {
		txs[i] = &entity.Transaction{
			ID:               uuid.New(),
			AppID:            appID,
			ShopifyGID:       "gid://shopify/AppTransaction/" + uuid.New().String(),
			MyshopifyDomain:  "store.myshopify.com",
			ChargeType:       valueobject.ChargeTypeRecurring,
			GrossAmountCents: 2999,
			NetAmountCents:   2549,
			Currency:         "USD",
			TransactionDate:  time.Now().AddDate(0, 0, -i),
			CreatedAt:        time.Now(),
		}
	}
	return txs
}

func TestTransactionHandler_List_DefaultPage(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/12345",
	}

	txs := makeTxTestData(appID, 50)
	txRepo := &mockTxRepo{transactions: txs}
	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}

	handler := NewTransactionHandler(txRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/transactions", nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	txList := resp["transactions"].([]interface{})
	if len(txList) != 20 {
		t.Errorf("expected 20 transactions (default pageSize), got %d", len(txList))
	}
	if resp["total"].(float64) != 50 {
		t.Errorf("expected total=50, got %v", resp["total"])
	}
	if resp["page"].(float64) != 1 {
		t.Errorf("expected page=1, got %v", resp["page"])
	}
	if resp["pageSize"].(float64) != 20 {
		t.Errorf("expected pageSize=20, got %v", resp["pageSize"])
	}
	if resp["totalPages"].(float64) != 3 {
		t.Errorf("expected totalPages=3, got %v", resp["totalPages"])
	}
}

func TestTransactionHandler_List_CustomPageSize(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/12345",
	}

	txs := makeTxTestData(appID, 50)
	txRepo := &mockTxRepo{transactions: txs}
	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}

	handler := NewTransactionHandler(txRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/transactions?page=2&pageSize=10", nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	txList := resp["transactions"].([]interface{})
	if len(txList) != 10 {
		t.Errorf("expected 10 transactions on page 2, got %d", len(txList))
	}
	if resp["page"].(float64) != 2 {
		t.Errorf("expected page=2, got %v", resp["page"])
	}
	if resp["totalPages"].(float64) != 5 {
		t.Errorf("expected totalPages=5, got %v", resp["totalPages"])
	}
}

func TestTransactionHandler_List_OutOfRangePage(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/12345",
	}

	txs := makeTxTestData(appID, 5)
	txRepo := &mockTxRepo{transactions: txs}
	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}

	handler := NewTransactionHandler(txRepo, partnerRepo, appRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/transactions?page=99", nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)

	txList := resp["transactions"].([]interface{})
	if len(txList) != 0 {
		t.Errorf("expected 0 transactions for out-of-range page, got %d", len(txList))
	}
	if resp["total"].(float64) != 5 {
		t.Errorf("expected total=5, got %v", resp["total"])
	}
}

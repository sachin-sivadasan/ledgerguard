package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock implementations
type mockTransactionFetcher struct {
	transactions []*entity.Transaction
	err          error
}

func (m *mockTransactionFetcher) FetchTransactions(ctx context.Context, accessToken string, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.err
}

type mockTransactionRepo struct {
	upsertCalls    int
	upsertBatchTxs []*entity.Transaction
	err            error
}

func (m *mockTransactionRepo) Upsert(ctx context.Context, tx *entity.Transaction) error {
	m.upsertCalls++
	return m.err
}

func (m *mockTransactionRepo) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	m.upsertBatchTxs = txs
	return m.err
}

func (m *mockTransactionRepo) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepo) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionRepo) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	return int64(len(m.upsertBatchTxs)), nil
}

type mockAppRepoForSync struct {
	app *entity.App
	err error
}

func (m *mockAppRepoForSync) Create(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepoForSync) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return m.app, m.err
}

func (m *mockAppRepoForSync) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	if m.app != nil {
		return []*entity.App{m.app}, nil
	}
	return nil, m.err
}

func (m *mockAppRepoForSync) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	return m.app, m.err
}

func (m *mockAppRepoForSync) Update(ctx context.Context, app *entity.App) error {
	return nil
}

func (m *mockAppRepoForSync) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockPartnerRepoForSync struct {
	account *entity.PartnerAccount
	err     error
}

func (m *mockPartnerRepoForSync) Create(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForSync) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockPartnerRepoForSync) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}

func (m *mockPartnerRepoForSync) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return nil, nil
}

func (m *mockPartnerRepoForSync) Update(ctx context.Context, account *entity.PartnerAccount) error {
	return nil
}

func (m *mockPartnerRepoForSync) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}

type mockDecryptorForSync struct {
	decrypted []byte
	err       error
}

func (m *mockDecryptorForSync) Decrypt(ciphertext []byte) ([]byte, error) {
	return m.decrypted, m.err
}

func TestSyncService_SyncApp_Success(t *testing.T) {
	appID := uuid.New()
	partnerAccountID := uuid.New()

	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
		PartnerAppID:     "gid://partners/App/123",
		Name:             "Test App",
	}

	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		PartnerID:            "org123",
		EncryptedAccessToken: []byte("encrypted"),
	}

	transactions := []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			AmountCents:     2999,
			Currency:        "USD",
			TransactionDate: time.Now(),
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/2",
			MyshopifyDomain: "store2.myshopify.com",
			ChargeType:      valueobject.ChargeTypeUsage,
			AmountCents:     500,
			Currency:        "USD",
			TransactionDate: time.Now(),
		},
	}

	fetcher := &mockTransactionFetcher{transactions: transactions}
	txRepo := &mockTransactionRepo{}
	appRepo := &mockAppRepoForSync{app: app}
	partnerRepo := &mockPartnerRepoForSync{account: partnerAccount}
	decryptor := &mockDecryptorForSync{decrypted: []byte("decrypted-token")}

	service := NewSyncService(fetcher, txRepo, appRepo, partnerRepo, decryptor)

	result, err := service.SyncApp(context.Background(), appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TransactionCount != 2 {
		t.Errorf("expected 2 transactions, got %d", result.TransactionCount)
	}

	if len(txRepo.upsertBatchTxs) != 2 {
		t.Errorf("expected 2 upsert calls, got %d", len(txRepo.upsertBatchTxs))
	}
}

func TestSyncService_SyncApp_AppNotFound(t *testing.T) {
	appRepo := &mockAppRepoForSync{err: errors.New("not found")}
	service := NewSyncService(nil, nil, appRepo, nil, nil)

	_, err := service.SyncApp(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSyncService_SyncApp_FetchError(t *testing.T) {
	appID := uuid.New()
	partnerAccountID := uuid.New()

	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
	}

	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		EncryptedAccessToken: []byte("encrypted"),
	}

	fetcher := &mockTransactionFetcher{err: errors.New("API error")}
	appRepo := &mockAppRepoForSync{app: app}
	partnerRepo := &mockPartnerRepoForSync{account: partnerAccount}
	decryptor := &mockDecryptorForSync{decrypted: []byte("token")}

	service := NewSyncService(fetcher, nil, appRepo, partnerRepo, decryptor)

	_, err := service.SyncApp(context.Background(), appID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSyncService_SyncApp_NoTransactions(t *testing.T) {
	appID := uuid.New()
	partnerAccountID := uuid.New()

	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
	}

	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		EncryptedAccessToken: []byte("encrypted"),
	}

	fetcher := &mockTransactionFetcher{transactions: []*entity.Transaction{}}
	txRepo := &mockTransactionRepo{}
	appRepo := &mockAppRepoForSync{app: app}
	partnerRepo := &mockPartnerRepoForSync{account: partnerAccount}
	decryptor := &mockDecryptorForSync{decrypted: []byte("token")}

	service := NewSyncService(fetcher, txRepo, appRepo, partnerRepo, decryptor)

	result, err := service.SyncApp(context.Background(), appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TransactionCount != 0 {
		t.Errorf("expected 0 transactions, got %d", result.TransactionCount)
	}
}

func TestSyncService_SyncAllApps(t *testing.T) {
	partnerAccountID := uuid.New()
	appID := uuid.New()

	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccountID,
		TrackingEnabled:  true,
	}

	partnerAccount := &entity.PartnerAccount{
		ID:                   partnerAccountID,
		EncryptedAccessToken: []byte("encrypted"),
	}

	transactions := []*entity.Transaction{
		{
			ID:         uuid.New(),
			AppID:      appID,
			ShopifyGID: "gid://shopify/Transaction/1",
			ChargeType: valueobject.ChargeTypeRecurring,
		},
	}

	fetcher := &mockTransactionFetcher{transactions: transactions}
	txRepo := &mockTransactionRepo{}
	appRepo := &mockAppRepoForSync{app: app}
	partnerRepo := &mockPartnerRepoForSync{account: partnerAccount}
	decryptor := &mockDecryptorForSync{decrypted: []byte("token")}

	service := NewSyncService(fetcher, txRepo, appRepo, partnerRepo, decryptor)

	results, err := service.SyncAllApps(context.Background(), partnerAccountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

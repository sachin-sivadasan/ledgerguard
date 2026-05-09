package processors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// ==================== Mocks ====================

// --- SyncJobRepo ---

type mockSyncJobRepo struct {
	jobs     map[uuid.UUID]*entity.SyncJob
	created  []*entity.SyncJob
	failedID uuid.UUID
	failMsg  string
}

func newMockSyncJobRepo() *mockSyncJobRepo {
	return &mockSyncJobRepo{jobs: make(map[uuid.UUID]*entity.SyncJob)}
}

func (m *mockSyncJobRepo) Create(_ context.Context, job *entity.SyncJob) error {
	m.jobs[job.ID] = job
	m.created = append(m.created, job)
	return nil
}
func (m *mockSyncJobRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.SyncJob, error) {
	if j, ok := m.jobs[id]; ok {
		return j, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockSyncJobRepo) FindByStatus(_ context.Context, status entity.SyncJobStatus) ([]*entity.SyncJob, error) {
	var result []*entity.SyncJob
	for _, j := range m.jobs {
		if j.Status == status {
			result = append(result, j)
		}
	}
	return result, nil
}
func (m *mockSyncJobRepo) FindActiveByAppIDAndType(_ context.Context, _ uuid.UUID, _ string) (*entity.SyncJob, error) {
	return nil, nil
}
func (m *mockSyncJobRepo) FindByParentJobID(_ context.Context, parentID uuid.UUID) ([]*entity.SyncJob, error) {
	var result []*entity.SyncJob
	for _, j := range m.jobs {
		if j.ParentJobID != nil && *j.ParentJobID == parentID {
			result = append(result, j)
		}
	}
	return result, nil
}
func (m *mockSyncJobRepo) ListByAppID(_ context.Context, _ uuid.UUID, _ string, _ string, _, _ int) ([]*entity.SyncJob, int, error) {
	return nil, 0, nil
}
func (m *mockSyncJobRepo) UpdateStatus(_ context.Context, id uuid.UUID, status entity.SyncJobStatus) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = status
	}
	return nil
}
func (m *mockSyncJobRepo) UpdateProgress(_ context.Context, id uuid.UUID, total, completed int) error {
	if j, ok := m.jobs[id]; ok {
		j.TotalItems = total
		j.CompletedItems = completed
	}
	return nil
}
func (m *mockSyncJobRepo) MarkStarted(_ context.Context, id uuid.UUID, workerID string) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusProcessing
		j.WorkerID = workerID
	}
	return nil
}
func (m *mockSyncJobRepo) MarkCompleted(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusCompleted
		now := time.Now().UTC()
		j.CompletedAt = &now
	}
	return nil
}
func (m *mockSyncJobRepo) MarkFailed(_ context.Context, id uuid.UUID, errMsg string) error {
	m.failedID = id
	m.failMsg = errMsg
	if j, ok := m.jobs[id]; ok {
		j.Status = entity.SyncJobStatusFailed
		j.ErrorMessage = errMsg
	}
	return nil
}
func (m *mockSyncJobRepo) MarkPendingIfProcessing(_ context.Context, id uuid.UUID) error {
	if j, ok := m.jobs[id]; ok {
		if j.Status == entity.SyncJobStatusProcessing {
			j.Status = entity.SyncJobStatusPending
		}
	}
	return nil
}

// --- AppRepo ---

type mockAppRepo struct {
	apps map[uuid.UUID]*entity.App
}

func (m *mockAppRepo) Create(_ context.Context, _ *entity.App) error         { return nil }
func (m *mockAppRepo) Update(_ context.Context, _ *entity.App) error         { return nil }
func (m *mockAppRepo) Delete(_ context.Context, _ uuid.UUID) error           { return nil }
func (m *mockAppRepo) FindByPartnerAccountID(_ context.Context, _ uuid.UUID) ([]*entity.App, error) {
	return nil, nil
}
func (m *mockAppRepo) FindByPartnerAppID(_ context.Context, _ uuid.UUID, _ string) (*entity.App, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockAppRepo) FindAllByPartnerAppID(_ context.Context, _ string) ([]*entity.App, error) {
	return nil, nil
}
func (m *mockAppRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.App, error) {
	if a, ok := m.apps[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}

// --- PartnerRepo ---

type mockPartnerRepo struct {
	accounts map[uuid.UUID]*entity.PartnerAccount // keyed by ID
}

func (m *mockPartnerRepo) Create(_ context.Context, _ *entity.PartnerAccount) error   { return nil }
func (m *mockPartnerRepo) Update(_ context.Context, _ *entity.PartnerAccount) error   { return nil }
func (m *mockPartnerRepo) Delete(_ context.Context, _ uuid.UUID) error                { return nil }
func (m *mockPartnerRepo) GetAllIDs(_ context.Context) ([]uuid.UUID, error)           { return nil, nil }
func (m *mockPartnerRepo) FindByUserID(_ context.Context, _ uuid.UUID) (*entity.PartnerAccount, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockPartnerRepo) FindByOrgID(_ context.Context, _ uuid.UUID) (*entity.PartnerAccount, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockPartnerRepo) FindByPartnerID(_ context.Context, _ string) (*entity.PartnerAccount, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockPartnerRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	if a, ok := m.accounts[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}

// --- Decryptor ---

type mockDecryptor struct{}

func (d *mockDecryptor) Decrypt(_ []byte) ([]byte, error) {
	return []byte("fake-token"), nil
}

// --- TransactionFetcher ---

type mockTransactionFetcher struct {
	transactions []*entity.Transaction
	err          error
}

func (m *mockTransactionFetcher) FetchTransactions(_ context.Context, _ string, _ uuid.UUID, _, _ time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.err
}

// --- TransactionRepo ---

type mockTransactionRepo struct {
	upserted []*entity.Transaction
}

func (m *mockTransactionRepo) Upsert(_ context.Context, _ *entity.Transaction) error { return nil }
func (m *mockTransactionRepo) UpsertBatch(_ context.Context, txs []*entity.Transaction) error {
	m.upserted = append(m.upserted, txs...)
	return nil
}
func (m *mockTransactionRepo) FindByAppID(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*entity.Transaction, error) {
	return m.upserted, nil
}
func (m *mockTransactionRepo) FindByShopifyGID(_ context.Context, _ string) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) CountByAppID(_ context.Context, _ uuid.UUID) (int64, error) {
	return int64(len(m.upserted)), nil
}
func (m *mockTransactionRepo) GetEarningsSummary(_ context.Context, _ uuid.UUID) (*repository.EarningsSummary, error) {
	return nil, nil
}
func (m *mockTransactionRepo) GetPendingByAvailableDate(_ context.Context, _ uuid.UUID) ([]repository.EarningsByDate, error) {
	return nil, nil
}
func (m *mockTransactionRepo) GetUpcomingAvailability(_ context.Context, _ uuid.UUID, _ int) ([]repository.EarningsByDate, error) {
	return nil, nil
}
func (m *mockTransactionRepo) FindByDomain(_ context.Context, _ uuid.UUID, _ string, _, _ time.Time) ([]*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) GetEarningsSummaryByDomain(_ context.Context, _ uuid.UUID, _ string) (*repository.EarningsSummary, error) {
	return nil, nil
}
func (m *mockTransactionRepo) FindByAppIDPaginated(_ context.Context, _ uuid.UUID, _ repository.TransactionFilters) (*repository.TransactionPage, error) {
	return &repository.TransactionPage{}, nil
}

// --- LedgerRebuilder ---

type mockLedgerRebuilder struct {
	rebuildCalled   bool
	backfillCalled  bool
	backfillResult  int
}

func (m *mockLedgerRebuilder) RebuildFromTransactions(_ context.Context, _ uuid.UUID, _ time.Time) (*domainservice.LedgerRebuildResult, error) {
	m.rebuildCalled = true
	return &domainservice.LedgerRebuildResult{}, nil
}
func (m *mockLedgerRebuilder) BackfillHistoricalSnapshots(_ context.Context, _ uuid.UUID, _ []*entity.Transaction) (int, error) {
	m.backfillCalled = true
	return m.backfillResult, nil
}

// --- EventFetcher ---

type mockEventFetcher struct {
	events []external.AppEvent
	err    error
}

func (m *mockEventFetcher) FetchAppEvents(_ context.Context, _, _, _, _ string) ([]external.AppEvent, error) {
	return m.events, m.err
}

// --- AppEventRepo ---

type mockAppEventRepo struct {
	upserted []*entity.AppEvent
}

func (m *mockAppEventRepo) UpsertBatch(_ context.Context, events []*entity.AppEvent) error {
	m.upserted = append(m.upserted, events...)
	return nil
}
func (m *mockAppEventRepo) FindByAppAndShop(_ context.Context, _ uuid.UUID, _ string) ([]*entity.AppEvent, error) {
	return nil, nil
}
func (m *mockAppEventRepo) FindByAppID(_ context.Context, _ uuid.UUID) ([]*entity.AppEvent, error) {
	return nil, nil
}
func (m *mockAppEventRepo) FindByAppIDPaginated(_ context.Context, _ uuid.UUID, _ repository.EventFilters) (*repository.EventPage, error) {
	return &repository.EventPage{}, nil
}

// --- SubscriptionRepo ---

type mockSubRepo struct {
	subs    []*entity.Subscription
	upserted []*entity.Subscription
}

func (m *mockSubRepo) Upsert(_ context.Context, sub *entity.Subscription) error {
	m.upserted = append(m.upserted, sub)
	return nil
}
func (m *mockSubRepo) FindByID(_ context.Context, _ uuid.UUID) (*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) FindByAppID(_ context.Context, _ uuid.UUID) ([]*entity.Subscription, error) {
	return m.subs, nil
}
func (m *mockSubRepo) FindByShopifyGID(_ context.Context, _ string) (*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) FindByAppIDAndDomain(_ context.Context, _ uuid.UUID, _ string) (*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) FindByRiskState(_ context.Context, _ uuid.UUID, _ valueobject.RiskState) ([]*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) DeleteByAppID(_ context.Context, _ uuid.UUID) error     { return nil }
func (m *mockSubRepo) SoftDeleteByAppID(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockSubRepo) FindDeletedByAppID(_ context.Context, _ uuid.UUID) ([]*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) RestoreByID(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockSubRepo) FindWithFilters(_ context.Context, _ uuid.UUID, _ repository.SubscriptionFilters) (*repository.SubscriptionPage, error) {
	return nil, nil
}
func (m *mockSubRepo) GetSummary(_ context.Context, _ uuid.UUID) (*repository.SubscriptionSummary, error) {
	return nil, nil
}
func (m *mockSubRepo) GetPriceStats(_ context.Context, _ uuid.UUID) (*repository.PriceStats, error) {
	return nil, nil
}

// --- ShopRepo ---

type mockShopRepo struct {
	shops   map[string]*entity.Shop
	upserted []*entity.Shop
}

func newMockShopRepo() *mockShopRepo {
	return &mockShopRepo{shops: make(map[string]*entity.Shop)}
}

func (m *mockShopRepo) Upsert(_ context.Context, shop *entity.Shop) error {
	m.upserted = append(m.upserted, shop)
	m.shops[shop.MyshopifyDomain] = shop
	return nil
}
func (m *mockShopRepo) FindByDomain(_ context.Context, domain string) (*entity.Shop, error) {
	if s, ok := m.shops[domain]; ok {
		return s, nil
	}
	return nil, nil
}
func (m *mockShopRepo) FindByDomains(_ context.Context, domains []string) (map[string]*entity.Shop, error) {
	result := make(map[string]*entity.Shop)
	for _, d := range domains {
		if s, ok := m.shops[d]; ok {
			result[d] = s
		}
	}
	return result, nil
}

// --- ShopBrandFetcher ---

type mockBrandFetcher struct {
	shop *entity.Shop
	err  error
}

func (m *mockBrandFetcher) FetchBrand(_ context.Context, domain string) (*entity.Shop, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.shop != nil {
		return m.shop, nil
	}
	return &entity.Shop{
		ID:              uuid.New(),
		MyshopifyDomain: domain,
		ShopName:        "Test Shop",
	}, nil
}

// --- ReviewScraper ---

type mockReviewScraper struct {
	reviews []external.ScrapedReview
	err     error
}

func (m *mockReviewScraper) ScrapeReviews(_ context.Context, _ string, _ int) ([]external.ScrapedReview, error) {
	return m.reviews, m.err
}

// --- AppReviewRepo ---

type mockReviewRepo struct {
	upserted []*entity.AppReview
}

func (m *mockReviewRepo) UpsertBatch(_ context.Context, reviews []*entity.AppReview) error {
	m.upserted = append(m.upserted, reviews...)
	return nil
}
func (m *mockReviewRepo) FindByAppID(_ context.Context, _ uuid.UUID, _, _ int) ([]*entity.AppReview, error) {
	return nil, nil
}
func (m *mockReviewRepo) CountByAppID(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

// ==================== Helpers ====================

func setupRedis(t *testing.T) (*redis.Client, *queue.LockManager, *queue.ProgressTracker) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	syncJobRepo := newMockSyncJobRepo()
	lm := queue.NewLockManager(client)
	pt := queue.NewProgressTracker(client, syncJobRepo, 0, 0)
	return client, lm, pt
}

func makePayload(appID, userID, partnerID uuid.UUID, jobType string) *queue.SyncJobPayload {
	return &queue.SyncJobPayload{
		JobID:            uuid.New(),
		AppID:            appID,
		UserID:           userID,
		PartnerAccountID: partnerID,
		JobType:          jobType,
		EnqueuedAt:       time.Now().UTC(),
	}
}

func setupProcessorContext(t *testing.T) (uuid.UUID, uuid.UUID, uuid.UUID, *mockAppRepo, *mockPartnerRepo, *mockDecryptor) {
	t.Helper()
	appID := uuid.New()
	userID := uuid.New()
	partnerID := uuid.New()

	appRepo := &mockAppRepo{apps: map[uuid.UUID]*entity.App{
		appID: {ID: appID, PartnerAccountID: partnerID, Name: "Test App", PartnerAppID: "gid://partners/App/123"},
	}}
	partnerRepo := &mockPartnerRepo{accounts: map[uuid.UUID]*entity.PartnerAccount{
		partnerID: {ID: partnerID, UserID: userID, PartnerID: "org-123", EncryptedAccessToken: []byte("enc")},
	}}

	return appID, userID, partnerID, appRepo, partnerRepo, &mockDecryptor{}
}

// ==================== TransactionProcessor Tests ====================

func TestTransactionProcessor_Type(t *testing.T) {
	p := &TransactionProcessor{}
	if p.Type() != entity.SyncJobTypeTransactionSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeTransactionSync, p.Type())
	}
}

func TestTransactionProcessor_Success(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	txRepo := &mockTransactionRepo{}
	ledger := &mockLedgerRebuilder{}

	fetcher := &mockTransactionFetcher{
		transactions: []*entity.Transaction{
			{ID: uuid.New(), AppID: appID, PartnerAppGID: "gid://partners/App/123", ShopifyGID: "tx1"},
			{ID: uuid.New(), AppID: appID, PartnerAppGID: "gid://partners/App/123", ShopifyGID: "tx2"},
		},
	}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeTransactionSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeTransactionSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewTransactionProcessor(fetcher, txRepo, appRepo, partnerRepo, decryptor, ledger, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(txRepo.upserted) != 2 {
		t.Errorf("Expected 2 transactions upserted, got %d", len(txRepo.upserted))
	}
	if !ledger.rebuildCalled {
		t.Error("Expected ledger rebuild to be called")
	}
}

func TestTransactionProcessor_FetchError(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	txRepo := &mockTransactionRepo{}

	fetcher := &mockTransactionFetcher{err: fmt.Errorf("API error")}
	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeTransactionSync)

	p := NewTransactionProcessor(fetcher, txRepo, appRepo, partnerRepo, decryptor, nil, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err == nil {
		t.Fatal("Expected error from failed fetch")
	}
}

func TestTransactionProcessor_Cancellation(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeTransactionSync)

	// Set cancellation flag before processing
	_ = lm.RequestCancellation(context.Background(), payload.JobID)

	p := NewTransactionProcessor(&mockTransactionFetcher{}, &mockTransactionRepo{}, appRepo, partnerRepo, decryptor, nil, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err == nil {
		t.Fatal("Expected cancellation error")
	}
}

// ==================== SnapshotProcessor Tests ====================

func TestSnapshotProcessor_Type(t *testing.T) {
	p := &SnapshotProcessor{}
	if p.Type() != entity.SyncJobTypeSnapshotSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeSnapshotSync, p.Type())
	}
}

func TestSnapshotProcessor_Success(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	txRepo := &mockTransactionRepo{
		upserted: []*entity.Transaction{
			{ID: uuid.New(), AppID: appID},
		},
	}
	ledger := &mockLedgerRebuilder{backfillResult: 5}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeSnapshotSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeSnapshotSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewSnapshotProcessor(txRepo, appRepo, partnerRepo, decryptor, ledger, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if !ledger.backfillCalled {
		t.Error("Expected backfill to be called")
	}
}

func TestSnapshotProcessor_NoTransactions(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	txRepo := &mockTransactionRepo{} // empty

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeSnapshotSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeSnapshotSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewSnapshotProcessor(txRepo, appRepo, partnerRepo, decryptor, nil, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	// Processor returns nil on success — worker handles MarkCompleted
}

// ==================== EventProcessor Tests ====================

func TestEventProcessor_Type(t *testing.T) {
	p := &EventProcessor{}
	if p.Type() != entity.SyncJobTypeEventSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeEventSync, p.Type())
	}
}

func TestEventProcessor_Success(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	appEventRepo := &mockAppEventRepo{}
	subRepo := &mockSubRepo{
		subs: []*entity.Subscription{
			{ID: uuid.New(), AppID: appID, ShopifyShopGID: "gid://shopify/Shop/1"},
			{ID: uuid.New(), AppID: appID, ShopifyShopGID: "gid://shopify/Shop/2"},
		},
	}

	now := time.Now().UTC()
	eventFetcher := &mockEventFetcher{
		events: []external.AppEvent{
			{Type: "RELATIONSHIP_INSTALLED", OccurredAt: now},
		},
	}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeEventSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeEventSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewEventProcessor(eventFetcher, appEventRepo, subRepo, appRepo, partnerRepo, decryptor, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// 2 subs * 1 event each = 2 events
	if len(appEventRepo.upserted) != 2 {
		t.Errorf("Expected 2 events, got %d", len(appEventRepo.upserted))
	}
}

func TestEventProcessor_SkipsEmptyShopGID(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	appEventRepo := &mockAppEventRepo{}
	subRepo := &mockSubRepo{
		subs: []*entity.Subscription{
			{ID: uuid.New(), AppID: appID, ShopifyShopGID: ""}, // empty
		},
	}
	eventFetcher := &mockEventFetcher{
		events: []external.AppEvent{{Type: "INSTALLED", OccurredAt: time.Now()}},
	}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeEventSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeEventSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewEventProcessor(eventFetcher, appEventRepo, subRepo, appRepo, partnerRepo, decryptor, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(appEventRepo.upserted) != 0 {
		t.Errorf("Expected 0 events for empty shop GID, got %d", len(appEventRepo.upserted))
	}
}

// ==================== StatusProcessor Tests ====================

func TestStatusProcessor_Type(t *testing.T) {
	p := &StatusProcessor{}
	if p.Type() != entity.SyncJobTypeStatusSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeStatusSync, p.Type())
	}
}

func TestStatusProcessor_UpdatesStatus(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, partnerRepo, decryptor := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	subRepo := &mockSubRepo{
		subs: []*entity.Subscription{
			{ID: uuid.New(), AppID: appID, ShopifyShopGID: "gid://shopify/Shop/1", Status: "ACTIVE"},
		},
	}
	eventFetcher := &mockEventFetcher{
		events: []external.AppEvent{
			{Type: "RELATIONSHIP_UNINSTALLED", OccurredAt: time.Now()},
		},
	}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeStatusSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeStatusSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewStatusProcessor(eventFetcher, subRepo, appRepo, partnerRepo, decryptor, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(subRepo.upserted) == 0 {
		t.Fatal("Expected subscription to be updated")
	}
	if subRepo.upserted[0].Status != "UNINSTALLED" {
		t.Errorf("Expected UNINSTALLED, got %s", subRepo.upserted[0].Status)
	}
	if subRepo.upserted[0].RiskState != valueobject.RiskStateChurned {
		t.Errorf("Expected churned risk state, got %s", subRepo.upserted[0].RiskState)
	}
}

// ==================== StoreProcessor Tests ====================

func TestStoreProcessor_Type(t *testing.T) {
	p := &StoreProcessor{}
	if p.Type() != entity.SyncJobTypeStoreSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeStoreSync, p.Type())
	}
}

func TestStoreProcessor_FetchesNewDomains(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, _, decryptor := setupProcessorContext(t)
	_ = decryptor // store processor doesn't use decryptor for fetching

	syncJobRepo := newMockSyncJobRepo()
	shopRepo := newMockShopRepo()
	subRepo := &mockSubRepo{
		subs: []*entity.Subscription{
			{ID: uuid.New(), AppID: appID, MyshopifyDomain: "shop1.myshopify.com"},
			{ID: uuid.New(), AppID: appID, MyshopifyDomain: "shop2.myshopify.com"},
		},
	}
	brandFetcher := &mockBrandFetcher{}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeStoreSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeStoreSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewStoreProcessor(brandFetcher, shopRepo, subRepo, appRepo, nil, nil, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(shopRepo.upserted) != 2 {
		t.Errorf("Expected 2 shops upserted, got %d", len(shopRepo.upserted))
	}
}

func TestStoreProcessor_SkipsExistingDomains(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, _, _ := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	shopRepo := newMockShopRepo()
	// Pre-existing shop
	shopRepo.shops["existing.myshopify.com"] = &entity.Shop{ID: uuid.New(), MyshopifyDomain: "existing.myshopify.com"}

	subRepo := &mockSubRepo{
		subs: []*entity.Subscription{
			{ID: uuid.New(), AppID: appID, MyshopifyDomain: "existing.myshopify.com"},
			{ID: uuid.New(), AppID: appID, MyshopifyDomain: "new.myshopify.com"},
		},
	}
	brandFetcher := &mockBrandFetcher{}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeStoreSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeStoreSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewStoreProcessor(brandFetcher, shopRepo, subRepo, appRepo, nil, nil, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Only new domain should be fetched
	if len(shopRepo.upserted) != 1 {
		t.Errorf("Expected 1 new shop upserted, got %d", len(shopRepo.upserted))
	}
}

func TestStoreProcessor_NoDomains(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, _, _ := setupProcessorContext(t)

	syncJobRepo := newMockSyncJobRepo()
	shopRepo := newMockShopRepo()
	subRepo := &mockSubRepo{
		subs: []*entity.Subscription{
			{ID: uuid.New(), AppID: appID, MyshopifyDomain: ""}, // no domain
		},
	}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeStoreSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeStoreSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewStoreProcessor(nil, shopRepo, subRepo, appRepo, nil, nil, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	// Processor returns nil — worker handles MarkCompleted
}

// ==================== ReviewProcessor Tests ====================

func TestReviewProcessor_Type(t *testing.T) {
	p := &ReviewProcessor{}
	if p.Type() != entity.SyncJobTypeReviewSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeReviewSync, p.Type())
	}
}

func TestReviewProcessor_Success(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, _, _ := setupProcessorContext(t)

	// Set app store slug
	appRepo.apps[appID].AppStoreSlug = "test-app"

	syncJobRepo := newMockSyncJobRepo()
	reviewRepo := &mockReviewRepo{}

	scraper := &mockReviewScraper{
		reviews: []external.ScrapedReview{
			{Author: "Alice", Rating: 5, Body: "Great app!", Date: time.Now(), Location: "US", TimeUsing: "1 year"},
			{Author: "Bob", Rating: 4, Body: "Good app", Date: time.Now(), Location: "UK", TimeUsing: "6 months"},
		},
	}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeReviewSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeReviewSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewReviewProcessor(scraper, reviewRepo, appRepo, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(reviewRepo.upserted) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(reviewRepo.upserted))
	}
}

func TestReviewProcessor_NoSlug(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, _, _ := setupProcessorContext(t)

	// No slug set
	appRepo.apps[appID].AppStoreSlug = ""

	syncJobRepo := newMockSyncJobRepo()

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeReviewSync)
	syncJobRepo.jobs[payload.JobID] = entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeReviewSync, 0)
	syncJobRepo.jobs[payload.JobID].ID = payload.JobID

	p := NewReviewProcessor(nil, nil, appRepo, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	// Processor returns nil — worker handles MarkCompleted
}

func TestReviewProcessor_ScrapeError(t *testing.T) {
	_, lm, pt := setupRedis(t)
	appID, userID, partnerID, appRepo, _, _ := setupProcessorContext(t)

	appRepo.apps[appID].AppStoreSlug = "test-app"

	syncJobRepo := newMockSyncJobRepo()
	scraper := &mockReviewScraper{err: fmt.Errorf("scrape failed")}

	payload := makePayload(appID, userID, partnerID, entity.SyncJobTypeReviewSync)

	p := NewReviewProcessor(scraper, nil, appRepo, syncJobRepo, lm, pt)
	err := p.Process(context.Background(), payload)
	if err == nil {
		t.Fatal("Expected error from failed scrape")
	}
}

// ==================== FullSyncProcessor Tests ====================

func TestFullSyncProcessor_Type(t *testing.T) {
	p := &FullSyncProcessor{}
	if p.Type() != entity.SyncJobTypeFullSync {
		t.Errorf("Expected %s, got %s", entity.SyncJobTypeFullSync, p.Type())
	}
}

func TestFullSyncProcessor_CreatesChildJobs(t *testing.T) {
	client, lm, pt := setupRedis(t)

	syncJobRepo := newMockSyncJobRepo()

	appID := uuid.New()
	userID := uuid.New()
	partnerID := uuid.New()

	parentJob := entity.NewSyncJob(appID, userID, partnerID, entity.SyncJobTypeFullSync, 0)
	syncJobRepo.jobs[parentJob.ID] = parentJob

	payload := &queue.SyncJobPayload{
		JobID:            parentJob.ID,
		AppID:            appID,
		UserID:           userID,
		PartnerAccountID: partnerID,
		JobType:          entity.SyncJobTypeFullSync,
		EnqueuedAt:       time.Now().UTC(),
	}

	p := NewFullSyncProcessor(syncJobRepo, client, lm, pt)

	// Run in goroutine since it will block on waitForChildren
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Simulate children completing immediately (we complete them as they're created)
	go func() {
		// Give it time to create children
		time.Sleep(200 * time.Millisecond)
		for _, j := range syncJobRepo.jobs {
			if j.ParentJobID != nil {
				j.Status = entity.SyncJobStatusCompleted
				now := time.Now().UTC()
				j.CompletedAt = &now
			}
		}
	}()

	err := p.Process(ctx, payload)
	if err != nil && ctx.Err() == nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should have created 6 child jobs (2 wave1 + 4 wave2)
	childCount := 0
	for _, j := range syncJobRepo.jobs {
		if j.ParentJobID != nil {
			childCount++
		}
	}

	// At minimum wave1 (2 children) should be created
	if childCount < 2 {
		t.Errorf("Expected at least 2 child jobs, got %d", childCount)
	}
}

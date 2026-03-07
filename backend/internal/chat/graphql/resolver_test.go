package graphql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// --- Mock Repositories ---

type mockSubscriptionRepo struct {
	subscriptions []*entity.Subscription
	subscription  *entity.Subscription
	page          *repository.SubscriptionPage
	err           error
}

func (m *mockSubscriptionRepo) Upsert(ctx context.Context, s *entity.Subscription) error {
	return nil
}
func (m *mockSubscriptionRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	return m.subscription, m.err
}
func (m *mockSubscriptionRepo) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	return m.subscriptions, m.err
}
func (m *mockSubscriptionRepo) FindByShopifyGID(ctx context.Context, gid string) (*entity.Subscription, error) {
	return m.subscription, m.err
}
func (m *mockSubscriptionRepo) FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, domain string) (*entity.Subscription, error) {
	return m.subscription, m.err
}
func (m *mockSubscriptionRepo) FindByRiskState(ctx context.Context, appID uuid.UUID, rs valueobject.RiskState) ([]*entity.Subscription, error) {
	return m.subscriptions, m.err
}
func (m *mockSubscriptionRepo) DeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	return nil
}
func (m *mockSubscriptionRepo) SoftDeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	return nil
}
func (m *mockSubscriptionRepo) FindDeletedByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	return nil, nil
}
func (m *mockSubscriptionRepo) RestoreByID(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockSubscriptionRepo) FindWithFilters(ctx context.Context, appID uuid.UUID, f repository.SubscriptionFilters) (*repository.SubscriptionPage, error) {
	return m.page, m.err
}
func (m *mockSubscriptionRepo) GetSummary(ctx context.Context, appID uuid.UUID) (*repository.SubscriptionSummary, error) {
	return nil, nil
}
func (m *mockSubscriptionRepo) GetPriceStats(ctx context.Context, appID uuid.UUID) (*repository.PriceStats, error) {
	return nil, nil
}

type mockTransactionRepo struct {
	transactions []*entity.Transaction
	err          error
}

func (m *mockTransactionRepo) Upsert(ctx context.Context, tx *entity.Transaction) error { return nil }
func (m *mockTransactionRepo) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	return nil
}
func (m *mockTransactionRepo) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.err
}
func (m *mockTransactionRepo) FindByShopifyGID(ctx context.Context, gid string) (*entity.Transaction, error) {
	return nil, nil
}
func (m *mockTransactionRepo) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockTransactionRepo) GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*repository.EarningsSummary, error) {
	return nil, nil
}
func (m *mockTransactionRepo) GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]repository.EarningsByDate, error) {
	return nil, nil
}
func (m *mockTransactionRepo) GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]repository.EarningsByDate, error) {
	return nil, nil
}
func (m *mockTransactionRepo) FindByDomain(ctx context.Context, appID uuid.UUID, domain string, from, to time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.err
}
func (m *mockTransactionRepo) GetEarningsSummaryByDomain(ctx context.Context, appID uuid.UUID, domain string) (*repository.EarningsSummary, error) {
	return nil, nil
}

type mockSnapshotRepo struct {
	snapshot  *entity.DailyMetricsSnapshot
	snapshots []*entity.DailyMetricsSnapshot
	err       error
}

func (m *mockSnapshotRepo) Upsert(ctx context.Context, s *entity.DailyMetricsSnapshot) error {
	return nil
}
func (m *mockSnapshotRepo) FindByAppIDAndDate(ctx context.Context, appID uuid.UUID, date time.Time) (*entity.DailyMetricsSnapshot, error) {
	return m.snapshot, m.err
}
func (m *mockSnapshotRepo) FindByAppIDRange(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.DailyMetricsSnapshot, error) {
	return m.snapshots, m.err
}
func (m *mockSnapshotRepo) FindLatestByAppID(ctx context.Context, appID uuid.UUID) (*entity.DailyMetricsSnapshot, error) {
	return m.snapshot, m.err
}

type mockAppRepo struct {
	app  *entity.App
	apps []*entity.App
	err  error
}

func (m *mockAppRepo) Create(ctx context.Context, app *entity.App) error { return nil }
func (m *mockAppRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return m.app, m.err
}
func (m *mockAppRepo) FindByPartnerAccountID(ctx context.Context, partnerAccountID uuid.UUID) ([]*entity.App, error) {
	return m.apps, m.err
}
func (m *mockAppRepo) FindByPartnerAppID(ctx context.Context, partnerAccountID uuid.UUID, partnerAppID string) (*entity.App, error) {
	return m.app, m.err
}
func (m *mockAppRepo) FindAllByPartnerAppID(ctx context.Context, partnerAppID string) ([]*entity.App, error) {
	return m.apps, m.err
}
func (m *mockAppRepo) Update(ctx context.Context, app *entity.App) error { return nil }
func (m *mockAppRepo) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

type mockPartnerAccountRepo struct {
	account *entity.PartnerAccount
	err     error
}

func (m *mockPartnerAccountRepo) Create(ctx context.Context, a *entity.PartnerAccount) error {
	return nil
}
func (m *mockPartnerAccountRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}
func (m *mockPartnerAccountRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error) {
	return m.account, m.err
}
func (m *mockPartnerAccountRepo) FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error) {
	return m.account, m.err
}
func (m *mockPartnerAccountRepo) Update(ctx context.Context, a *entity.PartnerAccount) error {
	return nil
}
func (m *mockPartnerAccountRepo) Delete(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *mockPartnerAccountRepo) GetAllIDs(ctx context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

type mockSubscriptionEventRepo struct {
	events []*entity.SubscriptionEvent
	err    error
}

func (m *mockSubscriptionEventRepo) Create(ctx context.Context, e *entity.SubscriptionEvent) error {
	return nil
}
func (m *mockSubscriptionEventRepo) FindBySubscriptionID(ctx context.Context, subID uuid.UUID) ([]*entity.SubscriptionEvent, error) {
	return m.events, m.err
}
func (m *mockSubscriptionEventRepo) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.SubscriptionEvent, error) {
	return m.events, m.err
}
func (m *mockSubscriptionEventRepo) FindChurnEvents(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.SubscriptionEvent, error) {
	return m.events, m.err
}
func (m *mockSubscriptionEventRepo) CountByEventType(ctx context.Context, appID uuid.UUID, from, to time.Time) (map[string]int, error) {
	return nil, nil
}

// --- Helper to build resolver ---

func newTestResolver(opts ...func(*Resolver)) *queryResolver {
	r := &Resolver{
		SubscriptionRepo:      &mockSubscriptionRepo{},
		TransactionRepo:       &mockTransactionRepo{},
		SnapshotRepo:          &mockSnapshotRepo{},
		AppRepo:               &mockAppRepo{},
		PartnerAccountRepo:    &mockPartnerAccountRepo{},
		SubscriptionEventRepo: &mockSubscriptionEventRepo{},
		RiskEngine:            service.NewRiskEngine(),
		MetricsEngine:         service.NewMetricsEngine(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return &queryResolver{r}
}

// --- Tests ---

func TestSubscriptions_ReturnsPagedResults(t *testing.T) {
	appID := uuid.New()
	sub1 := &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		MyshopifyDomain: "shop1.myshopify.com",
		ShopName:        "Shop One",
		Status:          "ACTIVE",
		RiskState:       valueobject.RiskStateSafe,
		BasePriceCents:  2999,
		Currency:        "USD",
		BillingInterval: valueobject.BillingIntervalMonthly,
	}
	sub2 := &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		MyshopifyDomain: "shop2.myshopify.com",
		ShopName:        "Shop Two",
		Status:          "ACTIVE",
		RiskState:       valueobject.RiskStateOneCycleMissed,
		BasePriceCents:  4999,
		Currency:        "USD",
		BillingInterval: valueobject.BillingIntervalMonthly,
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.SubscriptionRepo = &mockSubscriptionRepo{
			page: &repository.SubscriptionPage{
				Subscriptions: []*entity.Subscription{sub1, sub2},
				Total:         2,
				Page:          1,
				PageSize:      50,
			},
		}
	})

	result, err := resolver.Subscriptions(context.Background(), appID.String(), nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(result.Nodes))
	}
	if result.TotalCount != 2 {
		t.Errorf("expected totalCount 2, got %d", result.TotalCount)
	}
	if result.Nodes[0].Domain != "shop1.myshopify.com" {
		t.Errorf("expected domain shop1.myshopify.com, got %s", result.Nodes[0].Domain)
	}
	if result.Nodes[1].RiskState != RiskStateOneCycleMissed {
		t.Errorf("expected risk state ONE_CYCLE_MISSED, got %s", result.Nodes[1].RiskState)
	}
}

func TestSubscription_SingleByID(t *testing.T) {
	subID := uuid.New()
	sub := &entity.Subscription{
		ID:              subID,
		MyshopifyDomain: "mystore.myshopify.com",
		ShopName:        "My Store",
		Status:          "ACTIVE",
		RiskState:       valueobject.RiskStateSafe,
		BasePriceCents:  1999,
		Currency:        "USD",
		BillingInterval: valueobject.BillingIntervalMonthly,
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.SubscriptionRepo = &mockSubscriptionRepo{subscription: sub}
	})

	result, err := resolver.Subscription(context.Background(), subID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != subID.String() {
		t.Errorf("expected ID %s, got %s", subID, result.ID)
	}
	if result.ShopName != "My Store" {
		t.Errorf("expected shop name My Store, got %s", result.ShopName)
	}
}

func TestMetrics_ReturnsLatestSnapshot(t *testing.T) {
	appID := uuid.New()
	snapshot := &entity.DailyMetricsSnapshot{
		ID:                   uuid.New(),
		AppID:                appID,
		ActiveMRRCents:       150000,
		RevenueAtRiskCents:   25000,
		UsageRevenueCents:    10000,
		TotalRevenueCents:    185000,
		RenewalSuccessRate:   0.92,
		SafeCount:            45,
		OneCycleMissedCount:  3,
		TwoCyclesMissedCount: 1,
		ChurnedCount:         2,
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.SnapshotRepo = &mockSnapshotRepo{snapshot: snapshot}
	})

	result, err := resolver.Metrics(context.Background(), appID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ActiveMrrCents != 150000 {
		t.Errorf("expected activeMrrCents 150000, got %d", result.ActiveMrrCents)
	}
	if result.RevenueAtRiskCents != 25000 {
		t.Errorf("expected revenueAtRiskCents 25000, got %d", result.RevenueAtRiskCents)
	}
	if result.RenewalSuccessRate != 0.92 {
		t.Errorf("expected renewalSuccessRate 0.92, got %f", result.RenewalSuccessRate)
	}
	if result.SafeCount != 45 {
		t.Errorf("expected safeCount 45, got %d", result.SafeCount)
	}
}

func TestMetricsTrend_ReturnsSnapshots(t *testing.T) {
	appID := uuid.New()
	snapshots := []*entity.DailyMetricsSnapshot{
		{
			ID:                 uuid.New(),
			AppID:              appID,
			Date:               time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			ActiveMRRCents:     100000,
			RenewalSuccessRate: 0.90,
			RevenueAtRiskCents: 20000,
			SafeCount:          40,
			ChurnedCount:       1,
		},
		{
			ID:                 uuid.New(),
			AppID:              appID,
			Date:               time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			ActiveMRRCents:     120000,
			RenewalSuccessRate: 0.93,
			RevenueAtRiskCents: 15000,
			SafeCount:          43,
			ChurnedCount:       2,
		},
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.SnapshotRepo = &mockSnapshotRepo{snapshots: snapshots}
	})

	months := 3
	result, err := resolver.MetricsTrend(context.Background(), appID.String(), &months)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(result))
	}
	if result[0].ActiveMrrCents != 100000 {
		t.Errorf("expected first snapshot MRR 100000, got %d", result[0].ActiveMrrCents)
	}
	if result[1].ActiveMrrCents != 120000 {
		t.Errorf("expected second snapshot MRR 120000, got %d", result[1].ActiveMrrCents)
	}
}

func TestStoreHealth_ReturnsDomainHealth(t *testing.T) {
	appID := uuid.New()
	sub := &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		MyshopifyDomain: "healthy-store.myshopify.com",
		ShopName:        "Healthy Store",
		Status:          "ACTIVE",
		RiskState:       valueobject.RiskStateSafe,
		BasePriceCents:  2999,
		Currency:        "USD",
		BillingInterval: valueobject.BillingIntervalMonthly,
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.SubscriptionRepo = &mockSubscriptionRepo{subscription: sub}
	})

	result, err := resolver.StoreHealth(context.Background(), appID.String(), "healthy-store.myshopify.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Domain != "healthy-store.myshopify.com" {
		t.Errorf("expected domain healthy-store.myshopify.com, got %s", result.Domain)
	}
	if result.ShopName != "Healthy Store" {
		t.Errorf("expected shop name Healthy Store, got %s", result.ShopName)
	}
	if !result.IsPaid {
		t.Error("expected isPaid true for SAFE subscription")
	}
	if result.RiskState != RiskStateSafe {
		t.Errorf("expected risk state SAFE, got %s", result.RiskState)
	}
}

func TestEarnings_AggregatesTransactions(t *testing.T) {
	appID := uuid.New()
	txs := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, GrossAmountCents: 3000, NetAmountCents: 2550, Currency: "USD"},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, GrossAmountCents: 5000, NetAmountCents: 4250, Currency: "USD"},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, GrossAmountCents: 1000, NetAmountCents: 850, Currency: "USD"},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeOneTime, GrossAmountCents: 500, NetAmountCents: 425, Currency: "USD"},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRefund, GrossAmountCents: 200, NetAmountCents: 200, Currency: "USD"},
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.TransactionRepo = &mockTransactionRepo{transactions: txs}
	})

	result, err := resolver.Earnings(context.Background(), appID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Recurring: 2550 + 4250 = 6800
	if result.RecurringCents != 6800 {
		t.Errorf("expected recurringCents 6800, got %d", result.RecurringCents)
	}
	if result.UsageCents != 850 {
		t.Errorf("expected usageCents 850, got %d", result.UsageCents)
	}
	if result.OneTimeCents != 425 {
		t.Errorf("expected oneTimeCents 425, got %d", result.OneTimeCents)
	}
	if result.RefundCents != 200 {
		t.Errorf("expected refundCents 200, got %d", result.RefundCents)
	}
	// Total: 6800 + 850 + 425 - 200 = 7875
	if result.TotalCents != 7875 {
		t.Errorf("expected totalCents 7875, got %d", result.TotalCents)
	}
}

func TestRiskSummary_CalculatesFromSubscriptions(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		{ID: uuid.New(), AppID: appID, Status: "ACTIVE", RiskState: valueobject.RiskStateSafe, BasePriceCents: 2999, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), AppID: appID, Status: "ACTIVE", RiskState: valueobject.RiskStateSafe, BasePriceCents: 4999, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), AppID: appID, Status: "ACTIVE", RiskState: valueobject.RiskStateOneCycleMissed, BasePriceCents: 1999, BillingInterval: valueobject.BillingIntervalMonthly},
		{ID: uuid.New(), AppID: appID, Status: "ACTIVE", RiskState: valueobject.RiskStateChurned, BasePriceCents: 999, BillingInterval: valueobject.BillingIntervalMonthly},
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.SubscriptionRepo = &mockSubscriptionRepo{subscriptions: subs}
	})

	result, err := resolver.RiskSummary(context.Background(), appID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalSubscriptions != 4 {
		t.Errorf("expected totalSubscriptions 4, got %d", result.TotalSubscriptions)
	}
	if result.Safe != 2 {
		t.Errorf("expected safe 2, got %d", result.Safe)
	}
	if result.OneCycleMissed != 1 {
		t.Errorf("expected oneCycleMissed 1, got %d", result.OneCycleMissed)
	}
	if result.Churned != 1 {
		t.Errorf("expected churned 1, got %d", result.Churned)
	}
	// At-risk subscriptions: only ONE_CYCLE_MISSED (CHURNED is not "at risk")
	if len(result.AtRiskSubscriptions) != 1 {
		t.Errorf("expected 1 at-risk subscription, got %d", len(result.AtRiskSubscriptions))
	}
}

func TestApps_RequiresAuth(t *testing.T) {
	resolver := newTestResolver()

	// No user in context → error
	_, err := resolver.Apps(context.Background())
	if err == nil {
		t.Fatal("expected error when no user in context")
	}
	if err.Error() != "authentication required" {
		t.Errorf("expected 'authentication required' error, got: %v", err)
	}
}

func TestApps_ReturnsUserApps(t *testing.T) {
	userID := uuid.New()
	partnerAccountID := uuid.New()
	appID := uuid.New()

	resolver := newTestResolver(func(r *Resolver) {
		r.PartnerAccountRepo = &mockPartnerAccountRepo{
			account: &entity.PartnerAccount{
				ID:     partnerAccountID,
				UserID: userID,
			},
		}
		r.AppRepo = &mockAppRepo{
			apps: []*entity.App{
				{ID: appID, PartnerAccountID: partnerAccountID, PartnerAppID: "gid://partners/App/123", Name: "Test App", TrackingEnabled: true, InstallCount: 42},
			},
		}
	})

	user := &entity.User{ID: userID, Email: "test@example.com", Role: valueobject.RoleOwner}
	ctx := middleware.SetUserContext(context.Background(), user)

	result, err := resolver.Apps(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result))
	}
	if result[0].Name != "Test App" {
		t.Errorf("expected app name 'Test App', got '%s'", result[0].Name)
	}
	if result[0].InstallCount != 42 {
		t.Errorf("expected installCount 42, got %d", result[0].InstallCount)
	}
}

func TestTransactions_ReturnsPaginated(t *testing.T) {
	appID := uuid.New()
	txs := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, MyshopifyDomain: "shop.myshopify.com", ShopName: "Shop", ChargeType: valueobject.ChargeTypeRecurring, GrossAmountCents: 2999, NetAmountCents: 2549, Currency: "USD", CreatedDate: time.Now()},
		{ID: uuid.New(), AppID: appID, MyshopifyDomain: "shop.myshopify.com", ShopName: "Shop", ChargeType: valueobject.ChargeTypeUsage, GrossAmountCents: 500, NetAmountCents: 425, Currency: "USD", CreatedDate: time.Now()},
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.TransactionRepo = &mockTransactionRepo{transactions: txs}
	})

	result, err := resolver.Transactions(context.Background(), appID.String(), nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("expected totalCount 2, got %d", result.TotalCount)
	}
	if len(result.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(result.Nodes))
	}
	if result.PageInfo.HasNextPage {
		t.Error("expected hasNextPage false for 2 results with default limit 50")
	}
}

func TestTransactions_FilterByChargeType(t *testing.T) {
	appID := uuid.New()
	txs := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, MyshopifyDomain: "s.myshopify.com", ShopName: "S", ChargeType: valueobject.ChargeTypeRecurring, GrossAmountCents: 2999, NetAmountCents: 2549, Currency: "USD", CreatedDate: time.Now()},
		{ID: uuid.New(), AppID: appID, MyshopifyDomain: "s.myshopify.com", ShopName: "S", ChargeType: valueobject.ChargeTypeUsage, GrossAmountCents: 500, NetAmountCents: 425, Currency: "USD", CreatedDate: time.Now()},
	}

	resolver := newTestResolver(func(r *Resolver) {
		r.TransactionRepo = &mockTransactionRepo{transactions: txs}
	})

	ct := ChargeTypeRecurring
	result, err := resolver.Transactions(context.Background(), appID.String(), &ct, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("expected totalCount 1 after filtering, got %d", result.TotalCount)
	}
	if result.Nodes[0].ChargeType != ChargeTypeRecurring {
		t.Errorf("expected RECURRING, got %s", result.Nodes[0].ChargeType)
	}
}

func TestSubscriptions_InvalidAppID(t *testing.T) {
	resolver := newTestResolver()

	_, err := resolver.Subscriptions(context.Background(), "not-a-uuid", nil, nil, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid app ID")
	}
}

func TestApp_SingleByID(t *testing.T) {
	appID := uuid.New()
	resolver := newTestResolver(func(r *Resolver) {
		r.AppRepo = &mockAppRepo{
			app: &entity.App{ID: appID, Name: "My App", PartnerAppID: "gid://partners/App/456", TrackingEnabled: true, InstallCount: 10},
		}
	})

	result, err := resolver.App(context.Background(), appID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "My App" {
		t.Errorf("expected name 'My App', got '%s'", result.Name)
	}
	if result.InstallCount != 10 {
		t.Errorf("expected installCount 10, got %d", result.InstallCount)
	}
}

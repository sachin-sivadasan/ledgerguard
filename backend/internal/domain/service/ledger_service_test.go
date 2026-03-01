package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// Mock implementations
type mockTxRepoForLedger struct {
	transactions []*entity.Transaction
	err          error
}

func (m *mockTxRepoForLedger) Upsert(ctx context.Context, tx *entity.Transaction) error {
	return nil
}

func (m *mockTxRepoForLedger) UpsertBatch(ctx context.Context, txs []*entity.Transaction) error {
	return nil
}

func (m *mockTxRepoForLedger) FindByAppID(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error) {
	return m.transactions, m.err
}

func (m *mockTxRepoForLedger) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Transaction, error) {
	return nil, nil
}

func (m *mockTxRepoForLedger) CountByAppID(ctx context.Context, appID uuid.UUID) (int64, error) {
	return int64(len(m.transactions)), nil
}

func (m *mockTxRepoForLedger) GetEarningsSummary(ctx context.Context, appID uuid.UUID) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}

func (m *mockTxRepoForLedger) GetPendingByAvailableDate(ctx context.Context, appID uuid.UUID) ([]repository.EarningsByDate, error) {
	return nil, nil
}

func (m *mockTxRepoForLedger) GetUpcomingAvailability(ctx context.Context, appID uuid.UUID, days int) ([]repository.EarningsByDate, error) {
	return nil, nil
}

func (m *mockTxRepoForLedger) FindByDomain(ctx context.Context, appID uuid.UUID, domain string, from, to time.Time) ([]*entity.Transaction, error) {
	return nil, m.err
}

func (m *mockTxRepoForLedger) GetEarningsSummaryByDomain(ctx context.Context, appID uuid.UUID, domain string) (*repository.EarningsSummary, error) {
	return &repository.EarningsSummary{}, nil
}

type mockSubRepoForLedger struct {
	subscriptions []*entity.Subscription
	upsertCalls   int
	deleteCalls   int
	err           error
}

func (m *mockSubRepoForLedger) Upsert(ctx context.Context, subscription *entity.Subscription) error {
	m.upsertCalls++
	m.subscriptions = append(m.subscriptions, subscription)
	return m.err
}

func (m *mockSubRepoForLedger) FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	return nil, nil
}

func (m *mockSubRepoForLedger) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	return m.subscriptions, m.err
}

func (m *mockSubRepoForLedger) FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Subscription, error) {
	return nil, nil
}

func (m *mockSubRepoForLedger) FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, myshopifyDomain string) (*entity.Subscription, error) {
	return nil, nil
}

func (m *mockSubRepoForLedger) FindByRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.Subscription, error) {
	return nil, nil
}

func (m *mockSubRepoForLedger) DeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	m.deleteCalls++
	m.subscriptions = nil
	return m.err
}

func (m *mockSubRepoForLedger) FindWithFilters(ctx context.Context, appID uuid.UUID, filters repository.SubscriptionFilters) (*repository.SubscriptionPage, error) {
	return &repository.SubscriptionPage{
		Subscriptions: m.subscriptions,
		Total:         len(m.subscriptions),
		Page:          1,
		PageSize:      25,
		TotalPages:    1,
	}, nil
}

func (m *mockSubRepoForLedger) GetSummary(ctx context.Context, appID uuid.UUID) (*repository.SubscriptionSummary, error) {
	return &repository.SubscriptionSummary{}, nil
}

func (m *mockSubRepoForLedger) GetPriceStats(ctx context.Context, appID uuid.UUID) (*repository.PriceStats, error) {
	return &repository.PriceStats{}, nil
}

func (m *mockSubRepoForLedger) SoftDeleteByAppID(ctx context.Context, appID uuid.UUID) error {
	return nil
}

func (m *mockSubRepoForLedger) FindDeletedByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error) {
	return nil, nil
}

func (m *mockSubRepoForLedger) RestoreByID(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestLedgerService_RebuildFromTransactions_Success(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	transactions := []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     2999,
			Currency:        "USD",
			TransactionDate: now.AddDate(0, -1, 0), // 1 month ago
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/2",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     2999,
			Currency:        "USD",
			TransactionDate: now, // Today
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/3",
			MyshopifyDomain: "store2.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     4999,
			Currency:        "USD",
			TransactionDate: now,
		},
	}

	txRepo := &mockTxRepoForLedger{transactions: transactions}
	subRepo := &mockSubRepoForLedger{}

	service := NewLedgerService(txRepo, subRepo)

	result, err := service.RebuildFromTransactions(context.Background(), appID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SubscriptionsUpdated != 2 {
		t.Errorf("expected 2 subscriptions, got %d", result.SubscriptionsUpdated)
	}

	// MRR should be sum of both subscriptions (2999 + 4999)
	expectedMRR := int64(2999 + 4999)
	if result.TotalMRRCents != expectedMRR {
		t.Errorf("expected MRR %d, got %d", expectedMRR, result.TotalMRRCents)
	}

	if result.RiskSummary.SafeCount != 2 {
		t.Errorf("expected 2 safe subscriptions, got %d", result.RiskSummary.SafeCount)
	}

	if subRepo.deleteCalls != 1 {
		t.Errorf("expected 1 delete call, got %d", subRepo.deleteCalls)
	}

	if subRepo.upsertCalls != 2 {
		t.Errorf("expected 2 upsert calls, got %d", subRepo.upsertCalls)
	}
}

func TestLedgerService_RebuildFromTransactions_SeparatesRecurringAndUsage(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	transactions := []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     2999,
			Currency:        "USD",
			TransactionDate: now,
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/2",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeUsage,
			NetAmountCents:     500,
			Currency:        "USD",
			TransactionDate: now,
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/3",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeUsage,
			NetAmountCents:     300,
			Currency:        "USD",
			TransactionDate: now,
		},
	}

	txRepo := &mockTxRepoForLedger{transactions: transactions}
	subRepo := &mockSubRepoForLedger{}

	service := NewLedgerService(txRepo, subRepo)

	result, err := service.RebuildFromTransactions(context.Background(), appID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MRR should only include RECURRING (2999)
	if result.TotalMRRCents != 2999 {
		t.Errorf("expected MRR 2999 (RECURRING only), got %d", result.TotalMRRCents)
	}

	// Usage should be sum of USAGE transactions (500 + 300)
	expectedUsage := int64(500 + 300)
	if result.TotalUsageCents != expectedUsage {
		t.Errorf("expected usage %d, got %d", expectedUsage, result.TotalUsageCents)
	}
}

func TestLedgerService_RebuildFromTransactions_ComputesExpectedRenewalDate(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)
	lastChargeDate := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	transactions := []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     2999,
			Currency:        "USD",
			TransactionDate: lastChargeDate,
		},
	}

	txRepo := &mockTxRepoForLedger{transactions: transactions}
	subRepo := &mockSubRepoForLedger{}

	service := NewLedgerService(txRepo, subRepo)

	_, err := service.RebuildFromTransactions(context.Background(), appID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subRepo.subscriptions) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subRepo.subscriptions))
	}

	sub := subRepo.subscriptions[0]

	// Check last_recurring_charge_date
	if sub.LastRecurringChargeDate == nil {
		t.Fatal("expected last_recurring_charge_date to be set")
	}
	if !sub.LastRecurringChargeDate.Equal(lastChargeDate) {
		t.Errorf("expected last_recurring_charge_date %v, got %v", lastChargeDate, *sub.LastRecurringChargeDate)
	}

	// Check expected_next_charge_date (should be +1 month)
	if sub.ExpectedNextChargeDate == nil {
		t.Fatal("expected expected_next_charge_date to be set")
	}
	expectedNext := lastChargeDate.AddDate(0, 1, 0) // March 1, 2026
	if !sub.ExpectedNextChargeDate.Equal(expectedNext) {
		t.Errorf("expected expected_next_charge_date %v, got %v", expectedNext, *sub.ExpectedNextChargeDate)
	}
}

func TestLedgerService_RebuildFromTransactions_ClassifiesRiskState(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	// Transaction from 76 days ago
	// Expected next charge = 76 days ago + 30 days = 46 days ago
	// Days past due = 46 (should be ONE_CYCLE_MISSED: 31-60 range)
	oldChargeDate := now.AddDate(0, 0, -76)

	transactions := []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     2999,
			Currency:        "USD",
			TransactionDate: oldChargeDate,
		},
	}

	txRepo := &mockTxRepoForLedger{transactions: transactions}
	subRepo := &mockSubRepoForLedger{}

	service := NewLedgerService(txRepo, subRepo)

	result, err := service.RebuildFromTransactions(context.Background(), appID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RiskSummary.OneCycleMissedCount != 1 {
		t.Errorf("expected 1 one_cycle_missed, got %d", result.RiskSummary.OneCycleMissedCount)
	}

	if len(subRepo.subscriptions) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subRepo.subscriptions))
	}

	sub := subRepo.subscriptions[0]
	if sub.RiskState != valueobject.RiskStateOneCycleMissed {
		t.Errorf("expected risk state ONE_CYCLE_MISSED, got %s", sub.RiskState)
	}
}

func TestLedgerService_SeparateRevenue(t *testing.T) {
	appID := uuid.New()
	now := time.Now()

	transactions := []*entity.Transaction{
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, TransactionDate: now},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, TransactionDate: now},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeRecurring, TransactionDate: now},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeOneTime, TransactionDate: now},
		{ID: uuid.New(), AppID: appID, ChargeType: valueobject.ChargeTypeUsage, TransactionDate: now},
	}

	service := NewLedgerService(nil, nil)

	recurring, usage := service.SeparateRevenue(transactions)

	if len(recurring) != 2 {
		t.Errorf("expected 2 recurring transactions, got %d", len(recurring))
	}

	if len(usage) != 2 {
		t.Errorf("expected 2 usage transactions, got %d", len(usage))
	}
}

func TestLedgerService_RebuildFromTransactions_Deterministic(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	transactions := []*entity.Transaction{
		{
			ID:              uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store-b.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     2999,
			TransactionDate: now,
		},
		{
			ID:              uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/2",
			MyshopifyDomain: "store-a.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     4999,
			TransactionDate: now,
		},
	}

	// Run rebuild twice
	txRepo1 := &mockTxRepoForLedger{transactions: transactions}
	subRepo1 := &mockSubRepoForLedger{}
	service1 := NewLedgerService(txRepo1, subRepo1)
	result1, _ := service1.RebuildFromTransactions(context.Background(), appID, now)

	txRepo2 := &mockTxRepoForLedger{transactions: transactions}
	subRepo2 := &mockSubRepoForLedger{}
	service2 := NewLedgerService(txRepo2, subRepo2)
	result2, _ := service2.RebuildFromTransactions(context.Background(), appID, now)

	// Results should be identical
	if result1.TotalMRRCents != result2.TotalMRRCents {
		t.Errorf("non-deterministic MRR: %d vs %d", result1.TotalMRRCents, result2.TotalMRRCents)
	}

	if result1.SubscriptionsUpdated != result2.SubscriptionsUpdated {
		t.Errorf("non-deterministic subscription count: %d vs %d", result1.SubscriptionsUpdated, result2.SubscriptionsUpdated)
	}

	// Subscriptions should be in same order (sorted by domain)
	if len(subRepo1.subscriptions) != len(subRepo2.subscriptions) {
		t.Fatalf("different subscription counts")
	}

	for i := range subRepo1.subscriptions {
		if subRepo1.subscriptions[i].MyshopifyDomain != subRepo2.subscriptions[i].MyshopifyDomain {
			t.Errorf("non-deterministic order at %d: %s vs %s",
				i,
				subRepo1.subscriptions[i].MyshopifyDomain,
				subRepo2.subscriptions[i].MyshopifyDomain)
		}
	}
}

func TestLedgerService_RebuildFromTransactions_NoTransactions(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	txRepo := &mockTxRepoForLedger{transactions: []*entity.Transaction{}}
	subRepo := &mockSubRepoForLedger{}

	service := NewLedgerService(txRepo, subRepo)

	result, err := service.RebuildFromTransactions(context.Background(), appID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SubscriptionsUpdated != 0 {
		t.Errorf("expected 0 subscriptions, got %d", result.SubscriptionsUpdated)
	}

	if result.TotalMRRCents != 0 {
		t.Errorf("expected MRR 0, got %d", result.TotalMRRCents)
	}
}

func TestLedgerService_DetectsBillingInterval_Annual(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 2, 26, 12, 0, 0, 0, time.UTC)

	// Two transactions ~365 days apart = annual
	transactions := []*entity.Transaction{
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/1",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     29900, // $299/year
			TransactionDate: now.AddDate(-1, 0, 0), // 1 year ago
		},
		{
			ID:              uuid.New(),
			AppID:           appID,
			ShopifyGID:      "gid://shopify/Transaction/2",
			MyshopifyDomain: "store1.myshopify.com",
			ChargeType:      valueobject.ChargeTypeRecurring,
			NetAmountCents:     29900,
			TransactionDate: now,
		},
	}

	txRepo := &mockTxRepoForLedger{transactions: transactions}
	subRepo := &mockSubRepoForLedger{}

	service := NewLedgerService(txRepo, subRepo)

	_, err := service.RebuildFromTransactions(context.Background(), appID, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(subRepo.subscriptions) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subRepo.subscriptions))
	}

	sub := subRepo.subscriptions[0]
	if sub.BillingInterval != valueobject.BillingIntervalAnnual {
		t.Errorf("expected ANNUAL billing interval, got %s", sub.BillingInterval)
	}

	// MRR should be annual price / 12
	expectedMRR := int64(29900 / 12)
	if sub.MRRCents() != expectedMRR {
		t.Errorf("expected MRR %d, got %d", expectedMRR, sub.MRRCents())
	}
}

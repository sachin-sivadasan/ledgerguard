package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// LedgerRebuildResult contains the result of a ledger rebuild
type LedgerRebuildResult struct {
	AppID                uuid.UUID
	SubscriptionsUpdated int
	TotalMRRCents        int64
	TotalUsageCents      int64
	RiskSummary          RiskSummary
	RebuildAt            time.Time
	// Snapshot contains the daily metrics snapshot (if snapshotRepo is configured)
	Snapshot             *entity.DailyMetricsSnapshot
}

// RiskSummary contains counts of subscriptions by risk state
type RiskSummary struct {
	SafeCount             int
	OneCycleMissedCount   int
	TwoCyclesMissedCount  int
	ChurnedCount          int
}

// LedgerService handles deterministic ledger rebuilds
type LedgerService struct {
	txRepo       repository.TransactionRepository
	subRepo      repository.SubscriptionRepository
	snapshotRepo repository.DailyMetricsSnapshotRepository
	metrics      *MetricsEngine
}

func NewLedgerService(
	txRepo repository.TransactionRepository,
	subRepo repository.SubscriptionRepository,
) *LedgerService {
	return &LedgerService{
		txRepo:  txRepo,
		subRepo: subRepo,
		metrics: NewMetricsEngine(),
	}
}

// WithSnapshotRepository adds a snapshot repository for daily metrics storage
func (s *LedgerService) WithSnapshotRepository(repo repository.DailyMetricsSnapshotRepository) *LedgerService {
	s.snapshotRepo = repo
	return s
}

// RebuildFromTransactions rebuilds subscription state from transactions
// This is deterministic: same transactions → same subscription state
func (s *LedgerService) RebuildFromTransactions(ctx context.Context, appID uuid.UUID, now time.Time) (*LedgerRebuildResult, error) {
	// Fetch all transactions for the app (12-month window)
	from := now.AddDate(-1, 0, 0)
	transactions, err := s.txRepo.FindByAppID(ctx, appID, from, now)
	if err != nil {
		return nil, err
	}

	// Group transactions by domain (store)
	byDomain := s.groupTransactionsByDomain(transactions)

	// Rebuild subscriptions from transactions
	subscriptions := s.rebuildSubscriptions(appID, byDomain, now)

	// Delete existing subscriptions and insert rebuilt ones
	if err := s.subRepo.DeleteByAppID(ctx, appID); err != nil {
		return nil, err
	}

	var totalMRR int64
	var totalUsage int64
	riskSummary := RiskSummary{}

	for _, sub := range subscriptions {
		if err := s.subRepo.Upsert(ctx, sub); err != nil {
			return nil, err
		}

		// Accumulate MRR (only from ACTIVE subscriptions)
		if sub.IsActive() {
			totalMRR += sub.MRRCents()
		}

		// Count by risk state
		switch sub.RiskState {
		case valueobject.RiskStateSafe:
			riskSummary.SafeCount++
		case valueobject.RiskStateOneCycleMissed:
			riskSummary.OneCycleMissedCount++
		case valueobject.RiskStateTwoCyclesMissed:
			riskSummary.TwoCyclesMissedCount++
		case valueobject.RiskStateChurned:
			riskSummary.ChurnedCount++
		}
	}

	// Calculate total usage revenue
	totalUsage = s.sumUsageRevenue(transactions)

	result := &LedgerRebuildResult{
		AppID:                appID,
		SubscriptionsUpdated: len(subscriptions),
		TotalMRRCents:        totalMRR,
		TotalUsageCents:      totalUsage,
		RiskSummary:          riskSummary,
		RebuildAt:            now,
	}

	// Store daily metrics snapshot if repository is configured
	if s.snapshotRepo != nil && s.metrics != nil {
		snapshot := s.metrics.ComputeAllMetrics(appID, subscriptions, transactions, now)
		if err := s.snapshotRepo.Upsert(ctx, snapshot); err != nil {
			return nil, err
		}
		result.Snapshot = snapshot
	}

	return result, nil
}

// groupTransactionsByDomain groups transactions by myshopify_domain
func (s *LedgerService) groupTransactionsByDomain(transactions []*entity.Transaction) map[string][]*entity.Transaction {
	byDomain := make(map[string][]*entity.Transaction)
	for _, tx := range transactions {
		byDomain[tx.MyshopifyDomain] = append(byDomain[tx.MyshopifyDomain], tx)
	}
	return byDomain
}

// rebuildSubscriptions creates subscription records from transactions
func (s *LedgerService) rebuildSubscriptions(appID uuid.UUID, byDomain map[string][]*entity.Transaction, now time.Time) []*entity.Subscription {
	var subscriptions []*entity.Subscription

	for domain, txs := range byDomain {
		sub := s.buildSubscriptionFromTransactions(appID, domain, txs, now)
		if sub != nil {
			subscriptions = append(subscriptions, sub)
		}
	}

	// Sort for deterministic output
	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].MyshopifyDomain < subscriptions[j].MyshopifyDomain
	})

	return subscriptions
}

// buildSubscriptionFromTransactions builds a subscription from a store's transactions
func (s *LedgerService) buildSubscriptionFromTransactions(appID uuid.UUID, domain string, txs []*entity.Transaction, now time.Time) *entity.Subscription {
	// Sort transactions by date (oldest first for processing order)
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].TransactionDate.Before(txs[j].TransactionDate)
	})

	// Find recurring transactions
	var recurringTxs []*entity.Transaction
	for _, tx := range txs {
		if tx.ChargeType == valueobject.ChargeTypeRecurring {
			recurringTxs = append(recurringTxs, tx)
		}
	}

	// If no recurring transactions, no subscription to track
	if len(recurringTxs) == 0 {
		return nil
	}

	// Get the most recent recurring transaction
	lastRecurring := recurringTxs[len(recurringTxs)-1]

	// Detect billing interval - use from transaction if available, otherwise detect from pattern
	billingInterval := s.detectBillingInterval(recurringTxs)
	if lastRecurring.BillingInterval != "" {
		switch lastRecurring.BillingInterval {
		case "ANNUAL":
			billingInterval = valueobject.BillingIntervalAnnual
		case "MONTHLY", "EVERY_30_DAYS":
			billingInterval = valueobject.BillingIntervalMonthly
		}
	}

	// Create subscription
	// Use GrossAmountCents for subscription price (what customer pays)
	// If GrossAmountCents is not set, fall back to NetAmountCents
	basePriceCents := lastRecurring.GrossAmountCents
	if basePriceCents == 0 {
		basePriceCents = lastRecurring.NetAmountCents
	}

	// Determine subscription GID:
	// 1. Use real Shopify subscription GID if available from transaction
	// 2. Otherwise generate internal synthetic ID with lg_ prefix
	subscriptionGID := lastRecurring.SubscriptionGID
	if subscriptionGID == "" {
		// Generate synthetic ID - use lg_sub_ prefix to clearly distinguish from Shopify GIDs
		subscriptionGID = "lg_sub_" + uuid.NewSHA1(uuid.NameSpaceDNS, []byte(domain)).String()
	}

	// Determine subscription status from transaction or default to ACTIVE
	status := lastRecurring.SubscriptionStatus
	if status == "" {
		status = "ACTIVE"
	}

	sub := entity.NewSubscription(
		appID,
		subscriptionGID,
		domain,
		lastRecurring.ShopName, // Shop name from transaction
		"",                     // Plan name not available from transactions
		basePriceCents,
		lastRecurring.Currency,
		billingInterval,
	)

	// Set subscription status from transaction data
	sub.Status = status

	// Update from the most recent charge
	sub.UpdateFromRecurringCharge(lastRecurring.TransactionDate, basePriceCents)

	// Use subscription period end from transaction if available
	if lastRecurring.SubscriptionPeriodEnd != nil {
		sub.ExpectedNextChargeDate = lastRecurring.SubscriptionPeriodEnd
	}

	// Classify risk based on current date and status
	sub.ClassifyRisk(now)

	return sub
}

// detectBillingInterval detects MONTHLY vs ANNUAL from transaction pattern
func (s *LedgerService) detectBillingInterval(txs []*entity.Transaction) valueobject.BillingInterval {
	if len(txs) < 2 {
		return valueobject.BillingIntervalMonthly // Default
	}

	// Calculate average days between transactions
	var totalDays float64
	for i := 1; i < len(txs); i++ {
		days := txs[i].TransactionDate.Sub(txs[i-1].TransactionDate).Hours() / 24
		totalDays += days
	}
	avgDays := totalDays / float64(len(txs)-1)

	// If average is closer to 365 than 30, it's annual
	if avgDays > 180 {
		return valueobject.BillingIntervalAnnual
	}
	return valueobject.BillingIntervalMonthly
}

// sumUsageRevenue calculates total usage revenue from transactions
func (s *LedgerService) sumUsageRevenue(transactions []*entity.Transaction) int64 {
	var total int64
	for _, tx := range transactions {
		if tx.ChargeType == valueobject.ChargeTypeUsage {
			total += tx.AmountCents()
		}
	}
	return total
}

// SeparateRevenue separates transactions into RECURRING and USAGE streams
func (s *LedgerService) SeparateRevenue(transactions []*entity.Transaction) (recurring, usage []*entity.Transaction) {
	for _, tx := range transactions {
		switch tx.ChargeType {
		case valueobject.ChargeTypeRecurring:
			recurring = append(recurring, tx)
		case valueobject.ChargeTypeUsage:
			usage = append(usage, tx)
		}
	}
	return recurring, usage
}

// BackfillHistoricalSnapshots creates daily snapshots for each month in the transaction history
// This should be called after syncing transactions to populate historical metrics
func (s *LedgerService) BackfillHistoricalSnapshots(ctx context.Context, appID uuid.UUID, transactions []*entity.Transaction) (int, error) {
	if s.snapshotRepo == nil || s.metrics == nil || len(transactions) == 0 {
		return 0, nil
	}

	// Find date range from transactions
	var earliest, latest time.Time
	for _, tx := range transactions {
		if earliest.IsZero() || tx.TransactionDate.Before(earliest) {
			earliest = tx.TransactionDate
		}
		if latest.IsZero() || tx.TransactionDate.After(latest) {
			latest = tx.TransactionDate
		}
	}

	// Create snapshots for the last day of each month
	snapshotsCreated := 0
	current := time.Date(earliest.Year(), earliest.Month(), 1, 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()

	for current.Before(latest) || current.Equal(latest) {
		// Get end of month
		endOfMonth := time.Date(current.Year(), current.Month()+1, 0, 23, 59, 59, 0, time.UTC)

		// Don't create snapshots for future dates
		snapshotDate := endOfMonth
		if snapshotDate.After(today) {
			snapshotDate = today
		}

		// Filter transactions up to this date
		txsUpToDate := s.filterTransactionsUpTo(transactions, snapshotDate)
		if len(txsUpToDate) == 0 {
			current = current.AddDate(0, 1, 0)
			continue
		}

		// Build subscriptions from transactions up to this date
		byDomain := s.groupTransactionsByDomain(txsUpToDate)
		subscriptions := s.rebuildSubscriptions(appID, byDomain, snapshotDate)

		// Filter transactions for just this month (for revenue calculation)
		startOfMonth := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, time.UTC)
		txsThisMonth := s.filterTransactionsInRange(transactions, startOfMonth, endOfMonth)

		// Compute and store snapshot
		snapshot := s.metrics.ComputeAllMetrics(appID, subscriptions, txsThisMonth, snapshotDate)
		if err := s.snapshotRepo.Upsert(ctx, snapshot); err != nil {
			return snapshotsCreated, err
		}
		snapshotsCreated++

		// Move to next month
		current = current.AddDate(0, 1, 0)
	}

	return snapshotsCreated, nil
}

// filterTransactionsUpTo returns transactions on or before the given date
func (s *LedgerService) filterTransactionsUpTo(transactions []*entity.Transaction, date time.Time) []*entity.Transaction {
	var filtered []*entity.Transaction
	for _, tx := range transactions {
		if !tx.TransactionDate.After(date) {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

// filterTransactionsInRange returns transactions within the given date range
func (s *LedgerService) filterTransactionsInRange(transactions []*entity.Transaction, start, end time.Time) []*entity.Transaction {
	var filtered []*entity.Transaction
	for _, tx := range transactions {
		if (tx.TransactionDate.Equal(start) || tx.TransactionDate.After(start)) &&
			(tx.TransactionDate.Equal(end) || tx.TransactionDate.Before(end)) {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

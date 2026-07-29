package service

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// SyncHistoryStart is the floor date for a FULL sync/rebuild — an arbitrary early date
// comfortably before any real app transaction (no monetized data exists before an app's
// first sale, which is well after this floor). Using it as the `from` pulls an app's
// ENTIRE history — "from the beginning" — rather than a trailing window: for the Partner
// API fetch, cursor pagination + the createdAtMin filter return every transaction from
// this date; for the DB reads (rebuild / snapshot backfill) it selects all stored rows.
// So no real transaction is truncated. Incremental catch-up syncs still fetch a small
// LookbackDays delta.
var SyncHistoryStart = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)

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
	// Rebuild from the app's ENTIRE stored transaction history (deterministic full rebuild,
	// not a trailing window) — see SyncHistoryStart.
	transactions, err := s.txRepo.FindByAppID(ctx, appID, SyncHistoryStart, now)
	if err != nil {
		return nil, err
	}

	log.Printf("LedgerService: found %d transactions for app %s (from %s to %s)", len(transactions), appID, SyncHistoryStart.Format("2006-01-02"), now.Format("2006-01-02"))

	// Count by charge type for debugging
	typeCounts := map[string]int{}
	for _, tx := range transactions {
		typeCounts[tx.ChargeType.String()]++
	}
	log.Printf("LedgerService: charge type breakdown: %v", typeCounts)

	// Group transactions by domain (store)
	byDomain := s.groupTransactionsByDomain(transactions)
	log.Printf("LedgerService: grouped into %d unique domains", len(byDomain))

	// Rebuild subscriptions from transactions
	subscriptions := s.rebuildSubscriptions(appID, byDomain, now)
	log.Printf("LedgerService: rebuilt %d subscriptions", len(subscriptions))

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

	// Generate stable domain key (deterministic, survives reinstalls)
	stableDomainKey := "lg_sub_" + uuid.NewSHA1(uuid.NameSpaceDNS, []byte(domain)).String()

	// Use real Shopify subscription GID if available, otherwise fall back to stable key
	subscriptionGID := lastRecurring.SubscriptionGID
	if subscriptionGID == "" {
		subscriptionGID = stableDomainKey
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

	// Set shop GID for events lookup and stable domain key
	sub.ShopifyShopGID = lastRecurring.ShopifyShopGID
	sub.StableDomainKey = stableDomainKey

	// Set the real business start date = earliest recurring charge date. Computed as the
	// MIN over recurring txs of (CreatedDate, falling back to TransactionDate) so it
	// exactly matches migration 000043's backfill (MIN(COALESCE(created_date,
	// transaction_date))) — the two paths must agree. Distinct from CreatedAt (the
	// record-created timestamp, reset on every rebuild); reports use StartDate().
	var activatedAt time.Time
	for _, tx := range recurringTxs {
		cd := tx.CreatedDate
		if cd.IsZero() {
			cd = tx.TransactionDate
		}
		if activatedAt.IsZero() || cd.Before(activatedAt) {
			activatedAt = cd
		}
	}
	if !activatedAt.IsZero() {
		sub.ActivatedAt = &activatedAt
	}

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

// BackfillHistoricalSnapshots creates daily snapshots from the transaction history.
// This enables forecasting from day one — a user with 12 months of Shopify data
// gets ~365 daily snapshots immediately, well above the 90-point forecasting minimum.
//
// Optimized: O(n log n + 365*k) where k = avg domains per day, instead of O(365*n).
// Sorts transactions once, then uses pointer-based advancement so each tx is visited once.
// Collects snapshots in memory and batch-upserts at the end.
func (s *LedgerService) BackfillHistoricalSnapshots(ctx context.Context, appID uuid.UUID, transactions []*entity.Transaction) (int, error) {
	if s.snapshotRepo == nil || s.metrics == nil || len(transactions) == 0 {
		return 0, nil
	}

	// 1. Sort transactions by date once: O(n log n)
	sorted := make([]*entity.Transaction, len(transactions))
	copy(sorted, transactions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TransactionDate.Before(sorted[j].TransactionDate)
	})

	earliest := sorted[0].TransactionDate
	latest := sorted[len(sorted)-1].TransactionDate

	// 2. Pointer-based advancement through sorted transactions
	today := time.Now().UTC()
	current := time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 0, 0, 0, 0, time.UTC)
	txPtr := 0
	allTxsSoFar := make([]*entity.Transaction, 0, len(sorted))
	var batch []*entity.DailyMetricsSnapshot

	for !current.After(latest) {
		snapshotDate := current
		if snapshotDate.After(today) {
			break
		}

		// Advance pointer — each tx is visited exactly once across all days
		endOfDay := time.Date(snapshotDate.Year(), snapshotDate.Month(), snapshotDate.Day(), 23, 59, 59, 999999999, time.UTC)
		for txPtr < len(sorted) && !sorted[txPtr].TransactionDate.After(endOfDay) {
			allTxsSoFar = append(allTxsSoFar, sorted[txPtr])
			txPtr++
		}

		if len(allTxsSoFar) == 0 {
			current = current.AddDate(0, 0, 1)
			continue
		}

		// Rebuild subscriptions from cumulative slice
		byDomain := s.groupTransactionsByDomain(allTxsSoFar)
		subscriptions := s.rebuildSubscriptions(appID, byDomain, snapshotDate)

		// Revenue: trailing 30-day window
		windowStart := snapshotDate.AddDate(0, 0, -30)
		txsInWindow := s.filterTransactionsInRange(allTxsSoFar, windowStart, snapshotDate)

		snapshot := s.metrics.ComputeAllMetrics(appID, subscriptions, txsInWindow, snapshotDate)
		batch = append(batch, snapshot)

		current = current.AddDate(0, 0, 1)
	}

	// 3. Batch upsert all snapshots
	if len(batch) > 0 {
		if err := s.snapshotRepo.UpsertBatch(ctx, batch); err != nil {
			return 0, err
		}
	}

	return len(batch), nil
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

package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	domainEntity "github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
	revrepo "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/repository"
)

// ReadModelBuilder populates the CQRS read model for the Revenue API
type ReadModelBuilder struct {
	// Source repositories (main ledger)
	subscriptionRepo repository.SubscriptionRepository
	transactionRepo  repository.TransactionRepository

	// Target repositories (read model)
	subscriptionStatusRepo revrepo.SubscriptionStatusRepository
	usageStatusRepo        revrepo.UsageStatusRepository
}

// NewReadModelBuilder creates a new ReadModelBuilder
func NewReadModelBuilder(
	subscriptionRepo repository.SubscriptionRepository,
	transactionRepo repository.TransactionRepository,
	subscriptionStatusRepo revrepo.SubscriptionStatusRepository,
	usageStatusRepo revrepo.UsageStatusRepository,
) *ReadModelBuilder {
	return &ReadModelBuilder{
		subscriptionRepo:       subscriptionRepo,
		transactionRepo:        transactionRepo,
		subscriptionStatusRepo: subscriptionStatusRepo,
		usageStatusRepo:        usageStatusRepo,
	}
}

// RebuildForApp rebuilds the read model for a specific app
// This should be called after a ledger sync completes
func (b *ReadModelBuilder) RebuildForApp(ctx context.Context, appID uuid.UUID) error {
	log.Printf("ReadModelBuilder: rebuilding read model for app %s", appID)
	start := time.Now()

	// Rebuild subscription statuses
	if err := b.rebuildSubscriptionStatuses(ctx, appID); err != nil {
		log.Printf("ReadModelBuilder: failed to rebuild subscription statuses: %v", err)
		return err
	}

	// Rebuild usage statuses
	if err := b.rebuildUsageStatuses(ctx, appID); err != nil {
		log.Printf("ReadModelBuilder: failed to rebuild usage statuses: %v", err)
		return err
	}

	log.Printf("ReadModelBuilder: completed rebuild for app %s in %v", appID, time.Since(start))
	return nil
}

// rebuildSubscriptionStatuses rebuilds all subscription statuses for an app
func (b *ReadModelBuilder) rebuildSubscriptionStatuses(ctx context.Context, appID uuid.UUID) error {
	// Get all subscriptions for the app
	subscriptions, err := b.subscriptionRepo.FindByAppID(ctx, appID)
	if err != nil {
		return err
	}
	log.Printf("ReadModelBuilder: rebuilding for app %s — %d subscriptions", appID, len(subscriptions))

	// Convert to status entities
	statuses := make([]*entity.SubscriptionStatus, len(subscriptions))
	for i, sub := range subscriptions {
		statuses[i] = b.subscriptionToStatus(sub)
	}

	// Batch upsert
	if err := b.subscriptionStatusRepo.UpsertBatch(ctx, statuses); err != nil {
		return err
	}
	log.Printf("ReadModelBuilder: upserted %d subscription statuses for app %s", len(statuses), appID)
	return nil
}

// subscriptionToStatus converts a domain subscription to a status read model
func (b *ReadModelBuilder) subscriptionToStatus(sub *domainEntity.Subscription) *entity.SubscriptionStatus {
	now := time.Now().UTC()

	// Calculate months overdue
	monthsOverdue := 0
	if sub.ExpectedNextChargeDate != nil && now.After(*sub.ExpectedNextChargeDate) {
		days := int(now.Sub(*sub.ExpectedNextChargeDate).Hours() / 24)
		monthsOverdue = days / 30
	}

	// Determine if paid current cycle
	isPaidCurrentCycle := sub.Status == "ACTIVE" && sub.RiskState == valueobject.RiskStateSafe

	return &entity.SubscriptionStatus{
		ID:                       uuid.New(),
		ShopifyGID:               sub.ShopifyGID,
		AppID:                    sub.AppID,
		MyshopifyDomain:          sub.MyshopifyDomain,
		ShopName:                 sub.ShopName,
		PlanName:                 sub.PlanName,
		RiskState:                sub.RiskState,
		IsPaidCurrentCycle:       isPaidCurrentCycle,
		MonthsOverdue:            monthsOverdue,
		LastSuccessfulChargeDate: sub.LastRecurringChargeDate,
		ExpectedNextChargeDate:   sub.ExpectedNextChargeDate,
		Status:                   sub.Status,
		LastSyncedAt:             now,
	}
}

// rebuildUsageStatuses rebuilds all usage statuses for an app
func (b *ReadModelBuilder) rebuildUsageStatuses(ctx context.Context, appID uuid.UUID) error {
	// Get all subscriptions first (to map usage to subscriptions)
	subscriptions, err := b.subscriptionRepo.FindByAppID(ctx, appID)
	if err != nil {
		return err
	}

	// Get all transactions for the app (last 12 months)
	now := time.Now()
	from := now.AddDate(-1, 0, 0)
	allTransactions, err := b.transactionRepo.FindByAppID(ctx, appID, from, now)
	if err != nil {
		return err
	}

	// Filter to USAGE transactions only
	transactions := make([]*domainEntity.Transaction, 0)
	for _, txn := range allTransactions {
		if txn.ChargeType == valueobject.ChargeTypeUsage {
			transactions = append(transactions, txn)
		}
	}
	log.Printf("ReadModelBuilder: found %d usage transactions for app %s", len(transactions), appID)

	if len(transactions) == 0 {
		return nil
	}

	// Build subscription map by domain for matching
	subByDomain := make(map[string]*domainEntity.Subscription)
	for _, sub := range subscriptions {
		subByDomain[sub.MyshopifyDomain] = sub
	}

	// Convert to usage status entities
	statuses := make([]*entity.UsageStatus, 0, len(transactions))
	for _, txn := range transactions {
		// Find the parent subscription
		sub := subByDomain[txn.MyshopifyDomain]
		if sub == nil {
			// Skip usage without a matching subscription
			continue
		}

		status := b.transactionToUsageStatus(txn, sub)
		statuses = append(statuses, status)
	}

	if len(statuses) == 0 {
		return nil
	}

	// Batch upsert
	if err := b.usageStatusRepo.UpsertBatch(ctx, statuses); err != nil {
		return err
	}
	log.Printf("ReadModelBuilder: upserted %d usage statuses for app %s", len(statuses), appID)
	return nil
}

// transactionToUsageStatus converts a usage transaction to a status read model
func (b *ReadModelBuilder) transactionToUsageStatus(txn *domainEntity.Transaction, sub *domainEntity.Subscription) *entity.UsageStatus {
	// For AppUsageSale, the chargeId (stored in SubscriptionGID) is the actual
	// Shopify usage record GID (gid://shopify/AppUsageRecord/...).
	// Fall back to the Partner API transaction GID if chargeId is empty.
	usageGID := txn.ShopifyGID
	if txn.SubscriptionGID != "" {
		usageGID = txn.SubscriptionGID
	}

	return &entity.UsageStatus{
		ID:                     uuid.New(),
		ShopifyGID:             usageGID,
		SubscriptionShopifyGID: sub.ShopifyGID,
		SubscriptionID:         sub.ID,
		Billed:                 true, // If we have a transaction, it's billed
		BillingDate:            &txn.TransactionDate,
		AmountCents:            int(txn.NetAmountCents),
		Description:            "", // Not stored in transaction
		LastSyncedAt:           time.Now().UTC(),
	}
}

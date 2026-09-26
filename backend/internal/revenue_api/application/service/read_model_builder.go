package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainEntity "github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
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
	logging.FromContext(ctx).Info("rebuilding read model", zap.String("app_id", appID.String()))
	start := time.Now()

	// Rebuild subscription statuses
	if err := b.rebuildSubscriptionStatuses(ctx, appID); err != nil {
		logging.FromContext(ctx).Error("failed to rebuild subscription statuses", zap.String("app_id", appID.String()), zap.Error(err))
		return err
	}

	// Rebuild usage statuses
	if err := b.rebuildUsageStatuses(ctx, appID); err != nil {
		logging.FromContext(ctx).Error("failed to rebuild usage statuses", zap.String("app_id", appID.String()), zap.Error(err))
		return err
	}

	logging.FromContext(ctx).Info("completed read model rebuild", zap.String("app_id", appID.String()), zap.Duration("duration", time.Since(start)))
	return nil
}

// rebuildSubscriptionStatuses rebuilds all subscription statuses for an app
func (b *ReadModelBuilder) rebuildSubscriptionStatuses(ctx context.Context, appID uuid.UUID) error {
	// Get all subscriptions for the app
	subscriptions, err := b.subscriptionRepo.FindByAppID(ctx, appID)
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("rebuilding subscription statuses", zap.String("app_id", appID.String()), zap.Int("count", len(subscriptions)))

	// Convert to status entities
	statuses := make([]*entity.SubscriptionStatus, len(subscriptions))
	for i, sub := range subscriptions {
		statuses[i] = b.subscriptionToStatus(sub)
	}

	// Batch upsert
	if err := b.subscriptionStatusRepo.UpsertBatch(ctx, statuses); err != nil {
		return err
	}
	logging.FromContext(ctx).Info("upserted subscription statuses", zap.Int("count", len(statuses)), zap.String("app_id", appID.String()))
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

	// Rebuild usage statuses from the app's ENTIRE stored history (matches the ledger /
	// snapshot full-history rebuild — see SyncHistoryStart), not a trailing window.
	now := time.Now()
	allTransactions, err := b.transactionRepo.FindByAppID(ctx, appID, domainservice.SyncHistoryStart, now)
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
	logging.FromContext(ctx).Info("found usage transactions", zap.Int("count", len(transactions)), zap.String("app_id", appID.String()))

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
	logging.FromContext(ctx).Info("upserted usage statuses", zap.Int("count", len(statuses)), zap.String("app_id", appID.String()))
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

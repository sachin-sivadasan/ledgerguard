package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// ReadModelRebuilder rebuilds the Revenue API read model after sync
type ReadModelRebuilder interface {
	RebuildForApp(ctx context.Context, appID uuid.UUID) error
}

// TransactionProcessor handles transaction_sync jobs
type TransactionProcessor struct {
	fetcher          service.TransactionFetcher
	txRepo           repository.TransactionRepository
	appRepo          repository.AppRepository
	partnerRepo      repository.PartnerAccountRepository
	decryptor        queue.Decryptor
	ledger           service.LedgerRebuilder
	syncJobRepo      repository.SyncJobRepository
	lockManager      *queue.LockManager
	progress         *queue.ProgressTracker
	readModelBuilder ReadModelRebuilder
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(
	fetcher service.TransactionFetcher,
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor queue.Decryptor,
	ledger service.LedgerRebuilder,
	syncJobRepo repository.SyncJobRepository,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *TransactionProcessor {
	return &TransactionProcessor{
		fetcher:     fetcher,
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
		decryptor:   decryptor,
		ledger:      ledger,
		syncJobRepo: syncJobRepo,
		lockManager: lockManager,
		progress:    progress,
	}
}

// WithReadModelBuilder adds an optional read model builder for Revenue API
func (p *TransactionProcessor) WithReadModelBuilder(builder ReadModelRebuilder) *TransactionProcessor {
	p.readModelBuilder = builder
	return p
}

func (p *TransactionProcessor) Type() string { return entity.SyncJobTypeTransactionSync }

func (p *TransactionProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	// Common preamble
	pCtx, err := queue.PrepareProcessorContext(ctx, payload, p.appRepo, p.partnerRepo, p.decryptor)
	if err != nil {
		return err
	}

	// Check cancellation
	if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
		return fmt.Errorf("job cancelled")
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{Message: "Fetching transactions..."})

	now := time.Now().UTC()
	var from time.Time
	if payload.LookbackDays > 0 {
		from = now.AddDate(0, 0, -payload.LookbackDays) // incremental catch-up: small delta
	} else {
		from = domainservice.SyncHistoryStart // full sync: entire history from the beginning
	}

	fetchCtx := external.WithOrganizationID(ctx, pCtx.OrganizationID)
	fetchCtx = external.WithPartnerAppGID(fetchCtx, pCtx.App.PartnerAppID)

	transactions, err := p.fetcher.FetchTransactions(fetchCtx, pCtx.AccessToken, payload.AppID, from, now)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   len(transactions),
		Message: fmt.Sprintf("Fetched %d transactions, processing earnings...", len(transactions)),
	})

	// Check cancellation
	if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
		return fmt.Errorf("job cancelled")
	}

	// Process earnings
	earningsCalc := domainservice.NewEarningsCalculator()
	earningsCalc.ProcessTransactions(transactions, now)

	// Upsert transactions
	if len(transactions) > 0 {
		if err := p.txRepo.UpsertBatch(ctx, transactions); err != nil {
			return fmt.Errorf("failed to store transactions: %w", err)
		}
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:     len(transactions),
		Completed: len(transactions),
		Message:   "Rebuilding ledger...",
	})

	// Rebuild ledger
	if p.ledger != nil {
		if _, err := p.ledger.RebuildFromTransactions(ctx, payload.AppID, now); err != nil {
			return fmt.Errorf("failed to rebuild ledger: %w", err)
		}
	}

	// Rebuild Revenue API read model
	if p.readModelBuilder != nil {
		if err := p.readModelBuilder.RebuildForApp(ctx, payload.AppID); err != nil {
			log.Printf("[queue] TransactionProcessor: failed to rebuild read model: %v", err)
			// Non-fatal — don't fail the sync for read model issues
		}
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(transactions),
		Completed: len(transactions),
		Message:   "Transaction sync complete",
	})

	log.Printf("[queue] TransactionProcessor: synced %d transactions for app %s (job %s)", len(transactions), payload.AppID, payload.JobID)
	return nil
}

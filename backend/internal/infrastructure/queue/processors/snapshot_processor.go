package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// SnapshotProcessor handles snapshot_sync jobs (backfill historical snapshots)
type SnapshotProcessor struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
	decryptor   queue.Decryptor
	ledger      service.LedgerRebuilder
	syncJobRepo repository.SyncJobRepository
	lockManager *queue.LockManager
	progress    *queue.ProgressTracker
}

func NewSnapshotProcessor(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor queue.Decryptor,
	ledger service.LedgerRebuilder,
	syncJobRepo repository.SyncJobRepository,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *SnapshotProcessor {
	return &SnapshotProcessor{
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

func (p *SnapshotProcessor) Type() string { return entity.SyncJobTypeSnapshotSync }

func (p *SnapshotProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	p.progress.Update(ctx, payload.JobID, queue.Progress{Message: "Loading transactions for snapshots..."})

	now := time.Now().UTC()
	// Backfill snapshots across the app's ENTIRE stored history — see SyncHistoryStart.
	transactions, err := p.txRepo.FindByAppID(ctx, payload.AppID, domainservice.SyncHistoryStart, now)
	if err != nil {
		return fmt.Errorf("failed to load transactions: %w", err)
	}

	if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
		return fmt.Errorf("job cancelled")
	}

	if len(transactions) == 0 {
		p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{Message: "No transactions for snapshots"})
		return nil
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   len(transactions),
		Message: "Backfilling historical snapshots...",
	})

	snapshotCount, err := p.ledger.BackfillHistoricalSnapshots(ctx, payload.AppID, transactions)
	if err != nil {
		return fmt.Errorf("failed to backfill snapshots: %w", err)
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     snapshotCount,
		Completed: snapshotCount,
		Message:   fmt.Sprintf("Backfilled %d snapshots", snapshotCount),
	})

	log.Printf("[queue] SnapshotProcessor: backfilled %d snapshots for app %s (job %s)", snapshotCount, payload.AppID, payload.JobID)
	return nil
}

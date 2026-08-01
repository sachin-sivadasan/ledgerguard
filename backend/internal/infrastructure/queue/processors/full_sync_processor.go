package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

const childPollInterval = 5 * time.Second

// FullSyncProcessor orchestrates a full sync by creating and monitoring child jobs
type FullSyncProcessor struct {
	syncJobRepo repository.SyncJobRepository
	redisClient *redis.Client
	lockManager *queue.LockManager
	progress    *queue.ProgressTracker
}

func NewFullSyncProcessor(
	syncJobRepo repository.SyncJobRepository,
	redisClient *redis.Client,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *FullSyncProcessor {
	return &FullSyncProcessor{
		syncJobRepo: syncJobRepo,
		redisClient: redisClient,
		lockManager: lockManager,
		progress:    progress,
	}
}

func (p *FullSyncProcessor) Type() string { return entity.SyncJobTypeFullSync }

func (p *FullSyncProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	parentJob, err := p.syncJobRepo.FindByID(ctx, payload.JobID)
	if err != nil {
		return fmt.Errorf("failed to find parent job: %w", err)
	}

	totalChildren := 6
	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   totalChildren,
		Message: "Starting full sync — Wave 1 (transactions, reviews)...",
	})

	// Wave 1: transaction_sync, review_sync (no dependency on subscriptions)
	wave1Types := []struct {
		jobType    string
		entityType string
	}{
		{entity.SyncJobTypeTransactionSync, "transaction"},
		{entity.SyncJobTypeReviewSync, "review"},
	}

	var wave1JobIDs []string
	for _, w := range wave1Types {
		childJob := entity.NewChildSyncJob(parentJob, w.jobType, w.entityType)
		if err := p.syncJobRepo.Create(ctx, childJob); err != nil {
			return fmt.Errorf("failed to create child job %s: %w", w.jobType, err)
		}

		childPayload := &queue.SyncJobPayload{
			JobID:            childJob.ID,
			AppID:            childJob.AppID,
			UserID:           childJob.UserID,
			PartnerAccountID: childJob.PartnerAccountID,
			JobType:          childJob.JobType,
			ParentJobID:      childJob.ParentJobID,
			Priority:         childJob.Priority,
			EntityType:       childJob.EntityType,
			EnqueuedAt:       time.Now().UTC(),
		}

		if err := queue.Enqueue(ctx, p.redisClient, childPayload); err != nil {
			_ = p.syncJobRepo.MarkFailed(ctx, childJob.ID, err.Error())
			continue
		}
		wave1JobIDs = append(wave1JobIDs, childJob.ID.String())
	}

	// Bug 8 fix: Guard against empty wave1JobIDs (all enqueues failed)
	if len(wave1JobIDs) == 0 {
		return fmt.Errorf("failed to enqueue any wave 1 jobs")
	}

	// Wait for transaction_sync to complete (needed before Wave 2)
	if err := p.waitForChildren(ctx, payload, []string{wave1JobIDs[0]}, "Waiting for transaction sync..."); err != nil {
		return err
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:     totalChildren,
		Completed: 1,
		Message:   "Transaction sync done — starting Wave 2 (events, status, stores)...",
	})

	// Wave 2: depends on subscriptions from ledger rebuild (transaction_sync).
	// NOTE: snapshot_sync is intentionally NOT here — it must run in Wave 3, after
	// status_sync reconciles subscription risk, so the daily metrics snapshot
	// (Dashboard) matches the reconciled Subscriptions/Risk pages instead of the
	// pre-reconciliation charge-only risk (RISK-1).
	wave2Types := []struct {
		jobType    string
		entityType string
	}{
		{entity.SyncJobTypeEventSync, "event"},
		{entity.SyncJobTypeStatusSync, "subscription"},
		{entity.SyncJobTypeStoreSync, "store"},
	}

	var wave2JobIDs []string
	for _, w := range wave2Types {
		childJob := entity.NewChildSyncJob(parentJob, w.jobType, w.entityType)
		if err := p.syncJobRepo.Create(ctx, childJob); err != nil {
			return fmt.Errorf("failed to create child job %s: %w", w.jobType, err)
		}

		childPayload := &queue.SyncJobPayload{
			JobID:            childJob.ID,
			AppID:            childJob.AppID,
			UserID:           childJob.UserID,
			PartnerAccountID: childJob.PartnerAccountID,
			JobType:          childJob.JobType,
			ParentJobID:      childJob.ParentJobID,
			Priority:         childJob.Priority,
			EntityType:       childJob.EntityType,
			EnqueuedAt:       time.Now().UTC(),
		}

		if err := queue.Enqueue(ctx, p.redisClient, childPayload); err != nil {
			_ = p.syncJobRepo.MarkFailed(ctx, childJob.ID, err.Error())
			continue
		}
		wave2JobIDs = append(wave2JobIDs, childJob.ID.String())
	}

	// Wait for remaining Wave 1 + all Wave 2 jobs (status reconciliation must finish
	// before the snapshot).
	allRemaining := append(wave1JobIDs[1:], wave2JobIDs...)
	if err := p.waitForChildren(ctx, payload, allRemaining, "Waiting for events, status, stores..."); err != nil {
		return err
	}

	// Wave 3: snapshot — runs AFTER status_sync so the daily metrics snapshot
	// reflects the reconciled subscription risk (RISK-1 convergence).
	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:     totalChildren,
		Completed: 5,
		Message:   "Status reconciled — Wave 3 (daily snapshot)...",
	})

	snapshotJob := entity.NewChildSyncJob(parentJob, entity.SyncJobTypeSnapshotSync, "snapshot")
	if err := p.syncJobRepo.Create(ctx, snapshotJob); err != nil {
		return fmt.Errorf("failed to create snapshot child job: %w", err)
	}
	snapshotPayload := &queue.SyncJobPayload{
		JobID:            snapshotJob.ID,
		AppID:            snapshotJob.AppID,
		UserID:           snapshotJob.UserID,
		PartnerAccountID: snapshotJob.PartnerAccountID,
		JobType:          snapshotJob.JobType,
		ParentJobID:      snapshotJob.ParentJobID,
		Priority:         snapshotJob.Priority,
		EntityType:       snapshotJob.EntityType,
		EnqueuedAt:       time.Now().UTC(),
	}
	if err := queue.Enqueue(ctx, p.redisClient, snapshotPayload); err != nil {
		_ = p.syncJobRepo.MarkFailed(ctx, snapshotJob.ID, err.Error())
	} else if err := p.waitForChildren(ctx, payload, []string{snapshotJob.ID.String()}, "Waiting for daily snapshot..."); err != nil {
		return err
	}

	// Check children for final status
	children, err := p.syncJobRepo.FindByParentJobID(ctx, payload.JobID)
	if err != nil {
		return fmt.Errorf("failed to find children: %w", err)
	}

	hasFailure := false
	for _, child := range children {
		if child.Status == entity.SyncJobStatusFailed {
			hasFailure = true
			break
		}
	}

	if hasFailure {
		p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
			Total:     6,
			Completed: 6,
			Message:   "Full sync completed with some failures",
		})
		// Use UpdateStatus directly for partial_failure — worker won't call MarkCompleted for this case
		return p.syncJobRepo.UpdateStatus(ctx, payload.JobID, entity.SyncJobStatusPartialFailure)
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     totalChildren,
		Completed: totalChildren,
		Message:   "Full sync complete",
	})

	log.Printf("FullSyncProcessor: completed for app %s", payload.AppID)
	return nil
}

// waitForChildren polls until all specified children are in a terminal state
func (p *FullSyncProcessor) waitForChildren(ctx context.Context, payload *queue.SyncJobPayload, childIDs []string, progressMsg string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(childPollInterval):
		}

		if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
			return fmt.Errorf("job cancelled")
		}

		allDone := true
		children, err := p.syncJobRepo.FindByParentJobID(ctx, payload.JobID)
		if err != nil {
			continue
		}

		// Check only the children we're waiting for
		childSet := make(map[string]bool)
		for _, id := range childIDs {
			childSet[id] = true
		}

		for _, child := range children {
			if childSet[child.ID.String()] && !child.IsTerminal() {
				allDone = false
				break
			}
		}

		if allDone {
			return nil
		}

		p.progress.Update(ctx, payload.JobID, queue.Progress{Message: progressMsg})
	}
}

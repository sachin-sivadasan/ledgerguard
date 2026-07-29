package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// StatusProcessor handles status_sync jobs (enrich subscription status from events)
type StatusProcessor struct {
	eventFetcher service.EventFetcher
	subRepo      repository.SubscriptionRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
	decryptor    queue.Decryptor
	syncJobRepo  repository.SyncJobRepository
	lockManager  *queue.LockManager
	progress     *queue.ProgressTracker
}

func NewStatusProcessor(
	eventFetcher service.EventFetcher,
	subRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor queue.Decryptor,
	syncJobRepo repository.SyncJobRepository,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *StatusProcessor {
	return &StatusProcessor{
		eventFetcher: eventFetcher,
		subRepo:      subRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
		decryptor:    decryptor,
		syncJobRepo:  syncJobRepo,
		lockManager:  lockManager,
		progress:     progress,
	}
}

func (p *StatusProcessor) Type() string { return entity.SyncJobTypeStatusSync }

func (p *StatusProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	pCtx, err := queue.PrepareProcessorContext(ctx, payload, p.appRepo, p.partnerRepo, p.decryptor)
	if err != nil {
		return err
	}

	subscriptions, err := p.subRepo.FindByAppID(ctx, payload.AppID)
	if err != nil {
		return fmt.Errorf("failed to find subscriptions: %w", err)
	}

	log.Printf("[queue] StatusProcessor: processing %d subscriptions for app %s (job %s)", len(subscriptions), payload.AppID, payload.JobID)

	fetchCtx := external.WithOrganizationID(ctx, pCtx.OrganizationID)

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   len(subscriptions),
		Message: fmt.Sprintf("Enriching status for %d subscriptions...", len(subscriptions)),
	})

	updated := 0
	failed := 0 // shops whose status could not be refreshed (fetch or upsert error)
	for i, sub := range subscriptions {
		if sub.ShopifyShopGID == "" {
			log.Printf("[queue] StatusProcessor: skipping subscription %s — no ShopifyShopGID (job %s)", sub.ID, payload.JobID)
			continue
		}

		if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
			return fmt.Errorf("job cancelled")
		}

		events, err := p.eventFetcher.FetchAppEvents(fetchCtx, pCtx.OrganizationID, pCtx.AccessToken, pCtx.App.PartnerAppID, sub.ShopifyShopGID)
		if err != nil {
			log.Printf("[queue] StatusProcessor: error fetching events for shop %s: %v (job %s)", sub.ShopifyShopGID, err, payload.JobID)
			failed++
			continue
		}

		newStatus := external.GetLatestSubscriptionStatus(events)
		if newStatus != "" && newStatus != sub.Status {
			sub.Status = newStatus
			sub.UpdatedAt = time.Now().UTC()

			if newStatus == "UNINSTALLED" || newStatus == "CANCELLED" {
				sub.RiskState = valueobject.RiskStateChurned
			}

			if err := p.subRepo.Upsert(ctx, sub); err != nil {
				log.Printf("[queue] StatusProcessor: error persisting status for shop %s: %v (job %s)", sub.ShopifyShopGID, err, payload.JobID)
				failed++
				continue
			}
			updated++
		}

		p.progress.Update(ctx, payload.JobID, queue.Progress{
			Total:     len(subscriptions),
			Completed: i + 1,
			Message:   fmt.Sprintf("Processed %d/%d subscriptions", i+1, len(subscriptions)),
		})
	}

	// Surface partial failures: without the old 10-sub cap the per-shop loop now spans the
	// whole account, so a swallowed fetch/upsert error must not masquerade as a clean sync
	// (e.g. a token expiring mid-run would otherwise silently leave stale statuses).
	doneMsg := fmt.Sprintf("Updated %d subscription statuses", updated)
	if failed > 0 {
		doneMsg = fmt.Sprintf("Updated %d statuses, %d shops failed (see logs)", updated, failed)
		log.Printf("[queue] StatusProcessor: WARNING %d/%d shops failed to refresh for app %s (job %s)", failed, len(subscriptions), payload.AppID, payload.JobID)
	}
	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(subscriptions),
		Completed: len(subscriptions),
		Message:   doneMsg,
	})

	log.Printf("[queue] StatusProcessor: updated %d subscriptions (%d failed) for app %s (job %s)", updated, failed, payload.AppID, payload.JobID)
	return nil
}

package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// EventProcessor handles event_sync jobs (fetch app lifecycle events)
type EventProcessor struct {
	eventFetcher service.EventFetcher
	appEventRepo repository.AppEventRepository
	subRepo      repository.SubscriptionRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
	decryptor    queue.Decryptor
	syncJobRepo  repository.SyncJobRepository
	lockManager  *queue.LockManager
	progress     *queue.ProgressTracker
}

func NewEventProcessor(
	eventFetcher service.EventFetcher,
	appEventRepo repository.AppEventRepository,
	subRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor queue.Decryptor,
	syncJobRepo repository.SyncJobRepository,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *EventProcessor {
	return &EventProcessor{
		eventFetcher: eventFetcher,
		appEventRepo: appEventRepo,
		subRepo:      subRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
		decryptor:    decryptor,
		syncJobRepo:  syncJobRepo,
		lockManager:  lockManager,
		progress:     progress,
	}
}

func (p *EventProcessor) Type() string { return entity.SyncJobTypeEventSync }

func (p *EventProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	pCtx, err := queue.PrepareProcessorContext(ctx, payload, p.appRepo, p.partnerRepo, p.decryptor)
	if err != nil {
		return err
	}

	// Get subscriptions to iterate shop GIDs
	subscriptions, err := p.subRepo.FindByAppID(ctx, payload.AppID)
	if err != nil {
		return fmt.Errorf("failed to find subscriptions: %w", err)
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   len(subscriptions),
		Message: fmt.Sprintf("Fetching events for %d shops...", len(subscriptions)),
	})

	fetchCtx := external.WithOrganizationID(ctx, pCtx.OrganizationID)
	var allEvents []*entity.AppEvent
	completed := 0

	for _, sub := range subscriptions {
		if sub.ShopifyShopGID == "" {
			completed++
			continue
		}

		if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
			return fmt.Errorf("job cancelled")
		}

		events, err := p.eventFetcher.FetchAppEvents(fetchCtx, pCtx.OrganizationID, pCtx.AccessToken, pCtx.App.PartnerAppID, sub.ShopifyShopGID)
		if err != nil {
			completed++
			continue // Log but continue
		}

		for _, ev := range events {
			rawData, _ := json.Marshal(ev)
			allEvents = append(allEvents, entity.NewAppEvent(payload.AppID, sub.ShopifyShopGID, ev.Type, ev.OccurredAt, rawData))
		}

		completed++
		p.progress.Update(ctx, payload.JobID, queue.Progress{
			Total:     len(subscriptions),
			Completed: completed,
			Message:   fmt.Sprintf("Fetched events for %d/%d shops", completed, len(subscriptions)),
		})
	}

	// Upsert all events
	if len(allEvents) > 0 {
		if err := p.appEventRepo.UpsertBatch(ctx, allEvents); err != nil {
			return fmt.Errorf("failed to store events: %w", err)
		}
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(subscriptions),
		Completed: len(subscriptions),
		Message:   fmt.Sprintf("Stored %d events", len(allEvents)),
	})

	log.Printf("EventProcessor: stored %d events for app %s", len(allEvents), payload.AppID)
	_ = uuid.New() // suppress unused import if needed
	return p.syncJobRepo.MarkCompleted(ctx, payload.JobID)
}

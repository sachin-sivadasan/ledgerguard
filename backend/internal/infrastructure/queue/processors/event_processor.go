package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	// Map known (charged) shops' GID -> myshopify domain, so their events keep a
	// domain identifier the Store timeline / Events page can match on. Never-charged
	// shops (not in our subscriptions) fall back to the event's shop name/GID.
	subscriptions, err := p.subRepo.FindByAppID(ctx, payload.AppID)
	if err != nil {
		return fmt.Errorf("failed to find subscriptions: %w", err)
	}
	gidToDomain := make(map[string]string, len(subscriptions))
	for _, sub := range subscriptions {
		if sub.ShopifyShopGID != "" && sub.MyshopifyDomain != "" {
			gidToDomain[sub.ShopifyShopGID] = sub.MyshopifyDomain
		}
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Message: "Fetching app-wide lifecycle events...",
	})
	log.Printf("[queue] EventProcessor: fetching app-wide event stream for app %s (job %s)", payload.AppID, payload.JobID)

	// Fetch the ENTIRE app event stream (all shops, all lifecycle types) in one
	// paginated call — not per-shop. This captures every shop, including the free /
	// never-charged installs that have no subscription, so install metrics match
	// the Shopify Partner dashboard (previously we only saw charged shops).
	fetchCtx := external.WithOrganizationID(ctx, pCtx.OrganizationID)
	events, err := p.eventFetcher.FetchAppEvents(fetchCtx, pCtx.OrganizationID, pCtx.AccessToken, pCtx.App.PartnerAppID, "")
	if err != nil {
		return fmt.Errorf("failed to fetch app events: %w", err)
	}
	log.Printf("[queue] EventProcessor: fetched %d app-wide events for app %s — storing + computing installs (job %s)", len(events), payload.AppID, payload.JobID)

	if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
		return fmt.Errorf("job cancelled")
	}

	allEvents := make([]*entity.AppEvent, 0, len(events))
	for _, ev := range events {
		rawData, _ := json.Marshal(ev)
		shopIdentifier := gidToDomain[ev.ShopID]
		if shopIdentifier == "" {
			shopIdentifier = ev.ShopName
		}
		if shopIdentifier == "" {
			shopIdentifier = ev.ShopID
		}
		allEvents = append(allEvents, entity.NewAppEvent(payload.AppID, shopIdentifier, ev.Type, ev.OccurredAt, rawData))
	}

	if len(allEvents) > 0 {
		if err := p.appEventRepo.UpsertBatch(ctx, allEvents); err != nil {
			return fmt.Errorf("failed to store events: %w", err)
		}
	}

	// Derive install metrics from the relationship events and persist the active
	// install count (APPS-1: the Apps card's "installs" = currently-installed shops).
	activeInstalls, totalInstalls := external.CountInstalls(events)
	if pCtx.App.InstallCount != activeInstalls {
		pCtx.App.InstallCount = activeInstalls
		pCtx.App.UpdatedAt = time.Now().UTC()
		if err := p.appRepo.Update(ctx, pCtx.App); err != nil {
			log.Printf("[queue] EventProcessor: failed to persist install count (%d) for app %s: %v", activeInstalls, payload.AppID, err)
		}
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(allEvents),
		Completed: len(allEvents),
		Message:   fmt.Sprintf("Stored %d events; %d active installs (%d total)", len(allEvents), activeInstalls, totalInstalls),
	})

	log.Printf("[queue] EventProcessor: stored %d app-wide events, %d active / %d total installs for app %s (job %s)", len(allEvents), activeInstalls, totalInstalls, payload.AppID, payload.JobID)
	return nil
}

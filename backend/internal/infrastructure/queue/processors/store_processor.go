package processors

import (
	"context"
	"fmt"
	"log"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// StoreProcessor handles store_sync jobs (fetch shop brand data/logos)
type StoreProcessor struct {
	brandFetcher service.ShopBrandFetcher
	shopRepo     repository.ShopRepository
	subRepo      repository.SubscriptionRepository
	appRepo      repository.AppRepository
	partnerRepo  repository.PartnerAccountRepository
	decryptor    queue.Decryptor
	syncJobRepo  repository.SyncJobRepository
	lockManager  *queue.LockManager
	progress     *queue.ProgressTracker
}

func NewStoreProcessor(
	brandFetcher service.ShopBrandFetcher,
	shopRepo repository.ShopRepository,
	subRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor queue.Decryptor,
	syncJobRepo repository.SyncJobRepository,
	lockManager *queue.LockManager,
	progress *queue.ProgressTracker,
) *StoreProcessor {
	return &StoreProcessor{
		brandFetcher: brandFetcher,
		shopRepo:     shopRepo,
		subRepo:      subRepo,
		appRepo:      appRepo,
		partnerRepo:  partnerRepo,
		decryptor:    decryptor,
		syncJobRepo:  syncJobRepo,
		lockManager:  lockManager,
		progress:     progress,
	}
}

func (p *StoreProcessor) Type() string { return entity.SyncJobTypeStoreSync }

func (p *StoreProcessor) Process(ctx context.Context, payload *queue.SyncJobPayload) error {
	// We don't need the access token for storefront brand fetching, but validate the app exists
	_, err := p.appRepo.FindByID(ctx, payload.AppID)
	if err != nil {
		return fmt.Errorf("failed to find app %s: %w", payload.AppID, err)
	}

	subscriptions, err := p.subRepo.FindByAppID(ctx, payload.AppID)
	if err != nil {
		return fmt.Errorf("failed to find subscriptions: %w", err)
	}

	// Collect unique domains
	domainSet := make(map[string]bool)
	for _, sub := range subscriptions {
		if sub.MyshopifyDomain != "" {
			domainSet[sub.MyshopifyDomain] = true
		}
	}

	domains := make([]string, 0, len(domainSet))
	for d := range domainSet {
		domains = append(domains, d)
	}

	if len(domains) == 0 {
		p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{Message: "No domains to fetch brands for"})
		return nil
	}

	// Check which domains already exist
	existing, err := p.shopRepo.FindByDomains(ctx, domains)
	if err != nil {
		return fmt.Errorf("failed to check existing shops: %w", err)
	}

	// Filter to new domains only
	var newDomains []string
	for _, d := range domains {
		if _, found := existing[d]; !found {
			newDomains = append(newDomains, d)
		}
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   len(newDomains),
		Message: fmt.Sprintf("Fetching brand data for %d new domains...", len(newDomains)),
	})

	fetched := 0
	for i, domain := range newDomains {
		if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
			return fmt.Errorf("job cancelled")
		}

		shop, err := p.brandFetcher.FetchBrand(ctx, domain)
		if err != nil {
			continue // Not critical
		}

		if err := p.shopRepo.Upsert(ctx, shop); err != nil {
			continue
		}
		fetched++

		p.progress.Update(ctx, payload.JobID, queue.Progress{
			Total:     len(newDomains),
			Completed: i + 1,
			Message:   fmt.Sprintf("Fetched %d/%d shop brands", i+1, len(newDomains)),
		})
	}

	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(newDomains),
		Completed: len(newDomains),
		Message:   fmt.Sprintf("Fetched %d shop brands", fetched),
	})

	log.Printf("[queue] StoreProcessor: fetched %d brands for app %s (job %s)", fetched, payload.AppID, payload.JobID)
	return nil
}

package processors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// storeBrandWorkers controls concurrent HTTP fetches.
// Bounded by: mock server throughput (single-threaded Ruby) and
// DB pool (25 max conns shared with other handlers/workers).
const storeBrandWorkers = 10

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

	if len(newDomains) == 0 {
		p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{Message: "All shop brands already fetched"})
		return nil
	}

	p.progress.Update(ctx, payload.JobID, queue.Progress{
		Total:   len(newDomains),
		Message: fmt.Sprintf("Fetching brand data for %d new domains...", len(newDomains)),
	})

	// Use concurrent workers for large domain lists (whale persona: 50K shops)
	var fetched atomic.Int32
	var completed atomic.Int32
	sem := make(chan struct{}, storeBrandWorkers)
	var wg sync.WaitGroup

	for _, domain := range newDomains {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		default:
		}

		if int(completed.Load())%500 == 0 {
			if cancelled, _ := p.lockManager.IsCancelled(ctx, payload.JobID); cancelled {
				wg.Wait()
				return fmt.Errorf("job cancelled")
			}
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(d string) {
			defer wg.Done()
			defer func() { <-sem }()

			shop, err := p.brandFetcher.FetchBrand(ctx, d)
			if err != nil {
				completed.Add(1)
				return
			}
			if err := p.shopRepo.Upsert(ctx, shop); err != nil {
				completed.Add(1)
				return
			}
			fetched.Add(1)
			done := int(completed.Add(1))

			if done%500 == 0 || done == len(newDomains) {
				p.progress.Update(ctx, payload.JobID, queue.Progress{
					Total:     len(newDomains),
					Completed: done,
					Message:   fmt.Sprintf("Fetched %d/%d shop brands (%d successful)", done, len(newDomains), fetched.Load()),
				})
			}
		}(domain)
	}

	wg.Wait()

	totalFetched := int(fetched.Load())
	p.progress.ForceUpdate(ctx, payload.JobID, queue.Progress{
		Total:     len(newDomains),
		Completed: len(newDomains),
		Message:   fmt.Sprintf("Fetched %d shop brands", totalFetched),
	})

	log.Printf("[queue] StoreProcessor: fetched %d/%d brands for app %s (job %s)", totalFetched, len(newDomains), payload.AppID, payload.JobID)
	return nil
}

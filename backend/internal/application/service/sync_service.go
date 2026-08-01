package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
)

// TransactionFetcher interface for fetching transactions from external API
type TransactionFetcher interface {
	FetchTransactions(ctx context.Context, accessToken string, appID uuid.UUID, from, to time.Time) ([]*entity.Transaction, error)
}

// ShopBrandFetcher interface for fetching shop brand data
type ShopBrandFetcher interface {
	FetchBrand(ctx context.Context, myshopifyDomain string) (*entity.Shop, error)
}

// EventFetcher interface for fetching app lifecycle events
type EventFetcher interface {
	FetchAppEvents(ctx context.Context, organizationID, accessToken string, appGID string, shopGID string) ([]external.AppEvent, error)
}

// ReviewScraper interface for scraping app reviews
type ReviewScraper interface {
	ScrapeReviews(ctx context.Context, slug string, maxPages int) ([]external.ScrapedReview, error)
}

// Decryptor interface for decrypting tokens
type Decryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// LedgerRebuilder interface for rebuilding ledger after sync
type LedgerRebuilder interface {
	RebuildFromTransactions(ctx context.Context, appID uuid.UUID, now time.Time) (*domainservice.LedgerRebuildResult, error)
	BackfillHistoricalSnapshots(ctx context.Context, appID uuid.UUID, transactions []*entity.Transaction) (int, error)
	RefreshTodaySnapshot(ctx context.Context, appID uuid.UUID) error
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	AppID            uuid.UUID
	AppName          string
	TransactionCount int
	RiskSummary      *domainservice.RiskSummary
	RevenueAtRisk    int64
	TotalMRRCents    int64
	SyncedAt         time.Time
	Error            error
}

// ReadModelRebuilder rebuilds the Revenue API read model after sync
type ReadModelRebuilder interface {
	RebuildForApp(ctx context.Context, appID uuid.UUID) error
}

// SyncService handles synchronization of transactions from Partner API
type SyncService struct {
	fetcher          TransactionFetcher
	eventFetcher     EventFetcher
	brandFetcher     ShopBrandFetcher
	reviewScraper    ReviewScraper
	txRepo           repository.TransactionRepository
	subRepo          repository.SubscriptionRepository
	shopRepo         repository.ShopRepository
	reviewRepo       repository.AppReviewRepository
	appRepo          repository.AppRepository
	partnerRepo      repository.PartnerAccountRepository
	decryptor        Decryptor
	ledger           LedgerRebuilder
	readModelBuilder ReadModelRebuilder
}

func NewSyncService(
	fetcher TransactionFetcher,
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
	decryptor Decryptor,
	ledger LedgerRebuilder,
) *SyncService {
	return &SyncService{
		fetcher:     fetcher,
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
		decryptor:   decryptor,
		ledger:      ledger,
	}
}

// WithEventFetcher adds an event fetcher for subscription status enrichment
func (s *SyncService) WithEventFetcher(fetcher EventFetcher) *SyncService {
	s.eventFetcher = fetcher
	return s
}

// WithSubscriptionRepo adds a subscription repository for status updates
func (s *SyncService) WithSubscriptionRepo(repo repository.SubscriptionRepository) *SyncService {
	s.subRepo = repo
	return s
}

// WithShopBrandFetcher adds a brand fetcher and shop repo for logo fetching during sync
func (s *SyncService) WithShopBrandFetcher(fetcher ShopBrandFetcher, shopRepo repository.ShopRepository) *SyncService {
	s.brandFetcher = fetcher
	s.shopRepo = shopRepo
	return s
}

// WithReviewScraper adds a review scraper for fetching app store reviews during sync
func (s *SyncService) WithReviewScraper(scraper ReviewScraper, reviewRepo repository.AppReviewRepository) *SyncService {
	s.reviewScraper = scraper
	s.reviewRepo = reviewRepo
	return s
}

// WithReadModelBuilder adds a read model builder for Revenue API
func (s *SyncService) WithReadModelBuilder(builder ReadModelRebuilder) *SyncService {
	s.readModelBuilder = builder
	return s
}

// TriggerSync starts a background sync for a newly-selected app (fire-and-forget).
// Used as a fallback when queue-based sync is not enabled.
func (s *SyncService) TriggerSync(ctx context.Context, appID, userID, partnerAccountID uuid.UUID) error {
	go func() {
		bgCtx := context.Background()
		_, _ = s.SyncApp(bgCtx, appID)
	}()
	return nil
}

// SyncApp synchronizes transactions for a single app
func (s *SyncService) SyncApp(ctx context.Context, appID uuid.UUID) (*SyncResult, error) {
	// Check if fetcher is configured
	if s.fetcher == nil {
		return nil, fmt.Errorf("transaction fetcher not configured")
	}

	// Get app
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to find app: %w", err)
	}

	// Get partner account for the app
	partnerAccount, err := s.getPartnerAccountForApp(ctx, app.PartnerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner account: %w", err)
	}

	// Decrypt access token
	accessToken, err := s.decryptor.Decrypt(partnerAccount.EncryptedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Fetch the app's entire transaction history from the beginning — see SyncHistoryStart.
	now := time.Now().UTC()
	from := domainservice.SyncHistoryStart
	to := now

	// Add organization ID and app GID to context for the Partner API client
	fetchCtx := external.WithOrganizationID(ctx, partnerAccount.PartnerID)
	fetchCtx = external.WithPartnerAppGID(fetchCtx, app.PartnerAppID)

	// Fetch transactions from Partner API (filtered by app)
	transactions, err := s.fetcher.FetchTransactions(fetchCtx, string(accessToken), appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// Process earnings tracking for each transaction
	earningsCalc := domainservice.NewEarningsCalculator()
	earningsCalc.ProcessTransactions(transactions, now)

	// Store transactions (upsert for idempotency)
	if len(transactions) > 0 {
		if err := s.txRepo.UpsertBatch(ctx, transactions); err != nil {
			return nil, fmt.Errorf("failed to store transactions: %w", err)
		}
	}

	// Rebuild ledger and recalculate risk states
	var riskSummary *domainservice.RiskSummary
	var revenueAtRisk int64
	var totalMRR int64

	if s.ledger != nil {
		rebuildResult, err := s.ledger.RebuildFromTransactions(ctx, appID, now)
		if err != nil {
			return nil, fmt.Errorf("failed to rebuild ledger: %w", err)
		}
		riskSummary = &rebuildResult.RiskSummary
		totalMRR = rebuildResult.TotalMRRCents

		// Calculate revenue at risk (ONE_CYCLE_MISSED + TWO_CYCLES_MISSED MRR)
		// This would require access to subscriptions, simplified here
		revenueAtRisk = 0 // Will be calculated by caller if needed

		// Backfill historical snapshots from the app's ENTIRE stored history — see
		// SyncHistoryStart. Backfill is best-effort (not fatal to the sync), but errors are
		// logged rather than swallowed: over full history a silent failure would corrupt a
		// large span of the permanent snapshot audit trail (CLAUDE.md §9).
		if allTransactions, findErr := s.txRepo.FindByAppID(ctx, appID, domainservice.SyncHistoryStart, now); findErr != nil {
			log.Printf("SyncService: snapshot backfill skipped — failed to load transactions for app %s: %v", appID, findErr)
		} else if len(allTransactions) > 0 {
			if _, bfErr := s.ledger.BackfillHistoricalSnapshots(ctx, appID, allTransactions); bfErr != nil {
				log.Printf("SyncService: snapshot backfill failed for app %s: %v", appID, bfErr)
			}
		}

		// Enrich subscription status from app events (if configured)
		if s.eventFetcher != nil && s.subRepo != nil {
			_ = s.enrichSubscriptionStatus(fetchCtx, app, partnerAccount, string(accessToken))
			// Ignore enrichment errors - status defaults to ACTIVE

			// Backfill computed today's snapshot from charge-only risk; recompute it
			// now from the reconciled subscriptions so the Dashboard KPIs match the
			// Subscriptions/Risk pages (RISK-1). Best-effort, non-fatal.
			if bfErr := s.ledger.RefreshTodaySnapshot(ctx, appID); bfErr != nil {
				log.Printf("SyncService: refresh today snapshot failed for app %s: %v", appID, bfErr)
			}
		}

		// Fetch shop brand data for new domains (if configured)
		if s.brandFetcher != nil && s.shopRepo != nil && s.subRepo != nil {
			_ = s.fetchShopBrands(ctx, appID)
			// Ignore brand fetch errors - logos are not critical
		}

		// Scrape app store reviews (if configured and slug is set)
		if s.reviewScraper != nil && s.reviewRepo != nil && app.AppStoreSlug != "" {
			_ = s.scrapeAndStoreReviews(ctx, app)
			// Ignore review scrape errors - reviews are not critical
		}

		// Rebuild Revenue API read model
		if s.readModelBuilder != nil {
			_ = s.readModelBuilder.RebuildForApp(ctx, appID)
			// Non-fatal — don't fail the sync for read model issues
		}
	}

	return &SyncResult{
		AppID:            appID,
		AppName:          app.Name,
		TransactionCount: len(transactions),
		RiskSummary:      riskSummary,
		RevenueAtRisk:    revenueAtRisk,
		TotalMRRCents:    totalMRR,
		SyncedAt:         now,
	}, nil
}

// SyncAllApps synchronizes transactions for all apps of a partner account
func (s *SyncService) SyncAllApps(ctx context.Context, partnerAccountID uuid.UUID) ([]*SyncResult, error) {
	// Get all apps for the partner account
	apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find apps: %w", err)
	}

	var results []*SyncResult

	for _, app := range apps {
		if !app.TrackingEnabled {
			continue
		}

		result, err := s.SyncApp(ctx, app.ID)
		if err != nil {
			results = append(results, &SyncResult{
				AppID:         app.ID,
				AppName:       app.Name,
				SyncedAt:      time.Now().UTC(),
				Error:         err,
				RiskSummary:   nil,
				RevenueAtRisk: 0,
				TotalMRRCents: 0,
			})
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SyncService) getPartnerAccountForApp(ctx context.Context, partnerAccountID uuid.UUID) (*entity.PartnerAccount, error) {
	return s.partnerRepo.FindByID(ctx, partnerAccountID)
}

// enrichSubscriptionStatus updates subscription status based on app lifecycle events
// This provides accurate status (ACTIVE, CANCELLED, UNINSTALLED) instead of defaulting to ACTIVE
func (s *SyncService) enrichSubscriptionStatus(ctx context.Context, app *entity.App, partnerAccount *entity.PartnerAccount, accessToken string) error {
	// Get all subscriptions for this app
	subscriptions, err := s.subRepo.FindByAppID(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("failed to find subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		return nil
	}

	// For each subscription with a shop GID, fetch events and update status
	for _, sub := range subscriptions {
		if sub.ShopifyShopGID == "" {
			continue // Can't fetch events without shop GID
		}

		// Fetch events for this shop
		events, err := s.eventFetcher.FetchAppEvents(ctx, partnerAccount.PartnerID, accessToken, app.PartnerAppID, sub.ShopifyShopGID)
		if err != nil {
			// Log but continue - don't fail the whole sync for event fetch errors
			continue
		}

		// Reconcile the event-derived status against billing reality (see
		// Subscription.ApplyEventStatus): a terminal CANCELLED/UNINSTALLED churns
		// only when no recurring charge post-dates the event (else it's a stale/
		// plan-change cancel), and risk is re-derived from charge recency.
		newStatus, statusAt := external.GetLatestSubscriptionStatusWithTime(events)
		if sub.ApplyEventStatus(newStatus, statusAt, time.Now().UTC()) {
			// Upsert the updated subscription
			if err := s.subRepo.Upsert(ctx, sub); err != nil {
				continue // Log but don't fail
			}
		}
	}

	return nil
}

// scrapeAndStoreReviews scrapes recent reviews from the Shopify App Store and stores them
func (s *SyncService) scrapeAndStoreReviews(ctx context.Context, app *entity.App) error {
	// During sync, scrape only 2 pages (~20 reviews) to stay fast
	scraped, err := s.reviewScraper.ScrapeReviews(ctx, app.AppStoreSlug, 2)
	if err != nil {
		return fmt.Errorf("failed to scrape reviews for %s: %w", app.AppStoreSlug, err)
	}

	if len(scraped) == 0 {
		return nil
	}

	now := time.Now().UTC()
	var reviews []*entity.AppReview
	for _, sr := range scraped {
		if sr.Rating == 0 || sr.Date.IsZero() {
			continue
		}
		reviews = append(reviews, &entity.AppReview{
			ID:             uuid.New(),
			AppID:          app.ID,
			SourceReviewID: sr.SourceReviewID(),
			Author:         sr.Author,
			Rating:         sr.Rating,
			Body:           sr.Body,
			ReviewDate:     sr.Date,
			Location:       sr.Location,
			TimeUsing:      sr.TimeUsing,
			Source:         "shopify_app_store",
			ScrapedAt:      now,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	if len(reviews) > 0 {
		return s.reviewRepo.UpsertBatch(ctx, reviews)
	}
	return nil
}

// fetchShopBrands fetches brand data for shop domains not yet in the shops table
func (s *SyncService) fetchShopBrands(ctx context.Context, appID uuid.UUID) error {
	subscriptions, err := s.subRepo.FindByAppID(ctx, appID)
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

	if len(domainSet) == 0 {
		return nil
	}

	domains := make([]string, 0, len(domainSet))
	for d := range domainSet {
		domains = append(domains, d)
	}

	// Check which domains are already in the shops table
	existing, err := s.shopRepo.FindByDomains(ctx, domains)
	if err != nil {
		return fmt.Errorf("failed to find existing shops: %w", err)
	}

	// Fetch brand data only for new domains
	for _, domain := range domains {
		if _, found := existing[domain]; found {
			continue
		}

		shop, err := s.brandFetcher.FetchBrand(ctx, domain)
		if err != nil {
			// Log but continue - brand data is not critical
			continue
		}

		if err := s.shopRepo.Upsert(ctx, shop); err != nil {
			continue
		}
	}

	return nil
}

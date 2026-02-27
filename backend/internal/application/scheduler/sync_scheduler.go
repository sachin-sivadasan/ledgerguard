package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// SyncScheduler handles scheduled synchronization of transactions
type SyncScheduler struct {
	syncService *service.SyncService
	partnerRepo repository.PartnerAccountRepository
	interval    time.Duration
	stopCh      chan struct{}
	doneCh      chan struct{}
}

// NewSyncScheduler creates a new SyncScheduler with 12-hour interval
func NewSyncScheduler(
	syncService *service.SyncService,
	partnerRepo repository.PartnerAccountRepository,
) *SyncScheduler {
	return &SyncScheduler{
		syncService: syncService,
		partnerRepo: partnerRepo,
		interval:    12 * time.Hour,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *SyncScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop gracefully stops the scheduler
func (s *SyncScheduler) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

func (s *SyncScheduler) run(ctx context.Context) {
	defer close(s.doneCh)

	// Calculate time until next 00:00 or 12:00 UTC
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial sync
	s.syncAll(ctx)

	for {
		select {
		case <-ticker.C:
			s.syncAll(ctx)
		case <-s.stopCh:
			log.Println("Sync scheduler stopped")
			return
		case <-ctx.Done():
			log.Println("Sync scheduler context cancelled")
			return
		}
	}
}

func (s *SyncScheduler) syncAll(ctx context.Context) {
	log.Println("Starting scheduled sync...")

	// Get all unique partner account IDs from apps
	partnerAccountIDs, err := s.getPartnerAccountIDs(ctx)
	if err != nil {
		log.Printf("Failed to get partner accounts: %v", err)
		return
	}

	for _, partnerAccountID := range partnerAccountIDs {
		results, err := s.syncService.SyncAllApps(ctx, partnerAccountID)
		if err != nil {
			log.Printf("Failed to sync apps for partner %s: %v", partnerAccountID, err)
			continue
		}

		for _, result := range results {
			if result.Error != nil {
				log.Printf("Sync error for app %s: %v", result.AppName, result.Error)
			} else {
				log.Printf("Synced %d transactions for app %s", result.TransactionCount, result.AppName)
			}
		}
	}

	log.Println("Scheduled sync completed")
}

func (s *SyncScheduler) getPartnerAccountIDs(ctx context.Context) ([]uuid.UUID, error) {
	return s.partnerRepo.GetAllIDs(ctx)
}

// SetInterval allows customizing the sync interval (for testing)
func (s *SyncScheduler) SetInterval(interval time.Duration) {
	s.interval = interval
}

// RunOnce performs a single sync cycle (for testing)
func (s *SyncScheduler) RunOnce(ctx context.Context) {
	s.syncAll(ctx)
}

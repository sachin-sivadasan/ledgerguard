package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"go.uber.org/zap"
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
			logging.FromContext(ctx).Info("sync scheduler stopped")
			return
		case <-ctx.Done():
			logging.FromContext(ctx).Info("sync scheduler context cancelled")
			return
		}
	}
}

func (s *SyncScheduler) syncAll(ctx context.Context) {
	logging.FromContext(ctx).Info("starting scheduled sync")

	// Get all unique partner account IDs from apps
	partnerAccountIDs, err := s.getPartnerAccountIDs(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("failed to get partner accounts", zap.Error(err))
		return
	}

	for _, partnerAccountID := range partnerAccountIDs {
		results, err := s.syncService.SyncAllApps(ctx, partnerAccountID)
		if err != nil {
			logging.FromContext(ctx).Warn("failed to sync apps for partner", zap.String("partner_id", partnerAccountID.String()), zap.Error(err))
			continue
		}

		for _, result := range results {
			if result.Error != nil {
				logging.FromContext(ctx).Warn("sync error for app", zap.String("app_name", result.AppName), zap.Error(result.Error))
			} else {
				logging.FromContext(ctx).Info("synced transactions for app", zap.Int("count", result.TransactionCount), zap.String("app_name", result.AppName))
			}
		}
	}

	logging.FromContext(ctx).Info("scheduled sync completed")
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

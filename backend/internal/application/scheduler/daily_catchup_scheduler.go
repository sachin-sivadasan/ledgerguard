package scheduler

import (
	"context"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"go.uber.org/zap"
)

// DailyCatchupScheduler syncs the last N days of transactions and events
// to fill gaps caused by missed webhooks or downtime.
type DailyCatchupScheduler struct {
	queueSyncSvc  *service.QueueSyncService
	appRepo       repository.AppRepository
	partnerRepo   repository.PartnerAccountRepository
	targetHour    int // UTC hour to run (default 3)
	lookbackDays  int // days to look back (default 2)
	checkInterval time.Duration
	lastRunDate   string // YYYY-MM-DD to avoid double-runs
	stopCh        chan struct{}
	doneCh        chan struct{}
}

// NewDailyCatchupScheduler creates a new DailyCatchupScheduler.
func NewDailyCatchupScheduler(
	queueSyncSvc *service.QueueSyncService,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *DailyCatchupScheduler {
	return &DailyCatchupScheduler{
		queueSyncSvc:  queueSyncSvc,
		appRepo:       appRepo,
		partnerRepo:   partnerRepo,
		targetHour:    3,
		lookbackDays:  2,
		checkInterval: 15 * time.Minute,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
}

// Start begins the scheduler loop.
func (s *DailyCatchupScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop gracefully stops the scheduler.
func (s *DailyCatchupScheduler) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

func (s *DailyCatchupScheduler) run(ctx context.Context) {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// Check immediately on start
	s.check(ctx)

	for {
		select {
		case <-ticker.C:
			s.check(ctx)
		case <-s.stopCh:
			logging.FromContext(ctx).Info("daily catchup scheduler stopped")
			return
		case <-ctx.Done():
			logging.FromContext(ctx).Info("daily catchup scheduler context cancelled")
			return
		}
	}
}

func (s *DailyCatchupScheduler) check(ctx context.Context) {
	now := time.Now().UTC()
	if now.Hour() != s.targetHour {
		return
	}

	today := now.Format("2006-01-02")
	if today == s.lastRunDate {
		return
	}

	s.lastRunDate = today
	logging.FromContext(ctx).Info("running daily catchup sync", zap.Int("hour", s.targetHour), zap.Int("lookback_days", s.lookbackDays))

	count := s.enqueueAll(ctx, s.lookbackDays)
	logging.FromContext(ctx).Info("daily catchup complete", zap.Int("count", count))
}

// RunOnce triggers the catchup sync immediately for all apps.
// Returns the number of jobs enqueued.
func (s *DailyCatchupScheduler) RunOnce(ctx context.Context, lookbackDays int) int {
	if lookbackDays <= 0 {
		lookbackDays = s.lookbackDays
	}
	logging.FromContext(ctx).Info("manual trigger", zap.Int("lookback_days", lookbackDays))
	return s.enqueueAll(ctx, lookbackDays)
}

func (s *DailyCatchupScheduler) enqueueAll(ctx context.Context, lookbackDays int) int {
	partnerIDs, err := s.partnerRepo.GetAllIDs(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("failed to get partner accounts", zap.Error(err))
		return 0
	}

	jobCount := 0
	for _, partnerID := range partnerIDs {
		partner, err := s.partnerRepo.FindByID(ctx, partnerID)
		if err != nil {
			logging.FromContext(ctx).Warn("failed to find partner", zap.String("partner_id", partnerID.String()), zap.Error(err))
			continue
		}

		apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerID)
		if err != nil {
			logging.FromContext(ctx).Warn("failed to find apps for partner", zap.String("partner_id", partnerID.String()), zap.Error(err))
			continue
		}

		for _, app := range apps {
			// Enqueue transaction_sync with lookback
			if job, err := s.queueSyncSvc.EnqueueCatchupSync(ctx, app.ID, partner.UserID, partnerID, entity.SyncJobTypeTransactionSync, lookbackDays); err != nil {
				logging.FromContext(ctx).Warn("failed to enqueue transaction_sync for app", zap.String("app_id", app.ID.String()), zap.Error(err))
			} else if job != nil {
				jobCount++
			}

			// Enqueue event_sync with lookback
			if job, err := s.queueSyncSvc.EnqueueCatchupSync(ctx, app.ID, partner.UserID, partnerID, entity.SyncJobTypeEventSync, lookbackDays); err != nil {
				logging.FromContext(ctx).Warn("failed to enqueue event_sync for app", zap.String("app_id", app.ID.String()), zap.Error(err))
			} else if job != nil {
				jobCount++
			}
		}
	}

	return jobCount
}

// SetTargetHour changes the UTC hour at which the scheduler runs.
func (s *DailyCatchupScheduler) SetTargetHour(hour int) {
	s.targetHour = hour
}

// SetCheckInterval allows customizing the check interval (for testing).
func (s *DailyCatchupScheduler) SetCheckInterval(interval time.Duration) {
	s.checkInterval = interval
}

package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
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
			log.Println("[catchup] Daily catchup scheduler stopped")
			return
		case <-ctx.Done():
			log.Println("[catchup] Daily catchup scheduler context cancelled")
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
	log.Printf("[catchup] Running daily catchup sync (hour=%d, lookback=%d days)", s.targetHour, s.lookbackDays)

	count := s.enqueueAll(ctx, s.lookbackDays)
	log.Printf("[catchup] Daily catchup complete: enqueued %d jobs", count)
}

// RunOnce triggers the catchup sync immediately for all apps.
// Returns the number of jobs enqueued.
func (s *DailyCatchupScheduler) RunOnce(ctx context.Context, lookbackDays int) int {
	if lookbackDays <= 0 {
		lookbackDays = s.lookbackDays
	}
	log.Printf("[catchup] Manual trigger: lookback=%d days", lookbackDays)
	return s.enqueueAll(ctx, lookbackDays)
}

func (s *DailyCatchupScheduler) enqueueAll(ctx context.Context, lookbackDays int) int {
	partnerIDs, err := s.partnerRepo.GetAllIDs(ctx)
	if err != nil {
		log.Printf("[catchup] Failed to get partner accounts: %v", err)
		return 0
	}

	jobCount := 0
	for _, partnerID := range partnerIDs {
		partner, err := s.partnerRepo.FindByID(ctx, partnerID)
		if err != nil {
			log.Printf("[catchup] Failed to find partner %s: %v", partnerID, err)
			continue
		}

		apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerID)
		if err != nil {
			log.Printf("[catchup] Failed to find apps for partner %s: %v", partnerID, err)
			continue
		}

		for _, app := range apps {
			// Enqueue transaction_sync with lookback
			if job, err := s.queueSyncSvc.EnqueueCatchupSync(ctx, app.ID, partner.UserID, partnerID, entity.SyncJobTypeTransactionSync, lookbackDays); err != nil {
				log.Printf("[catchup] Failed to enqueue transaction_sync for app %s: %v", app.ID, err)
			} else if job != nil {
				jobCount++
			}

			// Enqueue event_sync with lookback
			if job, err := s.queueSyncSvc.EnqueueCatchupSync(ctx, app.ID, partner.UserID, partnerID, entity.SyncJobTypeEventSync, lookbackDays); err != nil {
				log.Printf("[catchup] Failed to enqueue event_sync for app %s: %v", app.ID, err)
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

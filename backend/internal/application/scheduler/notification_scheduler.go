package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// NotificationScheduler handles scheduled daily summary notifications
type NotificationScheduler struct {
	notificationSvc *service.NotificationService
	prefsRepo       repository.NotificationPreferencesRepository
	snapshotRepo    repository.DailyMetricsSnapshotRepository
	appRepo         repository.AppRepository
	partnerRepo     repository.PartnerAccountRepository
	checkInterval   time.Duration
	lastCheckedHour int
	stopCh          chan struct{}
	doneCh          chan struct{}
}

// NewNotificationScheduler creates a new NotificationScheduler
func NewNotificationScheduler(
	notificationSvc *service.NotificationService,
	prefsRepo repository.NotificationPreferencesRepository,
	snapshotRepo repository.DailyMetricsSnapshotRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *NotificationScheduler {
	return &NotificationScheduler{
		notificationSvc: notificationSvc,
		prefsRepo:       prefsRepo,
		snapshotRepo:    snapshotRepo,
		appRepo:         appRepo,
		partnerRepo:     partnerRepo,
		checkInterval:   15 * time.Minute,
		lastCheckedHour: -1,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *NotificationScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop gracefully stops the scheduler
func (s *NotificationScheduler) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

func (s *NotificationScheduler) run(ctx context.Context) {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// Check immediately on start
	s.checkAndSend(ctx)

	for {
		select {
		case <-ticker.C:
			s.checkAndSend(ctx)
		case <-s.stopCh:
			log.Println("Notification scheduler stopped")
			return
		case <-ctx.Done():
			log.Println("Notification scheduler context cancelled")
			return
		}
	}
}

func (s *NotificationScheduler) checkAndSend(ctx context.Context) {
	now := time.Now().UTC()
	currentHour := now.Hour()

	// Skip if we already checked this hour
	if currentHour == s.lastCheckedHour {
		return
	}

	s.lastCheckedHour = currentHour
	log.Printf("[notif] Checking for daily summary at hour %d UTC", currentHour)

	// Find users who have daily summary enabled at this hour
	userIDs, err := s.prefsRepo.FindUsersWithDailySummaryAtHour(ctx, currentHour)
	if err != nil {
		log.Printf("[notif] Failed to query users for daily summary: %v", err)
		return
	}

	if len(userIDs) == 0 {
		log.Printf("[notif] No users with daily summary at hour %d UTC", currentHour)
		return
	}

	log.Printf("[notif] Found %d users for daily summary at hour %d UTC", len(userIDs), currentHour)

	// Send daily summary to each user
	for _, userID := range userIDs {
		s.sendDailySummaryToUser(ctx, userID)
	}
}

func (s *NotificationScheduler) sendDailySummaryToUser(ctx context.Context, userID uuid.UUID) {
	// Get partner account for this user
	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil || partnerAccount == nil {
		log.Printf("[notif] No partner account for user %s, skipping", userID)
		return
	}

	// Get apps for the partner account
	apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerAccount.ID)
	if err != nil {
		log.Printf("[notif] Failed to find apps for user %s: %v", userID, err)
		return
	}

	log.Printf("[notif] User %s has %d apps, sending summaries", userID, len(apps))

	for _, app := range apps {
		snapshot, err := s.snapshotRepo.FindLatestByAppID(ctx, app.ID)
		if err != nil {
			log.Printf("[notif] No snapshot for app %s (%s), skipping: %v", app.Name, app.ID, err)
			continue
		}

		if err := s.notificationSvc.SendDailySummary(ctx, userID, app.Name, snapshot); err != nil {
			log.Printf("[notif] Failed to send summary for app %s to user %s: %v", app.Name, userID, err)
		} else {
			log.Printf("[notif] Sent daily summary for app %s to user %s (MRR: %d cents)", app.Name, userID, snapshot.ActiveMRRCents)
		}
	}
}

// SetCheckInterval allows customizing the check interval (for testing)
func (s *NotificationScheduler) SetCheckInterval(interval time.Duration) {
	s.checkInterval = interval
}

// RunOnce performs a single check cycle (for testing)
func (s *NotificationScheduler) RunOnce(ctx context.Context) {
	// Reset last checked hour to force a check
	s.lastCheckedHour = -1
	s.checkAndSend(ctx)
}

// RunForHour triggers daily summary for a specific UTC hour (admin/testing).
// Returns the number of users notified.
func (s *NotificationScheduler) RunForHour(ctx context.Context, hour int) int {
	log.Printf("[notif] Admin trigger: running daily summary for hour %d UTC", hour)

	userIDs, err := s.prefsRepo.FindUsersWithDailySummaryAtHour(ctx, hour)
	if err != nil {
		log.Printf("[notif] Admin trigger: failed to query users: %v", err)
		return 0
	}

	if len(userIDs) == 0 {
		log.Printf("[notif] Admin trigger: no users with daily summary at hour %d", hour)
		return 0
	}

	log.Printf("[notif] Admin trigger: sending to %d users", len(userIDs))
	for _, userID := range userIDs {
		s.sendDailySummaryToUser(ctx, userID)
	}

	return len(userIDs)
}

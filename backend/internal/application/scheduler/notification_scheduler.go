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
	log.Printf("Checking for daily summary notifications at hour %d UTC", currentHour)

	// Find users who have daily summary enabled at this hour
	userIDs, err := s.prefsRepo.FindUsersWithDailySummaryAtHour(ctx, currentHour)
	if err != nil {
		log.Printf("Failed to find users for daily summary: %v", err)
		return
	}

	if len(userIDs) == 0 {
		return
	}

	log.Printf("Found %d users for daily summary at hour %d", len(userIDs), currentHour)

	// Send daily summary to each user
	for _, userID := range userIDs {
		s.sendDailySummaryToUser(ctx, userID)
	}
}

func (s *NotificationScheduler) sendDailySummaryToUser(ctx context.Context, userID uuid.UUID) {
	// Get partner account for this user
	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil || partnerAccount == nil {
		log.Printf("No partner account found for user %s", userID)
		return
	}

	// Get apps for the partner account
	apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerAccount.ID)
	if err != nil {
		log.Printf("Failed to find apps for partner account %s: %v", partnerAccount.ID, err)
		return
	}

	// Send daily summary for each app
	for _, app := range apps {
		// Get latest snapshot for the app
		snapshot, err := s.snapshotRepo.FindLatestByAppID(ctx, app.ID)
		if err != nil {
			log.Printf("Failed to get latest snapshot for app %s: %v", app.Name, err)
			continue
		}

		// Send daily summary notification
		if err := s.notificationSvc.SendDailySummary(ctx, userID, app.Name, snapshot); err != nil {
			log.Printf("Failed to send daily summary for app %s to user %s: %v", app.Name, userID, err)
		} else {
			log.Printf("Sent daily summary for app %s to user %s", app.Name, userID)
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

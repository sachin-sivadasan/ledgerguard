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
			logging.FromContext(ctx).Info("notification scheduler stopped")
			return
		case <-ctx.Done():
			logging.FromContext(ctx).Info("notification scheduler context cancelled")
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
	logging.FromContext(ctx).Info("checking for daily summary", zap.Int("hour", currentHour))

	// Find users who have daily summary enabled at this hour
	userIDs, err := s.prefsRepo.FindUsersWithDailySummaryAtHour(ctx, currentHour)
	if err != nil {
		logging.FromContext(ctx).Error("failed to query users for daily summary", zap.Error(err))
		return
	}

	if len(userIDs) == 0 {
		logging.FromContext(ctx).Info("no users with daily summary at hour", zap.Int("hour", currentHour))
		return
	}

	logging.FromContext(ctx).Info("found users for daily summary", zap.Int("count", len(userIDs)), zap.Int("hour", currentHour))

	// Send daily summary to each user
	for _, userID := range userIDs {
		s.sendDailySummaryToUser(ctx, userID)
	}
}

func (s *NotificationScheduler) sendDailySummaryToUser(ctx context.Context, userID uuid.UUID) {
	// Get partner account for this user
	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil || partnerAccount == nil {
		logging.FromContext(ctx).Warn("no partner account for user, skipping", zap.String("user_id", userID.String()))
		return
	}

	// Get apps for the partner account
	apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerAccount.ID)
	if err != nil {
		logging.FromContext(ctx).Error("failed to find apps for user", zap.String("user_id", userID.String()), zap.Error(err))
		return
	}

	logging.FromContext(ctx).Info("user apps found, sending summaries", zap.String("user_id", userID.String()), zap.Int("count", len(apps)))

	for _, app := range apps {
		snapshot, err := s.snapshotRepo.FindLatestByAppID(ctx, app.ID)
		if err != nil {
			logging.FromContext(ctx).Warn("no snapshot for app, skipping", zap.String("app_name", app.Name), zap.String("app_id", app.ID.String()), zap.Error(err))
			continue
		}

		if err := s.notificationSvc.SendDailySummary(ctx, userID, app.Name, snapshot); err != nil {
			logging.FromContext(ctx).Warn("failed to send summary for app", zap.String("app_name", app.Name), zap.String("user_id", userID.String()), zap.Error(err))
		} else {
			logging.FromContext(ctx).Info("sent daily summary for app", zap.String("app_name", app.Name), zap.String("user_id", userID.String()), zap.Int64("mrr_cents", snapshot.ActiveMRRCents))
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
	logging.FromContext(ctx).Info("admin trigger: running daily summary for hour", zap.Int("hour", hour))

	userIDs, err := s.prefsRepo.FindUsersWithDailySummaryAtHour(ctx, hour)
	if err != nil {
		logging.FromContext(ctx).Error("admin trigger: failed to query users", zap.Error(err))
		return 0
	}

	if len(userIDs) == 0 {
		logging.FromContext(ctx).Info("admin trigger: no users with daily summary at hour", zap.Int("hour", hour))
		return 0
	}

	logging.FromContext(ctx).Info("admin trigger: sending to users", zap.Int("count", len(userIDs)))
	for _, userID := range userIDs {
		s.sendDailySummaryToUser(ctx, userID)
	}

	return len(userIDs)
}

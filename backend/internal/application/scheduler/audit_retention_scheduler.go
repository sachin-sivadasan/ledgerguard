package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
)

// AuditRetentionScheduler prunes old audit-log rows once per day.
type AuditRetentionScheduler struct {
	svc      *service.AuditRetentionService
	interval time.Duration
	stopCh   chan struct{}
}

// NewAuditRetentionScheduler creates a daily audit-retention scheduler.
func NewAuditRetentionScheduler(svc *service.AuditRetentionService) *AuditRetentionScheduler {
	return &AuditRetentionScheduler{
		svc:      svc,
		interval: 24 * time.Hour,
		stopCh:   make(chan struct{}),
	}
}

// Start runs an initial prune, then prunes every interval until Stop or ctx is done.
func (s *AuditRetentionScheduler) Start(ctx context.Context) {
	log.Println("Audit retention scheduler started (daily)")
	go func() {
		s.svc.PruneOnce(ctx, time.Now().UTC())
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.svc.PruneOnce(ctx, time.Now().UTC())
			}
		}
	}()
}

// Stop halts the scheduler.
func (s *AuditRetentionScheduler) Stop() {
	close(s.stopCh)
}

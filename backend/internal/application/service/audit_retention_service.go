package service

import (
	"context"
	"log"
	"time"
)

// NamedAuditPruner pairs an audit store's name with its prune function
// (satisfied by the org- and api-audit repositories' DeleteOlderThan).
type NamedAuditPruner struct {
	Name  string
	Prune func(ctx context.Context, cutoff time.Time) (int64, error)
}

// AuditRetentionService deletes audit-log rows older than a retention window.
// Disabled (no-op) when retentionDays <= 0 — audit logs are compliance records, so
// pruning is strictly opt-in; nothing is ever auto-deleted by default.
type AuditRetentionService struct {
	retentionDays int
	pruners       []NamedAuditPruner
}

// NewAuditRetentionService builds a retention service for the given stores.
func NewAuditRetentionService(retentionDays int, pruners ...NamedAuditPruner) *AuditRetentionService {
	return &AuditRetentionService{retentionDays: retentionDays, pruners: pruners}
}

// Enabled reports whether retention pruning is active.
func (s *AuditRetentionService) Enabled() bool {
	return s.retentionDays > 0 && len(s.pruners) > 0
}

// PruneOnce deletes audit rows older than retentionDays before now, across all
// stores. No-op when disabled. A failure on one store is logged and does not stop
// the others.
func (s *AuditRetentionService) PruneOnce(ctx context.Context, now time.Time) {
	if s.retentionDays <= 0 {
		return
	}
	cutoff := now.AddDate(0, 0, -s.retentionDays)
	for _, p := range s.pruners {
		n, err := p.Prune(ctx, cutoff)
		if err != nil {
			log.Printf("audit retention: prune %s failed: %v", p.Name, err)
			continue
		}
		if n > 0 {
			log.Printf("audit retention: pruned %d rows from %s (older than %s)", n, p.Name, cutoff.Format("2006-01-02"))
		}
	}
}

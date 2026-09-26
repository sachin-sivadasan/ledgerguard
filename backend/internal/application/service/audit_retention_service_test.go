package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAuditRetention_DisabledIsNoOp(t *testing.T) {
	calls := 0
	p := NamedAuditPruner{Name: "x", Prune: func(context.Context, time.Time) (int64, error) {
		calls++
		return 0, nil
	}}
	svc := NewAuditRetentionService(0, p)

	if svc.Enabled() {
		t.Error("expected retention disabled when retentionDays <= 0")
	}
	svc.PruneOnce(context.Background(), time.Now())
	if calls != 0 {
		t.Errorf("expected no prune calls when disabled, got %d", calls)
	}
}

func TestAuditRetention_PrunesWithCorrectCutoff(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	wantCutoff := now.AddDate(0, 0, -30)

	var gotCutoffs []time.Time
	mk := func(name string) NamedAuditPruner {
		return NamedAuditPruner{Name: name, Prune: func(_ context.Context, cutoff time.Time) (int64, error) {
			gotCutoffs = append(gotCutoffs, cutoff)
			return 3, nil
		}}
	}
	svc := NewAuditRetentionService(30, mk("org_audit_log"), mk("api_audit_log"))

	if !svc.Enabled() {
		t.Fatal("expected retention enabled")
	}
	svc.PruneOnce(context.Background(), now)

	if len(gotCutoffs) != 2 {
		t.Fatalf("expected both stores pruned, got %d", len(gotCutoffs))
	}
	for _, c := range gotCutoffs {
		if !c.Equal(wantCutoff) {
			t.Errorf("cutoff = %v, want %v (now - 30d)", c, wantCutoff)
		}
	}
}

func TestAuditRetention_OneStoreErrorDoesNotStopOthers(t *testing.T) {
	second := 0
	failing := NamedAuditPruner{Name: "bad", Prune: func(context.Context, time.Time) (int64, error) {
		return 0, errors.New("db down")
	}}
	ok := NamedAuditPruner{Name: "good", Prune: func(context.Context, time.Time) (int64, error) {
		second++
		return 1, nil
	}}
	svc := NewAuditRetentionService(7, failing, ok)

	svc.PruneOnce(context.Background(), time.Now())
	if second != 1 {
		t.Errorf("expected the second store to prune despite the first failing, got %d", second)
	}
}

package processors

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// TestJobLogger_AttachesJobFields is the "one sync's story" guard: every log line a
// processor emits must carry processor / app_id / job_id as structured keyword fields
// so a whole sync can be reconstructed from Elasticsearch by app_id or job_id — and the
// scoped logger must flow via ctx so downstream FromContext callers inherit the fields.
func TestJobLogger_AttachesJobFields(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	appID := uuid.New()
	jobID := uuid.New()
	payload := &queue.SyncJobPayload{AppID: appID, JobID: jobID}

	ctx := logging.ContextWithLogger(context.Background(), base)
	ctx, lg := jobLogger(ctx, "TransactionProcessor", payload)
	lg.Info("synced transactions", zap.Int("count", 3))

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["processor"] != "TransactionProcessor" {
		t.Errorf("processor = %v, want TransactionProcessor", fields["processor"])
	}
	if fields["app_id"] != appID.String() {
		t.Errorf("app_id = %v, want %s", fields["app_id"], appID)
	}
	if fields["job_id"] != jobID.String() {
		t.Errorf("job_id = %v, want %s", fields["job_id"], jobID)
	}
	if fields["count"] != int64(3) {
		t.Errorf("count = %v, want 3", fields["count"])
	}

	// request_id is omitted when the payload carries none (background/recovery job).
	if _, ok := fields["request_id"]; ok {
		t.Errorf("request_id should be absent when payload.RequestID is empty, got %v", fields["request_id"])
	}

	// The scoped logger must also flow via ctx to downstream FromContext callers.
	logging.FromContext(ctx).Info("downstream")
	if got := recorded.All()[1].ContextMap()["app_id"]; got != appID.String() {
		t.Errorf("ctx-propagated app_id = %v, want %s", got, appID)
	}
}

// TestJobLogger_IncludesRequestID: when the payload carries the enqueuing request's id
// (threaded through Redis), it must appear on the processor's log lines so a sync can be
// tied back to the HTTP request that triggered it.
func TestJobLogger_IncludesRequestID(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	const reqID = "req-enqueue-42"
	payload := &queue.SyncJobPayload{AppID: uuid.New(), JobID: uuid.New(), RequestID: reqID}

	ctx := logging.ContextWithLogger(context.Background(), base)
	_, lg := jobLogger(ctx, "TransactionProcessor", payload)
	lg.Info("synced")

	if got := recorded.All()[0].ContextMap()["request_id"]; got != reqID {
		t.Errorf("request_id = %v, want %q", got, reqID)
	}
}

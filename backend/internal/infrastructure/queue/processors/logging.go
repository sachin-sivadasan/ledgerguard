package processors

import (
	"context"

	"go.uber.org/zap"

	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/queue"
)

// jobLogger returns a logger scoped to one sync job, plus a context carrying it.
// The processor / app_id / job_id fields are the correlation keys that let us
// reconstruct "one sync's story" in Elasticsearch (docs/LOGS_MCP_SETUP.md, Phase 4) —
// they replace the hand-interpolated "for app %s (job %s)" suffixes the old
// log.Printf lines carried. Attaching the logger to ctx means downstream calls using
// logging.FromContext inherit the same fields for free.
func jobLogger(ctx context.Context, processor string, payload *queue.SyncJobPayload) (context.Context, *zap.Logger) {
	fields := []zap.Field{
		zap.String("processor", processor),
		zap.String("app_id", payload.AppID.String()),
		zap.String("job_id", payload.JobID.String()),
	}
	// request_id ties this sync back to the HTTP request that enqueued it (carried in
	// the payload through Redis). Empty for background/recovery jobs — omit it then.
	if payload.RequestID != "" {
		fields = append(fields, zap.String("request_id", payload.RequestID))
	}
	lg := logging.FromContext(ctx).With(fields...)
	return logging.ContextWithLogger(ctx, lg), lg
}

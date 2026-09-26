// Package logging provides the app's structured logger: zap with the ECS encoder so
// field names (@timestamp, log.level, message, service.*) match the Elasticsearch index
// template used by the logs-over-MCP setup (docs/LOGS_MCP_SETUP.md).
package logging

import (
	"context"
	"os"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

// newLogger builds an ECS-encoded zap logger writing to w. Exposed (unexported) for tests.
func newLogger(w zapcore.WriteSyncer, level, version string) *zap.Logger {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level)) // invalid/empty → stays Info
	core := ecszap.NewCore(ecszap.NewDefaultEncoderConfig(), w, lvl)
	return zap.New(core, zap.AddCaller()).With(
		zap.String("service.name", "ledgerguard-api"),
		zap.String("service.version", version),
	)
}

// New builds the base logger (ECS JSON to stdout).
func New(level, version string) *zap.Logger {
	return newLogger(zapcore.AddSync(os.Stdout), level, version)
}

// Init builds the logger and installs it globally: zap.L() returns it, and — via
// RedirectStdLog — every existing stdlib log.Printf/Println emits structured JSON too
// (so the 376 legacy call sites become structured immediately, migrating to the
// context logger opportunistically). Returns a cleanup func; call once at startup.
func Init(level, version string) (*zap.Logger, func()) {
	l := New(level, version)
	undoGlobals := zap.ReplaceGlobals(l)
	undoStd := zap.RedirectStdLog(l)
	return l, func() {
		undoStd()
		undoGlobals()
		_ = l.Sync()
	}
}

// ContextWithLogger stores a request-scoped logger (e.g. carrying request_id) on ctx.
func ContextWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns the request-scoped logger, or the global logger if none is set.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.L()
}

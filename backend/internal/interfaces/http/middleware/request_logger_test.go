package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/logging"
)

// TestRequestLogger_AttachesRequestID is the correlation-id guard: a downstream handler
// logging via logging.FromContext(ctx) must emit the request's chi request_id, so one
// request's log lines can be tied together.
func TestRequestLogger_AttachesRequestID(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	const reqID = "req-abc-123"
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.FromContext(r.Context()).Info("handled")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := logging.ContextWithLogger(req.Context(), base)   // seed the base logger RequestLogger builds on
	ctx = context.WithValue(ctx, chimw.RequestIDKey, reqID) // seed the id chimw.GetReqID reads
	req = req.WithContext(ctx)

	handler.ServeHTTP(httptest.NewRecorder(), req)

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry from the handler, got %d", len(entries))
	}
	if got := entries[0].ContextMap()["request_id"]; got != reqID {
		t.Errorf("request_id = %v, want %q", got, reqID)
	}
}

// TestRequestLogger_NoRequestID: without chi's RequestID upstream, request_id is empty
// (no panic) and the handler still runs.
func TestRequestLogger_NoRequestID(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	base := zap.New(core)

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.FromContext(r.Context()).Info("handled")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(logging.ContextWithLogger(req.Context(), base))

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if len(recorded.All()) != 1 {
		t.Fatalf("expected the handler to run and log once, got %d entries", len(recorded.All()))
	}
	if got := recorded.All()[0].ContextMap()["request_id"]; got != "" {
		t.Errorf("request_id = %v, want empty string when RequestID middleware absent", got)
	}
}

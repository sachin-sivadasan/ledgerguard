package logging

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// bufSyncer is a WriteSyncer over a strings.Builder for capturing log output.
type bufSyncer struct{ b *strings.Builder }

func (s bufSyncer) Write(p []byte) (int, error) { return s.b.Write(p) }
func (s bufSyncer) Sync() error                 { return nil }

func TestNewLogger_ECSFields(t *testing.T) {
	var b strings.Builder
	l := newLogger(bufSyncer{&b}, "info", "1.2.3")
	l.With(zap.String("request_id", "req-42")).Info("hello")

	var m map[string]any
	if err := json.Unmarshal([]byte(b.String()), &m); err != nil {
		t.Fatalf("log line is not valid JSON: %v\n%s", err, b.String())
	}
	// ECS-encoded field names + our base/context fields.
	for _, k := range []string{"@timestamp", "log.level", "message", "service.name", "service.version", "request_id"} {
		if _, ok := m[k]; !ok {
			t.Errorf("expected ECS field %q in output, got keys: %v", k, keys(m))
		}
	}
	if m["service.name"] != "ledgerguard-api" {
		t.Errorf("service.name = %v, want ledgerguard-api", m["service.name"])
	}
	if m["request_id"] != "req-42" {
		t.Errorf("request_id = %v, want req-42", m["request_id"])
	}
}

func TestFromContext_RoundTripAndFallback(t *testing.T) {
	// No logger on ctx → falls back to the global (never nil).
	if FromContext(context.Background()) == nil {
		t.Fatal("FromContext should fall back to a non-nil global logger")
	}
	// Round-trips a stored logger.
	l := zap.NewNop()
	ctx := ContextWithLogger(context.Background(), l)
	if FromContext(ctx) != l {
		t.Error("FromContext did not return the logger stored by ContextWithLogger")
	}
}

func TestNewLogger_LevelFiltering(t *testing.T) {
	var b strings.Builder
	l := newLogger(bufSyncer{&b}, "error", "dev")
	l.Info("suppressed") // below error
	l.Error("kept")
	out := b.String()
	if strings.Contains(out, "suppressed") {
		t.Error("info line should be filtered at error level")
	}
	if !strings.Contains(out, "kept") {
		t.Error("error line should be emitted at error level")
	}
}

// TestNewLogger_InvalidLevelDefaultsToInfo pins the incident-safety branch: a typo'd,
// non-empty LOG_LEVEL must fall back to Info (not silently downgrade to a level that
// swallows Info/Warn during an outage).
func TestNewLogger_InvalidLevelDefaultsToInfo(t *testing.T) {
	var b strings.Builder
	l := newLogger(bufSyncer{&b}, "debgu", "dev") // unparseable level
	l.Info("kept")
	if !strings.Contains(b.String(), "kept") {
		t.Error("invalid LOG_LEVEL should default to Info (info line must be emitted)")
	}
}

var _ zapcore.WriteSyncer = bufSyncer{}

func keys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

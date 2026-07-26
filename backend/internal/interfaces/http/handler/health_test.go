package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockDBChecker struct {
	healthy bool
}

func (m *mockDBChecker) Ping(ctx context.Context) error {
	if m.healthy {
		return nil
	}
	return context.DeadlineExceeded
}

func TestHealthHandler_Healthy(t *testing.T) {
	handler := NewHealthHandler(&mockDBChecker{healthy: true})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}

	if resp.Database != "connected" {
		t.Errorf("expected database 'connected', got '%s'", resp.Database)
	}
}

func TestHealthHandler_DatabaseUnhealthy(t *testing.T) {
	handler := NewHealthHandler(&mockDBChecker{healthy: false})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", resp.Status)
	}

	if resp.Database != "disconnected" {
		t.Errorf("expected database 'disconnected', got '%s'", resp.Database)
	}
}

// mockSchemaDBChecker implements both DBChecker and SchemaChecker.
type mockSchemaDBChecker struct {
	healthy   bool
	version   int64
	dirty     bool
	schemaErr error
}

func (m *mockSchemaDBChecker) Ping(ctx context.Context) error {
	if m.healthy {
		return nil
	}
	return context.DeadlineExceeded
}

func (m *mockSchemaDBChecker) MigrationStatus(ctx context.Context) (int64, bool, error) {
	return m.version, m.dirty, m.schemaErr
}

func TestHealthHandler_SchemaMigrated(t *testing.T) {
	handler := NewHealthHandler(&mockSchemaDBChecker{healthy: true, version: 40, dirty: false})
	rec := httptest.NewRecorder()
	handler.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "ok" || resp.Schema != "migrated (v40)" {
		t.Errorf("expected ok + 'migrated (v40)', got '%s' / '%s'", resp.Status, resp.Schema)
	}
}

func TestHealthHandler_SchemaDirty(t *testing.T) {
	handler := NewHealthHandler(&mockSchemaDBChecker{healthy: true, version: 17, dirty: true})
	rec := httptest.NewRecorder()
	handler.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for dirty schema, got %d", rec.Code)
	}
	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "degraded" || resp.Schema != "dirty at version 17" {
		t.Errorf("expected degraded + 'dirty at version 17', got '%s' / '%s'", resp.Status, resp.Schema)
	}
}

func TestHealthHandler_SchemaNotApplied(t *testing.T) {
	handler := NewHealthHandler(&mockSchemaDBChecker{healthy: true, schemaErr: context.DeadlineExceeded})
	rec := httptest.NewRecorder()
	handler.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when migrations not applied, got %d", rec.Code)
	}
	var resp HealthResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "degraded" || resp.Schema != "migrations not applied" {
		t.Errorf("expected degraded + 'migrations not applied', got '%s' / '%s'", resp.Status, resp.Schema)
	}
}

func TestHealthHandler_NilDB(t *testing.T) {
	handler := NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}

	if resp.Database != "not configured" {
		t.Errorf("expected database 'not configured', got '%s'", resp.Database)
	}
}

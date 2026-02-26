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

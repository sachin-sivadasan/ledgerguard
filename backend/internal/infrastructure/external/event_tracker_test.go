package external

import (
	"context"
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

func TestNoopTracker_Track(t *testing.T) {
	tracker := NewNoopTracker()
	// Should not panic
	tracker.Track(context.Background(), "user-123", "test_event", service.EventProperties{
		"key": "value",
	})
}

func TestNoopTracker_SetUserProperties(t *testing.T) {
	tracker := NewNoopTracker()
	// Should not panic
	tracker.SetUserProperties(context.Background(), "user-123", service.EventProperties{
		"$email": "test@example.com",
	})
}

func TestNoopTracker_ImplementsInterface(t *testing.T) {
	var _ service.EventTracker = (*NoopTracker)(nil)
}

func TestMixpanelClient_ImplementsInterface(t *testing.T) {
	var _ service.EventTracker = (*MixpanelClient)(nil)
}

func TestMixpanelClient_New(t *testing.T) {
	client := NewMixpanelClient("test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.token != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", client.token)
	}
}

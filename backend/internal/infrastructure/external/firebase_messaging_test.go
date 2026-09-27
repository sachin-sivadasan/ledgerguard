package external

import (
	"testing"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// TestBuildMessage_IncludesData pins the deep-link contract: the FCM message must
// carry the data payload AND keep the notification block (so clients that ignore
// data still get the visible alert).
func TestBuildMessage_IncludesData(t *testing.T) {
	data := map[string]string{"type": "risk_alert", "app_id": "app-123", "subscription_id": "sub-9"}
	msg := buildMessage("tok", entity.PlatformAndroid, "T", "B", data)

	if msg.Data["type"] != "risk_alert" || msg.Data["app_id"] != "app-123" || msg.Data["subscription_id"] != "sub-9" {
		t.Fatalf("expected deep-link data on message, got %#v", msg.Data)
	}
	if msg.Notification == nil || msg.Notification.Title != "T" || msg.Notification.Body != "B" {
		t.Fatal("notification block must remain alongside data")
	}
	if msg.Android == nil {
		t.Error("android platform config should be set for android tokens")
	}
}

// TestBuildMessage_NilData: a nil data map is valid (daily-summary style) and must
// not panic or fabricate keys.
func TestBuildMessage_NilData(t *testing.T) {
	msg := buildMessage("tok", entity.PlatformIOS, "T", "B", nil)
	if msg.Data != nil {
		t.Errorf("expected nil data, got %#v", msg.Data)
	}
	if msg.APNS == nil {
		t.Error("apns config should be set for ios tokens")
	}
}

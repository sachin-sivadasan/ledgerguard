package entity

import (
	"testing"
	"time"
)

// TestSubscription_StartDate verifies StartDate() returns ActivatedAt (the business
// start) when set, and falls back to CreatedAt (the record-created timestamp) otherwise.
func TestSubscription_StartDate(t *testing.T) {
	created := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	activated := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	withActivated := &Subscription{CreatedAt: created, ActivatedAt: &activated}
	if !withActivated.StartDate().Equal(activated) {
		t.Errorf("StartDate with ActivatedAt set: expected %v, got %v", activated, withActivated.StartDate())
	}

	noActivated := &Subscription{CreatedAt: created}
	if !noActivated.StartDate().Equal(created) {
		t.Errorf("StartDate fallback: expected CreatedAt %v, got %v", created, noActivated.StartDate())
	}
}

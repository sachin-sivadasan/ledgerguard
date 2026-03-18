package valueobject

import "testing"

func TestBillingSubscriptionStatus_IsValid(t *testing.T) {
	tests := []struct {
		status BillingSubscriptionStatus
		valid  bool
	}{
		{BillingSubscriptionStatusCreated, true},
		{BillingSubscriptionStatusActive, true},
		{BillingSubscriptionStatusPending, true},
		{BillingSubscriptionStatusHalted, true},
		{BillingSubscriptionStatusCancelled, true},
		{BillingSubscriptionStatusCompleted, true},
		{BillingSubscriptionStatus("INVALID"), false},
		{BillingSubscriptionStatus(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.valid {
				t.Errorf("BillingSubscriptionStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestBillingSubscriptionStatus_IsActive(t *testing.T) {
	if !BillingSubscriptionStatusActive.IsActive() {
		t.Error("ACTIVE.IsActive() should be true")
	}
	if BillingSubscriptionStatusCreated.IsActive() {
		t.Error("CREATED.IsActive() should be false")
	}
}

func TestBillingSubscriptionStatus_IsTerminal(t *testing.T) {
	if !BillingSubscriptionStatusCancelled.IsTerminal() {
		t.Error("CANCELLED.IsTerminal() should be true")
	}
	if !BillingSubscriptionStatusCompleted.IsTerminal() {
		t.Error("COMPLETED.IsTerminal() should be true")
	}
	if BillingSubscriptionStatusActive.IsTerminal() {
		t.Error("ACTIVE.IsTerminal() should be false")
	}
}

func TestParseBillingSubscriptionStatus(t *testing.T) {
	tests := []struct {
		input string
		want  BillingSubscriptionStatus
	}{
		{"CREATED", BillingSubscriptionStatusCreated},
		{"ACTIVE", BillingSubscriptionStatusActive},
		{"PENDING", BillingSubscriptionStatusPending},
		{"HALTED", BillingSubscriptionStatusHalted},
		{"CANCELLED", BillingSubscriptionStatusCancelled},
		{"COMPLETED", BillingSubscriptionStatusCompleted},
		{"UNKNOWN", BillingSubscriptionStatus("UNKNOWN")},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseBillingSubscriptionStatus(tt.input); got != tt.want {
				t.Errorf("ParseBillingSubscriptionStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

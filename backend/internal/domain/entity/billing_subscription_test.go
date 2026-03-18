package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestNewBillingSubscription(t *testing.T) {
	userID := uuid.New()
	bs := NewBillingSubscription(userID, "sub_123", "plan_456", "cust_789", valueobject.BillingPlanStarter, "https://rzp.io/test")

	if bs.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if bs.UserID != userID {
		t.Errorf("UserID = %v, want %v", bs.UserID, userID)
	}
	if bs.RazorpaySubscriptionID != "sub_123" {
		t.Errorf("RazorpaySubscriptionID = %q, want %q", bs.RazorpaySubscriptionID, "sub_123")
	}
	if bs.Plan != valueobject.BillingPlanStarter {
		t.Errorf("Plan = %v, want STARTER", bs.Plan)
	}
	if bs.Status != valueobject.BillingSubscriptionStatusCreated {
		t.Errorf("Status = %v, want CREATED", bs.Status)
	}
	if bs.AmountCents != 24900 {
		t.Errorf("AmountCents = %d, want 24900", bs.AmountCents)
	}
	if bs.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", bs.Currency)
	}
	if bs.ShortURL != "https://rzp.io/test" {
		t.Errorf("ShortURL = %q, want https://rzp.io/test", bs.ShortURL)
	}
}

func TestBillingSubscription_Activate(t *testing.T) {
	bs := NewBillingSubscription(uuid.New(), "sub_1", "plan_1", "cust_1", valueobject.BillingPlanPro, "")
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	bs.Activate(start, end)

	if bs.Status != valueobject.BillingSubscriptionStatusActive {
		t.Errorf("Status = %v, want ACTIVE", bs.Status)
	}
	if bs.CurrentPeriodStart == nil || !bs.CurrentPeriodStart.Equal(start) {
		t.Errorf("CurrentPeriodStart = %v, want %v", bs.CurrentPeriodStart, start)
	}
	if bs.CurrentPeriodEnd == nil || !bs.CurrentPeriodEnd.Equal(end) {
		t.Errorf("CurrentPeriodEnd = %v, want %v", bs.CurrentPeriodEnd, end)
	}
}

func TestBillingSubscription_Cancel(t *testing.T) {
	bs := NewBillingSubscription(uuid.New(), "sub_1", "plan_1", "cust_1", valueobject.BillingPlanStarter, "")
	bs.Cancel()
	if bs.Status != valueobject.BillingSubscriptionStatusCancelled {
		t.Errorf("Status = %v, want CANCELLED", bs.Status)
	}
}

func TestBillingSubscription_Halt(t *testing.T) {
	bs := NewBillingSubscription(uuid.New(), "sub_1", "plan_1", "cust_1", valueobject.BillingPlanStarter, "")
	bs.Halt()
	if bs.Status != valueobject.BillingSubscriptionStatusHalted {
		t.Errorf("Status = %v, want HALTED", bs.Status)
	}
}

func TestBillingSubscription_MarkPending(t *testing.T) {
	bs := NewBillingSubscription(uuid.New(), "sub_1", "plan_1", "cust_1", valueobject.BillingPlanStarter, "")
	bs.MarkPending()
	if bs.Status != valueobject.BillingSubscriptionStatusPending {
		t.Errorf("Status = %v, want PENDING", bs.Status)
	}
}

func TestBillingSubscription_MapToPlanTier(t *testing.T) {
	tests := []struct {
		plan valueobject.BillingPlan
		tier valueobject.PlanTier
	}{
		{valueobject.BillingPlanStarter, valueobject.PlanTierStarter},
		{valueobject.BillingPlanPro, valueobject.PlanTierPro},
		{valueobject.BillingPlan("UNKNOWN"), valueobject.PlanTierFree},
	}
	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			bs := NewBillingSubscription(uuid.New(), "sub_1", "plan_1", "cust_1", tt.plan, "")
			if got := bs.MapToPlanTier(); got != tt.tier {
				t.Errorf("MapToPlanTier() = %v, want %v", got, tt.tier)
			}
		})
	}
}

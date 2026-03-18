package valueobject

import "testing"

func TestBillingPlan_IsValid(t *testing.T) {
	tests := []struct {
		plan  BillingPlan
		valid bool
	}{
		{BillingPlanStarter, true},
		{BillingPlanPro, true},
		{BillingPlan("INVALID"), false},
		{BillingPlan(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.plan.String(), func(t *testing.T) {
			if got := tt.plan.IsValid(); got != tt.valid {
				t.Errorf("BillingPlan(%q).IsValid() = %v, want %v", tt.plan, got, tt.valid)
			}
		})
	}
}

func TestBillingPlan_PriceUSDCents(t *testing.T) {
	tests := []struct {
		plan  BillingPlan
		cents int
	}{
		{BillingPlanStarter, 24900},
		{BillingPlanPro, 49900},
		{BillingPlan("INVALID"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.plan.String(), func(t *testing.T) {
			if got := tt.plan.PriceUSDCents(); got != tt.cents {
				t.Errorf("BillingPlan(%q).PriceUSDCents() = %d, want %d", tt.plan, got, tt.cents)
			}
		})
	}
}

func TestParseBillingPlan(t *testing.T) {
	tests := []struct {
		input string
		want  BillingPlan
	}{
		{"STARTER", BillingPlanStarter},
		{"PRO", BillingPlanPro},
		{"UNKNOWN", BillingPlan("UNKNOWN")},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseBillingPlan(tt.input); got != tt.want {
				t.Errorf("ParseBillingPlan(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

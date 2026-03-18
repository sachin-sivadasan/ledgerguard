package valueobject

import "testing"

func TestPlanTier_IsValid(t *testing.T) {
	tests := []struct {
		tier  PlanTier
		valid bool
	}{
		{PlanTierFree, true},
		{PlanTierStarter, true},
		{PlanTierPro, true},
		{PlanTier("INVALID"), false},
		{PlanTier(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			if got := tt.tier.IsValid(); got != tt.valid {
				t.Errorf("PlanTier(%q).IsValid() = %v, want %v", tt.tier, got, tt.valid)
			}
		})
	}
}

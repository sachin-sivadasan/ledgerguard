package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func TestOrganization_MaxApps(t *testing.T) {
	tests := []struct {
		name     string
		planTier valueobject.PlanTier
		want     int
	}{
		{"FREE tier allows 1 app", valueobject.PlanTierFree, 1},
		{"STARTER tier allows 1 app", valueobject.PlanTierStarter, 1},
		{"PRO tier allows unlimited (0)", valueobject.PlanTierPro, 0},
		{"unknown tier defaults to 1", valueobject.PlanTier("UNKNOWN"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org := &Organization{
				ID:       uuid.New(),
				PlanTier: tt.planTier,
			}
			if got := org.MaxApps(); got != tt.want {
				t.Errorf("MaxApps() = %d, want %d", got, tt.want)
			}
		})
	}
}

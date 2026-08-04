package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// PlanLabelRepository persists developer-assigned plan labels per app.
type PlanLabelRepository interface {
	// FindByAppID returns all plan labels for an app (empty slice when none).
	FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.PlanLabel, error)
	// ReplaceAll atomically replaces the app's entire label set with the provided one
	// (the settings form saves the full set; cleared tiers are simply absent).
	ReplaceAll(ctx context.Context, appID uuid.UUID, labels []*entity.PlanLabel) error
}

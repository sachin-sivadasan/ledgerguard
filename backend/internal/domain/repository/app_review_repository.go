package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type AppReviewRepository interface {
	UpsertBatch(ctx context.Context, reviews []*entity.AppReview) error
	FindByAppID(ctx context.Context, appID uuid.UUID, limit, offset int) ([]*entity.AppReview, error)
	FindAllByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppReview, error)
	CountByAppID(ctx context.Context, appID uuid.UUID) (int, error)
}

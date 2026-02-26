package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type SubscriptionRepository interface {
	Upsert(ctx context.Context, subscription *entity.Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.Subscription, error)
	FindByShopifyGID(ctx context.Context, shopifyGID string) (*entity.Subscription, error)
	FindByAppIDAndDomain(ctx context.Context, appID uuid.UUID, myshopifyDomain string) (*entity.Subscription, error)
	FindByRiskState(ctx context.Context, appID uuid.UUID, riskState valueobject.RiskState) ([]*entity.Subscription, error)
	DeleteByAppID(ctx context.Context, appID uuid.UUID) error
}

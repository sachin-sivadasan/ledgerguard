package repository

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// ShopRepository provides access to shop brand data
type ShopRepository interface {
	Upsert(ctx context.Context, shop *entity.Shop) error
	FindByDomain(ctx context.Context, domain string) (*entity.Shop, error)
	FindByDomains(ctx context.Context, domains []string) (map[string]*entity.Shop, error)
}

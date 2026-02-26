package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type PartnerAccountRepository interface {
	Create(ctx context.Context, account *entity.PartnerAccount) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.PartnerAccount, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.PartnerAccount, error)
	FindByPartnerID(ctx context.Context, partnerID string) (*entity.PartnerAccount, error)
	Update(ctx context.Context, account *entity.PartnerAccount) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

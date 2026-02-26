package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}

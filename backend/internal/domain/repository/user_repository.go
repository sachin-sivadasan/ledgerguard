package repository

import (
	"context"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

type UserRepository interface {
	FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}

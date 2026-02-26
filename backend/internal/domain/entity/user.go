package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

type User struct {
	ID          uuid.UUID
	FirebaseUID string
	Email       string
	Role        valueobject.Role
	PlanTier    valueobject.PlanTier
	CreatedAt   time.Time
}

func NewUser(firebaseUID, email string) *User {
	return &User{
		ID:          uuid.New(),
		FirebaseUID: firebaseUID,
		Email:       email,
		Role:        valueobject.RoleOwner,
		PlanTier:    valueobject.PlanTierFree,
		CreatedAt:   time.Now().UTC(),
	}
}

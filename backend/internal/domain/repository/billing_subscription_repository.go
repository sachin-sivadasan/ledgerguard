package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
)

// BillingSubscriptionRepository defines persistence operations for billing subscriptions.
type BillingSubscriptionRepository interface {
	Create(ctx context.Context, bs *entity.BillingSubscription) error
	Update(ctx context.Context, bs *entity.BillingSubscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.BillingSubscription, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.BillingSubscription, error)
	FindByRazorpaySubscriptionID(ctx context.Context, razorpaySubID string) (*entity.BillingSubscription, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.BillingSubscription, error)
}

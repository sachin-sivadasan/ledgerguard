package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/entity"
	revrepo "github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/domain/repository"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrAppAccessDenied      = errors.New("access denied to this app")
)

// SubscriptionStatusService handles subscription status queries
type SubscriptionStatusService struct {
	statusRepo  revrepo.SubscriptionStatusRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewSubscriptionStatusService creates a new SubscriptionStatusService
func NewSubscriptionStatusService(
	statusRepo revrepo.SubscriptionStatusRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *SubscriptionStatusService {
	return &SubscriptionStatusService{
		statusRepo:  statusRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// GetByShopifyGID retrieves a subscription status by Shopify GID
// Returns ErrAppAccessDenied if the user doesn't own the app
func (s *SubscriptionStatusService) GetByShopifyGID(ctx context.Context, userID uuid.UUID, shopifyGID string) (*entity.SubscriptionStatus, error) {
	// Get the subscription status
	status, err := s.statusRepo.GetByShopifyGID(ctx, shopifyGID)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	// Verify user has access to this app
	if err := s.verifyAppAccess(ctx, userID, status.AppID); err != nil {
		return nil, err
	}

	return status, nil
}

// GetByDomain retrieves a subscription status by myshopify domain
func (s *SubscriptionStatusService) GetByDomain(ctx context.Context, userID uuid.UUID, domain string) (*entity.SubscriptionStatus, error) {
	// First, we need to find which apps this user has access to
	apps, err := s.getUserApps(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Try to find the subscription in any of the user's apps
	for _, appID := range apps {
		status, err := s.statusRepo.GetByDomain(ctx, appID, domain)
		if err == nil {
			return status, nil
		}
	}

	return nil, ErrSubscriptionNotFound
}

// GetByShopifyGIDs retrieves multiple subscription statuses by Shopify GIDs
func (s *SubscriptionStatusService) GetByShopifyGIDs(ctx context.Context, userID uuid.UUID, shopifyGIDs []string) (*entity.SubscriptionStatusBatchResponse, error) {
	if len(shopifyGIDs) == 0 {
		return &entity.SubscriptionStatusBatchResponse{
			Results:  []entity.SubscriptionStatusResponse{},
			NotFound: []string{},
		}, nil
	}

	// Get all statuses
	statuses, err := s.statusRepo.GetByShopifyGIDs(ctx, shopifyGIDs)
	if err != nil {
		return nil, err
	}

	// Get user's apps for access check
	userApps, err := s.getUserApps(ctx, userID)
	if err != nil {
		return nil, err
	}
	userAppSet := make(map[uuid.UUID]bool)
	for _, appID := range userApps {
		userAppSet[appID] = true
	}

	// Build response
	results := []entity.SubscriptionStatusResponse{}
	foundGIDs := make(map[string]bool)

	for _, status := range statuses {
		// Only include if user has access to this app
		if userAppSet[status.AppID] {
			results = append(results, status.ToResponse())
			foundGIDs[status.ShopifyGID] = true
		}
	}

	// Find not found GIDs
	notFound := []string{}
	for _, gid := range shopifyGIDs {
		if !foundGIDs[gid] {
			notFound = append(notFound, gid)
		}
	}

	return &entity.SubscriptionStatusBatchResponse{
		Results:  results,
		NotFound: notFound,
	}, nil
}

// getUserApps returns all app IDs the user has access to
func (s *SubscriptionStatusService) getUserApps(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Get partner account for user (currently single partner account per user)
	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all apps for the partner account
	apps, err := s.appRepo.FindByPartnerAccountID(ctx, partnerAccount.ID)
	if err != nil {
		return nil, err
	}

	appIDs := make([]uuid.UUID, len(apps))
	for i, app := range apps {
		appIDs[i] = app.ID
	}

	return appIDs, nil
}

// verifyAppAccess checks if a user has access to an app
func (s *SubscriptionStatusService) verifyAppAccess(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	// Get the app
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return ErrAppAccessDenied
	}

	// Get the partner account for user
	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrAppAccessDenied
	}

	// Check if the app belongs to the user's partner account
	if app.PartnerAccountID != partnerAccount.ID {
		return ErrAppAccessDenied
	}

	return nil
}

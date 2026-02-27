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
	ErrUsageNotFound = errors.New("usage record not found")
)

// UsageStatusService handles usage status queries
type UsageStatusService struct {
	usageRepo        revrepo.UsageStatusRepository
	subscriptionRepo revrepo.SubscriptionStatusRepository
	appRepo          repository.AppRepository
	partnerRepo      repository.PartnerAccountRepository
}

// NewUsageStatusService creates a new UsageStatusService
func NewUsageStatusService(
	usageRepo revrepo.UsageStatusRepository,
	subscriptionRepo revrepo.SubscriptionStatusRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *UsageStatusService {
	return &UsageStatusService{
		usageRepo:        usageRepo,
		subscriptionRepo: subscriptionRepo,
		appRepo:          appRepo,
		partnerRepo:      partnerRepo,
	}
}

// GetByShopifyGID retrieves a usage status by Shopify GID with parent subscription
func (s *UsageStatusService) GetByShopifyGID(ctx context.Context, userID uuid.UUID, shopifyGID string) (*entity.UsageStatusResponse, error) {
	// Get the usage status
	usage, err := s.usageRepo.GetByShopifyGID(ctx, shopifyGID)
	if err != nil {
		return nil, ErrUsageNotFound
	}

	// Get the parent subscription to verify access and include in response
	subscription, err := s.subscriptionRepo.GetByShopifyGID(ctx, usage.SubscriptionShopifyGID)
	if err != nil {
		return nil, ErrUsageNotFound
	}

	// Verify user has access to this app
	if err := s.verifyAppAccess(ctx, userID, subscription.AppID); err != nil {
		return nil, err
	}

	// Build response with subscription
	resp := usage.ToResponseWithSubscription(subscription)
	return &resp, nil
}

// GetByShopifyGIDs retrieves multiple usage statuses by Shopify GIDs
func (s *UsageStatusService) GetByShopifyGIDs(ctx context.Context, userID uuid.UUID, shopifyGIDs []string) (*entity.UsageStatusBatchResponse, error) {
	if len(shopifyGIDs) == 0 {
		return &entity.UsageStatusBatchResponse{
			Results:  []entity.UsageStatusResponse{},
			NotFound: []string{},
		}, nil
	}

	// Get all usage statuses
	usages, err := s.usageRepo.GetByShopifyGIDs(ctx, shopifyGIDs)
	if err != nil {
		return nil, err
	}

	// Collect unique subscription GIDs
	subGIDs := make(map[string]bool)
	for _, u := range usages {
		subGIDs[u.SubscriptionShopifyGID] = true
	}

	// Fetch all subscriptions
	subGIDList := make([]string, 0, len(subGIDs))
	for gid := range subGIDs {
		subGIDList = append(subGIDList, gid)
	}

	subscriptions, err := s.subscriptionRepo.GetByShopifyGIDs(ctx, subGIDList)
	if err != nil {
		return nil, err
	}

	// Build subscription map
	subMap := make(map[string]*entity.SubscriptionStatus)
	for _, sub := range subscriptions {
		subMap[sub.ShopifyGID] = sub
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
	results := []entity.UsageStatusResponse{}
	foundGIDs := make(map[string]bool)

	for _, usage := range usages {
		sub := subMap[usage.SubscriptionShopifyGID]
		// Only include if user has access to this app
		if sub != nil && userAppSet[sub.AppID] {
			results = append(results, usage.ToResponseWithSubscription(sub))
			foundGIDs[usage.ShopifyGID] = true
		}
	}

	// Find not found GIDs
	notFound := []string{}
	for _, gid := range shopifyGIDs {
		if !foundGIDs[gid] {
			notFound = append(notFound, gid)
		}
	}

	return &entity.UsageStatusBatchResponse{
		Results:  results,
		NotFound: notFound,
	}, nil
}

// getUserApps returns all app IDs the user has access to
func (s *UsageStatusService) getUserApps(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

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
func (s *UsageStatusService) verifyAppAccess(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return ErrAppAccessDenied
	}

	partnerAccount, err := s.partnerRepo.FindByUserID(ctx, userID)
	if err != nil {
		return ErrAppAccessDenied
	}

	if app.PartnerAccountID != partnerAccount.ID {
		return ErrAppAccessDenied
	}

	return nil
}

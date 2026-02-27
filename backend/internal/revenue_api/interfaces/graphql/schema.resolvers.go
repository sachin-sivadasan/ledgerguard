package graphql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/revenue_api/interfaces/http/middleware"
)

// SubscriptionStatus represents a subscription in GraphQL
type SubscriptionStatus struct {
	SubscriptionID           string     `json:"subscriptionId"`
	MyshopifyDomain          string     `json:"myshopifyDomain"`
	ShopName                 *string    `json:"shopName"`
	PlanName                 *string    `json:"planName"`
	RiskState                RiskState  `json:"riskState"`
	IsPaidCurrentCycle       bool       `json:"isPaidCurrentCycle"`
	MonthsOverdue            int        `json:"monthsOverdue"`
	LastSuccessfulChargeDate *time.Time `json:"lastSuccessfulChargeDate"`
	ExpectedNextChargeDate   *time.Time `json:"expectedNextChargeDate"`
	Status                   SubscriptionStatusEnum `json:"status"`
}

// UsageStatus represents a usage record in GraphQL
type UsageStatus struct {
	UsageID      string              `json:"usageId"`
	Billed       bool                `json:"billed"`
	BillingDate  *time.Time          `json:"billingDate"`
	AmountCents  int                 `json:"amountCents"`
	Description  *string             `json:"description"`
	Subscription *SubscriptionStatus `json:"subscription"`
}

// SubscriptionBatchResult is the result of a batch subscription lookup
type SubscriptionBatchResult struct {
	Results  []*SubscriptionStatus `json:"results"`
	NotFound []string              `json:"notFound"`
}

// UsageBatchResult is the result of a batch usage lookup
type UsageBatchResult struct {
	Results  []*UsageStatus `json:"results"`
	NotFound []string       `json:"notFound"`
}

// QueryResolver implements the Query resolvers
type QueryResolver struct {
	*Resolver
}

// Query returns the query resolver
func (r *Resolver) Query() *QueryResolver {
	return &QueryResolver{r}
}

// getUserID extracts the user ID from context (set by API key auth middleware)
func getUserID(ctx context.Context) (uuid.UUID, error) {
	apiKey := middleware.APIKeyFromContext(ctx)
	if apiKey == nil {
		return uuid.Nil, ErrUnauthorized
	}
	return apiKey.UserID, nil
}

// Subscription resolves a single subscription by Shopify GID
func (r *QueryResolver) Subscription(ctx context.Context, shopifyGid string) (*SubscriptionStatus, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	status, err := r.subscriptionService.GetByShopifyGID(ctx, userID, shopifyGid)
	if err != nil {
		return nil, err
	}

	return toGraphQLSubscription(status), nil
}

// SubscriptionByDomain resolves a subscription by myshopify domain
func (r *QueryResolver) SubscriptionByDomain(ctx context.Context, domain string) (*SubscriptionStatus, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	status, err := r.subscriptionService.GetByDomain(ctx, userID, domain)
	if err != nil {
		return nil, err
	}

	return toGraphQLSubscription(status), nil
}

// Subscriptions resolves multiple subscriptions by Shopify GIDs
func (r *QueryResolver) Subscriptions(ctx context.Context, shopifyGids []string) (*SubscriptionBatchResult, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := r.subscriptionService.GetByShopifyGIDs(ctx, userID, shopifyGids)
	if err != nil {
		return nil, err
	}

	gqlResults := make([]*SubscriptionStatus, len(result.Results))
	for i, s := range result.Results {
		gqlResults[i] = &SubscriptionStatus{
			SubscriptionID:           s.SubscriptionID,
			MyshopifyDomain:          s.MyshopifyDomain,
			ShopName:                 strPtr(s.ShopName),
			PlanName:                 strPtr(s.PlanName),
			RiskState:                RiskState(s.RiskState),
			IsPaidCurrentCycle:       s.IsPaidCurrentCycle,
			MonthsOverdue:            s.MonthsOverdue,
			LastSuccessfulChargeDate: s.LastSuccessfulChargeDate,
			ExpectedNextChargeDate:   s.ExpectedNextChargeDate,
			Status:                   SubscriptionStatusEnum(s.Status),
		}
	}

	return &SubscriptionBatchResult{
		Results:  gqlResults,
		NotFound: result.NotFound,
	}, nil
}

// Usage resolves a single usage record by Shopify GID
func (r *QueryResolver) Usage(ctx context.Context, shopifyGid string) (*UsageStatus, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	status, err := r.usageService.GetByShopifyGID(ctx, userID, shopifyGid)
	if err != nil {
		return nil, err
	}

	return toGraphQLUsage(status), nil
}

// Usages resolves multiple usage records by Shopify GIDs
func (r *QueryResolver) Usages(ctx context.Context, shopifyGids []string) (*UsageBatchResult, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := r.usageService.GetByShopifyGIDs(ctx, userID, shopifyGids)
	if err != nil {
		return nil, err
	}

	gqlResults := make([]*UsageStatus, len(result.Results))
	for i, u := range result.Results {
		gqlResults[i] = toGraphQLUsageFromResponse(&u)
	}

	return &UsageBatchResult{
		Results:  gqlResults,
		NotFound: result.NotFound,
	}, nil
}

// Helper functions

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toGraphQLSubscription(s interface{}) *SubscriptionStatus {
	// Type assert to the entity type
	type statusEntity interface {
		ToResponse() interface{}
	}

	if entity, ok := s.(statusEntity); ok {
		resp := entity.ToResponse()
		if r, ok := resp.(struct {
			SubscriptionID           string
			MyshopifyDomain          string
			ShopName                 string
			PlanName                 string
			RiskState                string
			IsPaidCurrentCycle       bool
			MonthsOverdue            int
			LastSuccessfulChargeDate *time.Time
			ExpectedNextChargeDate   *time.Time
			Status                   string
		}); ok {
			return &SubscriptionStatus{
				SubscriptionID:           r.SubscriptionID,
				MyshopifyDomain:          r.MyshopifyDomain,
				ShopName:                 strPtr(r.ShopName),
				PlanName:                 strPtr(r.PlanName),
				RiskState:                RiskState(r.RiskState),
				IsPaidCurrentCycle:       r.IsPaidCurrentCycle,
				MonthsOverdue:            r.MonthsOverdue,
				LastSuccessfulChargeDate: r.LastSuccessfulChargeDate,
				ExpectedNextChargeDate:   r.ExpectedNextChargeDate,
				Status:                   SubscriptionStatusEnum(r.Status),
			}
		}
	}
	return nil
}

func toGraphQLUsage(u interface{}) *UsageStatus {
	// Direct conversion from response pointer
	if resp, ok := u.(*struct {
		UsageID      string
		Billed       bool
		BillingDate  *time.Time
		AmountCents  int
		Description  string
		Subscription *struct {
			SubscriptionID     string
			MyshopifyDomain    string
			RiskState          string
			IsPaidCurrentCycle bool
		}
	}); ok {
		result := &UsageStatus{
			UsageID:     resp.UsageID,
			Billed:      resp.Billed,
			BillingDate: resp.BillingDate,
			AmountCents: resp.AmountCents,
			Description: strPtr(resp.Description),
		}
		if resp.Subscription != nil {
			result.Subscription = &SubscriptionStatus{
				SubscriptionID:     resp.Subscription.SubscriptionID,
				MyshopifyDomain:    resp.Subscription.MyshopifyDomain,
				RiskState:          RiskState(resp.Subscription.RiskState),
				IsPaidCurrentCycle: resp.Subscription.IsPaidCurrentCycle,
			}
		}
		return result
	}
	return nil
}

func toGraphQLUsageFromResponse(u interface{}) *UsageStatus {
	// Handle the entity.UsageStatusResponse type
	return toGraphQLUsage(u)
}

// ErrUnauthorized is returned when no API key is in context
var ErrUnauthorized = &GraphQLError{Message: "unauthorized: API key required", Code: "UNAUTHORIZED"}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string
	Code    string
}

func (e *GraphQLError) Error() string {
	return e.Message
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/external"
)

// BillingService handles B2B subscription billing via Razorpay.
type BillingService struct {
	razorpayClient  *external.RazorpayClient
	billingRepo     repository.BillingSubscriptionRepository
	userRepo        repository.UserRepository
	webhookSecret   string
	starterPlanID   string
	proPlanID       string
	tracker         domainservice.EventTracker
}

// SetTracker sets the event tracker for billing lifecycle events.
func (s *BillingService) SetTracker(t domainservice.EventTracker) {
	s.tracker = t
}

// NewBillingService creates a new BillingService.
func NewBillingService(
	razorpayClient *external.RazorpayClient,
	billingRepo repository.BillingSubscriptionRepository,
	userRepo repository.UserRepository,
	webhookSecret string,
	starterPlanID string,
	proPlanID string,
) *BillingService {
	return &BillingService{
		razorpayClient: razorpayClient,
		billingRepo:    billingRepo,
		userRepo:       userRepo,
		webhookSecret:  webhookSecret,
		starterPlanID:  starterPlanID,
		proPlanID:      proPlanID,
	}
}

// CheckoutResult contains the result of creating a checkout session.
type CheckoutResult struct {
	SubscriptionID string `json:"subscription_id"`
	ShortURL       string `json:"short_url"`
}

// BillingStatus contains the current billing status for a user.
type BillingStatus struct {
	Plan               string     `json:"plan"`
	Status             string     `json:"status"`
	AmountCents        int        `json:"amount_cents"`
	Currency           string     `json:"currency"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	ShortURL           string     `json:"short_url,omitempty"`
}

// CreateCheckout creates a Razorpay subscription and returns the hosted checkout URL.
func (s *BillingService) CreateCheckout(ctx context.Context, userID uuid.UUID, plan valueobject.BillingPlan) (*CheckoutResult, error) {
	if !plan.IsValid() {
		return nil, fmt.Errorf("invalid billing plan: %s", plan)
	}

	// Look up user for email
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Reuse existing Razorpay customer ID if available, otherwise create one
	var customerID string
	existingSubs, _ := s.billingRepo.FindByUserID(ctx, userID)
	if len(existingSubs) > 0 {
		customerID = existingSubs[0].RazorpayCustomerID
	}
	if customerID == "" {
		cust, err := s.razorpayClient.CreateCustomer(ctx, external.CreateCustomerRequest{
			Name:  user.Email,
			Email: user.Email,
		})
		if err != nil {
			return nil, fmt.Errorf("create razorpay customer: %w", err)
		}
		customerID = cust.ID
	}

	// Map plan to Razorpay plan ID
	razorpayPlanID := s.planIDForPlan(plan)
	if razorpayPlanID == "" {
		return nil, fmt.Errorf("no razorpay plan ID configured for plan: %s", plan)
	}

	// Create Razorpay subscription
	sub, err := s.razorpayClient.CreateSubscription(ctx, external.CreateSubscriptionRequest{
		PlanID:         razorpayPlanID,
		CustomerID:     customerID,
		TotalCount:     120, // 10 years of monthly billing
		CustomerNotify: 1,   // required for hosted checkout page
	})
	if err != nil {
		return nil, fmt.Errorf("create razorpay subscription: %w", err)
	}

	// Persist billing subscription
	bs := entity.NewBillingSubscription(
		userID,
		sub.ID,
		razorpayPlanID,
		customerID,
		plan,
		sub.ShortURL,
	)
	if err := s.billingRepo.Create(ctx, bs); err != nil {
		return nil, fmt.Errorf("save billing subscription: %w", err)
	}

	if s.tracker != nil {
		s.tracker.Track(ctx, userID.String(), "billing_subscription_created", domainservice.EventProperties{
			"plan":   string(plan),
			"amount": bs.AmountCents,
		})
	}

	return &CheckoutResult{
		SubscriptionID: sub.ID,
		ShortURL:       sub.ShortURL,
	}, nil
}

// GetBillingStatus returns the current billing status for a user.
func (s *BillingService) GetBillingStatus(ctx context.Context, userID uuid.UUID) (*BillingStatus, error) {
	bs, err := s.billingRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		// No active subscription — return free tier status
		return &BillingStatus{
			Plan:   "FREE",
			Status: "NONE",
		}, nil
	}

	return &BillingStatus{
		Plan:               bs.Plan.String(),
		Status:             bs.Status.String(),
		AmountCents:        bs.AmountCents,
		Currency:           bs.Currency,
		CurrentPeriodStart: bs.CurrentPeriodStart,
		CurrentPeriodEnd:   bs.CurrentPeriodEnd,
		ShortURL:           bs.ShortURL,
	}, nil
}

// RazorpayWebhookPayload represents the top-level Razorpay webhook event.
type RazorpayWebhookPayload struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

// RazorpaySubscriptionPayload is the subscription entity nested in the webhook payload.
type RazorpaySubscriptionPayload struct {
	Subscription struct {
		Entity external.RazorpaySubscription `json:"entity"`
	} `json:"subscription"`
}

// HandleWebhookEvent verifies the webhook signature and routes to the appropriate handler.
// Always returns nil (logs errors) to prevent Razorpay retries on processing failures.
func (s *BillingService) HandleWebhookEvent(ctx context.Context, body []byte, signature string) error {
	if !external.VerifyWebhookSignature(body, signature, s.webhookSecret) {
		return fmt.Errorf("invalid webhook signature")
	}

	var event RazorpayWebhookPayload
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("billing webhook: failed to parse event: %v", err)
		return nil
	}

	var subPayload RazorpaySubscriptionPayload
	if err := json.Unmarshal(event.Payload, &subPayload); err != nil {
		log.Printf("billing webhook: failed to parse subscription payload for event %s: %v", event.Event, err)
		return nil
	}

	razorpaySub := subPayload.Subscription.Entity
	log.Printf("billing webhook: event=%s subscription_id=%s", event.Event, razorpaySub.ID)

	bs, err := s.billingRepo.FindByRazorpaySubscriptionID(ctx, razorpaySub.ID)
	if err != nil {
		log.Printf("billing webhook: subscription not found for razorpay_id=%s: %v", razorpaySub.ID, err)
		return nil
	}

	switch event.Event {
	case "subscription.activated":
		s.handleActivated(ctx, bs, razorpaySub)
	case "subscription.charged":
		s.handleCharged(ctx, bs, razorpaySub)
	case "subscription.pending":
		s.handlePending(ctx, bs)
	case "subscription.halted":
		s.handleHalted(ctx, bs)
	case "subscription.cancelled":
		s.handleCancelled(ctx, bs)
	default:
		log.Printf("billing webhook: unhandled event type: %s", event.Event)
	}

	return nil
}

func (s *BillingService) handleActivated(ctx context.Context, bs *entity.BillingSubscription, rzpSub external.RazorpaySubscription) {
	periodStart, periodEnd := extractPeriod(rzpSub)
	bs.Activate(periodStart, periodEnd)

	if err := s.billingRepo.Update(ctx, bs); err != nil {
		log.Printf("billing webhook: failed to update subscription %s on activation: %v", bs.ID, err)
		return
	}

	if s.tracker != nil {
		s.tracker.Track(ctx, bs.UserID.String(), "billing_activated", domainservice.EventProperties{
			"plan": string(bs.Plan),
		})
	}

	// Update user's plan tier
	s.updateUserPlanTier(ctx, bs.UserID, bs.MapToPlanTier())
}

func (s *BillingService) handleCharged(ctx context.Context, bs *entity.BillingSubscription, rzpSub external.RazorpaySubscription) {
	periodStart, periodEnd := extractPeriod(rzpSub)
	bs.UpdatePeriod(periodStart, periodEnd)

	if err := s.billingRepo.Update(ctx, bs); err != nil {
		log.Printf("billing webhook: failed to update subscription %s on charge: %v", bs.ID, err)
	}
}

func (s *BillingService) handlePending(ctx context.Context, bs *entity.BillingSubscription) {
	bs.MarkPending()
	if err := s.billingRepo.Update(ctx, bs); err != nil {
		log.Printf("billing webhook: failed to update subscription %s to pending: %v", bs.ID, err)
	}
}

func (s *BillingService) handleHalted(ctx context.Context, bs *entity.BillingSubscription) {
	bs.Halt()
	if err := s.billingRepo.Update(ctx, bs); err != nil {
		log.Printf("billing webhook: failed to update subscription %s to halted: %v", bs.ID, err)
	}

	if s.tracker != nil {
		s.tracker.Track(ctx, bs.UserID.String(), "billing_payment_failed", domainservice.EventProperties{
			"plan": string(bs.Plan),
		})
	}
}

func (s *BillingService) handleCancelled(ctx context.Context, bs *entity.BillingSubscription) {
	bs.Cancel()
	if err := s.billingRepo.Update(ctx, bs); err != nil {
		log.Printf("billing webhook: failed to update subscription %s to cancelled: %v", bs.ID, err)
		return
	}

	if s.tracker != nil {
		s.tracker.Track(ctx, bs.UserID.String(), "billing_cancelled", domainservice.EventProperties{
			"plan": string(bs.Plan),
		})
	}

	// Downgrade user to free tier
	s.updateUserPlanTier(ctx, bs.UserID, valueobject.PlanTierFree)
}

func (s *BillingService) updateUserPlanTier(ctx context.Context, userID uuid.UUID, tier valueobject.PlanTier) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Printf("billing webhook: failed to find user %s for plan update: %v", userID, err)
		return
	}
	user.PlanTier = tier
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("billing webhook: failed to update user %s plan tier to %s: %v", userID, tier, err)
	}
}

func (s *BillingService) planIDForPlan(plan valueobject.BillingPlan) string {
	switch plan {
	case valueobject.BillingPlanStarter:
		return s.starterPlanID
	case valueobject.BillingPlanPro:
		return s.proPlanID
	default:
		return ""
	}
}

func extractPeriod(sub external.RazorpaySubscription) (time.Time, time.Time) {
	var start, end time.Time
	if sub.CurrentStart != nil {
		start = time.Unix(*sub.CurrentStart, 0).UTC()
	}
	if sub.CurrentEnd != nil {
		end = time.Unix(*sub.CurrentEnd, 0).UTC()
	}
	return start, end
}

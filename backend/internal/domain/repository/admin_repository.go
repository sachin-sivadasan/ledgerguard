package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AdminUserRow represents a user with aggregated admin-level info.
type AdminUserRow struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	Role                  string     `json:"role"`
	PlanTier              string     `json:"plan_tier"`
	CreatedAt             time.Time  `json:"created_at"`
	OnboardingCompletedAt *time.Time `json:"onboarding_completed_at,omitempty"`
	AppCount              int        `json:"app_count"`
	PartnerConnected      bool       `json:"partner_connected"`
}

// OnboardingFunnel represents aggregate onboarding funnel metrics.
type OnboardingFunnel struct {
	TotalUsers         int `json:"total_users"`
	PartnerConnected   int `json:"partner_connected"`
	AppSelected        int `json:"app_selected"`
	OnboardingComplete int `json:"onboarding_complete"`
}

// AdminSyncJobRow represents a sync job with related user/app info for admin view.
type AdminSyncJobRow struct {
	ID             uuid.UUID  `json:"id"`
	AppID          uuid.UUID  `json:"app_id"`
	AppName        string     `json:"app_name"`
	UserEmail      string     `json:"user_email"`
	JobType        string     `json:"job_type"`
	Status         string     `json:"status"`
	TotalItems     int        `json:"total_items"`
	CompletedItems int        `json:"completed_items"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AdminBillingRow represents a billing subscription with user info for admin view.
type AdminBillingRow struct {
	ID                 uuid.UUID  `json:"id"`
	UserEmail          string     `json:"user_email"`
	Plan               string     `json:"plan"`
	Status             string     `json:"status"`
	AmountCents        int        `json:"amount_cents"`
	Currency           string     `json:"currency"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// AdminRepository provides cross-tenant read-only queries for admin dashboards.
type AdminRepository interface {
	ListUsers(ctx context.Context) ([]AdminUserRow, error)
	GetOnboardingFunnel(ctx context.Context) (*OnboardingFunnel, error)
	ListSyncJobs(ctx context.Context, limit int) ([]AdminSyncJobRow, error)
	ListBillingSubscriptions(ctx context.Context) ([]AdminBillingRow, error)
}

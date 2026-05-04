package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

// mockAdminRepository implements repository.AdminRepository for testing.
type mockAdminRepository struct {
	users   []repository.AdminUserRow
	funnel  *repository.OnboardingFunnel
	jobs    []repository.AdminSyncJobRow
	billing []repository.AdminBillingRow
	err     error
}

func (m *mockAdminRepository) ListUsers(_ context.Context) ([]repository.AdminUserRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.users, nil
}

func (m *mockAdminRepository) GetOnboardingFunnel(_ context.Context) (*repository.OnboardingFunnel, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.funnel, nil
}

func (m *mockAdminRepository) ListSyncJobs(_ context.Context, _ int) ([]repository.AdminSyncJobRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.jobs, nil
}

func (m *mockAdminRepository) ListBillingSubscriptions(_ context.Context) ([]repository.AdminBillingRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.billing, nil
}

func TestAdminHandler_ListUsers(t *testing.T) {
	now := time.Now().UTC()
	repo := &mockAdminRepository{
		users: []repository.AdminUserRow{
			{
				ID:               uuid.New(),
				Email:            "alice@example.com",
				Role:             "OWNER",
				PlanTier:         "FREE",
				CreatedAt:        now,
				AppCount:         2,
				PartnerConnected: true,
			},
			{
				ID:               uuid.New(),
				Email:            "bob@example.com",
				Role:             "OWNER",
				PlanTier:         "STARTER",
				CreatedAt:        now,
				AppCount:         0,
				PartnerConnected: false,
			},
		},
	}

	h := NewAdminHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok || int(total) != 2 {
		t.Errorf("expected total=2, got %v", resp["total"])
	}

	users, ok := resp["users"].([]interface{})
	if !ok || len(users) != 2 {
		t.Errorf("expected 2 users, got %v", len(users))
	}
}

func TestAdminHandler_ListUsers_Error(t *testing.T) {
	repo := &mockAdminRepository{err: errors.New("db error")}
	h := NewAdminHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAdminHandler_OnboardingFunnel(t *testing.T) {
	repo := &mockAdminRepository{
		funnel: &repository.OnboardingFunnel{
			TotalUsers:         100,
			PartnerConnected:   75,
			AppSelected:        50,
			OnboardingComplete: 30,
		},
	}

	h := NewAdminHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/onboarding", nil)
	rec := httptest.NewRecorder()

	h.OnboardingFunnel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp repository.OnboardingFunnel
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp.TotalUsers != 100 {
		t.Errorf("expected TotalUsers=100, got %d", resp.TotalUsers)
	}
	if resp.PartnerConnected != 75 {
		t.Errorf("expected PartnerConnected=75, got %d", resp.PartnerConnected)
	}
	if resp.AppSelected != 50 {
		t.Errorf("expected AppSelected=50, got %d", resp.AppSelected)
	}
	if resp.OnboardingComplete != 30 {
		t.Errorf("expected OnboardingComplete=30, got %d", resp.OnboardingComplete)
	}
}

func TestAdminHandler_OnboardingFunnel_Error(t *testing.T) {
	repo := &mockAdminRepository{err: errors.New("db error")}
	h := NewAdminHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/onboarding", nil)
	rec := httptest.NewRecorder()

	h.OnboardingFunnel(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAdminHandler_ListSyncJobs(t *testing.T) {
	now := time.Now().UTC()
	repo := &mockAdminRepository{
		jobs: []repository.AdminSyncJobRow{
			{
				ID:        uuid.New(),
				AppID:     uuid.New(),
				AppName:   "TestApp",
				UserEmail: "alice@example.com",
				JobType:   "full_sync",
				Status:    "completed",
				CreatedAt: now,
			},
		},
	}

	h := NewAdminHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sync?limit=10", nil)
	rec := httptest.NewRecorder()

	h.ListSyncJobs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok || int(total) != 1 {
		t.Errorf("expected total=1, got %v", resp["total"])
	}
}

func TestAdminHandler_ListSyncJobs_Error(t *testing.T) {
	repo := &mockAdminRepository{err: errors.New("db error")}
	h := NewAdminHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/sync", nil)
	rec := httptest.NewRecorder()

	h.ListSyncJobs(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAdminHandler_ListBilling(t *testing.T) {
	now := time.Now().UTC()
	repo := &mockAdminRepository{
		billing: []repository.AdminBillingRow{
			{
				ID:          uuid.New(),
				UserEmail:   "alice@example.com",
				Plan:        "STARTER",
				Status:      "ACTIVE",
				AmountCents: 999,
				Currency:    "INR",
				CreatedAt:   now,
			},
		},
	}

	h := NewAdminHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/billing", nil)
	rec := httptest.NewRecorder()

	h.ListBilling(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok || int(total) != 1 {
		t.Errorf("expected total=1, got %v", resp["total"])
	}
}

func TestAdminHandler_ListBilling_Error(t *testing.T) {
	repo := &mockAdminRepository{err: errors.New("db error")}
	h := NewAdminHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/billing", nil)
	rec := httptest.NewRecorder()

	h.ListBilling(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

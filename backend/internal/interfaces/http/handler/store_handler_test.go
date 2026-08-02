package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// TestStoreHandler_List_UsesEventSourcedDates is the STORE-2 regression guard:
// first_install_date / last_interaction must come from the app_events stream, not
// the record-created CreatedAt/UpdatedAt (which are reset on every ledger rebuild).
func TestStoreHandler_List_UsesEventSourcedDates(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App", PartnerAppID: "gid://partners/App/1"}

	rebuildTime := time.Date(2026, 7, 30, 7, 18, 45, 0, time.UTC) // record-created "today"
	installDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	lastEvent := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	sub := &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		MyshopifyDomain: "store.myshopify.com",
		ShopifyShopGID:  "gid://shopify/Shop/1",
		Status:          "ACTIVE",
		CreatedAt:       rebuildTime,
		UpdatedAt:       rebuildTime,
	}

	// Events are keyed by the myshopify domain (how EventProcessor stores charged shops).
	events := []*entity.AppEvent{
		{ID: uuid.New(), AppID: appID, ShopifyShopGID: "store.myshopify.com", EventType: "RELATIONSHIP_INSTALLED", OccurredAt: installDate},
		{ID: uuid.New(), AppID: appID, ShopifyShopGID: "store.myshopify.com", EventType: "SUBSCRIPTION_CHARGE_ACTIVATED", OccurredAt: lastEvent},
	}

	handler := NewStoreHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{sub}},
		&mockTxRepo{},
		&mockEventRepo{events: events},
		&mockPartnerRepoForSub{account: partnerAccount},
		&mockAppRepoForSub{app: app},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/stores", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Stores []struct {
			FirstInstallDate string `json:"first_install_date"`
			LastInteraction  string `json:"last_interaction"`
		} `json:"stores"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Stores) != 1 {
		t.Fatalf("expected 1 store, got %d", len(resp.Stores))
	}

	if got, want := resp.Stores[0].FirstInstallDate, installDate.Format(time.RFC3339); got != want {
		t.Errorf("first_install_date = %q, want %q (event date, not CreatedAt)", got, want)
	}
	if got, want := resp.Stores[0].LastInteraction, lastEvent.Format(time.RFC3339); got != want {
		t.Errorf("last_interaction = %q, want %q (latest event, not UpdatedAt)", got, want)
	}
	if resp.Stores[0].FirstInstallDate == rebuildTime.Format(time.RFC3339) {
		t.Error("first_install_date still equals the record-created rebuild time (STORE-2 regression)")
	}
}

// TestStoreHandler_List_UsesPersistedRiskState is the STORE-1 regression guard:
// the store badge/health must reflect the persisted (reconciled) risk_state, not a
// live re-classification. A cancel-trap store (CANCELLED but persisted SAFE) must
// not be re-churned on the Stores page.
func TestStoreHandler_List_UsesPersistedRiskState(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App", PartnerAppID: "gid://partners/App/1"}

	past := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	sub := &entity.Subscription{
		ID: uuid.New(), AppID: appID, MyshopifyDomain: "trap.myshopify.com",
		Status: "CANCELLED", ExpectedNextChargeDate: &past, RiskState: valueobject.RiskStateSafe,
	}

	handler := NewStoreHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{sub}},
		&mockTxRepo{},
		nil, // appEventRepo optional
		&mockPartnerRepoForSub{account: partnerAccount},
		&mockAppRepoForSub{app: app},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/stores", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}))

	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Stores []struct {
			RiskState   string `json:"risk_state"`
			HealthScore int    `json:"health_score"`
		} `json:"stores"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Stores) != 1 {
		t.Fatalf("expected 1 store, got %d", len(resp.Stores))
	}
	if resp.Stores[0].RiskState != string(valueobject.RiskStateSafe) {
		t.Errorf("risk_state = %q, want SAFE (persisted; a live re-classify would churn the CANCELLED store)", resp.Stores[0].RiskState)
	}
	if resp.Stores[0].HealthScore != 90 {
		t.Errorf("health_score = %d, want 90 (derived from persisted SAFE)", resp.Stores[0].HealthScore)
	}
}

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// TestRiskHandler_Summary_UsesPersistedRiskState is the RISK-1b regression guard:
// the risk distribution must reflect the PERSISTED (reconciled) risk_state, not a
// live RiskEngine.ClassifyAll re-classification. The classic case is a cancel-trap
// subscription — Status CANCELLED but kept SAFE by ApplyEventStatus because a
// recurring charge post-dates the cancel. A naive re-classification would churn it
// (terminal → CHURNED); the persisted state (SAFE) is authoritative.
func TestRiskHandler_Summary_UsesPersistedRiskState(t *testing.T) {
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App", PartnerAppID: "gid://partners/App/1"}

	past := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // well past → ClassifyAll would churn

	subs := []*entity.Subscription{
		// Cancel-trap: CANCELLED status but persisted SAFE. Naive ClassifyAll would
		// mark this CHURNED (terminal) — the summary must keep it SAFE.
		{ID: uuid.New(), AppID: appID, MyshopifyDomain: "trap.myshopify.com", Status: "CANCELLED", ExpectedNextChargeDate: &past, RiskState: valueobject.RiskStateSafe},
		// A genuinely at-risk sub (persisted).
		{ID: uuid.New(), AppID: appID, MyshopifyDomain: "atrisk.myshopify.com", Status: "ACTIVE", ExpectedNextChargeDate: &past, RiskState: valueobject.RiskStateOneCycleMissed},
	}

	handler := NewRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		nil, // appEventRepo optional
		&mockPartnerRepoForSub{account: partnerAccount},
		&mockAppRepoForSub{app: app},
		domainservice.NewRiskEngine(),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/risk/summary", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}))

	rec := httptest.NewRecorder()
	handler.Summary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Distribution struct {
			Safe     int `json:"safe"`
			OneCycle int `json:"one_cycle"`
			Churned  int `json:"churned"`
		} `json:"distribution"`
		AtRiskStores []struct {
			ShopDomain      string   `json:"shop_domain"`
			InstalledAppIDs []string `json:"installed_app_ids"`
		} `json:"at_risk_stores"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Distribution.Safe != 1 {
		t.Errorf("safe = %d, want 1 (cancel-trap sub kept SAFE from persisted state)", resp.Distribution.Safe)
	}
	if resp.Distribution.Churned != 0 {
		t.Errorf("churned = %d, want 0 (naive re-classification would have churned the CANCELLED sub)", resp.Distribution.Churned)
	}
	if resp.Distribution.OneCycle != 1 {
		t.Errorf("one_cycle = %d, want 1", resp.Distribution.OneCycle)
	}
	if len(resp.AtRiskStores) != 1 || resp.AtRiskStores[0].ShopDomain != "atrisk.myshopify.com" {
		t.Errorf("at_risk_stores = %+v, want [atrisk.myshopify.com]", resp.AtRiskStores)
	}
	// RISK-2: installed_app_ids must be populated with the app (was empty []).
	if got := resp.AtRiskStores[0].InstalledAppIDs; len(got) != 1 || got[0] != appID.String() {
		t.Errorf("installed_app_ids = %v, want [%s]", got, appID.String())
	}
}

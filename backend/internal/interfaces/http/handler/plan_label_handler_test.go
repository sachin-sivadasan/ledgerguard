package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// mockPlanLabelRepo is a test double for repository.PlanLabelRepository.
type mockPlanLabelRepo struct {
	labels     []*entity.PlanLabel
	findErr    error
	replaceErr error
	saved      []*entity.PlanLabel // captured by ReplaceAll
}

func (m *mockPlanLabelRepo) FindByAppID(_ context.Context, _ uuid.UUID) ([]*entity.PlanLabel, error) {
	return m.labels, m.findErr
}

func (m *mockPlanLabelRepo) ReplaceAll(_ context.Context, _ uuid.UUID, labels []*entity.PlanLabel) error {
	if m.replaceErr != nil {
		return m.replaceErr
	}
	m.saved = labels
	return nil
}

func newPlanLabelHandler(appID uuid.UUID, pa *entity.PartnerAccount, subs []*entity.Subscription, repo *mockPlanLabelRepo) *PlanLabelHandler {
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	return NewPlanLabelHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		repo,
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
}

func doPlanLabels(t *testing.T, h *PlanLabelHandler, method string, appID uuid.UUID, pa *entity.PartnerAccount, body string, withUser bool) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *strings.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	} else {
		rdr = strings.NewReader("")
	}
	req := httptest.NewRequest(method, "/api/v1/apps/"+appID.String()+"/plan-labels", rdr)
	req = withURLParam(req, "appID", appID.String())
	if withUser {
		req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	}
	rec := httptest.NewRecorder()
	if method == http.MethodPut {
		h.PutPlanLabels(rec, req)
	} else {
		h.GetPlanLabels(rec, req)
	}
	return rec
}

// TestPlanLabels_GetDetectsTiers: GET returns the distinct price tiers among un-named subs,
// each with its pseudo-label + any saved custom label + active-customer count.
func TestPlanLabels_GetDetectsTiers(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	mo := func(domain string, price int64, state valueobject.RiskState) *entity.Subscription {
		s := atRiskSub(appID, domain, "", price, state)
		s.BillingInterval = valueobject.BillingIntervalMonthly
		return s
	}
	subs := []*entity.Subscription{
		mo("a.myshopify.com", 2900, valueobject.RiskStateSafe),
		mo("b.myshopify.com", 2900, valueobject.RiskStateSafe),
		mo("c.myshopify.com", 1900, valueobject.RiskStateSafe),
		mo("d.myshopify.com", 1900, valueobject.RiskStateChurned), // churned → excluded
	}
	repo := &mockPlanLabelRepo{labels: []*entity.PlanLabel{
		{AppID: appID, BillingInterval: valueobject.BillingIntervalMonthly, PriceCents: 2900, Label: "Starter"},
	}}
	rec := doPlanLabels(t, newPlanLabelHandler(appID, pa, subs, repo), http.MethodGet, appID, pa, "", true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp planLabelsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Tiers) != 2 {
		t.Fatalf("tiers = %d, want 2 price tiers", len(resp.Tiers))
	}
	// $29 tier leads (2 customers > 1); carries the saved "Starter" label.
	if resp.Tiers[0].PriceCents != 2900 || resp.Tiers[0].Customers != 2 || resp.Tiers[0].Label != "Starter" {
		t.Errorf("tier[0] = %+v, want $29/2 customers/Starter", resp.Tiers[0])
	}
	// $19 tier: 1 active customer (churned excluded), no saved label → pseudo only.
	if resp.Tiers[1].PriceCents != 1900 || resp.Tiers[1].Customers != 1 || resp.Tiers[1].Label != "" {
		t.Errorf("tier[1] = %+v, want $19/1 customer/no label", resp.Tiers[1])
	}
	if resp.Tiers[1].PseudoLabel != "$19.00/mo" {
		t.Errorf("tier[1] pseudo = %q, want $19.00/mo", resp.Tiers[1].PseudoLabel)
	}
}

// TestPlanLabels_PutSavesAndClears: PUT saves non-empty labels and drops cleared ones.
func TestPlanLabels_PutSavesAndClears(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	repo := &mockPlanLabelRepo{}
	body := `{"labels":[
		{"billingInterval":"MONTHLY","priceCents":2900,"label":"Starter"},
		{"billingInterval":"MONTHLY","priceCents":1900,"label":"  "}
	]}`
	rec := doPlanLabels(t, newPlanLabelHandler(appID, pa, nil, repo), http.MethodPut, appID, pa, body, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	// Only the non-blank label persists.
	if len(repo.saved) != 1 || repo.saved[0].Label != "Starter" || repo.saved[0].PriceCents != 2900 {
		t.Errorf("saved = %+v, want just the $29 Starter", repo.saved)
	}
}

func TestPlanLabels_PutRejectsDuplicateTier(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	body := `{"labels":[
		{"billingInterval":"MONTHLY","priceCents":2900,"label":"Starter"},
		{"billingInterval":"MONTHLY","priceCents":2900,"label":"Growth"}
	]}`
	rec := doPlanLabels(t, newPlanLabelHandler(appID, pa, nil, &mockPlanLabelRepo{}), http.MethodPut, appID, pa, body, true)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on duplicate tier, got %d", rec.Code)
	}
}

func TestPlanLabels_401WithoutUser(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	rec := doPlanLabels(t, newPlanLabelHandler(appID, pa, nil, &mockPlanLabelRepo{}), http.MethodGet, appID, pa, "", false)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

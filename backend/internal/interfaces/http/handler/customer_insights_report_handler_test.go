package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func newCustomerInsightsHandler(appID uuid.UUID, pa *entity.PartnerAccount, subs []*entity.Subscription, findErr error) *CustomerInsightsReportHandler {
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	return NewCustomerInsightsReportHandler(
		&mockSubscriptionRepo{subscriptions: subs, findAllErr: findErr},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		nil,
	)
}

func doCustomerInsights(t *testing.T, h *CustomerInsightsReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string, withUser bool) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/customer-insights"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	if withUser {
		req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	}
	rec := httptest.NewRecorder()
	h.GetCustomerInsights(rec, req)
	return rec
}

// baseInsightsSubs: an active base with two plans, a mix of risk, and one churned sub.
//
//	Pro   ($50): 2 SAFE + 1 at-risk  → $150 MRR, $50 at-risk
//	Basic ($10): 1 SAFE              → $10 MRR, band < $25
//	Basic ($10): 1 CHURNED           → excluded from the active base
func baseInsightsSubs(appID uuid.UUID) []*entity.Subscription {
	return []*entity.Subscription{
		atRiskSub(appID, "a.myshopify.com", "Pro", 5000, valueobject.RiskStateSafe),
		atRiskSub(appID, "b.myshopify.com", "Pro", 5000, valueobject.RiskStateSafe),
		atRiskSub(appID, "c.myshopify.com", "Pro", 5000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "d.myshopify.com", "Basic", 1000, valueobject.RiskStateSafe),
		atRiskSub(appID, "e.myshopify.com", "Basic", 1000, valueobject.RiskStateChurned),
	}
}

func TestCustomerInsights_SegmentsAndKPIs(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newCustomerInsightsHandler(appID, pa, baseInsightsSubs(appID), nil)
	rec := doCustomerInsights(t, h, appID, pa, "", true)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp customerInsightsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Active base = non-churned (4 subs), $160 MRR.
	if resp.TotalCustomers != 4 {
		t.Errorf("totalCustomers = %d, want 4 (churned excluded)", resp.TotalCustomers)
	}
	if resp.ActiveMrrCents != 16000 {
		t.Errorf("activeMrrCents = %d, want 16000", resp.ActiveMrrCents)
	}
	if resp.AtRiskCustomers != 1 || resp.AtRiskMrrCents != 5000 {
		t.Errorf("at-risk = %d/%d, want 1/5000", resp.AtRiskCustomers, resp.AtRiskMrrCents)
	}

	// Risk segments span the whole base, churned included.
	seg := map[string]riskSegment{}
	for _, s := range resp.RiskSegments {
		seg[s.RiskState] = s
	}
	if seg["SAFE"].Customers != 3 || seg["SAFE"].MrrCents != 11000 {
		t.Errorf("SAFE = %d/%d, want 3/11000", seg["SAFE"].Customers, seg["SAFE"].MrrCents)
	}
	if seg["AT_RISK"].Customers != 1 || seg["AT_RISK"].MrrCents != 5000 {
		t.Errorf("AT_RISK = %d/%d, want 1/5000", seg["AT_RISK"].Customers, seg["AT_RISK"].MrrCents)
	}
	if seg["CHURNED"].Customers != 1 || seg["CHURNED"].MrrCents != 1000 {
		t.Errorf("CHURNED = %d/%d, want 1/1000", seg["CHURNED"].Customers, seg["CHURNED"].MrrCents)
	}

	// Plan crosstab: Pro first (higher MRR), with the at-risk split surfaced.
	if len(resp.PlanRisk) != 2 || resp.PlanRisk[0].PlanName != "Pro" {
		t.Fatalf("planRisk = %+v, want Pro first", resp.PlanRisk)
	}
	pro := resp.PlanRisk[0]
	if pro.Customers != 3 || pro.SafeCount != 2 || pro.AtRiskCount != 1 || pro.MrrCents != 15000 || pro.AtRiskMrrCents != 5000 {
		t.Errorf("Pro row = %+v, want 3/safe2/atrisk1/mrr15000/atrisk5000", pro)
	}
}

func TestCustomerInsights_RevenueBands(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newCustomerInsightsHandler(appID, pa, baseInsightsSubs(appID), nil)
	rec := doCustomerInsights(t, h, appID, pa, "", true)
	var resp customerInsightsReport
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	band := map[string]revenueBand{}
	for _, b := range resp.RevenueBands {
		band[b.Label] = b
	}
	// Basic ($10) → "< $25"; the three Pro ($50) → "$50–$100".
	if band["< $25"].Customers != 1 || band["< $25"].MrrCents != 1000 {
		t.Errorf(`"< $25" = %d/%d, want 1/1000`, band["< $25"].Customers, band["< $25"].MrrCents)
	}
	if band["$50–$100"].Customers != 3 || band["$50–$100"].MrrCents != 15000 {
		t.Errorf(`"$50–$100" = %d/%d, want 3/15000`, band["$50–$100"].Customers, band["$50–$100"].MrrCents)
	}
	// 3 of 4 active customers sit in the $50–$100 band.
	if got := band["$50–$100"].PctOfCustomers; got < 0.74 || got > 0.76 {
		t.Errorf(`"$50–$100" pct = %v, want ~0.75`, got)
	}
	// Every band definition is always present (even empty ones), for a stable UI.
	if len(resp.RevenueBands) != len(revenueBandDefs) {
		t.Errorf("bands = %d, want %d (all bands always present)", len(resp.RevenueBands), len(revenueBandDefs))
	}
}

func TestCustomerInsights_TopCustomersOrderedAndActiveOnly(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newCustomerInsightsHandler(appID, pa, baseInsightsSubs(appID), nil)
	rec := doCustomerInsights(t, h, appID, pa, "", true)
	var resp customerInsightsReport
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	// 4 active customers (churned "e" excluded), highest MRR first.
	if len(resp.TopCustomers) != 4 {
		t.Fatalf("topCustomers = %d, want 4 (churned excluded)", len(resp.TopCustomers))
	}
	if resp.TopCustomers[0].MrrCents != 5000 {
		t.Errorf("top MRR = %d, want 5000 first", resp.TopCustomers[0].MrrCents)
	}
	if last := resp.TopCustomers[3]; last.MrrCents != 1000 || last.PlanName != "Basic" {
		t.Errorf("last = %+v, want the $10 Basic customer", last)
	}
	for _, c := range resp.TopCustomers {
		if c.RiskState == string(valueobject.RiskStateChurned) {
			t.Errorf("churned customer %s leaked into top customers", c.ShopName)
		}
	}
}

// TestCustomerInsights_PseudoPlanLabels: when subs have no plan name (the common case —
// the Partner API doesn't provide one), the crosstab segments by a price-tier pseudo-label
// instead of collapsing into one blank row.
func TestCustomerInsights_PseudoPlanLabels(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	// Two price tiers, no plan names.
	a := atRiskSub(appID, "a.myshopify.com", "", 5000, valueobject.RiskStateSafe)
	b := atRiskSub(appID, "b.myshopify.com", "", 5000, valueobject.RiskStateSafe)
	c := atRiskSub(appID, "c.myshopify.com", "", 1000, valueobject.RiskStateSafe)
	h := newCustomerInsightsHandler(appID, pa, []*entity.Subscription{a, b, c}, nil)
	rec := doCustomerInsights(t, h, appID, pa, "", true)

	var resp customerInsightsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Two distinct price tiers → two named rows, not one blank row.
	if len(resp.PlanRisk) != 2 {
		t.Fatalf("planRisk rows = %d, want 2 price tiers", len(resp.PlanRisk))
	}
	labels := map[string]bool{}
	for _, p := range resp.PlanRisk {
		if p.PlanName == "" {
			t.Errorf("empty plan label leaked into the crosstab")
		}
		labels[p.PlanName] = true
	}
	if !labels["$50.00/mo"] || !labels["$10.00/mo"] {
		t.Errorf("expected pseudo-labels $50.00/mo and $10.00/mo, got %v", labels)
	}
}

func TestCustomerInsights_401WithoutUser(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newCustomerInsightsHandler(appID, pa, baseInsightsSubs(appID), nil)
	rec := doCustomerInsights(t, h, appID, pa, "", false)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestCustomerInsights_503OnRepoError(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newCustomerInsightsHandler(appID, pa, nil, errors.New("db down"))
	rec := doCustomerInsights(t, h, appID, pa, "", true)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestCustomerInsights_CSV(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newCustomerInsightsHandler(appID, pa, baseInsightsSubs(appID), nil)
	rec := doCustomerInsights(t, h, appID, pa, "format=csv", true)

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("content-type = %q, want text/csv", ct)
	}
	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// Header + 2 plan rows.
	if len(rows) != 3 {
		t.Fatalf("CSV rows = %d, want 3 (header + 2 plans)", len(rows))
	}
	if rows[0][0] != "plan" || rows[1][0] != "Pro" {
		t.Errorf("CSV header/first row unexpected: %v / %v", rows[0], rows[1])
	}
}

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// atRiskSub is a small builder for at-risk subscription test data.
func atRiskSub(appID uuid.UUID, domain, plan string, cents int64, state valueobject.RiskState) *entity.Subscription {
	expected := time.Now().UTC().AddDate(0, 0, -45)
	return &entity.Subscription{
		ID:                     uuid.New(),
		AppID:                  appID,
		MyshopifyDomain:        domain,
		ShopName:               strings.Split(domain, ".")[0],
		PlanName:               plan,
		BasePriceCents:         cents,
		Currency:               "USD",
		RiskState:              state,
		ExpectedNextChargeDate: &expected,
	}
}

func newRevenueAtRiskRequest(t *testing.T, appID uuid.UUID, partnerAccount *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/revenue-at-risk"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func TestRevenueAtRisk_TotalsAndByState(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		atRiskSub(appID, "one-a.myshopify.com", "Pro", 5000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "one-b.myshopify.com", "Basic", 3000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "two-a.myshopify.com", "Pro", 4000, valueobject.RiskStateTwoCyclesMissed),
		// Safe sub must be excluded.
		atRiskSub(appID, "safe.myshopify.com", "Pro", 9999, valueobject.RiskStateSafe),
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp revenueAtRiskReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.ByState.OneCycleCents != 8000 {
		t.Errorf("oneCycleCents: expected 8000, got %d", resp.ByState.OneCycleCents)
	}
	if resp.ByState.TwoCycleCents != 4000 {
		t.Errorf("twoCycleCents: expected 4000, got %d", resp.ByState.TwoCycleCents)
	}
	if resp.TotalAtRiskCents != 12000 {
		t.Errorf("totalAtRiskCents: expected 12000, got %d", resp.TotalAtRiskCents)
	}
	if resp.Counts.OneCycle != 2 || resp.Counts.TwoCycle != 1 {
		t.Errorf("counts: expected {2,1}, got {%d,%d}", resp.Counts.OneCycle, resp.Counts.TwoCycle)
	}
	// Acceptance: sum of store MRRs == totalAtRiskCents.
	var storeSum int64
	for _, s := range resp.Stores {
		storeSum += s.MRRCents
	}
	if storeSum != resp.TotalAtRiskCents {
		t.Errorf("store MRR sum %d != totalAtRiskCents %d", storeSum, resp.TotalAtRiskCents)
	}
}

func TestRevenueAtRisk_RecoverableCalc(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		atRiskSub(appID, "one.myshopify.com", "Pro", 10000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "two.myshopify.com", "Pro", 10000, valueobject.RiskStateTwoCyclesMissed),
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	var resp revenueAtRiskReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// 10000*0.60 + 10000*0.25 = 6000 + 2500 = 8500
	if resp.RecoverableCents != 8500 {
		t.Errorf("recoverableCents: expected 8500, got %d", resp.RecoverableCents)
	}
	// Per-store recoverable uses each state's rate.
	for _, s := range resp.Stores {
		var want int64
		switch s.RiskState {
		case valueobject.RiskStateOneCycleMissed.String():
			want = 6000
		case valueobject.RiskStateTwoCyclesMissed.String():
			want = 2500
		}
		if s.RecoverableCents != want {
			t.Errorf("store %s recoverable: expected %d, got %d", s.Domain, want, s.RecoverableCents)
		}
	}
}

func TestRevenueAtRisk_StoresSortedByMRRDesc(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		atRiskSub(appID, "low.myshopify.com", "Pro", 1000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "high.myshopify.com", "Pro", 9000, valueobject.RiskStateTwoCyclesMissed),
		atRiskSub(appID, "mid.myshopify.com", "Pro", 5000, valueobject.RiskStateOneCycleMissed),
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	var resp revenueAtRiskReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Stores) != 3 {
		t.Fatalf("expected 3 stores, got %d", len(resp.Stores))
	}
	wantOrder := []int64{9000, 5000, 1000}
	for i, want := range wantOrder {
		if resp.Stores[i].MRRCents != want {
			t.Errorf("stores[%d].mrrCents: expected %d, got %d", i, want, resp.Stores[i].MRRCents)
		}
	}
}

func TestRevenueAtRisk_Empty(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	// Only safe subs → nothing at risk.
	subs := []*entity.Subscription{
		atRiskSub(appID, "safe.myshopify.com", "Pro", 5000, valueobject.RiskStateSafe),
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp revenueAtRiskReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.TotalAtRiskCents != 0 || resp.RecoverableCents != 0 {
		t.Errorf("expected zero totals, got total=%d recoverable=%d", resp.TotalAtRiskCents, resp.RecoverableCents)
	}
	if len(resp.Stores) != 0 {
		t.Errorf("expected empty stores, got %d", len(resp.Stores))
	}
	// stores/trend must serialize as [] (non-nil) for a clean JSON contract.
	if !strings.Contains(rec.Body.String(), `"stores":[]`) {
		t.Errorf("expected stores serialized as [], body: %s", rec.Body.String())
	}
}

func TestRevenueAtRisk_SegmentFilter(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		atRiskSub(appID, "pro.myshopify.com", "Pro", 5000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "basic.myshopify.com", "Basic", 3000, valueobject.RiskStateOneCycleMissed),
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "segment=plan:Pro")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	var resp revenueAtRiskReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Stores) != 1 || resp.Stores[0].PlanName != "Pro" {
		t.Fatalf("expected only Pro store, got %+v", resp.Stores)
	}
	if resp.TotalAtRiskCents != 5000 {
		t.Errorf("expected totalAtRiskCents=5000, got %d", resp.TotalAtRiskCents)
	}
}

func TestRevenueAtRisk_Trend(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snapshots := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, RevenueAtRiskCents: 150000},
		{ID: uuid.New(), AppID: appID, Date: d.AddDate(0, 0, 1), RevenueAtRiskCents: 160000},
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snapshots},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "from=2026-06-01&to=2026-07-31")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	var resp revenueAtRiskReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-07-01" || resp.Trend[0].AtRiskCents != 150000 {
		t.Errorf("trend[0]: expected {2026-07-01,150000}, got {%s,%d}", resp.Trend[0].Date, resp.Trend[0].AtRiskCents)
	}
}

func TestRevenueAtRisk_CSVFormat(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		atRiskSub(appID, "high.myshopify.com", "Pro", 9000, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "low.myshopify.com", "Basic", 1000, valueobject.RiskStateTwoCyclesMissed),
	}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "format=csv")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "revenue-at-risk.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header + 2 rows, got %d lines: %v", len(lines), lines)
	}
	header := "domain,shopName,mrrCents,riskState,daysLate,expectedChargeDate,planName,recoverableCents"
	if lines[0] != header {
		t.Errorf("header mismatch:\n got: %s\nwant: %s", lines[0], header)
	}
	// Sorted by MRR desc: high first.
	if !strings.HasPrefix(lines[1], "high.myshopify.com,") {
		t.Errorf("expected first row to be high MRR store, got %s", lines[1])
	}
}

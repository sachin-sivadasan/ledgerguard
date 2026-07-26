package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
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

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d records: %v", len(records), records)
	}
	wantHeader := []string{"domain", "shopName", "mrrCents", "riskState", "daysLate", "expectedChargeDate", "planName", "recoverableCents"}
	if len(records[0]) != len(wantHeader) {
		t.Fatalf("header column count: expected %d, got %d", len(wantHeader), len(records[0]))
	}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by MRR desc: high first.
	if records[1][0] != "high.myshopify.com" {
		t.Errorf("expected first row to be high MRR store, got %s", records[1][0])
	}
}

// TestRevenueAtRisk_AnnualMRRNormalization verifies annual subs are normalized to
// monthly (÷12) in totals and per-state sums, not counted at their full base price.
func TestRevenueAtRisk_AnnualMRRNormalization(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	annual := atRiskSub(appID, "annual.myshopify.com", "Pro", 1200, valueobject.RiskStateOneCycleMissed)
	annual.BillingInterval = valueobject.BillingIntervalAnnual
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{annual}},
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
	// 1200 annual ÷ 12 = 100 monthly.
	if resp.TotalAtRiskCents != 100 {
		t.Errorf("totalAtRiskCents: expected 100 (1200/12), got %d", resp.TotalAtRiskCents)
	}
	if resp.ByState.OneCycleCents != 100 {
		t.Errorf("oneCycleCents: expected 100, got %d", resp.ByState.OneCycleCents)
	}
	if len(resp.Stores) != 1 || resp.Stores[0].MRRCents != 100 {
		t.Fatalf("store MRR: expected 100, got %+v", resp.Stores)
	}
}

// TestRevenueAtRisk_RepoErrorReturns503 verifies infra repo failures surface as 503.
func TestRevenueAtRisk_RepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestRevenueAtRisk_NilExpectedChargeDate verifies a store with no expected charge
// date reports daysLate=0 and an empty expectedChargeDate.
func TestRevenueAtRisk_NilExpectedChargeDate(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	sub := atRiskSub(appID, "nodate.myshopify.com", "Pro", 5000, valueobject.RiskStateOneCycleMissed)
	sub.ExpectedNextChargeDate = nil
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{sub}},
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
	if len(resp.Stores) != 1 {
		t.Fatalf("expected 1 store, got %d", len(resp.Stores))
	}
	if resp.Stores[0].DaysLate != 0 {
		t.Errorf("daysLate: expected 0, got %d", resp.Stores[0].DaysLate)
	}
	if resp.Stores[0].ExpectedChargeDate != "" {
		t.Errorf("expectedChargeDate: expected empty, got %q", resp.Stores[0].ExpectedChargeDate)
	}
}

// TestRevenueAtRisk_CSVEscaping verifies a shopName containing a comma stays a single
// quoted field, keeping the row at 8 columns.
func TestRevenueAtRisk_CSVEscaping(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	sub := atRiskSub(appID, "comma.myshopify.com", "Pro", 5000, valueobject.RiskStateOneCycleMissed)
	sub.ShopName = "Acme, Inc."
	h := NewRevenueAtRiskHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{sub}},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newRevenueAtRiskRequest(t, appID, partnerAccount, "format=csv")
	rec := httptest.NewRecorder()
	h.GetRevenueAtRisk(rec, req)

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d records", len(records))
	}
	row := records[1]
	if len(row) != 8 {
		t.Fatalf("expected 8 columns, got %d: %v", len(row), row)
	}
	if row[1] != "Acme, Inc." {
		t.Errorf("shopName: expected %q, got %q", "Acme, Inc.", row[1])
	}
}

// TestParseDateRange unit-tests parseDateRange directly (same package).
func TestParseDateRange(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	// Default: no params → to≈now, from=to-30d.
	from, to := parseDateRange("", "", now)
	if !to.Equal(now) {
		t.Errorf("default to: expected %v, got %v", now, to)
	}
	if want := now.AddDate(0, 0, -30); !from.Equal(want) {
		t.Errorf("default from: expected %v, got %v", want, from)
	}

	// Valid parse of both.
	from, to = parseDateRange("2026-06-01", "2026-06-30", now)
	if from.Format(dateLayout) != "2026-06-01" || to.Format(dateLayout) != "2026-06-30" {
		t.Errorf("valid parse: got from=%s to=%s", from.Format(dateLayout), to.Format(dateLayout))
	}

	// Malformed from with valid to → from falls back to to-30d.
	from, to = parseDateRange("not-a-date", "2026-06-30", now)
	if to.Format(dateLayout) != "2026-06-30" {
		t.Errorf("to: expected 2026-06-30, got %s", to.Format(dateLayout))
	}
	if want := to.AddDate(0, 0, -30); !from.Equal(want) {
		t.Errorf("fallback from: expected %v, got %v", want, from)
	}
}

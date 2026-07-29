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

func newActiveCustomersRequest(t *testing.T, appID uuid.UUID, pa *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/active-customers"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func activeCustomersFixture(subs []*entity.Subscription, snaps []*entity.DailyMetricsSnapshot, findErr error) (uuid.UUID, *entity.PartnerAccount, *ActiveCustomersReportHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewActiveCustomersReportHandler(
		&mockSubscriptionRepo{subscriptions: subs, findAllErr: findErr},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doActiveCustomers(t *testing.T, h *ActiveCustomersReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := newActiveCustomersRequest(t, appID, pa, query)
	rec := httptest.NewRecorder()
	h.GetActiveCustomersReport(rec, req)
	return rec
}

func decodeActiveCustomers(t *testing.T, rec *httptest.ResponseRecorder) activeCustomersReport {
	t.Helper()
	var resp activeCustomersReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// acSnap builds a snapshot with the given per-state counts. TotalSubscriptions is the
// sum; active (non-churned) = safe + oneCycle + twoCycle = Total − churned.
func acSnap(appID uuid.UUID, date time.Time, safe, oneCycle, twoCycle, churned int) *entity.DailyMetricsSnapshot {
	return &entity.DailyMetricsSnapshot{
		ID: uuid.New(), AppID: appID, Date: date,
		SafeCount: safe, OneCycleMissedCount: oneCycle, TwoCyclesMissedCount: twoCycle,
		ChurnedCount: churned, TotalSubscriptions: safe + oneCycle + twoCycle + churned,
	}
}

// TestActiveCustomers_HeadlineFromLatestSnapshot: headline = latest in-range snapshot's
// active count (Total − churned) and equals the last trend point.
func TestActiveCustomers_HeadlineFromLatestSnapshot(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		acSnap(appID, d, 10, 1, 1, 2),           // active 12
		acSnap(appID, d.AddDate(0, 0, 1), 12, 2, 0, 2), // active 14
	}
	h := NewActiveCustomersReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doActiveCustomers(t, h, appID, pa, "from=2026-07-01&to=2026-07-15") // ≤31d → daily
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeActiveCustomers(t, rec)
	if resp.ActiveCustomers != 14 {
		t.Errorf("activeCustomers: expected 14 (latest snapshot Total−churned), got %d", resp.ActiveCustomers)
	}
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-07-01" || resp.Trend[0].ActiveCustomers != 12 {
		t.Errorf("trend[0]: expected {2026-07-01,12}, got {%s,%d}", resp.Trend[0].Date, resp.Trend[0].ActiveCustomers)
	}
	if resp.ActiveCustomers != resp.Trend[len(resp.Trend)-1].ActiveCustomers {
		t.Errorf("headline %d != last trend point %d", resp.ActiveCustomers, resp.Trend[len(resp.Trend)-1].ActiveCustomers)
	}
}

// TestActiveCustomers_FallbackToCurrentSubs: with no in-range snapshot, headline =
// current non-churned subscription count.
func TestActiveCustomers_FallbackToCurrentSubs(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 5000),
		safeSub(appID, "b.myshopify.com", "Pro", 5000),
		atRiskSub(appID, "c.myshopify.com", "Starter", 2000, valueobject.RiskStateOneCycleMissed),
		churnedSub(appID, "d.myshopify.com", "Pro", 5000),
		churnedSub(appID, "e.myshopify.com", "Pro", 5000),
	}
	aid, pa, h := activeCustomersFixture(subs, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-31")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeActiveCustomers(t, rec)
	if resp.ActiveCustomers != 3 {
		t.Errorf("activeCustomers: expected 3 non-churned (2 safe + 1 at-risk), got %d", resp.ActiveCustomers)
	}
	if len(resp.Trend) != 0 {
		t.Errorf("expected empty trend (no snapshots), got %d", len(resp.Trend))
	}
}

// TestActiveCustomers_NewChurnedNetInRange: New = StartDate in range; Churned = churn
// date in range; Net = new − churned. Out-of-range subs are excluded.
func TestActiveCustomers_NewChurnedNetInRange(t *testing.T) {
	appID := uuid.New()
	inRange := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	outStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	churnIn := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	churnOut := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{
		nnNewSub(appID, "new.myshopify.com", "Pro", 5000, inRange),      // new (StartDate 07-10)
		nnNewSub(appID, "old.myshopify.com", "Pro", 5000, outStart),     // not new (06-01)
		nnChurnedSub(appID, "churn.myshopify.com", outStart, churnIn),   // churned in range (created out → not new)
		nnChurnedSub(appID, "churn2.myshopify.com", outStart, churnOut), // churned out of range
	}
	aid, pa, h := activeCustomersFixture(subs, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-31")
	resp := decodeActiveCustomers(t, rec)
	if resp.NewCount != 1 {
		t.Errorf("newCount: expected 1, got %d", resp.NewCount)
	}
	if resp.ChurnedCount != 1 {
		t.Errorf("churnedCount: expected 1, got %d", resp.ChurnedCount)
	}
	if resp.NetChange != 0 {
		t.Errorf("netChange: expected 0 (1−1), got %d", resp.NetChange)
	}
}

// TestActiveCustomers_ActiveByPlan: non-churned subs grouped by plan (count + MRR +
// share of active count), sorted by MRR desc; churned excluded.
func TestActiveCustomers_ActiveByPlan(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 5000),
		safeSub(appID, "b.myshopify.com", "Pro", 5000),
		safeSub(appID, "c.myshopify.com", "Starter", 2000),
		atRiskSub(appID, "d.myshopify.com", "Enterprise", 8000, valueobject.RiskStateOneCycleMissed),
		churnedSub(appID, "e.myshopify.com", "Pro", 5000), // excluded
	}
	aid, pa, h := activeCustomersFixture(subs, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "")
	resp := decodeActiveCustomers(t, rec)
	if len(resp.Plans) != 3 {
		t.Fatalf("expected 3 plans (churned-only excluded), got %d: %+v", len(resp.Plans), resp.Plans)
	}
	// Sorted by MRR desc: Pro 10000, Enterprise 8000, Starter 2000.
	if resp.Plans[0].PlanName != "Pro" || resp.Plans[0].ActiveSubs != 2 || resp.Plans[0].MrrCents != 10000 {
		t.Errorf("plans[0]: expected {Pro,2,10000}, got %+v", resp.Plans[0])
	}
	if resp.Plans[1].PlanName != "Enterprise" || resp.Plans[1].ActiveSubs != 1 {
		t.Errorf("plans[1]: expected {Enterprise,1}, got %+v", resp.Plans[1])
	}
	// % of active COUNT: Pro 2/4 = 0.5.
	if resp.Plans[0].PctOfActive != 0.5 {
		t.Errorf("plans[0].pctOfActive: expected 0.5, got %v", resp.Plans[0].PctOfActive)
	}
}

// TestActiveCustomers_CSV: format=csv returns the per-plan table as CSV.
func TestActiveCustomers_CSV(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 5000),
		safeSub(appID, "b.myshopify.com", "Starter", 2000),
	}
	aid, pa, h := activeCustomersFixture(subs, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("content-type: expected text/csv, got %q", ct)
	}
	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(rows) != 3 { // header + 2 plans
		t.Fatalf("expected 3 csv rows (header + 2), got %d", len(rows))
	}
	if rows[0][0] != "plan" || rows[0][1] != "activeSubs" {
		t.Errorf("csv header: got %v", rows[0])
	}
}

// TestActiveCustomers_RepoError503: a subscription-repo failure yields 503 (ADR-042).
func TestActiveCustomers_RepoError503(t *testing.T) {
	aid, pa, h := activeCustomersFixture(nil, nil, errors.New("db down"))
	rec := doActiveCustomers(t, h, aid, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

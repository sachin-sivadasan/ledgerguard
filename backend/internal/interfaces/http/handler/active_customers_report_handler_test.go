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
		nil,
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
// TestActiveCustomers_HeadlineReconcilesWithPlansNotSnapshot proves the headline is the
// live non-churned count (== sum of the "Active by plan" table) and does NOT follow the
// latest snapshot, even when the snapshot's active tally diverges. This guards the fix
// for the snapshot-vs-live headline gap seen in prod (headline 1016 vs table sum 926):
// the trend stays snapshot-sourced, but the big number must reconcile with the table.
func TestActiveCustomers_HeadlineReconcilesWithPlansNotSnapshot(t *testing.T) {
	appID := uuid.New()
	// 3 live non-churned subs (+1 churned) → headline must be 3.
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 5000),
		safeSub(appID, "b.myshopify.com", "Pro", 5000),
		atRiskSub(appID, "c.myshopify.com", "Starter", 2000, valueobject.RiskStateOneCycleMissed),
		churnedSub(appID, "d.myshopify.com", "Pro", 5000),
	}
	// Snapshots claim a much higher active count (14) — the headline must ignore it.
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		acSnap(appID, d, 10, 1, 1, 2),                  // active 12
		acSnap(appID, d.AddDate(0, 0, 1), 12, 2, 0, 2), // active 14 (latest)
	}
	aid, pa, h := activeCustomersFixture(subs, snaps, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-15") // ≤31d → daily
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeActiveCustomers(t, rec)
	if resp.ActiveCustomers != 3 {
		t.Errorf("activeCustomers: expected 3 (live non-churned), got %d — headline must not follow the snapshot's 14", resp.ActiveCustomers)
	}
	planSum := 0
	for _, p := range resp.Plans {
		planSum += p.ActiveSubs
	}
	if resp.ActiveCustomers != planSum {
		t.Errorf("headline %d must reconcile with plan-table sum %d", resp.ActiveCustomers, planSum)
	}
	// Trend is still the (independent) snapshot series: 2 points, last = 14.
	if len(resp.Trend) != 2 || resp.Trend[len(resp.Trend)-1].ActiveCustomers != 14 {
		t.Errorf("trend should stay snapshot-sourced (2 pts, last=14), got %+v", resp.Trend)
	}
}

// TestActiveCustomers_HeadlineIsLiveNonChurned: headline = current non-churned
// subscription count, independent of any snapshots.
func TestActiveCustomers_HeadlineIsLiveNonChurned(t *testing.T) {
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

// TestActiveCustomers_AllChurnedHeadlineZero: subs exist but every one is churned →
// empty plan table, headline 0, yet in-range churn still moves Churned/Net negative.
// Guards that the live headline is a true sum of the (empty) plan breakdown, not a
// fallback to the raw sub count.
func TestActiveCustomers_AllChurnedHeadlineZero(t *testing.T) {
	appID := uuid.New()
	older := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	inRange := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{
		nnChurnedSub(appID, "a.myshopify.com", older, inRange), // churns in-window
		churnedSub(appID, "b.myshopify.com", "Pro", 5000),      // churned, out-of-window
	}
	aid, pa, h := activeCustomersFixture(subs, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-31")
	resp := decodeActiveCustomers(t, rec)
	if resp.ActiveCustomers != 0 {
		t.Errorf("activeCustomers: expected 0 (all churned), got %d", resp.ActiveCustomers)
	}
	if len(resp.Plans) != 0 {
		t.Errorf("expected empty plan table, got %d rows", len(resp.Plans))
	}
	if resp.ChurnedCount != 1 || resp.NetChange != -1 {
		t.Errorf("expected churned=1 net=-1, got churned=%d net=%d", resp.ChurnedCount, resp.NetChange)
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

// TestActiveCustomers_NewKeysOffActivatedAt: "New" counts by StartDate()=ActivatedAt
// (the real business start), NOT the record-created CreatedAt (which resets on rebuild).
func TestActiveCustomers_NewKeysOffActivatedAt(t *testing.T) {
	appID := uuid.New()
	inRange := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	outRange := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	// A: ActivatedAt in-range, CreatedAt OUT → counts as new (StartDate = ActivatedAt).
	a := safeSub(appID, "a.myshopify.com", "Pro", 5000)
	a.ActivatedAt = &inRange
	a.CreatedAt = outRange
	// B: ActivatedAt OUT, CreatedAt in-range → NOT new (ActivatedAt wins over CreatedAt).
	b := safeSub(appID, "b.myshopify.com", "Pro", 5000)
	b.ActivatedAt = &outRange
	b.CreatedAt = time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	aid, pa, h := activeCustomersFixture([]*entity.Subscription{a, b}, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-31")
	resp := decodeActiveCustomers(t, rec)
	if resp.NewCount != 1 {
		t.Errorf("newCount: expected 1 (A via ActivatedAt; B's ActivatedAt out of range), got %d", resp.NewCount)
	}
}

// TestActiveCustomers_DateBoundaryInclusive: the `to` day is inclusive (toExclusive =
// to+1); a start/churn on `to` counts, on `to`+1 does not.
func TestActiveCustomers_DateBoundaryInclusive(t *testing.T) {
	appID := uuid.New()
	onTo := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	past := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	older := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	startOnTo := safeSub(appID, "s1.myshopify.com", "Pro", 5000)
	startOnTo.ActivatedAt = &onTo
	startPastTo := safeSub(appID, "s2.myshopify.com", "Pro", 5000)
	startPastTo.ActivatedAt = &past
	churnOnTo := nnChurnedSub(appID, "c1.myshopify.com", older, onTo)   // churn on to
	churnPastTo := nnChurnedSub(appID, "c2.myshopify.com", older, past) // churn past to

	subs := []*entity.Subscription{startOnTo, startPastTo, churnOnTo, churnPastTo}
	aid, pa, h := activeCustomersFixture(subs, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-31")
	resp := decodeActiveCustomers(t, rec)
	if resp.NewCount != 1 {
		t.Errorf("newCount: expected 1 (start on `to` inclusive; on to+1 excluded), got %d", resp.NewCount)
	}
	if resp.ChurnedCount != 1 {
		t.Errorf("churnedCount: expected 1 (churn on `to` inclusive; on to+1 excluded), got %d", resp.ChurnedCount)
	}
}

// TestActiveCustomers_SameSubStartAndChurn: a single sub that started AND churned
// in-window counts toward BOTH new and churned (net nets to zero).
func TestActiveCustomers_SameSubStartAndChurn(t *testing.T) {
	appID := uuid.New()
	act := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	churn := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	s := atRiskSub(appID, "x.myshopify.com", "Pro", 1000, valueobject.RiskStateChurned)
	s.ActivatedAt = &act
	s.ExpectedNextChargeDate = &churn // churnedDateOf prefers this
	aid, pa, h := activeCustomersFixture([]*entity.Subscription{s}, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "from=2026-07-01&to=2026-07-31")
	resp := decodeActiveCustomers(t, rec)
	if resp.NewCount != 1 || resp.ChurnedCount != 1 || resp.NetChange != 0 {
		t.Errorf("same-sub start+churn: expected new=1 churned=1 net=0, got new=%d churned=%d net=%d",
			resp.NewCount, resp.ChurnedCount, resp.NetChange)
	}
}

// TestActiveCustomers_EmptyData: no subs + no snapshots → all zeros, USD default,
// empty plans/trend, and no divide-by-zero in pctOfActive.
func TestActiveCustomers_EmptyData(t *testing.T) {
	aid, pa, h := activeCustomersFixture(nil, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	resp := decodeActiveCustomers(t, rec)
	if resp.ActiveCustomers != 0 || resp.NewCount != 0 || resp.ChurnedCount != 0 || resp.NetChange != 0 {
		t.Errorf("empty: expected all-zero counts, got %+v", resp)
	}
	if resp.Currency != "USD" {
		t.Errorf("empty currency: expected USD, got %q", resp.Currency)
	}
	if len(resp.Plans) != 0 || len(resp.Trend) != 0 {
		t.Errorf("empty: expected no plans/trend, got %d plans / %d trend", len(resp.Plans), len(resp.Trend))
	}
}

// TestActiveCustomers_CurrencyFromSubs: currency = first non-empty sub currency.
func TestActiveCustomers_CurrencyFromSubs(t *testing.T) {
	appID := uuid.New()
	eur := safeSub(appID, "eu.myshopify.com", "Pro", 5000)
	eur.Currency = "EUR"
	aid, pa, h := activeCustomersFixture([]*entity.Subscription{eur}, nil, nil)
	rec := doActiveCustomers(t, h, aid, pa, "")
	resp := decodeActiveCustomers(t, rec)
	if resp.Currency != "EUR" {
		t.Errorf("currency: expected EUR, got %q", resp.Currency)
	}
}

// TestActiveCustomers_MonthlyGranularity: a >92-day window downsamples the trend to
// monthly buckets (last-per-bucket) and reports interval "month".
func TestActiveCustomers_MonthlyGranularity(t *testing.T) {
	appID := uuid.New()
	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		acSnap(appID, jan1, 10, 0, 0, 0),                   // Jan
		acSnap(appID, jan1.AddDate(0, 0, 15), 12, 0, 0, 0), // Jan (later → wins bucket)
		acSnap(appID, jan1.AddDate(0, 1, 0), 14, 0, 0, 0),  // Feb
	}
	appID2 := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID2, PartnerAccountID: pa.ID, Name: "Test App"}
	// Snapshots are keyed to appID but the mock ignores appID, so reuse them.
	_ = appID
	h := NewActiveCustomersReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		nil,
	)
	rec := doActiveCustomers(t, h, appID2, pa, "from=2026-01-01&to=2026-06-30") // >92d → month
	resp := decodeActiveCustomers(t, rec)
	if resp.Interval != "month" {
		t.Errorf("interval: expected month, got %q", resp.Interval)
	}
	if len(resp.Trend) != 2 { // Jan (last-per-bucket = 12) + Feb (14)
		t.Fatalf("expected 2 monthly points, got %d: %+v", len(resp.Trend), resp.Trend)
	}
	if resp.Trend[0].ActiveCustomers != 12 {
		t.Errorf("Jan bucket: expected last-per-bucket 12, got %d", resp.Trend[0].ActiveCustomers)
	}
}

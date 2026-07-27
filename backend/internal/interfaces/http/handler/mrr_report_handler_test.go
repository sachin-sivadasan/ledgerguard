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

// newMRRRequest builds an authenticated GET request for the MRR report.
func newMRRRequest(t *testing.T, appID uuid.UUID, partnerAccount *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/mrr"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func mrrFixture(subs []*entity.Subscription, snaps []*entity.DailyMetricsSnapshot) (uuid.UUID, *entity.PartnerAccount, *MRRReportHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doMRR(t *testing.T, h *MRRReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := newMRRRequest(t, appID, pa, query)
	rec := httptest.NewRecorder()
	h.GetMRRReport(rec, req)
	return rec
}

func decodeMRR(t *testing.T, rec *httptest.ResponseRecorder) mrrReport {
	t.Helper()
	var resp mrrReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// snap builds a snapshot with a given date and ActiveMRRCents.
func mrrSnap(appID uuid.UUID, date time.Time, activeMRR int64) *entity.DailyMetricsSnapshot {
	return &entity.DailyMetricsSnapshot{ID: uuid.New(), AppID: appID, Date: date, ActiveMRRCents: activeMRR}
}

// TestMRR_HeadlineFromLatestSnapshot verifies mrrCents is the latest in-range
// snapshot's ActiveMRRCents and equals the last trend point.
func TestMRR_HeadlineFromLatestSnapshot(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		mrrSnap(appID, d, 1180000),
		mrrSnap(appID, d.AddDate(0, 0, 1), 1248000),
	}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doMRR(t, h, appID, pa, "from=2026-06-01&to=2026-07-31")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := decodeMRR(t, rec)
	if resp.MrrCents != 1248000 {
		t.Errorf("mrrCents: expected 1248000, got %d", resp.MrrCents)
	}
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-07-01" || resp.Trend[0].MrrCents != 1180000 {
		t.Errorf("trend[0]: expected {2026-07-01,1180000}, got {%s,%d}", resp.Trend[0].Date, resp.Trend[0].MrrCents)
	}
	if resp.MrrCents != resp.Trend[len(resp.Trend)-1].MrrCents {
		t.Errorf("headline %d != last trend point %d", resp.MrrCents, resp.Trend[len(resp.Trend)-1].MrrCents)
	}
}

// TestMRR_MomChangePositive verifies a doubling of MRR yields momChangePct 1.0
// (a growth ratio, NOT clamped to [0,1]).
func TestMRR_MomChangePositive(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		mrrSnap(appID, d, 500000),
		mrrSnap(appID, d.AddDate(0, 0, 1), 1000000), // doubling → +1.0
	}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, ""))
	if resp.MomChangePct != 1.0 {
		t.Errorf("momChangePct: expected 1.0 (doubling, NOT clamped), got %v", resp.MomChangePct)
	}
}

// TestMRR_MomChangeNegative verifies a halving of MRR yields momChangePct -0.5
// (signed, can be negative — NOT clamped to [0,1]).
func TestMRR_MomChangeNegative(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		mrrSnap(appID, d, 1000000),
		mrrSnap(appID, d.AddDate(0, 0, 1), 500000), // halving → -0.5
	}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, ""))
	if resp.MomChangePct != -0.5 {
		t.Errorf("momChangePct: expected -0.5 (halving, NOT clamped), got %v", resp.MomChangePct)
	}
}

// TestMRR_MomChangeZeroGuards verifies momChangePct is 0 when baseline is 0 or
// there are fewer than 2 snapshots.
func TestMRR_MomChangeZeroGuards(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	// Baseline 0 → divide-by-zero guard.
	zeroBaseline := []*entity.DailyMetricsSnapshot{
		mrrSnap(appID, d, 0),
		mrrSnap(appID, d.AddDate(0, 0, 1), 500000),
	}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: zeroBaseline},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if got := decodeMRR(t, doMRR(t, h, appID, pa, "")).MomChangePct; got != 0 {
		t.Errorf("momChangePct with 0 baseline: expected 0, got %v", got)
	}

	// Single snapshot → fewer than 2.
	single := []*entity.DailyMetricsSnapshot{mrrSnap(appID, d, 500000)}
	h2 := NewMRRReportHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: single},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if got := decodeMRR(t, doMRR(t, h2, appID, pa, "")).MomChangePct; got != 0 {
		t.Errorf("momChangePct with 1 snapshot: expected 0, got %v", got)
	}
}

// TestMRR_NewMrrInRange verifies newMrrCents sums SAFE subs created in [from,to],
// including an annual sub normalized ÷12, and excludes out-of-range CreatedAt and
// non-SAFE subs.
func TestMRR_NewMrrInRange(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	inRange := safeSub(appID, "new.myshopify.com", "Pro", 5000)
	inRange.CreatedAt = from.AddDate(0, 0, 5)

	annual := safeSub(appID, "annual.myshopify.com", "Pro", 1200)
	annual.BillingInterval = valueobject.BillingIntervalAnnual // 1200/12 = 100
	annual.CreatedAt = from.AddDate(0, 0, 10)

	// Created before `from` → excluded.
	old := safeSub(appID, "old.myshopify.com", "Pro", 9999)
	old.CreatedAt = from.AddDate(0, 0, -10)

	// Created in range but not SAFE → excluded.
	risk := atRiskSub(appID, "risk.myshopify.com", "Pro", 8888, valueobject.RiskStateOneCycleMissed)
	risk.CreatedAt = from.AddDate(0, 0, 3)

	subs := []*entity.Subscription{inRange, annual, old, risk}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	// 5000 + 100 (annual/12) = 5100.
	if resp.NewMrrCents != 5100 {
		t.Errorf("newMrrCents: expected 5100, got %d", resp.NewMrrCents)
	}
}

// TestMRR_ChurnedMrrInRange verifies churnedMrrCents sums CHURNED subs whose
// churnedDateOf falls in [from,to] (incl. an annual sub ÷12) and excludes
// out-of-range churn dates. Returned as a positive number.
func TestMRR_ChurnedMrrInRange(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	inRangeChurn := from.AddDate(0, 0, 5)
	churnedIn := atRiskSub(appID, "c1.myshopify.com", "Pro", 4000, valueobject.RiskStateChurned)
	churnedIn.ExpectedNextChargeDate = &inRangeChurn

	annualChurnDate := from.AddDate(0, 0, 8)
	annualChurned := atRiskSub(appID, "annual.myshopify.com", "Pro", 1200, valueobject.RiskStateChurned)
	annualChurned.BillingInterval = valueobject.BillingIntervalAnnual // 1200/12 = 100
	annualChurned.ExpectedNextChargeDate = &annualChurnDate

	// Churn date before `from` → excluded.
	outChurn := from.AddDate(0, 0, -20)
	churnedOut := atRiskSub(appID, "c2.myshopify.com", "Pro", 9999, valueobject.RiskStateChurned)
	churnedOut.ExpectedNextChargeDate = &outChurn

	subs := []*entity.Subscription{churnedIn, annualChurned, churnedOut}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	// 4000 + 100 (annual/12) = 4100, positive.
	if resp.ChurnedMrrCents != 4100 {
		t.Errorf("churnedMrrCents: expected 4100, got %d", resp.ChurnedMrrCents)
	}
}

// TestMRR_MovementDateBoundary pins the [from,to] window edges for New/Churned MRR:
// created/churned exactly at `from` and any time on the `to` day are included; the day
// after `to` is excluded.
func TestMRR_MovementDateBoundary(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	fromMidnight := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	toAfternoon := time.Date(2026, 7, 31, 14, 30, 0, 0, time.UTC)
	dayAfter := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	nAtFrom := safeSub(appID, "n1.myshopify.com", "Pro", 1000)
	nAtFrom.CreatedAt = fromMidnight
	nToDay := safeSub(appID, "n2.myshopify.com", "Pro", 2000)
	nToDay.CreatedAt = toAfternoon
	nAfter := safeSub(appID, "n3.myshopify.com", "Pro", 4000)
	nAfter.CreatedAt = dayAfter

	cAtFrom := atRiskSub(appID, "c1.myshopify.com", "Pro", 500, valueobject.RiskStateChurned)
	cAtFrom.ExpectedNextChargeDate = &fromMidnight
	cAfter := atRiskSub(appID, "c2.myshopify.com", "Pro", 800, valueobject.RiskStateChurned)
	cAfter.ExpectedNextChargeDate = &dayAfter

	subs := []*entity.Subscription{nAtFrom, nToDay, nAfter, cAtFrom, cAfter}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	// New: 1000 (at from) + 2000 (to-day afternoon) = 3000; day-after 4000 excluded.
	if resp.NewMrrCents != 3000 {
		t.Errorf("newMrrCents boundary: expected 3000 (from + to-day, day-after excluded), got %d", resp.NewMrrCents)
	}
	// Churned: 500 (at from); day-after 800 excluded.
	if resp.ChurnedMrrCents != 500 {
		t.Errorf("churnedMrrCents boundary: expected 500 (from included, day-after excluded), got %d", resp.ChurnedMrrCents)
	}
}

// TestMRR_CurrencyNonUSD verifies a non-USD subscription currency surfaces.
func TestMRR_CurrencyNonUSD(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	eur := safeSub(appID, "eu.myshopify.com", "Pro", 5000)
	eur.Currency = "EUR"
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{eur}},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, ""))
	if resp.Currency != "EUR" {
		t.Errorf("currency: expected EUR, got %q", resp.Currency)
	}
}

// TestMRR_EmptyPlanNameBucket verifies a sub with an empty plan name forms its own bucket.
func TestMRR_EmptyPlanNameBucket(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	s := safeSub(appID, "noplan.myshopify.com", "", 4000)
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{s}},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, ""))
	if len(resp.Plans) != 1 || resp.Plans[0].PlanName != "" || resp.Plans[0].MrrCents != 4000 {
		t.Errorf("expected one empty-name plan bucket with 4000, got %+v", resp.Plans)
	}
}

// TestMRR_PlanGrouping verifies per-plan activeSubs, mrrCents, pctOfTotal and sort
// by MRR descending.
func TestMRR_PlanGrouping(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		// Pro: 2 safe (6000+2000=8000) + 1 churned (excluded from MRR).
		safeSub(appID, "p1.myshopify.com", "Pro", 6000),
		safeSub(appID, "p2.myshopify.com", "Pro", 2000),
		atRiskSub(appID, "p3.myshopify.com", "Pro", 1000, valueobject.RiskStateChurned),
		// Basic: 1 safe (2000).
		safeSub(appID, "b1.myshopify.com", "Basic", 2000),
	}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeMRR(t, doMRR(t, h, appID, pa, ""))
	if len(resp.Plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(resp.Plans))
	}
	// Total safe MRR = 8000 + 2000 = 10000.
	pro := resp.Plans[0]
	if pro.PlanName != "Pro" || pro.ActiveSubs != 2 || pro.MrrCents != 8000 {
		t.Errorf("plans[0] Pro: got %+v", pro)
	}
	if pro.PctOfTotal != 0.8 {
		t.Errorf("Pro pctOfTotal: expected 0.8, got %v", pro.PctOfTotal)
	}
	basic := resp.Plans[1]
	if basic.PlanName != "Basic" || basic.ActiveSubs != 1 || basic.MrrCents != 2000 || basic.PctOfTotal != 0.2 {
		t.Errorf("plans[1] Basic: got %+v", basic)
	}
}

// TestMRR_Empty verifies the empty case yields zero metrics and []-serialized slices.
func TestMRR_Empty(t *testing.T) {
	appID, pa, h := mrrFixture(nil, nil)
	rec := doMRR(t, h, appID, pa, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	resp := decodeMRR(t, rec)
	if resp.MrrCents != 0 || resp.MomChangePct != 0 || resp.NewMrrCents != 0 || resp.ChurnedMrrCents != 0 {
		t.Errorf("expected zero metrics, got %+v", resp)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"plans":[]`) {
		t.Errorf("expected plans serialized as [], body: %s", body)
	}
	if !strings.Contains(body, `"trend":[]`) {
		t.Errorf("expected trend serialized as [], body: %s", body)
	}
}

// TestMRR_SubRepoErrorReturns503 verifies subscription-repo failures surface as 503.
func TestMRR_SubRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doMRR(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestMRR_SnapshotRepoErrorReturns503 verifies snapshot-repo failures surface as 503.
func TestMRR_SnapshotRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{safeSub(appID, "a.myshopify.com", "Pro", 5000)}},
		&mockSnapshotRepoForForecast{rangeErr: errors.New("snapshot db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doMRR(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestMRR_CSVFormat verifies CSV output: header + one row per plan (sorted desc),
// with pctOfTotal formatted to 4 decimals. Also exercises plan-name escaping.
func TestMRR_CSVFormat(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		safeSub(appID, "p1.myshopify.com", `Pro, "Annual"`, 9000),
		safeSub(appID, "b1.myshopify.com", "Basic", 1000),
	}
	h := NewMRRReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doMRR(t, h, appID, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "mrr.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"plan", "activeSubs", "mrrCents", "pctOfTotal"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by MRR desc: escaped-name Pro first, stays one column.
	if len(records[1]) != 4 {
		t.Fatalf("expected 4 columns, got %d: %v", len(records[1]), records[1])
	}
	if records[1][0] != `Pro, "Annual"` {
		t.Errorf("plan column: expected %q, got %q", `Pro, "Annual"`, records[1][0])
	}
	// 9000 / 10000 = 0.9, formatted to 4 decimals.
	if records[1][3] != "0.9000" {
		t.Errorf("pctOfTotal format: expected 0.9000, got %q", records[1][3])
	}
}

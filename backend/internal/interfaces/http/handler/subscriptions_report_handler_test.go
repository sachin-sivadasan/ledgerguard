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

// newSubscriptionsRequest builds an authenticated GET request for the Subscriptions report.
func newSubscriptionsRequest(t *testing.T, appID uuid.UUID, pa *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/subscriptions"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func newSubscriptionsHandler(appID uuid.UUID, pa *entity.PartnerAccount, subs []*entity.Subscription, snaps []*entity.DailyMetricsSnapshot) *SubscriptionsReportHandler {
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	return NewSubscriptionsReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
}

func doSubscriptions(t *testing.T, h *SubscriptionsReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.GetSubscriptions(rec, newSubscriptionsRequest(t, appID, pa, query))
	return rec
}

func decodeSubscriptions(t *testing.T, rec *httptest.ResponseRecorder) subscriptionsReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp subscriptionsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// subsSnap builds a snapshot carrying only the TotalSubscriptions denominator used by
// the churn rate (date recent so it lands inside the handler's trailing window).
func subsSnap(appID uuid.UUID, total int) *entity.DailyMetricsSnapshot {
	return &entity.DailyMetricsSnapshot{
		ID:                 uuid.New(),
		AppID:              appID,
		Date:               time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		TotalSubscriptions: total,
	}
}

// TestSubscriptions_ArpuIsActiveMrrOverActiveSubs verifies ARPU = ActiveMRR ÷ active subs.
func TestSubscriptions_ArpuIsActiveMrrOverActiveSubs(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 3000),
		safeSub(appID, "b.myshopify.com", "Pro", 2000),
		safeSub(appID, "c.myshopify.com", "Pro", 1000),
	}
	h := newSubscriptionsHandler(appID, pa, subs, nil)
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.ActiveSubs != 3 || resp.ActiveMrrCents != 6000 {
		t.Fatalf("expected activeSubs 3 / activeMrr 6000, got %d / %d", resp.ActiveSubs, resp.ActiveMrrCents)
	}
	if resp.ArpuCents != 2000 { // 6000 / 3
		t.Errorf("arpuCents: expected 2000, got %d", resp.ArpuCents)
	}
}

// TestSubscriptions_ArpuFlooredIntegerDivision verifies ARPU floors to whole cents.
func TestSubscriptions_ArpuFlooredIntegerDivision(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 4000),
		safeSub(appID, "b.myshopify.com", "Pro", 3000),
		safeSub(appID, "c.myshopify.com", "Pro", 3000),
	}
	h := newSubscriptionsHandler(appID, pa, subs, nil)
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	// 10000 / 3 = 3333.33… → floored to 3333.
	if resp.ArpuCents != 3333 {
		t.Errorf("arpuCents (floored): expected 3333, got %d", resp.ArpuCents)
	}
}

// TestSubscriptions_ArpuFloorsNotRounds pins that ARPU FLOORS (integer division) rather
// than rounding: 10000 ÷ 6 = 1666.67, which floors to 1666 but would round to 1667 — so
// a round-instead-of-floor regression is caught.
func TestSubscriptions_ArpuFloorsNotRounds(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 2000),
		safeSub(appID, "b.myshopify.com", "Pro", 2000),
		safeSub(appID, "c.myshopify.com", "Pro", 2000),
		safeSub(appID, "d.myshopify.com", "Pro", 2000),
		safeSub(appID, "e.myshopify.com", "Pro", 1000),
		safeSub(appID, "f.myshopify.com", "Pro", 1000),
	}
	h := newSubscriptionsHandler(appID, pa, subs, nil)
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	// 10000 / 6 = 1666.67 → floor 1666 (round would give 1667).
	if resp.ArpuCents != 1666 {
		t.Errorf("arpuCents: expected 1666 (floored, not rounded to 1667), got %d", resp.ArpuCents)
	}
}

// TestSubscriptions_LtvIsArpuOverChurnRate verifies LTV = ARPU ÷ churn rate, where the
// churn rate is churned subs ÷ snapshot total (the shared churnRate definition), and
// that the quotient is ROUNDED to the nearest cent (2000/0.2 must be 10000, not 9999).
func TestSubscriptions_LtvIsArpuOverChurnRate(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 3000),
		safeSub(appID, "b.myshopify.com", "Pro", 2000),
		safeSub(appID, "c.myshopify.com", "Pro", 1000),
		atRiskSub(appID, "d.myshopify.com", "Pro", 5000, valueobject.RiskStateChurned),
		atRiskSub(appID, "e.myshopify.com", "Pro", 5000, valueobject.RiskStateChurned),
	}
	// churnedCount 2, snapshot total 10 → rate 0.2. ARPU = 6000/3 = 2000. LTV = 10000.
	h := newSubscriptionsHandler(appID, pa, subs, []*entity.DailyMetricsSnapshot{subsSnap(appID, 10)})
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.ChurnRate != 0.2 {
		t.Fatalf("churnRate: expected 0.2, got %v", resp.ChurnRate)
	}
	if resp.ArpuCents != 2000 {
		t.Fatalf("arpuCents: expected 2000, got %d", resp.ArpuCents)
	}
	if resp.LtvCents != 10000 {
		t.Errorf("ltvCents: expected 10000 (rounded, 2000/0.2), got %d", resp.LtvCents)
	}
}

// TestSubscriptions_LtvRoundedNotTruncated pins that LTV ROUNDS to the nearest cent
// rather than truncating. ARPU 1000 ÷ churn 0.7 = 1428.57… → rounds to 1429; a plain
// int64() truncation (or math.Floor) would give 1428, so this catches that regression.
func TestSubscriptions_LtvRoundedNotTruncated(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		// 1 safe sub @ 1000 → ARPU 1000.
		safeSub(appID, "a.myshopify.com", "Pro", 1000),
	}
	// 7 churned subs, snapshot total 10 → rate 0.7.
	for range 7 {
		subs = append(subs, atRiskSub(appID, "c.myshopify.com", "Pro", 500, valueobject.RiskStateChurned))
	}
	h := newSubscriptionsHandler(appID, pa, subs, []*entity.DailyMetricsSnapshot{subsSnap(appID, 10)})
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.ChurnRate != 0.7 {
		t.Fatalf("churnRate: expected 0.7, got %v", resp.ChurnRate)
	}
	if resp.ArpuCents != 1000 {
		t.Fatalf("arpuCents: expected 1000, got %d", resp.ArpuCents)
	}
	// 1000 / 0.7 = 1428.57… → rounds to 1429 (truncation would give 1428).
	if resp.LtvCents != 1429 {
		t.Errorf("ltvCents: expected 1429 (rounded, not truncated to 1428), got %d", resp.LtvCents)
	}
}

// TestSubscriptions_LtvZeroWhenNoChurn verifies LTV is 0 (undefined) when the churn
// rate is 0 — both when no subs are churned and when there is no snapshot at all.
func TestSubscriptions_LtvZeroWhenNoChurn(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{safeSub(appID, "a.myshopify.com", "Pro", 3000)}

	// No churned subs, snapshot present → rate 0 → LTV 0.
	h := newSubscriptionsHandler(appID, pa, subs, []*entity.DailyMetricsSnapshot{subsSnap(appID, 10)})
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.ChurnRate != 0 || resp.LtvCents != 0 {
		t.Errorf("no churn: expected rate 0 / ltv 0, got %v / %d", resp.ChurnRate, resp.LtvCents)
	}

	// No snapshot at all → denominator 0 → rate 0 → LTV 0.
	churned := atRiskSub(appID, "b.myshopify.com", "Pro", 5000, valueobject.RiskStateChurned)
	h2 := newSubscriptionsHandler(appID, pa, []*entity.Subscription{subs[0], churned}, nil)
	resp2 := decodeSubscriptions(t, doSubscriptions(t, h2, appID, pa, ""))
	if resp2.ChurnRate != 0 || resp2.LtvCents != 0 {
		t.Errorf("no snapshot: expected rate 0 / ltv 0, got %v / %d", resp2.ChurnRate, resp2.LtvCents)
	}
}

// TestSubscriptions_ChurnRateClampedWhenStale verifies the churn rate clamps to 1.0
// when the live churned count exceeds a stale snapshot total (shared churnRate helper),
// so LTV never exceeds ARPU on stale data.
func TestSubscriptions_ChurnRateClampedWhenStale(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 4000),
		atRiskSub(appID, "b.myshopify.com", "Pro", 1000, valueobject.RiskStateChurned),
		atRiskSub(appID, "c.myshopify.com", "Pro", 1000, valueobject.RiskStateChurned),
		atRiskSub(appID, "d.myshopify.com", "Pro", 1000, valueobject.RiskStateChurned),
	}
	// churnedCount 3 > snapshot total 2 → rate clamped to 1.0. ARPU = 4000. LTV = 4000.
	h := newSubscriptionsHandler(appID, pa, subs, []*entity.DailyMetricsSnapshot{subsSnap(appID, 2)})
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.ChurnRate != 1.0 {
		t.Fatalf("churnRate: expected clamp to 1.0, got %v", resp.ChurnRate)
	}
	if resp.LtvCents != resp.ArpuCents {
		t.Errorf("ltv at rate 1.0 should equal arpu %d, got %d", resp.ArpuCents, resp.LtvCents)
	}
}

// TestSubscriptions_ActiveExcludesNonSafe verifies only SAFE subs count toward active
// subs and ActiveMRR; at-risk and churned subs are excluded from both.
func TestSubscriptions_ActiveExcludesNonSafe(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 5000),
		atRiskSub(appID, "b.myshopify.com", "Pro", 9999, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "c.myshopify.com", "Pro", 8888, valueobject.RiskStateChurned),
	}
	h := newSubscriptionsHandler(appID, pa, subs, nil)
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.ActiveSubs != 1 || resp.ActiveMrrCents != 5000 {
		t.Errorf("expected 1 active sub / 5000 MRR (non-SAFE excluded), got %d / %d", resp.ActiveSubs, resp.ActiveMrrCents)
	}
}

// TestSubscriptions_PlanCompositionAndSort verifies per-plan activeSubs, mrr, ARPU,
// LTV (using the app-level churn rate) and pctOfSubs, sorted by active-sub count desc.
func TestSubscriptions_PlanCompositionAndSort(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		// Pro: 2 safe (5000 + 3000 = 8000).
		safeSub(appID, "p1.myshopify.com", "Pro", 5000),
		safeSub(appID, "p2.myshopify.com", "Pro", 3000),
		// Basic: 3 safe (1000 each = 3000).
		safeSub(appID, "b1.myshopify.com", "Basic", 1000),
		safeSub(appID, "b2.myshopify.com", "Basic", 1000),
		safeSub(appID, "b3.myshopify.com", "Basic", 1000),
		// 2 churned (any plan) drive churnedCount = 2.
		atRiskSub(appID, "x1.myshopify.com", "Pro", 9999, valueobject.RiskStateChurned),
		atRiskSub(appID, "x2.myshopify.com", "Basic", 9999, valueobject.RiskStateChurned),
	}
	// churnedCount 2, snapshot total 20 → rate 0.1.
	h := newSubscriptionsHandler(appID, pa, subs, []*entity.DailyMetricsSnapshot{subsSnap(appID, 20)})
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))

	if resp.ActiveSubs != 5 || resp.ActiveMrrCents != 11000 {
		t.Fatalf("expected 5 active / 11000 MRR, got %d / %d", resp.ActiveSubs, resp.ActiveMrrCents)
	}
	if len(resp.Plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(resp.Plans))
	}
	// Sorted by activeSubs desc: Basic (3) first, then Pro (2).
	basic := resp.Plans[0]
	if basic.PlanName != "Basic" || basic.ActiveSubs != 3 || basic.MrrCents != 3000 {
		t.Fatalf("plans[0] Basic: got %+v", basic)
	}
	if basic.ArpuCents != 1000 || basic.PctOfSubs != 0.6 { // 3000/3, 3/5
		t.Errorf("Basic arpu/pct: expected 1000 / 0.6, got %d / %v", basic.ArpuCents, basic.PctOfSubs)
	}
	if basic.LtvCents != 10000 { // 1000 / 0.1
		t.Errorf("Basic ltv: expected 10000, got %d", basic.LtvCents)
	}
	pro := resp.Plans[1]
	if pro.PlanName != "Pro" || pro.ActiveSubs != 2 || pro.MrrCents != 8000 {
		t.Fatalf("plans[1] Pro: got %+v", pro)
	}
	if pro.ArpuCents != 4000 || pro.PctOfSubs != 0.4 { // 8000/2, 2/5
		t.Errorf("Pro arpu/pct: expected 4000 / 0.4, got %d / %v", pro.ArpuCents, pro.PctOfSubs)
	}
	if pro.LtvCents != 40000 { // 4000 / 0.1
		t.Errorf("Pro ltv: expected 40000, got %d", pro.LtvCents)
	}
}

// TestSubscriptions_EmptyPlanNameBucket verifies a sub with an empty plan name forms its
// own bucket (labeled by its price tier, not dropped) — the Partner API gives no plan
// name, so the report falls back to a "$40.00/mo" pseudo-label rather than a blank row.
func TestSubscriptions_EmptyPlanNameBucket(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	s := safeSub(appID, "noplan.myshopify.com", "", 4000)
	h := newSubscriptionsHandler(appID, pa, []*entity.Subscription{s}, nil)
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if len(resp.Plans) != 1 || resp.Plans[0].PlanName != "$40.00/mo" || resp.Plans[0].ActiveSubs != 1 {
		t.Errorf("expected one price-tier bucket ($40.00/mo) with 1 sub, got %+v", resp.Plans)
	}
}

// TestSubscriptions_CurrencyNonUSD verifies a non-USD subscription currency surfaces.
func TestSubscriptions_CurrencyNonUSD(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	eur := safeSub(appID, "eu.myshopify.com", "Pro", 5000)
	eur.Currency = "EUR"
	h := newSubscriptionsHandler(appID, pa, []*entity.Subscription{eur}, nil)
	resp := decodeSubscriptions(t, doSubscriptions(t, h, appID, pa, ""))
	if resp.Currency != "EUR" {
		t.Errorf("currency: expected EUR, got %q", resp.Currency)
	}
}

// TestSubscriptions_Empty verifies the empty case yields zero metrics and a
// []-serialized plans slice (never null).
func TestSubscriptions_Empty(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newSubscriptionsHandler(appID, pa, nil, nil)
	rec := doSubscriptions(t, h, appID, pa, "")
	resp := decodeSubscriptions(t, rec)
	if resp.ActiveSubs != 0 || resp.ArpuCents != 0 || resp.LtvCents != 0 || resp.ChurnRate != 0 {
		t.Errorf("expected zero metrics, got %+v", resp)
	}
	if !strings.Contains(rec.Body.String(), `"plans":[]`) {
		t.Errorf("expected plans serialized as [], body: %s", rec.Body.String())
	}
}

// TestSubscriptions_SubRepoErrorReturns503 verifies subscription-repo failures surface as 503.
func TestSubscriptions_SubRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewSubscriptionsReportHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doSubscriptions(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSubscriptions_SnapshotRepoErrorReturns503 verifies snapshot-repo failures surface as 503.
func TestSubscriptions_SnapshotRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewSubscriptionsReportHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{safeSub(appID, "a.myshopify.com", "Pro", 5000)}},
		&mockSnapshotRepoForForecast{rangeErr: errors.New("snapshot db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doSubscriptions(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSubscriptions_CSVFormat verifies CSV output: header + one row per plan (sorted by
// active subs desc), pctOfSubs formatted to 4 decimals, and plan-name escaping.
func TestSubscriptions_CSVFormat(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	subs := []*entity.Subscription{
		safeSub(appID, "p1.myshopify.com", `Pro, "Annual"`, 5000),
		safeSub(appID, "p2.myshopify.com", `Pro, "Annual"`, 3000),
		safeSub(appID, "b1.myshopify.com", "Basic", 1000),
	}
	h := newSubscriptionsHandler(appID, pa, subs, nil)
	rec := doSubscriptions(t, h, appID, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "subscriptions.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"plan", "activeSubs", "mrrCents", "arpuCents", "ltvCents", "pctOfSubs"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by activeSubs desc: escaped-name Pro (2 subs) first, stays one column.
	if len(records[1]) != 6 {
		t.Fatalf("expected 6 columns, got %d: %v", len(records[1]), records[1])
	}
	if records[1][0] != `Pro, "Annual"` {
		t.Errorf("plan column: expected %q, got %q", `Pro, "Annual"`, records[1][0])
	}
	// Pro pctOfSubs = 2/3 = 0.6667 (4 decimals).
	if records[1][5] != "0.6667" {
		t.Errorf("pctOfSubs format: expected 0.6667, got %q", records[1][5])
	}
}

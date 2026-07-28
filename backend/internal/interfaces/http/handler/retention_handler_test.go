package handler

import (
	"context"
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
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// mockAppEventRepo implements repository.AppEventRepository for retention tests.
type mockAppEventRepo struct {
	events []*entity.AppEvent
	err    error
}

func (m *mockAppEventRepo) UpsertBatch(ctx context.Context, events []*entity.AppEvent) error {
	return nil
}
func (m *mockAppEventRepo) FindByAppAndShop(ctx context.Context, appID uuid.UUID, shopGID string) ([]*entity.AppEvent, error) {
	return nil, nil
}
func (m *mockAppEventRepo) FindByAppID(ctx context.Context, appID uuid.UUID) ([]*entity.AppEvent, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.events, nil
}
func (m *mockAppEventRepo) FindByAppIDPaginated(ctx context.Context, appID uuid.UUID, filters repository.EventFilters) (*repository.EventPage, error) {
	return nil, nil
}

// safeSub builds a SAFE (retained) subscription for retention tests.
func safeSub(appID uuid.UUID, domain, plan string, cents int64) *entity.Subscription {
	return atRiskSub(appID, domain, plan, cents, valueobject.RiskStateSafe)
}

func newRetentionRequest(t *testing.T, appID uuid.UUID, partnerAccount *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/retention"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func retentionFixture(subs []*entity.Subscription, snaps []*entity.DailyMetricsSnapshot, events []*entity.AppEvent) (uuid.UUID, *entity.PartnerAccount, *RetentionHandler) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)
	return appID, partnerAccount, h
}

func doRetention(t *testing.T, h *RetentionHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := newRetentionRequest(t, appID, pa, query)
	rec := httptest.NewRecorder()
	h.GetRetention(rec, req)
	return rec
}

// TestRetention_RenewalRateFromLatestSnapshot verifies the headline renewalRate is
// the latest snapshot's RenewalSuccessRate and equals the last trend point.
func TestRetention_RenewalRateFromLatestSnapshot(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, RenewalSuccessRate: 0.90},
		{ID: uuid.New(), AppID: appID, Date: d.AddDate(0, 0, 1), RenewalSuccessRate: 0.92},
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	// ≤31-day window → daily granularity, so each snapshot is its own trend point.
	rec := doRetention(t, h, appID, pa, "from=2026-07-01&to=2026-07-15")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.RenewalRate != 0.92 {
		t.Errorf("renewalRate: expected 0.92, got %v", resp.RenewalRate)
	}
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-07-01" || resp.Trend[0].RenewalRate != 0.90 {
		t.Errorf("trend[0]: expected {2026-07-01,0.9}, got {%s,%v}", resp.Trend[0].Date, resp.Trend[0].RenewalRate)
	}
	// Headline equals the last trend point.
	if resp.RenewalRate != resp.Trend[len(resp.Trend)-1].RenewalRate {
		t.Errorf("headline %v != last trend point %v", resp.RenewalRate, resp.Trend[len(resp.Trend)-1].RenewalRate)
	}
}

// TestRetention_RenewalRateClamped verifies rates outside [0,1] are clamped.
func TestRetention_RenewalRateClamped(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, RenewalSuccessRate: -0.5},
		{ID: uuid.New(), AppID: appID, Date: d.AddDate(0, 0, 1), RenewalSuccessRate: 1.7},
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Latest is 1.7 → clamped to 1.
	if resp.RenewalRate != 1 {
		t.Errorf("renewalRate: expected 1 (clamped), got %v", resp.RenewalRate)
	}
	if resp.Trend[0].RenewalRate != 0 {
		t.Errorf("trend[0]: expected 0 (clamped from -0.5), got %v", resp.Trend[0].RenewalRate)
	}
	if resp.Trend[1].RenewalRate != 1 {
		t.Errorf("trend[1]: expected 1 (clamped from 1.7), got %v", resp.Trend[1].RenewalRate)
	}
}

// TestRetention_NoSnapshotZeroRate verifies renewalRate is 0 when no snapshot in range.
func TestRetention_NoSnapshotZeroRate(t *testing.T) {
	appID, pa, h := retentionFixture(
		[]*entity.Subscription{safeSub(uuid.New(), "a.myshopify.com", "Pro", 5000)},
		nil, nil,
	)
	rec := doRetention(t, h, appID, pa, "")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.RenewalRate != 0 {
		t.Errorf("renewalRate: expected 0 (no snapshot), got %v", resp.RenewalRate)
	}
	if len(resp.Trend) != 0 {
		t.Errorf("expected empty trend, got %d", len(resp.Trend))
	}
}

// TestRetention_RetainedMrrSafeOnly verifies retained MRR sums only SAFE subs,
// including an annual sub normalized ÷12, and excludes at-risk/churned.
func TestRetention_RetainedMrrSafeOnly(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	annual := safeSub(appID, "annual.myshopify.com", "Pro", 1200)
	annual.BillingInterval = valueobject.BillingIntervalAnnual // 1200/12 = 100
	subs := []*entity.Subscription{
		safeSub(appID, "a.myshopify.com", "Pro", 5000),
		annual,
		atRiskSub(appID, "risk.myshopify.com", "Pro", 9999, valueobject.RiskStateOneCycleMissed),
		atRiskSub(appID, "churned.myshopify.com", "Pro", 8888, valueobject.RiskStateChurned),
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// 5000 + 100 (annual/12) = 5100. At-risk/churned excluded.
	if resp.RetainedMrrCents != 5100 {
		t.Errorf("retainedMrrCents: expected 5100, got %d", resp.RetainedMrrCents)
	}
}

// TestRetention_Reactivations verifies distinct-shop reactivation counting: matches
// REACTIVAT* case-insensitively, dedupes by shop, and excludes out-of-range and
// non-reactivation events.
func TestRetention_Reactivations(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	inRange := from.AddDate(0, 0, 5)
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "REACTIVATED", OccurredAt: inRange},
		// Same shop again in range → still 1 distinct.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "reactivation", OccurredAt: inRange.AddDate(0, 0, 1)},
		// Different shop, case-insensitive match.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "shop_reactivated", OccurredAt: inRange},
		// Out of range (before from) → excluded.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "REACTIVATED", OccurredAt: from.AddDate(0, 0, -5)},
		// Wrong event type → excluded.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/4", EventType: "INSTALLED", OccurredAt: inRange},
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "from=2026-07-01&to=2026-07-31")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// shop/1 (deduped) + shop/2 = 2 distinct reactivations.
	if resp.Reactivations != 2 {
		t.Errorf("reactivations: expected 2, got %d", resp.Reactivations)
	}
}

// TestRetention_ReactivationsBoundary verifies the date-range edges: an event at
// exactly `from` (midnight) is included, an event any time on the `to` day is
// included (the whole day, not just midnight), and an event on the day after `to`
// is excluded. Also proves a shop with both a reactivation and a non-reactivation
// event still counts once.
func TestRetention_ReactivationsBoundary(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	fromMidnight := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	toDayAfternoon := time.Date(2026, 7, 31, 14, 30, 0, 0, time.UTC) // same day as to=2026-07-31
	dayAfterTo := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		// Exactly at `from` midnight → included.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "REACTIVATED", OccurredAt: fromMidnight},
		// Afternoon of the `to` day → must be included (the bug this guards against).
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "REACTIVATED", OccurredAt: toDayAfternoon},
		// shop/2 also has a non-reactivation event → still counts once.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "INSTALLED", OccurredAt: toDayAfternoon},
		// Day after `to` → excluded.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "REACTIVATED", OccurredAt: dayAfterTo},
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "from=2026-07-01&to=2026-07-31")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// shop/1 (at `from`) + shop/2 (on `to` day) = 2; shop/3 (day after) excluded.
	if resp.Reactivations != 2 {
		t.Errorf("reactivations: expected 2 (from-edge + to-day included, day-after excluded), got %d", resp.Reactivations)
	}
}

// TestRetention_CurrencyNonUSD verifies a non-USD subscription currency surfaces.
func TestRetention_CurrencyNonUSD(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	eur := safeSub(appID, "eu.myshopify.com", "Pro", 5000)
	eur.Currency = "EUR"
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{eur}},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doRetention(t, h, appID, pa, "")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Currency != "EUR" {
		t.Errorf("currency: expected EUR, got %q", resp.Currency)
	}
}

// TestRetention_EmptyPlanNameBucket verifies subs with an empty plan name form their
// own bucket rather than being dropped.
func TestRetention_EmptyPlanNameBucket(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		safeSub(appID, "noplan.myshopify.com", "", 4000),
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doRetention(t, h, appID, pa, "")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Plans) != 1 || resp.Plans[0].PlanName != "" || resp.Plans[0].RetainedMrrCents != 4000 {
		t.Errorf("expected one empty-name plan bucket with 4000 retained, got %+v", resp.Plans)
	}
}

// TestRetention_PlanGrouping verifies per-plan activeSubs, renewalRate and retained
// MRR, sorted by retained MRR descending.
func TestRetention_PlanGrouping(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		// Pro: 2 safe (5000+3000) + 1 churned → 2/3 renewal, retained 8000.
		safeSub(appID, "p1.myshopify.com", "Pro", 5000),
		safeSub(appID, "p2.myshopify.com", "Pro", 3000),
		atRiskSub(appID, "p3.myshopify.com", "Pro", 1000, valueobject.RiskStateChurned),
		// Basic: 1 safe (2000) of 1 → 1.0 renewal, retained 2000.
		safeSub(appID, "b1.myshopify.com", "Basic", 2000),
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "")
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(resp.Plans))
	}
	// Sorted by retained MRR desc: Pro (8000) first, Basic (2000) second.
	pro := resp.Plans[0]
	if pro.PlanName != "Pro" || pro.ActiveSubs != 2 || pro.RetainedMrrCents != 8000 {
		t.Errorf("plans[0] Pro: got %+v", pro)
	}
	// 2 safe / 3 total ≈ 0.6667.
	if pro.RenewalRate < 0.66 || pro.RenewalRate > 0.67 {
		t.Errorf("Pro renewalRate: expected ~0.6667, got %v", pro.RenewalRate)
	}
	basic := resp.Plans[1]
	if basic.PlanName != "Basic" || basic.ActiveSubs != 1 || basic.RetainedMrrCents != 2000 || basic.RenewalRate != 1 {
		t.Errorf("plans[1] Basic: got %+v", basic)
	}
}

// TestRetention_Empty verifies the empty case yields zero metrics and []-serialized
// slices (non-nil).
func TestRetention_Empty(t *testing.T) {
	appID, pa, h := retentionFixture(nil, nil, nil)
	rec := doRetention(t, h, appID, pa, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp retentionReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.RenewalRate != 0 || resp.RetainedMrrCents != 0 || resp.Reactivations != 0 {
		t.Errorf("expected zero metrics, got rate=%v mrr=%d react=%d",
			resp.RenewalRate, resp.RetainedMrrCents, resp.Reactivations)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"plans":[]`) {
		t.Errorf("expected plans serialized as [], body: %s", body)
	}
	if !strings.Contains(body, `"trend":[]`) {
		t.Errorf("expected trend serialized as [], body: %s", body)
	}
}

// TestRetention_SubRepoErrorReturns503 verifies subscription-repo failures surface as 503.
func TestRetention_SubRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doRetention(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestRetention_SnapshotRepoErrorReturns503 verifies snapshot-repo failures surface as 503.
func TestRetention_SnapshotRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{safeSub(appID, "a.myshopify.com", "Pro", 5000)}},
		&mockSnapshotRepoForForecast{rangeErr: errors.New("snapshot db down")},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doRetention(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestRetention_EventRepoErrorReturns503 verifies event-repo failures surface as 503.
func TestRetention_EventRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{safeSub(appID, "a.myshopify.com", "Pro", 5000)}},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{err: errors.New("events db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doRetention(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestRetention_CSVFormat verifies CSV output: header + one row per plan (sorted),
// with the renewalRate formatted to 4 decimals.
func TestRetention_CSVFormat(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		safeSub(appID, "p1.myshopify.com", "Pro", 9000),
		safeSub(appID, "b1.myshopify.com", "Basic", 1000),
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "retention.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"plan", "activeSubs", "renewalRate", "retainedMrrCents"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by retained MRR desc: Pro first.
	if records[1][0] != "Pro" {
		t.Errorf("expected first row Pro, got %s", records[1][0])
	}
	// renewalRate formatted to 4 decimals.
	if records[1][2] != "1.0000" {
		t.Errorf("renewalRate format: expected 1.0000, got %q", records[1][2])
	}
}

// TestRetention_CSVEscaping verifies plan names with commas/quotes stay one column.
func TestRetention_CSVEscaping(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		safeSub(appID, "acme.myshopify.com", `Pro, "Annual"`, 5000),
	}
	h := NewRetentionHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	rec := doRetention(t, h, appID, pa, "format=csv")
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d: %v", len(records), records)
	}
	if len(records[1]) != 4 {
		t.Fatalf("expected 4 columns, got %d: %v", len(records[1]), records[1])
	}
	if records[1][0] != `Pro, "Annual"` {
		t.Errorf("plan column: expected %q, got %q", `Pro, "Annual"`, records[1][0])
	}
}

package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// nnNewSub builds a SAFE subscription created at `createdAt` (the "new" signal).
func nnNewSub(appID uuid.UUID, domain, plan string, cents int64, createdAt time.Time) *entity.Subscription {
	s := safeSub(appID, domain, plan, cents)
	s.CreatedAt = createdAt
	return s
}

// nnChurnedSub builds a CHURNED subscription with a given created + churn (effective) date.
func nnChurnedSub(appID uuid.UUID, domain string, createdAt, churnedAt time.Time) *entity.Subscription {
	s := atRiskSub(appID, domain, "Pro", 1000, valueobject.RiskStateChurned)
	s.CreatedAt = createdAt
	s.ExpectedNextChargeDate = &churnedAt // churnedDateOf prefers ExpectedNextChargeDate
	return s
}

func nnFixture(subs []*entity.Subscription) (uuid.UUID, *entity.PartnerAccount, *NetNewSubsReportHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewNetNewSubsReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doNetNew(t *testing.T, h *NetNewSubsReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/net-new-subscriptions"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.GetNetNewSubs(rec, req)
	return rec
}

func decodeNetNew(t *testing.T, rec *httptest.ResponseRecorder) netNewSubsReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp netNewSubsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

var (
	nnJul5  = time.Date(2026, 7, 5, 9, 0, 0, 0, time.UTC)
	nnJul10 = time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	nnJul12 = time.Date(2026, 7, 12, 9, 0, 0, 0, time.UTC)
	nnJul15 = time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	nnJul20 = time.Date(2026, 7, 20, 9, 0, 0, 0, time.UTC)
	nnJul25 = time.Date(2026, 7, 25, 9, 0, 0, 0, time.UTC)
	nnJan   = time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC) // before range
	nnJun   = time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC) // before range
	nnRange = "from=2026-07-01&to=2026-07-31"
)

// TestNetNewSubs_KPIsAndNet verifies new (created in range), churned (churnedDateOf in
// range), and net = new − churned, with out-of-range starts/churns excluded.
func TestNetNewSubs_KPIsAndNet(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		nnNewSub(appID, "a.myshopify.com", "Pro", 4900, nnJul5),
		nnNewSub(appID, "b.myshopify.com", "Pro", 3900, nnJul10),
		nnNewSub(appID, "c.myshopify.com", "Pro", 1900, nnJul15),
		nnChurnedSub(appID, "d.myshopify.com", nnJan, nnJul20),   // churned in range
		nnNewSub(appID, "old.myshopify.com", "Pro", 5000, nnJun), // created before range → not new
		nnChurnedSub(appID, "e.myshopify.com", nnJan, nnJun),     // churned before range → not churned
	}
	_, pa, h := nnFixture(subs)
	resp := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange))
	if resp.NewSubs != 3 {
		t.Errorf("newSubs: expected 3, got %d", resp.NewSubs)
	}
	if resp.Churned != 1 {
		t.Errorf("churned: expected 1, got %d", resp.Churned)
	}
	if resp.Net != 2 {
		t.Errorf("net: expected +2, got %d", resp.Net)
	}
}

// TestNetNewSubs_StartAndChurnSamePeriodCountsBoth verifies a sub that both starts and
// churns in the window counts in BOTH new and churned (net contribution 0).
func TestNetNewSubs_StartAndChurnSamePeriodCountsBoth(t *testing.T) {
	appID := uuid.New()
	s := nnChurnedSub(appID, "x.myshopify.com", nnJul5, nnJul25) // created AND churned in range
	_, pa, h := nnFixture([]*entity.Subscription{s})
	resp := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange))
	if resp.NewSubs != 1 || resp.Churned != 1 || resp.Net != 0 {
		t.Errorf("start+churn same period: expected new 1 / churned 1 / net 0, got %+v", resp)
	}
}

// TestNetNewSubs_DateBoundary pins the [from,to] window: created at `from` and any time
// on the `to` day are included; the day after `to` is excluded.
func TestNetNewSubs_DateBoundary(t *testing.T) {
	appID := uuid.New()
	fromMidnight := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	toAfternoon := time.Date(2026, 7, 31, 14, 30, 0, 0, time.UTC)
	dayAfter := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{
		nnNewSub(appID, "a.myshopify.com", "Pro", 1000, fromMidnight),
		nnNewSub(appID, "b.myshopify.com", "Pro", 2000, toAfternoon),
		nnNewSub(appID, "c.myshopify.com", "Pro", 3000, dayAfter), // excluded
	}
	_, pa, h := nnFixture(subs)
	resp := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange))
	if resp.NewSubs != 2 {
		t.Errorf("newSubs boundary: expected 2 (from + to-day, day-after excluded), got %d", resp.NewSubs)
	}
}

// TestNetNewSubs_DailyTrend verifies per-day new/churned/net buckets, ascending, and a
// churn-only day (new=0).
func TestNetNewSubs_DailyTrend(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		nnNewSub(appID, "a.myshopify.com", "Pro", 1000, nnJul12), // later day, added first
		nnNewSub(appID, "b.myshopify.com", "Pro", 1000, nnJul10),
		nnNewSub(appID, "c.myshopify.com", "Pro", 1000, nnJul10),
		nnChurnedSub(appID, "d.myshopify.com", nnJan, nnJul10), // churn on day1 (created out of range)
	}
	_, pa, h := nnFixture(subs)
	resp := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange))
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend days, got %d: %+v", len(resp.Trend), resp.Trend)
	}
	if resp.Trend[0].Date != "2026-07-10" || resp.Trend[0].New != 2 || resp.Trend[0].Churned != 1 || resp.Trend[0].Net != 1 {
		t.Errorf("trend[0]: expected {2026-07-10, new2, churned1, net1}, got %+v", resp.Trend[0])
	}
	if resp.Trend[1].Date != "2026-07-12" || resp.Trend[1].New != 1 || resp.Trend[1].Churned != 0 || resp.Trend[1].Net != 1 {
		t.Errorf("trend[1]: expected {2026-07-12, new1, churned0, net1}, got %+v", resp.Trend[1])
	}
}

// TestNetNewSubs_RecentNewestFirstAndFields verifies the recent-new-subs table is
// newest-first (by CreatedAt) and carries domain/plan/MRR/started.
func TestNetNewSubs_RecentNewestFirstAndFields(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		nnNewSub(appID, "older.myshopify.com", "Starter", 1900, nnJul5),
		nnNewSub(appID, "newer.myshopify.com", "Pro", 4900, nnJul20),
	}
	_, pa, h := nnFixture(subs)
	resp := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange))
	if len(resp.NewStores) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(resp.NewStores))
	}
	first := resp.NewStores[0]
	if first.Domain != "newer.myshopify.com" || first.PlanName != "Pro" || first.MrrCents != 4900 || first.Started != "2026-07-20" {
		t.Errorf("row[0]: expected newest {newer.myshopify.com, Pro, 4900, 2026-07-20}, got %+v", first)
	}
	if resp.NewStores[1].Domain != "older.myshopify.com" {
		t.Errorf("row[1]: expected older sub second, got %q", resp.NewStores[1].Domain)
	}
}

// TestNetNewSubs_RecentCapped verifies the table caps at recentNewSubsLimit (keeping the
// newest) while NewSubs counts every in-range sub.
func TestNetNewSubs_RecentCapped(t *testing.T) {
	appID := uuid.New()
	n := recentNewSubsLimit + 5
	base := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	subs := make([]*entity.Subscription, 0, n)
	for i := range n {
		// i increases with CreatedAt → shop/(n-1) is the newest.
		subs = append(subs, nnNewSub(appID, fmt.Sprintf("shop%d.myshopify.com", i), "Pro", 1000, base.Add(time.Duration(i)*time.Hour)))
	}
	_, pa, h := nnFixture(subs)
	resp := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange))
	if resp.NewSubs != n {
		t.Errorf("newSubs: expected %d (all counted), got %d", n, resp.NewSubs)
	}
	if len(resp.NewStores) != recentNewSubsLimit {
		t.Fatalf("newStores: expected capped at %d, got %d", recentNewSubsLimit, len(resp.NewStores))
	}
	if resp.NewStores[0].Domain != fmt.Sprintf("shop%d.myshopify.com", n-1) {
		t.Errorf("newStores[0]: expected newest shop%d, got %q", n-1, resp.NewStores[0].Domain)
	}
	shown := make(map[string]bool, len(resp.NewStores))
	for _, r := range resp.NewStores {
		shown[r.Domain] = true
	}
	for i := range 5 { // oldest 5 dropped
		if shown[fmt.Sprintf("shop%d.myshopify.com", i)] {
			t.Errorf("expected oldest shop%d to be dropped by the cap", i)
		}
	}
}

// TestNetNewSubs_CurrencyFirstNonEmpty verifies the currency is the FIRST non-empty (the
// break-loop avoids the "USD"-sentinel-collision bug) and defaults to USD.
func TestNetNewSubs_CurrencyFirstNonEmpty(t *testing.T) {
	appID := uuid.New()
	eur := nnNewSub(appID, "a.myshopify.com", "Pro", 1000, nnJul5)
	eur.Currency = "EUR"
	usd := nnNewSub(appID, "b.myshopify.com", "Pro", 1000, nnJul10) // Currency "USD"
	_, pa, h := nnFixture([]*entity.Subscription{eur, usd})
	if got := decodeNetNew(t, doNetNew(t, h, appID, pa, nnRange)).Currency; got != "EUR" {
		t.Errorf("currency: expected EUR (first non-empty, not overwritten by later USD), got %q", got)
	}

	// Default USD when none carry a currency.
	empty := nnNewSub(appID, "c.myshopify.com", "Pro", 1000, nnJul5)
	empty.Currency = ""
	_, pa2, h2 := nnFixture([]*entity.Subscription{empty})
	if got := decodeNetNew(t, doNetNew(t, h2, appID, pa2, nnRange)).Currency; got != "USD" {
		t.Errorf("currency: expected USD default, got %q", got)
	}
}

// TestNetNewSubs_Empty verifies the empty case yields zeros and []-serialized slices.
func TestNetNewSubs_Empty(t *testing.T) {
	appID, pa, h := nnFixture(nil)
	rec := doNetNew(t, h, appID, pa, "")
	resp := decodeNetNew(t, rec)
	if resp.NewSubs != 0 || resp.Churned != 0 || resp.Net != 0 {
		t.Errorf("expected zero KPIs, got %+v", resp)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"trend":[]`) || !strings.Contains(body, `"newStores":[]`) {
		t.Errorf("expected trend and newStores serialized as [], body: %s", body)
	}
}

// TestNetNewSubs_RepoErrorReturns503 verifies subscription-repo failures surface as 503.
func TestNetNewSubs_RepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewNetNewSubsReportHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if rec := doNetNew(t, h, appID, pa, ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

// TestNetNewSubs_Unauthenticated verifies a missing user yields 401.
func TestNetNewSubs_Unauthenticated(t *testing.T) {
	appID, _, h := nnFixture(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/reports/net-new-subscriptions", nil)
	req = withURLParam(req, "appID", appID.String())
	rec := httptest.NewRecorder()
	h.GetNetNewSubs(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestNetNewSubs_CSVFormat verifies CSV output: header + one row per new sub (newest
// first) with comma-safe escaping.
func TestNetNewSubs_CSVFormat(t *testing.T) {
	appID := uuid.New()
	sub := nnNewSub(appID, "p1.myshopify.com", `Pro, "Annual"`, 4900, nnJul20)
	_, pa, h := nnFixture([]*entity.Subscription{sub})
	rec := doNetNew(t, h, appID, pa, nnRange+"&format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "net-new-subscriptions.csv") {
		t.Errorf("Content-Disposition = %q, want net-new-subscriptions.csv", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d: %v", len(records), records)
	}
	wantHeader := []string{"domain", "shopName", "plan", "mrrCents", "started"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], want)
		}
	}
	if records[1][0] != "p1.myshopify.com" || records[1][2] != `Pro, "Annual"` || records[1][3] != "4900" || records[1][4] != "2026-07-20" {
		t.Errorf("row = %v, want {p1.myshopify.com, ..., Pro-Annual, 4900, 2026-07-20}", records[1])
	}
}

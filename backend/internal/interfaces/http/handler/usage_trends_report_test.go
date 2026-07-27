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

// Fixed ISO-week Mondays used across these tests so bucketing is deterministic.
var (
	utWeek1Mon = time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)  // Monday
	utWeek2Mon = time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC) // Monday
)

// usageTrendsTx builds a USAGE Transaction for usage-trends tests. GrossAmountCents is
// set distinct from net (net+777) so any net→gross regression in the sums is caught —
// the report must use AmountCents() (net) everywhere.
func usageTrendsTx(appID uuid.UUID, domain, shopName string, netCents int64, txDate time.Time) *entity.Transaction {
	return &entity.Transaction{
		ID:               uuid.New(),
		AppID:            appID,
		ShopifyGID:       "gid://shopify/AppTransaction/" + uuid.New().String(),
		MyshopifyDomain:  domain,
		ShopName:         shopName,
		ChargeType:       valueobject.ChargeTypeUsage,
		GrossAmountCents: netCents + 777, // deliberately != net
		NetAmountCents:   netCents,
		Currency:         "USD",
		TransactionDate:  txDate,
		CreatedAt:        time.Now(),
	}
}

func newUsageTrendsRequest(t *testing.T, appID uuid.UUID, pa *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/usage-trends"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func newUsageTrendsHandler(appID uuid.UUID, pa *entity.PartnerAccount, txRepo *mockTxRepo) *UsageTrendsReportHandler {
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	return NewUsageTrendsReportHandler(
		txRepo,
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
}

func doUsageTrends(t *testing.T, h *UsageTrendsReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := newUsageTrendsRequest(t, appID, pa, query)
	rec := httptest.NewRecorder()
	h.GetUsageTrends(rec, req)
	return rec
}

func decodeUsageTrends(t *testing.T, rec *httptest.ResponseRecorder) usageTrendsReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp usageTrendsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// TestUsageTrends_OnlyUsageCounted verifies non-USAGE charge types (RECURRING/ONE_TIME/
// REFUND) are excluded from the headline total, the weekly trend, and the store rows —
// usage momentum stays strictly separated from MRR.
func TestUsageTrends_OnlyUsageCounted(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	usage := usageTrendsTx(appID, "a.myshopify.com", "Acme", 30000, utWeek1Mon)
	recurring := usageTrendsTx(appID, "a.myshopify.com", "Acme", 99999, utWeek1Mon)
	recurring.ChargeType = valueobject.ChargeTypeRecurring
	oneTime := usageTrendsTx(appID, "b.myshopify.com", "Beta", 88888, utWeek1Mon)
	oneTime.ChargeType = valueobject.ChargeTypeOneTime
	refund := usageTrendsTx(appID, "c.myshopify.com", "Gamma", 5555, utWeek1Mon)
	refund.ChargeType = valueobject.ChargeTypeRefund

	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: []*entity.Transaction{usage, recurring, oneTime, refund}})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))

	if resp.UsageMrrEquivCents != 30000 {
		t.Errorf("usageMrrEquivCents: expected 30000 (USAGE only), got %d", resp.UsageMrrEquivCents)
	}
	if resp.ActiveStores != 1 {
		t.Errorf("activeStores: expected 1 (only the USAGE store), got %d", resp.ActiveStores)
	}
	if len(resp.Stores) != 1 || resp.Stores[0].Domain != "a.myshopify.com" {
		t.Errorf("expected only the USAGE store, got %+v", resp.Stores)
	}
	if len(resp.WeeklyTrend) != 1 || resp.WeeklyTrend[0].UsageCents != 30000 {
		t.Errorf("weeklyTrend: expected one bucket of 30000, got %+v", resp.WeeklyTrend)
	}
}

// TestUsageTrends_WeeklyBucketing verifies txs are grouped into the correct ISO-week
// buckets (keyed by the Monday of the tx's week) with per-week sums, ascending by
// weekStart.
func TestUsageTrends_WeeklyBucketing(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	txs := []*entity.Transaction{
		// Week 1 (Mon 2026-07-06 .. Sun 2026-07-12): 10000 + 4000 = 14000.
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 10000, utWeek1Mon),
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 4000, time.Date(2026, 7, 12, 23, 0, 0, 0, time.UTC)), // Sunday of week1
		// Week 2 (Mon 2026-07-13 ..): 24000.
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 24000, utWeek2Mon),
	}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: txs})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))

	if len(resp.WeeklyTrend) != 2 {
		t.Fatalf("expected 2 weekly buckets, got %d: %+v", len(resp.WeeklyTrend), resp.WeeklyTrend)
	}
	if resp.WeeklyTrend[0].WeekStart != "2026-07-06" || resp.WeeklyTrend[0].UsageCents != 14000 {
		t.Errorf("weeklyTrend[0]: expected {2026-07-06,14000}, got %+v", resp.WeeklyTrend[0])
	}
	if resp.WeeklyTrend[1].WeekStart != "2026-07-13" || resp.WeeklyTrend[1].UsageCents != 24000 {
		t.Errorf("weeklyTrend[1]: expected {2026-07-13,24000}, got %+v", resp.WeeklyTrend[1])
	}
}

// TestUsageTrends_WowSignedPositiveAndNegative verifies wowChangePct is the SIGNED,
// UNCLAMPED ratio: a >2x growth week yields >+1.0 (proves no min(r,1.0) clamp) and a
// halving week yields -0.5.
func TestUsageTrends_WowSignedPositiveAndNegative(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}

	// Growth: week1=10000 → week2=35000 ⇒ (35000-10000)/10000 = +2.5 (>1, so a clamp
	// to 1.0 would be caught).
	grow := []*entity.Transaction{
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 10000, utWeek1Mon),
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 35000, utWeek2Mon),
	}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: grow})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))
	if resp.WowChangePct != 2.5 {
		t.Errorf("wowChangePct (2.5x growth): expected +2.5 (unclamped), got %v", resp.WowChangePct)
	}

	// Shrink: week1=10000 → week2=5000 ⇒ (5000-10000)/10000 = -0.5.
	shrink := []*entity.Transaction{
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 10000, utWeek1Mon),
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 5000, utWeek2Mon),
	}
	h2 := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: shrink})
	resp2 := decodeUsageTrends(t, doUsageTrends(t, h2, appID, pa, ""))
	if resp2.WowChangePct != -0.5 {
		t.Errorf("wowChangePct (halving): expected -0.5 (signed, not clamped), got %v", resp2.WowChangePct)
	}
}

// TestUsageTrends_WowZeroCases verifies wowChangePct is 0 when there are fewer than 2
// weekly buckets, and 0 when the prior week's total is <= 0.
func TestUsageTrends_WowZeroCases(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}

	// Single week → 0.
	single := []*entity.Transaction{usageTrendsTx(appID, "a.myshopify.com", "Acme", 10000, utWeek1Mon)}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: single})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))
	if resp.WowChangePct != 0 {
		t.Errorf("wowChangePct (<2 weeks): expected 0, got %v", resp.WowChangePct)
	}

	// Prior week sums to 0 (a +N and a -N usage adjustment net to zero) → guard returns 0.
	pos := usageTrendsTx(appID, "a.myshopify.com", "Acme", 5000, utWeek1Mon)
	neg := usageTrendsTx(appID, "a.myshopify.com", "Acme", -5000, utWeek1Mon)
	latest := usageTrendsTx(appID, "a.myshopify.com", "Acme", 8000, utWeek2Mon)
	h2 := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: []*entity.Transaction{pos, neg, latest}})
	resp2 := decodeUsageTrends(t, doUsageTrends(t, h2, appID, pa, ""))
	if resp2.WowChangePct != 0 {
		t.Errorf("wowChangePct (prior week 0): expected 0, got %v", resp2.WowChangePct)
	}
}

// TestUsageTrends_ActiveStoresDistinctDomains verifies activeStores counts distinct
// domains with >= 1 USAGE tx (repeat domains counted once; non-USAGE excluded).
func TestUsageTrends_ActiveStoresDistinctDomains(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	recurring := usageTrendsTx(appID, "recurring.myshopify.com", "R", 9999, utWeek1Mon)
	recurring.ChargeType = valueobject.ChargeTypeRecurring
	txs := []*entity.Transaction{
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 1000, utWeek1Mon),
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 2000, utWeek2Mon), // same domain, 2nd week
		usageTrendsTx(appID, "b.myshopify.com", "Beta", 3000, utWeek1Mon),
		recurring, // must not count as an active usage store
	}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: txs})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))
	if resp.ActiveStores != 2 {
		t.Errorf("activeStores: expected 2 distinct USAGE domains, got %d", resp.ActiveStores)
	}
}

// TestUsageTrends_PerStoreGroupingSortAndWow verifies per-store usageCents aggregation,
// per-store WoW from its own weekly buckets, and sort by usageCents descending.
func TestUsageTrends_PerStoreGroupingSortAndWow(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	txs := []*entity.Transaction{
		// high: week1=6000, week2=9000 ⇒ usage 15000, wow (9000-6000)/6000 = +0.5.
		usageTrendsTx(appID, "high.myshopify.com", "", 6000, utWeek1Mon),
		usageTrendsTx(appID, "high.myshopify.com", "High", 9000, utWeek2Mon), // supplies shop name
		// low: single week ⇒ usage 2000, wow 0 (<2 weeks).
		usageTrendsTx(appID, "low.myshopify.com", "Low", 2000, utWeek1Mon),
	}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: txs})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))

	if len(resp.Stores) != 2 {
		t.Fatalf("expected 2 stores, got %d: %+v", len(resp.Stores), resp.Stores)
	}
	high := resp.Stores[0] // sorted by usage desc
	if high.Domain != "high.myshopify.com" || high.UsageCents != 15000 {
		t.Errorf("stores[0]: expected high 15000, got %+v", high)
	}
	if high.ShopName != "High" {
		t.Errorf("expected first-non-empty shop name High, got %q", high.ShopName)
	}
	if high.WowPct != 0.5 {
		t.Errorf("high per-store wowPct: expected +0.5, got %v", high.WowPct)
	}
	low := resp.Stores[1]
	if low.Domain != "low.myshopify.com" || low.UsageCents != 2000 || low.WowPct != 0 {
		t.Errorf("stores[1]: expected low {2000, wow 0}, got %+v", low)
	}
}

// TestUsageTrends_UsageMrrEquivIsWindowSum verifies usageMrrEquivCents is the Σ of net
// USAGE amounts across the whole window (all weeks, all stores).
func TestUsageTrends_UsageMrrEquivIsWindowSum(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	txs := []*entity.Transaction{
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 12000, utWeek1Mon),
		usageTrendsTx(appID, "b.myshopify.com", "Beta", 34000, utWeek2Mon),
		usageTrendsTx(appID, "a.myshopify.com", "Acme", 58000, utWeek2Mon),
	}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: txs})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))
	if resp.UsageMrrEquivCents != 104000 {
		t.Errorf("usageMrrEquivCents: expected 104000 (window sum), got %d", resp.UsageMrrEquivCents)
	}
}

// TestUsageTrends_CurrencyDefaultAndFirstNonEmpty verifies currency defaults to USD when
// empty and otherwise surfaces the first non-empty currency.
func TestUsageTrends_CurrencyDefaultAndFirstNonEmpty(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	t1 := usageTrendsTx(appID, "a.myshopify.com", "Acme", 1000, utWeek1Mon)
	t1.Currency = ""
	t2 := usageTrendsTx(appID, "b.myshopify.com", "Beta", 2000, utWeek1Mon)
	t2.Currency = "EUR"
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: []*entity.Transaction{t1, t2}})
	resp := decodeUsageTrends(t, doUsageTrends(t, h, appID, pa, ""))
	if resp.Currency != "EUR" {
		t.Errorf("currency: expected EUR (first non-empty), got %q", resp.Currency)
	}
}

// TestUsageTrends_Empty verifies the empty case yields zero headlines, USD default, and
// []-serialized (non-nil) weeklyTrend/stores slices.
func TestUsageTrends_Empty(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{})
	rec := doUsageTrends(t, h, appID, pa, "")
	resp := decodeUsageTrends(t, rec)

	if resp.UsageMrrEquivCents != 0 || resp.WowChangePct != 0 || resp.ActiveStores != 0 {
		t.Errorf("expected zero headlines, got %+v", resp)
	}
	if resp.Currency != "USD" {
		t.Errorf("expected USD default currency, got %q", resp.Currency)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"weeklyTrend":[]`) {
		t.Errorf("expected weeklyTrend serialized as [], body: %s", body)
	}
	if !strings.Contains(body, `"stores":[]`) {
		t.Errorf("expected stores serialized as [], body: %s", body)
	}
}

// TestUsageTrends_TxRepoErrorReturns503 verifies transaction-repo failures surface as 503.
func TestUsageTrends_TxRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{findErr: errors.New("db down")})
	rec := doUsageTrends(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUsageTrends_CSVFormatAndEscaping verifies CSV output: header + one row per store
// (sorted by usage desc), wowPct formatted to 4 decimals, and that a comma/quote-laden
// shop name stays a single field.
func TestUsageTrends_CSVFormatAndEscaping(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	txs := []*entity.Transaction{
		// big: week1=8000, week2=12000 ⇒ usage 20000, wow +0.5.
		usageTrendsTx(appID, "big.myshopify.com", `Acme, "Inc"`, 8000, utWeek1Mon),
		usageTrendsTx(appID, "big.myshopify.com", `Acme, "Inc"`, 12000, utWeek2Mon),
		usageTrendsTx(appID, "small.myshopify.com", "Small", 1000, utWeek1Mon),
	}
	h := newUsageTrendsHandler(appID, pa, &mockTxRepo{transactions: txs})
	rec := doUsageTrends(t, h, appID, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "usage-trends.csv") {
		t.Errorf("expected filename usage-trends.csv in Content-Disposition, got %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"domain", "shopName", "usageCents", "wowPct"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by usage desc: big first.
	if records[1][0] != "big.myshopify.com" {
		t.Errorf("expected first row big.myshopify.com, got %s", records[1][0])
	}
	if len(records[1]) != 4 {
		t.Fatalf("expected 4 columns, got %d: %v", len(records[1]), records[1])
	}
	if records[1][1] != `Acme, "Inc"` {
		t.Errorf("shopName column: expected %q, got %q", `Acme, "Inc"`, records[1][1])
	}
	if records[1][2] != "20000" {
		t.Errorf("usageCents column: expected 20000, got %s", records[1][2])
	}
	if records[1][3] != "0.5000" {
		t.Errorf("wowPct column: expected 0.5000, got %s", records[1][3])
	}
}

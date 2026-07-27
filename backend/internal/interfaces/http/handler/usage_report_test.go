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

// usageTx builds a Transaction for usage-report tests. GrossAmountCents is set to a
// distinct value (!= net) so any accidental net→gross regression in the report sums
// is caught — the report must use AmountCents() (net) everywhere.
func usageTx(appID uuid.UUID, domain, shopName string, ct valueobject.ChargeType, netCents int64) *entity.Transaction {
	return &entity.Transaction{
		ID:               uuid.New(),
		AppID:            appID,
		ShopifyGID:       "gid://shopify/AppTransaction/" + uuid.New().String(),
		MyshopifyDomain:  domain,
		ShopName:         shopName,
		ChargeType:       ct,
		GrossAmountCents: netCents + 777, // deliberately != net
		NetAmountCents:   netCents,
		Currency:         "USD",
		TransactionDate:  time.Now(),
		CreatedAt:        time.Now(),
	}
}

func newUsageRequest(t *testing.T, appID uuid.UUID, pa *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/usage"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func usageFixture(txs []*entity.Transaction, snaps []*entity.DailyMetricsSnapshot) (uuid.UUID, *entity.PartnerAccount, *UsageReportHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: txs},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doUsage(t *testing.T, h *UsageReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := newUsageRequest(t, appID, pa, query)
	rec := httptest.NewRecorder()
	h.GetUsageReport(rec, req)
	return rec
}

// TestUsage_HeadlineSumsUsageAndOneTimeOnly verifies usageCents sums ONLY USAGE,
// oneTimeCents sums ONLY ONE_TIME, and chargesCount counts ONLY USAGE+ONE_TIME —
// RECURRING and REFUND transactions are excluded from every headline. Distinct
// amounts ensure a USAGE↔ONE_TIME swap would be caught.
func TestUsage_HeadlineSumsUsageAndOneTimeOnly(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	txs := []*entity.Transaction{
		usageTx(appID, "a.myshopify.com", "Acme", valueobject.ChargeTypeUsage, 30000),
		usageTx(appID, "a.myshopify.com", "Acme", valueobject.ChargeTypeUsage, 1200),
		usageTx(appID, "b.myshopify.com", "Beta", valueobject.ChargeTypeOneTime, 6400),
		// Excluded — RECURRING belongs to MRR, must not count.
		usageTx(appID, "a.myshopify.com", "Acme", valueobject.ChargeTypeRecurring, 99999),
		// Excluded — REFUND is a negative adjustment, must not count.
		usageTx(appID, "b.myshopify.com", "Beta", valueobject.ChargeTypeRefund, 5555),
	}
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: txs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp usageReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.UsageCents != 31200 {
		t.Errorf("usageCents: expected 31200 (usage only), got %d", resp.UsageCents)
	}
	if resp.OneTimeCents != 6400 {
		t.Errorf("oneTimeCents: expected 6400 (one-time only), got %d", resp.OneTimeCents)
	}
	// 2 usage + 1 one-time = 3; recurring/refund excluded.
	if resp.ChargesCount != 3 {
		t.Errorf("chargesCount: expected 3 (usage+one-time only), got %d", resp.ChargesCount)
	}
}

// TestUsage_PerStoreGroupingAndSort verifies per-store usage/oneTime/chargeCount
// aggregation, first-non-empty shop name, and sort by usageCents descending.
func TestUsage_PerStoreGroupingAndSort(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	txs := []*entity.Transaction{
		// low.myshopify.com: usage 1000, one-time 500, 2 charges.
		usageTx(appID, "low.myshopify.com", "Low", valueobject.ChargeTypeUsage, 1000),
		usageTx(appID, "low.myshopify.com", "Low", valueobject.ChargeTypeOneTime, 500),
		// high.myshopify.com: usage 8000+2000=10000, one-time 300, 3 charges.
		// First tx has empty shop name; second supplies "High".
		usageTx(appID, "high.myshopify.com", "", valueobject.ChargeTypeUsage, 8000),
		usageTx(appID, "high.myshopify.com", "High", valueobject.ChargeTypeUsage, 2000),
		usageTx(appID, "high.myshopify.com", "High", valueobject.ChargeTypeOneTime, 300),
		// Recurring must not create a store bucket or bump counts.
		usageTx(appID, "recurring-only.myshopify.com", "R", valueobject.ChargeTypeRecurring, 7777),
	}
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: txs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "")
	var resp usageReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Stores) != 2 {
		t.Fatalf("expected 2 stores (recurring-only excluded), got %d: %+v", len(resp.Stores), resp.Stores)
	}
	// Sorted by usage desc: high first.
	high := resp.Stores[0]
	if high.Domain != "high.myshopify.com" || high.UsageCents != 10000 || high.OneTimeCents != 300 || high.ChargeCount != 3 {
		t.Errorf("stores[0] high: got %+v", high)
	}
	if high.ShopName != "High" {
		t.Errorf("expected first-non-empty shop name High, got %q", high.ShopName)
	}
	low := resp.Stores[1]
	if low.Domain != "low.myshopify.com" || low.UsageCents != 1000 || low.OneTimeCents != 500 || low.ChargeCount != 2 {
		t.Errorf("stores[1] low: got %+v", low)
	}
}

// TestUsage_EmptyDomainBucket verifies transactions with an empty domain form their
// own bucket rather than being dropped.
func TestUsage_EmptyDomainBucket(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	txs := []*entity.Transaction{
		usageTx(appID, "", "", valueobject.ChargeTypeUsage, 4000),
	}
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: txs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "")
	var resp usageReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Stores) != 1 || resp.Stores[0].Domain != "" || resp.Stores[0].UsageCents != 4000 {
		t.Errorf("expected one empty-domain bucket with 4000 usage, got %+v", resp.Stores)
	}
}

// TestUsage_TrendFromSnapshots verifies the trend has one point per snapshot
// (ascending), carrying the snapshot's UsageRevenueCents.
func TestUsage_TrendFromSnapshots(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snaps := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, UsageRevenueCents: 300000},
		{ID: uuid.New(), AppID: appID, Date: d.AddDate(0, 0, 1), UsageRevenueCents: 312000},
	}
	h := NewUsageReportHandler(
		&mockTxRepo{},
		&mockSnapshotRepoForForecast{snapshots: snaps},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "from=2026-06-01&to=2026-07-31")
	var resp usageReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-07-01" || resp.Trend[0].UsageCents != 300000 {
		t.Errorf("trend[0]: expected {2026-07-01,300000}, got {%s,%d}", resp.Trend[0].Date, resp.Trend[0].UsageCents)
	}
	if resp.Trend[1].UsageCents != 312000 {
		t.Errorf("trend[1].usageCents: expected 312000, got %d", resp.Trend[1].UsageCents)
	}
}

// TestUsage_CurrencyDefaultAndFirstNonEmpty verifies currency defaults to USD when no
// tx carries one, and otherwise surfaces the first non-empty currency.
func TestUsage_CurrencyDefaultAndFirstNonEmpty(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}

	// First tx has empty currency, second is EUR → EUR surfaces.
	t1 := usageTx(appID, "a.myshopify.com", "Acme", valueobject.ChargeTypeUsage, 1000)
	t1.Currency = ""
	t2 := usageTx(appID, "b.myshopify.com", "Beta", valueobject.ChargeTypeUsage, 2000)
	t2.Currency = "EUR"
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: []*entity.Transaction{t1, t2}},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "")
	var resp usageReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Currency != "EUR" {
		t.Errorf("currency: expected EUR (first non-empty), got %q", resp.Currency)
	}
}

// TestUsage_Empty verifies the empty case yields zero headlines and []-serialized
// slices (non-nil) with a USD default currency.
func TestUsage_Empty(t *testing.T) {
	appID, pa, h := usageFixture(nil, nil)
	rec := doUsage(t, h, appID, pa, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp usageReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.UsageCents != 0 || resp.OneTimeCents != 0 || resp.ChargesCount != 0 {
		t.Errorf("expected zero headlines, got usage=%d oneTime=%d count=%d",
			resp.UsageCents, resp.OneTimeCents, resp.ChargesCount)
	}
	if resp.Currency != "USD" {
		t.Errorf("expected USD default currency, got %q", resp.Currency)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"stores":[]`) {
		t.Errorf("expected stores serialized as [], body: %s", body)
	}
	if !strings.Contains(body, `"trend":[]`) {
		t.Errorf("expected trend serialized as [], body: %s", body)
	}
}

// TestUsage_TxRepoErrorReturns503 verifies transaction-repo failures surface as 503.
func TestUsage_TxRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUsageReportHandler(
		&mockTxRepo{findErr: errors.New("db down")},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUsage_SnapshotRepoErrorReturns503 verifies snapshot-repo failures surface as 503.
func TestUsage_SnapshotRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: []*entity.Transaction{usageTx(appID, "a.myshopify.com", "Acme", valueobject.ChargeTypeUsage, 1000)}},
		&mockSnapshotRepoForForecast{rangeErr: errors.New("snapshot db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUsage_CSVFormatAndEscaping verifies CSV output: header + one row per store
// (sorted by usage desc), and that a domain/shop name containing a comma stays a
// single field.
func TestUsage_CSVFormatAndEscaping(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	txs := []*entity.Transaction{
		usageTx(appID, "big.myshopify.com", `Acme, "Inc"`, valueobject.ChargeTypeUsage, 9000),
		usageTx(appID, "big.myshopify.com", `Acme, "Inc"`, valueobject.ChargeTypeOneTime, 100),
		usageTx(appID, "small.myshopify.com", "Small", valueobject.ChargeTypeUsage, 1000),
	}
	h := NewUsageReportHandler(
		&mockTxRepo{transactions: txs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUsage(t, h, appID, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "usage.csv") {
		t.Errorf("expected filename usage.csv in Content-Disposition, got %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"domain", "shopName", "usageCents", "oneTimeCents", "chargeCount"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by usage desc: big first.
	if records[1][0] != "big.myshopify.com" {
		t.Errorf("expected first row big.myshopify.com, got %s", records[1][0])
	}
	if len(records[1]) != 5 {
		t.Fatalf("expected 5 columns, got %d: %v", len(records[1]), records[1])
	}
	// The comma/quote-laden shop name must stay one field.
	if records[1][1] != `Acme, "Inc"` {
		t.Errorf("shopName column: expected %q, got %q", `Acme, "Inc"`, records[1][1])
	}
	if records[1][2] != "9000" || records[1][3] != "100" || records[1][4] != "2" {
		t.Errorf("big row values: got usage=%s oneTime=%s count=%s", records[1][2], records[1][3], records[1][4])
	}
}

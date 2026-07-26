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

// churnedSub is a small builder for churned subscription test data.
func churnedSub(appID uuid.UUID, domain, plan string, cents int64) *entity.Subscription {
	now := time.Now().UTC()
	churnedDate := now.AddDate(0, 0, -100)
	return &entity.Subscription{
		ID:                     uuid.New(),
		AppID:                  appID,
		MyshopifyDomain:        domain,
		ShopName:               strings.Split(domain, ".")[0],
		PlanName:               plan,
		BasePriceCents:         cents,
		Currency:               "USD",
		RiskState:              valueobject.RiskStateChurned,
		ExpectedNextChargeDate: &churnedDate,
		CreatedAt:              now.AddDate(0, 0, -200),
	}
}

func newChurnRequest(t *testing.T, appID uuid.UUID, partnerAccount *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/churn"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

// TestChurn_MrrLostAndCount verifies the sum of churned MRR (incl. an annual sub
// normalized ÷12) and the churned count.
func TestChurn_MrrLostAndCount(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	annual := churnedSub(appID, "annual.myshopify.com", "Pro", 1200)
	annual.BillingInterval = valueobject.BillingIntervalAnnual
	subs := []*entity.Subscription{
		churnedSub(appID, "a.myshopify.com", "Pro", 5000),
		churnedSub(appID, "b.myshopify.com", "Basic", 3000),
		annual, // 1200 / 12 = 100
		// Non-churned subs are filtered out by FindByRiskState.
		atRiskSub(appID, "safe.myshopify.com", "Pro", 9999, valueobject.RiskStateSafe),
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.ChurnedCount != 3 {
		t.Errorf("churnedCount: expected 3, got %d", resp.ChurnedCount)
	}
	// 5000 + 3000 + 100 (annual/12) = 8100
	if resp.ChurnedMrrLostCents != 8100 {
		t.Errorf("churnedMrrLostCents: expected 8100, got %d", resp.ChurnedMrrLostCents)
	}
	if resp.Currency != "USD" {
		t.Errorf("currency: expected USD, got %q", resp.Currency)
	}
	// Sum of store MRR-lost == churnedMrrLostCents.
	var storeSum int64
	for _, s := range resp.Stores {
		storeSum += s.MRRLostCents
	}
	if storeSum != resp.ChurnedMrrLostCents {
		t.Errorf("store MRR-lost sum %d != churnedMrrLostCents %d", storeSum, resp.ChurnedMrrLostCents)
	}
}

// TestChurn_ChurnRateFromSnapshot verifies churnRate = churnedCount ÷ latest
// snapshot TotalSubscriptions.
func TestChurn_ChurnRateFromSnapshot(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		churnedSub(appID, "a.myshopify.com", "Pro", 5000),
		churnedSub(appID, "b.myshopify.com", "Basic", 3000),
	}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snapshots := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, ChurnedCount: 1, TotalSubscriptions: 5},
		// Latest snapshot: 2 churned of 8 total → 0.25.
		{ID: uuid.New(), AppID: appID, Date: d.AddDate(0, 0, 1), ChurnedCount: 2, TotalSubscriptions: 8},
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{snapshots: snapshots},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "from=2026-06-01&to=2026-07-31")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// churnedCount=2, latest snapshot total=8 → 2/8 = 0.25.
	if resp.ChurnRate != 0.25 {
		t.Errorf("churnRate: expected 0.25, got %v", resp.ChurnRate)
	}
	// Trend uses each snapshot's own churnedCount/total.
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend points, got %d", len(resp.Trend))
	}
	if resp.Trend[0].Date != "2026-07-01" || resp.Trend[0].ChurnRate != 0.2 {
		t.Errorf("trend[0]: expected {2026-07-01,0.2}, got {%s,%v}", resp.Trend[0].Date, resp.Trend[0].ChurnRate)
	}
	if resp.Trend[1].ChurnRate != 0.25 {
		t.Errorf("trend[1].churnRate: expected 0.25, got %v", resp.Trend[1].ChurnRate)
	}
}

// TestChurn_ChurnRateZeroWhenNoSnapshot verifies churnRate is 0 (not a divide-by-
// zero panic) when there is no snapshot to source TotalSubscriptions from.
func TestChurn_ChurnRateZeroWhenNoSnapshot(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		churnedSub(appID, "a.myshopify.com", "Pro", 5000),
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{}, // no snapshots
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ChurnRate != 0 {
		t.Errorf("churnRate: expected 0 (no snapshot), got %v", resp.ChurnRate)
	}
	if resp.ChurnedCount != 1 {
		t.Errorf("churnedCount: expected 1, got %d", resp.ChurnedCount)
	}
}

// TestChurn_ChurnRateGuardsZeroTotal verifies a snapshot with TotalSubscriptions=0
// yields churnRate 0 rather than a divide-by-zero.
func TestChurn_ChurnRateGuardsZeroTotal(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		churnedSub(appID, "a.myshopify.com", "Pro", 5000),
	}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snapshots := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, ChurnedCount: 0, TotalSubscriptions: 0},
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{snapshots: snapshots},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ChurnRate != 0 {
		t.Errorf("churnRate: expected 0 (zero total), got %v", resp.ChurnRate)
	}
	if len(resp.Trend) != 1 || resp.Trend[0].ChurnRate != 0 {
		t.Errorf("trend churnRate: expected 0, got %+v", resp.Trend)
	}
}

// TestChurn_StoresSortedByMrrLostDesc verifies the ranked store list is sorted by
// MRR lost descending.
func TestChurn_StoresSortedByMrrLostDesc(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		churnedSub(appID, "low.myshopify.com", "Pro", 1000),
		churnedSub(appID, "high.myshopify.com", "Pro", 9000),
		churnedSub(appID, "mid.myshopify.com", "Pro", 5000),
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Stores) != 3 {
		t.Fatalf("expected 3 stores, got %d", len(resp.Stores))
	}
	wantOrder := []int64{9000, 5000, 1000}
	for i, want := range wantOrder {
		if resp.Stores[i].MRRLostCents != want {
			t.Errorf("stores[%d].mrrLostCents: expected %d, got %d", i, want, resp.Stores[i].MRRLostCents)
		}
	}
	// Tenure: created 200d ago, churned 100d ago → 100 days.
	if resp.Stores[0].TenureDays != 100 {
		t.Errorf("tenureDays: expected 100, got %d", resp.Stores[0].TenureDays)
	}
}

// TestChurn_Empty verifies the empty case: no churned subs → zero metrics and
// stores/trend serialize as [] (non-nil).
func TestChurn_Empty(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	// Only a safe sub → nothing churned.
	subs := []*entity.Subscription{
		atRiskSub(appID, "safe.myshopify.com", "Pro", 5000, valueobject.RiskStateSafe),
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ChurnedCount != 0 || resp.ChurnedMrrLostCents != 0 || resp.ChurnRate != 0 {
		t.Errorf("expected zero metrics, got count=%d mrr=%d rate=%v",
			resp.ChurnedCount, resp.ChurnedMrrLostCents, resp.ChurnRate)
	}
	if len(resp.Stores) != 0 {
		t.Errorf("expected empty stores, got %d", len(resp.Stores))
	}
	// stores/trend must serialize as [] (non-nil).
	body := rec.Body.String()
	if !strings.Contains(body, `"stores":[]`) {
		t.Errorf("expected stores serialized as [], body: %s", body)
	}
	if !strings.Contains(body, `"trend":[]`) {
		t.Errorf("expected trend serialized as [], body: %s", body)
	}
}

// TestChurn_RepoErrorReturns503 verifies infra repo failures surface as 503.
func TestChurn_RepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	h := NewChurnHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestChurn_ChurnedDateFallback verifies churnedDateOf falls back to
// LastRecurringChargeDate when ExpectedNextChargeDate is nil, and emits an empty
// churnedDate (with tenure measured to now) when both are nil.
func TestChurn_ChurnedDateFallback(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	now := time.Now().UTC()
	lastCharge := now.AddDate(0, 0, -50)

	fallback := churnedSub(appID, "fallback.myshopify.com", "Pro", 9000)
	fallback.ExpectedNextChargeDate = nil
	fallback.LastRecurringChargeDate = &lastCharge

	bothNil := churnedSub(appID, "bothnil.myshopify.com", "Pro", 1000)
	bothNil.ExpectedNextChargeDate = nil
	bothNil.LastRecurringChargeDate = nil

	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{fallback, bothNil}},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	byDomain := map[string]churnStore{}
	for _, s := range resp.Stores {
		byDomain[s.Domain] = s
	}

	// Fallback uses LastRecurringChargeDate for the churned date.
	fb := byDomain["fallback.myshopify.com"]
	if fb.ChurnedDate != lastCharge.Format(dateLayout) {
		t.Errorf("fallback churnedDate: expected %q, got %q", lastCharge.Format(dateLayout), fb.ChurnedDate)
	}

	// Both nil → empty churnedDate; tenure measured to now (created 200d ago).
	bn := byDomain["bothnil.myshopify.com"]
	if bn.ChurnedDate != "" {
		t.Errorf("both-nil churnedDate: expected empty, got %q", bn.ChurnedDate)
	}
	if bn.TenureDays < 199 || bn.TenureDays > 201 {
		t.Errorf("both-nil tenureDays: expected ~200 (to now), got %d", bn.TenureDays)
	}
}

// TestChurn_SnapshotRepoErrorReturns503 verifies the snapshot-repo error branch
// (distinct from the subscription-repo branch) also surfaces as 503.
func TestChurn_SnapshotRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{
			churnedSub(appID, "a.myshopify.com", "Pro", 5000),
		}},
		&mockSnapshotRepoForForecast{rangeErr: errors.New("snapshot db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestChurn_ChurnRateClampedToOne verifies a stale/behind snapshot (live churned
// count exceeds the snapshot's total) never yields a churnRate above 1.0.
func TestChurn_ChurnRateClampedToOne(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	// 3 live churned subs, but the latest snapshot only knows of 2 total.
	subs := []*entity.Subscription{
		churnedSub(appID, "a.myshopify.com", "Pro", 5000),
		churnedSub(appID, "b.myshopify.com", "Pro", 3000),
		churnedSub(appID, "c.myshopify.com", "Pro", 1000),
	}
	d := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	snapshots := []*entity.DailyMetricsSnapshot{
		{ID: uuid.New(), AppID: appID, Date: d, ChurnedCount: 1, TotalSubscriptions: 2},
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{snapshots: snapshots},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	var resp churnReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// 3 churned / 2 total = 1.5 → clamped to 1.0.
	if resp.ChurnRate != 1 {
		t.Errorf("churnRate: expected 1 (clamped), got %v", resp.ChurnRate)
	}
}

// TestChurn_CSVEscaping verifies free-text fields containing commas/quotes are
// properly quoted (encoding/csv), so a shopName like "Acme, Inc." stays one column.
func TestChurn_CSVEscaping(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	tricky := churnedSub(appID, "acme.myshopify.com", `Pro, "Annual"`, 5000)
	tricky.ShopName = "Acme, Inc."
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{tricky}},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "format=csv")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d records: %v", len(records), records)
	}
	// Row must keep column integrity despite the embedded comma/quote.
	if len(records[1]) != 6 {
		t.Fatalf("expected 6 columns, got %d: %v", len(records[1]), records[1])
	}
	if records[1][1] != "Acme, Inc." {
		t.Errorf("shopName column: expected %q, got %q", "Acme, Inc.", records[1][1])
	}
	if records[1][5] != `Pro, "Annual"` {
		t.Errorf("planName column: expected %q, got %q", `Pro, "Annual"`, records[1][5])
	}
}

// TestChurn_CSVFormat verifies CSV output: header + one row per store, sorted by
// MRR lost desc.
func TestChurn_CSVFormat(t *testing.T) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	subs := []*entity.Subscription{
		churnedSub(appID, "high.myshopify.com", "Pro", 9000),
		churnedSub(appID, "low.myshopify.com", "Basic", 1000),
	}
	h := NewChurnHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockSnapshotRepoForForecast{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)

	req := newChurnRequest(t, appID, partnerAccount, "format=csv")
	rec := httptest.NewRecorder()
	h.GetChurn(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "churn.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d records: %v", len(records), records)
	}
	wantHeader := []string{"domain", "shopName", "mrrLostCents", "churnedDate", "tenureDays", "planName"}
	if len(records[0]) != len(wantHeader) {
		t.Fatalf("header column count: expected %d, got %d", len(wantHeader), len(records[0]))
	}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Sorted by MRR lost desc: high first.
	if records[1][0] != "high.myshopify.com" {
		t.Errorf("expected first row to be high MRR store, got %s", records[1][0])
	}
}

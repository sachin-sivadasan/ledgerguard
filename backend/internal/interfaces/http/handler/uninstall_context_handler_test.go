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

// uninstallSub builds a subscription correlated by ShopifyShopGID for uninstall tests.
func uninstallSub(appID uuid.UUID, shopGID, domain, plan string, state valueobject.RiskState, createdAt time.Time) *entity.Subscription {
	return &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		ShopifyShopGID:  shopGID,
		MyshopifyDomain: domain,
		PlanName:        plan,
		Currency:        "USD",
		Status:          "UNINSTALLED",
		RiskState:       state,
		CreatedAt:       createdAt,
	}
}

func newUninstallRequest(t *testing.T, appID uuid.UUID, partnerAccount *entity.PartnerAccount, query string) *http.Request {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/uninstall-context"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return req
}

func uninstallFixture(subs []*entity.Subscription, events []*entity.AppEvent) (uuid.UUID, *entity.PartnerAccount, *UninstallContextHandler) {
	appID := uuid.New()
	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: partnerAccount.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: partnerAccount},
	)
	return appID, partnerAccount, h
}

func doUninstall(t *testing.T, h *UninstallContextHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := newUninstallRequest(t, appID, pa, query)
	rec := httptest.NewRecorder()
	h.GetUninstallContext(rec, req)
	return rec
}

func decodeUninstall(t *testing.T, rec *httptest.ResponseRecorder) uninstallReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp uninstallReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// TestUninstall_EventTypeFiltering verifies UNINSTALL* matches case-insensitively and
// non-uninstall event types are excluded.
func TestUninstall_EventTypeFiltering(t *testing.T) {
	appID := uuid.New()
	when := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "app_uninstall", OccurredAt: when},
		// Non-uninstall → excluded.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "INSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/4", EventType: "REACTIVATED", OccurredAt: when},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)

	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Uninstalls != 2 {
		t.Errorf("uninstalls: expected 2 (shop/1 + shop/2), got %d", resp.Uninstalls)
	}
}

// TestUninstall_DateRangeBoundary verifies from-edge inclusive, whole to-day inclusive,
// day-after-to excluded.
func TestUninstall_DateRangeBoundary(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	fromMidnight := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	toAfternoon := time.Date(2026, 7, 31, 14, 30, 0, 0, time.UTC)
	dayAfterTo := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: fromMidnight},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "APP_UNINSTALLED", OccurredAt: toAfternoon},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "APP_UNINSTALLED", OccurredAt: dayAfterTo},
	}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Uninstalls != 2 {
		t.Errorf("uninstalls: expected 2 (from-edge + to-day included, day-after excluded), got %d", resp.Uninstalls)
	}
}

// TestUninstall_DedupLatestEvent verifies a shop with multiple uninstall events in range
// yields one row using the LATEST event date.
func TestUninstall_DedupLatestEvent(t *testing.T) {
	appID := uuid.New()
	early := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	late := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: early},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: late},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Uninstalls != 1 {
		t.Fatalf("uninstalls: expected 1 (deduped), got %d", resp.Uninstalls)
	}
	if resp.Stores[0].UninstalledDate != "2026-07-20" {
		t.Errorf("expected latest uninstall date 2026-07-20, got %s", resp.Stores[0].UninstalledDate)
	}
}

// TestUninstall_StateMapping verifies correlation → state mapping for each RiskState,
// FROZEN status, and a missing subscription (Unknown).
func TestUninstall_StateMapping(t *testing.T) {
	appID := uuid.New()
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	when := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)

	frozen := uninstallSub(appID, "gid://shop/frozen", "frozen.myshopify.com", "Pro", valueobject.RiskStateSafe, created)
	frozen.Status = "FROZEN"

	subs := []*entity.Subscription{
		uninstallSub(appID, "gid://shop/safe", "safe.myshopify.com", "Pro", valueobject.RiskStateSafe, created),
		uninstallSub(appID, "gid://shop/one", "one.myshopify.com", "Pro", valueobject.RiskStateOneCycleMissed, created),
		uninstallSub(appID, "gid://shop/two", "two.myshopify.com", "Pro", valueobject.RiskStateTwoCyclesMissed, created),
		uninstallSub(appID, "gid://shop/churn", "churn.myshopify.com", "Pro", valueobject.RiskStateChurned, created),
		frozen,
	}
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/safe", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/one", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/two", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/churn", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/frozen", EventType: "APP_UNINSTALLED", OccurredAt: when},
		// No matching sub → Unknown.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/nosub", EventType: "APP_UNINSTALLED", OccurredAt: when},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))

	got := map[string]string{}
	for _, s := range resp.Stores {
		got[s.Domain] = s.StateBeforeUninstall
	}
	want := map[string]string{
		"safe.myshopify.com":   "Healthy",
		"one.myshopify.com":    "At-Risk",
		"two.myshopify.com":    "At-Risk",
		"churn.myshopify.com":  "At-Risk",
		"frozen.myshopify.com": "Frozen",
		"gid://shop/nosub":     "Unknown", // falls back to shop GID as domain
	}
	for domain, wantState := range want {
		if got[domain] != wantState {
			t.Errorf("state for %s: expected %s, got %s", domain, wantState, got[domain])
		}
	}
}

// TestUninstall_WereAtRiskPct verifies wereAtRiskPct = atRisk ÷ correlated, with an
// Unknown (uncorrelated) row excluded from the denominator.
func TestUninstall_WereAtRiskPct(t *testing.T) {
	appID := uuid.New()
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	when := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{
		uninstallSub(appID, "gid://shop/1", "a.myshopify.com", "Pro", valueobject.RiskStateChurned, created),
		uninstallSub(appID, "gid://shop/2", "b.myshopify.com", "Pro", valueobject.RiskStateOneCycleMissed, created),
		uninstallSub(appID, "gid://shop/3", "c.myshopify.com", "Pro", valueobject.RiskStateSafe, created),
	}
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "APP_UNINSTALLED", OccurredAt: when},
		// Uncorrelated (no sub) → excluded from denominator.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/nosub", EventType: "APP_UNINSTALLED", OccurredAt: when},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Uninstalls != 4 {
		t.Fatalf("uninstalls: expected 4 rows, got %d", resp.Uninstalls)
	}
	// 2 at-risk / 3 correlated (nosub excluded) ≈ 0.6667.
	if resp.WereAtRiskPct < 0.66 || resp.WereAtRiskPct > 0.67 {
		t.Errorf("wereAtRiskPct: expected ~0.6667, got %v", resp.WereAtRiskPct)
	}
}

// TestUninstall_MedianTenureOdd verifies the median over correlated rows with an odd count.
func TestUninstall_MedianTenureOdd(t *testing.T) {
	appID := uuid.New()
	when := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	// Tenures (days÷30): 30d→1.0, 90d→3.0, 300d→10.0. Median = 3.0.
	subs := []*entity.Subscription{
		uninstallSub(appID, "gid://shop/1", "a.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -30)),
		uninstallSub(appID, "gid://shop/2", "b.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -90)),
		uninstallSub(appID, "gid://shop/3", "c.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -300)),
	}
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "APP_UNINSTALLED", OccurredAt: when},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.MedianTenureMonths != 3.0 {
		t.Errorf("medianTenureMonths (odd): expected 3.0, got %v", resp.MedianTenureMonths)
	}
}

// TestUninstall_MedianTenureEven verifies the median (avg of two middles) with an even count.
func TestUninstall_MedianTenureEven(t *testing.T) {
	appID := uuid.New()
	when := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	// Tenures: 30d→1.0, 60d→2.0, 90d→3.0, 120d→4.0. Median = (2.0+3.0)/2 = 2.5.
	subs := []*entity.Subscription{
		uninstallSub(appID, "gid://shop/1", "a.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -30)),
		uninstallSub(appID, "gid://shop/2", "b.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -60)),
		uninstallSub(appID, "gid://shop/3", "c.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -90)),
		uninstallSub(appID, "gid://shop/4", "d.myshopify.com", "Pro", valueobject.RiskStateSafe, when.AddDate(0, 0, -120)),
	}
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "APP_UNINSTALLED", OccurredAt: when},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/4", EventType: "APP_UNINSTALLED", OccurredAt: when},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.MedianTenureMonths != 2.5 {
		t.Errorf("medianTenureMonths (even): expected 2.5, got %v", resp.MedianTenureMonths)
	}
}

// TestUninstall_SortedByDateDesc verifies stores are sorted by uninstalledDate descending.
func TestUninstall_SortedByDateDesc(t *testing.T) {
	appID := uuid.New()
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/2", EventType: "APP_UNINSTALLED", OccurredAt: time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/3", EventType: "APP_UNINSTALLED", OccurredAt: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	resp := decodeUninstall(t, doUninstall(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	got := []string{resp.Stores[0].UninstalledDate, resp.Stores[1].UninstalledDate, resp.Stores[2].UninstalledDate}
	want := []string{"2026-07-25", "2026-07-15", "2026-07-05"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("stores order: expected %v, got %v", want, got)
			break
		}
	}
}

// TestUninstall_Empty verifies the empty case yields zeros and []-serialized stores.
func TestUninstall_Empty(t *testing.T) {
	appID, pa, h := uninstallFixture(nil, nil)
	rec := doUninstall(t, h, appID, pa, "")
	resp := decodeUninstall(t, rec)
	if resp.Uninstalls != 0 || resp.WereAtRiskPct != 0 || resp.MedianTenureMonths != 0 {
		t.Errorf("expected zero metrics, got %+v", resp)
	}
	if !strings.Contains(rec.Body.String(), `"stores":[]`) {
		t.Errorf("expected stores serialized as [], body: %s", rec.Body.String())
	}
}

// TestUninstall_SubRepoErrorReturns503 verifies subscription-repo failures surface as 503.
func TestUninstall_SubRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUninstall(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUninstall_EventRepoErrorReturns503 verifies event-repo failures surface as 503.
func TestUninstall_EventRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: []*entity.Subscription{uninstallSub(appID, "gid://shop/1", "a.myshopify.com", "Pro", valueobject.RiskStateSafe, time.Now())}},
		&mockAppEventRepo{err: errors.New("events db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUninstall(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestUninstall_CSVFormat verifies CSV output: header, one row per shop, blank tenure
// for uncorrelated rows, and escaping of commas/quotes in free-text columns.
func TestUninstall_CSVFormat(t *testing.T) {
	appID := uuid.New()
	created := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	when := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{
		uninstallSub(appID, "gid://shop/1", "acme.myshopify.com", `Pro, "Annual"`, valueobject.RiskStateChurned, created),
	}
	events := []*entity.AppEvent{
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/1", EventType: "APP_UNINSTALLED", OccurredAt: when},
		// Uncorrelated → blank tenure cell.
		{ID: uuid.New(), ShopifyShopGID: "gid://shop/nosub", EventType: "APP_UNINSTALLED", OccurredAt: when},
	}
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewUninstallContextHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doUninstall(t, h, appID, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "uninstall-context.csv") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"domain", "stateBeforeUninstall", "planName", "tenureMonths", "uninstalledDate"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d]: expected %q, got %q", i, want, records[0][i])
		}
	}
	// Find the correlated row (acme) and verify the plan escaping stayed one column.
	var acme []string
	for _, rec := range records[1:] {
		if rec[0] == "acme.myshopify.com" {
			acme = rec
		}
	}
	if acme == nil {
		t.Fatalf("acme row not found in %v", records)
	}
	if len(acme) != 5 {
		t.Fatalf("expected 5 columns, got %d: %v", len(acme), acme)
	}
	if acme[2] != `Pro, "Annual"` {
		t.Errorf("plan column: expected %q, got %q", `Pro, "Annual"`, acme[2])
	}
	// Uncorrelated row has a blank tenure cell.
	for _, rec := range records[1:] {
		if rec[0] == "gid://shop/nosub" && rec[3] != "" {
			t.Errorf("expected blank tenure for uncorrelated row, got %q", rec[3])
		}
	}
}

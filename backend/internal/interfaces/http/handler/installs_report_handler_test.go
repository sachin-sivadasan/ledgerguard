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

// installEvt builds an app event for the installs report tests.
func installEvt(shopGID, eventType string, when time.Time) *entity.AppEvent {
	return &entity.AppEvent{ID: uuid.New(), ShopifyShopGID: shopGID, EventType: eventType, OccurredAt: when}
}

// installsSub builds a subscription correlated by ShopifyShopGID (for domain resolution).
func installsSub(appID uuid.UUID, shopGID, domain string) *entity.Subscription {
	return &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		ShopifyShopGID:  shopGID,
		MyshopifyDomain: domain,
		Currency:        "USD",
	}
}

func installsFixture(subs []*entity.Subscription, events []*entity.AppEvent) (uuid.UUID, *entity.PartnerAccount, *InstallsReportHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewInstallsReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doInstalls(t *testing.T, h *InstallsReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/installs"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.GetInstalls(rec, req)
	return rec
}

func decodeInstalls(t *testing.T, rec *httptest.ResponseRecorder) installsReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp installsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

var insDay = time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

// TestInstalls_KPIsAndUninstallSubstringTrap verifies the install/uninstall counts, net,
// exclusion of non-lifecycle events, AND that RELATIONSHIP_UNINSTALLED is classified as
// an Uninstall (its "INSTALL" substring must NOT make it an install).
func TestInstalls_KPIsAndUninstallSubstringTrap(t *testing.T) {
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", insDay),
		installEvt("gid://shop/2", "RELATIONSHIP_INSTALLED", insDay),
		installEvt("gid://shop/3", "RELATIONSHIP_INSTALLED", insDay),
		installEvt("gid://shop/4", "RELATIONSHIP_UNINSTALLED", insDay),     // must count as Uninstall
		installEvt("gid://shop/5", "SUBSCRIPTION_CHARGE_ACCEPTED", insDay), // excluded
	}
	appID, pa, h := installsFixture(nil, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 3 {
		t.Errorf("installs: expected 3, got %d", resp.Installs)
	}
	if resp.Uninstalls != 1 {
		t.Errorf("uninstalls: expected 1 (RELATIONSHIP_UNINSTALLED must NOT be an install), got %d", resp.Uninstalls)
	}
	if resp.Net != 2 {
		t.Errorf("net: expected +2, got %d", resp.Net)
	}
}

// TestInstalls_NonLifecycleEventsExcluded verifies that the other real event types the
// (unfiltered) sync writes — RELATIONSHIP_REACTIVATED / RELATIONSHIP_DEACTIVATED and
// SUBSCRIPTION_CHARGE_* — are excluded from the counts, trend and table. Only the exact
// INSTALLED/UNINSTALLED types count.
func TestInstalls_NonLifecycleEventsExcluded(t *testing.T) {
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", insDay),
		installEvt("gid://shop/2", "RELATIONSHIP_REACTIVATED", insDay),
		installEvt("gid://shop/3", "RELATIONSHIP_DEACTIVATED", insDay),
		installEvt("gid://shop/4", "SUBSCRIPTION_CHARGE_ACCEPTED", insDay),
		installEvt("gid://shop/5", "SUBSCRIPTION_CHARGE_CANCELED", insDay),
	}
	appID, pa, h := installsFixture(nil, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 || resp.Uninstalls != 0 {
		t.Errorf("expected only the 1 install counted (reactivate/deactivate/charge excluded), got %+v", resp)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event row, got %d", len(resp.Events))
	}
}

// TestInstalls_NetCanBeNegative verifies net = installs − uninstalls can go negative.
func TestInstalls_NetCanBeNegative(t *testing.T) {
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", insDay),
		installEvt("gid://shop/2", "RELATIONSHIP_UNINSTALLED", insDay),
		installEvt("gid://shop/3", "RELATIONSHIP_UNINSTALLED", insDay),
	}
	appID, pa, h := installsFixture(nil, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Net != -1 {
		t.Errorf("net: expected -1 (1 install − 2 uninstalls), got %d", resp.Net)
	}
}

// TestInstalls_DateBoundary pins the [from,to] window: events at `from` and any time on
// the `to` day are included; the day after `to` is excluded.
func TestInstalls_DateBoundary(t *testing.T) {
	fromMidnight := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	toAfternoon := time.Date(2026, 7, 31, 14, 30, 0, 0, time.UTC)
	dayAfter := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", fromMidnight),
		installEvt("gid://shop/2", "RELATIONSHIP_INSTALLED", toAfternoon),
		installEvt("gid://shop/3", "RELATIONSHIP_INSTALLED", dayAfter), // excluded
	}
	appID, pa, h := installsFixture(nil, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 2 {
		t.Errorf("installs boundary: expected 2 (from + to-day, day-after excluded), got %d", resp.Installs)
	}
}

// TestInstalls_DailyTrend verifies same-day events aggregate into one trend point with
// separate install/uninstall counts, sorted ascending by date.
func TestInstalls_DailyTrend(t *testing.T) {
	d1 := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	d1b := time.Date(2026, 7, 10, 18, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 7, 12, 9, 0, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", d2), // later day, added first
		installEvt("gid://shop/2", "RELATIONSHIP_INSTALLED", d1),
		installEvt("gid://shop/3", "RELATIONSHIP_INSTALLED", d1b),
		installEvt("gid://shop/4", "RELATIONSHIP_UNINSTALLED", d1),
	}
	appID, pa, h := installsFixture(nil, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if len(resp.Trend) != 2 {
		t.Fatalf("expected 2 trend days, got %d: %+v", len(resp.Trend), resp.Trend)
	}
	if resp.Trend[0].Date != "2026-07-10" || resp.Trend[0].Installs != 2 || resp.Trend[0].Uninstalls != 1 {
		t.Errorf("trend[0]: expected {2026-07-10, 2, 1}, got %+v", resp.Trend[0])
	}
	if resp.Trend[1].Date != "2026-07-12" || resp.Trend[1].Installs != 1 || resp.Trend[1].Uninstalls != 0 {
		t.Errorf("trend[1]: expected {2026-07-12, 1, 0}, got %+v", resp.Trend[1])
	}
}

// TestInstalls_RecentEventsNewestFirstAndDomain verifies the recent-events table is
// newest-first, labels Install/Uninstall, and resolves the domain from the correlated
// subscription (falling back to the raw shop GID when uncorrelated).
func TestInstalls_RecentEventsNewestFirstAndDomain(t *testing.T) {
	older := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 7, 20, 9, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{installsSub(uuid.New(), "gid://shop/known", "acme.myshopify.com")}
	events := []*entity.AppEvent{
		installEvt("gid://shop/known", "RELATIONSHIP_INSTALLED", older),
		installEvt("gid://shop/unknown", "RELATIONSHIP_UNINSTALLED", newer),
	}
	appID, pa, h := installsFixture(subs, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if len(resp.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(resp.Events))
	}
	// Newest first: the uninstall on Jul 20 leads.
	if resp.Events[0].Event != "Uninstall" || resp.Events[0].Date != "2026-07-20" {
		t.Errorf("events[0]: expected {Uninstall, 2026-07-20}, got %+v", resp.Events[0])
	}
	if resp.Events[0].Domain != "gid://shop/unknown" { // no sub → raw GID fallback
		t.Errorf("events[0].domain: expected raw GID fallback, got %q", resp.Events[0].Domain)
	}
	if resp.Events[1].Event != "Install" || resp.Events[1].Domain != "acme.myshopify.com" {
		t.Errorf("events[1]: expected {Install, acme.myshopify.com}, got %+v", resp.Events[1])
	}
}

// TestInstalls_RecentEventsCapped verifies the recent-events table caps at
// recentInstallEventsLimit while the KPIs still count every event, AND that the cap keeps
// the NEWEST rows (dropping the oldest) — a distinct domain/time per event pins which 50
// survive, so a "keep oldest 50" regression is caught.
func TestInstalls_RecentEventsCapped(t *testing.T) {
	n := recentInstallEventsLimit + 5
	events := make([]*entity.AppEvent, 0, n)
	base := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	for i := range n {
		// i increases with OccurredAt → shop/(n-1) is the newest event.
		events = append(events, installEvt(fmt.Sprintf("gid://shop/%d", i), "RELATIONSHIP_INSTALLED", base.Add(time.Duration(i)*time.Minute)))
	}
	appID, pa, h := installsFixture(nil, events)
	resp := decodeInstalls(t, doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != n {
		t.Errorf("installs: expected %d (all counted), got %d", n, resp.Installs)
	}
	if len(resp.Events) != recentInstallEventsLimit {
		t.Fatalf("events: expected capped at %d, got %d", recentInstallEventsLimit, len(resp.Events))
	}
	// Newest survives at the top; the 5 oldest are dropped by the cap.
	if resp.Events[0].Domain != fmt.Sprintf("gid://shop/%d", n-1) {
		t.Errorf("events[0]: expected newest gid://shop/%d, got %q", n-1, resp.Events[0].Domain)
	}
	shown := make(map[string]bool, len(resp.Events))
	for _, e := range resp.Events {
		shown[e.Domain] = true
	}
	for i := range 5 {
		if shown[fmt.Sprintf("gid://shop/%d", i)] {
			t.Errorf("expected oldest event gid://shop/%d to be dropped by the cap", i)
		}
	}
}

// TestInstalls_Empty verifies the empty case yields zeros and []-serialized slices.
func TestInstalls_Empty(t *testing.T) {
	appID, pa, h := installsFixture(nil, nil)
	rec := doInstalls(t, h, appID, pa, "")
	resp := decodeInstalls(t, rec)
	if resp.Installs != 0 || resp.Uninstalls != 0 || resp.Net != 0 {
		t.Errorf("expected zero KPIs, got %+v", resp)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"trend":[]`) || !strings.Contains(body, `"events":[]`) {
		t.Errorf("expected trend and events serialized as [], body: %s", body)
	}
}

// TestInstalls_SubRepoErrorReturns503 verifies subscription-repo failures surface as 503.
func TestInstalls_SubRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewInstallsReportHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if rec := doInstalls(t, h, appID, pa, ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

// TestInstalls_EventRepoErrorReturns503 verifies event-repo failures surface as 503.
func TestInstalls_EventRepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewInstallsReportHandler(
		&mockSubscriptionRepo{},
		&mockAppEventRepo{err: errors.New("events db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if rec := doInstalls(t, h, appID, pa, ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

// TestInstalls_Unauthenticated verifies a missing user yields 401.
func TestInstalls_Unauthenticated(t *testing.T) {
	appID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/reports/installs", nil)
	req = withURLParam(req, "appID", appID.String())
	rec := httptest.NewRecorder()
	_, _, h := installsFixture(nil, nil)
	h.GetInstalls(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestInstalls_CSVFormat verifies CSV output: header + one row per recent event
// (newest first), with domain/event/date and comma-safe escaping.
func TestInstalls_CSVFormat(t *testing.T) {
	subs := []*entity.Subscription{installsSub(uuid.New(), "gid://shop/1", `Acme, Inc.myshopify.com`)}
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", insDay),
	}
	appID, pa, h := installsFixture(subs, events)
	rec := doInstalls(t, h, appID, pa, "from=2026-07-01&to=2026-07-31&format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "installs.csv") {
		t.Errorf("Content-Disposition = %q, want installs.csv", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d: %v", len(records), records)
	}
	wantHeader := []string{"store", "event", "date"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], want)
		}
	}
	if records[1][0] != `Acme, Inc.myshopify.com` || records[1][1] != "Install" || records[1][2] != "2026-07-10" {
		t.Errorf("row = %v, want {Acme, Inc.myshopify.com, Install, 2026-07-10}", records[1])
	}
}

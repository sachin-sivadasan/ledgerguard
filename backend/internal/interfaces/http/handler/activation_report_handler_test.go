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

var actDay = time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)

// actSub builds a subscription; charged=true sets LastRecurringChargeDate (the
// "reached a recurring charge" = Paid signal).
func actSub(appID uuid.UUID, shopGID, domain string, charged bool) *entity.Subscription {
	s := &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		ShopifyShopGID:  shopGID,
		MyshopifyDomain: domain,
		Currency:        "USD",
	}
	if charged {
		d := actDay
		s.LastRecurringChargeDate = &d
	}
	return s
}

func activationFixture(subs []*entity.Subscription, events []*entity.AppEvent) (uuid.UUID, *entity.PartnerAccount, *ActivationReportHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewActivationReportHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppEventRepo{events: events},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doActivation(t *testing.T, h *ActivationReportHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/reports/activation"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.GetActivation(rec, req)
	return rec
}

func decodeActivation(t *testing.T, rec *httptest.ResponseRecorder) activationReport {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp activationReport
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

func approxEq(a, b float64) bool {
	d := a - b
	return d < 1e-6 && d > -1e-6
}

// TestActivation_InstallsToPaid: the 2-stage funnel = distinct lifetime installs →
// distinct paying shops. 4 installs, 2 of which pay → 50%.
func TestActivation_InstallsToPaid(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "gid://shop/a", "a.myshopify.com", true),  // paid
		actSub(appID, "gid://shop/b", "b.myshopify.com", true),  // paid
		actSub(appID, "gid://shop/c", "c.myshopify.com", false), // installed, not paid
	}
	events := []*entity.AppEvent{
		installEvt("a.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("b.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("c.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("d.myshopify.com", "RELATIONSHIP_INSTALLED", actDay), // install only, no sub
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, ""))

	if resp.Installs != 4 || resp.Paid != 2 {
		t.Fatalf("expected installs=4 paid=2, got %d/%d", resp.Installs, resp.Paid)
	}
	if !approxEq(resp.OverallPct, 0.5) {
		t.Errorf("overallPct: expected 0.5, got %v", resp.OverallPct)
	}
	if len(resp.Stages) != 2 || resp.Stages[0].Key != "installs" || resp.Stages[1].Key != "paid" {
		t.Fatalf("expected 2 stages [installs, paid], got %+v", resp.Stages)
	}
	if resp.Stages[0].Count != 4 || resp.Stages[1].Count != 2 {
		t.Errorf("stage counts disagree with scalars: %+v", resp.Stages)
	}
	if !approxEq(resp.Stages[1].PctOfPrior, 0.5) {
		t.Errorf("paid stage pctOfPrior: expected 0.5, got %v", resp.Stages[1].PctOfPrior)
	}
}

// TestActivation_MatchesConversionHelper is the RPT-ACTIVATION-1 guard: the funnel is
// computed via the SAME helper as the Installs report's conversion headline, so the two
// must be identical for any input.
func TestActivation_MatchesConversionHelper(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "gid://shop/a", "a.myshopify.com", true),
		actSub(appID, "gid://shop/b", "b.myshopify.com", false),
	}
	events := []*entity.AppEvent{
		installEvt("a.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("b.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("c.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
	}
	_, conv := computeLifecycleAndConversion(events, subs)

	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, ""))
	if resp.Installs != conv.Installs || resp.Paid != conv.Paid || !approxEq(resp.OverallPct, conv.Rate) {
		t.Errorf("funnel (%d/%d/%v) diverged from conversion helper (%d/%d/%v)",
			resp.Installs, resp.Paid, resp.OverallPct, conv.Installs, conv.Paid, conv.Rate)
	}
}

// TestActivation_DefragmentsShopKeys: a shop whose install events are stored under both a
// ShopName (free era) and its domain (paid era) counts as ONE install, not two — inherited
// from the shared conversion helper's canonicalization.
func TestActivation_DefragmentsShopKeys(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "gid://shop/1", "acme.myshopify.com", true),
	}
	// same shop under two keys.
	subs[0].ShopName = "Acme Store"
	events := []*entity.AppEvent{
		installEvt("Acme Store", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("acme.myshopify.com", "RELATIONSHIP_INSTALLED", actDay.Add(time.Hour)),
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, ""))
	if resp.Installs != 1 || resp.Paid != 1 {
		t.Fatalf("two keys for one shop must dedup to installs=1 paid=1, got %d/%d", resp.Installs, resp.Paid)
	}
}

// TestActivation_EmptyData: no events/subs → all zeros, 2 stages, no divide-by-zero.
func TestActivation_EmptyData(t *testing.T) {
	aid, pa, h := activationFixture(nil, nil)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, ""))
	if resp.Installs != 0 || resp.Paid != 0 || resp.OverallPct != 0 {
		t.Errorf("expected all zero, got installs=%d paid=%d pct=%v", resp.Installs, resp.Paid, resp.OverallPct)
	}
	if len(resp.Stages) != 2 {
		t.Errorf("expected 2 stages even when empty, got %d", len(resp.Stages))
	}
}

// TestActivation_CSV: CSV lists the two stages with counts and percentages.
func TestActivation_CSV(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "gid://shop/a", "a.myshopify.com", true),
		actSub(appID, "gid://shop/b", "b.myshopify.com", false),
	}
	events := []*entity.AppEvent{
		installEvt("a.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("b.myshopify.com", "RELATIONSHIP_INSTALLED", actDay),
	}
	aid, pa, h := activationFixture(subs, events)
	rec := doActivation(t, h, aid, pa, "format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("content-type: expected text/csv, got %q", ct)
	}
	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(rows) != 3 { // header + 2 stages
		t.Fatalf("expected 3 CSV rows, got %d: %v", len(rows), rows)
	}
	if rows[0][0] != "stage" || rows[1][0] != "Installs" || rows[2][0] != "Paid / Recurring" {
		t.Errorf("unexpected CSV shape: %v", rows)
	}
	if rows[1][1] != "2" || rows[2][1] != "1" {
		t.Errorf("CSV counts wrong (want installs=2 paid=1): %v", rows)
	}
	// paid = 0.5 of installs (both pctOfPrior and pctOfInstalls).
	if rows[2][2] != "0.5000" || rows[2][3] != "0.5000" {
		t.Errorf("paid row pct wrong (want 0.5/0.5): %v", rows[2])
	}
}

func TestActivation_RepoError503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}

	// Subscription repo failure → 503.
	h1 := NewActivationReportHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("subs db down")},
		&mockAppEventRepo{},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if rec := doActivation(t, h1, appID, pa, ""); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("subs error: expected 503, got %d", rec.Code)
	}

	// Event repo failure → 503.
	h2 := NewActivationReportHandler(
		&mockSubscriptionRepo{},
		&mockAppEventRepo{err: errors.New("events db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	if rec := doActivation(t, h2, appID, pa, ""); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("events error: expected 503, got %d", rec.Code)
	}
}

func TestActivation_Unauthorized401(t *testing.T) {
	appID := uuid.New()
	_, _, h := activationFixture(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/reports/activation", nil)
	req = withURLParam(req, "appID", appID.String())
	rec := httptest.NewRecorder()
	h.GetActivation(rec, req) // no user in context
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

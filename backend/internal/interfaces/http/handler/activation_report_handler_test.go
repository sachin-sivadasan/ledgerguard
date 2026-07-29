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

// acctEvt builds a SUBSCRIPTION_CHARGE_ACCEPTED app event.
func acctEvt(shopKey string, when time.Time) *entity.AppEvent {
	return &entity.AppEvent{ID: uuid.New(), ShopifyShopGID: shopKey, EventType: "SUBSCRIPTION_CHARGE_ACCEPTED", OccurredAt: when}
}

// actSub builds a subscription correlated by ShopifyShopGID; charged=true sets
// LastRecurringChargeDate (the "reached a recurring charge" signal).
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

// TestActivation_FullFunnel: 5 installs → 3 accepted (started) → 2 charged (paid), with
// nested subsets and correct conversion percentages.
func TestActivation_FullFunnel(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "s1", "s1.myshopify.com", true),  // started + paid
		actSub(appID, "s2", "s2.myshopify.com", true),  // started + paid
		actSub(appID, "s3", "s3.myshopify.com", false), // accepted but not charged → started, not paid
	}
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s2", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s3", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s4", "RELATIONSHIP_INSTALLED", actDay), // install only
		installEvt("s5", "RELATIONSHIP_INSTALLED", actDay), // install only
		acctEvt("s1", actDay),
		acctEvt("s2", actDay),
		acctEvt("s3", actDay),
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))

	if resp.Installs != 5 || resp.Started != 3 || resp.Paid != 2 {
		t.Fatalf("stages: expected installs=5 started=3 paid=2, got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
	if !approxEq(resp.InstallToSubPct, 3.0/5.0) {
		t.Errorf("installToSubPct: expected 0.6, got %v", resp.InstallToSubPct)
	}
	if !approxEq(resp.SubToPaidPct, 2.0/3.0) {
		t.Errorf("subToPaidPct: expected 0.667, got %v", resp.SubToPaidPct)
	}
	if !approxEq(resp.OverallPct, 2.0/5.0) {
		t.Errorf("overallPct: expected 0.4, got %v", resp.OverallPct)
	}
	// Stages array mirrors the KPIs and is nested/narrowing.
	if len(resp.Stages) != 3 || resp.Stages[0].Count < resp.Stages[1].Count || resp.Stages[1].Count < resp.Stages[2].Count {
		t.Errorf("stages must be nested/narrowing, got %+v", resp.Stages)
	}
	if resp.Stages[0].Key != "installs" || resp.Stages[1].Key != "started" || resp.Stages[2].Key != "paid" {
		t.Errorf("stage keys wrong: %+v", resp.Stages)
	}
}

// TestActivation_StagesAgreeWithScalars: the Stages array and the scalar
// installs/started/paid + pcts are two representations of one funnel; pin that they never
// disagree, and that nesting holds (paid ⊆ started ⊆ installs). Guards the two-sources-of-
// truth risk the frontend depends on (it reads both).
func TestActivation_StagesAgreeWithScalars(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "s1", "s1.myshopify.com", true),
		actSub(appID, "s2", "s2.myshopify.com", true),
		actSub(appID, "s3", "s3.myshopify.com", false),
	}
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s2", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s3", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s4", "RELATIONSHIP_INSTALLED", actDay),
		acctEvt("s1", actDay),
		acctEvt("s2", actDay),
		acctEvt("s3", actDay),
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))

	if resp.Paid > resp.Started || resp.Started > resp.Installs {
		t.Fatalf("nesting violated: paid=%d started=%d installs=%d", resp.Paid, resp.Started, resp.Installs)
	}
	if len(resp.Stages) != 3 {
		t.Fatalf("expected 3 stages, got %d", len(resp.Stages))
	}
	if resp.Stages[0].Count != resp.Installs || resp.Stages[1].Count != resp.Started || resp.Stages[2].Count != resp.Paid {
		t.Errorf("stage counts disagree with scalars: stages=%+v scalars=%d/%d/%d", resp.Stages, resp.Installs, resp.Started, resp.Paid)
	}
	if !approxEq(resp.Stages[0].PctOfPrior, 1.0) ||
		!approxEq(resp.Stages[1].PctOfPrior, resp.InstallToSubPct) ||
		!approxEq(resp.Stages[2].PctOfPrior, resp.SubToPaidPct) {
		t.Errorf("stage pctOfPrior disagrees with scalar pcts: stages=%+v install→sub=%v sub→paid=%v",
			resp.Stages, resp.InstallToSubPct, resp.SubToPaidPct)
	}
}

// TestActivation_StartedNeedsAcceptedEvent: a charged subscription whose shop has NO
// SUBSCRIPTION_CHARGE_ACCEPTED event does NOT count as started (started is event-sourced,
// not derived from the subscriptions table — that's what keeps started ≠ paid).
func TestActivation_StartedNeedsAcceptedEvent(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{actSub(appID, "s1", "s1.myshopify.com", true)}
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay), // installed + charged sub, but no ACCEPTED event
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 || resp.Started != 0 || resp.Paid != 0 {
		t.Fatalf("expected installs=1 started=0 paid=0 (no accepted event), got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
	// started=0 with installs>0: SubToPaidPct divides by zero → guarded to 0.
	if resp.SubToPaidPct != 0 || resp.Stages[2].PctOfPrior != 0 {
		t.Errorf("subToPaidPct must be 0 when started=0, got %v (stage pctOfPrior %v)", resp.SubToPaidPct, resp.Stages[2].PctOfPrior)
	}
}

// TestActivation_AcceptedButNotCharged: a shop that installed + accepted but whose
// subscription never reached a recurring charge counts as started, not paid.
func TestActivation_AcceptedButNotCharged(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{actSub(appID, "s1", "s1.myshopify.com", false)} // sub exists but not charged
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay),
		acctEvt("s1", actDay),
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 || resp.Started != 1 || resp.Paid != 0 {
		t.Fatalf("expected installs=1 started=1 paid=0, got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
}

// TestActivation_CohortGating: a shop that accepted + is charged but has NO install event
// in-window is excluded entirely (the funnel is the install cohort).
func TestActivation_CohortGating(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{actSub(appID, "s1", "s1.myshopify.com", true)}
	events := []*entity.AppEvent{
		acctEvt("s1", actDay), // accepted + charged, but never an install event
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 0 || resp.Started != 0 || resp.Paid != 0 {
		t.Fatalf("expected all 0 (no install event), got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
}

// TestActivation_CanonicalisesShopIdentity: an install event keyed by shop GID and an
// accepted event keyed by domain (for the same shop, which has a subscription) collapse to
// one funnel identity → counted as started/paid.
func TestActivation_CanonicalisesShopIdentity(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{actSub(appID, "gid://shop/1", "shop1.myshopify.com", true)}
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", actDay), // GID-keyed (webhook path)
		acctEvt("shop1.myshopify.com", actDay),                       // domain-keyed (sync path)
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 || resp.Started != 1 || resp.Paid != 1 {
		t.Fatalf("expected 1/1/1 after canonicalisation, got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
}

// TestActivation_CanonCollisionDedupsInstalls: two install events for the SAME shop, one
// keyed by GID and one by domain, must dedup to a single install (not double-count). This
// guards the canonicalisation that underpins every count the report produces.
func TestActivation_CanonCollisionDedupsInstalls(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{actSub(appID, "gid://shop/1", "shop1.myshopify.com", true)}
	events := []*entity.AppEvent{
		installEvt("gid://shop/1", "RELATIONSHIP_INSTALLED", actDay),        // GID-keyed
		installEvt("shop1.myshopify.com", "RELATIONSHIP_INSTALLED", actDay), // domain-keyed, same shop
		acctEvt("gid://shop/1", actDay),
	}
	aid, pa, h := activationFixture(subs, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 || resp.Started != 1 || resp.Paid != 1 {
		t.Fatalf("two raw keys for one shop must dedup to 1 install, got installs=%d started=%d paid=%d", resp.Installs, resp.Started, resp.Paid)
	}
}

// TestActivation_AcceptedNoSubIsNotPaid: an install + accepted event for a shop with NO
// subscription row at all counts as started but never paid (paidShopKeys stays empty).
func TestActivation_AcceptedNoSubIsNotPaid(t *testing.T) {
	events := []*entity.AppEvent{
		installEvt("s9", "RELATIONSHIP_INSTALLED", actDay),
		acctEvt("s9", actDay),
	}
	aid, pa, h := activationFixture(nil, events) // no subscriptions
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 || resp.Started != 1 || resp.Paid != 0 {
		t.Fatalf("orphan accept: expected installs=1 started=1 paid=0, got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
}

// TestActivation_DateBoundaryInclusive: install on the `to` day is counted; on to+1 is not.
func TestActivation_DateBoundaryInclusive(t *testing.T) {
	onTo := time.Date(2026, 7, 31, 23, 0, 0, 0, time.UTC)
	pastTo := time.Date(2026, 8, 1, 0, 30, 0, 0, time.UTC)
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", onTo),
		installEvt("s2", "RELATIONSHIP_INSTALLED", pastTo),
	}
	aid, pa, h := activationFixture(nil, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 {
		t.Errorf("installs: expected 1 (on-`to` inclusive, to+1 excluded), got %d", resp.Installs)
	}
}

// TestActivation_DistinctShops: repeated install events for the same shop count once.
func TestActivation_DistinctShops(t *testing.T) {
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay.Add(time.Hour)),
	}
	aid, pa, h := activationFixture(nil, events)
	resp := decodeActivation(t, doActivation(t, h, aid, pa, "from=2026-07-01&to=2026-07-31"))
	if resp.Installs != 1 {
		t.Errorf("installs: expected 1 distinct shop, got %d", resp.Installs)
	}
}

// TestActivation_EmptyData: no events/subs → all zeros, all pcts 0 (no divide-by-zero),
// stages present with zero counts.
func TestActivation_EmptyData(t *testing.T) {
	aid, pa, h := activationFixture(nil, nil)
	rec := doActivation(t, h, aid, pa, "")
	resp := decodeActivation(t, rec)
	if resp.Installs != 0 || resp.Started != 0 || resp.Paid != 0 {
		t.Errorf("expected all-zero counts, got %d/%d/%d", resp.Installs, resp.Started, resp.Paid)
	}
	if resp.OverallPct != 0 || resp.InstallToSubPct != 0 || resp.SubToPaidPct != 0 {
		t.Errorf("expected all-zero pcts, got %v/%v/%v", resp.OverallPct, resp.InstallToSubPct, resp.SubToPaidPct)
	}
	if len(resp.Stages) != 3 {
		t.Errorf("expected 3 stages even when empty, got %d", len(resp.Stages))
	}
}

// TestActivation_CSV: CSV attachment lists the three stages with counts and percentages.
func TestActivation_CSV(t *testing.T) {
	appID := uuid.New()
	subs := []*entity.Subscription{
		actSub(appID, "s1", "s1.myshopify.com", true),
		actSub(appID, "s2", "s2.myshopify.com", false),
	}
	events := []*entity.AppEvent{
		installEvt("s1", "RELATIONSHIP_INSTALLED", actDay),
		installEvt("s2", "RELATIONSHIP_INSTALLED", actDay),
		acctEvt("s1", actDay),
		acctEvt("s2", actDay),
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
	if len(rows) != 4 { // header + 3 stages
		t.Fatalf("expected 4 CSV rows, got %d: %v", len(rows), rows)
	}
	if rows[0][0] != "stage" || rows[1][0] != "Installs" || rows[3][0] != "Paid / Recurring" {
		t.Errorf("unexpected CSV shape: %v", rows)
	}
	if rows[1][1] != "2" || rows[2][1] != "2" || rows[3][1] != "1" {
		t.Errorf("CSV counts wrong (want installs=2 started=2 paid=1): %v", rows)
	}
	// pctOfPrior (col 2) + pctOfInstalls (col 3): installs=1.0/1.0; paid=0.5 of started, 0.5 of installs.
	if rows[1][2] != "1.0000" || rows[1][3] != "1.0000" {
		t.Errorf("installs row pct wrong: %v", rows[1])
	}
	if rows[3][2] != "0.5000" || rows[3][3] != "0.5000" {
		t.Errorf("paid row pct wrong (want pctOfPrior=0.5 pctOfInstalls=0.5): %v", rows[3])
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

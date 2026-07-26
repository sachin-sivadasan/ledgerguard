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

// cohortSub builds a subscription with explicit created/updated timestamps and risk
// state, for deterministic cohort tests. buildCohorts groups by CreatedAt month; for
// churned subs isActiveAt prefers churnedDateOf (ExpectedNextChargeDate/
// LastRecurringChargeDate) and falls back to UpdatedAt only when those are unset.
func cohortSub(appID uuid.UUID, created, updated time.Time, state valueobject.RiskState) *entity.Subscription {
	return &entity.Subscription{
		ID:              uuid.New(),
		AppID:           appID,
		MyshopifyDomain: "shop.myshopify.com",
		Currency:        "USD",
		RiskState:       state,
		CreatedAt:       created,
		UpdatedAt:       updated,
	}
}

func month(year int, m time.Month) time.Time {
	return time.Date(year, m, 1, 0, 0, 0, 0, time.UTC)
}

// --- buildCohorts unit tests (deterministic via injected `now`) ---

// TestBuildCohorts_GroupsAndM0IsFull verifies subs are grouped by creation month and
// M0 retention is always 100%.
func TestBuildCohorts_GroupsAndM0IsFull(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	// All created on the 1st so M0 (checked at the cohort month's 1st) captures them.
	subs := []*entity.Subscription{
		cohortSub(appID, month(2026, 5), now, valueobject.RiskStateSafe),
		cohortSub(appID, month(2026, 5), now, valueobject.RiskStateSafe),
		cohortSub(appID, month(2026, 6), now, valueobject.RiskStateSafe),
	}

	cohorts := buildCohorts(subs, 6, now)
	if len(cohorts) != 2 {
		t.Fatalf("expected 2 cohorts (May, June), got %d: %+v", len(cohorts), cohorts)
	}
	// Chronological order: May first.
	if cohorts[0].CohortMonth != "2026-05" || cohorts[0].InitialStores != 2 {
		t.Errorf("cohort[0]: expected 2026-05 with 2 stores, got %+v", cohorts[0])
	}
	if cohorts[1].CohortMonth != "2026-06" || cohorts[1].InitialStores != 1 {
		t.Errorf("cohort[1]: expected 2026-06 with 1 store, got %+v", cohorts[1])
	}
	// M0 is 100% for every cohort.
	for _, c := range cohorts {
		if len(c.RetentionPcts) == 0 || c.RetentionPcts[0] != 100 {
			t.Errorf("cohort %s: expected M0=100, got %+v", c.CohortMonth, c.RetentionPcts)
		}
	}
}

// TestBuildCohorts_ChurnedDropsOutAtRightMonth verifies a churned sub counts as
// retained up to its (deterministic) churn date, then drops the retention pct. The
// churn date is taken from ExpectedNextChargeDate, not UpdatedAt.
func TestBuildCohorts_ChurnedDropsOutAtRightMonth(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	cohortMonth := month(2026, 5)
	// Churned; expected charge (churn date) mid-June → active at M1 (June 1), gone by
	// M2 (July 1). UpdatedAt is set to `now` to prove it is NOT used as the churn date.
	churned := cohortSub(appID, cohortMonth, now, valueobject.RiskStateChurned)
	churnDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	churned.ExpectedNextChargeDate = &churnDate
	subs := []*entity.Subscription{
		cohortSub(appID, cohortMonth, now, valueobject.RiskStateSafe), // stays safe
		churned,
	}

	cohorts := buildCohorts(subs, 6, now)
	if len(cohorts) != 1 {
		t.Fatalf("expected 1 cohort, got %d", len(cohorts))
	}
	pcts := cohorts[0].RetentionPcts
	// M0: cohort baseline → 100. M1 = June 1: churn June 15 is after → 100.
	// M2 = July 1: churn June 15 is before → only the safe sub → 50.
	if len(pcts) < 3 {
		t.Fatalf("expected >=3 retention points, got %+v", pcts)
	}
	if pcts[0] != 100 {
		t.Errorf("M0: expected 100, got %v", pcts[0])
	}
	if pcts[1] != 100 {
		t.Errorf("M1: expected 100 (churn after M1), got %v", pcts[1])
	}
	if pcts[2] != 50 {
		t.Errorf("M2: expected 50 (churned sub dropped), got %v", pcts[2])
	}
}

// TestBuildCohorts_M0FullForMidMonthCreated is the regression guard for the M0 fix: a
// sub created mid-month (the realistic case) must still yield M0 = 100% (cohort
// baseline), not ~0% from measuring activity at the month's 1st instant.
func TestBuildCohorts_M0FullForMidMonthCreated(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	subs := []*entity.Subscription{
		cohortSub(appID, time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC), now, valueobject.RiskStateSafe),
		cohortSub(appID, time.Date(2026, 5, 28, 9, 30, 0, 0, time.UTC), now, valueobject.RiskStateSafe),
	}
	cohorts := buildCohorts(subs, 6, now)
	if len(cohorts) != 1 {
		t.Fatalf("expected 1 cohort, got %d", len(cohorts))
	}
	if cohorts[0].RetentionPcts[0] != 100 {
		t.Errorf("M0 for mid-month-created subs: expected 100, got %v", cohorts[0].RetentionPcts[0])
	}
}

// TestBuildCohorts_ChurnFallsBackToUpdatedAt verifies that when a churned sub has no
// charge dates, isActiveAt falls back to UpdatedAt as the churn instant.
func TestBuildCohorts_ChurnFallsBackToUpdatedAt(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	cohortMonth := month(2026, 5)
	// No ExpectedNextChargeDate/LastRecurringChargeDate → falls back to UpdatedAt (June 15).
	churned := cohortSub(appID, cohortMonth, time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), valueobject.RiskStateChurned)
	subs := []*entity.Subscription{
		cohortSub(appID, cohortMonth, now, valueobject.RiskStateSafe),
		churned,
	}
	cohorts := buildCohorts(subs, 6, now)
	pcts := cohorts[0].RetentionPcts
	if len(pcts) < 3 || pcts[1] != 100 || pcts[2] != 50 {
		t.Errorf("fallback-to-UpdatedAt churn: expected M1=100, M2=50, got %+v", pcts)
	}
}

// TestBuildCohorts_ExcludesOlderThanCutoff verifies subs created before the window
// cutoff are excluded.
func TestBuildCohorts_ExcludesOlderThanCutoff(t *testing.T) {
	appID := uuid.New()
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	// months=6 → cutoff = 2026-02 (Feb..Jul window).
	subs := []*entity.Subscription{
		cohortSub(appID, month(2026, 1), now, valueobject.RiskStateSafe), // Jan → before cutoff, excluded
		cohortSub(appID, month(2026, 2), now, valueobject.RiskStateSafe), // Feb → included (cutoff month)
		cohortSub(appID, month(2026, 7), now, valueobject.RiskStateSafe), // Jul → included
	}

	cohorts := buildCohorts(subs, 6, now)
	if len(cohorts) != 2 {
		t.Fatalf("expected 2 cohorts (Feb, Jul), got %d: %+v", len(cohorts), cohorts)
	}
	if cohorts[0].CohortMonth != "2026-02" {
		t.Errorf("expected first cohort 2026-02, got %s", cohorts[0].CohortMonth)
	}
}

// TestBuildCohorts_Empty verifies no subscriptions yields an empty (non-nil) slice.
func TestBuildCohorts_Empty(t *testing.T) {
	now := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	cohorts := buildCohorts(nil, 6, now)
	if cohorts == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(cohorts) != 0 {
		t.Errorf("expected empty, got %+v", cohorts)
	}
}

// --- GetCohorts end-to-end tests (httptest) ---

func cohortFixture(subs []*entity.Subscription) (uuid.UUID, *entity.PartnerAccount, *CohortHandler) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewCohortHandler(
		&mockSubscriptionRepo{subscriptions: subs},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	return appID, pa, h
}

func doCohorts(t *testing.T, h *CohortHandler, appID uuid.UUID, pa *entity.PartnerAccount, query string) *httptest.ResponseRecorder {
	t.Helper()
	url := "/api/v1/apps/" + appID.String() + "/cohorts"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.GetCohorts(rec, req)
	return rec
}

type cohortsResponse struct {
	Cohorts []entity.CohortData `json:"cohorts"`
}

// TestGetCohorts_JSON verifies the end-to-end JSON contract.
func TestGetCohorts_JSON(t *testing.T) {
	appID, pa, h := cohortFixture(nil)
	// Rebuild fixture with app-scoped subs (need appID from fixture).
	subs := []*entity.Subscription{
		cohortSub(appID, month(2026, 5), time.Now().UTC(), valueobject.RiskStateSafe),
	}
	h = replaceCohortSubs(h, subs)

	rec := doCohorts(t, h, appID, pa, "months=6")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp cohortsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Cohorts) != 1 || resp.Cohorts[0].CohortMonth != "2026-05" {
		t.Errorf("expected one 2026-05 cohort, got %+v", resp.Cohorts)
	}
}

// replaceCohortSubs swaps the subscription repo so a fixture can be built with subs
// scoped to the fixture's appID.
func replaceCohortSubs(h *CohortHandler, subs []*entity.Subscription) *CohortHandler {
	h.subscriptionRepo = &mockSubscriptionRepo{subscriptions: subs}
	return h
}

// TestGetCohorts_MonthsClampDefault verifies the months param clamps to [2,24] with a
// default of 6 and out-of-range values ignored (falling back to default).
func TestGetCohorts_MonthsClampDefault(t *testing.T) {
	appID, pa, h := cohortFixture(nil)
	// One sub 12 months back and one 3 months back; with default 6 only the recent one
	// is in-window, proving the out-of-range param was ignored (kept default 6).
	now := time.Now().UTC()
	subs := []*entity.Subscription{
		cohortSub(appID, now.AddDate(0, -12, 0), now, valueobject.RiskStateSafe),
		cohortSub(appID, now.AddDate(0, -3, 0), now, valueobject.RiskStateSafe),
	}
	h = replaceCohortSubs(h, subs)

	// months=100 is out of range → ignored → default 6 → only the -3mo sub qualifies.
	rec := doCohorts(t, h, appID, pa, "months=100")
	var resp cohortsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Cohorts) != 1 {
		t.Errorf("months=100 (ignored, default 6): expected 1 in-window cohort, got %d: %+v", len(resp.Cohorts), resp.Cohorts)
	}

	// months=24 is valid → both subs qualify.
	rec = doCohorts(t, h, appID, pa, "months=24")
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Cohorts) != 2 {
		t.Errorf("months=24: expected 2 cohorts, got %d: %+v", len(resp.Cohorts), resp.Cohorts)
	}
}

// TestGetCohorts_MonthsLowerBoundAndInvalid verifies a below-min (months=1) and a
// non-numeric (months=abc) param both fall back to the default window of 6.
func TestGetCohorts_MonthsLowerBoundAndInvalid(t *testing.T) {
	appID, pa, h := cohortFixture(nil)
	now := time.Now().UTC()
	subs := []*entity.Subscription{
		cohortSub(appID, now.AddDate(0, -3, 0), now, valueobject.RiskStateSafe),  // within default-6
		cohortSub(appID, now.AddDate(0, -12, 0), now, valueobject.RiskStateSafe), // outside default-6
	}
	h = replaceCohortSubs(h, subs)

	for _, q := range []string{"months=1", "months=abc"} {
		rec := doCohorts(t, h, appID, pa, q)
		var resp cohortsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("%s decode: %v", q, err)
		}
		// Default 6 → only the -3mo sub is in-window; proves the bad param was ignored.
		if len(resp.Cohorts) != 1 {
			t.Errorf("%s (should fall back to default 6): expected 1 cohort, got %d: %+v", q, len(resp.Cohorts), resp.Cohorts)
		}
	}
}

// TestGetCohorts_EmptyList verifies no subscriptions serializes as {"cohorts":[]}.
func TestGetCohorts_EmptyList(t *testing.T) {
	appID, pa, h := cohortFixture(nil)
	rec := doCohorts(t, h, appID, pa, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"cohorts":[]`) {
		t.Errorf("expected cohorts serialized as [], body: %s", rec.Body.String())
	}
}

// TestGetCohorts_RepoErrorReturns503 verifies subscription-repo failures surface as 503
// (ADR-042), the key new behavior of this hardening.
func TestGetCohorts_RepoErrorReturns503(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewCohortHandler(
		&mockSubscriptionRepo{findAllErr: errors.New("db down")},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	rec := doCohorts(t, h, appID, pa, "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestGetCohorts_CSVFormat verifies CSV output: header (cohort,initialStores,M0..Mmax),
// one row per cohort with retention pcts formatted to 1 decimal, and ragged shorter-lived
// cohorts padded so every row has the same column count.
func TestGetCohorts_CSVFormat(t *testing.T) {
	appID, pa, h := cohortFixture(nil)
	// Two cohorts of different ages → different-length retention rows (ragged).
	// Snap created dates to the 1st of their month so M0 (checked at the cohort
	// month's 1st) is 100%.
	realNow := time.Now().UTC()
	older := realNow.AddDate(0, -5, 0)
	newer := realNow.AddDate(0, -1, 0)
	subs := []*entity.Subscription{
		// Older cohort (~5 months) → longer retention row.
		cohortSub(appID, month(older.Year(), older.Month()), realNow, valueobject.RiskStateSafe),
		// Newer cohort (~1 month) → shorter retention row.
		cohortSub(appID, month(newer.Year(), newer.Month()), realNow, valueobject.RiskStateSafe),
	}
	h = replaceCohortSubs(h, subs)

	rec := doCohorts(t, h, appID, pa, "months=24&format=csv")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "cohorts.csv") {
		t.Errorf("expected filename cohorts.csv in Content-Disposition, got %q", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	// Header starts with cohort,initialStores,M0...
	if records[0][0] != "cohort" || records[0][1] != "initialStores" || records[0][2] != "M0" {
		t.Errorf("header prefix wrong: %v", records[0])
	}
	// Column integrity: every row (including the padded newer cohort) has the same
	// column count as the header.
	cols := len(records[0])
	for i, row := range records {
		if len(row) != cols {
			t.Errorf("row %d: expected %d columns, got %d: %v", i, cols, len(row), row)
		}
	}
	// M0 formatted to 1 decimal.
	if records[1][2] != "100.0" {
		t.Errorf("M0 format: expected 100.0, got %q", records[1][2])
	}
	// The newer (shorter) cohort must have trailing empty pads in its later M columns.
	last := records[len(records)-1]
	if last[cols-1] != "" {
		t.Errorf("expected trailing pad empty for ragged newer cohort, got %q", last[cols-1])
	}
}

// TestGetCohorts_CSVEmpty verifies the CSV path with no cohorts writes just a header
// (cohort,initialStores with no M columns) and no data rows.
func TestGetCohorts_CSVEmpty(t *testing.T) {
	appID, pa, h := cohortFixture(nil)
	rec := doCohorts(t, h, appID, pa, "format=csv")
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected header only, got %d rows: %v", len(records), records)
	}
	if len(records[0]) != 2 || records[0][0] != "cohort" || records[0][1] != "initialStores" {
		t.Errorf("expected header [cohort initialStores], got %v", records[0])
	}
}

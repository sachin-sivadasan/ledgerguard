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

// payoutScheduleFixture wires a PayoutScheduleReportHandler with the shared mock repos
// and returns a ready-to-serve authenticated GET request.
func payoutScheduleFixture(t *testing.T, txs []*entity.Transaction, findErr error, query string) (*PayoutScheduleReportHandler, *http.Request) {
	t.Helper()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewPayoutScheduleReportHandler(
		&mockTxRepo{transactions: txs, findErr: findErr},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	url := "/api/v1/apps/" + appID.String() + "/reports/payout-schedule"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return h, req
}

func servePayoutSchedule(t *testing.T, txs []*entity.Transaction, query string) (*httptest.ResponseRecorder, payoutScheduleReport) {
	t.Helper()
	h, req := payoutScheduleFixture(t, txs, nil, query)
	rec := httptest.NewRecorder()
	h.GetPayoutSchedule(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var report payoutScheduleReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	return rec, report
}

var (
	psAvail1 = time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	psAvail2 = time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
	psAvail3 = time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	psCreate = time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
)

// psTx builds a transaction carrying the fields the payout schedule reads (reuses the
// earnings builder for the net/status/available-date shape).
func psTx(net int64, status entity.EarningsStatus, available time.Time) *entity.Transaction {
	return earningsTx("shop.myshopify.com", "Shop", "USD", net+100, net, status, psCreate, available)
}

// TestPayoutSchedule_KPIsAndGrouping verifies the AVAILABLE/PENDING KPIs, per-(date,
// status) grouping with summed amounts + counts, PAID_OUT exclusion, and NextPayoutDate.
func TestPayoutSchedule_KPIsAndGrouping(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(2000, entity.EarningsStatusAvailable, psAvail1),
		psTx(980, entity.EarningsStatusAvailable, psAvail1), // same date+status → one row
		psTx(1240, entity.EarningsStatusPending, psAvail2),
		psTx(610, entity.EarningsStatusPending, psAvail3),
		psTx(5000, entity.EarningsStatusPaidOut, psAvail1), // excluded (Payout History)
	}
	_, report := servePayoutSchedule(t, txs, "")

	if report.UpcomingPayoutCents != 2980 { // 2000 + 980 available
		t.Errorf("upcomingPayoutCents = %d, want 2980", report.UpcomingPayoutCents)
	}
	if report.PendingCents != 1850 { // 1240 + 610 pending
		t.Errorf("pendingCents = %d, want 1850", report.PendingCents)
	}
	if report.NextPayoutDate != "2026-07-30" {
		t.Errorf("nextPayoutDate = %q, want 2026-07-30", report.NextPayoutDate)
	}
	if len(report.Rows) != 3 { // paid-out excluded; two available collapse to one row
		t.Fatalf("expected 3 rows, got %d: %+v", len(report.Rows), report.Rows)
	}
	first := report.Rows[0]
	if first.AvailableDate != "2026-07-30" || first.AmountCents != 2980 || first.ChargeCount != 2 || first.Status != "Available" {
		t.Errorf("row[0] = %+v, want {2026-07-30, 2980, 2, Available}", first)
	}
}

// TestPayoutSchedule_Reconciles verifies upcoming + pending equals the sum of all row
// amounts (no phantom row outside a KPI, no KPI money missing from the rows).
func TestPayoutSchedule_Reconciles(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(2000, entity.EarningsStatusAvailable, psAvail1),
		psTx(1240, entity.EarningsStatusPending, psAvail2),
		psTx(610, entity.EarningsStatusPending, psAvail3),
		psTx(5000, entity.EarningsStatusPaidOut, psAvail1),
	}
	_, report := servePayoutSchedule(t, txs, "")
	var rowSum int64
	for _, r := range report.Rows {
		rowSum += r.AmountCents
	}
	if rowSum != report.UpcomingPayoutCents+report.PendingCents {
		t.Errorf("row sum %d != upcoming(%d)+pending(%d)", rowSum, report.UpcomingPayoutCents, report.PendingCents)
	}
}

// TestPayoutSchedule_UnknownStatusExcludedAndReconciles verifies an unrecognized status
// is dropped from KPIs and rows, keeping reconciliation intact.
func TestPayoutSchedule_UnknownStatusExcludedAndReconciles(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(2000, entity.EarningsStatusAvailable, psAvail1),
		psTx(1240, entity.EarningsStatusPending, psAvail2),
		psTx(9999, entity.EarningsStatus("DISPUTED"), psAvail3),
		psTx(8888, entity.EarningsStatus(""), psAvail3),
	}
	_, report := servePayoutSchedule(t, txs, "")
	if report.UpcomingPayoutCents != 2000 || report.PendingCents != 1240 {
		t.Errorf("KPIs = upcoming %d / pending %d, want 2000 / 1240 (unknown excluded)", report.UpcomingPayoutCents, report.PendingCents)
	}
	if len(report.Rows) != 2 {
		t.Fatalf("expected 2 rows (unknown excluded), got %d", len(report.Rows))
	}
	var rowSum int64
	for _, r := range report.Rows {
		rowSum += r.AmountCents
	}
	if rowSum != report.UpcomingPayoutCents+report.PendingCents {
		t.Errorf("reconciliation broken: row sum %d != %d", rowSum, report.UpcomingPayoutCents+report.PendingCents)
	}
}

// TestPayoutSchedule_SortAvailableBeforePendingSameDate verifies rows sharing a date are
// ordered Available (sooner-paying) before Pending.
func TestPayoutSchedule_SortAvailableBeforePendingSameDate(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(500, entity.EarningsStatusPending, psAvail1),
		psTx(1000, entity.EarningsStatusAvailable, psAvail1),
	}
	_, report := servePayoutSchedule(t, txs, "")
	if len(report.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(report.Rows))
	}
	if report.Rows[0].Status != "Available" || report.Rows[1].Status != "Pending" {
		t.Errorf("same-date order = [%s, %s], want [Available, Pending]", report.Rows[0].Status, report.Rows[1].Status)
	}
	if report.Rows[0].AvailableDate != "2026-07-30" || report.Rows[1].AvailableDate != "2026-07-30" {
		t.Errorf("both rows should be 2026-07-30, got %q / %q", report.Rows[0].AvailableDate, report.Rows[1].AvailableDate)
	}
}

// TestPayoutSchedule_UnscheduledRowSinksLast verifies a row with a zero AvailableDate
// serializes "" and sorts to the bottom, and that NextPayoutDate ignores it.
func TestPayoutSchedule_UnscheduledRowSinksLast(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(700, entity.EarningsStatusAvailable, time.Time{}), // no available date
		psTx(500, entity.EarningsStatusPending, psAvail1),
	}
	_, report := servePayoutSchedule(t, txs, "")
	if len(report.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(report.Rows))
	}
	last := report.Rows[len(report.Rows)-1]
	if last.AvailableDate != "" {
		t.Errorf("unscheduled row should sink last with empty date, got %q", last.AvailableDate)
	}
	if report.NextPayoutDate != "2026-07-30" {
		t.Errorf("nextPayoutDate = %q, want 2026-07-30 (unscheduled row ignored)", report.NextPayoutDate)
	}
}

// TestPayoutSchedule_CurrencyDefaultAndFirstNonEmpty covers the currency fallbacks.
func TestPayoutSchedule_CurrencyDefaultAndFirstNonEmpty(t *testing.T) {
	// Default USD when none carry a currency.
	noCur := earningsTx("a.myshopify.com", "Acme", "", 100, 90, entity.EarningsStatusPending, psCreate, psAvail1)
	_, report := servePayoutSchedule(t, []*entity.Transaction{noCur}, "")
	if report.Currency != "USD" {
		t.Errorf("currency = %s, want USD", report.Currency)
	}

	// First non-empty wins.
	eur := earningsTx("b.myshopify.com", "Beta", "EUR", 200, 180, entity.EarningsStatusAvailable, psCreate, psAvail1)
	_, report2 := servePayoutSchedule(t, []*entity.Transaction{noCur, eur}, "")
	if report2.Currency != "EUR" {
		t.Errorf("currency = %s, want EUR", report2.Currency)
	}
}

// TestPayoutSchedule_EmptyCase verifies zero KPIs, empty NextPayoutDate and a
// []-serialized rows slice.
func TestPayoutSchedule_EmptyCase(t *testing.T) {
	rec, report := servePayoutSchedule(t, []*entity.Transaction{}, "")
	if report.UpcomingPayoutCents != 0 || report.PendingCents != 0 || report.NextPayoutDate != "" {
		t.Errorf("expected zero KPIs and empty nextPayoutDate, got %+v", report)
	}
	if !strings.Contains(rec.Body.String(), `"rows":[]`) {
		t.Errorf("expected rows serialized as [], body: %s", rec.Body.String())
	}
}

// TestPayoutSchedule_OnlyPaidOutYieldsEmpty verifies that when every charge is already
// paid out, the schedule is empty (all belong to Payout History).
func TestPayoutSchedule_OnlyPaidOutYieldsEmpty(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(2000, entity.EarningsStatusPaidOut, psAvail1),
		psTx(1000, entity.EarningsStatusPaidOut, psAvail2),
	}
	_, report := servePayoutSchedule(t, txs, "")
	if report.UpcomingPayoutCents != 0 || report.PendingCents != 0 || len(report.Rows) != 0 {
		t.Errorf("expected empty schedule for all-paid-out, got %+v", report)
	}
}

// TestPayoutSchedule_RepoErrorReturns503 verifies transaction-repo failures surface as 503.
func TestPayoutSchedule_RepoErrorReturns503(t *testing.T) {
	h, req := payoutScheduleFixture(t, nil, errors.New("db down"), "")
	rec := httptest.NewRecorder()
	h.GetPayoutSchedule(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

// TestPayoutSchedule_Unauthenticated verifies a missing user yields 401.
func TestPayoutSchedule_Unauthenticated(t *testing.T) {
	h, req := payoutScheduleFixture(t, nil, nil, "")
	req = httptest.NewRequest(http.MethodGet, req.URL.String(), nil)
	req = withURLParam(req, "appID", uuid.New().String())
	rec := httptest.NewRecorder()
	h.GetPayoutSchedule(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestPayoutSchedule_CSVFormat verifies the CSV header, one row per group, and asc sort.
func TestPayoutSchedule_CSVFormat(t *testing.T) {
	txs := []*entity.Transaction{
		psTx(1240, entity.EarningsStatusPending, psAvail2),
		psTx(2000, entity.EarningsStatusAvailable, psAvail1),
	}
	h, req := payoutScheduleFixture(t, txs, nil, "format=csv")
	rec := httptest.NewRecorder()
	h.GetPayoutSchedule(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "payout-schedule.csv") {
		t.Errorf("Content-Disposition = %q, want payout-schedule.csv", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"availableDate", "amountCents", "chargeCount", "status"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], want)
		}
	}
	// Sorted ascending by date: 2026-07-30 (Available) before 2026-08-06 (Pending).
	if records[1][0] != "2026-07-30" || records[1][3] != "Available" {
		t.Errorf("row[1] = %v, want 2026-07-30/Available first", records[1])
	}
}

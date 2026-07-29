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

// payoutHistoryFixture wires a PayoutHistoryReportHandler with the shared mock repos
// and returns a ready-to-serve authenticated GET request.
func payoutHistoryFixture(t *testing.T, txs []*entity.Transaction, findErr error, query string) (*PayoutHistoryReportHandler, *http.Request) {
	t.Helper()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewPayoutHistoryReportHandler(
		&mockTxRepo{transactions: txs, findErr: findErr},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
	)
	url := "/api/v1/apps/" + appID.String() + "/reports/payout-history"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return h, req
}

func servePayoutHistory(t *testing.T, txs []*entity.Transaction, query string) (*httptest.ResponseRecorder, payoutHistoryReport) {
	t.Helper()
	h, req := payoutHistoryFixture(t, txs, nil, query)
	rec := httptest.NewRecorder()
	h.GetPayoutHistory(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var report payoutHistoryReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	return rec, report
}

var (
	phJun1      = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	phJun2      = time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	phMay1      = time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	phApr1      = time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	phPaidJul5  = time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
	phPaidJul15 = time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	phPaidJun5  = time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	phIngestJul = time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC) // sync/ingestion time (CreatedAt)
)

// phTx builds a PAID_OUT-shaped transaction (reuses the earnings builder), with a
// caller-supplied charge month (set as CreatedDate — the field the report groups by,
// NOT CreatedAt) and availability date.
func phTx(net int64, status entity.EarningsStatus, charge, available time.Time) *entity.Transaction {
	tx := earningsTx("shop.myshopify.com", "Shop", "USD", net+100, net, status, charge, available)
	tx.CreatedDate = charge // the Shopify charge date the period grouping keys off
	return tx
}

// TestPayoutHistory_KPIsAndPeriodGrouping verifies Total/Count/Avg and per-month grouping
// with summed amounts + counts, and that PaidDate is the latest available date in a period.
func TestPayoutHistory_KPIsAndPeriodGrouping(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(2000, entity.EarningsStatusPaidOut, phJun1, phPaidJul5),
		phTx(1120, entity.EarningsStatusPaidOut, phJun2, phPaidJul15), // same period, later paid date
		phTx(2780, entity.EarningsStatusPaidOut, phMay1, phPaidJun5),
		phTx(9999, entity.EarningsStatusPending, phJun1, phPaidJul5), // not paid → excluded
	}
	_, report := servePayoutHistory(t, txs, "")

	if report.TotalPaidCents != 5900 { // 3120 (Jun) + 2780 (May)
		t.Errorf("totalPaidCents = %d, want 5900", report.TotalPaidCents)
	}
	if report.PayoutCount != 2 {
		t.Errorf("payoutCount = %d, want 2", report.PayoutCount)
	}
	if report.AvgPayoutCents != 2950 { // 5900 / 2
		t.Errorf("avgPayoutCents = %d, want 2950", report.AvgPayoutCents)
	}
	if len(report.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d: %+v", len(report.Rows), report.Rows)
	}
	jun := report.Rows[0]
	if jun.Period != "2026-06" || jun.AmountCents != 3120 || jun.ChargeCount != 2 {
		t.Errorf("row[0] = %+v, want {2026-06, 3120, 2}", jun)
	}
	if jun.AvailableDate != "2026-07-15" { // latest availability date in the period
		t.Errorf("row[0].availableDate = %q, want 2026-07-15 (latest in period)", jun.AvailableDate)
	}
}

// TestPayoutHistory_GroupsByChargeDateNotIngestion pins that periods key off the Shopify
// charge date (CreatedDate), NOT the row-ingestion time (CreatedAt). A charge billed in
// May but synced (CreatedAt) in July must land in the May period — reverting the grouping
// to CreatedAt would fail this.
func TestPayoutHistory_GroupsByChargeDateNotIngestion(t *testing.T) {
	tx := earningsTx("shop.myshopify.com", "Shop", "USD", 2100, 2000, entity.EarningsStatusPaidOut, phIngestJul, phPaidJul5)
	tx.CreatedDate = phMay1 // real Shopify charge date (May); CreatedAt above is the July sync time
	_, report := servePayoutHistory(t, []*entity.Transaction{tx}, "")
	if len(report.Rows) != 1 || report.Rows[0].Period != "2026-05" {
		t.Fatalf("expected the charge in its CreatedDate month 2026-05 (not the CreatedAt sync month 2026-07), got %+v", report.Rows)
	}
}

// TestPayoutHistory_ChargeDateFallsBackToTransactionDate pins the TransactionDate fallback
// when CreatedDate is unset.
func TestPayoutHistory_ChargeDateFallsBackToTransactionDate(t *testing.T) {
	tx := earningsTx("shop.myshopify.com", "Shop", "USD", 2100, 2000, entity.EarningsStatusPaidOut, phIngestJul, phPaidJul5)
	// CreatedDate zero; TransactionDate set to April → period comes from TransactionDate.
	tx.TransactionDate = phApr1
	_, report := servePayoutHistory(t, []*entity.Transaction{tx}, "")
	if len(report.Rows) != 1 || report.Rows[0].Period != "2026-04" {
		t.Fatalf("expected fallback to TransactionDate month 2026-04, got %+v", report.Rows)
	}
}

// TestPayoutHistory_SortPeriodDescending verifies rows are ordered most-recent period first.
func TestPayoutHistory_SortPeriodDescending(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(100, entity.EarningsStatusPaidOut, phApr1, phPaidJun5),
		phTx(200, entity.EarningsStatusPaidOut, phJun1, phPaidJul5),
		phTx(300, entity.EarningsStatusPaidOut, phMay1, phPaidJun5),
	}
	_, report := servePayoutHistory(t, txs, "")
	wantOrder := []string{"2026-06", "2026-05", "2026-04"}
	if len(report.Rows) != len(wantOrder) {
		t.Fatalf("expected %d rows, got %d", len(wantOrder), len(report.Rows))
	}
	for i, want := range wantOrder {
		if report.Rows[i].Period != want {
			t.Errorf("row[%d].period = %q, want %q (desc)", i, report.Rows[i].Period, want)
		}
	}
}

// TestPayoutHistory_OnlyPaidOutCounted verifies PENDING/AVAILABLE are excluded from
// KPIs and rows (they belong to Payout Schedule).
func TestPayoutHistory_OnlyPaidOutCounted(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(2000, entity.EarningsStatusPaidOut, phJun1, phPaidJul5),
		phTx(9999, entity.EarningsStatusPending, phJun1, phPaidJul5),
		phTx(8888, entity.EarningsStatusAvailable, phMay1, phPaidJun5),
	}
	_, report := servePayoutHistory(t, txs, "")
	if report.TotalPaidCents != 2000 || report.PayoutCount != 1 || len(report.Rows) != 1 {
		t.Errorf("expected only the paid-out tx counted (2000/1/1 row), got %+v", report)
	}
}

// TestPayoutHistory_PaidDateEmptyWhenNoAvailableDate verifies a paid tx with a zero
// available date yields an empty PaidDate (UI renders "—") without dropping the money.
func TestPayoutHistory_PaidDateEmptyWhenNoAvailableDate(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(2000, entity.EarningsStatusPaidOut, phJun1, time.Time{}),
	}
	_, report := servePayoutHistory(t, txs, "")
	if len(report.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(report.Rows))
	}
	if report.Rows[0].AvailableDate != "" {
		t.Errorf("availableDate = %q, want \"\" (zero availability date)", report.Rows[0].AvailableDate)
	}
	if report.TotalPaidCents != 2000 { // money still counted
		t.Errorf("totalPaidCents = %d, want 2000 (money not dropped)", report.TotalPaidCents)
	}
}

// TestPayoutHistory_AvgPayoutFloored verifies Avg = Total ÷ count using integer floor.
func TestPayoutHistory_AvgPayoutFloored(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(1000, entity.EarningsStatusPaidOut, phJun1, phPaidJul5),
		phTx(1000, entity.EarningsStatusPaidOut, phMay1, phPaidJun5),
		phTx(1001, entity.EarningsStatusPaidOut, phApr1, phPaidJun5),
	}
	_, report := servePayoutHistory(t, txs, "")
	// Total 3001 / 3 periods = 1000.33 → floored to 1000.
	if report.AvgPayoutCents != 1000 {
		t.Errorf("avgPayoutCents = %d, want 1000 (floored)", report.AvgPayoutCents)
	}
}

// TestPayoutHistory_AvgPayoutFloorsNotRounds pins that Avg FLOORS rather than rounding:
// 2001 ÷ 2 = 1000.5 floors to 1000 but rounds to 1001 — catches a round mutation.
func TestPayoutHistory_AvgPayoutFloorsNotRounds(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(1000, entity.EarningsStatusPaidOut, phJun1, phPaidJul5),
		phTx(1001, entity.EarningsStatusPaidOut, phMay1, phPaidJun5),
	}
	_, report := servePayoutHistory(t, txs, "")
	// Total 2001 / 2 periods = 1000.5 → floor 1000 (round would give 1001).
	if report.AvgPayoutCents != 1000 {
		t.Errorf("avgPayoutCents = %d, want 1000 (floored, not rounded to 1001)", report.AvgPayoutCents)
	}
}

// TestPayoutHistory_CurrencyDefaultAndFirstNonEmpty covers the currency fallbacks.
func TestPayoutHistory_CurrencyDefaultAndFirstNonEmpty(t *testing.T) {
	noCur := earningsTx("a.myshopify.com", "Acme", "", 100, 90, entity.EarningsStatusPaidOut, phJun1, phPaidJul5)
	_, report := servePayoutHistory(t, []*entity.Transaction{noCur}, "")
	if report.Currency != "USD" {
		t.Errorf("currency = %s, want USD", report.Currency)
	}

	eur := earningsTx("b.myshopify.com", "Beta", "EUR", 200, 180, entity.EarningsStatusPaidOut, phMay1, phPaidJun5)
	_, report2 := servePayoutHistory(t, []*entity.Transaction{noCur, eur}, "")
	if report2.Currency != "EUR" {
		t.Errorf("currency = %s, want EUR", report2.Currency)
	}
}

// TestPayoutHistory_EmptyCase verifies zero KPIs and a []-serialized rows slice.
func TestPayoutHistory_EmptyCase(t *testing.T) {
	rec, report := servePayoutHistory(t, []*entity.Transaction{}, "")
	if report.TotalPaidCents != 0 || report.PayoutCount != 0 || report.AvgPayoutCents != 0 {
		t.Errorf("expected zero KPIs, got %+v", report)
	}
	if !strings.Contains(rec.Body.String(), `"rows":[]`) {
		t.Errorf("expected rows serialized as [], body: %s", rec.Body.String())
	}
}

// TestPayoutHistory_NoPaidOutYieldsEmpty verifies an all-upcoming set produces an empty
// history with a zero (not divide-by-zero) average.
func TestPayoutHistory_NoPaidOutYieldsEmpty(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(2000, entity.EarningsStatusPending, phJun1, phPaidJul5),
		phTx(1000, entity.EarningsStatusAvailable, phMay1, phPaidJun5),
	}
	_, report := servePayoutHistory(t, txs, "")
	if report.TotalPaidCents != 0 || report.PayoutCount != 0 || report.AvgPayoutCents != 0 || len(report.Rows) != 0 {
		t.Errorf("expected empty history for no-paid-out, got %+v", report)
	}
}

// TestPayoutHistory_RepoErrorReturns503 verifies transaction-repo failures surface as 503.
func TestPayoutHistory_RepoErrorReturns503(t *testing.T) {
	h, req := payoutHistoryFixture(t, nil, errors.New("db down"), "")
	rec := httptest.NewRecorder()
	h.GetPayoutHistory(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

// TestPayoutHistory_Unauthenticated verifies a missing user yields 401.
func TestPayoutHistory_Unauthenticated(t *testing.T) {
	h, req := payoutHistoryFixture(t, nil, nil, "")
	req = httptest.NewRequest(http.MethodGet, req.URL.String(), nil)
	req = withURLParam(req, "appID", uuid.New().String())
	rec := httptest.NewRecorder()
	h.GetPayoutHistory(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestPayoutHistory_CSVFormat verifies the CSV header, one row per period, and desc sort.
func TestPayoutHistory_CSVFormat(t *testing.T) {
	txs := []*entity.Transaction{
		phTx(2780, entity.EarningsStatusPaidOut, phMay1, phPaidJun5),
		phTx(3120, entity.EarningsStatusPaidOut, phJun1, phPaidJul5),
	}
	h, req := payoutHistoryFixture(t, txs, nil, "format=csv")
	rec := httptest.NewRecorder()
	h.GetPayoutHistory(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "payout-history.csv") {
		t.Errorf("Content-Disposition = %q, want payout-history.csv", cd)
	}
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected header + 2 rows, got %d: %v", len(records), records)
	}
	wantHeader := []string{"period", "amountCents", "chargeCount", "availableDate"}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], want)
		}
	}
	// Sorted period descending: 2026-06 before 2026-05.
	if records[1][0] != "2026-06" || records[2][0] != "2026-05" {
		t.Errorf("rows = %q, %q; want 2026-06 then 2026-05 (desc)", records[1][0], records[2][0])
	}
}

// TestPayoutHistory_Paging verifies the per-period rows are paged server-side: no paging
// returns every period with rowsTotal = full count; a limit/offset window returns just
// that slice (descending-period order preserved); KPIs ignore paging; and an offset past
// the end yields an empty window with total unchanged.
func TestPayoutHistory_Paging(t *testing.T) {
	// 6 distinct charge months → 6 distinct rows, sorted period descending.
	const n = 6
	avail := phPaidJul5
	txs := make([]*entity.Transaction, 0, n)
	for i := 0; i < n; i++ {
		charge := time.Date(2026, time.Month(1+i), 10, 0, 0, 0, 0, time.UTC)
		txs = append(txs, phTx(int64(1000+i), entity.EarningsStatusPaidOut, charge, avail))
	}

	// No paging → all rows, rowsTotal = full count.
	_, all := servePayoutHistory(t, txs, "")
	if len(all.Rows) != n || all.RowsTotal != n {
		t.Fatalf("no-paging: got %d rows / total %d, want %d / %d", len(all.Rows), all.RowsTotal, n, n)
	}
	wantTotalPaid := all.TotalPaidCents

	// limit=2&offset=1 → the [1,3) window (descending period), total still n.
	_, page := servePayoutHistory(t, txs, "limit=2&offset=1")
	if page.RowsTotal != n {
		t.Errorf("paged rowsTotal = %d, want %d (full count)", page.RowsTotal, n)
	}
	if len(page.Rows) != 2 {
		t.Fatalf("paged rows = %d, want 2", len(page.Rows))
	}
	// Descending: rows[0] is the latest month (2026-06); offset 1 → 2026-05.
	if page.Rows[0].Period != "2026-05" {
		t.Errorf("paged rows[0].period = %q, want 2026-05", page.Rows[0].Period)
	}
	// KPIs must reflect the FULL set regardless of paging.
	if page.TotalPaidCents != wantTotalPaid {
		t.Errorf("paged totalPaidCents = %d, want %d (KPIs must ignore paging)", page.TotalPaidCents, wantTotalPaid)
	}

	// offset past the end → empty (non-nil) window, total unchanged.
	_, beyond := servePayoutHistory(t, txs, "limit=2&offset=99")
	if len(beyond.Rows) != 0 || beyond.RowsTotal != n {
		t.Errorf("beyond-end: got %d rows / total %d, want 0 / %d", len(beyond.Rows), beyond.RowsTotal, n)
	}
}

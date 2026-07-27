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

// earningsTestFixture wires an EarningsReportHandler with the shared mock repos and
// returns a ready-to-serve GET request carrying an authenticated owner.
func earningsTestFixture(t *testing.T, txs []*entity.Transaction, findErr error, query string) (*EarningsReportHandler, *http.Request) {
	t.Helper()

	partnerAccount := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{
		ID:               appID,
		PartnerAccountID: partnerAccount.ID,
		Name:             "Test App",
		PartnerAppID:     "gid://partners/App/12345",
	}

	txRepo := &mockTxRepo{transactions: txs, findErr: findErr}
	partnerRepo := &mockPartnerRepoForSub{account: partnerAccount}
	appRepo := &mockAppRepoForSub{app: app}
	handler := NewEarningsReportHandler(txRepo, appRepo, partnerRepo)

	url := "/api/v1/apps/" + appID.String() + "/reports/earnings"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return handler, req
}

// earningsTx builds a transaction with the fields the earnings report reads.
func earningsTx(domain, shopName, currency string, gross, net int64, status entity.EarningsStatus, created, available time.Time) *entity.Transaction {
	return &entity.Transaction{
		ID:               uuid.New(),
		AppID:            uuid.New(),
		MyshopifyDomain:  domain,
		ShopName:         shopName,
		Currency:         currency,
		ChargeType:       valueobject.ChargeTypeRecurring,
		GrossAmountCents: gross,
		NetAmountCents:   net,
		CreatedAt:        created,
		AvailableDate:    available,
		EarningsStatus:   status,
	}
}

func serveEarnings(t *testing.T, txs []*entity.Transaction, query string) (*httptest.ResponseRecorder, earningsReport) {
	t.Helper()
	handler, req := earningsTestFixture(t, txs, nil, query)
	rec := httptest.NewRecorder()
	handler.GetEarningsReport(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var report earningsReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	return rec, report
}

var (
	dayC1 = time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	dayC2 = time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	dayC3 = time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	dayA  = time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)
)

func mixedEarningsTxs() []*entity.Transaction {
	return []*entity.Transaction{
		earningsTx("a.myshopify.com", "Acme", "USD", 4900, 3920, entity.EarningsStatusPending, dayC1, dayA),
		earningsTx("b.myshopify.com", "Beta", "USD", 2000, 1500, entity.EarningsStatusAvailable, dayC2, dayA),
		earningsTx("c.myshopify.com", "Gamma", "USD", 1000, 800, entity.EarningsStatusPaidOut, dayC3, dayA),
	}
}

func TestEarningsReport_KPISumsByStatus(t *testing.T) {
	_, report := serveEarnings(t, mixedEarningsTxs(), "")

	if report.PendingCents != 3920 {
		t.Errorf("pendingCents = %d, want 3920", report.PendingCents)
	}
	if report.AvailableCents != 1500 {
		t.Errorf("availableCents = %d, want 1500", report.AvailableCents)
	}
	if report.PaidOutCents != 800 {
		t.Errorf("paidOutCents = %d, want 800", report.PaidOutCents)
	}
	if report.NetEarningsCents != 3920+1500+800 {
		t.Errorf("netEarningsCents = %d, want 6220", report.NetEarningsCents)
	}
}

// TestEarningsReport_UnknownStatusExcludedAndReconciles verifies a tx with an
// unrecognized EarningsStatus is dropped from BOTH the KPIs and the charge table, so
// the visible rows' net sums to NetEarningsCents (no phantom row that isn't in a KPI).
func TestEarningsReport_UnknownStatusExcludedAndReconciles(t *testing.T) {
	txs := append(mixedEarningsTxs(),
		earningsTx("x.myshopify.com", "Xenon", "USD", 5000, 4000, entity.EarningsStatus("DISPUTED"), dayC1, dayA),
		earningsTx("y.myshopify.com", "Yttrium", "USD", 3000, 2400, entity.EarningsStatus(""), dayC2, dayA),
	)
	_, report := serveEarnings(t, txs, "")

	// Only the 3 known-status charges survive; the 2 unknown ones are dropped.
	if len(report.Charges) != 3 {
		t.Fatalf("charges: expected 3 (unknown-status excluded), got %d", len(report.Charges))
	}
	var chargeNetSum int64
	for _, c := range report.Charges {
		chargeNetSum += c.NetCents
	}
	if chargeNetSum != report.NetEarningsCents {
		t.Errorf("charge net sum %d != netEarningsCents %d (reconciliation broken)", chargeNetSum, report.NetEarningsCents)
	}
	// Net still equals only the known buckets (unknown 4000/2400 excluded).
	if report.NetEarningsCents != 3920+1500+800 {
		t.Errorf("netEarningsCents = %d, want 6220 (unknown-status net excluded)", report.NetEarningsCents)
	}
}

// TestEarningsReport_ZeroAvailableDateBlank verifies an unset AvailableDate serializes
// as "" (not "0001-01-01") so the UI can render its "—" fallback.
func TestEarningsReport_ZeroAvailableDateBlank(t *testing.T) {
	tx := earningsTx("a.myshopify.com", "Acme", "USD", 4900, 3920, entity.EarningsStatusPaidOut, dayC1, time.Time{})
	_, report := serveEarnings(t, []*entity.Transaction{tx}, "")
	if len(report.Charges) != 1 {
		t.Fatalf("expected 1 charge, got %d", len(report.Charges))
	}
	if report.Charges[0].AvailableDate != "" {
		t.Errorf("zero AvailableDate: expected \"\", got %q", report.Charges[0].AvailableDate)
	}
}

func TestEarningsReport_StatusLabelMapping(t *testing.T) {
	_, report := serveEarnings(t, mixedEarningsTxs(), "")

	labelByDomain := map[string]string{}
	for _, c := range report.Charges {
		labelByDomain[c.Domain] = c.Status
	}
	cases := map[string]string{
		"a.myshopify.com": "Pending",
		"b.myshopify.com": "Available",
		"c.myshopify.com": "Paid",
	}
	for domain, want := range cases {
		if got := labelByDomain[domain]; got != want {
			t.Errorf("status for %s = %q, want %q", domain, got, want)
		}
	}
}

func TestEarningsReport_ChargesSortedDateDesc(t *testing.T) {
	_, report := serveEarnings(t, mixedEarningsTxs(), "")

	wantOrder := []string{"2026-07-24", "2026-07-22", "2026-07-20"}
	if len(report.Charges) != len(wantOrder) {
		t.Fatalf("expected %d charges, got %d", len(wantOrder), len(report.Charges))
	}
	for i, want := range wantOrder {
		if report.Charges[i].Date != want {
			t.Errorf("charge[%d].date = %s, want %s", i, report.Charges[i].Date, want)
		}
	}
}

func TestEarningsReport_ChargeFields(t *testing.T) {
	_, report := serveEarnings(t, mixedEarningsTxs(), "")

	var acme *earningsCharge
	for i := range report.Charges {
		if report.Charges[i].Domain == "a.myshopify.com" {
			acme = &report.Charges[i]
		}
	}
	if acme == nil {
		t.Fatal("acme charge not found")
	}
	if acme.ShopName != "Acme" || acme.GrossCents != 4900 || acme.NetCents != 3920 {
		t.Errorf("unexpected acme charge: %+v", *acme)
	}
	if acme.AvailableDate != "2026-08-06" {
		t.Errorf("availableDate = %s, want 2026-08-06", acme.AvailableDate)
	}
}

func TestEarningsReport_CurrencyDefaultsUSD(t *testing.T) {
	txs := []*entity.Transaction{
		earningsTx("a.myshopify.com", "Acme", "", 100, 90, entity.EarningsStatusPending, dayC1, dayA),
	}
	_, report := serveEarnings(t, txs, "")
	if report.Currency != "USD" {
		t.Errorf("currency = %s, want USD (default)", report.Currency)
	}
}

func TestEarningsReport_CurrencyPicksFirstNonEmpty(t *testing.T) {
	txs := []*entity.Transaction{
		earningsTx("a.myshopify.com", "Acme", "", 100, 90, entity.EarningsStatusPending, dayC1, dayA),
		earningsTx("b.myshopify.com", "Beta", "EUR", 200, 180, entity.EarningsStatusAvailable, dayC2, dayA),
	}
	_, report := serveEarnings(t, txs, "")
	if report.Currency != "EUR" {
		t.Errorf("currency = %s, want EUR (first non-empty)", report.Currency)
	}
}

func TestEarningsReport_EmptyCase(t *testing.T) {
	rec, report := serveEarnings(t, []*entity.Transaction{}, "")

	if report.NetEarningsCents != 0 || report.PendingCents != 0 || report.AvailableCents != 0 || report.PaidOutCents != 0 {
		t.Errorf("expected all-zero KPIs, got %+v", report)
	}
	if report.Currency != "USD" {
		t.Errorf("currency = %s, want USD", report.Currency)
	}
	// charges must serialize as [] not null.
	if !strings.Contains(rec.Body.String(), `"charges":[]`) {
		t.Errorf("expected charges to serialize as [], body: %s", rec.Body.String())
	}
}

func TestEarningsReport_RepoErrorReturns503(t *testing.T) {
	handler, req := earningsTestFixture(t, nil, errors.New("db down"), "")
	rec := httptest.NewRecorder()
	handler.GetEarningsReport(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestEarningsReport_Unauthenticated(t *testing.T) {
	handler, req := earningsTestFixture(t, mixedEarningsTxs(), nil, "")
	// Strip the user from context.
	req = httptest.NewRequest(http.MethodGet, req.URL.String(), nil)
	req = withURLParam(req, "appID", uuid.New().String())
	rec := httptest.NewRecorder()
	handler.GetEarningsReport(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestEarningsReport_CSVFormatAndEscaping(t *testing.T) {
	txs := []*entity.Transaction{
		earningsTx(`weird"quote.myshopify.com`, "Acme, Inc.", "USD", 4900, 3920, entity.EarningsStatusPending, dayC1, dayA),
	}
	handler, req := earningsTestFixture(t, txs, nil, "format=csv")
	rec := httptest.NewRecorder()
	handler.GetEarningsReport(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "earnings.csv") {
		t.Errorf("Content-Disposition = %q, want earnings.csv", cd)
	}

	// Re-parse the CSV: a shopName with a comma must remain a single field.
	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d records", len(records))
	}
	header := records[0]
	wantHeader := []string{"date", "store", "domain", "gross", "net", "status", "availableDate"}
	if len(header) != len(wantHeader) {
		t.Fatalf("header = %v, want %v", header, wantHeader)
	}
	for i, h := range wantHeader {
		if header[i] != h {
			t.Errorf("header[%d] = %q, want %q", i, header[i], h)
		}
	}
	row := records[1]
	if len(row) != 7 {
		t.Fatalf("data row has %d fields, want 7: %v", len(row), row)
	}
	if row[1] != "Acme, Inc." {
		t.Errorf("store field = %q, want %q (comma must stay in one field)", row[1], "Acme, Inc.")
	}
	if row[2] != `weird"quote.myshopify.com` {
		t.Errorf("domain field = %q, want the embedded-quote domain in one field", row[2])
	}
	if row[3] != "4900" || row[4] != "3920" {
		t.Errorf("gross/net = %q/%q, want 4900/3920", row[3], row[4])
	}
	if row[5] != "Pending" {
		t.Errorf("status = %q, want Pending", row[5])
	}
}

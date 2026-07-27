package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

// revenueMixFixture wires a RevenueMixReportHandler with the shared mock repos and
// returns a ready-to-serve GET request carrying an authenticated owner.
func revenueMixFixture(t *testing.T, txs []*entity.Transaction, findErr error, query string) (*RevenueMixReportHandler, *http.Request) {
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
	handler := NewRevenueMixReportHandler(txRepo, appRepo, partnerRepo)

	url := "/api/v1/apps/" + appID.String() + "/reports/revenue-mix"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	user := &entity.User{ID: partnerAccount.UserID, Role: valueobject.RoleOwner}
	req = req.WithContext(contextWithUser(req.Context(), user))
	return handler, req
}

// mixTx builds a transaction with the fields the revenue-mix report reads: ChargeType,
// NetAmountCents (via AmountCents) and Currency.
func mixTx(ct valueobject.ChargeType, netCents int64, currency string) *entity.Transaction {
	return &entity.Transaction{
		ID:             uuid.New(),
		AppID:          uuid.New(),
		ChargeType:     ct,
		NetAmountCents: netCents,
		// Deliberately different from net so a regression to summing gross instead of
		// net (AmountCents) would change every total and fail the tests.
		GrossAmountCents: netCents + 9999,
		Currency:         currency,
	}
}

func serveRevenueMix(t *testing.T, txs []*entity.Transaction, query string) (*httptest.ResponseRecorder, revenueMixReport) {
	t.Helper()
	handler, req := revenueMixFixture(t, txs, nil, query)
	rec := httptest.NewRecorder()
	handler.GetRevenueMix(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var report revenueMixReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	return rec, report
}

// mixedRevenueTxs: recurring 5000+1000=6000, usage 3000, one-time 1000, refund 500.
// gross = 10000, net = 9500. Usage != One-time so a Usage↔One-time case-label swap
// would be caught.
func mixedRevenueTxs() []*entity.Transaction {
	return []*entity.Transaction{
		mixTx(valueobject.ChargeTypeRecurring, 5000, "USD"),
		mixTx(valueobject.ChargeTypeRecurring, 1000, "USD"),
		mixTx(valueobject.ChargeTypeUsage, 3000, "USD"),
		mixTx(valueobject.ChargeTypeOneTime, 1000, "USD"),
		mixTx(valueobject.ChargeTypeRefund, 500, "USD"),
	}
}

func TestRevenueMix_PerTypeSums(t *testing.T) {
	_, report := serveRevenueMix(t, mixedRevenueTxs(), "")

	if report.RecurringCents != 6000 {
		t.Errorf("recurringCents = %d, want 6000", report.RecurringCents)
	}
	if report.UsageCents != 3000 {
		t.Errorf("usageCents = %d, want 3000", report.UsageCents)
	}
	if report.OneTimeCents != 1000 {
		t.Errorf("oneTimeCents = %d, want 1000", report.OneTimeCents)
	}
	if report.RefundCents != 500 {
		t.Errorf("refundCents = %d, want 500", report.RefundCents)
	}
}

func TestRevenueMix_GrossExcludesRefund(t *testing.T) {
	_, report := serveRevenueMix(t, mixedRevenueTxs(), "")

	// gross = recurring + usage + oneTime, refund excluded.
	if report.GrossCents != 6000+3000+1000 {
		t.Errorf("grossCents = %d, want 10000 (refund excluded)", report.GrossCents)
	}
}

func TestRevenueMix_NetIsGrossMinusRefund(t *testing.T) {
	_, report := serveRevenueMix(t, mixedRevenueTxs(), "")

	if report.NetCents != report.GrossCents-report.RefundCents {
		t.Errorf("netCents = %d, want gross-refund = %d", report.NetCents, report.GrossCents-report.RefundCents)
	}
	if report.NetCents != 9500 {
		t.Errorf("netCents = %d, want 9500", report.NetCents)
	}
}

func TestRevenueMix_SegmentPctsSumToOneAndKnownSplit(t *testing.T) {
	_, report := serveRevenueMix(t, mixedRevenueTxs(), "")

	if len(report.Segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(report.Segments))
	}

	pctByType := map[string]float64{}
	amtByType := map[string]int64{}
	var sum float64
	for _, s := range report.Segments {
		pctByType[s.Type] = s.Pct
		amtByType[s.Type] = s.AmountCents
		sum += s.Pct
	}

	// Recurring 6000/10000 = 0.6.
	if got := pctByType["Recurring"]; math.Abs(got-0.6) > 1e-9 {
		t.Errorf("Recurring pct = %v, want 0.6", got)
	}
	if got := pctByType["Usage"]; math.Abs(got-0.3) > 1e-9 {
		t.Errorf("Usage pct = %v, want 0.3", got)
	}
	if got := pctByType["One-time"]; math.Abs(got-0.1) > 1e-9 {
		t.Errorf("One-time pct = %v, want 0.1", got)
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Errorf("segment pcts sum = %v, want ~1.0", sum)
	}

	// Amounts are of the positive streams.
	if amtByType["Recurring"] != 6000 || amtByType["Usage"] != 3000 || amtByType["One-time"] != 1000 {
		t.Errorf("segment amounts = %+v, want 6000/3000/1000", amtByType)
	}
	// Refund is NOT a composition segment — the 3 segment amounts sum to gross.
	if amtByType["Recurring"]+amtByType["Usage"]+amtByType["One-time"] != report.GrossCents {
		t.Errorf("segment amounts should sum to gross %d; refund must not be a segment", report.GrossCents)
	}
}

// TestRevenueMix_UnknownChargeTypeExcludedFromGross verifies a tx with an unrecognized
// ChargeType is excluded from gross/segments (and logged) rather than silently counted.
func TestRevenueMix_UnknownChargeTypeExcludedFromGross(t *testing.T) {
	txs := []*entity.Transaction{
		mixTx(valueobject.ChargeTypeRecurring, 5000, "USD"),
		mixTx(valueobject.ChargeType("DISPUTED"), 9000, "USD"),
		mixTx(valueobject.ChargeType(""), 3000, "USD"),
	}
	_, report := serveRevenueMix(t, txs, "")
	// Only the recognized RECURRING tx counts; the unknown ones are excluded from gross.
	if report.RecurringCents != 5000 || report.GrossCents != 5000 {
		t.Errorf("recurring=%d gross=%d, want 5000/5000 (unknown types excluded)", report.RecurringCents, report.GrossCents)
	}
	if report.UsageCents != 0 || report.OneTimeCents != 0 || report.RefundCents != 0 {
		t.Errorf("unknown types leaked into a bucket: %+v", report)
	}
}

// TestRevenueMix_RecurringUsageSeparation verifies a USAGE tx never leaks into the
// recurring bucket (Revenue Classification: RECURRING and USAGE strictly separated).
func TestRevenueMix_RecurringUsageSeparation(t *testing.T) {
	txs := []*entity.Transaction{
		mixTx(valueobject.ChargeTypeUsage, 7777, "USD"),
	}
	_, report := serveRevenueMix(t, txs, "")

	if report.RecurringCents != 0 {
		t.Errorf("recurringCents = %d, want 0 (USAGE must not land in recurring)", report.RecurringCents)
	}
	if report.UsageCents != 7777 {
		t.Errorf("usageCents = %d, want 7777", report.UsageCents)
	}
}

// TestRevenueMix_EmptyPctGuard verifies gross=0 yields pct 0 for all segments with no
// divide-by-zero panic, and segments serialize as a present array with zero values.
func TestRevenueMix_EmptyPctGuard(t *testing.T) {
	rec, report := serveRevenueMix(t, []*entity.Transaction{}, "")

	if report.GrossCents != 0 || report.NetCents != 0 {
		t.Errorf("empty: gross=%d net=%d, want 0/0", report.GrossCents, report.NetCents)
	}
	if len(report.Segments) != 3 {
		t.Fatalf("expected 3 segments even when empty, got %d", len(report.Segments))
	}
	for _, s := range report.Segments {
		if s.AmountCents != 0 || s.Pct != 0 {
			t.Errorf("empty segment %s: amount=%d pct=%v, want 0/0", s.Type, s.AmountCents, s.Pct)
		}
	}
	if report.Currency != "USD" {
		t.Errorf("currency = %s, want USD", report.Currency)
	}
	// segments must serialize as a JSON array, not null.
	if !strings.Contains(rec.Body.String(), `"segments":[`) {
		t.Errorf("expected segments to serialize as array, body: %s", rec.Body.String())
	}
}

func TestRevenueMix_CurrencyDefaultsUSD(t *testing.T) {
	txs := []*entity.Transaction{
		mixTx(valueobject.ChargeTypeRecurring, 100, ""),
	}
	_, report := serveRevenueMix(t, txs, "")
	if report.Currency != "USD" {
		t.Errorf("currency = %s, want USD (default)", report.Currency)
	}
}

func TestRevenueMix_CurrencyPicksFirstNonEmpty(t *testing.T) {
	txs := []*entity.Transaction{
		mixTx(valueobject.ChargeTypeRecurring, 100, ""),
		mixTx(valueobject.ChargeTypeUsage, 200, "EUR"),
	}
	_, report := serveRevenueMix(t, txs, "")
	if report.Currency != "EUR" {
		t.Errorf("currency = %s, want EUR (first non-empty)", report.Currency)
	}
}

func TestRevenueMix_RepoErrorReturns503(t *testing.T) {
	handler, req := revenueMixFixture(t, nil, errors.New("db down"), "")
	rec := httptest.NewRecorder()
	handler.GetRevenueMix(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestRevenueMix_Unauthenticated(t *testing.T) {
	handler, req := revenueMixFixture(t, mixedRevenueTxs(), nil, "")
	// Strip the user from context.
	req = httptest.NewRequest(http.MethodGet, req.URL.String(), nil)
	req = withURLParam(req, "appID", uuid.New().String())
	rec := httptest.NewRecorder()
	handler.GetRevenueMix(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRevenueMix_CSVFormat(t *testing.T) {
	handler, req := revenueMixFixture(t, mixedRevenueTxs(), nil, "format=csv")
	rec := httptest.NewRecorder()
	handler.GetRevenueMix(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "revenue-mix.csv") {
		t.Errorf("Content-Disposition = %q, want revenue-mix.csv", cd)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// header + 3 segments + refund + net = 6 rows.
	if len(records) != 6 {
		t.Fatalf("expected 6 records (header+3 segments+refund+net), got %d: %v", len(records), records)
	}

	wantHeader := []string{"type", "amountCents", "pct"}
	for i, h := range wantHeader {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}

	// Recurring segment row: 6000 / 0.6000.
	if records[1][0] != "Recurring" || records[1][1] != "6000" || records[1][2] != "0.6000" {
		t.Errorf("recurring row = %v, want [Recurring 6000 0.6000]", records[1])
	}
	// Refund row: negative amount, blank pct.
	if records[4][0] != "Refund" || records[4][1] != "-500" || records[4][2] != "" {
		t.Errorf("refund row = %v, want [Refund -500 <blank>]", records[4])
	}
	// Net row.
	if records[5][0] != "Net" || records[5][1] != "9500" || records[5][2] != "" {
		t.Errorf("net row = %v, want [Net 9500 <blank>]", records[5])
	}
}

// TestRevenueMix_CSVNoRefundRowWhenZero verifies the Refund row is omitted when there
// are no refunds (header + 3 segments + net = 5 rows).
func TestRevenueMix_CSVNoRefundRowWhenZero(t *testing.T) {
	txs := []*entity.Transaction{
		mixTx(valueobject.ChargeTypeRecurring, 5000, "USD"),
		mixTx(valueobject.ChargeTypeUsage, 5000, "USD"),
	}
	handler, req := revenueMixFixture(t, txs, nil, "format=csv")
	rec := httptest.NewRecorder()
	handler.GetRevenueMix(rec, req)

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) != 5 {
		t.Fatalf("expected 5 records (no refund row), got %d: %v", len(records), records)
	}
	if records[4][0] != "Net" {
		t.Errorf("last row = %v, want Net (no Refund row)", records[4])
	}
}

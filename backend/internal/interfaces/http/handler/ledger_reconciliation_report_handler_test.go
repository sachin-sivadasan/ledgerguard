package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func doRecon(t *testing.T, txs []*entity.Transaction) map[string]any {
	t.Helper()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App"}
	h := NewLedgerReconciliationReportHandler(
		&mockTxRepo{transactions: txs},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		domainservice.NewFeeVerificationService(),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/reports/ledger-reconciliation?months=1", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetLedgerReconciliation(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// TestRecon_ItemizedBalanced: a fully-itemized sale (the shape sync now produces) — gross
// = net + revenue_share + processing — reconciles to a zero residual.
func TestRecon_ItemizedBalanced(t *testing.T) {
	appID := uuid.New()
	tx := &entity.Transaction{
		ID: uuid.New(), AppID: appID,
		GrossAmountCents:   10000,
		NetAmountCents:     8200, // 10000 − 1500 revenue share − 300 processing
		ShopifyFeeCents:    1500,
		ProcessingFeeCents: 300,
		TransactionDate:    time.Now().UTC(),
		Currency:           "USD",
	}
	resp := doRecon(t, []*entity.Transaction{tx})

	if resp["reconciled"] != true {
		t.Errorf("reconciled = %v, want true (buckets close: 8200 + 1500 + 300 = 10000)", resp["reconciled"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["revenue_share_cents"].(float64) != 1500 {
		t.Errorf("revenue_share_cents = %v, want 1500", m["revenue_share_cents"])
	}
	if m["processing_cents"].(float64) != 300 {
		t.Errorf("processing_cents = %v, want 300", m["processing_cents"])
	}
	if m["accounted_cents"].(float64) != 10000 {
		t.Errorf("accounted_cents = %v, want 10000 (net + share + processing)", m["accounted_cents"])
	}
	if m["residual_cents"].(float64) != 0 {
		t.Errorf("residual_cents = %v, want 0", m["residual_cents"])
	}
}

// TestRecon_UnaccountedResidualFlagged: money the three buckets don't explain (here a row
// whose processing/fee never synced, so net + share + processing < gross) is flagged — the
// honest residual signal, e.g. a refund/credit whose fee reversal is missing.
func TestRecon_UnaccountedResidualFlagged(t *testing.T) {
	appID := uuid.New()
	tx := &entity.Transaction{
		ID: uuid.New(), AppID: appID,
		GrossAmountCents:   10000,
		NetAmountCents:     8200, // Shopify kept 1800, but only 1500 is itemized...
		ShopifyFeeCents:    1500,
		ProcessingFeeCents: 0, // ...and processing never synced → 300 unaccounted
		TransactionDate:    time.Now().UTC(),
		Currency:           "USD",
	}
	resp := doRecon(t, []*entity.Transaction{tx})

	if resp["reconciled"] != false {
		t.Errorf("reconciled = %v, want false (8200 + 1500 + 0 = 9700 ≠ gross 10000)", resp["reconciled"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["residual_cents"].(float64) != 300 {
		t.Errorf("residual_cents = %v, want 300 (gross − accounted)", m["residual_cents"])
	}
	if resp["months_flagged"].(float64) != 1 {
		t.Errorf("months_flagged = %v, want 1", resp["months_flagged"])
	}
}

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

// TestRecon_Balanced: net == gross − fee (the Shopify identity holds) → reconciled.
func TestRecon_Balanced(t *testing.T) {
	appID := uuid.New()
	resp := doRecon(t, []*entity.Transaction{feeTx(appID, 10000, 1500)}) // net 8500 = 10000 − 1500

	if resp["reconciled"] != true {
		t.Errorf("reconciled = %v, want true (identity holds)", resp["reconciled"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["expected_net_cents"].(float64) != 8500 {
		t.Errorf("expected_net_cents = %v, want 8500 (gross − fee)", m["expected_net_cents"])
	}
	if m["residual_cents"].(float64) != 0 {
		t.Errorf("residual_cents = %v, want 0", m["residual_cents"])
	}
}

// TestRecon_MissingFeeFlagged: a month with no fee recorded but net < gross doesn't
// satisfy gross = net + fee → flagged (the ledger and Shopify disagree).
func TestRecon_MissingFeeFlagged(t *testing.T) {
	appID := uuid.New()
	tx := &entity.Transaction{
		ID: uuid.New(), AppID: appID,
		GrossAmountCents: 10000,
		NetAmountCents:   8500, // Shopify kept 1500, but ShopifyFeeCents is 0 (unsynced)
		ShopifyFeeCents:  0,
		TransactionDate:  time.Now().UTC(),
		Currency:         "USD",
	}
	resp := doRecon(t, []*entity.Transaction{tx})

	if resp["reconciled"] != false {
		t.Errorf("reconciled = %v, want false (net 8500 ≠ gross 10000 − fee 0)", resp["reconciled"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["residual_cents"].(float64) != -1500 {
		t.Errorf("residual_cents = %v, want -1500", m["residual_cents"])
	}
	if resp["months_flagged"].(float64) != 1 {
		t.Errorf("months_flagged = %v, want 1", resp["months_flagged"])
	}
}

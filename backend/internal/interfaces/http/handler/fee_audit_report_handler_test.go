package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
)

func doFeeAudit(t *testing.T, tier valueobject.RevenueShareTier, txs []*entity.Transaction, query string) map[string]any {
	t.Helper()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App", RevenueShareTier: tier}
	h := NewFeeAuditReportHandler(
		&mockTxRepo{transactions: txs},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		domainservice.NewFeeVerificationService(),
	)
	url := "/api/v1/apps/" + appID.String() + "/reports/fee-audit"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetFeeAudit(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// TestFeeAudit_CleanMonth: consistent 15% cut → detected 15, tier_matches true, no
// flagged months, and savings vs the 20% default surfaced.
func TestFeeAudit_CleanMonth(t *testing.T) {
	appID := uuid.New()
	resp := doFeeAudit(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 1500)}, "months=1")

	if resp["detected_fee_pct"].(float64) != 15 {
		t.Errorf("detected_fee_pct = %v, want 15", resp["detected_fee_pct"])
	}
	if resp["tier_matches"] != true {
		t.Errorf("tier_matches = %v, want true", resp["tier_matches"])
	}
	if resp["flagged_months"].(float64) != 0 {
		t.Errorf("flagged_months = %v, want 0", resp["flagged_months"])
	}
	if resp["total_cut_cents"].(float64) != 1500 {
		t.Errorf("total_cut_cents = %v, want 1500", resp["total_cut_cents"])
	}
	// savings vs 20% default = 10000*20% − 1500 = 500.
	if resp["savings_vs_default_cents"].(float64) != 500 {
		t.Errorf("savings_vs_default_cents = %v, want 500", resp["savings_vs_default_cents"])
	}
	if len(resp["months"].([]any)) != 1 {
		t.Errorf("expected 1 month, got %d", len(resp["months"].([]any)))
	}
}

// TestFeeAudit_TierMismatch: configured 0% but Shopify retains 15% → detected 15,
// tier_matches false (the RPT-FEES-2 signal surfaced in the audit).
func TestFeeAudit_TierMismatch(t *testing.T) {
	appID := uuid.New()
	resp := doFeeAudit(t, valueobject.RevenueShareTierSmallDev0,
		[]*entity.Transaction{feeTx(appID, 10000, 1500)}, "months=1")

	if resp["configured_fee_pct"].(float64) != 0 {
		t.Errorf("configured_fee_pct = %v, want 0", resp["configured_fee_pct"])
	}
	if resp["detected_fee_pct"].(float64) != 15 {
		t.Errorf("detected_fee_pct = %v, want 15", resp["detected_fee_pct"])
	}
	if resp["tier_matches"] != false {
		t.Errorf("tier_matches = %v, want false (configured 0%% ≠ actual 15%%)", resp["tier_matches"])
	}
}

func TestFeeAudit_CSV(t *testing.T) {
	appID := uuid.New()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID2 := uuid.New()
	app := &entity.App{ID: appID2, PartnerAccountID: pa.ID, Name: "Test App", RevenueShareTier: valueobject.RevenueShareTierSmallDev15}
	h := NewFeeAuditReportHandler(
		&mockTxRepo{transactions: []*entity.Transaction{feeTx(appID, 10000, 1500)}},
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		domainservice.NewFeeVerificationService(),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID2.String()+"/reports/fee-audit?months=1&format=csv", nil)
	req = withURLParam(req, "appID", appID2.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetFeeAudit(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("content-type = %q, want text/csv", ct)
	}
	if !strings.Contains(rec.Body.String(), "month,gross_cents,shopify_cut_cents") {
		t.Errorf("CSV header missing: %s", rec.Body.String())
	}
}

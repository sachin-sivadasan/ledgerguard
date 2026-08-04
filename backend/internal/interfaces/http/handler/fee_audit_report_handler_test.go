package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// TestFeeAudit_OffTierMonthFlagged: a month whose rate sits BETWEEN tiers (12% ≈ a
// transition/mischarge) doesn't match its nearest tier (15%) → flagged. Per-month
// snapping is what makes this robust to the $1M-crossing window (RPT-FEES-2 nuance).
func TestFeeAudit_OffTierMonthFlagged(t *testing.T) {
	appID := uuid.New()
	resp := doFeeAudit(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 1200)}, "months=1") // 12% actual

	m := resp["months"].([]any)[0].(map[string]any)
	if m["expected_cut_cents"].(float64) != 1500 {
		t.Errorf("expected_cut_cents = %v, want 1500 (12%% snaps to the 15%% tier)", m["expected_cut_cents"])
	}
	if m["fee_guard_ok"] != false {
		t.Errorf("fee_guard_ok = %v, want false (12%% is off-tier)", m["fee_guard_ok"])
	}
}

// TestFeeAudit_ZeroTierMonthClean: a legitimately 0% month (pre-$1M) snaps to the 0%
// tier and passes — it must NOT be flagged just because later months are 15%.
func TestFeeAudit_ZeroTierMonthClean(t *testing.T) {
	appID := uuid.New()
	resp := doFeeAudit(t, valueobject.RevenueShareTierSmallDev0,
		[]*entity.Transaction{feeTx(appID, 10000, 0)}, "months=1")

	m := resp["months"].([]any)[0].(map[string]any)
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true (a clean 0%% month is on-tier)", m["fee_guard_ok"])
	}
	if resp["flagged_months"].(float64) != 0 {
		t.Errorf("flagged_months = %v, want 0", resp["flagged_months"])
	}
}

// TestFeeAudit_ProcessingFeeDoesNotShiftTier is the guard for the derived-processing-fee
// change: tier detection must key off the REVENUE-SHARE rate (shopifyFee ÷ gross), never
// the total effective fee. A real month is 15% revenue share + ~3% processing = 18%
// effective; if detection ever snapped on the effective rate it would round 18% → the 20%
// tier and wrongly report a mismatch. With processing populated, detected must stay 15.
func TestFeeAudit_ProcessingFeeDoesNotShiftTier(t *testing.T) {
	appID := uuid.New()
	tx := &entity.Transaction{
		ID:                 uuid.New(),
		AppID:              appID,
		GrossAmountCents:   10000,
		ShopifyFeeCents:    1500, // 15% revenue share
		ProcessingFeeCents: 300,  // 3% processing baked into net (derived at sync)
		NetAmountCents:     8200, // 10000 − 1500 − 300
		TransactionDate:    time.Now().UTC(),
		Currency:           "USD",
	}
	resp := doFeeAudit(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{tx}, "months=1")

	if resp["detected_fee_pct"].(float64) != 15 {
		t.Errorf("detected_fee_pct = %v, want 15 (must ignore the 3%% processing fee)", resp["detected_fee_pct"])
	}
	if resp["tier_matches"] != true {
		t.Errorf("tier_matches = %v, want true (15%% configured == 15%% detected)", resp["tier_matches"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true (15%% cut is on-tier; processing is not a mischarge)", m["fee_guard_ok"])
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

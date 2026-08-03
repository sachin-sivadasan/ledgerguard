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

func feeTx(appID uuid.UUID, grossCents, shopifyFeeCents int64) *entity.Transaction {
	return &entity.Transaction{
		ID:               uuid.New(),
		AppID:            appID,
		GrossAmountCents: grossCents,
		NetAmountCents:   grossCents - shopifyFeeCents,
		ShopifyFeeCents:  shopifyFeeCents,
		TransactionDate:  time.Now().UTC(),
		Currency:         "USD",
	}
}

func doMonthlyFees(t *testing.T, tier valueobject.RevenueShareTier, txs []*entity.Transaction) map[string]any {
	t.Helper()
	pa := &entity.PartnerAccount{ID: uuid.New(), UserID: uuid.New()}
	appID := uuid.New()
	app := &entity.App{ID: appID, PartnerAccountID: pa.ID, Name: "Test App", RevenueShareTier: tier}
	h := NewFeeHandler(
		&mockAppRepoForSub{app: app},
		&mockPartnerRepoForSub{account: pa},
		&mockTxRepo{transactions: txs},
		domainservice.NewFeeVerificationService(),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/apps/"+appID.String()+"/fees/monthly?months=1", nil)
	req = withURLParam(req, "appID", appID.String())
	req = req.WithContext(contextWithUser(req.Context(), &entity.User{ID: pa.UserID, Role: valueobject.RoleOwner}))
	rec := httptest.NewRecorder()
	h.GetMonthlyProfitBreakdown(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

// TestMonthlyFees_GuardMatches: actual shopifyFee equals the tier's revenue share
// (15% of gross) → guard consistent, real shopify_cut populated (RPT-FEES-1).
func TestMonthlyFees_GuardMatches(t *testing.T) {
	appID := uuid.New()
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 1500)})

	if resp["detected_fee_pct"].(float64) != 15 {
		t.Errorf("detected_fee_pct = %v, want 15", resp["detected_fee_pct"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["shopify_cut_cents"].(float64) != 1500 {
		t.Errorf("shopify_cut_cents = %v, want 1500 (real, from shopifyFee)", m["shopify_cut_cents"])
	}
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true (consistent 15%%)", m["fee_guard_ok"])
	}
}

// TestMonthlyFees_DetectsActualTierAndFlagsMismatch is the RPT-FEES-2 guard: when the
// CONFIGURED tier (0%) doesn't match what Shopify actually retains (~15%), the report
// detects the real rate from the data, reports the configured-vs-actual mismatch
// (tier_matches=false), and the per-month guard uses the DETECTED rate (so consistent
// months don't false-alarm).
func TestMonthlyFees_DetectsActualTierAndFlagsMismatch(t *testing.T) {
	appID := uuid.New()
	// App configured 0%, but Shopify actually retained 15% ($15 of $100).
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev0,
		[]*entity.Transaction{feeTx(appID, 10000, 1500)})

	if resp["configured_fee_pct"].(float64) != 0 {
		t.Errorf("configured_fee_pct = %v, want 0", resp["configured_fee_pct"])
	}
	if resp["detected_fee_pct"].(float64) != 15 {
		t.Errorf("detected_fee_pct = %v, want 15 (snapped from the 15%% actual)", resp["detected_fee_pct"])
	}
	if resp["tier_matches"] != false {
		t.Errorf("tier_matches = %v, want false (configured 0%% ≠ actual 15%%)", resp["tier_matches"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["expected_cut_cents"].(float64) != 1500 {
		t.Errorf("expected_cut_cents = %v, want 1500 (gross × DETECTED 15%%, not configured 0%%)", m["expected_cut_cents"])
	}
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true (data is self-consistent at 15%%)", m["fee_guard_ok"])
	}
}

// TestMonthlyFees_ConfiguredMatchesActual: a correctly-configured 15% app whose data is
// 15% → tier_matches true.
func TestMonthlyFees_ConfiguredMatchesActual(t *testing.T) {
	appID := uuid.New()
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 1500)})
	if resp["tier_matches"] != true {
		t.Errorf("tier_matches = %v, want true (configured 15%% == actual 15%%)", resp["tier_matches"])
	}
}

// TestSnapToTierRate pins the rate→tier snapping to the nearest real Shopify rate.
func TestSnapToTierRate(t *testing.T) {
	cases := map[float64]float64{
		0: 0, 3.1: 0, 9.7: 15, 14: 15, 15.02: 15, 17: 15, 18.5: 20, 20: 20, 22: 20,
	}
	for in, want := range cases {
		if got := snapToTierRate(in); got != want {
			t.Errorf("snapToTierRate(%v) = %v, want %v", in, got, want)
		}
	}
}

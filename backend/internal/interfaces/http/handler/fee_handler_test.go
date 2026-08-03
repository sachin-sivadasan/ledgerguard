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
// (15% of gross) → fee_guard_ok true, variance 0, real shopify_cut populated (RPT-FEES-1).
func TestMonthlyFees_GuardMatches(t *testing.T) {
	appID := uuid.New()
	// tier 15% of $100 gross = $15.00 retained (pure revenue share, no processing).
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 1500)})

	if resp["expected_fee_pct"].(float64) != 15 {
		t.Errorf("expected_fee_pct = %v, want 15", resp["expected_fee_pct"])
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["shopify_cut_cents"].(float64) != 1500 {
		t.Errorf("shopify_cut_cents = %v, want 1500 (real, from shopifyFee)", m["shopify_cut_cents"])
	}
	if m["expected_cut_cents"].(float64) != 1500 {
		t.Errorf("expected_cut_cents = %v, want 1500 (revenue share)", m["expected_cut_cents"])
	}
	if m["fee_variance_cents"].(float64) != 0 {
		t.Errorf("fee_variance_cents = %v, want 0", m["fee_variance_cents"])
	}
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true", m["fee_guard_ok"])
	}
}

// TestMonthlyFees_GuardFlagsOvercharge: Shopify retained 20% but the app's tier is 15%
// → variance beyond 1% of gross trips the Fee Guard (catches a wrong rate / misconfigured
// tier, exactly the case seen live where a 0%-configured app was charged ~15%).
func TestMonthlyFees_GuardFlagsOvercharge(t *testing.T) {
	appID := uuid.New()
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 2000)}) // 20% actual vs 15% expected

	m := resp["months"].([]any)[0].(map[string]any)
	if m["fee_variance_cents"].(float64) != 500 {
		t.Errorf("fee_variance_cents = %v, want 500 (2000 actual − 1500 expected)", m["fee_variance_cents"])
	}
	if m["fee_guard_ok"] != false {
		t.Errorf("fee_guard_ok = %v, want false (overcharge exceeds 1%% of gross)", m["fee_guard_ok"])
	}
}

// TestMonthlyFees_ZeroTierNoFee: a correctly-configured 0% app that Shopify retains ~0
// from → guard passes.
func TestMonthlyFees_ZeroTierNoFee(t *testing.T) {
	appID := uuid.New()
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev0,
		[]*entity.Transaction{feeTx(appID, 10000, 0)})

	m := resp["months"].([]any)[0].(map[string]any)
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true (0%% tier, 0 fee)", m["fee_guard_ok"])
	}
}

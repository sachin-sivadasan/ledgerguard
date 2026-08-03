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

// TestMonthlyFees_GuardMatches: actual shopifyFee equals the expected TOTAL tier fee
// (revenue share 15% + processing 2.9% = 17.9%) → fee_guard_ok true, variance 0, real
// shopify_cut populated (RPT-FEES-1).
func TestMonthlyFees_GuardMatches(t *testing.T) {
	appID := uuid.New()
	// tier 15% + 2.9% processing = 17.9% of $100 gross = $17.90 retained.
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 1790)})

	if p := resp["expected_fee_pct"].(float64); p < 17.8 || p > 18.0 {
		t.Errorf("expected_fee_pct = %v, want ~17.9 (15%% share + 2.9%% processing)", p)
	}
	m := resp["months"].([]any)[0].(map[string]any)
	if m["shopify_cut_cents"].(float64) != 1790 {
		t.Errorf("shopify_cut_cents = %v, want 1790 (real, from shopifyFee)", m["shopify_cut_cents"])
	}
	if m["expected_cut_cents"].(float64) != 1790 {
		t.Errorf("expected_cut_cents = %v, want 1790 (revenue share + processing)", m["expected_cut_cents"])
	}
	if m["fee_variance_cents"].(float64) != 0 {
		t.Errorf("fee_variance_cents = %v, want 0", m["fee_variance_cents"])
	}
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true", m["fee_guard_ok"])
	}
}

// TestMonthlyFees_GuardFlagsOvercharge: Shopify retained 22.9% (a 20%-tier rate) but the
// app's tier expects 17.9% → variance beyond 1% of gross trips the Fee Guard.
func TestMonthlyFees_GuardFlagsOvercharge(t *testing.T) {
	appID := uuid.New()
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev15,
		[]*entity.Transaction{feeTx(appID, 10000, 2290)}) // 22.9% actual vs 17.9% expected

	m := resp["months"].([]any)[0].(map[string]any)
	if m["fee_variance_cents"].(float64) != 500 {
		t.Errorf("fee_variance_cents = %v, want 500 (2290 actual − 1790 expected)", m["fee_variance_cents"])
	}
	if m["fee_guard_ok"] != false {
		t.Errorf("fee_guard_ok = %v, want false (overcharge exceeds 1%% of gross)", m["fee_guard_ok"])
	}
}

// TestMonthlyFees_ZeroTierProcessingNotFlagged: the RPT-FEES-1 misfire guard — a
// 0%-revenue-share app still has ~2.9% processing retained; the guard must NOT flag it
// (expected includes processing, so a 2.9% actual is within tolerance).
func TestMonthlyFees_ZeroTierProcessingNotFlagged(t *testing.T) {
	appID := uuid.New()
	// SMALL_DEV_0 (0% share); Shopify retained ~2.9% processing on $100 = ~$2.90.
	resp := doMonthlyFees(t, valueobject.RevenueShareTierSmallDev0,
		[]*entity.Transaction{feeTx(appID, 10000, 290)})

	m := resp["months"].([]any)[0].(map[string]any)
	if m["fee_guard_ok"] != true {
		t.Errorf("fee_guard_ok = %v, want true (0%% tier + 2.9%% processing must not misfire)", m["fee_guard_ok"])
	}
}

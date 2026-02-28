package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// StoreHealthHandler handles store health detail endpoints
type StoreHealthHandler struct {
	subscriptionRepo repository.SubscriptionRepository
	transactionRepo  repository.TransactionRepository
	partnerRepo      repository.PartnerAccountRepository
	appRepo          repository.AppRepository
}

// NewStoreHealthHandler creates a new StoreHealthHandler
func NewStoreHealthHandler(
	subscriptionRepo repository.SubscriptionRepository,
	transactionRepo repository.TransactionRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *StoreHealthHandler {
	return &StoreHealthHandler{
		subscriptionRepo: subscriptionRepo,
		transactionRepo:  transactionRepo,
		partnerRepo:      partnerRepo,
		appRepo:          appRepo,
	}
}

// StoreHealthResponse represents the store health API response
type StoreHealthResponse struct {
	Subscription *SubscriptionResponse   `json:"subscription"`
	Transactions []TransactionResponse   `json:"transactions"`
	Earnings     *EarningsSummaryResponse `json:"earnings"`
}

// SubscriptionResponse represents subscription data in response
type SubscriptionResponse struct {
	ID                   string     `json:"id"`
	ShopifyGID           string     `json:"shopify_gid"`
	MyshopifyDomain      string     `json:"myshopify_domain"`
	ShopName             string     `json:"shop_name"`
	PlanName             string     `json:"plan_name"`
	BasePriceCents       int64      `json:"base_price_cents"`
	BillingInterval      string     `json:"billing_interval"`
	RiskState            string     `json:"risk_state"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	LastChargeDate       *time.Time `json:"last_charge_date,omitempty"`
	ExpectedNextCharge   *time.Time `json:"expected_next_charge,omitempty"`
}

// TransactionResponse represents transaction data in response
type TransactionResponse struct {
	ID               string    `json:"id"`
	ShopifyGID       string    `json:"shopify_gid"`
	ChargeType       string    `json:"charge_type"`
	GrossAmountCents int64     `json:"gross_amount_cents"`
	NetAmountCents   int64     `json:"net_amount_cents"`
	Currency         string    `json:"currency"`
	TransactionDate  time.Time `json:"transaction_date"`
	EarningsStatus   string    `json:"earnings_status"`
	AvailableDate    time.Time `json:"available_date,omitempty"`
}

// EarningsSummaryResponse represents earnings summary in response
type EarningsSummaryResponse struct {
	PendingCents   int64 `json:"pending_cents"`
	AvailableCents int64 `json:"available_cents"`
	PaidOutCents   int64 `json:"paid_out_cents"`
	TotalCents     int64 `json:"total_cents"`
}

// GetStoreHealth returns health details for a specific store
// GET /api/v1/apps/{appID}/stores/{domain}/health
func (h *StoreHealthHandler) GetStoreHealth(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	// Get partner account
	partnerAccount, err := h.partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "no partner account found")
		return
	}

	// Get numeric appID from URL and construct full GID
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "app ID is required")
		return
	}
	fullAppGID := subscriptionAppGIDPrefix + appIDStr

	// Find app by partner app ID (GID)
	app, err := h.appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "app not found")
		return
	}

	// Get domain from URL
	domain := chi.URLParam(r, "domain")
	if domain == "" {
		writeJSONError(w, http.StatusBadRequest, "domain is required")
		return
	}

	// Fetch subscription by domain
	subscription, err := h.subscriptionRepo.FindByAppIDAndDomain(r.Context(), app.ID, domain)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "subscription not found for this domain")
		return
	}

	// Fetch transactions for the last 3 months
	now := time.Now().UTC()
	from := now.AddDate(0, -3, 0)
	transactions, err := h.transactionRepo.FindByDomain(r.Context(), app.ID, domain, from, now)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	// Fetch earnings summary for this store
	earnings, err := h.transactionRepo.GetEarningsSummaryByDomain(r.Context(), app.ID, domain)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch earnings summary")
		return
	}

	// Build response
	response := StoreHealthResponse{
		Subscription: subscriptionToResponse(subscription),
		Transactions: transactionsToResponse(transactions),
		Earnings:     earningsToResponse(earnings),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func subscriptionToResponse(sub *entity.Subscription) *SubscriptionResponse {
	if sub == nil {
		return nil
	}
	return &SubscriptionResponse{
		ID:                 sub.ID.String(),
		ShopifyGID:         sub.ShopifyGID,
		MyshopifyDomain:    sub.MyshopifyDomain,
		ShopName:           sub.ShopName,
		PlanName:           sub.PlanName,
		BasePriceCents:     sub.BasePriceCents,
		BillingInterval:    string(sub.BillingInterval),
		RiskState:          string(sub.RiskState),
		Status:             sub.Status,
		CreatedAt:          sub.CreatedAt,
		LastChargeDate:     sub.LastRecurringChargeDate,
		ExpectedNextCharge: sub.ExpectedNextChargeDate,
	}
}

func transactionsToResponse(txs []*entity.Transaction) []TransactionResponse {
	responses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = TransactionResponse{
			ID:               tx.ID.String(),
			ShopifyGID:       tx.ShopifyGID,
			ChargeType:       tx.ChargeType.String(),
			GrossAmountCents: tx.GrossAmountCents,
			NetAmountCents:   tx.NetAmountCents,
			Currency:         tx.Currency,
			TransactionDate:  tx.TransactionDate,
			EarningsStatus:   string(tx.EarningsStatus),
			AvailableDate:    tx.AvailableDate,
		}
	}
	return responses
}

func earningsToResponse(earnings *repository.EarningsSummary) *EarningsSummaryResponse {
	if earnings == nil {
		return nil
	}
	return &EarningsSummaryResponse{
		PendingCents:   earnings.PendingCents,
		AvailableCents: earnings.AvailableCents,
		PaidOutCents:   earnings.PaidOutCents,
		TotalCents:     earnings.PendingCents + earnings.AvailableCents + earnings.PaidOutCents,
	}
}

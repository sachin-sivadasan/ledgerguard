package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
)

type TransactionHandler struct {
	txRepo      repository.TransactionRepository
	partnerRepo repository.PartnerAccountRepository
	appRepo     repository.AppRepository
}

func NewTransactionHandler(
	txRepo repository.TransactionRepository,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *TransactionHandler {
	return &TransactionHandler{
		txRepo:      txRepo,
		partnerRepo: partnerRepo,
		appRepo:     appRepo,
	}
}

// List returns transactions for an app within a date range.
// GET /api/v1/apps/{appID}/transactions
// Query params: start, end (ISO dates, default last 365 days)
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	appID, lookupErr := lookupAppID(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	now := time.Now().UTC()
	from := now.AddDate(-1, 0, 0)
	to := now

	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			from = t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			from = t
		}
	}
	if s := r.URL.Query().Get("end"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			to = t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			to = t
		}
	}

	txs, err := h.txRepo.FindByAppID(r.Context(), appID, from, to)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}

	type txJSON struct {
		ID               string `json:"id"`
		Date             string `json:"date"`
		ShopDomain       string `json:"shop_domain"`
		ChargeType       string `json:"charge_type"`
		AppID            string `json:"app_id"`
		GrossAmountCents int64  `json:"gross_amount_cents"`
		NetAmountCents   int64  `json:"net_amount_cents"`
	}

	items := make([]txJSON, 0, len(txs))
	for _, tx := range txs {
		items = append(items, txJSON{
			ID:               tx.ID.String(),
			Date:             tx.TransactionDate.Format(time.RFC3339),
			ShopDomain:       tx.MyshopifyDomain,
			ChargeType:       string(tx.ChargeType),
			AppID:            tx.AppID.String(),
			GrossAmountCents: tx.GrossAmountCents,
			NetAmountCents:   tx.NetAmountCents,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": items,
	})
}

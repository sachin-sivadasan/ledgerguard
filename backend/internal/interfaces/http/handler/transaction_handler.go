package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	domainservice "github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
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

// List returns paginated transactions for an app.
// GET /api/v1/apps/{appID}/transactions
// Query params: start, end (ISO dates), page (default 1), pageSize (default 20, max 100), chargeType
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	filters := parseTransactionFilters(r)

	result, err := h.txRepo.FindByAppIDPaginated(r.Context(), app.ID, filters)
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

	items := make([]txJSON, 0, len(result.Transactions))
	for _, tx := range result.Transactions {
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
		"total":        result.Total,
		"page":         result.Page,
		"pageSize":     result.PageSize,
		"totalPages":   result.TotalPages,
	})
}

// Summary returns aggregate Gross/Net/Shopify-Cut totals over the FULL filtered set
// (server-side), so the UI totals don't reflect only the loaded page.
// GET /api/v1/apps/{appID}/transactions/summary  (same start/end/chargeType filters as List)
func (h *TransactionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	summary, err := h.txRepo.GetTransactionSummary(r.Context(), app.ID, parseTransactionFilters(r))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch transaction summary")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"grossCents":     summary.GrossCents,
		"netCents":       summary.NetCents,
		"shopifyCutCents": summary.GrossCents - summary.NetCents,
		"count":          summary.Count,
	})
}

// parseTransactionFilters builds TransactionFilters from the request. The default window
// is the app's ENTIRE stored history (SyncHistoryStart) — the list previously defaulted to
// the last 12 months, which under-counted now that we sync full history. Overridable via
// ?start/?end.
func parseTransactionFilters(r *http.Request) repository.TransactionFilters {
	filters := repository.TransactionFilters{
		From:     domainservice.SyncHistoryStart,
		To:       time.Now().UTC(),
		Page:     1,
		PageSize: 20,
	}
	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filters.From = t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.From = t
		}
	}
	if s := r.URL.Query().Get("end"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filters.To = t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.To = t
		}
	}
	if s := r.URL.Query().Get("chargeType"); s != "" {
		filters.ChargeType = s
	}
	if s := r.URL.Query().Get("page"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 {
			filters.Page = parsed
		}
	}
	if s := r.URL.Query().Get("pageSize"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 100 {
			filters.PageSize = parsed
		}
	}
	return filters
}

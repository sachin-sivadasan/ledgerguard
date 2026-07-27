package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// EarningsReportHandler serves the "Earnings" report (REPORTS.md — net earnings
// split by payout status: pending / available / paid, plus a per-charge table).
// Mirrors the RetentionHandler structure and reuses the stateless
// service.EarningsCalculator over the stored, pipeline-computed EarningsStatus.
type EarningsReportHandler struct {
	txRepo      repository.TransactionRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewEarningsReportHandler constructs an EarningsReportHandler.
func NewEarningsReportHandler(
	txRepo repository.TransactionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *EarningsReportHandler {
	return &EarningsReportHandler{
		txRepo:      txRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// earningsCharge is a single per-transaction row in the earnings report.
type earningsCharge struct {
	Date          string `json:"date"`
	Domain        string `json:"domain"`
	ShopName      string `json:"shopName"`
	GrossCents    int64  `json:"grossCents"`
	NetCents      int64  `json:"netCents"`
	Status        string `json:"status"`
	AvailableDate string `json:"availableDate"`
}

// earningsReport is the full JSON contract for the Earnings report.
type earningsReport struct {
	Currency         string           `json:"currency"`
	NetEarningsCents int64            `json:"netEarningsCents"`
	PendingCents     int64            `json:"pendingCents"`
	AvailableCents   int64            `json:"availableCents"`
	PaidOutCents     int64            `json:"paidOutCents"`
	Charges          []earningsCharge `json:"charges"`
}

// GetEarningsReport returns the Earnings report for an app.
// GET /api/v1/apps/{appID}/reports/earnings?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *EarningsReportHandler) GetEarningsReport(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeJSONError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	now := time.Now().UTC()
	from, to := parseDateRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"), now)

	txs, err := h.txRepo.FindByAppID(r.Context(), app.ID, from, to)
	if err != nil {
		writeEarningsRepoError(w, "FindByAppID", err)
		return
	}

	calc := service.NewEarningsCalculator()
	summary := calc.SummarizeEarnings(txs)

	report := earningsReport{
		Currency:         earningsCurrency(txs),
		NetEarningsCents: summary.TotalCents(),
		PendingCents:     summary.PendingCents,
		AvailableCents:   summary.AvailableCents,
		PaidOutCents:     summary.PaidOutCents,
		Charges:          buildEarningsCharges(txs),
	}

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeEarningsChargesCSV(w, report.Charges)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("earnings: encode report: %v", err)
	}
}

// writeEarningsRepoError logs a repository failure and responds 503. The transaction
// repo has no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeEarningsRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("earnings: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// earningsCurrency returns the first transaction's non-empty currency, defaulting
// to "USD" when no transaction carries one.
func earningsCurrency(txs []*entity.Transaction) string {
	for _, tx := range txs {
		if tx.Currency != "" {
			return tx.Currency
		}
	}
	return "USD"
}

// buildEarningsCharges converts each transaction to a report row, sorted by charge
// date (CreatedAt) descending. Initialized with make so an empty set serializes []
// rather than null.
func buildEarningsCharges(txs []*entity.Transaction) []earningsCharge {
	charges := make([]earningsCharge, 0, len(txs))
	for _, tx := range txs {
		charges = append(charges, earningsCharge{
			Date:          tx.CreatedAt.Format(dateLayout),
			Domain:        tx.MyshopifyDomain,
			ShopName:      tx.ShopName,
			GrossCents:    tx.GrossAmountCents,
			NetCents:      tx.NetAmountCents,
			Status:        earningsStatusLabel(tx.EarningsStatus),
			AvailableDate: tx.AvailableDate.Format(dateLayout),
		})
	}
	sort.SliceStable(charges, func(i, j int) bool {
		return charges[i].Date > charges[j].Date
	})
	return charges
}

// earningsStatusLabel maps a stored EarningsStatus to a UI-friendly label. Unknown
// statuses pass through title-cased so nothing is silently dropped.
func earningsStatusLabel(status entity.EarningsStatus) string {
	switch status {
	case entity.EarningsStatusPending:
		return "Pending"
	case entity.EarningsStatusAvailable:
		return "Available"
	case entity.EarningsStatusPaidOut:
		return "Paid"
	default:
		s := string(status)
		if s == "" {
			return ""
		}
		return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
}

// writeEarningsChargesCSV writes the per-charge table as a CSV attachment. Uses
// encoding/csv so free-text shop names/domains with commas/quotes stay one column.
func writeEarningsChargesCSV(w http.ResponseWriter, charges []earningsCharge) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="earnings.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"date", "store", "gross", "net", "status", "availableDate"})
	for _, c := range charges {
		_ = cw.Write([]string{
			c.Date,
			c.ShopName,
			strconv.FormatInt(c.GrossCents, 10),
			strconv.FormatInt(c.NetCents, 10),
			c.Status,
			c.AvailableDate,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("earnings: write CSV: %v", err)
	}
}

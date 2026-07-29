package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// ActivationReportHandler serves the "Activation" report (REPORTS.md — Growth,
// Archetype E funnel): the install-to-paid conversion funnel. It joins app-events ↔
// subscriptions to count three NESTED stages of the install cohort:
//
//	Installs             = distinct shops with a RELATIONSHIP_INSTALLED event in-window
//	Started Subscription = those that ALSO have a SUBSCRIPTION_CHARGE_ACCEPTED event
//	                       (merchant approved a plan / began billing)
//	Paid / Recurring     = those whose subscription ALSO reached a first recurring charge
//
// Sourcing the middle stage from the SUBSCRIPTION_CHARGE_ACCEPTED event (rather than the
// subscriptions table) is what keeps "Started" distinct from "Paid": a subscription
// record only exists once ≥1 RECURRING charge lands (ledger rebuild), so a merchant who
// accepted a charge whose first recurring charge is pending/failed sits in Started but
// not Paid. Mirrors InstallsReportHandler (events + subscriptions, no snapshots).
//
// Data caveat: app-events are ingested per-subscription and dev-capped (see future.md),
// so the funnel renders sparse/empty with current live data until event coverage widens.
type ActivationReportHandler struct {
	subRepo     repository.SubscriptionRepository
	eventRepo   repository.AppEventRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewActivationReportHandler constructs an ActivationReportHandler.
func NewActivationReportHandler(
	subRepo repository.SubscriptionRepository,
	eventRepo repository.AppEventRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *ActivationReportHandler {
	return &ActivationReportHandler{
		subRepo:     subRepo,
		eventRepo:   eventRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// activationStage is a single funnel stage row.
type activationStage struct {
	Key        string  `json:"key"`   // "installs" | "started" | "paid"
	Label      string  `json:"label"` // human label for the funnel bar
	Count      int     `json:"count"`
	PctOfPrior float64 `json:"pctOfPrior"` // stage ÷ prior stage (installs = 1.0; 0 when prior is 0)
}

// activationReport is the full JSON contract for the Activation funnel report.
type activationReport struct {
	Installs int `json:"installs"`
	Started  int `json:"started"`
	Paid     int `json:"paid"`
	// OverallPct = paid ÷ installs; InstallToSubPct = started ÷ installs;
	// SubToPaidPct = paid ÷ started. Fractions in [0,1]; the frontend formats as %.
	OverallPct      float64           `json:"overallPct"`
	InstallToSubPct float64           `json:"installToSubPct"`
	SubToPaidPct    float64           `json:"subToPaidPct"`
	Stages          []activationStage `json:"stages"`
}

// GetActivation returns the Activation funnel report for an app.
// GET /api/v1/apps/{appID}/reports/activation?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *ActivationReportHandler) GetActivation(w http.ResponseWriter, r *http.Request) {
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

	subs, err := h.subRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeActivationRepoError(w, "FindByAppID", err)
		return
	}

	events, err := h.eventRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeActivationRepoError(w, "FindByAppID(events)", err)
		return
	}

	report := buildActivationReport(events, subs, from, to)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeActivationCSV(w, report)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("activation: encode report: %v", err)
	}
}

// writeActivationRepoError logs a repository failure and responds 503. These repos have
// no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeActivationRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("activation: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// buildActivationReport computes the three nested funnel stages over the [from,to] window
// (whole `to` day inclusive, matching the other event reports' boundary).
//
// Shop identity: app-events store a shop key that may be a real shop GID (webhook path) or
// a myshopify domain (per-subscription sync path) for the SAME shop. To collapse both to
// one identity, each event's shop key is canonicalised to the correlated subscription's
// domain when the join hits (else the raw key is used). This keeps a shop's install event
// and its charge-accepted event on the same funnel identity.
func buildActivationReport(events []*entity.AppEvent, subs []*entity.Subscription, from, to time.Time) activationReport {
	toExclusive := to.AddDate(0, 0, 1)
	subsByShop := indexSubsByShop(subs)

	// canon resolves a raw event shop key to a stable identity (domain when a subscription
	// correlates), so install (often GID-keyed) and charge-accepted (often domain-keyed)
	// events for one shop share a key.
	canon := func(shopKey string) string {
		if sub := subsByShop[shopKey]; sub != nil && sub.MyshopifyDomain != "" {
			return sub.MyshopifyDomain
		}
		return shopKey
	}

	installShops := map[string]bool{}
	acceptedShops := map[string]bool{}
	for _, e := range events {
		if e.OccurredAt.Before(from) || !e.OccurredAt.Before(toExclusive) {
			continue
		}
		switch strings.ToUpper(strings.TrimSpace(e.EventType)) {
		case "RELATIONSHIP_INSTALLED":
			installShops[canon(e.ShopifyShopGID)] = true
		case "SUBSCRIPTION_CHARGE_ACCEPTED":
			acceptedShops[canon(e.ShopifyShopGID)] = true
		}
	}

	// paidShopKeys: shops whose subscription reached a first recurring charge. Keyed by the
	// domain (the canonical identity, matching canon's output) and — only when the sub has
	// no domain — by its raw shop GID, so a funnel shop matches whichever identity its
	// events carry. (canon(sub.ShopifyShopGID) resolves back to this sub's domain when it
	// has one, so the second insert is a no-op then and only adds the raw GID otherwise.)
	paidShopKeys := map[string]bool{}
	for _, sub := range subs {
		if !subReachedRecurringCharge(sub) {
			continue
		}
		if sub.MyshopifyDomain != "" {
			paidShopKeys[sub.MyshopifyDomain] = true
		}
		if sub.ShopifyShopGID != "" {
			paidShopKeys[canon(sub.ShopifyShopGID)] = true
		}
	}

	// Nested stages: paid ⊆ started ⊆ installs.
	installs := len(installShops)
	started, paid := 0, 0
	for shop := range installShops {
		if !acceptedShops[shop] {
			continue
		}
		started++
		if paidShopKeys[shop] {
			paid++
		}
	}

	return activationReport{
		Installs:        installs,
		Started:         started,
		Paid:            paid,
		OverallPct:      ratio(paid, installs),
		InstallToSubPct: ratio(started, installs),
		SubToPaidPct:    ratio(paid, started),
		Stages: []activationStage{
			{Key: "installs", Label: "Installs", Count: installs, PctOfPrior: 1.0},
			{Key: "started", Label: "Started Subscription", Count: started, PctOfPrior: ratio(started, installs)},
			{Key: "paid", Label: "Paid / Recurring", Count: paid, PctOfPrior: ratio(paid, started)},
		},
	}
}

// subReachedRecurringCharge reports whether a subscription reached a first recurring
// charge — the "Paid / Recurring" signal. LastRecurringChargeDate is set whenever the
// ledger rebuild sees a RECURRING transaction for the store.
func subReachedRecurringCharge(sub *entity.Subscription) bool {
	return sub != nil && sub.LastRecurringChargeDate != nil
}

// ratio returns num/den as a fraction in [0,1], or 0 when den is 0 (no divide-by-zero).
func ratio(num, den int) float64 {
	if den <= 0 {
		return 0
	}
	return float64(num) / float64(den)
}

// writeActivationCSV writes the funnel stages as a CSV attachment.
func writeActivationCSV(w http.ResponseWriter, report activationReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="activation.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"stage", "count", "pctOfPrior", "pctOfInstalls"})
	for _, s := range report.Stages {
		_ = cw.Write([]string{
			s.Label,
			strconv.Itoa(s.Count),
			strconv.FormatFloat(s.PctOfPrior, 'f', 4, 64),
			strconv.FormatFloat(ratio(s.Count, report.Installs), 'f', 4, 64),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("activation: write CSV: %v", err)
	}
}

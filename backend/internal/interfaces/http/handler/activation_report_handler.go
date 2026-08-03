package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// ActivationReportHandler serves the "Activation" report (REPORTS.md — Growth):
// the all-time install→paid conversion funnel. Two stages, computed via the SAME
// helper as the Installs report's conversion headline so the two never disagree:
//
//	Installs         = distinct shops that ever installed (lifetime base, de-fragmented)
//	Paid / Recurring = those that reached a first recurring charge (a paying subscription)
//
// RPT-ACTIVATION-1: the previous 3-stage funnel keyed its middle "Started" stage on
// SUBSCRIPTION_CHARGE_ACCEPTED events, which this account's Partner stream barely emits
// (~120 shops), throttling paid to 40 vs the real ~2,931. The reliable paid signal is the
// subscriptions table (real recurring charges), not the sparse charge-events, so the
// unreliable middle stage was dropped.
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
	Key        string  `json:"key"`   // "installs" | "paid"
	Label      string  `json:"label"` // human label for the funnel bar
	Count      int     `json:"count"`
	PctOfPrior float64 `json:"pctOfPrior"` // stage ÷ prior stage (installs = 1.0; 0 when prior is 0)
}

// activationReport is the full JSON contract for the Activation funnel report.
type activationReport struct {
	Installs int `json:"installs"`
	Paid     int `json:"paid"`
	// OverallPct = paid ÷ installs, a fraction in [0,1] (frontend formats as %).
	OverallPct float64           `json:"overallPct"`
	Stages     []activationStage `json:"stages"`
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

	report := buildActivationReport(events, subs)

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

// buildActivationReport computes the all-time 2-stage Installs → Paid funnel. It reuses
// computeLifecycleAndConversion (the Installs report's conversion helper) so the funnel
// and the Installs page's conversion headline are identical by construction — same
// de-fragmented lifetime install base, same distinct-paying-shop count.
func buildActivationReport(events []*entity.AppEvent, subs []*entity.Subscription) activationReport {
	_, conv := computeLifecycleAndConversion(events, subs)
	installs, paid := conv.Installs, conv.Paid

	return activationReport{
		Installs:   installs,
		Paid:       paid,
		OverallPct: conv.Rate,
		Stages: []activationStage{
			{Key: "installs", Label: "Installs", Count: installs, PctOfPrior: 1.0},
			{Key: "paid", Label: "Paid / Recurring", Count: paid, PctOfPrior: conv.Rate},
		},
	}
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

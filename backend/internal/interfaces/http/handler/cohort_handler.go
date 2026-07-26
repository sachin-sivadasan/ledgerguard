package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type CohortHandler struct {
	subscriptionRepo repository.SubscriptionRepository
	appRepo          repository.AppRepository
	partnerRepo      repository.PartnerAccountRepository
}

func NewCohortHandler(
	subscriptionRepo repository.SubscriptionRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *CohortHandler {
	return &CohortHandler{
		subscriptionRepo: subscriptionRepo,
		appRepo:          appRepo,
		partnerRepo:      partnerRepo,
	}
}

// GetCohorts returns cohort retention data for the specified app.
// GET /api/v1/apps/{appID}/cohorts?months=6
func (h *CohortHandler) GetCohorts(w http.ResponseWriter, r *http.Request) {
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

	months := 6
	if m := r.URL.Query().Get("months"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 2 && parsed <= 24 {
			months = parsed
		}
	}

	// Fetch all subscriptions for this app
	subscriptions, err := h.subscriptionRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeCohortRepoError(w, "FindByAppID", err)
		return
	}

	now := time.Now()
	cohorts := buildCohorts(subscriptions, months, now)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeCohortsCSV(w, cohorts)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"cohorts": cohorts,
	}); err != nil {
		log.Printf("cohorts: encode response: %v", err)
	}
}

// writeCohortRepoError logs a repository failure and responds 503. This repo has no
// not-found sentinel — every error is an infrastructure failure (ADR-042), matching
// writeRetentionRepoError.
func writeCohortRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("cohorts: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// writeCohortsCSV writes the cohort retention heatmap as a CSV attachment. The header
// is cohort,initialStores,M0,M1,...,Mmax where max is the longest RetentionPcts across
// all cohorts. Ragged rows (a cohort with fewer months) are padded with empty strings
// so every row has the same column count. Uses encoding/csv for correct escaping.
func writeCohortsCSV(w http.ResponseWriter, cohorts []entity.CohortData) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="cohorts.csv"`)

	maxMonths := 0
	for _, c := range cohorts {
		if len(c.RetentionPcts) > maxMonths {
			maxMonths = len(c.RetentionPcts)
		}
	}

	header := make([]string, 0, 2+maxMonths)
	header = append(header, "cohort", "initialStores")
	for m := 0; m < maxMonths; m++ {
		header = append(header, "M"+strconv.Itoa(m))
	}

	cw := csv.NewWriter(w)
	_ = cw.Write(header)
	for _, c := range cohorts {
		row := make([]string, 0, len(header))
		row = append(row, c.CohortMonth, strconv.Itoa(c.InitialStores))
		for m := 0; m < maxMonths; m++ {
			if m < len(c.RetentionPcts) {
				row = append(row, strconv.FormatFloat(c.RetentionPcts[m], 'f', 1, 64))
			} else {
				row = append(row, "") // pad ragged row
			}
		}
		_ = cw.Write(row)
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("cohorts: write CSV: %v", err)
	}
}

// buildCohorts groups subscriptions by creation month and calculates retention.
func buildCohorts(subscriptions []*entity.Subscription, months int, now time.Time) []entity.CohortData {
	// Group subscriptions by cohort month (creation month)
	type cohortEntry struct {
		month         string
		monthTime     time.Time
		subscriptions []*entity.Subscription
	}

	cohortMap := map[string]*cohortEntry{}
	cutoffMonth := time.Date(now.Year(), now.Month()-time.Month(months-1), 1, 0, 0, 0, 0, time.UTC)

	for _, sub := range subscriptions {
		cohortMonth := time.Date(sub.CreatedAt.Year(), sub.CreatedAt.Month(), 1, 0, 0, 0, 0, time.UTC)
		if cohortMonth.Before(cutoffMonth) {
			continue
		}
		key := fmt.Sprintf("%d-%02d", cohortMonth.Year(), cohortMonth.Month())
		if cohortMap[key] == nil {
			cohortMap[key] = &cohortEntry{month: key, monthTime: cohortMonth}
		}
		cohortMap[key].subscriptions = append(cohortMap[key].subscriptions, sub)
	}

	// Sort cohorts chronologically
	sortedKeys := make([]string, 0, len(cohortMap))
	for k := range cohortMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Calculate retention for each cohort
	result := make([]entity.CohortData, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		entry := cohortMap[key]
		initial := len(entry.subscriptions)
		if initial == 0 {
			continue
		}

		// Calculate months elapsed since cohort month
		monthsElapsed := monthsBetween(entry.monthTime, now)

		retentionPcts := make([]float64, 0, monthsElapsed+1)
		for m := 0; m <= monthsElapsed; m++ {
			checkDate := entry.monthTime.AddDate(0, m, 0)
			active := 0
			for _, sub := range entry.subscriptions {
				if isActiveAt(sub, checkDate) {
					active++
				}
			}
			pct := float64(active) / float64(initial) * 100
			retentionPcts = append(retentionPcts, pct)
		}

		result = append(result, entity.CohortData{
			CohortMonth:   key,
			InitialStores: initial,
			RetentionPcts: retentionPcts,
		})
	}

	return result
}

func monthsBetween(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	return years*12 + months
}

// isActiveAt checks if a subscription was still active at the given date.
// A subscription is active if it's not churned (risk state is not CHURNED)
// or if it was churned after the check date.
func isActiveAt(sub *entity.Subscription, date time.Time) bool {
	// If subscription was created after the check date, not active
	if sub.CreatedAt.After(date) {
		return false
	}

	// If subscription is currently active (safe or at-risk), it was active at this date
	if sub.RiskState == valueobject.RiskStateSafe ||
		sub.RiskState == valueobject.RiskStateOneCycleMissed ||
		sub.RiskState == valueobject.RiskStateTwoCyclesMissed {
		return true
	}

	// For churned subscriptions, check if they churned after the date
	// Use UpdatedAt as a proxy for when the churn happened
	if sub.RiskState == valueobject.RiskStateChurned {
		return sub.UpdatedAt.After(date)
	}

	return true
}

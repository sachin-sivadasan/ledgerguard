package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// recentInstallEventsLimit caps the recent-events table. KPIs and the trend use ALL
// in-range events; only this table is truncated (newest first), and the true totals
// remain visible in the Installs/Uninstalls KPIs.
const recentInstallEventsLimit = 50

// InstallsReportHandler serves the "Installs" report (REPORTS.md — Growth, Archetype A):
// install vs uninstall activity over a period with a daily trend and a recent-events
// table. Derived from RELATIONSHIP_INSTALLED / RELATIONSHIP_UNINSTALLED app-events,
// correlated to subscriptions via ShopifyShopGID only to resolve a human-readable store
// domain. Mirrors the UninstallContextHandler structure (events + subscriptions, no
// snapshots).
type InstallsReportHandler struct {
	subRepo     repository.SubscriptionRepository
	eventRepo   repository.AppEventRepository
	appRepo     repository.AppRepository
	partnerRepo repository.PartnerAccountRepository
}

// NewInstallsReportHandler constructs an InstallsReportHandler.
func NewInstallsReportHandler(
	subRepo repository.SubscriptionRepository,
	eventRepo repository.AppEventRepository,
	appRepo repository.AppRepository,
	partnerRepo repository.PartnerAccountRepository,
) *InstallsReportHandler {
	return &InstallsReportHandler{
		subRepo:     subRepo,
		eventRepo:   eventRepo,
		appRepo:     appRepo,
		partnerRepo: partnerRepo,
	}
}

// installTrendPoint is a single day in the install/uninstall time series.
type installTrendPoint struct {
	Date       string `json:"date"`
	Installs   int    `json:"installs"`
	Uninstalls int    `json:"uninstalls"`
}

// installEvent is a single row in the recent install/uninstall events table.
type installEvent struct {
	Domain string `json:"domain"`
	Event  string `json:"event"` // "Install" or "Uninstall"
	Date   string `json:"date"`
}

// installsReport is the full JSON contract for the Installs report.
type installsReport struct {
	Installs   int                 `json:"installs"`
	Uninstalls int                 `json:"uninstalls"`
	Net        int                 `json:"net"`
	Trend      []installTrendPoint `json:"trend"`
	Events     []installEvent      `json:"events"`
}

// GetInstalls returns the Installs report for an app.
// GET /api/v1/apps/{appID}/reports/installs?from=YYYY-MM-DD&to=YYYY-MM-DD&format=csv
func (h *InstallsReportHandler) GetInstalls(w http.ResponseWriter, r *http.Request) {
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
		writeInstallsRepoError(w, "FindByAppID", err)
		return
	}

	events, err := h.eventRepo.FindByAppID(r.Context(), app.ID)
	if err != nil {
		writeInstallsRepoError(w, "FindByAppID(events)", err)
		return
	}

	report := buildInstallsReport(events, indexSubsByShop(subs), from, to)

	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		writeInstallEventsCSV(w, report.Events)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		log.Printf("installs: encode report: %v", err)
	}
}

// writeInstallsRepoError logs a repository failure and responds 503. These repos have
// no not-found sentinel — every error is an infrastructure failure (ADR-042).
func writeInstallsRepoError(w http.ResponseWriter, op string, err error) {
	log.Printf("installs: repo error in %s: %v", op, err)
	writeJSONError(w, http.StatusServiceUnavailable, "service temporarily unavailable")
}

// installEventKind classifies an app event as "Install", "Uninstall", or "" (neither).
// It matches the exact Shopify relationship event types rather than a substring, because
// (a) "RELATIONSHIP_UNINSTALLED" contains the substring "INSTALL" (a substring match
// would misclassify every uninstall), and (b) the synced app_events table is NOT filtered
// to lifecycle types — it also holds RELATIONSHIP_REACTIVATED / RELATIONSHIP_DEACTIVATED
// and SUBSCRIPTION_CHARGE_* events, which are intentionally excluded here (they are not
// installs/uninstalls). Exact matching keeps a future "RELATIONSHIP_REINSTALLED"-style
// type from silently counting as an install.
func installEventKind(eventType string) string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "RELATIONSHIP_INSTALLED":
		return "Install"
	case "RELATIONSHIP_UNINSTALLED":
		return "Uninstall"
	default:
		return ""
	}
}

// buildInstallsReport counts install/uninstall events in the [from,to] window (whole
// `to` day inclusive, matching the other event reports' boundary), builds a daily trend
// (only days with activity, ascending), and a recent-events table (newest first, capped
// at recentInstallEventsLimit). Net = installs − uninstalls. A store domain is resolved
// from the correlated subscription when the join hits, falling back to the event's stored
// shop identifier (which may be a myshopify domain or a GID depending on the sync source).
func buildInstallsReport(events []*entity.AppEvent, subsByShop map[string]*entity.Subscription, from, to time.Time) installsReport {
	toExclusive := to.AddDate(0, 0, 1)

	type dayAgg struct{ installs, uninstalls int }
	byDay := map[string]*dayAgg{}

	var installs, uninstalls int
	matched := make([]*entity.AppEvent, 0)
	kindOf := map[*entity.AppEvent]string{}

	for _, e := range events {
		kind := installEventKind(e.EventType)
		if kind == "" {
			continue
		}
		if e.OccurredAt.Before(from) || !e.OccurredAt.Before(toExclusive) {
			continue
		}

		day := e.OccurredAt.Format(dateLayout)
		d, ok := byDay[day]
		if !ok {
			d = &dayAgg{}
			byDay[day] = d
		}
		if kind == "Uninstall" {
			uninstalls++
			d.uninstalls++
		} else {
			installs++
			d.installs++
		}
		matched = append(matched, e)
		kindOf[e] = kind
	}

	// Trend: only days with activity, ascending (date keys are YYYY-MM-DD → lexical
	// order == chronological).
	days := make([]string, 0, len(byDay))
	for day := range byDay {
		days = append(days, day)
	}
	sort.Strings(days)
	trend := make([]installTrendPoint, 0, len(days))
	for _, day := range days {
		d := byDay[day]
		trend = append(trend, installTrendPoint{Date: day, Installs: d.installs, Uninstalls: d.uninstalls})
	}

	// Recent events: newest first (by exact OccurredAt), capped.
	sort.SliceStable(matched, func(i, j int) bool {
		return matched[i].OccurredAt.After(matched[j].OccurredAt)
	})
	recent := make([]installEvent, 0, len(matched))
	for i, e := range matched {
		if i >= recentInstallEventsLimit {
			break
		}
		domain := e.ShopifyShopGID
		if sub := subsByShop[e.ShopifyShopGID]; sub != nil && sub.MyshopifyDomain != "" {
			domain = sub.MyshopifyDomain
		}
		recent = append(recent, installEvent{
			Domain: domain,
			Event:  kindOf[e],
			Date:   e.OccurredAt.Format(dateLayout),
		})
	}

	return installsReport{
		Installs:   installs,
		Uninstalls: uninstalls,
		Net:        installs - uninstalls,
		Trend:      trend,
		Events:     recent,
	}
}

// writeInstallEventsCSV writes the recent install/uninstall events as a CSV attachment.
func writeInstallEventsCSV(w http.ResponseWriter, events []installEvent) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="installs.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"store", "event", "date"})
	for _, e := range events {
		_ = cw.Write([]string{e.Domain, e.Event, e.Date})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("installs: write CSV: %v", err)
	}
}

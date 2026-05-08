package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// ExportHandler handles data export endpoints
type ExportHandler struct {
	exportService *service.ExportService
	auditService  *service.AuditService
	partnerRepo   repository.PartnerAccountRepository
	appRepo       repository.AppRepository
}

// NewExportHandler creates a new ExportHandler
func NewExportHandler(
	exportService *service.ExportService,
	auditService *service.AuditService,
	partnerRepo repository.PartnerAccountRepository,
	appRepo repository.AppRepository,
) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		auditService:  auditService,
		partnerRepo:   partnerRepo,
		appRepo:       appRepo,
	}
}

// ExportTransactions handles GET /api/v1/apps/{appID}/export/transactions
// Query params: start (required), end (required), format (optional: csv|json, default: csv)
func (h *ExportHandler) ExportTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user
	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeExportError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeExportError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Parse date range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		writeExportError(w, http.StatusBadRequest, "start and end dates are required (YYYY-MM-DD)")
		return
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		writeExportError(w, http.StatusBadRequest, "invalid start date format (expected YYYY-MM-DD)")
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		writeExportError(w, http.StatusBadRequest, "invalid end date format (expected YYYY-MM-DD)")
		return
	}
	// Include the entire end day
	end = end.Add(24*time.Hour - time.Second)

	// Parse format
	format := service.ExportFormatCSV
	if f := r.URL.Query().Get("format"); f != "" {
		switch f {
		case "csv":
			format = service.ExportFormatCSV
		case "json":
			format = service.ExportFormatJSON
		default:
			writeExportError(w, http.StatusBadRequest, "invalid format (expected csv or json)")
			return
		}
	}

	// Log export request
	h.auditService.LogExportRequest(ctx, user.ID, string(service.ExportTypeTransactions), &app.ID, r.RemoteAddr, r.UserAgent())

	// Perform export
	result, err := h.exportService.ExportTransactions(ctx, app.ID, start, end, format)
	if err != nil {
		writeExportError(w, http.StatusInternalServerError, "failed to export transactions")
		return
	}

	// Write response
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	w.Header().Set("X-Record-Count", formatInt(result.RecordCount))
	w.Write(result.Data)
}

// ExportSubscriptions handles GET /api/v1/apps/{appID}/export/subscriptions
// Query params: format (optional: csv|json, default: csv)
func (h *ExportHandler) ExportSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user
	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeExportError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeExportError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Parse format
	format := service.ExportFormatCSV
	if f := r.URL.Query().Get("format"); f != "" {
		switch f {
		case "csv":
			format = service.ExportFormatCSV
		case "json":
			format = service.ExportFormatJSON
		default:
			writeExportError(w, http.StatusBadRequest, "invalid format (expected csv or json)")
			return
		}
	}

	// Log export request
	h.auditService.LogExportRequest(ctx, user.ID, string(service.ExportTypeSubscriptions), &app.ID, r.RemoteAddr, r.UserAgent())

	// Perform export
	result, err := h.exportService.ExportSubscriptions(ctx, app.ID, format)
	if err != nil {
		writeExportError(w, http.StatusInternalServerError, "failed to export subscriptions")
		return
	}

	// Write response
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	w.Header().Set("X-Record-Count", formatInt(result.RecordCount))
	w.Write(result.Data)
}

// ExportMetrics handles GET /api/v1/apps/{appID}/export/metrics
// Query params: start (required), end (required), format (optional: csv|json, default: csv)
func (h *ExportHandler) ExportMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user
	user := middleware.UserFromContext(ctx)
	if user == nil {
		writeExportError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	app, lookupErr := resolveAppFromRequest(r, h.partnerRepo, h.appRepo)
	if lookupErr != nil {
		writeExportError(w, lookupErr.statusCode, lookupErr.message)
		return
	}

	// Parse date range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		writeExportError(w, http.StatusBadRequest, "start and end dates are required (YYYY-MM-DD)")
		return
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		writeExportError(w, http.StatusBadRequest, "invalid start date format (expected YYYY-MM-DD)")
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		writeExportError(w, http.StatusBadRequest, "invalid end date format (expected YYYY-MM-DD)")
		return
	}

	// Parse format
	format := service.ExportFormatCSV
	if f := r.URL.Query().Get("format"); f != "" {
		switch f {
		case "csv":
			format = service.ExportFormatCSV
		case "json":
			format = service.ExportFormatJSON
		default:
			writeExportError(w, http.StatusBadRequest, "invalid format (expected csv or json)")
			return
		}
	}

	// Log export request
	h.auditService.LogExportRequest(ctx, user.ID, string(service.ExportTypeMetrics), &app.ID, r.RemoteAddr, r.UserAgent())

	// Perform export
	result, err := h.exportService.ExportMetrics(ctx, app.ID, start, end, format)
	if err != nil {
		writeExportError(w, http.StatusInternalServerError, "failed to export metrics")
		return
	}

	// Write response
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	w.Header().Set("X-Record-Count", formatInt(result.RecordCount))
	w.Write(result.Data)
}

// formatInt formats an integer as a string
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

// writeExportError writes an error response for export endpoints
func writeExportError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}

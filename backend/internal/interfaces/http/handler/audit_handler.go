package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type OrgAuditHandler struct {
	auditService *service.OrgAuditService
}

func NewOrgAuditHandler(auditService *service.OrgAuditService) *OrgAuditHandler {
	return &OrgAuditHandler{auditService: auditService}
}

// ListAuditLog returns paginated audit entries for an org.
// GET /api/v1/orgs/:orgId/audit-log?limit=50&offset=0
func (h *OrgAuditHandler) ListAuditLog(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	if org == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	page, err := h.auditService.GetAuditLog(r.Context(), org.ID, limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch audit log")
		return
	}

	entries := make([]map[string]interface{}, 0, len(page.Entries))
	for _, e := range page.Entries {
		entry := map[string]interface{}{
			"id":         e.ID.String(),
			"actor_id":   e.ActorID.String(),
			"action":     e.Action,
			"created_at": e.CreatedAt,
		}
		if e.TargetType != "" {
			entry["target_type"] = e.TargetType
		}
		if e.TargetID != nil {
			entry["target_id"] = e.TargetID.String()
		}
		if e.Metadata != nil {
			entry["metadata"] = json.RawMessage(e.Metadata)
		}
		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"total":   page.Total,
		"limit":   page.Limit,
		"offset":  page.Offset,
	})
}

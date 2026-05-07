package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/valueobject"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type OrgHandler struct {
	orgService *service.OrgService
}

func NewOrgHandler(orgService *service.OrgService) *OrgHandler {
	return &OrgHandler{orgService: orgService}
}

// CreateOrg creates a new organization.
// POST /api/v1/orgs
func (h *OrgHandler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	org, err := h.orgService.CreateOrganization(r.Context(), req.Name, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(orgToJSON(org))
}

// ListOrgs returns all organizations the user belongs to.
// GET /api/v1/orgs
func (h *OrgHandler) ListOrgs(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	memberships, err := h.orgService.ListUserOrganizations(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}

	type orgMemberJSON struct {
		OrgID  string `json:"org_id"`
		Role   string `json:"role"`
		Status string `json:"status"`
	}

	items := make([]orgMemberJSON, 0, len(memberships))
	for _, m := range memberships {
		items = append(items, orgMemberJSON{
			OrgID:  m.OrgID.String(),
			Role:   string(m.Role),
			Status: string(m.Status),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"organizations": items,
	})
}

// GetOrg returns org details.
// GET /api/v1/orgs/:orgId
func (h *OrgHandler) GetOrg(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	if org == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgToJSON(org))
}

// UpdateOrg updates org name/settings.
// PUT /api/v1/orgs/:orgId
func (h *OrgHandler) UpdateOrg(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	if org == nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name != "" {
		org.Name = req.Name
	}

	if err := h.orgService.UpdateOrganization(r.Context(), org, user.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update organization")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgToJSON(org))
}

// DeleteOrg deletes an organization.
// DELETE /api/v1/orgs/:orgId
func (h *OrgHandler) DeleteOrg(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	if org == nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	if err := h.orgService.DeleteOrganization(r.Context(), org.ID, user.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to delete organization")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListMembers returns all members of an org.
// GET /api/v1/orgs/:orgId/members
func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	if org == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	members, err := h.orgService.ListMembers(r.Context(), org.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	items := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		item := map[string]interface{}{
			"id":        m.ID.String(),
			"user_id":   m.UserID.String(),
			"role":      string(m.Role),
			"status":    string(m.Status),
			"joined_at": m.JoinedAt,
		}
		if m.SuspendedAt != nil {
			item["suspended_at"] = m.SuspendedAt
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": items,
	})
}

// InviteMember sends an invitation.
// POST /api/v1/orgs/:orgId/invitations
func (h *OrgHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	if org == nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	role := valueobject.ParseOrgRole(req.Role)
	if role == "" {
		role = valueobject.OrgRoleViewer
	}

	inv, err := h.orgService.InviteMember(r.Context(), org.ID, req.Email, role, user.ID)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "failed to invite member"
		if err == service.ErrMemberLimitReached {
			status = http.StatusPaymentRequired
			msg = err.Error()
		}
		writeJSONError(w, status, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         inv.ID.String(),
		"email":      inv.Email,
		"role":       string(inv.Role),
		"token":      inv.Token,
		"expires_at": inv.ExpiresAt,
	})
}

// RevokeInvitation cancels a pending invitation.
// DELETE /api/v1/orgs/:orgId/invitations/:id
func (h *OrgHandler) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	invID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid invitation ID")
		return
	}

	if err := h.orgService.RevokeInvitation(r.Context(), invID, user.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to revoke invitation")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AcceptInvitation accepts an invitation via token.
// POST /api/v1/invitations/:token/accept
func (h *OrgHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		writeJSONError(w, http.StatusBadRequest, "token is required")
		return
	}

	member, err := h.orgService.AcceptInvitation(r.Context(), token, user.ID)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "failed to accept invitation"
		switch err {
		case service.ErrInvitationExpired:
			status = http.StatusGone
			msg = err.Error()
		case service.ErrInvitationNotPending:
			status = http.StatusConflict
			msg = err.Error()
		case service.ErrAlreadyMember:
			status = http.StatusConflict
			msg = err.Error()
		}
		writeJSONError(w, status, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"member_id": member.ID.String(),
		"org_id":    member.OrgID.String(),
		"role":      string(member.Role),
	})
}

// RemoveMember removes a member from the org.
// DELETE /api/v1/orgs/:orgId/members/:userId
func (h *OrgHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	if err := h.orgService.RemoveMember(r.Context(), org.ID, memberID, user.ID); err != nil {
		status := http.StatusInternalServerError
		msg := "failed to remove member"
		if err == service.ErrCannotRemoveOwner {
			status = http.StatusForbidden
			msg = err.Error()
		}
		writeJSONError(w, status, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangeRole updates a member's role.
// PUT /api/v1/orgs/:orgId/members/:userId/role
func (h *OrgHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role := valueobject.ParseOrgRole(req.Role)
	if role == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid role")
		return
	}

	if err := h.orgService.ChangeRole(r.Context(), org.ID, memberID, user.ID, role); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to change role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SuspendMember suspends a member.
// PUT /api/v1/orgs/:orgId/members/:userId/suspend
func (h *OrgHandler) SuspendMember(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	if err := h.orgService.SuspendMember(r.Context(), org.ID, memberID, user.ID); err != nil {
		status := http.StatusInternalServerError
		msg := "failed to suspend member"
		if err == service.ErrCannotSuspendOwner {
			status = http.StatusForbidden
			msg = err.Error()
		}
		writeJSONError(w, status, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnsuspendMember restores a suspended member.
// PUT /api/v1/orgs/:orgId/members/:userId/unsuspend
func (h *OrgHandler) UnsuspendMember(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	if err := h.orgService.UnsuspendMember(r.Context(), org.ID, memberID, user.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to unsuspend member")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateNotificationPrefs updates a member's notification preferences.
// PUT /api/v1/orgs/:orgId/members/:userId/notifications
func (h *OrgHandler) UpdateNotificationPrefs(w http.ResponseWriter, r *http.Request) {
	memberID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid member ID")
		return
	}

	var prefs json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.orgService.UpdateNotificationPrefs(r.Context(), memberID, prefs); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update notification preferences")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ConfigureWebhook sets the org webhook URL.
// PUT /api/v1/orgs/:orgId/webhooks
func (h *OrgHandler) ConfigureWebhook(w http.ResponseWriter, r *http.Request) {
	org := middleware.OrgFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	if org == nil || user == nil {
		writeJSONError(w, http.StatusNotFound, "organization not found")
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}

	if err := h.orgService.ConfigureWebhook(r.Context(), org.ID, req.URL, user.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to configure webhook")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func orgToJSON(org *entity.Organization) map[string]interface{} {
	return map[string]interface{}{
		"id":         org.ID.String(),
		"name":       org.Name,
		"slug":       org.Slug,
		"plan_tier":  string(org.PlanTier),
		"created_at": org.CreatedAt,
	}
}

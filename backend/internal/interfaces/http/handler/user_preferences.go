package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// UserPreferences represents dashboard preferences for a user
type UserPreferences struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	PrimaryKpis      []string   `json:"primary_kpis"`
	SecondaryWidgets []string   `json:"secondary_widgets"`
	DefaultAppID     *uuid.UUID `json:"default_app_id,omitempty"`
	SelectedOrgID    *uuid.UUID `json:"selected_org_id,omitempty"`
	AutoSync         bool       `json:"auto_sync"`
	SyncFrequency    string     `json:"sync_frequency"`
	WorkspaceName    string     `json:"workspace_name"`
	Currency         string     `json:"currency"`
	Timezone         string     `json:"timezone"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UserPreferencesHandler handles user preferences endpoints
type UserPreferencesHandler struct {
	db *pgxpool.Pool
}

// NewUserPreferencesHandler creates a new handler
func NewUserPreferencesHandler(db *pgxpool.Pool) *UserPreferencesHandler {
	return &UserPreferencesHandler{db: db}
}

// defaultPreferences returns the default dashboard preferences
func defaultPreferences() *UserPreferences {
	return &UserPreferences{
		PrimaryKpis: []string{
			"active_mrr",
			"renewal_success_rate",
			"revenue_at_risk",
			"usage_revenue",
		},
		SecondaryWidgets: []string{
			"mrr_trend",
			"risk_distribution_chart",
			"forecast",
			"revenue_mix_chart",
			"weekly_activity",
		},
	}
}

// GetDashboardPreferences returns the user's dashboard preferences
// GET /api/v1/user/preferences/dashboard
func (h *UserPreferencesHandler) GetDashboardPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	prefs, err := h.findByUserID(r.Context(), user.ID)
	if err != nil {
		// Return defaults if no preferences found
		prefs = defaultPreferences()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"primary_kpis":      prefs.PrimaryKpis,
		"secondary_widgets": prefs.SecondaryWidgets,
	})
}

// SaveDashboardPreferences saves or updates the user's dashboard preferences
// PUT /api/v1/user/preferences/dashboard
func (h *UserPreferencesHandler) SaveDashboardPreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		PrimaryKpis      []string `json:"primary_kpis"`
		SecondaryWidgets []string `json:"secondary_widgets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate primary_kpis (max 4)
	if len(req.PrimaryKpis) > 4 {
		writeJSONError(w, http.StatusBadRequest, "primary_kpis cannot exceed 4 items")
		return
	}

	err := h.upsert(r.Context(), user.ID, req.PrimaryKpis, req.SecondaryWidgets)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save preferences")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "preferences saved",
		"primary_kpis":      req.PrimaryKpis,
		"secondary_widgets": req.SecondaryWidgets,
	})
}

// findByUserID retrieves preferences for a user
func (h *UserPreferencesHandler) findByUserID(ctx context.Context, userID uuid.UUID) (*UserPreferences, error) {
	query := `
		SELECT id, user_id, primary_kpis, secondary_widgets, default_app_id, selected_org_id,
			auto_sync, sync_frequency, workspace_name, currency, timezone,
			created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`

	var prefs UserPreferences
	err := h.db.QueryRow(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.PrimaryKpis,
		&prefs.SecondaryWidgets,
		&prefs.DefaultAppID,
		&prefs.SelectedOrgID,
		&prefs.AutoSync,
		&prefs.SyncFrequency,
		&prefs.WorkspaceName,
		&prefs.Currency,
		&prefs.Timezone,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &prefs, nil
}

// upsert creates or updates preferences for a user
func (h *UserPreferencesHandler) upsert(ctx context.Context, userID uuid.UUID, primaryKpis, secondaryWidgets []string) error {
	query := `
		INSERT INTO user_preferences (user_id, primary_kpis, secondary_widgets, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			primary_kpis = EXCLUDED.primary_kpis,
			secondary_widgets = EXCLUDED.secondary_widgets,
			updated_at = NOW()
	`

	_, err := h.db.Exec(ctx, query, userID, primaryKpis, secondaryWidgets)
	return err
}

// GetDefaultApp returns the user's default app preference
// GET /api/v1/user/preferences/default-app
func (h *UserPreferencesHandler) GetDefaultApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	prefs, err := h.findByUserID(r.Context(), user.ID)
	if err != nil || prefs.DefaultAppID == nil {
		// No default app set
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"default_app_id": nil,
			"message":        "no default app set",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"default_app_id": prefs.DefaultAppID.String(),
	})
}

// SetDefaultApp sets the user's default app preference
// PUT /api/v1/user/preferences/default-app
func (h *UserPreferencesHandler) SetDefaultApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		DefaultAppID *string `json:"default_app_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var appID *uuid.UUID
	if req.DefaultAppID != nil && *req.DefaultAppID != "" {
		parsed, err := uuid.Parse(*req.DefaultAppID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid app_id format")
			return
		}
		appID = &parsed
	}

	err := h.setDefaultApp(r.Context(), user.ID, appID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save default app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if appID != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":        "default app updated",
			"default_app_id": appID.String(),
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":        "default app cleared",
			"default_app_id": nil,
		})
	}
}

// GetSelectedOrg returns the user's selected organization preference
// GET /api/v1/user/preferences/selected-org
func (h *UserPreferencesHandler) GetSelectedOrg(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	prefs, err := h.findByUserID(r.Context(), user.ID)
	if err != nil || prefs.SelectedOrgID == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"selected_org_id": nil,
			"message":         "no organization selected",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"selected_org_id": prefs.SelectedOrgID.String(),
	})
}

// SetSelectedOrg sets the user's selected organization preference
// PUT /api/v1/user/preferences/selected-org
func (h *UserPreferencesHandler) SetSelectedOrg(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		SelectedOrgID *string `json:"selected_org_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var orgID *uuid.UUID
	if req.SelectedOrgID != nil && *req.SelectedOrgID != "" {
		parsed, err := uuid.Parse(*req.SelectedOrgID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid org_id format")
			return
		}
		orgID = &parsed
	}

	err := h.setSelectedOrg(r.Context(), user.ID, orgID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save selected organization")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if orgID != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "selected organization updated",
			"selected_org_id": orgID.String(),
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "selected organization cleared",
			"selected_org_id": nil,
		})
	}
}

// setSelectedOrg updates the selected org in the database
func (h *UserPreferencesHandler) setSelectedOrg(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error {
	query := `
		INSERT INTO user_preferences (user_id, primary_kpis, secondary_widgets, selected_org_id, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			selected_org_id = EXCLUDED.selected_org_id,
			updated_at = NOW()
	`

	defaults := defaultPreferences()
	_, err := h.db.Exec(ctx, query, userID, defaults.PrimaryKpis, defaults.SecondaryWidgets, orgID)
	return err
}

// setDefaultApp updates the default app in the database
func (h *UserPreferencesHandler) setDefaultApp(ctx context.Context, userID uuid.UUID, appID *uuid.UUID) error {
	query := `
		INSERT INTO user_preferences (user_id, primary_kpis, secondary_widgets, default_app_id, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			default_app_id = EXCLUDED.default_app_id,
			updated_at = NOW()
	`

	defaults := defaultPreferences()
	_, err := h.db.Exec(ctx, query, userID, defaults.PrimaryKpis, defaults.SecondaryWidgets, appID)
	return err
}

// GetSyncWorkspacePreferences returns sync schedule + workspace settings
// GET /api/v1/user/preferences/settings
func (h *UserPreferencesHandler) GetSyncWorkspacePreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	prefs, err := h.findByUserID(r.Context(), user.ID)
	if err != nil {
		// Return defaults
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"auto_sync":      true,
			"sync_frequency": "Every 6 hours",
			"workspace_name": "My Shopify Apps",
			"currency":       "USD",
			"timezone":       "America/New_York",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"auto_sync":      prefs.AutoSync,
		"sync_frequency": prefs.SyncFrequency,
		"workspace_name": prefs.WorkspaceName,
		"currency":       prefs.Currency,
		"timezone":       prefs.Timezone,
	})
}

// SaveSyncWorkspacePreferences saves sync schedule + workspace settings
// PUT /api/v1/user/preferences/settings
func (h *UserPreferencesHandler) SaveSyncWorkspacePreferences(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		AutoSync      bool   `json:"auto_sync"`
		SyncFrequency string `json:"sync_frequency"`
		WorkspaceName string `json:"workspace_name"`
		Currency      string `json:"currency"`
		Timezone      string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.upsertSyncWorkspace(r.Context(), user.ID, req.AutoSync, req.SyncFrequency, req.WorkspaceName, req.Currency, req.Timezone)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "settings saved",
		"auto_sync":      req.AutoSync,
		"sync_frequency": req.SyncFrequency,
		"workspace_name": req.WorkspaceName,
		"currency":       req.Currency,
		"timezone":       req.Timezone,
	})
}

// upsertSyncWorkspace creates or updates sync/workspace settings
func (h *UserPreferencesHandler) upsertSyncWorkspace(ctx context.Context, userID uuid.UUID, autoSync bool, syncFrequency, workspaceName, currency, timezone string) error {
	query := `
		INSERT INTO user_preferences (user_id, primary_kpis, secondary_widgets, auto_sync, sync_frequency, workspace_name, currency, timezone, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			auto_sync = EXCLUDED.auto_sync,
			sync_frequency = EXCLUDED.sync_frequency,
			workspace_name = EXCLUDED.workspace_name,
			currency = EXCLUDED.currency,
			timezone = EXCLUDED.timezone,
			updated_at = NOW()
	`

	defaults := defaultPreferences()
	_, err := h.db.Exec(ctx, query, userID, defaults.PrimaryKpis, defaults.SecondaryWidgets, autoSync, syncFrequency, workspaceName, currency, timezone)
	return err
}

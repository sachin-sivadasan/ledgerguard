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
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	PrimaryKpis      []string  `json:"primary_kpis"`
	SecondaryWidgets []string  `json:"secondary_widgets"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
			"renewal_success_rate",
			"active_mrr",
			"revenue_at_risk",
			"churned",
		},
		SecondaryWidgets: []string{
			"usage_revenue",
			"total_revenue",
			"revenue_mix_chart",
			"risk_distribution_chart",
			"earnings_timeline",
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
		SELECT id, user_id, primary_kpis, secondary_widgets, created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`

	var prefs UserPreferences
	err := h.db.QueryRow(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.PrimaryKpis,
		&prefs.SecondaryWidgets,
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

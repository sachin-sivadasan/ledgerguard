package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DBChecker interface {
	Ping(ctx context.Context) error
}

// SchemaChecker is optionally implemented by the DB to report migration state.
// When available, the health check also fails on a dirty or unapplied schema
// (a pool ping alone can't detect a half-migrated database).
type SchemaChecker interface {
	MigrationStatus(ctx context.Context) (version int64, dirty bool, err error)
}

type HealthHandler struct {
	db DBChecker
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Schema   string `json:"schema,omitempty"`
}

func NewHealthHandler(db DBChecker) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:   "ok",
		Database: "not configured",
	}
	statusCode := http.StatusOK

	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			resp.Status = "degraded"
			resp.Database = "disconnected"
			statusCode = http.StatusServiceUnavailable
		} else {
			resp.Database = "connected"

			// If the DB can report migration state, fail on a dirty or
			// unapplied schema — a half-migrated DB pings fine but errors on
			// real queries, so a pool ping alone would falsely report healthy.
			if sc, ok := h.db.(SchemaChecker); ok {
				version, dirty, err := sc.MigrationStatus(ctx)
				switch {
				case err != nil:
					resp.Status = "degraded"
					resp.Schema = "migrations not applied"
					statusCode = http.StatusServiceUnavailable
				case dirty:
					resp.Status = "degraded"
					resp.Schema = fmt.Sprintf("dirty at version %d", version)
					statusCode = http.StatusServiceUnavailable
				default:
					resp.Schema = fmt.Sprintf("migrated (v%d)", version)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
